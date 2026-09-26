// Package roles answers the endpoints for the roles users hold: global roles,
// which belong to no application, and application roles, which belong to
// one. What a role lets someone do is the applications' business; these
// endpoints only say which roles exist, what they include, and who holds them.
//
// Every request is allowed or not by the role's scope: what the administrator
// holds for its application, or for the whole panel when the role is global.
package roles

import (
	"errors"
	"log/slog"
	"net/http"
	"slices"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"loginer/internal/api/audit"
	"loginer/internal/api/respond"
	"loginer/internal/api/session"
	"loginer/internal/model"
	"loginer/internal/store"
)

// Handler holds what these endpoints need.
type Handler struct {
	store *store.Store
	audit audit.Recorder
	log   *slog.Logger
}

// New returns a Handler.
func New(st *store.Store, recorder audit.Recorder, log *slog.Logger) *Handler {
	return &Handler{store: st, audit: recorder, log: log}
}

// The values of the scope query parameter.
const (
	globalScope      = "global"
	applicationScope = "application"
)

// List returns a page of roles, sorted by name, each with how many users hold
// it and every role it includes once inheritance is followed.
//
// Two query parameters say which roles. scope is "global" for the global
// roles, "application" for the roles of every application the administrator
// can see, and empty for both. application narrows to one application's
// roles, whatever scope says.
func (h *Handler) List(c *gin.Context) {
	ctx := c.Request.Context()
	query := listQuery(c)

	reach := session.Reach(c, model.PermApplicationsRead)

	switch scope := c.Query("scope"); {
	case c.Query("application") != "":
		app, err := uuid.Parse(c.Query("application"))
		if err != nil {
			respond.BadRequest(c, "that is not an application id")
			return
		}

		if !session.Allowed(c, model.PermApplicationsRead, app) {
			respond.NotFound(c, "no such application")
			return
		}

		query.Applications = []uuid.UUID{app}
	case scope == globalScope:
		query.Global = true
		query.Applications = []uuid.UUID{}
	case scope == applicationScope:
		query.Applications = reach
	case scope == "":
		query.Global = true
		query.Applications = reach
	default:
		respond.BadRequest(c, `scope must be "global" or "application"`)
		return
	}

	roles, total, err := h.store.UserRoles(ctx, query)
	if err != nil {
		respond.Failure(c, h.log, err, "listing roles failed")
		return
	}

	details, err := h.details(c, roles)
	if err != nil {
		respond.Failure(c, h.log, err, "resolving roles failed")
		return
	}

	c.JSON(http.StatusOK, newPageResponse(roles, details, total, query))
}

// Get returns one role.
func (h *Handler) Get(c *gin.Context) {
	role, ok := h.find(c, "")
	if !ok {
		return
	}

	h.answer(c, http.StatusOK, role)
}

// Create adds a role: a global one when application_id is null, otherwise
// one of that application's.
func (h *Handler) Create(c *gin.Context) {
	var req roleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respond.Fail(c, respond.InvalidBody)
		return
	}

	if app := req.ApplicationID; app != nil {
		if !session.Allowed(c, model.PermApplicationsRead, *app) {
			respond.BadRequest(c, "application_id: no such application")
			return
		}

		if _, err := h.store.Application(c.Request.Context(), *app); err != nil {
			if errors.Is(err, store.ErrNotFound) {
				respond.BadRequest(c, "application_id: no such application")
				return
			}

			respond.Failure(c, h.log, err, "loading the role's application failed")
			return
		}
	}

	if !session.AllowedScope(c, model.PermUserRolesWrite, req.ApplicationID) {
		respond.Failure(c, h.log, respond.Forbidden, "")
		return
	}

	role, err := h.build(c, &req, nil)
	if err != nil {
		respond.Failure(c, h.log, err, "checking a role failed")
		return
	}

	if err := h.store.CreateUserRole(c.Request.Context(), role); err != nil {
		h.respondWrite(c, err, "creating role failed")
		return
	}

	h.audit.Record(c, "user_role.created", targetType, role.ID.String())

	h.answer(c, http.StatusCreated, role)
}

// Update replaces a role's name, description, default flag and what it
// includes. Its scope stays as it is: a role does not move between global and
// an application, or between applications.
func (h *Handler) Update(c *gin.Context) {
	role, ok := h.find(c, model.PermUserRolesWrite)
	if !ok {
		return
	}

	var req roleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respond.Fail(c, respond.InvalidBody)
		return
	}

	if _, err := h.build(c, &req, role); err != nil {
		respond.Failure(c, h.log, err, "checking a role failed")
		return
	}

	if err := h.store.SaveUserRole(c.Request.Context(), role); err != nil {
		h.respondWrite(c, err, "updating role failed")
		return
	}

	h.audit.Record(c, "user_role.updated", targetType, role.ID.String())

	h.answer(c, http.StatusOK, role)
}

// Delete removes a role for good. The users who held it keep everything else
// they hold, and the roles that included it simply stop.
func (h *Handler) Delete(c *gin.Context) {
	role, ok := h.find(c, model.PermUserRolesWrite)
	if !ok {
		return
	}

	if err := h.store.DeleteUserRole(c.Request.Context(), role); err != nil {
		respond.Failure(c, h.log, err, "deleting role failed")
		return
	}

	h.audit.Record(c, "user_role.deleted", targetType, role.ID.String())

	c.Status(http.StatusNoContent)
}

