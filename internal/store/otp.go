package store

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"loginer/internal/cache"
	"loginer/internal/model"
)

// OTPSettings returns how the codes this server emails behave.
//
// As with the mail settings and admin_security: one row, read by its age, and
// the default for a database that holds none — so the pages that ask for a
// code go on working on an installation whose row was removed by hand.
func (s *Store) OTPSettings(ctx context.Context) (*model.OTPSettings, error) {
	settings, err := cached(ctx, s, cache.OTPSettings, "settings", func() (model.OTPSettings, error) {
		var settings model.OTPSettings

		err := translate(s.db.WithContext(ctx).Order("created_at").First(&settings).Error)
		switch {
		case err == nil:
			return settings, nil
		case !errors.Is(err, ErrNotFound):
			return settings, err
		}

		return model.DefaultOTPSettings(), nil
	})
	if err != nil {
		return nil, err
	}

	return &settings, nil
}

// EnsureOTPSettings writes the row a fresh installation starts with and
// leaves one that already has it alone.
func (s *Store) EnsureOTPSettings(ctx context.Context) error {
	var count int64
	if err := s.db.WithContext(ctx).Model(&model.OTPSettings{}).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	settings := model.DefaultOTPSettings()

	return s.forgetting(ctx, translate(s.db.WithContext(ctx).Create(&settings).Error), cache.OTPSettings)
}

// SaveOTPSettings writes the settings back.
func (s *Store) SaveOTPSettings(ctx context.Context, settings *model.OTPSettings) error {
	return s.forgetting(ctx, translate(s.db.WithContext(ctx).Save(settings).Error), cache.OTPSettings)
}

// CreateLoginCode stores a sign-in waiting for an emailed code, and forgets
// any the same user had waiting: asking for a code makes the one before it
// useless, so a message somebody has already read cannot be typed back.
func (s *Store) CreateLoginCode(ctx context.Context, code *model.LoginCode) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		err := tx.Unscoped().
			Where("user_id = ? AND used_at IS NULL", code.UserID).
			Delete(&model.LoginCode{}).Error
		if err != nil {
			return translate(err)
		}

		return translate(tx.Create(code).Error)
	})
}

// LoginCodeByHash finds the sign-in a page's handle names, with the user it
// is for loaded: whoever is typing the code is who the session will be made
// for, and both are needed together every time.
func (s *Store) LoginCodeByHash(ctx context.Context, hash string) (*model.LoginCode, error) {
	var code model.LoginCode

	err := translate(s.db.WithContext(ctx).
		Preload("User").
		Where("handle_hash = ?", hash).
		First(&code).Error)
	if err != nil {
		return nil, err
	}

	return &code, nil
}

// RecordLoginCodeAttempt counts one typed code against the sign-in and
// answers how many have been counted, so the caller can say whether there are
// any guesses left.
//
// Every typed code takes its guess here before it is compared, the right one
// included, and a sign-in with none left gives none: false, and nothing
// counted. The count is added to in the database rather than read and
// written back, and only while it is under `max`, so guesses sent at once
// are each counted and no more than `max` of them are ever compared —
// written back from the row each request loaded, twenty sent together would
// count as one, and all twenty would be checked.
func (s *Store) RecordLoginCodeAttempt(ctx context.Context, code *model.LoginCode, max int) (int, bool, error) {
	var rows []struct {
		Attempts int
	}

	err := s.db.WithContext(ctx).Raw(
		`UPDATE login_codes SET attempts = attempts + 1
		WHERE id = ? AND used_at IS NULL AND attempts < ?
		RETURNING attempts`,
		code.ID, max,
	).Scan(&rows).Error
	if err != nil {
		return 0, false, translate(err)
	}
	if len(rows) == 0 {
		return code.Attempts, false, nil
	}

	code.Attempts = rows[0].Attempts

	return code.Attempts, true, nil
}

// ConsumeLoginCode marks a code used. A code is used once, so the update is
// conditional: two requests typing the right code at the same moment, and
// only one of them makes a session.
func (s *Store) ConsumeLoginCode(ctx context.Context, code *model.LoginCode, now time.Time) (bool, error) {
	result := s.db.WithContext(ctx).
		Model(&model.LoginCode{}).
		Where("id = ? AND used_at IS NULL", code.ID).
		Update("used_at", now)
	if result.Error != nil {
		return false, translate(result.Error)
	}

	if result.RowsAffected == 0 {
		return false, nil
	}

	code.UsedAt = &now

	return true, nil
}

// ResendLoginCode replaces the code a sign-in is waiting for, without
// starting the sign-in again: the handle the page holds stays as it is, and
// the guesses already spent stay spent — asking for another message is not a
// way to start the count over.
func (s *Store) ResendLoginCode(ctx context.Context, code *model.LoginCode, hash string, sentAt, expiresAt time.Time) error {
	err := translate(s.db.WithContext(ctx).Model(code).Updates(map[string]any{
		"code_hash":  hash,
		"sent_at":    sentAt,
		"expires_at": expiresAt,
	}).Error)
	if err != nil {
		return err
	}

	code.CodeHash = hash
	code.SentAt = sentAt
	code.ExpiresAt = expiresAt

	return nil
}
