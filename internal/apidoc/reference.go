// Package apidoc writes the API reference from the code that serves the API.
//
// Nothing about an endpoint is written twice. The route table in
// internal/api/server.go already says which paths exist, on which server, and
// behind which guards; each handler's doc comment already says what it does;
// its request type already says which fields it reads and what it holds them
// to; the problems it can answer with are already defined next to it. So the
// reference is read out of those, with the type checker, rather than kept
// beside them — and TestTheReferenceIsCurrent fails when what is committed is
// behind the code, the way `bun run check` fails when the panel's tokens are
// behind its theme.
//
// It writes two things: an OpenAPI 3.1 document for each server, which the
// public server serves at /openapi.json for other applications' tooling, and
// Markdown pages the docs app (web/docs) renders beside the hand-written
// guides. `make docs` runs it.
package apidoc

// Reference is everything the two servers answer, as read from the code.
type Reference struct {
	Servers  []Server
	Problems []Problem
}

// Server is one of the two listeners, and the routes mounted on it.
type Server struct {
	// Key is what the files are named after: "public" or "admin".
	Key      string
	Title    string
	Sections []Section
}

// Operations is every operation on the server, in the order they are mounted.
func (s Server) Operations() []Operation {
	var out []Operation
	for _, section := range s.Sections {
		out = append(out, section.Operations...)
	}
	return out
}

// Section is the routes one handler package answers: a subject, in the
// route table's words.
type Section struct {
	// Slug is the package's name, which is what the page is called.
	Slug  string
	Title string
	// Intro is the line under the page's title.
	Intro      string
	Operations []Operation
}

// Operation is one method on one path.
type Operation struct {
	Method string
	// Path is the path as Gin writes it, with :parameters.
	Path string
	// Handler is package.Method, which is also where to read the code.
	Handler string
	// Summary is the first sentence of the handler's doc comment, and
	// Description the rest of it.
	Summary     string
	Description string
	Access      Access
	// RateLimit is which of the per-address limits the route is held to:
	// "" for none, "sign-in" or "token".
	RateLimit string
	// OriginChecked says a change to this route is only taken from the app's
	// own origin — the CSRF check every /api/v1 route passes.
	OriginChecked bool

	PathParams  []Param
	QueryParams []Param
	// FormParams are read from an application/x-www-form-urlencoded body.
	FormParams []Param
	Headers    []Param
	// BasicAuth says the handler reads HTTP Basic credentials: a client's
	// id and secret.
	BasicAuth bool

	// Body is the JSON the handler binds, when it binds one.
	Body      *Schema
	Responses []Response
	Problems  []Problem
}

// Access is who may call an operation.
type Access struct {
	// Session is whose signed-in session the request has to carry: "" for
	// nobody's, "user", "admin", or "admin-setup" for an administrator who may
	// still be part way through signing in.
	Session string
	// Token is which of this server's APIs an access token may call the
	// route as, instead of the session: "admin" for an admin-cli token,
	// "account" for a user's account API token, or "" for none.
	Token string
	// Permissions are the panel permissions the route checks; any one of
	// them is enough. Anywhere says holding one for a single application
	// is enough to reach the route, and the handler narrows from there.
	Permissions []string
	Anywhere    bool
	SuperAdmin  bool
}

// Param is a parameter read from the path, the query, a form or a header.
type Param struct {
	Name        string
	Description string
	Required    bool
	Schema      *Schema
}

// Response is one status an operation answers with on purpose.
type Response struct {
	Status int
	// ContentType is empty for an answer without a body, such as a redirect.
	ContentType string
	Schema      *Schema
}

// Problem is an error a server answers with: see respond.Problem.
type Problem struct {
	Status  int
	Code    string
	English string
	// Apps is "id", "console" or both: which app the problem is shown in,
	// and so which server can give it.
	Apps []string
}
