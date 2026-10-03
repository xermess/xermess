// Package apidoc writes the API reference from the code that serves the API:
// the route table, each handler's doc comment, request types and the problems
// defined beside them. TestTheReferenceIsCurrent fails when the committed
// reference is behind.
//
// It writes an OpenAPI 3.1 document per server and Markdown pages for web/docs.
// `make docs` runs it.
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
	// Session is the session the route requires: "", "user", "admin", or
	// "admin-setup" (may be mid sign-in).
	Session string
	// Token is the API an access token may call the route as instead: "admin",
	// "account" or "".
	Token string
	// Permissions the route checks (any one suffices). Anywhere means holding
	// one for a single application is enough.
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
