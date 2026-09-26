package roles

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"loginer/internal/api/query"
	"loginer/internal/store"
)

// targetType is what these records are called in the activity log.
const targetType = "user_role"

// maxPageSize caps how many roles one request can ask for. It is higher than
// for users: the panel asks for every role it can see at once to offer them
// as choices.
const maxPageSize = 500

// roleRequest is the body of the create and update endpoints.
//
// Inherits holds ids, and replaces what the role had: an update sends the
// whole role, the same as for a user.
type roleRequest struct {
	// ApplicationID is the application the role belongs to, or null for a
	// global role. It is read on create only: a role keeps its scope.
	ApplicationID *uuid.UUID `json:"application_id"`

	Name        string `json:"name" validate:"required,rolename"`
	Description string `json:"description" validate:"max=255"`
	// IsDefault gives the role to every user created without roles of their
	// own.
	IsDefault bool `json:"is_default"`

	Inherits []uuid.UUID `json:"inherits"`

	// APIScopes are the ids of the API scopes the role grants, and replace
	// what it granted. Left out, the role keeps its grants — so an
	// administrator who cannot see APIs can still edit the rest of it.
	APIScopes *[]uuid.UUID `json:"api_scopes"`
}

// clean tidies what can be tidied, so the rules see the values that would
// actually be stored.
func (r *roleRequest) clean() {
	r.Name = strings.ToLower(strings.TrimSpace(r.Name))
	r.Description = strings.TrimSpace(r.Description)
}

// listQuery reads the search box, the filter and the page out of the query
// string, holding each to something sensible.
func listQuery(c *gin.Context) store.RoleQuery {
	return store.RoleQuery{
		Search:  c.Query("search"),
		Limit:   query.Int(c, "limit", 50, maxPageSize),
		Offset:  query.Int(c, "offset", 0, query.MaxOffset),
		Default: query.Bool(c, "default"),
	}
}
