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

// The tables below are the provider's memory between requests: the keys
// tokens are signed with, a sign-in that is under way, the code it ends with,
// the refresh tokens and sessions that outlive it, and password resets.
//
// Every secret that is handed to a browser or a client — a code, a refresh
// token, a session cookie, a reset link — is stored as a SHA-256 hash only, so
// a copy of the database cannot be replayed. They are 256 random bits, which
// is why a fast hash is right, as it is for the admin sessions.

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
	// EmailVerificationLifetime is how long a link confirming an address
	// works. Longer than a reset: nobody is locked out while it waits, and it
	// is often opened on another device, later.
	EmailVerificationLifetime = 24 * time.Hour
)

// SigningKey is one private key tokens are signed with. The key is stored
// encrypted with the server's secret; see jose.Sealer.
type SigningKey struct {
	// KID is what a token's header names the key by, and what the JWKS lists.
	KID       string    `gorm:"primaryKey;size:64"`
	Algorithm string    `gorm:"size:16;not null;index"`
	Sealed    []byte    `gorm:"not null"`
	CreatedAt time.Time `gorm:"not null"`
	// RetiredAt is set when a newer key replaces this one. A retired key
	// still verifies, and is still published, so tokens it signed stay valid
	// until they expire.
	RetiredAt *time.Time
}

// TableName pins the table name.
func (SigningKey) TableName() string {
	return "signing_keys"
}

// AuthorizationRequest is a sign-in under way: what an application asked the
// authorization endpoint for, kept while the user signs in on the login page.
// The page knows it by an opaque handle, whose hash is the row's key.
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
	AuthTime            time.Time `gorm:"not null"`

	ExpiresAt time.Time `gorm:"not null;index"`
	// UsedAt is set when the code is exchanged. A code presented a second
	// time is refused, and the tokens the first exchange issued are revoked
	// (RFC 6749 section 4.1.2): the second presenter may be the thief.
	UsedAt *time.Time
}

// TableName pins the table name.
func (AuthorizationCode) TableName() string {
	return "authorization_codes"
}

// RefreshToken is a refresh token. Using one replaces it with a new one in the
// same family; presenting a replaced one again revokes the whole family, since
// only a copy could still be holding it (RFC 9700 section 4.14.2).
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

	Scope    string    `gorm:"type:text;not null"`
	Audience string    `gorm:"size:255"`
	AuthTime time.Time `gorm:"not null"`

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

// UserSession is a user being signed in at this server, in the browser that
// holds its cookie. It is what lets the authorization endpoint skip the login
// page for someone who signed in a moment ago.
type UserSession struct {
	Base

	TokenHash string    `gorm:"size:64;uniqueIndex;not null"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;index"`
	User      *User     `gorm:"constraint:OnDelete:CASCADE"`
	AuthTime  time.Time `gorm:"not null"`

	// Method is the way in that made this session, so an application whose
	// login flow asks for more than it proved can say so (LoginFlow.Accepts).
	Method SignInMethod `gorm:"type:varchar(16);not null"`

	ExpiresAt time.Time `gorm:"not null;index"`
	RevokedAt *time.Time
	IP        string `gorm:"size:45"`
	UserAgent string `gorm:"size:255"`
}

// SignInMethod is the way in that made a session: what the person did to prove
// they were themselves.
//
// It is kept on the row because a session outlives the sign-in that made it.
// One cookie carries it to every application, while the rules it was made
// under are one login flow's — the flow of whichever application the person
// happened to sign in through. An application whose flow asks for more has to
// be able to tell, which is what LoginFlow.Accepts is for.
type SignInMethod string

const (
	// MethodPassword is an address and a password typed on the sign-in page.
	// An account made there is signed in the same way.
	MethodPassword SignInMethod = "password"

	// MethodEmailCode is a password, and then a code emailed to the address
	// (StepEmailCode).
	MethodEmailCode SignInMethod = "email_code"

	// MethodSocial is an account with one of the providers on the Social
	// page, which proved the address itself.
	MethodSocial SignInMethod = "social"

	// MethodSSO is an organisation's own identity provider.
	MethodSSO SignInMethod = "sso"
)

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

// EmailVerification is a link sent to an address to prove it belongs to
// whoever signs in with it: what a login flow that requires a verified
// address sends an account that has not proved its own, and what a new
// account is sent where the flow says to.
//
// It is also how an address is changed. NewEmail set makes the link a pending
// change rather than a confirmation: the link goes to the address somebody
// typed, and only using it moves the account there. That way round nobody can
// take an account by typing an address they cannot read, and nobody loses one
// by typing an address they meant to spell differently.
type EmailVerification struct {
	Base

	TokenHash string    `gorm:"size:64;uniqueIndex;not null"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;index"`
	User      *User     `gorm:"constraint:OnDelete:CASCADE"`
	ExpiresAt time.Time `gorm:"not null;index"`
	UsedAt    *time.Time

	// NewEmail is the address to move the account to, for a link that is a
	// change. Empty is a link that only confirms the address the account
	// already has.
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

// NewSecret returns a secret to hand out — a code, a token, a handle — and the
// hash to store for it: 256 random bits, base64url encoded so it is safe in a
// URL as it is.
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

// VerifyPKCE reports whether a code verifier matches the challenge sent with
// the authorization request. The verifier has to be 43 to 128 characters of
// the unreserved set (RFC 7636 section 4.1).
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
