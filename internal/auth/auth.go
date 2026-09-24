// Package auth signs administrators in and out — with a password and, when
// they have one or it is required, a second factor — and keeps the record of
// what they did.
package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"xermess/internal/jose"
	"xermess/internal/model"
	"xermess/internal/store"
)

// SessionLifetime is how long a session lasts before the administrator has to
// sign in again.
const SessionLifetime = 12 * time.Hour

// MaxFailedLogins wrong passwords in a row lock an account for LockoutDuration.
// The lock is short on purpose: long enough to make guessing hopeless, not so
// long that someone else's guessing keeps an administrator out for good.
const (
	MaxFailedLogins = 5
	LockoutDuration = 15 * time.Minute
)

// dummyHash is compared against when a username does not exist, so that an
// unknown username takes as long to refuse as a wrong password. It has to be
// a real hash at the real cost: a malformed one is refused before any hashing
// is done, which would answer unknown usernames measurably faster.
var dummyHash = sync.OnceValue(func() []byte {
	hash, err := bcrypt.GenerateFromPassword([]byte("not a password anyone has"), bcrypt.DefaultCost)
	if err != nil {
		panic(fmt.Sprintf("auth: make the dummy hash: %v", err))
	}

	return hash
})

// ErrInvalidCredentials is returned for a username that does not exist, a
// wrong password, and an account that may not sign in. They are one error on
// purpose: telling them apart tells an attacker which usernames are real.
var ErrInvalidCredentials = errors.New("invalid credentials")

// ErrNoSession is returned when a request carries no usable session.
var ErrNoSession = errors.New("no active session")

// State is how far a session has got.
type State string

const (
	// StateNone is no usable session at all.
	StateNone State = "none"
	// StateSignedIn is a session that may use the panel.
	StateSignedIn State = "signed_in"
	// StateMFA is a session whose password was right, waiting for a code
	// from the administrator's authenticator. It can do nothing but give one.
	StateMFA State = "mfa"
	// StateEnroll is a session whose password was right, for an
	// administrator who has to set up two-factor sign-in before anything
	// else. It can do nothing but that.
	StateEnroll State = "enroll"
)

// How long a session may wait half signed in. Long enough to find a phone, or
// to set an authenticator up; not so long that a password alone keeps a door
// ajar.
const (
	challengeLifetime = 10 * time.Minute
	enrolmentLifetime = 30 * time.Minute
)

// Service signs administrators in and out. Every query it makes goes through
// the store, so this file is about what signing in means rather than about
// how the rows are fetched.
type Service struct {
	store  *store.Store
	sealer *jose.Sealer
	log    *slog.Logger

	// issuer names the panel in authenticator apps.
	issuer string
}

// New returns a Service backed by the given store. `sealer` encrypts TOTP
// secrets; `issuer` is what authenticator apps list the account under; `log`
// reports what cannot be returned, such as a failed audit write.
//
// Whether a second factor is compulsory is not passed in: it is a setting a
// super admin changes in the panel, read from the database each time it
// matters, so turning it on takes effect on the next sign-in rather than on
// the next restart.
func New(st *store.Store, sealer *jose.Sealer, log *slog.Logger, issuer string) *Service {
	return &Service{store: st, sealer: sealer, log: log, issuer: issuer}
}

// MFARequired reports whether every administrator must use a second factor.
//
// An installation that has not said starts with no — see
// model.DefaultAdminSecurity. A database that cannot be *read* is not the
// same thing: the answer is unknown, and an unknown answer to "must this
// person prove who they are twice" is yes.
func (s *Service) MFARequired(ctx context.Context) bool {
	security, err := s.store.AdminSecurity(ctx)
	if err != nil {
		s.log.Error("reading the admin security settings failed", "error", err)
		return true
	}

	return security.MFARequired
}

// Request describes where a call came from, which is recorded on the session
// and in the log.
type Request struct {
	IP        string
	UserAgent string
}

