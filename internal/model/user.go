package model

import (
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type User struct {
	Base

	Email         string `gorm:"uniqueIndex;size:255;not null" json:"email"`
	EmailVerified bool   `gorm:"not null;default:false" json:"email_verified"`
	FirstName     string `gorm:"size:100" json:"first_name"`
	LastName      string `gorm:"size:100" json:"last_name"`
	IsActive      bool   `gorm:"not null" json:"is_active"`

	// PasswordHash is empty for a user who has never been given a password.
	// The password itself is never stored, and the hash never leaves the
	// server.
	PasswordHash string `gorm:"size:255;not null;default:''" json:"-"`
	// IsTemporaryPassword marks a password an administrator chose for the
	// user, which the user is expected to replace the next time they sign in.
	IsTemporaryPassword bool `gorm:"not null;default:false" json:"is_temporary_password"`
	// HasPassword tells the panel whether a password is set, without telling
	// it anything about the password.
	HasPassword bool `gorm:"-" json:"has_password"`

	// SocialAccounts are the providers this user signs in with. Like
	// HasPassword it is no column: the panel is told what a record amounts
	// to, and the rows themselves are user_identities.
	SocialAccounts []SocialAccount `gorm:"-" json:"social_accounts"`

	// LastLoginAt is when the user last signed in to an application.
	LastLoginAt *time.Time `json:"last_login_at"`
	// FailedLoginCount and LockedUntil lock the account for a while after
	// repeated wrong passwords, as for administrators.
	FailedLoginCount int        `gorm:"not null;default:0" json:"-"`
	LockedUntil      *time.Time `json:"locked_until"`

	Data map[string]any `gorm:"serializer:json" json:"data"`

	// Roles are the roles given to this user directly. What they add up to,
	// inheritance included, is RoleGraph.Effective.
	Roles []UserRole `gorm:"many2many:user_role_members;constraint:OnDelete:CASCADE" json:"roles"`
}

func (User) TableName() string {
	return "users"
}

// ErrPasswordTooLong is returned for a password bcrypt cannot hash: it reads
// no further than 72 bytes, and refuses rather than quietly ignoring the rest.
var ErrPasswordTooLong = errors.New("password must be at most 72 bytes")

// AfterFind fills in HasPassword for a record read from the database.
func (u *User) AfterFind(*gorm.DB) error {
	u.HasPassword = u.PasswordHash != ""
	return nil
}

// SetPassword replaces the user's password with a hash of the one given.
func (u *User) SetPassword(password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if errors.Is(err, bcrypt.ErrPasswordTooLong) {
		return ErrPasswordTooLong
	}
	if err != nil {
		return err
	}

	u.PasswordHash = string(hash)
	u.HasPassword = true

	return nil
}

// CanSignIn reports whether the user may sign in now: active, and not locked
// after too many wrong passwords.
func (u User) CanSignIn(now time.Time) bool {
	return u.IsActive && (u.LockedUntil == nil || now.After(*u.LockedUntil))
}

func (u User) FullName() string {
	switch {
	case u.FirstName != "" && u.LastName != "":
		return u.FirstName + " " + u.LastName
	case u.FirstName != "":
		return u.FirstName
	case u.LastName != "":
		return u.LastName
	default:
		return u.Email
	}
}
