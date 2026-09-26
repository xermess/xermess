package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"loginer/internal/api/respond"
)

// MaxBodyBytes is how much of a request body either server reads.
//
// Nothing here is an upload. The longest body anybody sends is a language's
// text or an identity provider's SAML metadata, tens of kilobytes at most, so
// a megabyte leaves room for those to grow several times over — and still
// keeps a caller from having a body read into memory until there is none
// left, which needs no account and no session to try.
const MaxBodyBytes = 1 << 20

// BodyTooLarge is what a caller sending more than that is told.
var BodyTooLarge = respond.Define(http.StatusRequestEntityTooLarge, "body_too_large", respond.Both)

// BodyLimit stops a request body from being read past `max` bytes.
//
// A body that says its length up front is refused before any of it is read.
// One that does not — a chunked body, or one whose Content-Length is a lie —
// is cut short by MaxBytesReader as the handler reads it, and the handler
// reports it as a body it could not read.
func BodyLimit(max int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.ContentLength > max {
			respond.Abort(c, BodyTooLarge)
			return
		}

		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, max)

		c.Next()
	}
}
