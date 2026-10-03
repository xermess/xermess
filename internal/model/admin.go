package model

import (
	"errors"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// Admin is a member of staff who can sign in to the admin panel.
type Admin struct {
	Base

	Username  string `gorm:"uniqueIndex;size:100;not null" json:"username"`
	Email     string `gorm:"uniqueIndex;size:255;not null" json:"email"`
	FirstName string `gorm:"size:100;not null" json:"first_name"`
	LastName  string `gorm:"size:100;not null" json:"last_name"`
	// AvatarURL is an absolute http(s) address of a picture of them, shown
	// in the panel's header and menus; empty, their initials stand in.
	AvatarURL    string `gorm:"size:512;not null;default:''" json:"avatar_url"`
	PasswordHash string `gorm:"size:255;not null" json:"-"`
	IsActive     bool   `gorm:"-" json:"is_active"`
	Status       Status `gorm:"type:varchar(32);index;not null;default:invited" json:"status"`

	EmailVerifiedAt *time.Time `json:"email_verified_at,omitempty"`
	LastLoginAt     *time.Time `json:"last_login_at,omitempty"`
	LastLoginIP     string     `gorm:"size:45" json:"last_login_ip,omitempty"`

	// FailedLoginCount and LockedUntil back a lockout after repeated failures.
	FailedLoginCount int        `gorm:"not null;default:0" json:"-"`
	LockedUntil      *time.Time `json:"locked_until,omitempty"`

	// Assignments are the roles held, panel-wide or per application. Service
	// marks an application calling with a token (ServiceAdmin); it is never
	// stored.
	Service bool `gorm:"-" json:"-"`

	Assignments []AdminRoleAssignment `gorm:"constraint:OnDelete:CASCADE" json:"-"`
	Sessions    []AdminSession        `gorm:"constraint:OnDelete:CASCADE" json:"-"`
	MFA         []MFA                 `gorm:"constraint:OnDelete:CASCADE" json:"-"`
}

// TableName pins the table name so renaming the struct cannot silently rename
// the table.
func (Admin) TableName() string {
	return "admins"
}

// MinAdminPasswordLength is stricter than a user's, since an admin account
// opens the panel.
const MinAdminPasswordLength = 10

// SetPassword stores a bcrypt hash; a password bcrypt cannot hash is
// ErrPasswordTooLong.
func (a *Admin) SetPassword(password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if errors.Is(err, bcrypt.ErrPasswordTooLong) {
		return ErrPasswordTooLong
	}
	if err != nil {
		return err
	}

	a.PasswordHash = string(hash)

	return nil
}

// FullName is the admin's display name.
func (a Admin) FullName() string {
	return strings.TrimSpace(a.FirstName + " " + a.LastName)
}

// CanSignIn reports whether the account is in a state that allows signing in.
func (a Admin) CanSignIn(now time.Time) bool {
	if a.Status != StatusActive {
		return false
	}
	return a.LockedUntil == nil || now.After(*a.LockedUntil)
}

// HasRole reports whether the admin holds the given role for the whole panel.
// Assignments must be loaded for this to mean anything.
func (a Admin) HasRole(name string) bool {
	for _, assignment := range a.Assignments {
		if assignment.Global() && assignment.Role.Name == name {
			return true
		}
	}
	return false
}

// IsSuperAdmin reports whether the admin holds super_admin, which is what
// managing other administrators and their roles needs.
func (a Admin) IsSuperAdmin() bool {
	return a.HasRole(RoleSuperAdmin)
}

// HasPermission reports whether the admin may do something across the whole
// panel: some role assigned for the whole panel grants it.
func (a Admin) HasPermission(name string) bool {
	return a.HasPermissionFor(name, nil)
}

// HasPermissionFor reports whether the admin may act on one application: via a
// panel-wide role, or a role scoped to it for a scopable permission.
func (a Admin) HasPermissionFor(name string, application *uuid.UUID) bool {
	for _, assignment := range a.Assignments {
		if assignment.Grants(name, application) {
			return true
		}
	}
	return false
}

// HasPermissionAnywhere reports whether the admin may act on at least one
// application.
func (a Admin) HasPermissionAnywhere(name string) bool {
	if a.HasPermission(name) {
		return true
	}

	for _, assignment := range a.Assignments {
		if !assignment.Global() && assignment.Grants(name, assignment.ApplicationID) {
			return true
		}
	}
	return false
}

// ApplicationsWith returns all=true for a panel-wide grant, otherwise the ids
// of applications a scoped role grants it for.
func (a Admin) ApplicationsWith(name string) (all bool, ids []uuid.UUID) {
	if a.HasPermission(name) {
		return true, nil
	}

	ids = []uuid.UUID{}
	for _, assignment := range a.Assignments {
		if !assignment.Global() && assignment.Grants(name, assignment.ApplicationID) {
			ids = append(ids, *assignment.ApplicationID)
		}
	}
	return false, ids
}

// Permissions is every catalog permission the admin holds for the whole
// panel, in catalog order.
func (a Admin) Permissions() []string {
	granted := []string{}
	for _, name := range AdminPermissionNames() {
		if a.HasPermission(name) {
			granted = append(granted, name)
		}
	}
	return granted
}

// ScopedPermissions lists, per application, the scopable permissions held there
// beyond the panel-wide ones.
func (a Admin) ScopedPermissions() map[uuid.UUID][]string {
	scoped := map[uuid.UUID][]string{}

	for _, assignment := range a.Assignments {
		if assignment.Global() {
			continue
		}

		app := *assignment.ApplicationID
		for _, name := range AdminPermissionNames() {
			if a.HasPermission(name) || !assignment.Grants(name, &app) {
				continue
			}
			if !slices.Contains(scoped[app], name) {
				scoped[app] = append(scoped[app], name)
			}
		}
	}

	return scoped
}

// ServiceAdmin is an application calling the admin API with a token, presented
// as an administrator holding exactly its token's permissions and never a super
// admin. Username is its client ID, which the activity log shows.
func ServiceAdmin(app Application, permissions []string) *Admin {
	return &Admin{
		Username:  app.ClientID,
		FirstName: app.Name,
		IsActive:  true,
		Service:   true,
		Status:    StatusActive,
		Assignments: []AdminRoleAssignment{{
			Role: AdminRole{Name: "application:" + app.ClientID, Permissions: permissions},
		}},
	}
}
