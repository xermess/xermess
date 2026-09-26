package session

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"loginer/internal/brand"
)

// UserCookie carries a user's session at this server: what lets the
// authorization endpoint sign someone in to a second application without
// asking for their password again. It is a different cookie from the
// administrators' on purpose — signing in to an application must never sign
// anyone in to the panel.
const UserCookie = brand.UserSessionCookie

// SetUser stores a user's session token in the browser, until the session
// expires — which the login flow that made it decides.
// `remember` is the "stay signed in" box: without it the cookie carries no
// Max-Age and the browser drops it when its window closes, so a machine
// somebody was passing through forgets them. The session itself lasts as long
// as the login flow says either way — this only decides how long the browser
// holds on to it.
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

// maxAge is seconds, as http.Cookie counts them: above zero sets Max-Age,
// zero leaves it off — a cookie for this browser window — and below zero
// deletes the cookie.
func writeUser(c *gin.Context, value string, maxAge int, secure bool) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     UserCookie,
		Value:    value,
		Path:     "/",
		MaxAge:   maxAge,
		HttpOnly: true,
		Secure:   secure,
		// Lax, so the cookie is sent when an application sends the browser
		// to the authorization endpoint — a top-level navigation — and not
		// with a request another site's script makes.
		SameSite: http.SameSiteLaxMode,
	})
}

// SignInStateCookie ties a sign-in started at a provider — a social one, or an
// organisation's — to the browser that started it.
//
// The state travels to the provider and comes back in the address, so whoever
// holds one can present it in any browser: their own state and their own code,
// walked through somebody else's browser, would sign that browser in as them,
// and everything the person did next would go into the attacker's account.
// The cookie is the half that cannot be put in another person's browser, and
// the callback refuses a state that does not match it.
const SignInStateCookie = brand.SignInStateCookie

// signInStatePath keeps the cookie to the provider endpoints, which are the
// only ones that ever read it.
const signInStatePath = "/oauth2"

// SetSignInState remembers the state of a sign-in just started. It is a
// session cookie: what bounds the sign-in is the row in the database, and the
// cookie only has to last the trip to the provider and back.
func SetSignInState(c *gin.Context, state string, secure bool) {
	writeSignInState(c, state, 0, secure)
}

// ClearSignInState forgets it. The callback clears it however it ends, so a
// state cannot be presented twice even before the row expires.
func ClearSignInState(c *gin.Context, secure bool) {
	writeSignInState(c, "", -1, secure)
}

// SignInState is what the browser carries, or "" for a browser that started
// no sign-in.
func SignInState(c *gin.Context) string {
	value, _ := c.Cookie(SignInStateCookie)
	return value
}

func writeSignInState(c *gin.Context, value string, maxAge int, secure bool) {
	// Most providers answer with a redirect the browser follows, but some
	// answer by posting a form — Apple, and every SAML one. Lax carries the
	// cookie on the first and not on the second, so where it can be Secure it
	// is None, which is the only way a browser sends a cookie with a
	// cross-site POST and which browsers only accept with Secure. Served over
	// plain http, which is development, it stays Lax: the redirect providers
	// work there, and the ones that post do not.
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
