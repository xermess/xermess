// Package respond writes the answers that are not the happy path, so every
// endpoint fails the same way:
//
//	{"error": "Wrong email or password.", "code": "invalid_credentials"}
//
// `code` names the problem (problem.go) so an app can say it in the reader's
// language, `params` carries what the sentence needs, and `error` is the
// sentence in English for whoever calls the API directly. Nothing about the
// server's insides is ever in any of them.
package respond

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Fault is a problem with what was sent: the status to answer with, the code
// and parameters an app translates, and the English sentence. Anything that
// is not a Fault is the server's own fault, and the caller is told nothing
// more than that something went wrong.
//
// A Fault without a code is one not yet given one: an app shows its English
// as it stands. New ones are made from a Problem.
type Fault struct {
	Status  int
	Code    string
	Params  map[string]any
	Message string
}

func (f Fault) Error() string {
	return f.Message
}

// Fail answers with a problem, its parameters given as name, value pairs.
func Fail(c *gin.Context, problem Problem, pairs ...any) {
	Write(c, problem.With(pairs...))
}

// Abort is Fail for middleware: it answers and stops the chain.
func Abort(c *gin.Context, problem Problem, pairs ...any) {
	Write(c, problem.With(pairs...))
	c.Abort()
}

// Write answers with a fault as it stands.
func Write(c *gin.Context, fault Fault) {
	body := gin.H{"error": fault.Message}
	if fault.Code != "" {
		body["code"] = fault.Code
	}
	if len(fault.Params) > 0 {
		body["params"] = fault.Params
	}

	c.JSON(fault.Status, body)
}

// BadRequest says the request itself was wrong, in English only. Prefer a
// Problem: this is for the admin endpoints not yet given codes.
func BadRequest(c *gin.Context, message string) {
	Error(c, http.StatusBadRequest, message)
}

// NotFound says there is no such thing, in English only (see BadRequest).
func NotFound(c *gin.Context, message string) {
	Error(c, http.StatusNotFound, message)
}

// Forbidden is the Fault for an administrator whose roles do not allow what
// they asked, for handlers that can only tell once they know which
// application the request is about.
var Forbidden = NotAllowed.Fault(nil)

// Conflict says the request clashes with what is already stored, in English
// only (see BadRequest).
func Conflict(c *gin.Context, message string) {
	Error(c, http.StatusConflict, message)
}

// Error answers with a status and an English message, and no code.
func Error(c *gin.Context, status int, message string) {
	c.JSON(status, gin.H{"error": message})
}

// Failure answers whatever the error turns out to be: a Fault is passed on to
// the caller as it stands, and anything else is logged with `note` and
// answered with the internal problem.
//
// Handlers call this instead of picking a status themselves, which is what
// keeps a database error from ever reaching a browser.
func Failure(c *gin.Context, log *slog.Logger, err error, note string) {
	var fault Fault
	if errors.As(err, &fault) {
		Write(c, fault)
		return
	}

	log.Error(note, "error", err)
	Write(c, Internal.Fault(nil))
}
