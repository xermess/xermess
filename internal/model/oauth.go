package model

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// The provider's state between requests. Every secret given to a browser or
// client (code, refresh token, session cookie, reset link) is 256 random bits
// stored only as a SHA-256 hash, so a database copy cannot be replayed.

// Lifetimes of what the provider hands out that is not configured per
// application.
const (
	// AuthorizationRequestLifetime is how long someone has to sign in once an
	// application sent them to the authorization endpoint.
	AuthorizationRequestLifetime = 30 * time.Minute
	// AuthorizationCodeLifetime is how long a client has to exchange a code.
	// RFC 6749 recommends ten minutes at most; a client exchanges it at once.
	AuthorizationCodeLifetime = 2 * time.Minute
	// PasswordResetLifetime is how long a reset link works.
	PasswordResetLifetime = time.Hour
	// EmailVerificationLifetime is longer than a reset's: nothing is locked
	// while it waits, and it is often opened later on another device.
	EmailVerificationLifetime = 24 * time.Hour
)

// SigningKey is one private key tokens are signed with. The key is stored
// encrypted with the server's secret; see jose.Sealer.
type SigningKey struct {
	// KID is what a token's header names the key by, and what the JWKS lists.
	KID        string    `gorm:"column:kid;primaryKey;size:64"`
	Algorithm  string    `gorm:"size:16;not null;index"`
	PrivateKey []byte    `gorm:"not null"`
	CreatedAt  time.Time `gorm:"not null"`
	// RetiredAt is set when a newer key takes over; a retired key is still
	// published, so its tokens verify until they expire.
	RetiredAt *time.Time
}

// TableName pins the table name.
func (SigningKey) TableName() string {
	return "signing_keys"
}

// AuthorizationRequest is a sign-in under way, kept while the user signs in.
// The page knows it by an opaque handle whose hash is the key.
type AuthorizationRequest struct {
	Base

	HandleHash    string       `gorm:"size:64;uniqueIndex;not null"`
	ApplicationID uuid.UUID    `gorm:"type:uuid;not null;index"`
	Application   *Application `gorm:"constraint:OnDelete:CASCADE"`

	RedirectURI         string `gorm:"size:512;not null"`
	Scope               string `gorm:"type:text;not null"`
	State               string `gorm:"type:text"`
	Nonce               string `gorm:"type:text"`
	CodeChallenge       string `gorm:"size:128"`
	CodeChallengeMethod string `gorm:"size:8"`
	Audience            string `gorm:"size:255"`
	LoginHint           string `gorm:"size:255"`

	ExpiresAt time.Time `gorm:"not null;index"`
	// CompletedAt is set once a code has been issued for it, so the handle
	// cannot be used to sign in a second time.
	CompletedAt *time.Time
}

// TableName pins the table name.
func (AuthorizationRequest) TableName() string {
	return "authorization_requests"
}

// Usable reports whether the request can still be signed in to.
func (r AuthorizationRequest) Usable(now time.Time) bool {
	return r.CompletedAt == nil && now.Before(r.ExpiresAt)
}

// AuthorizationCode is the code the authorization endpoint redirects back
// with, carrying everything the token endpoint needs to issue tokens.
type AuthorizationCode struct {
	Base

	CodeHash      string       `gorm:"size:64;uniqueIndex;not null"`
	ApplicationID uuid.UUID    `gorm:"type:uuid;not null;index"`
	Application   *Application `gorm:"constraint:OnDelete:CASCADE"`
	UserID        uuid.UUID    `gorm:"type:uuid;not null;index"`
	User          *User        `gorm:"constraint:OnDelete:CASCADE"`
	SessionID     *uuid.UUID   `gorm:"type:uuid"`

	RedirectURI         string    `gorm:"size:512;not null"`
	Scope               string    `gorm:"type:text;not null"`
	Nonce               string    `gorm:"type:text"`
	CodeChallenge       string    `gorm:"size:128"`
	CodeChallengeMethod string    `gorm:"size:8"`
	Audience            string    `gorm:"size:255"`
	AuthenticatedAt     time.Time `gorm:"not null"`

	ExpiresAt time.Time `gorm:"not null;index"`
	// UsedAt is set on exchange. A second presentation is refused and revokes
	// what the first issued (RFC 6749 4.1.2).
	UsedAt *time.Time
}

// TableName pins the table name.
func (AuthorizationCode) TableName() string {
	return "authorization_codes"
}

