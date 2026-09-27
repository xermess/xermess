package store

import (
	"context"
	"errors"

	"loginer/internal/cache"
	"loginer/internal/model"
)

// AdminSecurity returns how administrators are made to sign in.
//
// A database that holds no row has not been configured, and gets the default
// — which is what the server writes on its first start anyway. A database
// that cannot be read is a different thing, and is an error: the caller
// decides what to do when the answer is unknown, and internal/auth fails
// closed there.
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

// EnsureAdminSecurity writes the row a fresh installation starts with, from
// the configuration, and leaves an installation that already has one alone:
// after the first start the setting is the panel's, not the file's.
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

// AdminMFACounts is how many administrators there are, and how many of them
// have an authenticator confirmed, for the panel to say what requiring one
// would mean.
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
