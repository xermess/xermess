// Package account answers the endpoints behind the id app, the users' app:
// signing in, creating an account and resetting a forgotten password, and —
// once signed in — managing the account itself.
//
// The first half is the JSON side of the authorization endpoint. That
// endpoint sends a browser to the sign-in page with a handle for the sign-in
// under way; the page asks here what to show, and posts what the user typed.
// Once the user is signed in, the answer says where to send the browser next —
// back to the application, with a code.
//
// The second half sits behind RequireSession, and only ever reaches the
// account whose session cookie the request carries.
package account

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"xermess/internal/api/respond"
	"xermess/internal/api/session"
	"xermess/internal/oidc"
	"xermess/internal/store"
)

// Handler holds what these endpoints need.
type Handler struct {
	provider *oidc.Service
	log      *slog.Logger
	secure   bool
}

// New returns a Handler. `secure` sets the Secure flag on the session cookie.
func New(provider *oidc.Service, log *slog.Logger, secure bool) *Handler {
	return &Handler{provider: provider, log: log, secure: secure}
}

// Request describes a sign-in under way: the application it is for, as its
// sign-in page shows it.
func (h *Handler) Request(c *gin.Context) {
	pending, err := h.provider.Pending(c.Request.Context(), c.Param("handle"))
	if err != nil {
		h.fail(c, err, "loading a sign-in request failed")
		return
	}

	c.JSON(http.StatusOK, newRequestResponse(pending))
}

// Application describes an application by its client id, for the signed-out
// page to offer a way back to it.
func (h *Handler) Application(c *gin.Context) {
	app, err := h.provider.ApplicationByClientID(c.Request.Context(), c.Param("client_id"))
	if errors.Is(err, store.ErrNotFound) {
		respond.NotFound(c, "no such application")
		return
	}
	if err != nil {
		h.fail(c, err, "loading an application failed")
		return
	}

	c.JSON(http.StatusOK, gin.H{"application": app})
}

// Organization describes the organisation these pages sign users in for: who
// the account belongs to, where to ask for help, and the agreements accepted
// by making one. Every sign-in page asks for it, with or without a sign-in
// under way, so it needs no session and names nothing about the caller.
func (h *Handler) Organization(c *gin.Context) {
	organization, err := h.provider.Organization(c.Request.Context())
	if err != nil {
		h.fail(c, err, "loading the organization failed")
		return
	}

	c.JSON(http.StatusOK, gin.H{"organization": organization})
}

// Login signs a user in and, for a sign-in under way, says where to go next.
func (h *Handler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respond.BadRequest(c, "the request body is not valid")
		return
	}
	if err := req.validate(); err != nil {
		respond.Failure(c, h.log, err, "validating a sign-in failed")
		return
	}

	result, err := h.provider.SignIn(c.Request.Context(), req.Email, req.Password, client(c))
	if err != nil {
		h.fail(c, err, "signing a user in failed")
		return
	}

	h.signedIn(c, req.Request, result)
}

// Register creates an account for a sign-in under way, and signs it in.
func (h *Handler) Register(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respond.BadRequest(c, "the request body is not valid")
		return
	}
	if err := req.validate(); err != nil {
		respond.Failure(c, h.log, err, "validating a registration failed")
		return
	}

	result, err := h.provider.Register(c.Request.Context(), oidc.Registration{
		Request:       req.Request,
		Email:         req.Email,
		Password:      req.Password,
		FirstName:     req.FirstName,
		LastName:      req.LastName,
		AcceptedTerms: req.AcceptTerms,
	}, client(c))
	if err != nil {
		h.fail(c, err, "registering a user failed")
		return
	}

	h.signedIn(c, req.Request, result)
}

// signedIn finishes a sign-in or registration: the cookie, and where to go.
func (h *Handler) signedIn(c *gin.Context, request string, result *oidc.SignInResult) {
	if result.ResetToken != "" {
		c.JSON(http.StatusOK, signedInResponse{PasswordChangeRequired: true, ResetToken: result.ResetToken})
		return
	}

	session.SetUser(c, result.Token, h.secure)

	if request == "" {
		c.JSON(http.StatusOK, signedInResponse{})
		return
	}

	location, err := h.provider.Continue(c.Request.Context(), request, result.Session)
	if err != nil {
		h.fail(c, err, "continuing a sign-in failed")
		return
	}

	c.JSON(http.StatusOK, signedInResponse{RedirectTo: location})
}

// ForgotPassword sends a reset link. It answers the same whether or not the
// address has an account.
func (h *Handler) ForgotPassword(c *gin.Context) {
	var req forgotRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respond.BadRequest(c, "the request body is not valid")
		return
	}
	if err := req.validate(); err != nil {
		respond.Failure(c, h.log, err, "validating a password reset request failed")
		return
	}

	if err := h.provider.ForgotPassword(c.Request.Context(), req.Email, req.Request, client(c)); err != nil {
		h.fail(c, err, "starting a password reset failed")
		return
	}

	c.JSON(http.StatusAccepted, gin.H{"status": "sent"})
}

// CheckReset says whether a reset link still works.
func (h *Handler) CheckReset(c *gin.Context) {
	err := h.provider.CheckReset(c.Request.Context(), c.Query("token"))
	if err != nil && !errors.Is(err, oidc.ErrResetInvalid) {
		h.fail(c, err, "checking a reset link failed")
		return
	}

	c.JSON(http.StatusOK, gin.H{"valid": err == nil})
}

