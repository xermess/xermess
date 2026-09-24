package auth

import (
	"time"

	"github.com/google/uuid"

	"xermess/internal/model"
)

// adminResponse is an administrator as the browser sees it. It is built by
// hand rather than returning the model, so a column added later cannot
// accidentally start being published.
type adminResponse struct {
	ID        string     `json:"id"`
	Username  string     `json:"username"`
	Email     string     `json:"email"`
	FullName  string     `json:"full_name"`
	FirstName string     `json:"first_name"`
	LastName  string     `json:"last_name"`
	Status    string     `json:"status"`
	LastLogin *time.Time `json:"last_login_at,omitempty"`

	// Roles are the names of the roles held for the whole panel.
	Roles []string `json:"roles"`

	// Permissions are what those roles allow, ScopedPermissions what roles
	// held for one application add there, and IsSuperAdmin whether they may
	// manage administrators. The panel shows and hides its controls by
	// these; the server checks them regardless.
	Permissions       []string               `json:"permissions"`
	ScopedPermissions map[uuid.UUID][]string `json:"scoped_permissions"`
	IsSuperAdmin      bool                   `json:"is_super_admin"`

	// MFAEnabled says whether they sign in with a second factor.
	MFAEnabled bool `json:"mfa_enabled"`
}

func newAdminResponse(a *model.AdminUser) adminResponse {
	roles := []string{}
	for _, assignment := range a.Assignments {
		if assignment.Global() {
			roles = append(roles, assignment.Role.Name)
		}
	}

	return adminResponse{
		ID:        a.ID.String(),
		Username:  a.Username,
		Email:     a.Email,
		FullName:  a.FullName(),
		FirstName: a.FirstName,
		LastName:  a.LastName,
		Status:    string(a.Status),
		Roles:     roles,
		LastLogin: a.LastLoginAt,

		Permissions:       a.Permissions(),
		ScopedPermissions: a.ScopedPermissions(),
		IsSuperAdmin:      a.IsSuperAdmin(),
		MFAEnabled:        a.HasMFA(),
	}
}

// sessionResponse is one of the places an administrator is signed in.
type sessionResponse struct {
	ID        string    `json:"id"`
	IP        string    `json:"ip"`
	UserAgent string    `json:"user_agent"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
	Active    bool      `json:"active"`
}

func newSessionResponses(sessions []model.AdminUserSession) []sessionResponse {
	now := time.Now()
	out := make([]sessionResponse, 0, len(sessions))

	for _, s := range sessions {
		out = append(out, sessionResponse{
			ID:        s.ID.String(),
			IP:        s.IP,
			UserAgent: s.UserAgent,
			CreatedAt: s.CreatedAt,
			ExpiresAt: s.ExpiresAt,
			Active:    s.IsActive(now),
		})
	}

	return out
}
