package store

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"loginer/internal/model"
)

// User sessions for the Sessions page are never counted or offset-paged: they
// are read newest first by (created_at, id) and continued from the last shown,
// so every page costs the same.

// SessionQuery narrows the list of active sessions.
type SessionQuery struct {
	// Search is the start of an address — "ada" finds ada@example.com — so it
	// can use the prefix index on users.email rather than read every user.
	Search string
	// UserID keeps one user's sessions.
	UserID uuid.UUID
	// After continues a list: the sessions started before this one.
	After uuid.UUID
	Limit int
}

// ActiveSession is a session still signing somebody in, with who they are.
type ActiveSession struct {
	model.UserSession
	Email     string
	FirstName string
	LastName  string
}

// ActiveSessions lists the sessions still signing somebody in, newest first.
func (s *Store) ActiveSessions(ctx context.Context, q SessionQuery, now time.Time) ([]ActiveSession, error) {
	query := s.db.WithContext(ctx).
		Table("user_sessions").
		Select("user_sessions.*, users.email, users.first_name, users.last_name").
		Joins("JOIN users ON users.id = user_sessions.user_id AND users.deleted_at IS NULL").
		Where("user_sessions.deleted_at IS NULL AND user_sessions.revoked_at IS NULL AND user_sessions.expires_at > ?", now)

	if search := model.NormalizeEmail(q.Search); search != "" {
		query = query.Where("users.email LIKE ?", likeEscaper.Replace(search)+"%")
	}
	if q.UserID != uuid.Nil {
		query = query.Where("user_sessions.user_id = ?", q.UserID)
	}
	if q.After != uuid.Nil {
		query = query.Where(
			"(user_sessions.created_at, user_sessions.id) < (SELECT created_at, id FROM user_sessions WHERE id = ?)",
			q.After,
		)
	}

	var sessions []ActiveSession
	err := query.
		Order("user_sessions.created_at DESC, user_sessions.id DESC").
		Limit(q.Limit).
		Scan(&sessions).Error

	return sessions, err
}

// UserSession returns one session by id, whatever its state.
func (s *Store) UserSession(ctx context.Context, id uuid.UUID) (*model.UserSession, error) {
	var session model.UserSession
	if err := s.db.WithContext(ctx).First(&session, "id = ?", id).Error; err != nil {
		return nil, translate(err)
	}

	return &session, nil
}

// SignOutUser ends every session and revokes every refresh token of a user in
// one transaction, so applications are signed out too.
func (s *Store) SignOutUser(ctx context.Context, user uuid.UUID, at time.Time) (sessions, tokens int64, err error) {
	var ended []model.UserSession

	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var err error
		if ended, err = revokeUserSessionsIn(tx, at, "user_id = ? AND expires_at > ?", user, at); err != nil {
			return err
		}
		sessions = int64(len(ended))

		result := tx.Model(&model.RefreshToken{}).
			Where("user_id = ? AND revoked_at IS NULL AND expires_at > ?", user, at).
			Update("revoked_at", at)
		if result.Error != nil {
			return result.Error
		}
		tokens = result.RowsAffected

		return tx.Exec(pendingEmailChangesUsed, map[string]any{"at": at, "user": user}).Error
	})
	if err == nil {
		s.putUserSessions(ctx, ended...)
	}

	return sessions, tokens, err
}
