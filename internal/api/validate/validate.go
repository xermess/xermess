// Package validate checks that a request says what it has to say, before a
// handler acts on it.
//
// The rules are struct tags, read by go-playground/validator:
//
//	type loginRequest struct {
//		Username string `json:"username" validate:"required"`
//	}
//
// What this package adds is the answer. A failed rule comes back as a
// respond.Fault carrying a 400 and the problem `validation.<rule>`, with the
// field named the way the request named it and whatever the rule was held
// to —
//
//	{"code": "validation.max", "params": {"field": "first_name", "max": "100"}}
//
// — which an app says in the reader's language ("{field} must be at most
// {max} characters"), and the English beside it says for everyone else. Not
// "Key: 'Email' Error:Field validation for 'Email' failed on the 'email' tag".
package validate

import (
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"unicode"

	"github.com/go-playground/validator/v10"

	"xermess/internal/api/respond"
)

// instance is shared: building one is expensive, and it is safe to use from
// every request at once.
var instance = newValidator()

// rules are the problem each rule answers with. The built-in rules either
// server uses are listed here; a package adds its own with Register.
var rules = map[string]respond.Problem{
	"required":   respond.Define(http.StatusBadRequest, "validation.required", respond.Both),
	"email":      respond.Define(http.StatusBadRequest, "validation.email", respond.Both),
	"max":        respond.Define(http.StatusBadRequest, "validation.max", respond.Both),
	"min":        respond.Define(http.StatusBadRequest, "validation.min", respond.Both),
	"oneof":      respond.Define(http.StatusBadRequest, "validation.oneof", respond.Both),
	"startswith": respond.Define(http.StatusBadRequest, "validation.startswith", respond.Both),
	"eqfield":    respond.Define(http.StatusBadRequest, "validation.eqfield", respond.Both),
}

// invalid is the answer for a rule with no problem of its own.
var invalid = respond.Define(http.StatusBadRequest, "validation.invalid", respond.Both)

func newValidator() *validator.Validate {
	v := validator.New(validator.WithRequiredStructEnabled())

	// Name a field in an error the way the request named it, so the panel can
	// show the message next to the box someone typed in.
	v.RegisterTagNameFunc(func(field reflect.StructField) string {
		name := strings.Split(field.Tag.Get("json"), ",")[0]
		if name == "" || name == "-" {
			return field.Name
		}

		return name
	})

	return v
}

// Register adds a rule of our own, for something the library cannot know
// about — what a field name may look like, which types exist. Breaking it is
// the problem `validation.<tag>`, which the panel's catalog has to say: the
// rules registered this way are the admin API's.
//
// Call it from the init of the package whose rule it is, so the rule and the
// thing it describes stay together.
func Register(tag string, valid func(value string) bool) {
	if err := instance.RegisterValidation(tag, func(fl validator.FieldLevel) bool {
		return valid(fl.Field().String())
	}); err != nil {
		// Only a bad tag name gets here, which is a mistake in our own code
		// rather than anything a request can cause.
		panic(fmt.Sprintf("validate: register %q: %v", tag, err))
	}

	rules[tag] = respond.Define(http.StatusBadRequest, "validation."+tag, respond.Admin)
}

// Struct checks a request against its tags. It returns a respond.Fault, so a
// handler passes whatever comes back to respond.Failure and is done.
func Struct(v any) error {
	err := instance.Struct(v)
	if err == nil {
		return nil
	}

	var broken validator.ValidationErrors
	if !errors.As(err, &broken) {
		// The value was not a struct at all: our mistake, not the caller's.
		return err
	}

	return Fault(broken[0])
}

// Fault is the answer for one broken rule. Only the first is reported: a page
// shows one line, and a form is easier to fix one thing at a time.
func Fault(broken validator.FieldError) respond.Fault {
	problem, ok := rules[broken.Tag()]
	if !ok {
		problem = invalid
	}

	params := map[string]any{"field": broken.Field()}

	// A rule held to something says what: the length it wanted, the values
	// it takes, the field it has to match.
	if param := broken.Param(); param != "" {
		switch broken.Tag() {
		case "max", "min":
			params[broken.Tag()] = param
		case "oneof":
			params["values"] = strings.ReplaceAll(param, " ", ", ")
		case "startswith":
			params["prefix"] = param
		case "eqfield":
			params["other"] = toSnake(param)
		}
	}

	return problem.Fault(params)
}

// toSnake spells a Go field name the way the request does — ConfirmPassword
// as confirm_password — for a rule whose parameter names another field.
func toSnake(name string) string {
	var b strings.Builder

	for i, r := range name {
		if unicode.IsUpper(r) {
			if i > 0 {
				b.WriteByte('_')
			}
			r = unicode.ToLower(r)
		}
		b.WriteRune(r)
	}

	return b.String()
}

// The four below read one field of a PATCH: the value sent, or `current` when
// none was — the record's own value on an update, and the default on a
// create. A plain value cannot tell "false", "" or 0 from "not sent", so a
// request carries pointers and these turn them back into values.
//
// They live here rather than beside each handler because every subject that
// takes a PATCH needs the same four, and six copies of "if sent == nil" is
// six places for one of them to start behaving differently.

// Flag reads an on-or-off field. An API client that omits a flag should not
// switch it off.
func Flag(sent *bool, current bool) bool {
	if sent == nil {
		return current
	}

	return *sent
}

// Text reads a field of words, trimmed. A client that does not know about a
// field should not clear it by not mentioning it.
func Text(sent *string, current string) string {
	if sent == nil {
		return current
	}

	return strings.TrimSpace(*sent)
}

// Lower is Text for the fields that are compared, dialled or put in an
// address rather than read: a host name, an email address, a short name.
func Lower(sent *string, current string) string {
	return strings.ToLower(Text(sent, current))
}

// Number reads a field that counts: a port, a length, a lifetime.
func Number(sent *int, current int) int {
	if sent == nil {
		return current
	}

	return *sent
}
