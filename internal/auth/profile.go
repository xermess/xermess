package auth

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"loginer/internal/model"
)

// ErrWrongPassword is a current password that is not theirs. Changing the
// address they sign in with, or the password, takes the one they have: a
// session left open on someone else's screen is not enough to take the
// account over.
var ErrWrongPassword = errors.New("the current password is not right")

// Profile is what an administrator may change about themselves.
type Profile struct {
	FirstName string
	LastName  string
	Email     string
}

// UpdateProfile changes the signed-in administrator's name and address. The
// address is also what they sign in with, so a new one needs their current
// password; a name alone does not. A taken address is store.ErrDuplicate.
func (s *Service) UpdateProfile(
	ctx context.Context,
	admin *model.AdminUser,
	profile Profile,
	currentPassword string,
	req Request,
) error {
	if profile.Email != admin.Email {
		if err := s.confirmPassword(ctx, admin, currentPassword, req); err != nil {
			return err
		}
	}

	admin.FirstName = profile.FirstName
	admin.LastName = profile.LastName
	admin.Email = profile.Email
	admin.Username = profile.Email

	return s.store.SaveOwnAccount(ctx, admin)
}

// ChangePassword sets a new password for the signed-in administrator, given
// the one they have, and ends every session they have open except the one
// the change came from — the one that proved it knew the old password. A
// password bcrypt cannot hash is model.ErrPasswordTooLong.
func (s *Service) ChangePassword(
	ctx context.Context,
	admin *model.AdminUser,
	session uuid.UUID,
	current, next string,
	req Request,
) error {
	if err := s.confirmPassword(ctx, admin, current, req); err != nil {
		return err
	}

	if err := admin.SetPassword(next); err != nil {
		return err
	}
	if err := s.store.SaveOwnAccount(ctx, admin); err != nil {
		return err
	}

	return s.store.RevokeOtherSessionsFor(ctx, admin.ID, session, time.Now())
}

// confirmPassword checks that a password is the administrator's own, and
// counts a wrong one against the account.
//
// The counting is the point. Without it this was the one place an
// administrator's password could be tried without limit — from a session
// already in the panel, which is exactly the position somebody is in who found
// a screen left open, and the password is what they need to make the account
// theirs for good. A wrong one here now costs what a wrong one at the sign-in
// page costs: a step towards the lockout, and a line in the activity log.
//
// MaxFailedLogins of them lock the account, and a locked account's sessions
// stop working with it (model.AdminUser.CanSignIn, Service.Session). So an
// administrator who mistypes their own password five times is shut out of the
// panel for LockoutDuration — the same price the sign-in page charges for the
// same mistake, and the reason somebody holding a stolen session cannot sit
// there guessing.
func (s *Service) confirmPassword(ctx context.Context, admin *model.AdminUser, password string, req Request) error {
	if password != "" && bcrypt.CompareHashAndPassword([]byte(admin.PasswordHash), []byte(password)) == nil {
		return nil
	}

	locked, err := s.store.RecordFailedLogin(ctx, admin, time.Now(), MaxFailedLogins, LockoutDuration)
	if err != nil {
		return err
	}

	reason := "wrong current password"
	if locked {
		reason = "wrong current password; locked after too many attempts"
	}
	s.record(ctx, &admin.ID, admin.Username, "admin.login_failed", req, reason)

	return ErrWrongPassword
}
