package auth

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"loginer/internal/model"
)

// ErrWrongPassword is a wrong current password. Changing the address or
// password requires it, so an unattended session cannot take the account.
var ErrWrongPassword = errors.New("the current password is not right")

// Profile is what an administrator may change about themselves.
type Profile struct {
	FirstName string
	LastName  string
	Email     string
	AvatarURL string
}

// UpdateProfile changes the administrator's name and address; a new address
// needs the current password. A taken address is store.ErrDuplicate.
func (s *Service) UpdateProfile(ctx context.Context, admin *model.Admin, profile Profile, currentPassword string) error {
	if err := s.withPassword(ctx, admin); err != nil {
		return err
	}

	if profile.Email != admin.Email {
		if err := matches(admin, currentPassword); err != nil {
			return err
		}
	}

	admin.FirstName = profile.FirstName
	admin.LastName = profile.LastName
	admin.Email = profile.Email
	admin.Username = profile.Email
	admin.AvatarURL = profile.AvatarURL

	return s.store.SaveOwnAccount(ctx, admin)
}

// ChangePassword sets a new password given the current one and ends every other
// session. A password bcrypt cannot hash is model.ErrPasswordTooLong.
func (s *Service) ChangePassword(ctx context.Context, admin *model.Admin, session uuid.UUID, current, next string) error {
	if err := s.withPassword(ctx, admin); err != nil {
		return err
	}

	if err := matches(admin, current); err != nil {
		return err
	}

	if err := admin.SetPassword(next); err != nil {
		return err
	}
	if err := s.store.SaveOwnAccount(ctx, admin); err != nil {
		return err
	}

	_, err := s.store.RevokeOtherSessionsFor(ctx, admin.ID, session, time.Now())
	return err
}

// withPassword loads the password hash, which cached administrators do not
// carry.
func (s *Service) withPassword(ctx context.Context, admin *model.Admin) error {
	stored, err := s.store.AdminByID(ctx, admin.ID)
	if err != nil {
		return err
	}

	admin.PasswordHash = stored.PasswordHash

	return nil
}

// matches says whether a password is the administrator's own.
func matches(admin *model.Admin, password string) error {
	if password == "" || bcrypt.CompareHashAndPassword([]byte(admin.PasswordHash), []byte(password)) != nil {
		return ErrWrongPassword
	}

	return nil
}
