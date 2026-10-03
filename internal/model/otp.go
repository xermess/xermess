package model

import (
	"crypto/rand"
	"crypto/subtle"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/google/uuid"
)

// OTPSettings is the single row deciding how emailed one-time codes behave,
// read where codes are made and checked. Authenticator-app codes are RFC 6238
// (internal/totp) and are not configurable.
type OTPSettings struct {
	Base

	// CodeLength is how many digits a code has.
	CodeLength int `gorm:"not null" json:"code_length"`

	// LifetimeMinutes is how long a code works for.
	LifetimeMinutes int `gorm:"not null" json:"lifetime_minutes"`

	// MaxAttempts is how many wrong guesses spend a code.
	MaxAttempts int `gorm:"not null" json:"max_attempts"`

	// ResendSeconds is the wait before another code can go to the same address,
	// so resending cannot spam a stranger.
	ResendSeconds int `gorm:"not null" json:"resend_seconds"`
}

// TableName pins the table name.
func (OTPSettings) TableName() string {
	return "otp_settings"
}

// The bounds the panel is held to, and the reasons for them.
const (
	// MinOTPLength: below four digits the guesses become a lottery.
	MinOTPLength = 4
	// MaxOTPLength is ten, past which people copy rather than read.
	MaxOTPLength = 10

	// MaxOTPLifetimeMinutes is an hour. A code that outlives the message it
	// came in is a password sitting in an inbox.
	MaxOTPLifetimeMinutes = 60

	// MaxOTPAttempts is ten guesses, and MaxOTPResendSeconds fifteen
	// minutes — past which the button is not a wait but a wall.
	MaxOTPAttempts      = 10
	MaxOTPResendSeconds = 900
)

// DefaultOTPSettings: six digits, ten minutes, five guesses, a minute between
// messages.
func DefaultOTPSettings() OTPSettings {
	return OTPSettings{
		CodeLength:      6,
		LifetimeMinutes: 10,
		MaxAttempts:     5,
		ResendSeconds:   60,
	}
}

// Lifetime is how long a code works for.
func (o OTPSettings) Lifetime() time.Duration {
	return time.Duration(o.LifetimeMinutes) * time.Minute
}

// Resend is how long to wait before another code may be sent.
func (o OTPSettings) Resend() time.Duration {
	return time.Duration(o.ResendSeconds) * time.Second
}

// Validate reports the first thing wrong with the settings.
func (o OTPSettings) Validate() error {
	switch {
	case o.CodeLength < MinOTPLength || o.CodeLength > MaxOTPLength:
		return fmt.Errorf("code_length must be between %d and %d", MinOTPLength, MaxOTPLength)
	case o.LifetimeMinutes < 1 || o.LifetimeMinutes > MaxOTPLifetimeMinutes:
		return fmt.Errorf("lifetime_minutes must be between 1 and %d", MaxOTPLifetimeMinutes)
	case o.MaxAttempts < 1 || o.MaxAttempts > MaxOTPAttempts:
		return fmt.Errorf("max_attempts must be between 1 and %d", MaxOTPAttempts)
	case o.ResendSeconds < 0 || o.ResendSeconds > MaxOTPResendSeconds:
		return fmt.Errorf("resend_seconds must be between 0 and %d", MaxOTPResendSeconds)
	}

	return nil
}

// LoginCode is a sign-in waiting for an emailed code. The page holds only the
// handle, so a code alone is useless without the browser that asked for it.
type LoginCode struct {
	Base

	// HandleHash is what the page comes back with, hashed. The handle is the
	// secret here: a six-digit code is guessable and this is not.
	HandleHash string `gorm:"size:64;uniqueIndex;not null"`

	// CodeHash is the code itself, hashed. It is only ever compared.
	CodeHash string `gorm:"size:64;not null"`

	UserID uuid.UUID `gorm:"type:uuid;not null;index"`
	User   *User     `gorm:"constraint:OnDelete:CASCADE"`

	// Request is the sign-in under way, so the session returns to the asking
	// application; empty for the account itself.
	Request string `gorm:"size:64"`

	// RememberMe is the "stay signed in" choice made before the code was asked
	// for.
	RememberMe bool `gorm:"not null"`

	// Attempts is how many codes have been typed: every one is counted before
	// it is compared, the right one included.
	Attempts int `gorm:"not null"`

	// SentAt is when the last message went out, which is what the wait
	// before another is counted from.
	SentAt time.Time `gorm:"not null"`

	ExpiresAt time.Time `gorm:"not null;index"`
	UsedAt    *time.Time
}

// TableName pins the table name.
func (LoginCode) TableName() string {
	return "login_codes"
}

// Usable reports whether a code can still be typed: not used, not expired,
// and not guessed at more than it allows.
func (c LoginCode) Usable(now time.Time, maxAttempts int) bool {
	return c.UsedAt == nil && now.Before(c.ExpiresAt) && c.Attempts < maxAttempts
}

// NewOTP returns a crypto/rand code of `length` digits and its hash.
func NewOTP(length int) (code, hash string, err error) {
	var b strings.Builder
	for range length {
		digit, err := rand.Int(rand.Reader, big.NewInt(10))
		if err != nil {
			return "", "", fmt.Errorf("generate a one-time code: %w", err)
		}
		b.WriteString(digit.String())
	}

	code = b.String()

	return code, HashSecret(code), nil
}

// MatchesOTP reports whether a typed code is the one that was sent, comparing
// the hashes in constant time so nothing is learnt from how long it took.
func (c LoginCode) MatchesOTP(typed string) bool {
	typed = strings.Map(func(r rune) rune {
		if r >= '0' && r <= '9' {
			return r
		}
		return -1
	}, typed)

	if typed == "" {
		return false
	}

	return subtle.ConstantTimeCompare([]byte(c.CodeHash), []byte(HashSecret(typed))) == 1
}
