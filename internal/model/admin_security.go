package model

// AdminSecurity is how administrators are made to sign in: the settings that
// apply to every one of them, rather than to any single account.
//
// There is one row, like the organisation's. It starts as the installation's
// configuration says (LOGINER_ADMIN_MFA), and is a super admin's to change
// afterwards — so turning two-factor sign-in on for everybody does not mean
// editing a file and restarting the server.
type AdminSecurity struct {
	Base

	// MFARequired makes every administrator set up an authenticator before
	// they can do anything else, and stops any of them turning it off again.
	MFARequired bool `gorm:"not null" json:"mfa_required"`
}

// TableName pins the table name.
func (AdminSecurity) TableName() string {
	return "admin_security"
}

// DefaultAdminSecurity is what an installation starts with when nothing says
// otherwise: a second factor for whoever wants one, and nobody made to.
//
// Off is the default because the alternative decides for people who have not
// been asked: the first administrator would be sent to set up an
// authenticator before they had seen the panel, on a server they may only be
// trying out. A super admin requires them of everybody on the Administrators
// page — which is the deliberate act it should be — and each administrator is
// then taken to /admin/mfa-setup to enrol before they may do anything else.
func DefaultAdminSecurity() AdminSecurity {
	return AdminSecurity{MFARequired: false}
}
