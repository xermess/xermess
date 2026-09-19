package session

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// UserCookie carries a user's session at this server: what lets the
// authorization endpoint sign someone in to a second application without
// asking for their password again. It is a different cookie from the
// administrators' on purpose — signing in to an application must never sign
// anyone in to the panel.
const UserCookie = "xermess_user_session"

// SetUser stores a user's session token in the browser, until the session
// expires — which the login flow that made it decides.
func SetUser(c *gin.Context, token string, expires time.Time, secure bool) {
	writeUser(c, token, int(time.Until(expires).Seconds()), secure)
}

// ClearUser removes it again.
func ClearUser(c *gin.Context, secure bool) {
	writeUser(c, "", -1, secure)
}

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
