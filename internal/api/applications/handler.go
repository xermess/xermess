// Package applications answers the endpoints for the apps and services that
// sign their users in through this server with OAuth 2.0 and OpenID Connect:
// registering them, changing how they may authenticate, and rotating their
// secrets.
package applications

import (
	"errors"
	"log/slog"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"loginer/internal/api/audit"
	"loginer/internal/api/respond"
	"loginer/internal/brand"
	"loginer/internal/api/session"
	"loginer/internal/model"
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

// List returns a page of the applications the administrator can see, sorted
// by name, each with how many roles it defines.
func (h *Handler) List(c *gin.Context) {
	ctx := c.Request.Context()

	query := listQuery(c)
	query.Only = session.Reach(c, model.PermApplicationsRead)

	apps, total, err := h.store.Applications(ctx, query)
	if err != nil {
		respond.Failure(c, h.log, err, "listing applications failed")
		return
	}

	counts, err := h.store.ApplicationRoleCounts(ctx)
	if err != nil {
		respond.Failure(c, h.log, err, "counting application roles failed")
		return
	}

	c.JSON(http.StatusOK, newPageResponse(apps, counts, total, query))
}

// Get returns one application.
func (h *Handler) Get(c *gin.Context) {
	app, ok := h.find(c, model.PermApplicationsRead)
	if !ok {
		return
	}

	c.JSON(http.StatusOK, gin.H{"application": h.withCount(c, app)})
}

// Create registers an application. A confidential client is given its secret
// here, in the answer, and never again: only its hash is kept.
func (h *Handler) Create(c *gin.Context) {
	var req applicationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respond.Fail(c, respond.InvalidBody)
		return
	}

	app := &model.Application{
		Type:              model.ApplicationType(req.Type),
		IsEnabled:         true,
		AssertRoles:       true,
		RequirePKCE:       true,
		AllowRegistration: true,
	}
	if err := req.applyTo(app, true); err != nil {
		respond.Failure(c, h.log, err, "checking an application failed")
		return
	}

	clientID, err := model.NewClientID()
	if err != nil {
		respond.Failure(c, h.log, err, "making a client id failed")
		return
	}
	app.ClientID = clientID

	secret, err := issueIfNeeded(app)
	if err != nil {
		respond.Failure(c, h.log, err, "making a client secret failed")
		return
	}

	if err := h.store.CreateApplication(c.Request.Context(), app); err != nil {
		respond.Failure(c, h.log, err, "creating application failed")
		return
	}

	h.audit.Record(c, "application.created", targetType, app.ID.String())

	c.JSON(http.StatusCreated, newSecretResponse(h.withCount(c, app), secret))
}

// Update replaces an application's settings. Its type and client id stay as
// they are: every copy of the app already in use depends on both.
func (h *Handler) Update(c *gin.Context) {
	app, ok := h.find(c, model.PermApplicationsWrite)
	if !ok {
		return
	}
	if !h.changeable(c, app) {
		return
	}

	var req applicationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respond.Fail(c, respond.InvalidBody)
		return
	}

	if req.Type != "" && model.ApplicationType(req.Type) != app.Type {
		respond.BadRequest(c, "type cannot be changed: register a new application instead")
		return
	}

	if err := req.applyTo(app, false); err != nil {
		respond.Failure(c, h.log, err, "checking an application failed")
		return
	}

	// Moving from no secret to one — which only a confidential client can do
	// — needs a secret to exist, so it is issued and shown now.
	secret, err := issueIfNeeded(app)
	if err != nil {
		respond.Failure(c, h.log, err, "making a client secret failed")
		return
	}

	if err := h.store.SaveApplication(c.Request.Context(), app); err != nil {
		respond.Failure(c, h.log, err, "updating application failed")
		return
	}

	h.audit.Record(c, "application.updated", targetType, app.ID.String())

	c.JSON(http.StatusOK, newSecretResponse(h.withCount(c, app), secret))
}

// RotateSecret replaces a confidential client's secret. The old one stops
// working at once, so the app has to be given the new one straight away.
func (h *Handler) RotateSecret(c *gin.Context) {
	app, ok := h.find(c, model.PermApplicationsWrite)
	if !ok {
		return
	}
	if !h.changeable(c, app) {
		return
	}

	secret, err := app.IssueSecret(time.Now())
	if errors.Is(err, model.ErrNoSecret) {
		respond.BadRequest(c, "a public client has no secret to rotate")
		return
	}
	if err != nil {
		respond.Failure(c, h.log, err, "making a client secret failed")
		return
	}

	if err := h.store.SaveApplication(c.Request.Context(), app); err != nil {
		respond.Failure(c, h.log, err, "saving the new secret failed")
		return
	}

	h.audit.Record(c, "application.secret_rotated", targetType, app.ID.String())

	c.JSON(http.StatusOK, newSecretResponse(h.withCount(c, app), secret))
}

