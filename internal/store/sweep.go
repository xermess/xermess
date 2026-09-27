package store

import (
	"context"
	"log/slog"
	"time"

	"gorm.io/gorm"

	"loginer/internal/cache"
	"loginer/internal/model"
)

// What the server stores and stops needing: every sign-in leaves a session,
// a code, tokens and a line in the activity log behind, and nothing else
// ever deletes them. Swept, the tables hold what is still alive rather than
// everything there has ever been — which at millions of users is the
// difference between tables that stay the size of the user base and tables
// that grow by it every month.

// sweepInterval is how often the sweep runs. Nothing it removes can be used
// once it has expired, so how long it lingers is only a matter of space.
const sweepInterval = time.Hour

// sweepBatch is how many rows one DELETE removes. Deleting millions of rows
// in one statement holds its locks until it is done; in batches, the
// sign-ins writing the same tables are never kept waiting for long.
const sweepBatch = 5000

// expiring are the records that end at `expires_at`: sign-in requests and
// codes, tokens, sessions, reset and verification links, and the social and SSO sign-ins
// nobody came back from. A refresh token shares its family's expiry, so one
// is only ever removed once its whole family can no longer be used — reuse
// detection never loses a token it would have caught.
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

// Sweep removes every record that expired before `now` and, when
// `auditBefore` is not zero, the activity log entries older than it. It
// answers how many rows it removed.
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

// deleteInBatches deletes the rows of a table matching `where`, sweepBatch at
// a time, until there are none left. For a table of sessions (`kind` is not
// empty), the sessions deleted are removed from the session database too.
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

// KeepSwept runs Sweep once at startup and then every sweepInterval until ctx
// ends, keeping the activity log for `retention` (zero keeps all of it). main
// runs it for as long as the server serves. Each server process sweeps on its
// own; two sweeping at once delete the same rows, which is harmless.
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
