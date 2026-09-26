package store

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"loginer/internal/model"
)

// OTPSettings returns how the codes this server emails behave.
//
// As with the mail settings and admin_security: one row, read by its age, and
// the default for a database that holds none — so the pages that ask for a
// code go on working on an installation whose row was removed by hand.
func (s *Store) OTPSettings(ctx context.Context) (*model.OTPSettings, error) {
	var settings model.OTPSettings

	err := translate(s.db.WithContext(ctx).Order("created_at").First(&settings).Error)
	switch {
	case err == nil:
		return &settings, nil
	case !errors.Is(err, ErrNotFound):
		return nil, err
	}

	settings = model.DefaultOTPSettings()

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

	return translate(s.db.WithContext(ctx).Create(&settings).Error)
}

// SaveOTPSettings writes the settings back.
func (s *Store) SaveOTPSettings(ctx context.Context, settings *model.OTPSettings) error {
	return translate(s.db.WithContext(ctx).Save(settings).Error)
}

// CreateLoginCode stores a sign-in waiting for an emailed code, and forgets
// any the same user had waiting: asking for a code makes the one before it
// useless, so a message somebody has already read cannot be typed back.
func (s *Store) CreateLoginCode(ctx context.Context, code *model.LoginCode) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		err := tx.Unscoped().
			Where("user_id = ? AND consumed_at IS NULL", code.UserID).
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

// ClaimLoginCodeAttempt takes one of the guesses a waiting sign-in has left
// and answers how many have now been taken. `ok` is false when there were
// none left, which is the whole point of the method: the guess is claimed
// before the code is looked at, so a code cannot be guessed more times than
// the settings allow by sending the guesses at the same moment.
//
// Reading the count and writing it back would not hold. Ten requests that
// read `attempts` as 2 all write 3, and every one of them then compares a
// code against a sign-in that was supposed to have two guesses left. Here the
// count is raised by the database, and the row that comes back is the one
// this request is entitled to — the same reasoning as RecordUserFailedLogin.
//
// A code already spent is not excluded: raising the count on a sign-in that
// has been used changes nothing, since ConsumeLoginCode refuses to use it
// twice. Leaving it out means a false answer has exactly one meaning — the
// guesses are gone.
func (s *Store) ClaimLoginCodeAttempt(ctx context.Context, code *model.LoginCode, maxAttempts int) (int, bool, error) {
	var row struct {
		Attempts int
	}

	result := s.db.WithContext(ctx).Raw(`
		UPDATE login_codes SET attempts = attempts + 1
		WHERE id = @id AND attempts < @max
		RETURNING attempts`,
		map[string]any{"id": code.ID, "max": maxAttempts},
	).Scan(&row)
	if result.Error != nil {
		return 0, false, translate(result.Error)
	}
	if result.RowsAffected == 0 {
		return code.Attempts, false, nil
	}

	code.Attempts = row.Attempts

	return row.Attempts, true, nil
}

// ConsumeLoginCode marks a code used. A code is used once, so the update is
// conditional: two requests typing the right code at the same moment, and
// only one of them makes a session.
func (s *Store) ConsumeLoginCode(ctx context.Context, code *model.LoginCode, now time.Time) (bool, error) {
	result := s.db.WithContext(ctx).
		Model(&model.LoginCode{}).
		Where("id = ? AND consumed_at IS NULL", code.ID).
		Update("consumed_at", now)
	if result.Error != nil {
		return false, translate(result.Error)
	}

	if result.RowsAffected == 0 {
		return false, nil
	}

	code.ConsumedAt = &now

	return true, nil
}

// ResendLoginCode replaces the code a sign-in is waiting for, without
// starting the sign-in again: the handle the page holds stays as it is, and
// the guesses already spent stay spent — asking for another message is not a
// way to start the count over.
//
// `expires_at` is not among the columns written, and that is the point: the
// sign-in keeps the deadline it was given when it was held, so no number of
// messages extends it.
func (s *Store) ResendLoginCode(ctx context.Context, code *model.LoginCode, hash string, sentAt time.Time) error {
	err := translate(s.db.WithContext(ctx).Model(code).Updates(map[string]any{
		"code_hash": hash,
		"sent_at":   sentAt,
	}).Error)
	if err != nil {
		return err
	}

	code.CodeHash = hash
	code.SentAt = sentAt

	return nil
}
