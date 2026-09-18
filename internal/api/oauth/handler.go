// Package oauth answers the OAuth 2.0 and OpenID Connect provider endpoints:
// discovery and keys, authorization, token, userinfo, logout, revocation and
// introspection.
//
// These are not JSON API endpoints like the rest. They follow their RFCs: the
// token endpoint takes a form and answers errors as {"error", "error_description"},
// the authorization and logout endpoints answer with redirects, and userinfo
// reports a bad token in WWW-Authenticate. What each one decides is in
// internal/oidc; this package only reads and writes HTTP.
package oauth

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"xermess/internal/api/session"
	"xermess/internal/oidc"
)

// Handler holds what these endpoints need.
type Handler struct {
	provider *oidc.Service
	log      *slog.Logger
	secure   bool
}

// New returns a Handler. `secure` sets the Secure flag on the user session
// cookie the logout endpoint clears.
func New(provider *oidc.Service, log *slog.Logger, secure bool) *Handler {
	return &Handler{provider: provider, log: log, secure: secure}
}

// Discovery answers the OpenID Connect discovery document.
func (h *Handler) Discovery(c *gin.Context) {
	c.Header("Cache-Control", "public, max-age=300")
	c.JSON(http.StatusOK, h.provider.Discovery(c.Request.Context()))
}

// JWKS answers the public keys tokens are signed with.
func (h *Handler) JWKS(c *gin.Context) {
	keys, err := h.provider.JWKS(c.Request.Context())
	if err != nil {
		h.serverError(c, err, "loading the signing keys failed")
		return
	}

	// Short enough that a new key reaches every API within minutes.
	c.Header("Cache-Control", "public, max-age=300")
	c.JSON(http.StatusOK, keys)
}

// Authorize starts a sign-in for an application, and redirects: to the
// sign-in page, or straight back with a code for a user already signed in.
func (h *Handler) Authorize(c *gin.Context) {
	ctx := c.Request.Context()

	params, err := authorizeParams(c)
	if err != nil {
		c.String(http.StatusBadRequest, "the request could not be read")
		return
	}

	current, err := h.provider.SessionFor(ctx, cookie(c))
	if err != nil {
		h.serverError(c, err, "loading the user session failed")
		return
	}

	location, err := h.provider.Authorize(ctx, params, current)
	if err != nil {
		h.serverError(c, err, "authorization failed")
		return
	}

	c.Header("Cache-Control", "no-store")
	c.Redirect(http.StatusFound, location)
}

// SocialStart sends the browser to a provider to sign in there.
func (h *Handler) SocialStart(c *gin.Context) {
	location, err := h.provider.StartSocial(
		c.Request.Context(),
		c.Param("slug"),
		c.Query("request"),
		c.Query("next"),
	)
	if err != nil {
		h.socialFailed(c, err, "starting a social sign-in failed")
		return
	}

	c.Header("Cache-Control", "no-store")
	c.Redirect(http.StatusFound, location)
}

// SocialCallback is where the provider sends the browser back to.
//
// It answers GET and POST: nearly every provider redirects with the code in
// the query, and Apple posts it as a form when a name or an address was asked
// for.
func (h *Handler) SocialCallback(c *gin.Context) {
	ctx := c.Request.Context()

	code, state := c.Query("code"), c.Query("state")
	if c.Request.Method == http.MethodPost {
		code, state = c.PostForm("code"), c.PostForm("state")
	}

	// The provider refused, or the person changed their mind there. What it
	// said is for the log; the page says the sign-in did not happen.
	if refused := c.Query("error"); refused != "" {
		h.log.Info("a social provider refused a sign-in",
			"provider", c.Param("slug"), "error", refused,
			"description", c.Query("error_description"))

		c.Redirect(http.StatusFound, h.provider.SocialErrorPage(oidc.ErrSocialUpstream))
		return
	}

	result, err := h.provider.CompleteSocial(ctx, c.Param("slug"), code, state, client(c))
	if err != nil {
		h.socialFailed(c, err, "completing a social sign-in failed")
		return
	}

	session.SetUser(c, result.SignIn.Token, h.secure)

	c.Header("Cache-Control", "no-store")
	c.Redirect(http.StatusFound, h.provider.SocialLanding(ctx, result))
}

// socialFailed sends the browser to the sign-in app's error page. What the
// person is told is the provider service's to decide; anything that is not
// one of its own errors is the server's fault and is logged.
func (h *Handler) socialFailed(c *gin.Context, err error, note string) {
	if !oidc.IsSocialFailure(err) {
		h.log.Error(note, "provider", c.Param("slug"), "error", err)
	}

	c.Header("Cache-Control", "no-store")
	c.Redirect(http.StatusFound, h.provider.SocialErrorPage(err))
}

