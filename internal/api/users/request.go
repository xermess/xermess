package users

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"loginer/internal/api/query"
	"loginer/internal/store"
)

// targetType is what these records are called in the activity log.
const targetType = "user"

// maxPageSize caps how many users one request can ask for.
const maxPageSize = 200

// maxApplications caps how many applications a user's roles are reported for.
const maxApplications = 500

// assignRequest is the body of the endpoint that gives a user roles: the ids
// of the roles to add, global and application roles alike.
type assignRequest struct {
	Roles []uuid.UUID `json:"roles"`
}

// userRequest is the body of the create and update endpoints.
//
// The named fields are the built-in ones: they are columns of the record, so
// they are spelled out here and held to the rules in the tags. Data carries
// the additional fields an organisation added, whose rules live in
// user_fields and are applied in validation.go.
type userRequest struct {
	Email string `json:"email" validate:"required,email,max=255"`
	// EmailVerified and IsActive may be left out: a new user is then
	// unverified and active, and an existing one keeps what they had.
	EmailVerified *bool `json:"email_verified"`

	FirstName string `json:"first_name" validate:"max=100"`
	LastName  string `json:"last_name" validate:"max=100"`

	IsActive *bool `json:"is_active"`

	// Password is required when creating a user. When updating one it is
	// optional: left empty, the password the user has is kept.
	Password        string `json:"password" validate:"max=72"`
	ConfirmPassword string `json:"confirm_password" validate:"eqfield=Password"`
	// IsTemporaryPassword marks the password as one the user has to replace.
	// Left out, a new user's password is not temporary and an existing
	// user's keeps what it was.
	IsTemporaryPassword *bool `json:"is_temporary_password"`

	Data map[string]any `json:"data"`
}

// clean tidies what can be tidied, so the rules below see the values that
// would actually be stored.
func (r *userRequest) clean() {
	r.Email = trimmedLower(r.Email)
	r.FirstName = strings.TrimSpace(r.FirstName)
	r.LastName = strings.TrimSpace(r.LastName)
}

// listQuery reads the search box, the filter and the page out of the query
// string, holding each to something sensible.
func listQuery(c *gin.Context) store.UserQuery {
	q := store.UserQuery{
		Search:   c.Query("search"),
		Limit:    query.Int(c, "limit", 50, maxPageSize),
		Offset:   query.Int(c, "offset", 0, query.MaxOffset),
		Verified: query.Bool(c, "verified"),
	}

	if role, err := uuid.Parse(c.Query("role")); err == nil {
		q.Role = &role
	}

	return q
}

// trimmedLower is how an email is stored: no stray spaces, one case, so two
// spellings of the same address cannot both be registered.
func trimmedLower(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}