// Login checks the credentials and starts a session. The token it returns is
// the only copy: the database keeps a hash of it, so a leaked database cannot
// be used to sign in.
//
// A right password is not always a finished sign-in. An administrator with a
// second factor gets a session in StateMFA, and one who must set a factor up
// gets StateEnroll; either can do nothing else until that step is done.
func (s *Service) Login(ctx context.Context, username, password string, req Request) (string, *model.AdminUser, State, error) {
	admin, err := s.store.AdminByUsername(ctx, username)

	switch {
	case errors.Is(err, store.ErrNotFound):
		// Still hash something, so a missing username and a wrong password
		// take the same time to answer.
		_ = bcrypt.CompareHashAndPassword(dummyHash(), []byte(password))
		s.record(ctx, nil, username, "admin.login_failed", req, "unknown username")
		return "", nil, StateNone, ErrInvalidCredentials
	case err != nil:
		return "", nil, StateNone, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(admin.PasswordHash), []byte(password)); err != nil {
		locked, err := s.store.RecordFailedLogin(ctx, admin, time.Now(), MaxFailedLogins, LockoutDuration)
		if err != nil {
			return "", nil, StateNone, err
		}

		reason := "wrong password"
		if locked {
			reason = "wrong password; locked after too many attempts"
		}
		s.record(ctx, &admin.ID, admin.Username, "admin.login_failed", req, reason)

		return "", nil, StateNone, ErrInvalidCredentials
	}

	if !admin.CanSignIn(time.Now()) {
		reason := string(admin.Status)
		if admin.Status == model.StatusActive {
			reason = "locked"
		}
		s.record(ctx, &admin.ID, admin.Username, "admin.login_blocked", req, reason)

		return "", nil, StateNone, ErrInvalidCredentials
	}

	state, lifetime := StateSignedIn, SessionLifetime
	switch {
	case admin.HasMFA():
		state, lifetime = StateMFA, challengeLifetime
	case s.MFARequired(ctx):
		state, lifetime = StateEnroll, enrolmentLifetime
	}

	token, err := newToken()
	if err != nil {
		return "", nil, StateNone, err
	}

	session := model.AdminUserSession{
		AdminUserID: admin.ID,
		TokenHash:   hashToken(token),
		ExpiresAt:   time.Now().Add(lifetime),
		MFAPassed:   state == StateSignedIn,
		UserAgent:   req.UserAgent,
		IP:          req.IP,
	}
	if err := s.store.CreateSession(ctx, &session); err != nil {
		return "", nil, StateNone, err
	}

	if state == StateSignedIn {
		if err := s.signedIn(ctx, admin, req, ""); err != nil {
			return "", nil, StateNone, err
		}
	}

	return token, admin, state, nil
}

// signedIn finishes a sign-in: the last sign-in is recorded, the wrong
// passwords before it forgotten, and the log told how it happened.
func (s *Service) signedIn(ctx context.Context, admin *model.AdminUser, req Request, method string) error {
	if err := s.store.MarkAdminSignedIn(ctx, admin, time.Now(), req.IP); err != nil {
		return err
	}

	s.recordWith(ctx, &admin.ID, admin.Username, "admin.login", req, map[string]any{"second_factor": method})

	return nil
}

