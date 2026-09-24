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

// OTPSettings is how the one-time codes this server emails behave: how long
// one is, how long it lasts, how many guesses it takes, and how soon another
// may be asked for.
//
// There is one row, like the organisation's. It is seeded with the defaults
// below on the first start and is a super admin's afterwards, and it is read
// where a code is made and checked rather than held from startup — so making
// codes shorter for a shop, or longer for a staff tool, takes effect on the
// next sign-in.
//
// The codes an authenticator app shows are not these. Those are RFC 6238,
// whose parameters every app assumes (internal/totp), and changing them would
// mean apps showing the wrong codes; these are the codes the server sends to
// an address itself.
type OTPSettings struct {
	Base

	// Length is how many digits a code has.
	Length int `gorm:"not null" json:"length"`

	// LifetimeMinutes is how long a code works for.
	LifetimeMinutes int `gorm:"not null" json:"lifetime_minutes"`

	// MaxAttempts is how many wrong guesses a code takes before it is spent.
	// Without it a six-digit code is a million guesses away from anybody
	// patient.
	MaxAttempts int `gorm:"not null" json:"max_attempts"`

	// ResendSeconds is how long somebody waits before another code can be
	// sent to the same address. It is what keeps the "send it again" button
	// from being a way to post mail to a stranger.
	ResendSeconds int `gorm:"not null" json:"resend_seconds"`
}

// TableName pins the table name.
func (OTPSettings) TableName() string {
	return "otp_settings"
}

// The bounds the panel is held to, and the reasons for them.
const (
	// MinOTPLength is four digits, which is short enough to be worth
	// refusing below: ten thousand codes and five guesses is a lottery
	// somebody would run.
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

// DefaultOTPSettings is what an installation starts with: six digits for ten
// minutes, five guesses, and a minute between messages. It is what almost
// every service that emails a code does, which is the point — somebody
// signing in has seen it before.
func DefaultOTPSettings() OTPSettings {
	return OTPSettings{
		Length:          6,
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
	case o.Length < MinOTPLength || o.Length > MaxOTPLength:
		return fmt.Errorf("length must be between %d and %d", MinOTPLength, MaxOTPLength)
	case o.LifetimeMinutes < 1 || o.LifetimeMinutes > MaxOTPLifetimeMinutes:
		return fmt.Errorf("lifetime_minutes must be between 1 and %d", MaxOTPLifetimeMinutes)
	case o.MaxAttempts < 1 || o.MaxAttempts > MaxOTPAttempts:
		return fmt.Errorf("max_attempts must be between 1 and %d", MaxOTPAttempts)
	case o.ResendSeconds < 0 || o.ResendSeconds > MaxOTPResendSeconds:
		return fmt.Errorf("resend_seconds must be between 0 and %d", MaxOTPResendSeconds)
	}

	return nil
}

// LoginCode is a sign-in waiting for a code that was emailed: who it is for,
// what they have to type, and how far they have got with it.
//
// The row is the sign-in itself, held between the password being accepted and
// the session being made. The page carries HandleHash's handle and nothing
// else, so a code on its own is no use to whoever read the message over
// somebody's shoulder — they would need the browser it was asked from too.
type LoginCode struct {
	Base

	// HandleHash is what the page comes back with, hashed. The handle is the
	// secret here: a six-digit code is guessable and this is not.
	HandleHash string `gorm:"size:64;uniqueIndex;not null"`

	// CodeHash is the code itself, hashed. It is only ever compared.
	CodeHash string `gorm:"size:64;not null"`

	UserID uuid.UUID `gorm:"type:uuid;not null;index"`
	User   *User     `gorm:"constraint:OnDelete:CASCADE"`

	// Request is the sign-in under way this belongs to, so the session made
	// once the code is right goes back to the application that asked. Empty
	// when somebody is signing in to their account itself.
	Request string `gorm:"size:64"`

	// Remember is the "stay signed in" box as it was ticked before the code
	// was asked for: the sign-in is one act, and the answer to it should not
	// depend on where the code arrived.
	Remember bool `gorm:"not null"`

	// Attempts is how many wrong codes have been typed.
	Attempts int `gorm:"not null"`

	// SentAt is when the last message went out, which is what the wait
	// before another is counted from.
	SentAt time.Time `gorm:"not null"`

	ExpiresAt  time.Time `gorm:"not null;index"`
	ConsumedAt *time.Time
}

// TableName pins the table name.
func (LoginCode) TableName() string {
	return "login_codes"
}

// Usable reports whether a code can still be typed: not used, not expired,
// and not guessed at more than it allows.
func (c LoginCode) Usable(now time.Time, maxAttempts int) bool {
	return c.ConsumedAt == nil && now.Before(c.ExpiresAt) && c.Attempts < maxAttempts
}

// NewOTP returns a code of `length` digits and its hash. Every digit comes
// from crypto/rand: a code somebody could work out from the last one is not a
// second factor.
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
