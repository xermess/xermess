// Package auth answers the endpoints that sign an administrator in and out,
// and the ones that say who they are: the panel's whole idea of a session.
package auth

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"xermess/internal/api/audit"
	"xermess/internal/api/respond"
	"xermess/internal/api/session"
	authsvc "xermess/internal/auth"
	"xermess/internal/model"
	"xermess/internal/store"
)

// The problems these endpoints answer with that are their own.
var (
	wrongPassword    = respond.Define(http.StatusBadRequest, "admin_wrong_password", respond.Admin)
	emailTaken       = respond.Define(http.StatusConflict, "admin_email_taken", respond.Admin)
	passwordTooShort = respond.Define(http.StatusBadRequest, "admin_password_too_short", respond.Admin)
	passwordTooLong  = respond.Define(http.StatusBadRequest, "admin_password_too_long", respond.Admin)
)

// targetType is what an administrator is called in the activity log.
const targetType = "admin_user"

// Handler holds what these endpoints need.
type Handler struct {
	auth  *authsvc.Service
	store *store.Store
	audit audit.Recorder
	log   *slog.Logger

	// secure sets the cookie's Secure flag: true once served over HTTPS.
	secure bool
}

// New returns a Handler.
func New(service *authsvc.Service, st *store.Store, recorder audit.Recorder, log *slog.Logger, secure bool) *Handler {
	return &Handler{auth: service, store: st, audit: recorder, log: log, secure: secure}
}

// Login checks the credentials and sets the session cookie.
func (h *Handler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respond.Fail(c, respond.InvalidBody)
		return
	}

	if err := req.validate(); err != nil {
		respond.Failure(c, h.log, err, "validating a sign-in failed")
		return
	}

	token, admin, state, err := h.auth.Login(c.Request.Context(), req.Username, req.Password, requestOf(c))
	switch {
	case errors.Is(err, authsvc.ErrInvalidCredentials):
		// One message for every kind of failure: which usernames exist is not
		// something a sign-in page should reveal.
		respond.Error(c, http.StatusUnauthorized, "wrong username or password")
		return
	case err != nil:
		respond.Failure(c, h.log, err, "login failed")
		return
	}

	session.Set(c, token, h.secure)

	// A right password with a second step still to go says which step; the
	// administrator is not shown until they are through it.
	if state != authsvc.StateSignedIn {
		c.JSON(http.StatusOK, gin.H{"next": state})
		return
	}

	c.JSON(http.StatusOK, gin.H{"admin": newAdminResponse(admin)})
}

// State says how far the session this browser carries has got, so the sign-in
// page knows whether to ask for a password, a code, or to set up an
// authenticator. It needs no session: saying there is none is an answer.
func (h *Handler) State(c *gin.Context) {
	token, _ := c.Cookie(session.Cookie)

	_, _, state, err := h.auth.Session(c.Request.Context(), token)
	if err != nil && !errors.Is(err, authsvc.ErrNoSession) {
		respond.Failure(c, h.log, err, "reading the session failed")
		return
	}

	c.JSON(http.StatusOK, gin.H{"state": state, "mfa_required": h.auth.MFARequired(c.Request.Context())})
}

// VerifyMFA finishes a sign-in waiting for a second factor.
func (h *Handler) VerifyMFA(c *gin.Context) {
	var req codeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respond.Fail(c, respond.InvalidBody)
		return
	}

	token, _ := c.Cookie(session.Cookie)

	admin, err := h.auth.VerifySignIn(c.Request.Context(), token, req.Code, requestOf(c))
	switch {
	case errors.Is(err, authsvc.ErrNoSession):
		respond.Error(c, http.StatusUnauthorized, "your sign-in has expired; enter your password again")
		return
	case errors.Is(err, authsvc.ErrInvalidCode):
		respond.Error(c, http.StatusUnauthorized, err.Error())
		return
	case err != nil:
		respond.Failure(c, h.log, err, "verifying a second factor failed")
		return
	}

	c.JSON(http.StatusOK, gin.H{"admin": newAdminResponse(admin)})
}

// Logout revokes the session and clears the cookie.
func (h *Handler) Logout(c *gin.Context) {
	token, _ := c.Cookie(session.Cookie)

	if err := h.auth.Logout(c.Request.Context(), token, requestOf(c)); err != nil {
		h.log.Error("logout failed", "error", err)
	}

	session.Clear(c, h.secure)

	c.JSON(http.StatusOK, gin.H{"status": "signed out"})
}

// Me returns the signed-in administrator, and is what the browser calls to
// find out whether it still has a session.
func (h *Handler) Me(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"admin": newAdminResponse(session.Admin(c))})
}

// UpdateMe changes the caller's own name and address. It reaches nothing
// else about the account — not the roles, not the status — so every
// administrator may use it, whatever their roles allow.
func (h *Handler) UpdateMe(c *gin.Context) {
	var req profileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respond.Fail(c, respond.InvalidBody)
		return
	}
	if err := req.validate(); err != nil {
		respond.Failure(c, h.log, err, "validating a profile failed")
		return
	}

	admin := session.Admin(c)
	profile := authsvc.Profile{FirstName: req.FirstName, LastName: req.LastName, Email: req.Email}

	err := h.auth.UpdateProfile(c.Request.Context(), admin, profile, req.CurrentPassword)
	switch {
	case errors.Is(err, authsvc.ErrWrongPassword):
		respond.Fail(c, wrongPassword)
		return
	case errors.Is(err, store.ErrDuplicate):
		respond.Fail(c, emailTaken)
		return
	case err != nil:
		respond.Failure(c, h.log, err, "updating a profile failed")
		return
	}

	h.audit.Record(c, "admin.profile_updated", targetType, admin.ID.String())

	c.JSON(http.StatusOK, gin.H{"admin": newAdminResponse(admin)})
}

// ChangePassword sets the caller's own password, given the one they have.
// Every other session they have open ends; the one they are using stays.
func (h *Handler) ChangePassword(c *gin.Context) {
	var req passwordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respond.Fail(c, respond.InvalidBody)
		return
	}
	if err := req.validate(); err != nil {
		respond.Failure(c, h.log, err, "validating a password change failed")
		return
	}

	admin := session.Admin(c)

	err := h.auth.ChangePassword(c.Request.Context(), admin, session.ID(c), req.CurrentPassword, req.NewPassword)
	switch {
	case errors.Is(err, authsvc.ErrWrongPassword):
		respond.Fail(c, wrongPassword)
		return
	case errors.Is(err, model.ErrPasswordTooLong):
		respond.Fail(c, passwordTooLong)
		return
	case err != nil:
		respond.Failure(c, h.log, err, "changing a password failed")
		return
	}

	h.audit.Record(c, "admin.password_changed", targetType, admin.ID.String())

	c.JSON(http.StatusOK, gin.H{"status": "changed"})
}

// Sessions lists the caller's own sessions, so they can see where they are
// signed in.
func (h *Handler) Sessions(c *gin.Context) {
	sessions, err := h.store.SessionsFor(c.Request.Context(), session.Admin(c).ID, 20)
	if err != nil {
		respond.Failure(c, h.log, err, "listing sessions failed")
		return
	}

	c.JSON(http.StatusOK, gin.H{"sessions": newSessionResponses(sessions)})
}

// requestOf describes where the call came from, for the session and the log.
func requestOf(c *gin.Context) authsvc.Request {
	return authsvc.Request{
		IP:        c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
	}
}
