package migrations

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(upSessions, downSessions)
}

// The indexes the panel's Sessions page reads by (store.ActiveSessions):
// sessions newest first, continued from the last one shown, and users by the
// start of their address. The unique index on users.email cannot answer
// "starts with" unless the database collates as C; text_pattern_ops can,
// whatever it collates as.
var sessionIndexes = []struct{ name, create string }{
	{"idx_user_sessions_recent", "CREATE INDEX IF NOT EXISTS idx_user_sessions_recent ON user_sessions (created_at DESC, id DESC)"},
	{"idx_users_email_prefix", "CREATE INDEX IF NOT EXISTS idx_users_email_prefix ON users (email text_pattern_ops)"},
}

func upSessions(ctx context.Context, tx *sql.Tx) error {
	for _, index := range sessionIndexes {
		if _, err := tx.ExecContext(ctx, index.create); err != nil {
			return fmt.Errorf("create %s: %w", index.name, err)
		}
	}

	return nil
}

func downSessions(ctx context.Context, tx *sql.Tx) error {
	for _, index := range sessionIndexes {
		if _, err := tx.ExecContext(ctx, "DROP INDEX IF EXISTS "+index.name); err != nil {
			return fmt.Errorf("drop %s: %w", index.name, err)
		}
	}

	return nil
}
