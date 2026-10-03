package model

// RoleSuperAdmin grants everything, including managing administrators, and
// cannot be edited or removed.
const RoleSuperAdmin = "super_admin"

// AdminRole is a named set of permissions from AdminPermissions, made by a
// super admin.
type AdminRole struct {
	Base

	Name        string `gorm:"type:varchar(64);uniqueIndex;not null" json:"name"`
	Description string `gorm:"size:255" json:"description,omitempty"`

	// Permissions are the names this role grants, from AdminPermissions.
	// super_admin keeps this empty: it is granted everything by name.
	Permissions []string `gorm:"type:text;serializer:json" json:"permissions"`
}

// TableName pins the table name.
func (AdminRole) TableName() string {
	return "admin_roles"
}

// IsSuperAdmin reports whether this is the built-in super_admin role.
func (r AdminRole) IsSuperAdmin() bool {
	return r.Name == RoleSuperAdmin
}

// Grants reports whether holding this role allows a permission.
func (r AdminRole) Grants(permission string) bool {
	if r.IsSuperAdmin() {
		return true
	}

	for _, granted := range r.Permissions {
		if granted == permission {
			return true
		}
	}

	return false
}
