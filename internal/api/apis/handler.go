// Package apis answers the endpoints for resource servers: the services
// applications ask for access tokens to call, each with an audience and the
// scopes its tokens may carry.
package apis

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"loginer/internal/api/audit"
	"loginer/internal/api/respond"
	"loginer/internal/api/session"
	"loginer/internal/model"
	"loginer/internal/oidc"
	"loginer/internal/store"
)

// Handler holds what these endpoints need.
type Handler struct {
	store *store.Store
	audit audit.Recorder
	log   *slog.Logger

	// issuer is the provider's: the iss its tokens carry.
	issuer string
}

// New returns a Handler. `issuer` is LOGINER_ISSUER, which tokens name the
// server by.
func New(st *store.Store, recorder audit.Recorder, log *slog.Logger, issuer string) *Handler {
	return &Handler{store: st, audit: recorder, log: log, issuer: issuer}
}

// List returns every API matching the search, sorted by name, with its scopes
// and how many applications may use it.
func (h *Handler) List(c *gin.Context) {
	ctx := c.Request.Context()

	apis, err := h.store.APIs(ctx, c.Query("search"))
	if err != nil {
		respond.Failure(c, h.log, err, "listing APIs failed")
		return
	}

	details, err := h.details(c)
	if err != nil {
		respond.Failure(c, h.log, err, "counting API applications and roles failed")
		return
	}

	c.JSON(http.StatusOK, newListResponse(apis, details))
}

// Get returns one API.
func (h *Handler) Get(c *gin.Context) {
	api, ok := h.find(c)
	if !ok {
		return
	}

	c.JSON(http.StatusOK, gin.H{"api": h.withCount(c, api)})
}

// Create registers an API with its scopes.
func (h *Handler) Create(c *gin.Context) {
	var req apiRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respond.Fail(c, respond.InvalidBody)
		return
	}

	api := &model.API{EnforceRoles: true}
	if err := req.applyTo(api, true); err != nil {
		respond.Failure(c, h.log, err, "checking an API failed")
		return
	}

	if err := h.store.CreateAPI(c.Request.Context(), api); err != nil {
		h.respondWrite(c, err, "creating API failed")
		return
	}

	h.audit.Record(c, "api.created", targetType, api.ID.String())

	h.answer(c, http.StatusCreated, api.ID)
}

// Update replaces an API's name, description, role enforcement and scopes. Its
// identifier stays as it is: every service checking tokens compares against
// it.
func (h *Handler) Update(c *gin.Context) {
	api, ok := h.find(c)
	if !ok {
		return
	}

	var req apiRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respond.Fail(c, respond.InvalidBody)
		return
	}

	if req.Identifier != "" && req.Identifier != api.Identifier {
		respond.BadRequest(c, "identifier cannot be changed: register a new API instead")
		return
	}

	if err := req.applyTo(api, false); err != nil {
		respond.Failure(c, h.log, err, "checking an API failed")
		return
	}

	if err := h.store.SaveAPI(c.Request.Context(), api); err != nil {
		h.respondWrite(c, err, "updating API failed")
		return
	}

	h.audit.Record(c, "api.updated", targetType, api.ID.String())

	h.answer(c, http.StatusOK, api.ID)
}

// Delete removes an API, its scopes, every application's authorisation for it
// and every role's grant of its scopes. Tokens already issued for it stay
// valid until they expire.
func (h *Handler) Delete(c *gin.Context) {
	api, ok := h.find(c)
	if !ok {
		return
	}

	if err := h.store.DeleteAPI(c.Request.Context(), api); err != nil {
		respond.Failure(c, h.log, err, "deleting API failed")
		return
	}

	h.audit.Record(c, "api.deleted", targetType, api.ID.String())

	c.Status(http.StatusNoContent)
}

// answer writes an API as stored, read back so new scopes carry their ids.
func (h *Handler) answer(c *gin.Context, status int, id uuid.UUID) {
	api, err := h.store.API(c.Request.Context(), id)
	if err != nil {
		respond.Failure(c, h.log, err, "reading the API back failed")
		return
	}

	c.JSON(status, gin.H{"api": h.withCount(c, api)})
}

// Applications lists the applications the administrator can see, with what
// each may do with the API. Granting and revoking access goes through the
// application's own endpoints, which check the administrator may change it.
func (h *Handler) Applications(c *gin.Context) {
	api, ok := h.find(c)
	if !ok {
		return
	}

	apps, err := h.store.APIApplications(c.Request.Context(), api.ID, session.Reach(c, model.PermApplicationsRead))
	if err != nil {
		respond.Failure(c, h.log, err, "listing the API's applications failed")
		return
	}

	c.JSON(http.StatusOK, newApplicationsResponse(apps))
}

// Logs lists what has happened to or involving the API, newest first:
// changes to it, and applications gaining or losing access. Token requests
// join these once the token endpoint issues tokens.
func (h *Handler) Logs(c *gin.Context) {
	api, ok := h.find(c)
	if !ok {
		return
	}

	events, err := h.store.APIAuditLog(c.Request.Context(), api.ID, maxLogEntries)
	if err != nil {
		respond.Failure(c, h.log, err, "reading the API's log failed")
		return
	}

	c.JSON(http.StatusOK, newLogResponse(events))
}

// details reads what every API response carries beyond the API: counts, and
// where this server is. A count that cannot be read fails the list, but not a
// write that has already happened; see withCount.
func (h *Handler) details(c *gin.Context) (apiDetails, error) {
	ctx := c.Request.Context()

	applications, err := h.store.APIApplicationCounts(ctx)
	if err != nil {
		return apiDetails{}, err
	}

	roles, err := h.store.APIRoleCounts(ctx)
	if err != nil {
		return apiDetails{}, err
	}

	return apiDetails{
		applications: applications,
		roles:        roles,
		issuer:       h.issuer,
		jwksURI:      h.issuer + oidc.PathJWKS,
	}, nil
}

// withCount builds the response for one API with its counts. A count that
// cannot be read is not worth failing the request for.
func (h *Handler) withCount(c *gin.Context, api *model.API) apiResponse {
	details, err := h.details(c)
	if err != nil {
		h.log.Error("counting API applications and roles failed", "error", err)
		details = apiDetails{issuer: h.issuer, jwksURI: h.issuer + oidc.PathJWKS}
	}

	return newAPIResponse(*api, details)
}

// find loads the API named in the path, answering the request itself if there
// is no such API.
func (h *Handler) find(c *gin.Context) (*model.API, bool) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respond.BadRequest(c, "that is not an API id")
		return nil, false
	}

	api, err := h.store.API(c.Request.Context(), id)

	switch {
	case errors.Is(err, store.ErrNotFound):
		respond.NotFound(c, "no such API")
		return nil, false
	case err != nil:
		respond.Failure(c, h.log, err, "loading API failed")
		return nil, false
	}

	return api, true
}

// respondWrite turns a failed write into an answer, telling a taken identifier
// apart from anything else because that one is the writer's to fix.
func (h *Handler) respondWrite(c *gin.Context, err error, note string) {
	if errors.Is(err, store.ErrDuplicate) {
		respond.Conflict(c, "an API with that identifier already exists")
		return
	}

	respond.Failure(c, h.log, err, note)
}
