package respond

import (
	"fmt"
	"net/http"
	"sort"
	"sync"

	"loginer/i18n"
)

// Problem is one thing that can go wrong, as an app is told about it: the
// status, a code, and which apps may be shown it.
//
// The code is what makes the answer translatable. Every error the API gives
// carries one —
//
//	{"error": "Wrong email or password.", "code": "invalid_credentials"}
//
// — and the app shows `error.<code>` from its own catalog, in the reader's
// language, with `params` filled in. The English sentence is there for
// whoever calls the API without an app, and it is not written here: it is the
// base language's text for the same key, so it exists once, in
// i18n/<app>/en/, and a translator and the server can never disagree
// about what an error says.
//
// A problem is defined next to the code that returns it, with Define, and
// TestErrorCodesMatchTheCatalogs holds the definitions and the catalogs to
// each other: every problem has its key in the catalog of every app it is
// for, and every `error.*` key in a catalog is a problem the server can give.
type Problem struct {
	Status int
	Code   string
	// Apps are the apps whose pages can be shown this problem, and so whose
	// catalogs have to say it: the sign-in pages for the public API, the
	// panel for the admin one, both for what either server can answer.
	Apps []i18n.App
}

// Who a problem is for.
var (
	Public = []i18n.App{i18n.ID}
	Admin  = []i18n.App{i18n.Console}
	Both   = []i18n.App{i18n.ID, i18n.Console}
)

var (
	registryMu sync.Mutex
	registry   = map[string]Problem{}
)

// Define adds a problem the server can give. It is called once, from a
// package-level var next to the handler that returns it; defining a code
// twice is a mistake in our own code, and stops the server at startup.
func Define(status int, code string, apps []i18n.App) Problem {
	registryMu.Lock()
	defer registryMu.Unlock()

	if _, taken := registry[code]; taken {
		panic(fmt.Sprintf("respond: the problem %q is defined twice", code))
	}

	problem := Problem{Status: status, Code: code, Apps: apps}
	registry[code] = problem

	return problem
}

// Lookup finds a problem by its code.
func Lookup(code string) (Problem, bool) {
	registryMu.Lock()
	defer registryMu.Unlock()

	problem, ok := registry[code]

	return problem, ok
}

// Problems is every problem defined so far, by code.
func Problems() []Problem {
	registryMu.Lock()
	defer registryMu.Unlock()

	out := make([]Problem, 0, len(registry))
	for _, problem := range registry {
		out = append(out, problem)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Code < out[j].Code })

	return out
}

// Key is the catalog key an app shows for a problem.
func (p Problem) Key() string {
	return "error." + p.Code
}

// English is the problem's sentence in the base language, parameters filled
// in. A problem whose key is missing from the catalog says its code, which
// the tests stop from ever shipping.
func (p Problem) English(params map[string]any) string {
	for _, app := range p.Apps {
		if text, ok := i18n.Text(app, p.Key()); ok {
			return i18n.Fill(text, params)
		}
	}

	return p.Code
}

// With is the problem as an error a handler can return, with its parameters
// given as name, value pairs:
//
//	return passwordTooShort.With("min", 8)
func (p Problem) With(pairs ...any) Fault {
	params := map[string]any{}
	for i := 0; i+1 < len(pairs); i += 2 {
		params[fmt.Sprint(pairs[i])] = pairs[i+1]
	}

	return p.Fault(params)
}

// Fault is the problem as an error, with its parameters.
func (p Problem) Fault(params map[string]any) Fault {
	if len(params) == 0 {
		params = nil
	}

	return Fault{Status: p.Status, Code: p.Code, Params: params, Message: p.English(params)}
}

// The problems either server can give, whatever the endpoint.
var (
	InvalidBody = Define(http.StatusBadRequest, "invalid_body", Both)
	NotFoundAny = Define(http.StatusNotFound, "not_found", Both)
	Internal    = Define(http.StatusInternalServerError, "internal", Both)
	NotSignedIn = Define(http.StatusUnauthorized, "not_signed_in", Both)
	NotAllowed  = Define(http.StatusForbidden, "forbidden", Both)

	// LanguageNotFound is a language code nothing has: the sign-in pages
	// asking for one that was removed, or the panel for one it never had.
	LanguageNotFound = Define(http.StatusNotFound, "language_not_found", Both)
)
