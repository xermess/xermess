package admins

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"loginer/internal/api/query"
	"loginer/internal/model"
	"loginer/internal/store"
)

// targetType is what these records are called in the activity log.
const targetType = "admin_user"

// maxPageSize caps how many administrators one request can ask for.
const maxPageSize = 200

// adminRequest is the body of the create and update endpoints.
//
// There is no username: the address is the account, as it is for the first
// administrator. Assignments replace every role the administrator held.
type adminRequest struct {
	Email     string `json:"email" validate:"required,email,max=255"`
	FirstName string `json:"first_name" validate:"required,max=100"`
	LastName  string `json:"last_name" validate:"max=100"`

	// Status says whether the account may sign in. Only an active one may.
	Status string `json:"status" validate:"required,oneof=active suspended disabled"`

	// Password is required when creating an administrator. When updating one
	// it is optional: left empty, the password they have is kept.
	Password        string `json:"password" validate:"max=72"`
	ConfirmPassword string `json:"confirm_password" validate:"eqfield=Password"`

	Assignments []assignmentRequest `json:"assignments"`
}

// assignmentRequest gives the administrator one role, for the whole panel
// when ApplicationID is null or for that one application.
type assignmentRequest struct {
	RoleID        uuid.UUID  `json:"role_id"`
	ApplicationID *uuid.UUID `json:"application_id"`
}

// clean tidies what can be tidied. The password is taken as it was typed.
func (r *adminRequest) clean() {
	r.Email = strings.ToLower(strings.TrimSpace(r.Email))
	r.FirstName = strings.TrimSpace(r.FirstName)
	r.LastName = strings.TrimSpace(r.LastName)
	r.Status = strings.TrimSpace(r.Status)
}

// listQuery reads the search box, the filters and the page out of the query
// string, holding each to something sensible.
func listQuery(c *gin.Context) store.AdminQuery {
	q := store.AdminQuery{
		Search: c.Query("search"),
		Limit:  query.Int(c, "limit", 50, maxPageSize),
		Offset: query.Int(c, "offset", 0, query.MaxOffset),
	}

	if status := model.Status(c.Query("status")); status.Valid() {
		q.Status = status
	}

	if role, err := uuid.Parse(c.Query("role")); err == nil {
		q.Role = &role
	}

	return q
}

// securityRequest is what a change to the settings that apply to every
// administrator sends. The flag is a pointer so a request that does not
// mention it leaves it alone.
type securityRequest struct {
	MFARequired *bool `json:"mfa_required"`
}
