package store

import (
	"context"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"xermess/internal/model"
)

// SSOConnections returns every connection, by name.
func (s *Store) SSOConnections(ctx context.Context) ([]model.SSOConnection, error) {
	var connections []model.SSOConnection
	err := s.db.WithContext(ctx).Order("name").Find(&connections).Error

	return connections, err
}

// SSOConnection returns one connection by id.
func (s *Store) SSOConnection(ctx context.Context, id uuid.UUID) (*model.SSOConnection, error) {
	var connection model.SSOConnection
	if err := s.db.WithContext(ctx).First(&connection, "id = ?", id).Error; err != nil {
		return nil, translate(err)
	}

	return &connection, nil
}

// SSOConnectionBySlug returns the connection an address names.
func (s *Store) SSOConnectionBySlug(ctx context.Context, slug string) (*model.SSOConnection, error) {
	var connection model.SSOConnection
	if err := s.db.WithContext(ctx).First(&connection, "slug = ?", slug).Error; err != nil {
		return nil, translate(err)
	}

	return &connection, nil
}

// SSOConnectionForEmail returns the enabled connection that owns an address's
// domain, or ErrNotFound. A domain belongs to one connection at most, which
// the panel keeps true (SSODomainsTaken).
func (s *Store) SSOConnectionForEmail(ctx context.Context, email string) (*model.SSOConnection, error) {
	_, domain, ok := strings.Cut(strings.ToLower(strings.TrimSpace(email)), "@")
	if !ok || domain == "" {
		return nil, ErrNotFound
	}

	var connection model.SSOConnection
	err := s.db.WithContext(ctx).
		Where("enabled = ? AND jsonb_exists(domains, ?)", true, domain).
		First(&connection).Error
	if err != nil {
		return nil, translate(err)
	}

	return &connection, nil
}

// SSOButtons are the enabled connections that put a button on the sign-in
// page, by name.
func (s *Store) SSOButtons(ctx context.Context) ([]model.SSOConnection, error) {
	var connections []model.SSOConnection
	err := s.db.WithContext(ctx).
		Where("enabled = ? AND show_on_login = ?", true, true).
		Order("name").
		Find(&connections).Error

	return connections, err
}

// AnySSOConnectionEnabled reports whether an address can lead to a
// connection — an enabled one with domains — so the sign-in page offers "Sign
// in with SSO" only where it leads somewhere.
func (s *Store) AnySSOConnectionEnabled(ctx context.Context) (bool, error) {
	var count int64
	err := s.db.WithContext(ctx).Model(&model.SSOConnection{}).
		Where("enabled = ? AND jsonb_array_length(domains) > 0", true).
		Limit(1).Count(&count).Error

	return count > 0, err
}

// SSODomainsTaken is which of these domains another connection already has.
func (s *Store) SSODomainsTaken(ctx context.Context, domains []string, except uuid.UUID) ([]string, error) {
	connections, err := s.SSOConnections(ctx)
	if err != nil {
		return nil, err
	}

	var taken []string
	for _, connection := range connections {
		if connection.ID == except {
			continue
		}
		for _, domain := range domains {
			if slices.Contains(connection.Domains, domain) && !slices.Contains(taken, domain) {
				taken = append(taken, domain)
			}
		}
	}

	return taken, nil
}

// CreateSSOConnection adds a connection.
func (s *Store) CreateSSOConnection(ctx context.Context, connection *model.SSOConnection) error {
	return translate(s.db.WithContext(ctx).Create(connection).Error)
}

// SaveSSOConnection writes a connection back. `newProvider` drops the
// identities held at it as well, in the same transaction: a subject is only
// unique within the provider that issued it, so one from a provider the
// connection has been pointed away from could name somebody else at the new
// one. Their accounts stay, and link again by address on the next sign-in.
func (s *Store) SaveSSOConnection(ctx context.Context, connection *model.SSOConnection, newProvider bool) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if newProvider {
			err := tx.Unscoped().Where("connection_id = ?", connection.ID).Delete(&model.SSOIdentity{}).Error
			if err != nil {
				return translate(err)
			}
		}

		return translate(tx.Save(connection).Error)
	})
}

// DeleteSSOConnection removes a connection, the identities held at it and any
// sign-in to it under way. The accounts themselves stay: they can still sign
// in however else they can, or be given a password.
func (s *Store) DeleteSSOConnection(ctx context.Context, connection *model.SSOConnection) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, child := range []any{&model.SSOIdentity{}, &model.SSOLogin{}} {
			if err := tx.Unscoped().Where("connection_id = ?", connection.ID).Delete(child).Error; err != nil {
				return translate(err)
			}
		}

		return translate(tx.Unscoped().Delete(connection).Error)
	})
}

// SSOIdentityCounts is how many people have signed in through each
// connection, by connection id.
func (s *Store) SSOIdentityCounts(ctx context.Context) (map[uuid.UUID]int64, error) {
	var rows []struct {
		ConnectionID uuid.UUID
		Count        int64
	}

	err := s.db.WithContext(ctx).
		Model(&model.SSOIdentity{}).
		Select("connection_id, COUNT(*) AS count").
		Group("connection_id").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	out := make(map[uuid.UUID]int64, len(rows))
	for _, row := range rows {
		out[row.ConnectionID] = row.Count
	}

	return out, nil
}

// SSOIdentity returns the identity a connection's subject is, or ErrNotFound.
func (s *Store) SSOIdentity(ctx context.Context, connection uuid.UUID, subject string) (*model.SSOIdentity, error) {
	var identity model.SSOIdentity
	err := s.db.WithContext(ctx).First(&identity, "connection_id = ? AND subject = ?", connection, subject).Error
	if err != nil {
		return nil, translate(err)
	}

	return &identity, nil
}

// CreateSSOIdentity connects a person at a connection's provider to a user.
func (s *Store) CreateSSOIdentity(ctx context.Context, identity *model.SSOIdentity) error {
	return translate(s.db.WithContext(ctx).Create(identity).Error)
}

// MarkSSOIdentityUsed notes a sign-in through an identity, and keeps the
// address the provider gave this time.
func (s *Store) MarkSSOIdentityUsed(ctx context.Context, identity *model.SSOIdentity, email string, at time.Time) error {
	return translate(s.db.WithContext(ctx).
		Model(identity).
		Updates(map[string]any{"last_login_at": at, "email": email}).Error)
}

// CreateSSOLogin records a sign-in sent to a connection's provider. The ones
// nobody came back from are swept once they expire (Sweep).
func (s *Store) CreateSSOLogin(ctx context.Context, login *model.SSOLogin) error {
	return translate(s.db.WithContext(ctx).Create(login).Error)
}

// TakeSSOLogin returns the sign-in a state belongs to and deletes it, so a
// provider's answer is only ever accepted once. A state that is not there, or
// has expired, is ErrNotFound.
func (s *Store) TakeSSOLogin(ctx context.Context, stateHash string, now time.Time) (*model.SSOLogin, error) {
	var login model.SSOLogin

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
