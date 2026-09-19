// Package csrf stops another site from changing anything with a signed-in
// user's or administrator's cookie.
//
// The session cookies are SameSite=Lax, which keeps other sites' requests from
// carrying them — but "site" means the registrable domain, so any subdomain of
// it counts as the same site: a blog on blog.mywebsite.com, a user's page on
// a shared host. And the JSON endpoints bind any body as JSON, so a plain
// HTML form posting text/plain would reach them without a CORS preflight.
//
// Two checks close that, on every request that can change something:
//
//   - a body has to be application/json, which a form cannot send;
//   - a browser has to say it came from an origin that is allowed: the Origin
//     header, or failing that Sec-Fetch-Site. A request with neither is not a
//     browser's — a script, curl, another server — and carries no victim's
//     cookie, so it is let through to authenticate like any other.
package csrf

import (
	"mime"
	"net/http"
	"slices"

	"github.com/gin-gonic/gin"

	"xermess/internal/api/respond"
)

// What a refused request is told.
var (
	unsupportedBody = respond.Define(http.StatusUnsupportedMediaType, "unsupported_body", respond.Both)
	crossOrigin     = respond.Define(http.StatusForbidden, "cross_origin", respond.Both)
	crossSite       = respond.Define(http.StatusForbidden, "cross_site", respond.Both)
)

// New guards the routes it is mounted on. `origins` are the exact origins
// allowed to make changes: the app served on them, and any CORS origin
// configured.
func New(origins []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		switch c.Request.Method {
		case http.MethodGet, http.MethodHead, http.MethodOptions:
			c.Next()
			return
		}

		if hasBody(c.Request) {
			media, _, _ := mime.ParseMediaType(c.GetHeader("Content-Type"))
			if media != "application/json" {
				respond.Abort(c, unsupportedBody)
				return
			}
		}

		if origin := c.GetHeader("Origin"); origin != "" {
			if !slices.Contains(origins, origin) {
				respond.Abort(c, crossOrigin, "origin", origin)
				return
			}
		} else if site := c.GetHeader("Sec-Fetch-Site"); site != "" && site != "same-origin" && site != "none" {
			respond.Abort(c, crossSite)
			return
		}

		c.Next()
	}
}

func hasBody(r *http.Request) bool {
	return r.ContentLength > 0 || len(r.TransferEncoding) > 0
}
