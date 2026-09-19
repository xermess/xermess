package store

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"xermess/internal/cache"
	"xermess/internal/model"
)

// SocialProviders returns the configured providers, in the order their
// buttons are shown. `enabledOnly` is what the sign-in pages ask for.
func (s *Store) SocialProviders(ctx context.Context, enabledOnly bool) ([]model.SocialProvider, error) {
	query := s.db.WithContext(ctx).Order("position, created_at")
	if enabledOnly {
		query = query.Where("enabled = ?", true)
	}

	var providers []model.SocialProvider
	err := query.Find(&providers).Error

	return providers, err
}

// SocialButton is an enabled provider as the sign-in pages draw its button.
type SocialButton struct {
	Slug string `json:"slug"`
	Name string `json:"name"`
	Kind string `json:"kind"`
}

// SocialButtons is the enabled providers, in order, as buttons. Every sign-in
// page asks for them, so they are read through the cache — as buttons and
// not as providers, because a provider carries its sealed client secret and
// that has no business in a cache.
func (s *Store) SocialButtons(ctx context.Context) ([]SocialButton, error) {
	return cached(ctx, s, cache.SocialButtons, "enabled", func() ([]SocialButton, error) {
		providers, err := s.SocialProviders(ctx, true)
		if err != nil {
			return nil, err
		}

		buttons := make([]SocialButton, 0, len(providers))
		for _, provider := range providers {
			buttons = append(buttons, SocialButton{Slug: provider.Slug, Name: provider.Name, Kind: string(provider.Kind)})
		}

		return buttons, nil
	})
}

// SocialProvider returns one provider by id.
func (s *Store) SocialProvider(ctx context.Context, id uuid.UUID) (*model.SocialProvider, error) {
	var provider model.SocialProvider
	if err := s.db.WithContext(ctx).First(&provider, "id = ?", id).Error; err != nil {
		return nil, translate(err)
	}

	return &provider, nil
}

// SocialProviderBySlug returns the provider a callback address names.
func (s *Store) SocialProviderBySlug(ctx context.Context, slug string) (*model.SocialProvider, error) {
	var provider model.SocialProvider
	if err := s.db.WithContext(ctx).First(&provider, "slug = ?", slug).Error; err != nil {
		return nil, translate(err)
	}

	return &provider, nil
}

// CreateSocialProvider adds a provider.
func (s *Store) CreateSocialProvider(ctx context.Context, provider *model.SocialProvider) error {
	if err := translate(s.db.WithContext(ctx).Create(provider).Error); err != nil {
		return err
	}

	s.forget(ctx, cache.SocialButtons)

	return nil
}

// SaveSocialProvider writes a provider back.
func (s *Store) SaveSocialProvider(ctx context.Context, provider *model.SocialProvider) error {
	if err := translate(s.db.WithContext(ctx).Save(provider).Error); err != nil {
		return err
	}

	s.forget(ctx, cache.SocialButtons)

	return nil
}

// DeleteSocialProvider removes a provider, and with it every identity held at
// it: without the provider there is no way to check those identities again,
// and the accounts themselves stay.
func (s *Store) DeleteSocialProvider(ctx context.Context, provider *model.SocialProvider) error {
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		err := tx.Unscoped().Where("provider_id = ?", provider.ID).Delete(&model.UserIdentity{}).Error
		if err != nil {
			return err
		}

		return tx.Unscoped().Delete(provider).Error
	})
	if err != nil {
		return err
	}

	s.forget(ctx, cache.SocialButtons)

	return nil
}

// NextSocialPosition is where a new provider's button goes: after the last.
func (s *Store) NextSocialPosition(ctx context.Context) int {
	var last model.SocialProvider
	if err := s.db.WithContext(ctx).Order("position DESC").First(&last).Error; err != nil {
		return 1
	}

	return last.Position + 1
}

// SocialIdentityCounts is how many users hold an identity at each provider,
// by provider id, for the panel's list.
func (s *Store) SocialIdentityCounts(ctx context.Context) (map[uuid.UUID]int64, error) {
	var rows []struct {
		ProviderID uuid.UUID
		Count      int64
	}

	err := s.db.WithContext(ctx).
		Model(&model.UserIdentity{}).
		Select("provider_id, count(*) AS count").
		Group("provider_id").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	counts := make(map[uuid.UUID]int64, len(rows))
	for _, row := range rows {
		counts[row.ProviderID] = row.Count
	}

	return counts, nil
}

