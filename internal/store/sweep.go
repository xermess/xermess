package store

import (
	"context"
	"log/slog"
	"time"

	"gorm.io/gorm"

	"loginer/internal/cache"
	"loginer/internal/model"
)

// Sweeping removes expired sessions, codes, tokens and old audit lines, so
// tables track live data rather than growing with every sign-in.

// sweepInterval is how often the sweep runs. Nothing it removes can be used
// once it has expired, so how long it lingers is only a matter of space.
const sweepInterval = time.Hour

// sweepBatch keeps each DELETE small so its locks never hold sign-ins up for
// long.
const sweepBatch = 5000

// expiring are the records that end at `expires_at`. A refresh token shares its
// family's expiry, so reuse detection never loses a token it would have caught.
var expiring = []any{
	&model.AuthorizationRequest{},
	&model.AuthorizationCode{},
	&model.RefreshToken{},
	&model.UserSession{},
	&model.PasswordReset{},
	&model.EmailVerification{},
	&model.LoginCode{},
	&model.AdminSession{},
	&model.SocialLogin{},
	&model.SSOLogin{},
}

// Sweep removes records that expired before `now` and, unless `auditBefore` is
// zero, older activity log entries. It returns the count removed.
func (s *Store) Sweep(ctx context.Context, now, auditBefore time.Time) (int64, error) {
	var removed int64

	for _, table := range expiring {
		n, err := s.deleteInBatches(ctx, table, sessionKinds[tableName(s.db, table)], "expires_at < ?", now)
		removed += n
		if err != nil {
			return removed, err
		}
	}

	if !auditBefore.IsZero() {
		n, err := s.deleteInBatches(ctx, &model.AuditLog{}, "", "created_at < ?", auditBefore)
		removed += n
		if err != nil {
			return removed, err
		}
	}

	return removed, nil
}

// sessionKinds names the tables whose rows are also kept in the session
// database, by the kind they are kept under there.
var sessionKinds = map[string]string{
	"user_sessions":  cache.UserSession,
	"admin_sessions": cache.AdminSession,
}

// tableName is the table a model is stored in.
func tableName(db *gorm.DB, model any) string {
	statement := &gorm.Statement{DB: db}
	if err := statement.Parse(model); err != nil {
		return ""
	}
	return statement.Schema.Table
}

// deleteInBatches deletes matching rows sweepBatch at a time. For a session
// table (`kind` set) it also drops them from the cache.
func (s *Store) deleteInBatches(ctx context.Context, table any, kind, where string, arg any) (int64, error) {
	var removed int64

	for {
		batch := s.db.Model(table).Unscoped().Select("id").Where(where, arg).Limit(sweepBatch)

		var (
			deleted int64
			err     error
		)
		if kind == "" {
			result := s.db.WithContext(ctx).Unscoped().Where("id IN (?)", batch).Delete(table)
			deleted, err = result.RowsAffected, result.Error
		} else {
			// The table name is the model's, never anything a request said.
			var hashes []string
			err = s.db.WithContext(ctx).
				Raw("DELETE FROM "+tableName(s.db, table)+" WHERE id IN (?) RETURNING token_hash", batch).
				Scan(&hashes).Error
			deleted = int64(len(hashes))
			s.sessions.DropSessions(ctx, kind, hashes...)
		}
		if err != nil {
			return removed, translate(err)
		}

		removed += deleted
		if deleted < sweepBatch {
			return removed, nil
		}
	}
}

// KeepSwept sweeps at startup and every sweepInterval until ctx ends, keeping
// the log for `retention` (zero keeps all). Concurrent sweeps on several
// servers are harmless.
func (s *Store) KeepSwept(ctx context.Context, retention time.Duration, log *slog.Logger) {
	sweep := func() {
		now := time.Now()

		var auditBefore time.Time
		if retention > 0 {
			auditBefore = now.Add(-retention)
		}

		removed, err := s.Sweep(ctx, now, auditBefore)
		switch {
		case err != nil && ctx.Err() == nil:
			log.Error("sweeping expired records failed", "error", err, "removed", removed)
		case removed > 0:
			log.Info("swept expired records", "removed", removed)
		}
	}

	sweep()

	ticker := time.NewTicker(sweepInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			sweep()
		}
	}
}
