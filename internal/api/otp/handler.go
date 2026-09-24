// Package otp answers the endpoints for the one-time codes this server emails
// people as they sign in: how long a code is, how long it lasts, how many
// guesses it takes, and how soon another may be asked for.
//
// There is one record of it, like the organisation's, so there is no list and
// nothing to create or delete: the two endpoints read the settings and write
// them back. Which sign-ins ask for a code is not here — that is the login
// flow's emailed code step, on the Login flows page.
package otp

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"xermess/internal/api/audit"
	"xermess/internal/api/respond"
	"xermess/internal/store"
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

// Get returns the settings, and which login flows ask for a code — so the
// page can say whether any of this is being used, and lead to the flow that
// uses it.
func (h *Handler) Get(c *gin.Context) {
	ctx := c.Request.Context()

	settings, err := h.store.OTPSettings(ctx)
	if err != nil {
		respond.Failure(c, h.log, err, "loading the one-time code settings failed")
		return
	}

	flows, err := h.store.LoginFlows(ctx)
	if err != nil {
		respond.Failure(c, h.log, err, "listing login flows failed")
		return
	}

	c.JSON(http.StatusOK, newResponse(*settings, flows))
}

// Update changes the settings. What a request leaves out is left as it is.
func (h *Handler) Update(c *gin.Context) {
	ctx := c.Request.Context()

	settings, err := h.store.OTPSettings(ctx)
	if err != nil {
		respond.Failure(c, h.log, err, "loading the one-time code settings failed")
		return
	}

	var req settingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respond.Fail(c, respond.InvalidBody)
		return
	}

	changed, err := req.applyTo(settings)
	if err != nil {
		respond.Failure(c, h.log, err, "validating the one-time code settings failed")
		return
	}

	flows, err := h.store.LoginFlows(ctx)
	if err != nil {
		respond.Failure(c, h.log, err, "listing login flows failed")
		return
	}

	if len(changed) == 0 {
		c.JSON(http.StatusOK, newResponse(*settings, flows))
		return
	}

	if err := h.store.SaveOTPSettings(ctx, settings); err != nil {
		respond.Failure(c, h.log, err, "updating the one-time code settings failed")
		return
	}

	// A code that is shorter or lasts longer is a change to how hard this
	// server is to get into, so the log says which way it moved.
	h.audit.RecordWith(c, "otp.settings_updated", targetType, targetID, map[string]any{
		"fields":           changed,
		"length":           settings.Length,
		"lifetime_minutes": settings.LifetimeMinutes,
		"max_attempts":     settings.MaxAttempts,
	})

	c.JSON(http.StatusOK, newResponse(*settings, flows))
}
