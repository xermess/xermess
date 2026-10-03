package store

import (
	"context"
	"errors"

	"loginer/internal/cache"
	"loginer/internal/model"
)

// AdminSecurity returns the administrators' sign-in settings, or the default
// when no row exists. A read failure is an error, and internal/auth fails
// closed on it.
func (s *Store) AdminSecurity(ctx context.Context) (*model.AdminSecurity, error) {
	security, err := cached(ctx, s, cache.AdminSecurity, "settings", func() (model.AdminSecurity, error) {
		var security model.AdminSecurity

		err := translate(s.db.WithContext(ctx).Order("created_at").First(&security).Error)
		switch {
		case err == nil:
			return security, nil
		case !errors.Is(err, ErrNotFound):
			return security, err
		}

		return model.DefaultAdminSecurity(), nil
	})
	if err != nil {
		return nil, err
	}

	return &security, nil
}

// EnsureAdminSecurity seeds the row from configuration on a fresh installation
// and leaves an existing one alone.
func (s *Store) EnsureAdminSecurity(ctx context.Context, mfaRequired bool) error {
	var count int64
	if err := s.db.WithContext(ctx).Model(&model.AdminSecurity{}).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	security := model.AdminSecurity{RequireMFA: mfaRequired}

	return s.forgetting(ctx, translate(s.db.WithContext(ctx).Create(&security).Error), cache.AdminSecurity)
}

// SaveAdminSecurity writes the settings back.
func (s *Store) SaveAdminSecurity(ctx context.Context, security *model.AdminSecurity) error {
	return s.forgetting(ctx, translate(s.db.WithContext(ctx).Save(security).Error), cache.AdminSecurity)
}

// AdminMFACounts returns how many administrators exist and how many have a
// confirmed authenticator.
func (s *Store) AdminMFACounts(ctx context.Context) (total, withMFA int64, err error) {
	if err = s.db.WithContext(ctx).Model(&model.Admin{}).Count(&total).Error; err != nil {
		return 0, 0, err
	}

	err = s.db.WithContext(ctx).
		Model(&model.Admin{}).
		Where("id IN (?)", s.db.Model(&model.MFA{}).
			Select("admin_id").
			Where("confirmed_at IS NOT NULL AND deleted_at IS NULL")).
		Count(&withMFA).Error

	return total, withMFA, err
}