// Token issues tokens.
func (h *Handler) Token(c *gin.Context) {
	noStore(c)

	params, err := tokenParams(c)
	if err != nil {
		h.oauthError(c, err)
		return
	}

	tokens, err := h.provider.Token(c.Request.Context(), params)
	if err != nil {
		h.oauthError(c, err)
		return
	}

	c.JSON(http.StatusOK, tokens)
}

// UserInfo answers the claims about the user an access token is for.
func (h *Handler) UserInfo(c *gin.Context) {
	noStore(c)

	token, err := bearer(c)
	if err != nil {
		h.bearerError(c, err)
		return
	}

	claims, err := h.provider.UserInfo(c.Request.Context(), token)
	if err != nil {
		h.bearerError(c, err)
		return
	}

	c.JSON(http.StatusOK, claims)
}

// Logout signs the user out of this server, and redirects.
func (h *Handler) Logout(c *gin.Context) {
	params, err := logoutParams(c)
	if err != nil {
		c.String(http.StatusBadRequest, "the request could not be read")
		return
	}

	location, err := h.provider.Logout(c.Request.Context(), params, cookie(c), client(c))
	if err != nil {
		h.serverError(c, err, "logout failed")
		return
	}

	session.ClearUser(c, h.secure)
	c.Header("Cache-Control", "no-store")
	c.Redirect(http.StatusFound, location)
}

// Revoke revokes a refresh token.
func (h *Handler) Revoke(c *gin.Context) {
	noStore(c)

	auth, token, err := tokenOnlyParams(c)
	if err != nil {
		h.oauthError(c, err)
		return
	}

	if err := h.provider.Revoke(c.Request.Context(), auth, token); err != nil {
		h.oauthError(c, err)
		return
	}

	c.Status(http.StatusOK)
}

// Introspect says whether a token is active.
func (h *Handler) Introspect(c *gin.Context) {
	noStore(c)

	auth, token, err := tokenOnlyParams(c)
	if err != nil {
		h.oauthError(c, err)
		return
	}

	result, err := h.provider.Introspect(c.Request.Context(), auth, token)
	if err != nil {
		h.oauthError(c, err)
		return
	}

	c.JSON(http.StatusOK, result)
}

// oauthError answers an error in the token endpoint's format. Anything that
// is not an OAuth error is the server's, and says no more than server_error.
func (h *Handler) oauthError(c *gin.Context, err error) {
	var failure *oidc.Error
	if !errors.As(err, &failure) {
		h.log.Error("an oauth endpoint failed", "error", err, "path", c.FullPath())
		c.JSON(http.StatusInternalServerError, errorResponse{Error: oidc.ErrServerError, Description: "something went wrong"})
		return
	}

	// A client that tried HTTP Basic is told which scheme to use again
	// (RFC 6749 section 5.2).
	if failure.Status == http.StatusUnauthorized {
		c.Header("WWW-Authenticate", `Basic realm="xermess"`)
	}

	c.JSON(failure.Status, errorResponse{Error: failure.Code, Description: failure.Description})
}

// bearerError answers a userinfo error the way RFC 6750 says: in the
// WWW-Authenticate header, with the same in the body for people reading it.
func (h *Handler) bearerError(c *gin.Context, err error) {
	var failure *oidc.Error
	if !errors.As(err, &failure) {
		h.log.Error("userinfo failed", "error", err)
		c.JSON(http.StatusInternalServerError, errorResponse{Error: oidc.ErrServerError, Description: "something went wrong"})
		return
	}

	c.Header("WWW-Authenticate", `Bearer realm="xermess", error="`+failure.Code+`", error_description="`+quoteSafe(failure.Description)+`"`)
	c.JSON(failure.Status, errorResponse{Error: failure.Code, Description: failure.Description})
}

// serverError is for the endpoints that answer with a page rather than JSON.
func (h *Handler) serverError(c *gin.Context, err error, note string) {
	h.log.Error(note, "error", err)
	c.String(http.StatusInternalServerError, "something went wrong")
}

func noStore(c *gin.Context) {
	c.Header("Cache-Control", "no-store")
	c.Header("Pragma", "no-cache")
}

func cookie(c *gin.Context) string {
	value, _ := c.Cookie(session.UserCookie)
	return value
}

func client(c *gin.Context) oidc.Client {
	return oidc.Client{IP: c.ClientIP(), UserAgent: c.Request.UserAgent()}
}