// Delete removes an application, the roles it defines, and everyone's hold on
// them. Tokens it has already issued stay valid until they expire.
func (h *Handler) Delete(c *gin.Context) {
	app, ok := h.find(c, model.PermApplicationsWrite)
	if !ok {
		return
	}

	// The server makes admin-cli again on its next start, so deleting it
	// would only lose its scopes; turning it off is how to stop it.
	if app.ClientID == model.AdminCLIClientID {
		respond.Fail(c, systemApplication)
		return
	}

	if err := h.store.DeleteApplication(c.Request.Context(), app); err != nil {
		respond.Failure(c, h.log, err, "deleting application failed")
		return
	}

	h.audit.Record(c, "application.deleted", targetType, app.ID.String())

	c.Status(http.StatusNoContent)
}

// systemApplication is the answer to deleting admin-cli.
var systemApplication = respond.Define(http.StatusConflict, "system_application", respond.Admin)

// changeable says whether the administrator may change this application or
// rotate its secret, answering the request itself when they may not.
//
// Only a super admin may authorise an application for the admin API
// (AuthorizeAPI) — but the power is in the secret, not in the authorisation.
// Whoever holds the secret of an application the admin API trusts gets a
// client credentials token carrying every scope it was allowed, and with it
// every permission a super admin ever gave it, so rotating that secret, or
// turning the application back on, re-adding a grant and lengthening its
// tokens, has to be a super admin's too. Anything else about such an
// application is a super admin's to change as well: the panel's own
// `applications.write` is a permission over ordinary applications.
func (h *Handler) changeable(c *gin.Context, app *model.Application) bool {
	if session.Admin(c).IsSuperAdmin() {
		return true
	}

	audience, err := h.store.AudienceFor(c.Request.Context(), app.ID, brand.AdminAPIIdentifier)
	if err != nil {
		// The admin API is made at startup, so not finding it is as much a
		// fault as any other — and the safe answer to a fault is no.
		respond.Failure(c, h.log, err, "checking an application's access to the admin API failed")
		return false
	}
	if audience.Authorized {
		respond.Fail(c, systemApplicationChange)
		return false
	}

	return true
}

// systemApplicationChange is the answer to changing an application the admin
// API trusts, or rotating its secret, without being a super admin.
var systemApplicationChange = respond.Define(http.StatusForbidden, "system_application_change", respond.Admin)

// APIAccess lists every API with what the application may do with it:
// whether it is authorised to ask for tokens for it, and which scopes.
func (h *Handler) APIAccess(c *gin.Context) {
	app, ok := h.find(c, model.PermApplicationsRead)
	if !ok {
		return
	}

	access, err := h.store.ApplicationAPIAccess(c.Request.Context(), app.ID)
	if err != nil {
		respond.Failure(c, h.log, err, "listing API access failed")
		return
	}

	c.JSON(http.StatusOK, newAccessResponse(access))
}

// AuthorizeAPI lets the application ask for tokens for an API, and replaces
// the API scopes it may ask for with the ones named. Those scopes are the
// ceiling on every token the application gets for the API.
func (h *Handler) AuthorizeAPI(c *gin.Context) {
	ctx := c.Request.Context()

	app, ok := h.find(c, model.PermApplicationsWrite)
	if !ok {
		return
	}

	api, ok := h.findAPI(c)
	if !ok {
		return
	}

	// A system API's scopes are power over this server itself: the admin
	// API's are the permission catalog, and the account API's act on any
	// user who signs in to the application. Handing them to an application
	// its administrator controls would give that administrator — who may
	// hold applications.write for this one application alone — every
	// permission in the panel, so only a super admin may.
	if api.System != "" && !session.Admin(c).IsSuperAdmin() {
		respond.Fail(c, systemAPIAccess)
		return
	}

	var req authorizeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respond.Fail(c, respond.InvalidBody)
		return
	}

	for _, id := range req.Scopes {
		if !slices.ContainsFunc(api.Scopes, func(scope model.APIScope) bool { return scope.ID == id }) {
			respond.BadRequest(c, "scopes: one of them is not a scope of "+api.Identifier)
			return
		}
	}

	if err := h.store.AuthorizeApplicationAPI(ctx, app.ID, api.ID, req.Scopes); err != nil {
		respond.Failure(c, h.log, err, "authorising the API failed")
		return
	}

	h.audit.RecordWith(c, "application.api_authorized", targetType, app.ID.String(), apiMetadata(app, api))

	h.APIAccess(c)
}

// systemAPIAccess is the answer to giving or taking an application's access to
// one of the server's own APIs without being a super admin.
var systemAPIAccess = respond.Define(http.StatusForbidden, "system_api_access", respond.Admin)