// build checks a submitted role and returns the role to write. Passing an
// existing role fills that one in instead of making a new one, whose scope is
// then the request's.
//
// The included roles are named by id; every one has to exist, be one the
// administrator can see, and be one the role may include. Including is
// refused when it would make a role include itself.
func (h *Handler) build(c *gin.Context, req *roleRequest, into *model.UserRole) (*model.UserRole, error) {
	ctx := c.Request.Context()

	if err := req.validate(); err != nil {
		return nil, err
	}

	inherits, err := h.store.UserRolesByID(ctx, req.Inherits)
	if errors.Is(err, store.ErrNotFound) {
		return nil, respond.Fault{Status: http.StatusBadRequest, Message: "inherits: one of them does not exist"}
	}
	if err != nil {
		return nil, err
	}

	role := into
	if role == nil {
		role = &model.UserRole{ApplicationID: req.ApplicationID}
	}

	for _, inherited := range inherits {
		if !session.SeesRole(c, inherited) {
			return nil, respond.Fault{Status: http.StatusBadRequest, Message: "inherits: one of them does not exist"}
		}

		// Including a role gives it to everyone holding this one, so adding
		// it takes the right to manage roles where it belongs. Otherwise an
		// administrator of one application could slip a global role into one
		// of its roles and hand that out. What the role already included
		// stays, whoever put it there.
		added := !slices.ContainsFunc(role.Inherits, func(held model.UserRole) bool { return held.ID == inherited.ID })
		if added && !session.AllowedScope(c, model.PermUserRolesWrite, inherited.ApplicationID) {
			return nil, respond.Fault{
				Status:  http.StatusForbidden,
				Message: "inherits: your roles do not allow you to include " + inherited.Name,
			}
		}

		if !role.MayInherit(inherited) {
			return nil, respond.Fault{
				Status: http.StatusBadRequest,
				Message: "inherits: " + inherited.Name + " belongs to another application; " +
					"an application role can include global roles and its own application's roles",
			}
		}
	}

	if into != nil && len(inherits) > 0 {
		graph, err := h.store.RoleGraph(ctx)
		if err != nil {
			return nil, err
		}

		if graph.WouldCycle(role.ID, req.Inherits) {
			return nil, respond.Fault{
				Status:  http.StatusBadRequest,
				Message: "inherits: a role cannot include itself, or a role that includes it",
			}
		}
	}

	if req.APIScopes != nil {
		if !session.Admin(c).HasPermission(model.PermAPIsRead) {
			return nil, respond.Fault{
				Status:  http.StatusForbidden,
				Message: "api_scopes: your roles do not allow you to see APIs",
			}
		}

		scopes, err := h.store.APIScopesByID(ctx, *req.APIScopes)
		if errors.Is(err, store.ErrNotFound) {
			return nil, respond.Fault{Status: http.StatusBadRequest, Message: "api_scopes: one of them does not exist"}
		}
		if err != nil {
			return nil, err
		}

		role.APIScopes = scopes
	}

	role.Name = req.Name
	role.Description = req.Description
	role.IsDefault = req.IsDefault
	role.Inherits = inherits

	return role, nil
}

// details is what a page of roles says beyond the rows: how many users hold
// each, and what each includes once inheritance is followed.
func (h *Handler) details(c *gin.Context, roles []model.UserRole) (roleDetails, error) {
	ctx := c.Request.Context()

	ids := make([]uuid.UUID, 0, len(roles))
	for _, role := range roles {
		ids = append(ids, role.ID)
	}

	members, err := h.store.RoleMemberCounts(ctx, ids)
	if err != nil {
		return roleDetails{}, err
	}

	graph, err := h.store.RoleGraph(ctx)
	if err != nil {
		return roleDetails{}, err
	}

	return roleDetails{members: members, graph: graph, visible: func(role model.UserRole) bool {
		return session.SeesRole(c, role)
	}}, nil
}

// answer writes one role with its details.
func (h *Handler) answer(c *gin.Context, status int, role *model.UserRole) {
	details, err := h.details(c, []model.UserRole{*role})
	if err != nil {
		respond.Failure(c, h.log, err, "resolving role failed")
		return
	}

	// The graph was read after the write, so it has the role as stored; this
	// is the same role, and saves relying on that.
	details.graph[role.ID] = *role

	c.JSON(status, gin.H{"role": newRoleResponse(*role, details)})
}

// find loads the role named in the path and, when `permission` is given,
// checks the administrator holds it for the role's scope, answering the
// request itself otherwise. A role the administrator cannot see is not found.
func (h *Handler) find(c *gin.Context, permission string) (*model.UserRole, bool) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respond.BadRequest(c, "that is not a role id")
		return nil, false
	}

	role, err := h.store.UserRole(c.Request.Context(), id)

	switch {
	case errors.Is(err, store.ErrNotFound), err == nil && !session.SeesRole(c, *role):
		respond.NotFound(c, "no such role")
		return nil, false
	case err != nil:
		respond.Failure(c, h.log, err, "loading role failed")
		return nil, false
	case permission != "" && !session.AllowedScope(c, permission, role.ApplicationID):
		respond.Failure(c, h.log, respond.Forbidden, "")
		return nil, false
	}

	return role, true
}

// respondWrite turns a failed write into an answer, telling a duplicate name
// apart from anything else because that one is the writer's to fix.
func (h *Handler) respondWrite(c *gin.Context, err error, note string) {
	if errors.Is(err, store.ErrDuplicate) {
		respond.Conflict(c, "a role with that name already exists in this scope")
		return
	}

	respond.Failure(c, h.log, err, note)
}