// Session returns the session a token carries, whatever state it is in, with
// its administrator.
func (s *Service) Session(ctx context.Context, token string) (*model.AdminUser, *model.AdminUserSession, State, error) {
	if token == "" {
		return nil, nil, StateNone, ErrNoSession
	}

	session, err := s.store.SessionByTokenHash(ctx, hashToken(token))
	switch {
	case errors.Is(err, store.ErrNotFound):
		return nil, nil, StateNone, ErrNoSession
	case err != nil:
		return nil, nil, StateNone, err
	}

	now := time.Now()
	if session.RevokedAt != nil || !now.Before(session.ExpiresAt) {
		return nil, nil, StateNone, ErrNoSession
	}

	admin, err := s.store.AdminByID(ctx, session.AdminUserID)
	if err != nil || !admin.CanSignIn(now) {
		return nil, nil, StateNone, ErrNoSession
	}

	switch {
	case session.MFAPassed && s.MFARequired(ctx) && !admin.HasMFA():
		// Signed in before two-factor sign-in was required, or had it reset:
		// the session is good for setting it up and nothing else.
		return admin, session, StateEnroll, nil
	case session.MFAPassed:
		return admin, session, StateSignedIn, nil
	case admin.HasMFA():
		return admin, session, StateMFA, nil
	case s.MFARequired(ctx):
		return admin, session, StateEnroll, nil
	default:
		// Waiting for a factor that has since been removed, where none is
		// required: sign in again.
		return nil, nil, StateNone, ErrNoSession
	}
}

// Authenticate returns the administrator a fully signed-in session belongs
// to, and the session. It is what the middleware calls on every request to
// the panel's API; a session half way through signing in is no session here.
func (s *Service) Authenticate(ctx context.Context, token string) (*model.AdminUser, *model.AdminUserSession, error) {
	admin, session, state, err := s.Session(ctx, token)
	if err != nil {
		return nil, nil, err
	}

	if state != StateSignedIn {
		return nil, nil, ErrNoSession
	}

	return admin, session, nil
}

// Logout revokes the session the token belongs to. Signing out twice is not
// an error.
func (s *Service) Logout(ctx context.Context, token string, req Request) error {
	if token == "" {
		return nil
	}

	session, err := s.store.SessionByTokenHash(ctx, hashToken(token))
	switch {
	case errors.Is(err, store.ErrNotFound):
		return nil
	case err != nil:
		return err
	}

	if err := s.store.RevokeSession(ctx, session, time.Now()); err != nil {
		return err
	}

	if admin, err := s.store.AdminByID(ctx, session.AdminUserID); err == nil {
		s.record(ctx, &admin.ID, admin.Username, "admin.logout", req, "")
	}

	return nil
}

// record writes one line to the activity log. A failure to write the log must
// not fail the request that caused it, so the error is swallowed on purpose;
// the caller has already done the thing being recorded.
func (s *Service) record(ctx context.Context, adminID *uuid.UUID, actor, action string, req Request, note string) {
	var metadata map[string]any
	if note != "" {
		metadata = map[string]any{"reason": note}
	}

	s.recordWith(ctx, adminID, actor, action, req, metadata)
}

// recordWith is record with metadata of the caller's choosing. Empty values are
// left out.
func (s *Service) recordWith(ctx context.Context, adminID *uuid.UUID, actor, action string, req Request, metadata map[string]any) {
	entry := model.AuditLog{
		AdminUserID: adminID,
		ActorEmail:  actor,
		Action:      action,
		TargetType:  "admin_user",
		IP:          req.IP,
		UserAgent:   req.UserAgent,
	}
	for key, value := range metadata {
		if value != "" && value != nil {
			if entry.Metadata == nil {
				entry.Metadata = map[string]any{}
			}
			entry.Metadata[key] = value
		}
	}
	if adminID != nil {
		entry.TargetID = adminID.String()
	}

	// The action itself has happened; a lost log line must not undo it, but
	// it must not go unnoticed either.
	if err := s.store.WriteAudit(ctx, &entry); err != nil {
		s.log.Error("writing the activity log failed", "error", err, "action", action)
	}
}

// newToken returns a session token: 256 bits of randomness, hex encoded.
func newToken() (string, error) {
	var b [32]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", fmt.Errorf("generate session token: %w", err)
	}
	return hex.EncodeToString(b[:]), nil
}

// hashToken is what gets stored. SHA-256 is right here, unlike for passwords:
// the token is long and random, so it cannot be guessed by brute force.
func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
