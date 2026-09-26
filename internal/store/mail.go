package store

import (
	"context"
	"errors"

	"loginer/internal/model"
)

// MailSettings returns how this installation sends email.
//
// There is one row, so it is read by nothing but its age: the oldest is the
// one the first start seeded. A database that holds none has not been
// configured and gets the default, which sends nothing — an installation
// whose row was removed by hand logs its messages rather than failing every
// sign-up.
//
// It is read from the database each time rather than cached: email is sent
// rarely enough that a query costs nothing, and the settings carry a sealed
// password that has no business in a second place.
func (s *Store) MailSettings(ctx context.Context) (*model.MailSettings, error) {
	var settings model.MailSettings

	err := translate(s.db.WithContext(ctx).Order("created_at").First(&settings).Error)
	switch {
	case err == nil:
		return &settings, nil
	case !errors.Is(err, ErrNotFound):
		return nil, err
	}

	settings = model.DefaultMailSettings()

	return &settings, nil
}

// EnsureMailSettings writes the row a fresh installation starts with, from
// the configuration, and leaves an installation that already has one alone:
// after the first start the mail server is the panel's, not the file's.
func (s *Store) EnsureMailSettings(ctx context.Context, settings model.MailSettings) error {
	var count int64
	if err := s.db.WithContext(ctx).Model(&model.MailSettings{}).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	return translate(s.db.WithContext(ctx).Create(&settings).Error)
}

// SaveMailSettings writes the settings back.
func (s *Store) SaveMailSettings(ctx context.Context, settings *model.MailSettings) error {
	return translate(s.db.WithContext(ctx).Save(settings).Error)
}
