// Package middleware holds what every request passes through, whichever route
// it is on, and the order they pass through them in.
//
// Only the pieces that are the same for every deployment live here: the log
// line a request leaves behind, and turning a panic into an answer. Anything
// that depends on configuration — which browser origins may call the API, who
// is signed in — is built elsewhere and passed in.
package middleware

import (
	"log/slog"

	"github.com/gin-gonic/gin"
)

// Chain is everything a request passes through, in the order it passes
// through them: logged first, so a request that panics is still reported;
// recovered next, so a panic becomes a 500 rather than a dead connection;
// then the cross-origin rules, which the caller builds from its configuration
// and hands in; and last the limit on how much body will be read, which comes
// after the preflight answer so a browser's check is never refused for a body
// it did not send.
//
// Spread it into gin's Use:
//
//	r.Use(middleware.Chain(log, cors.New(origins))...)
func Chain(log *slog.Logger, cors gin.HandlerFunc) []gin.HandlerFunc {
	return []gin.HandlerFunc{
		RequestLogger(log),
		gin.Recovery(),
		cors,
		BodyLimit(MaxBodyBytes),
	}
}
