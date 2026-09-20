package model

import (
	"time"

	"github.com/google/uuid"
)

// MFA is one second factor enrolled by an admin. An admin may have more than
// one, so a lost phone is not a lockout.
type MFA struct {
	Base

	AdminUserID uuid.UUID `gorm:"type:uuid;index;not null" json:"admin_user_id"`
	AdminUser   AdminUser `gorm:"foreignKey:AdminUserID" json:"-"`

	Method MFAMethod `gorm:"type:varchar(32);index;not null" json:"method"`
	Label  string    `gorm:"size:100" json:"label,omitempty"`

	// Secret is the TOTP seed, or the WebAuthn credential. Encrypt it before
	// it is written; the database should never hold it in the clear.
	Secret string `gorm:"size:255;not null" json:"-"`

	// RecoveryCodes holds the hashes of the one-time recovery codes, never the
	// codes themselves.
	RecoveryCodes []string `gorm:"serializer:json" json:"-"`

	ConfirmedAt *time.Time `json:"confirmed_at,omitempty"`
	// LastUsedAt is the start of the last TOTP step this factor accepted, so
	// that step can be read back exactly (UsedStep) and never accepted twice.
	// It is not a note of when the factor was last used by any means: a
	// recovery code leaves it alone, or the step just gone would look spent.
	LastUsedAt *time.Time `json:"last_used_at,omitempty"`
}

// TableName pins the table name.
func (MFA) TableName() string {
	return "mfa"
}

// IsConfirmed reports whether enrolment was finished. An unconfirmed factor
// must not be accepted at sign-in.
func (m MFA) IsConfirmed() bool {
	return m.ConfirmedAt != nil
}

// HasMFA reports whether the administrator has a confirmed second factor. The
// store loads only confirmed factors onto an administrator, so any is one.
func (a AdminUser) HasMFA() bool {
	for _, factor := range a.MFA {
		if factor.IsConfirmed() {
			return true
		}
	}
	return false
}

// UsedStep is the last TOTP step a factor accepted. LastUsedAt is written as
// the start of that step, so the step can be read back from it exactly.
func (m MFA) UsedStep(period time.Duration) int64 {
	if m.LastUsedAt == nil {
		return 0
	}
	return m.LastUsedAt.Unix() / int64(period/time.Second)
}
