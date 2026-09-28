// Package reference serves each server's OpenAPI document, so an
// application's developers can point their tooling at the server itself:
// import it into Postman or Insomnia, or generate a client from it.
//
// The documents are written by `make docs` (internal/apidoc) from the route
// table and the handlers, and embedded here; what the server adds at startup
// is its own address, which only it knows.
package reference

import (
	_ "embed"
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
)

var (
	//go:embed public.json
	publicDocument []byte
	//go:embed admin.json
	adminDocument []byte
)

// Handler answers one server's document.
type Handler struct {
	document []byte
}

// NewPublic is the public server's document, addressed at the issuer.
func NewPublic(issuer string) (*Handler, error) {
	return addressed(publicDocument, issuer)
}

// NewAdmin is the admin server's document, addressed at the panel's origin,
// which is where the admin API is reached from.
func NewAdmin(origin string) (*Handler, error) {
	return addressed(adminDocument, origin)
}

func addressed(document []byte, url string) (*Handler, error) {
	var parsed map[string]any
	if err := json.Unmarshal(document, &parsed); err != nil {
		return nil, err
	}
	if url != "" {
		parsed["servers"] = []map[string]string{{"url": url}}
	}

	out, err := json.MarshalIndent(parsed, "", "  ")
	if err != nil {
		return nil, err
	}
	return &Handler{document: out}, nil
}

// Document answers the OpenAPI document.
func (h *Handler) Document(c *gin.Context) {
	c.Header("Cache-Control", "public, max-age=300")
	c.Data(http.StatusOK, "application/json; charset=utf-8", h.document)
}