// UserIdentity returns the identity a provider's subject belongs to.
func (s *Store) UserIdentity(ctx context.Context, providerID uuid.UUID, subject string) (*model.UserIdentity, error) {
	var identity model.UserIdentity
	err := s.db.WithContext(ctx).
		First(&identity, "provider_id = ? AND subject = ?", providerID, subject).Error
	if err != nil {
		return nil, translate(err)
	}

	return &identity, nil
}

// UserIdentities returns every provider one user can sign in with.
func (s *Store) UserIdentities(ctx context.Context, userID uuid.UUID) ([]model.UserIdentity, error) {
	var identities []model.UserIdentity
	err := s.db.WithContext(ctx).
		Preload("Provider").
		Where("user_id = ?", userID).
		Order("created_at").
		Find(&identities).Error

	return identities, err
}

// CreateUserIdentity connects an account at a provider to a user.
func (s *Store) CreateUserIdentity(ctx context.Context, identity *model.UserIdentity) error {
	return translate(s.db.WithContext(ctx).Create(identity).Error)
}

// MarkIdentityUsed notes that a user signed in with an identity, and keeps
// the address the provider gave this time.
func (s *Store) MarkIdentityUsed(ctx context.Context, identity *model.UserIdentity, email string, at time.Time) error {
	identity.LastLoginAt = &at
	identity.Email = email

	return translate(s.db.WithContext(ctx).
		Model(identity).
		Updates(map[string]any{"last_login_at": at, "email": email}).Error)
}

// DeleteUserIdentity disconnects a provider from a user's account.
func (s *Store) DeleteUserIdentity(ctx context.Context, identity *model.UserIdentity) error {
	return s.db.WithContext(ctx).Unscoped().Delete(identity).Error
}

// CreateSocialLogin records a sign-in sent to a provider.
func (s *Store) CreateSocialLogin(ctx context.Context, login *model.SocialLogin) error {
	return translate(s.db.WithContext(ctx).Create(login).Error)
}

// TakeSocialLogin returns the sign-in a state belongs to and deletes it, so
// an answer from a provider is only ever accepted once. A state that is not
// there, or has expired, is ErrNotFound.
func (s *Store) TakeSocialLogin(ctx context.Context, stateHash string, now time.Time) (*model.SocialLogin, error) {
	var login model.SocialLogin

	err := s.db.WithContext(ctx).
		Clauses(clause.Returning{}).
		Where("state_hash = ? AND expires_at > ?", stateHash, now).
		Delete(&login).Error
	if err != nil {
		return nil, translate(err)
	}

	if login.ID == uuid.Nil {
		return nil, ErrNotFound
	}

	return &login, nil
}

// PurgeSocialLogins removes the sign-ins nobody came back from.
func (s *Store) PurgeSocialLogins(ctx context.Context, now time.Time) error {
	return s.db.WithContext(ctx).
		Unscoped().
		Where("expires_at <= ?", now).
		Delete(&model.SocialLogin{}).Error
}

// SocialAccountsFor is the providers each of these users signs in with, by
// user id. It is one query for a whole page of users rather than one each.
func (s *Store) SocialAccountsFor(ctx context.Context, userIDs []uuid.UUID) (map[uuid.UUID][]model.SocialAccount, error) {
	accounts := map[uuid.UUID][]model.SocialAccount{}
	if len(userIDs) == 0 {
		return accounts, nil
	}

	var rows []struct {
		ID          uuid.UUID
		UserID      uuid.UUID
		Provider    string
		Slug        string
		Kind        model.SocialKind
		Email       string
		ConnectedAt time.Time
		LastLoginAt *time.Time
	}

	err := s.db.WithContext(ctx).
		Table("user_identities AS i").
		Select(`i.id, i.user_id, p.name AS provider, p.slug, p.kind,
			i.email, i.created_at AS connected_at, i.last_login_at`).
		Joins("JOIN social_providers p ON p.id = i.provider_id AND p.deleted_at IS NULL").
		Where("i.user_id IN ? AND i.deleted_at IS NULL", userIDs).
		Order("p.position, i.created_at").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	for _, row := range rows {
		accounts[row.UserID] = append(accounts[row.UserID], model.SocialAccount{
			ID:          row.ID,
			Provider:    row.Provider,
			Slug:        row.Slug,
			Kind:        row.Kind,
			Email:       row.Email,
			ConnectedAt: row.ConnectedAt,
			LastLoginAt: row.LastLoginAt,
		})
	}

	return accounts, nil
}

// UserIdentityByID returns one identity, for disconnecting it.
func (s *Store) UserIdentityByID(ctx context.Context, id uuid.UUID) (*model.UserIdentity, error) {
	var identity model.UserIdentity
	if err := s.db.WithContext(ctx).First(&identity, "id = ?", id).Error; err != nil {
		return nil, translate(err)
	}

	return &identity, nil
}