// RefreshToken rotates on use; presenting a replaced one revokes the whole
// family (RFC 9700 4.14.2).
type RefreshToken struct {
	Base

	TokenHash     string       `gorm:"size:64;uniqueIndex;not null"`
	FamilyID      uuid.UUID    `gorm:"type:uuid;not null;index"`
	ApplicationID uuid.UUID    `gorm:"type:uuid;not null;index"`
	Application   *Application `gorm:"constraint:OnDelete:CASCADE"`
	UserID        uuid.UUID    `gorm:"type:uuid;not null;index"`
	User          *User        `gorm:"constraint:OnDelete:CASCADE"`
	// CodeID is the code the family started from, so a replayed code can
	// revoke what it was exchanged for.
	CodeID *uuid.UUID `gorm:"type:uuid;index"`

	Scope           string    `gorm:"type:text;not null"`
	Audience        string    `gorm:"size:255"`
	AuthenticatedAt time.Time `gorm:"not null"`

	// ExpiresAt is the family's: a rotated token keeps its predecessor's, so
	// rotating cannot keep a session alive for ever.
	ExpiresAt time.Time `gorm:"not null;index"`
	RevokedAt *time.Time
}

// TableName pins the table name.
func (RefreshToken) TableName() string {
	return "refresh_tokens"
}

// Usable reports whether the token can still be exchanged.
func (t RefreshToken) Usable(now time.Time) bool {
	return t.RevokedAt == nil && now.Before(t.ExpiresAt)
}

// UserSession is a user signed in in one browser, letting the authorization
// endpoint skip the login page.
type UserSession struct {
	Base

	TokenHash       string    `gorm:"size:64;uniqueIndex;not null"`
	UserID          uuid.UUID `gorm:"type:uuid;not null;index"`
	User            *User     `gorm:"constraint:OnDelete:CASCADE"`
	AuthenticatedAt time.Time `gorm:"not null"`
	ExpiresAt       time.Time `gorm:"not null;index"`
	RevokedAt       *time.Time
	IP              string `gorm:"size:45"`
	UserAgent       string `gorm:"size:255"`
}

// TableName pins the table name.
func (UserSession) TableName() string {
	return "user_sessions"
}

// Active reports whether the session still signs its user in.
func (s UserSession) Active(now time.Time) bool {
	return s.RevokedAt == nil && now.Before(s.ExpiresAt)
}

// PasswordReset is a password reset link that was sent.
type PasswordReset struct {
	Base

	TokenHash string    `gorm:"size:64;uniqueIndex;not null"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;index"`
	User      *User     `gorm:"constraint:OnDelete:CASCADE"`
	ExpiresAt time.Time `gorm:"not null;index"`
	UsedAt    *time.Time
}

// TableName pins the table name.
func (PasswordReset) TableName() string {
	return "password_resets"
}

// Usable reports whether the link still works.
func (r PasswordReset) Usable(now time.Time) bool {
	return r.UsedAt == nil && now.Before(r.ExpiresAt)
}

// EmailVerification is a link proving an address. With NewEmail set it is a
// pending change of address: the account moves only when the new inbox uses the
// link.
type EmailVerification struct {
	Base

	TokenHash string    `gorm:"size:64;uniqueIndex;not null"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;index"`
	User      *User     `gorm:"constraint:OnDelete:CASCADE"`
	ExpiresAt time.Time `gorm:"not null;index"`
	UsedAt    *time.Time

	// NewEmail is the address to move to; empty confirms the current address.
	NewEmail string `gorm:"size:255"`
}

// IsChange reports whether the link moves the account to another address.
func (v EmailVerification) IsChange() bool {
	return v.NewEmail != ""
}

// TableName pins the table name.
func (EmailVerification) TableName() string {
	return "email_verifications"
}

// Usable reports whether the link still works.
func (v EmailVerification) Usable(now time.Time) bool {
	return v.UsedAt == nil && now.Before(v.ExpiresAt)
}

// NewSecret returns a 256-bit URL-safe secret to hand out and the hash to
// store.
func NewSecret() (secret, hash string, err error) {
	var b [32]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", "", fmt.Errorf("generate secret: %w", err)
	}

	secret = base64.RawURLEncoding.EncodeToString(b[:])

	return secret, HashSecret(secret), nil
}

// HashSecret is what is stored for a secret, and what it is looked up by.
func HashSecret(secret string) string {
	sum := sha256.Sum256([]byte(secret))
	return hex.EncodeToString(sum[:])
}

// PKCE methods. Only S256 is accepted: "plain" protects nothing against an
// attacker who can read the authorization request (RFC 9700 section 2.1.1).
const PKCES256 = "S256"

// VerifyPKCE checks a verifier (43 to 128 unreserved characters, RFC 7636 4.1)
// against the S256 challenge.
func VerifyPKCE(challenge, verifier string) bool {
	if len(verifier) < 43 || len(verifier) > 128 {
		return false
	}

	for _, r := range verifier {
		unreserved := r >= 'A' && r <= 'Z' || r >= 'a' && r <= 'z' || r >= '0' && r <= '9' || strings.ContainsRune("-._~", r)
		if !unreserved {
			return false
		}
	}

	sum := sha256.Sum256([]byte(verifier))

	return base64.RawURLEncoding.EncodeToString(sum[:]) == challenge
}
