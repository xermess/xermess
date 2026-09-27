package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AdminRoleAssignment gives an administrator a role, either for the whole
// panel or for one application.
//
// The same role can be assigned more than once with different scopes — a
// "moderator" of shop and of blog — which is how Zitadel's project managers
// and Keycloak's per-client admin roles work. A scoped assignment grants only
// the role's scopable permissions, and only for its application; super_admin
// is only ever assigned for the whole panel.
type AdminRoleAssignment struct {
	ID        uuid.UUID `gorm:"type:uuid;primarykey" json:"id"`
	CreatedAt time.Time `json:"created_at"`

	AdminID uuid.UUID `gorm:"type:uuid;not null;index" json:"admin_id"`
	Admin   *Admin    `gorm:"constraint:OnDelete:CASCADE" json:"-"`

	RoleID uuid.UUID `gorm:"type:uuid;not null;index" json:"role_id"`
	Role   AdminRole `gorm:"constraint:OnDelete:CASCADE" json:"role"`

	// ApplicationID is the application the role is held for, or nil for the
	// whole panel.
	ApplicationID *uuid.UUID   `gorm:"type:uuid;index" json:"application_id"`
	Application   *Application `gorm:"constraint:OnDelete:CASCADE" json:"application,omitempty"`
}

// TableName pins the table name.
func (AdminRoleAssignment) TableName() string {
	return "admin_role_assignments"
}

// BeforeCreate gives the row its id.
func (a *AdminRoleAssignment) BeforeCreate(*gorm.DB) error {
	if a.ID != uuid.Nil {
		return nil
	}

	id, err := newID()
	if err != nil {
		return err
	}
	a.ID = id

	return nil
}

// Global reports whether the assignment covers the whole panel.
func (a AdminRoleAssignment) Global() bool {
	return a.ApplicationID == nil
}

// Grants reports whether the assignment allows a permission for an
// application. A nil application asks about the panel as a whole, which only
// a global assignment can answer yes to.
func (a AdminRoleAssignment) Grants(permission string, application *uuid.UUID) bool {
	if a.Global() {
		return a.Role.Grants(permission)
	}

	return application != nil &&
		*a.ApplicationID == *application &&
		IsScopablePermission(permission) &&
		a.Role.Grants(permission)
}