// ResetPassword sets a new password through a reset link.
func (h *Handler) ResetPassword(c *gin.Context) {
	var req resetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respond.BadRequest(c, "the request body is not valid")
		return
	}
	if err := req.validate(); err != nil {
		respond.Failure(c, h.log, err, "validating a password reset failed")
		return
	}

	if err := h.provider.ResetPassword(c.Request.Context(), req.Token, req.Password, client(c)); err != nil {
		h.fail(c, err, "resetting a password failed")
		return
	}

	// Every session the user had has ended, this browser's included.
	session.ClearUser(c, h.secure)

	c.JSON(http.StatusOK, gin.H{"status": "reset"})
}

// Logout signs the user out of this server in this browser.
func (h *Handler) Logout(c *gin.Context) {
	token, _ := c.Cookie(session.UserCookie)

	if err := h.provider.SignOut(c.Request.Context(), token, client(c)); err != nil {
		h.log.Error("signing a user out failed", "error", err)
	}

	session.ClearUser(c, h.secure)

	c.JSON(http.StatusOK, gin.H{"status": "signed out"})
}

// sessionKey is the Gin context key RequireSession stores the session under.
const sessionKey = "user_session"

// RequireSession refuses the request unless it carries a user's session, and
// hands the session to the handlers behind it.
func (h *Handler) RequireSession(c *gin.Context) {
	token, _ := c.Cookie(session.UserCookie)

	current, err := h.provider.SessionFor(c.Request.Context(), token)
	if err != nil {
		h.fail(c, err, "loading the user session failed")
		c.Abort()
		return
	}
	if current == nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "not signed in"})
		return
	}

	c.Set(sessionKey, current)
	c.Next()
}

func signedInSession(c *gin.Context) *oidc.Session {
	value, _ := c.Get(sessionKey)
	current, _ := value.(*oidc.Session)
	return current
}

// Me returns the signed-in user.
func (h *Handler) Me(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"user": newUserResponse(signedInSession(c).User)})
}

// UpdateMe changes the signed-in user's name.
func (h *Handler) UpdateMe(c *gin.Context) {
	var req profileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respond.BadRequest(c, "the request body is not valid")
		return
	}
	if err := req.validate(); err != nil {
		respond.Failure(c, h.log, err, "validating a profile failed")
		return
	}

	user, err := h.provider.UpdateProfile(c.Request.Context(), signedInSession(c), oidc.Profile{
		FirstName: req.FirstName,
		LastName:  req.LastName,
	}, client(c))
	if err != nil {
		h.fail(c, err, "updating a profile failed")
		return
	}

	c.JSON(http.StatusOK, gin.H{"user": newUserResponse(user)})
}

// ChangePassword replaces the signed-in user's password.
func (h *Handler) ChangePassword(c *gin.Context) {
	var req passwordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respond.BadRequest(c, "the request body is not valid")
		return
	}
	if err := req.validate(); err != nil {
		respond.Failure(c, h.log, err, "validating a password change failed")
		return
	}

	err := h.provider.ChangePassword(c.Request.Context(), signedInSession(c), req.CurrentPassword, req.NewPassword, client(c))
	if err != nil {
		h.fail(c, err, "changing a password failed")
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "changed"})
}

// Sessions lists where the signed-in user is signed in.
func (h *Handler) Sessions(c *gin.Context) {
	sessions, err := h.provider.Sessions(c.Request.Context(), signedInSession(c))
	if err != nil {
		h.fail(c, err, "listing sessions failed")
		return
	}

	c.JSON(http.StatusOK, gin.H{"sessions": sessions})
}

// EndSession signs one of the user's other browsers out.
func (h *Handler) EndSession(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respond.NotFound(c, "no such session")
		return
	}

	if err := h.provider.EndSession(c.Request.Context(), signedInSession(c), id, client(c)); err != nil {
		h.fail(c, err, "ending a session failed")
		return
	}

	c.Status(http.StatusNoContent)
}

// Applications lists the applications that can still act for the user.
func (h *Handler) Applications(c *gin.Context) {
	apps, err := h.provider.ConnectedApplications(c.Request.Context(), signedInSession(c))
	if err != nil {
		h.fail(c, err, "listing connected applications failed")
		return
	}

	c.JSON(http.StatusOK, gin.H{"applications": apps})
}

// Disconnect revokes an application's refresh tokens for the user.
func (h *Handler) Disconnect(c *gin.Context) {
	err := h.provider.DisconnectApplication(c.Request.Context(), signedInSession(c), c.Param("client_id"), client(c))
	if err != nil {
		h.fail(c, err, "disconnecting an application failed")
		return
	}

	c.Status(http.StatusNoContent)
}

// fail answers the provider's errors with the status each means, and anything
// else as the server's own failure.
func (h *Handler) fail(c *gin.Context, err error, note string) {
	var field *oidc.FieldError

	switch {
	case errors.As(err, &field):
		respond.BadRequest(c, field.Message)
	case errors.Is(err, oidc.ErrInvalidCredentials):
		respond.Error(c, http.StatusUnauthorized, err.Error())
	case errors.Is(err, oidc.ErrRequestExpired), errors.Is(err, oidc.ErrResetInvalid):
		respond.Error(c, http.StatusGone, err.Error())
	case errors.Is(err, oidc.ErrRegistrationClosed):
		respond.Error(c, http.StatusForbidden, err.Error())
	case errors.Is(err, oidc.ErrEmailTaken):
		respond.Conflict(c, err.Error())
	case errors.Is(err, oidc.ErrWrongPassword):
		respond.BadRequest(c, err.Error())
	case errors.Is(err, oidc.ErrNotYours):
		respond.NotFound(c, "no such session or application")
	default:
		respond.Failure(c, h.log, err, note)
	}
}

func client(c *gin.Context) oidc.Client {
	return oidc.Client{IP: c.ClientIP(), UserAgent: c.Request.UserAgent()}
}
