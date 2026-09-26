package oidc

import (
	"context"
	"errors"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"loginer/internal/model"
	"loginer/internal/store"
)

// What a signed-in user can do about their own account, in the id app: their
// name, their password, where they are signed in, and which applications can
// still act for them. Every method takes the session the request carries, so
// a user only ever reaches their own account.

// ErrWrongPassword is a current password that does not match, when changing
// it. It counts towards the lockout, as a wrong password at sign-in does.
var ErrWrongPassword = problem("wrong_password", "the current password is wrong")

// confirmPassword checks a password somebody typed to prove an account is
// theirs. It guards the two things a session on its own should not be enough
// for — replacing the password, and moving the address the account signs in
// with — because both of them are how an account is taken over for good, and a
// session may be one left open on a screen somebody walked away from.
//
// A wrong password counts towards the lockout, exactly as it does at the
// sign-in page: otherwise this would be somewhere to try passwords without a
// limit, with the address already known.
//
// An account with no password cannot prove itself this way, so it can do
// neither. That is not a hole but the same answer the sign-in page gives it:
// whoever holds such an account arrived through a provider or a reset link, and
// their address is an administrator's to change (internal/api/users).
func (s *Service) confirmPassword(ctx context.Context, user *model.User, password string, client Client) error {
	if user.PasswordHash != "" && bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)) == nil {
		return nil
	}

	if _, err := s.store.RecordUserFailedLogin(ctx, user, s.now(), maxFailedLogins, lockoutDuration); err != nil {
		return err
	}
	s.record(ctx, user, user.Email, "user.login_failed", client, map[string]any{"reason": "wrong current password"})

	return ErrWrongPassword
}

// ErrNotYours is a session or application that does not belong to the signed-in
// user, or does not exist: the two are one answer.
var ErrNotYours = problem("not_yours", "not the user's")

// ErrPasswordUnchanged is a new password that is the current one.
var ErrPasswordUnchanged = problem("password_unchanged", "the new password is the current one")

// ErrCurrentSession is ending the session the request itself carries, which
// is signing out.
var ErrCurrentSession = problem("current_session", "that is the session in use")

// Profile is what a user may change about themselves. The email is the account,
// so it is not among them.
type Profile struct {
	FirstName string
	LastName  string
}

// UpdateProfile saves a user's name.
func (s *Service) UpdateProfile(ctx context.Context, session *Session, p Profile, client Client) (*model.User, error) {
	user := session.User
	user.FirstName = strings.TrimSpace(p.FirstName)
	user.LastName = strings.TrimSpace(p.LastName)

	if err := s.store.SaveUser(ctx, user); err != nil {
		return nil, err
	}

	s.record(ctx, user, user.Email, "user.profile_updated", client, nil)

	return user, nil
}

// ChangePassword replaces a user's password after checking the current one.
// Every other session the user has ends, and every refresh token applications
// hold for them is revoked: whoever might have known the old password is
// signed out everywhere but here.
func (s *Service) ChangePassword(ctx context.Context, session *Session, current, next string, client Client) error {
	user := session.User
	now := s.now()

	if err := s.confirmPassword(ctx, user, current, client); err != nil {
		return err
	}

	if current == next {
		return ErrPasswordUnchanged
	}

	if err := user.SetPassword(next); errors.Is(err, model.ErrPasswordTooLong) {
		return ErrPasswordTooLong
	} else if err != nil {
		return err
	}
	user.IsTemporaryPassword = false

	if err := s.store.ChangeUserPassword(ctx, user, session.Record.ID, now); err != nil {
		return err
	}

	s.record(ctx, user, user.Email, "user.password_changed_self", client, nil)

	return nil
}

// SessionInfo is one browser a user is signed in on.
type SessionInfo struct {
	ID        uuid.UUID `json:"id"`
	IP        string    `json:"ip"`
	UserAgent string    `json:"user_agent"`
	SignedIn  time.Time `json:"signed_in_at"`
	ExpiresAt time.Time `json:"expires_at"`
	// Current marks the session the request itself carries.
	Current bool `json:"current"`
}

// Sessions lists the user's active sessions, newest first.
func (s *Service) Sessions(ctx context.Context, session *Session) ([]SessionInfo, error) {
	records, err := s.store.ActiveUserSessions(ctx, session.User.ID, s.now())
	if err != nil {
		return nil, err
	}

	out := make([]SessionInfo, 0, len(records))
	for _, record := range records {
		out = append(out, SessionInfo{
			ID:        record.ID,
			IP:        record.IP,
			UserAgent: record.UserAgent,
			SignedIn:  record.AuthTime,
			ExpiresAt: record.ExpiresAt,
			Current:   record.ID == session.Record.ID,
		})
	}

	return out, nil
}

// EndSession signs one of the user's other browsers out. The current one is
// ended by signing out, which also clears its cookie.
func (s *Service) EndSession(ctx context.Context, session *Session, id uuid.UUID, client Client) error {
	if id == session.Record.ID {
		return ErrCurrentSession
	}

	records, err := s.store.ActiveUserSessions(ctx, session.User.ID, s.now())
	if err != nil {
		return err
	}

	if !slices.ContainsFunc(records, func(r model.UserSession) bool { return r.ID == id }) {
		return ErrNotYours
	}

	if err := s.store.RevokeUserSession(ctx, id, s.now()); err != nil {
		return err
	}

	s.record(ctx, session.User, session.User.Email, "user.session_revoked", client, nil)

	return nil
}

// ConnectedApplication is an application that can still act for the user:
// one holding a refresh token for them.
type ConnectedApplication struct {
	ClientID   string    `json:"client_id"`
	Name       string    `json:"name"`
	LogoURI    string    `json:"logo_uri"`
	ClientURI  string    `json:"client_uri"`
	Scopes     []string  `json:"scopes"`
	Authorized time.Time `json:"authorized_at"`
	LastUsed   time.Time `json:"last_used_at"`
}

// ConnectedApplications lists the applications holding refresh tokens for the
// user, most recently used first.
func (s *Service) ConnectedApplications(ctx context.Context, session *Session) ([]ConnectedApplication, error) {
	grants, err := s.store.UserGrants(ctx, session.User.ID, s.now())
	if err != nil {
		return nil, err
	}

	out := make([]ConnectedApplication, 0, len(grants))
	for _, grant := range grants {
		scopes := []string{}
		for _, scope := range strings.Fields(strings.Join(grant.Scopes, " ")) {
			if !slices.Contains(scopes, scope) {
				scopes = append(scopes, scope)
			}
		}

		out = append(out, ConnectedApplication{
			ClientID:   grant.Application.ClientID,
			Name:       grant.Application.Name,
			LogoURI:    grant.Application.LogoURI,
			ClientURI:  grant.Application.ClientURI,
			Scopes:     scopes,
			Authorized: grant.FirstIssued,
			LastUsed:   grant.LastIssued,
		})
	}

	return out, nil
}

// DisconnectApplication revokes every refresh token an application holds for
// the user. Access tokens it already has keep working until they expire.
func (s *Service) DisconnectApplication(ctx context.Context, session *Session, clientID string, client Client) error {
	app, err := s.store.ApplicationByClientID(ctx, clientID)
	if errors.Is(err, store.ErrNotFound) {
		return ErrNotYours
	}
	if err != nil {
		return err
	}

	revoked, err := s.store.RevokeUserGrant(ctx, session.User.ID, app.ID, s.now())
	if err != nil {
		return err
	}
	if revoked == 0 {
		return ErrNotYours
	}

	s.record(ctx, session.User, session.User.Email, "user.application_disconnected", client, map[string]any{"application": app.Name})

	return nil
}
