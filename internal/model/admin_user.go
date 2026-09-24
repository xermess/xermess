package model

import (
	"errors"
	"slices"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// AdminUser is a member of staff who can sign in to the admin panel.
type AdminUser struct {
	Base

	Username     string `gorm:"uniqueIndex;size:100;not null" json:"username"`
	Email        string `gorm:"uniqueIndex;size:255;not null" json:"email"`
	FirstName    string `gorm:"size:100;not null" json:"first_name"`
	LastName     string `gorm:"size:100;not null" json:"last_name"`
	PasswordHash string `gorm:"size:255;not null" json:"-"`
	IsActive     bool   `gorm:"-" json:"is_active"`
	Status       Status `gorm:"type:varchar(32);index;not null;default:invited" json:"status"`

	EmailVerifiedAt *time.Time `json:"email_verified_at,omitempty"`
	LastLoginAt     *time.Time `json:"last_login_at,omitempty"`
	LastLoginIP     string     `gorm:"size:45" json:"last_login_ip,omitempty"`

	// FailedLoginCount and LockedUntil back a lockout after repeated failures.
	FailedLoginCount int        `gorm:"not null;default:0" json:"-"`
	LockedUntil      *time.Time `json:"locked_until,omitempty"`

	// Assignments are the roles the administrator holds, each for the whole
	// panel or for one application.
	Assignments []AdminRoleAssignment `gorm:"constraint:OnDelete:CASCADE" json:"-"`
	Sessions    []AdminUserSession    `gorm:"constraint:OnDelete:CASCADE" json:"-"`
	MFA         []MFA                 `gorm:"constraint:OnDelete:CASCADE" json:"-"`
}

// TableName pins the table name so renaming the struct cannot silently rename
// the table.
func (AdminUser) TableName() string {
	return "admin_users"
}

// MinAdminPasswordLength is the shortest password an administrator may have,
// whoever sets it: an admin account opens the panel, so it is held to more
// than a user's.
const MinAdminPasswordLength = 10

// SetPassword replaces the administrator's password with a hash of the one
// given. The password itself is never stored; one bcrypt cannot hash is
// ErrPasswordTooLong, the same as a user's.
func (a *AdminUser) SetPassword(password string) error {
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
func (a AdminUser) FullName() string {
	return a.FirstName + " " + a.LastName
}

// CanSignIn reports whether the account is in a state that allows signing in.
func (a AdminUser) CanSignIn(now time.Time) bool {
	if a.Status != StatusActive {
		return false
	}
	return a.LockedUntil == nil || now.After(*a.LockedUntil)
}

// HasRole reports whether the admin holds the given role for the whole panel.
// Assignments must be loaded for this to mean anything.
func (a AdminUser) HasRole(name string) bool {
	for _, assignment := range a.Assignments {
		if assignment.Global() && assignment.Role.Name == name {
			return true
		}
	}
	return false
}

// IsSuperAdmin reports whether the admin holds super_admin, which is what
// managing other administrators and their roles needs.
func (a AdminUser) IsSuperAdmin() bool {
	return a.HasRole(RoleSuperAdmin)
}

// HasPermission reports whether the admin may do something across the whole
// panel: some role assigned for the whole panel grants it.
func (a AdminUser) HasPermission(name string) bool {
	return a.HasPermissionFor(name, nil)
}

// HasPermissionFor reports whether the admin may do something to one
// application: a role assigned for the whole panel grants it, or one assigned
// for that application does and the permission can be scoped.
func (a AdminUser) HasPermissionFor(name string, application *uuid.UUID) bool {
	for _, assignment := range a.Assignments {
		if assignment.Grants(name, application) {
			return true
		}
	}
	return false
}

// HasPermissionAnywhere reports whether the admin may do something to at least
// one application, which is what opening a list that is then narrowed to
// their applications needs.
func (a AdminUser) HasPermissionAnywhere(name string) bool {
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

// ApplicationsWith says which applications the admin may do something to:
// every one when a whole-panel role grants it, otherwise the ids of the
// applications a scoped role grants it for.
func (a AdminUser) ApplicationsWith(name string) (all bool, ids []uuid.UUID) {
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
func (a AdminUser) Permissions() []string {
	granted := []string{}
	for _, name := range AdminPermissionNames() {
		if a.HasPermission(name) {
			granted = append(granted, name)
		}
	}
	return granted
}

// ScopedPermissions is, for each application the admin holds a role for, the
// scopable permissions they have there beyond what they hold for the whole
// panel.
func (a AdminUser) ScopedPermissions() map[uuid.UUID][]string {
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
