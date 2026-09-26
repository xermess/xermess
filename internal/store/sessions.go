package store

import (
	"context"
	"time"

	"github.com/google/uuid"

	"loginer/internal/model"
)

// CreateSession starts a session.
func (s *Store) CreateSession(ctx context.Context, session *model.AdminUserSession) error {
	return s.db.WithContext(ctx).Create(session).Error
}

// SessionByTokenHash finds the session a token belongs to. Only the hash is
// ever stored, so this is what the middleware looks up on every request.
func (s *Store) SessionByTokenHash(ctx context.Context, hash string) (*model.AdminUserSession, error) {
	var session model.AdminUserSession
	if err := s.db.WithContext(ctx).Where("token_hash = ?", hash).First(&session).Error; err != nil {
		return nil, translate(err)
	}

	return &session, nil
}

// RevokeSession ends a session. The row stays, so the sessions list can still
// show where someone was signed in.
func (s *Store) RevokeSession(ctx context.Context, session *model.AdminUserSession, at time.Time) error {
	return s.db.WithContext(ctx).Model(session).Update("revoked_at", at).Error
}

// SessionsFor returns an administrator's own sessions, newest first.
func (s *Store) SessionsFor(ctx context.Context, adminID uuid.UUID, limit int) ([]model.AdminUserSession, error) {
	var sessions []model.AdminUserSession
	err := s.db.WithContext(ctx).
		Where("admin_user_id = ?", adminID).
		Order("created_at DESC").
		Limit(limit).
		Find(&sessions).Error

	return sessions, err
}

// RevokeOtherSessionsFor ends every open session an administrator has but
// one: the one they changed their password from, which would otherwise sign
// them out of the page they were using.
func (s *Store) RevokeOtherSessionsFor(ctx context.Context, adminID, keep uuid.UUID, at time.Time) error {
	return s.db.WithContext(ctx).
		Model(&model.AdminUserSession{}).
		Where("admin_user_id = ? AND id <> ? AND revoked_at IS NULL", adminID, keep).
		Update("revoked_at", at).Error
}
