// Package respond writes every non-happy answer the same way:
//
//	{"error": "Wrong email or password.", "code": "invalid_credentials"}
//
// `code` names the problem for translation, `params` fills its sentence, and
// `error` is English for direct API callers. Server internals never appear.
package respond

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Fault is a problem with the request: a status, a translatable code and
// params, and the English sentence. Anything that is not a Fault is reported
// only as an internal error.
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

// Forbidden is for handlers that only know which application a request concerns
// after loading it.
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

// Failure passes a Fault through and logs anything else with `note`, answering
// with the internal problem, so database errors never reach a browser.
func Failure(c *gin.Context, log *slog.Logger, err error, note string) {
	var fault Fault
	if errors.As(err, &fault) {
		Write(c, fault)
		return
	}

	log.Error(note, "error", err)
	Write(c, Internal.Fault(nil))
}
