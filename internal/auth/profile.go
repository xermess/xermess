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
func (s *Service) UpdateProfile(ctx context.Context, admin *model.AdminUser, profile Profile, currentPassword string) error {
	if profile.Email != admin.Email {
		if err := matches(admin, currentPassword); err != nil {
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
func (s *Service) ChangePassword(ctx context.Context, admin *model.AdminUser, session uuid.UUID, current, next string) error {
	if err := matches(admin, current); err != nil {
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

// matches says whether a password is the administrator's own.
func matches(admin *model.AdminUser, password string) error {
	if password == "" || bcrypt.CompareHashAndPassword([]byte(admin.PasswordHash), []byte(password)) != nil {
		return ErrWrongPassword
	}

	return nil
}
