// Package model holds the database models. Every table in the database is
// declared here as a struct, and registered in All so it gets migrated.
package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Base carries the columns every table has. Embed it in each model.
type Base struct {
	ID        uuid.UUID      `gorm:"type:uuid;primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// BeforeCreate gives a row its id before it is written. GORM finds this hook
// through the embedded Base, so every model that embeds Base gets it.
func (b *Base) BeforeCreate(*gorm.DB) error {
	if b.ID != uuid.Nil {
		return nil
	}

	id, err := newID()
	if err != nil {
		return err
	}
	b.ID = id

	return nil
}

// newID returns a version 7 UUID: random, but with a timestamp in the leading
// bits, so ids sort by creation time and index writes stay local.
func newID() (uuid.UUID, error) {
	return uuid.NewV7()
}

// All returns every model, in the order they should be migrated: a model must
// come after anything it references. Add new models here, or they will not get
// a table.
func All() []any {
	return []any{
		&Organization{},
		&AdminSecurity{},
		&MailSettings{},
		&OTPSettings{},
		&LoginFlow{},
		&Language{},
		&Translation{},
		&SocialProvider{},
		&Role{},
		&Application{},
		&AdminUser{},
		&AdminRoleAssignment{},
		&AdminUserSession{},
		&MFA{},
		&AuditLog{},
		&UserField{},
		&API{},
		&APIScope{},
		&ApplicationAPI{},
		&ApplicationAPIScope{},
		&UserRole{},
		&User{},
		&SigningKey{},
		&AuthorizationRequest{},
		&AuthorizationCode{},
		&RefreshToken{},
		&UserSession{},
		&PasswordReset{},
		&EmailVerification{},
		&LoginCode{},
		&UserIdentity{},
		&SocialLogin{},
		&SSOConnection{},
		&SSOIdentity{},
		&SSOLogin{},
	}
}
