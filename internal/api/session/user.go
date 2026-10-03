package session

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"loginer/internal/brand"
)

// UserCookie carries a user's own session for single sign-on. It is
// deliberately not the administrators' cookie: signing in to an application
// must never sign anyone in to the panel.
const UserCookie = brand.UserSessionCookie

// SetUser stores the session cookie. Without `remember` ("stay signed in") it
// has no Max-Age and disappears with the browser window; the session's own
// lifetime is the login flow's either way.
func SetUser(c *gin.Context, token string, expires time.Time, remember, secure bool) {
	maxAge := 0
	if remember {
		maxAge = int(time.Until(expires).Seconds())
	}

	writeUser(c, token, maxAge, secure)
}

// ClearUser removes it again.
func ClearUser(c *gin.Context, secure bool) {
	writeUser(c, "", -1, secure)
}

// maxAge is seconds: positive sets Max-Age, zero makes a browser-session
// cookie, negative deletes it.
func writeUser(c *gin.Context, value string, maxAge int, secure bool) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     UserCookie,
		Value:    value,
		Path:     "/",
		MaxAge:   maxAge,
		HttpOnly: true,
		Secure:   secure,
		// Lax: sent on the top-level navigation to the authorization endpoint,
		// not on other sites' script requests.
		SameSite: http.SameSiteLaxMode,
	})
}

// SignInStateCookie binds a provider sign-in to the browser that started it.
// Without it, an attacker could walk a victim's browser through the attacker's
// own state and code and sign the victim into the attacker's account.
const SignInStateCookie = brand.SignInStateCookie

// signInStatePath keeps the cookie to the provider endpoints, which are the
// only ones that ever read it.
const signInStatePath = "/oauth2"

// SetSignInState stores the state of a sign-in just started, as a
// browser-session cookie; the database row bounds its lifetime.
func SetSignInState(c *gin.Context, state string, secure bool) {
	writeSignInState(c, state, 0, secure)
}

// ClearSignInState forgets it. The callback clears it however it ends, so a
// state cannot be presented twice even before the row expires.
func ClearSignInState(c *gin.Context, secure bool) {
	writeSignInState(c, "", -1, secure)
}

// SignInState is the browser's state cookie, or "". Two cookies of the same
// name (which a sibling subdomain can plant) count as none, so the sign-in
// fails closed.
func SignInState(c *gin.Context) string {
	return single(c, SignInStateCookie)
}

// UserToken is the user's session cookie, or "" when absent or duplicated (a
// planted sibling-subdomain cookie cannot stand in).
func UserToken(c *gin.Context) string {
	return single(c, UserCookie)
}

// single reads a cookie only when exactly one of that name was sent.
func single(c *gin.Context, name string) string {
	named := c.Request.CookiesNamed(name)
	if len(named) != 1 {
		return ""
	}

	return named[0].Value
}

func writeSignInState(c *gin.Context, value string, maxAge int, secure bool) {
	// Apple and SAML providers post their answer cross-site, which only
	// SameSite=None cookies survive, and None requires Secure. Over plain http
	// (development) it stays Lax.
	sameSite := http.SameSiteLaxMode
	if secure {
		sameSite = http.SameSiteNoneMode
	}

	http.SetCookie(c.Writer, &http.Cookie{
		Name:     SignInStateCookie,
		Value:    value,
		Path:     signInStatePath,
		MaxAge:   maxAge,
		HttpOnly: true,
		Secure:   secure,
		SameSite: sameSite,
	})
}