// RevokeAPI stops the application asking for tokens for an API.
func (h *Handler) RevokeAPI(c *gin.Context) {
	app, ok := h.find(c, model.PermApplicationsWrite)
	if !ok {
		return
	}

	api, ok := h.findAPI(c)
	if !ok {
		return
	}

	// The same lock as on giving the access: an administrator of one
	// application should not be able to cut admin-cli off from the admin API
	// when only a super admin could give it back.
	if api.System != "" && !session.Admin(c).IsSuperAdmin() {
		respond.Fail(c, systemAPIAccess)
		return
	}

	if err := h.store.RevokeApplicationAPI(c.Request.Context(), app.ID, api.ID); err != nil {
		respond.Failure(c, h.log, err, "revoking the API failed")
		return
	}

	h.audit.RecordWith(c, "application.api_revoked", targetType, app.ID.String(), apiMetadata(app, api))

	h.APIAccess(c)
}

// TokenPreview shows what a token request would amount to — whether a token
// is issued, the decision on every scope, and the claims each token carries —
// without issuing anything. It runs the same evaluation the token endpoint
// will, so what it shows is what an application will get.
func (h *Handler) TokenPreview(c *gin.Context) {
	ctx := c.Request.Context()

	app, ok := h.find(c, model.PermApplicationsRead)
	if !ok {
		return
	}

	var req previewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respond.Fail(c, respond.InvalidBody)
		return
	}

	request := model.TokenRequest{
		Application: *app,
		Requested:   strings.Fields(req.Scope),
		Issuer:      h.issuer,
		Now:         time.Now(),
	}

	if req.UserID != nil {
		if !session.Admin(c).HasPermission(model.PermUsersRead) {
			respond.Failure(c, h.log, respond.Forbidden, "")
			return
		}

		user, err := h.store.User(ctx, *req.UserID)
		if errors.Is(err, store.ErrNotFound) {
			respond.BadRequest(c, "user_id: no such user")
			return
		}
		if err != nil {
			respond.Failure(c, h.log, err, "loading the user failed")
			return
		}

		roles, err := h.store.EffectiveRoles(ctx, user)
		if err != nil {
			respond.Failure(c, h.log, err, "resolving roles failed")
			return
		}

		request.User = user
		request.Roles = roles
	}

	if identifier := strings.TrimSpace(req.Audience); identifier != "" {
		audience, err := h.store.AudienceFor(ctx, app.ID, identifier)
		if errors.Is(err, store.ErrNotFound) {
			respond.BadRequest(c, "audience: no API has that identifier")
			return
		}
		if err != nil {
			respond.Failure(c, h.log, err, "loading the audience failed")
			return
		}

		request.API = audience.API
		request.Authorized = audience.Authorized
		request.Allowed = audience.Allowed
	}

	c.JSON(http.StatusOK, gin.H{"preview": model.EvaluateToken(request)})
}

// apiMetadata is what an access change records beyond its target, so the API's
// own log can find it too.
func apiMetadata(app *model.Application, api *model.API) map[string]any {
	return map[string]any{
		"api_id":      api.ID.String(),
		"api":         api.Identifier,
		"application": app.Name,
	}
}

// findAPI loads the API named in the path, answering the request itself if
// there is no such API.
func (h *Handler) findAPI(c *gin.Context) (*model.API, bool) {
	id, err := uuid.Parse(c.Param("api"))
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

// withCount builds the response for one application with how many roles it
// defines. A count that cannot be read is not worth failing the request for.
func (h *Handler) withCount(c *gin.Context, app *model.Application) applicationResponse {
	counts, err := h.store.ApplicationRoleCounts(c.Request.Context())
	if err != nil {
		h.log.Error("counting application roles failed", "error", err)
	}

	return newApplicationResponse(*app, counts[app.ID])
}

// find loads the application named in the path and checks the administrator
// holds `permission` for it, answering the request itself otherwise. An
// application they cannot read is not found, rather than forbidden, so its
// existence is not given away.
func (h *Handler) find(c *gin.Context, permission string) (*model.Application, bool) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respond.BadRequest(c, "that is not an application id")
		return nil, false
	}

	if !session.Allowed(c, model.PermApplicationsRead, id) {
		respond.NotFound(c, "no such application")
		return nil, false
	}

	if !session.Allowed(c, permission, id) {
		respond.Failure(c, h.log, respond.Forbidden, "")
		return nil, false
	}

	app, err := h.store.Application(c.Request.Context(), id)

	switch {
	case errors.Is(err, store.ErrNotFound):
		respond.NotFound(c, "no such application")
		return nil, false
	case err != nil:
		respond.Failure(c, h.log, err, "loading application failed")
		return nil, false
	}

	return app, true
}

// issueIfNeeded gives a confidential client that has no secret yet its first
// one, and returns it. Anything else returns "".
func issueIfNeeded(app *model.Application) (string, error) {
	if !app.HasSecret() || app.ClientSecretHash != "" {
		return "", nil
	}

	return app.IssueSecret(time.Now())
}
