package model

// AdminSecurity is the single row of settings for how administrators sign in.
// It is seeded from LOGINER_ADMIN_MFA and owned by super admins afterwards.
type AdminSecurity struct {
	Base

	// RequireMFA makes every administrator set up an authenticator before
	// they can do anything else, and stops any of them turning it off again.
	RequireMFA bool `gorm:"not null" json:"require_mfa"`
}

// TableName pins the table name.
func (AdminSecurity) TableName() string {
	return "admin_security"
}

// DefaultAdminSecurity makes a second factor optional, so the first
// administrator is not forced to enrol on a server they are only trying out.
// Requiring it is a deliberate choice on the Administrators page.
func DefaultAdminSecurity() AdminSecurity {
	return AdminSecurity{RequireMFA: false}
}
