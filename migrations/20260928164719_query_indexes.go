package migrations

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(upQueryIndexes, downQueryIndexes)
}

// The indexes the hot queries were missing, found by reading their plans on
// a database of a few hundred thousand users: each of these read its whole
// table on every page it serves.
//
//   - The users page sorts by created_at, and search matches anywhere in the
//     record (store.Users), which only a trigram index can serve.
//   - GORM gives a many-to-many table a primary key on both columns, which
//     only serves lookups by the first. Each is also read by its second:
//     users holding a role, role member counts, and the cascades that clear
//     an API's or a role's rows.
//   - An API's activity (store.APIAuditLog) finds entries by their target,
//     and by the API an entry's metadata names.
//
// And one set goes. Every table has an index on deleted_at, from model.Base,
// but nothing is ever soft-deleted — every delete here is a hard one — so the
// column is always null, no query can use the index, and every insert still
// pays to maintain it.
var queryIndexes = []struct{ name, create string }{
	{"idx_users_created", `CREATE INDEX idx_users_created ON users (created_at DESC)`},
	{"idx_user_role_members_role", `CREATE INDEX idx_user_role_members_role ON user_role_members (user_role_id)`},
	{"idx_user_role_inherits_inherited", `CREATE INDEX idx_user_role_inherits_inherited ON user_role_inherits (inherited_role_id)`},
	{"idx_user_role_api_scopes_scope", `CREATE INDEX idx_user_role_api_scopes_scope ON user_role_api_scopes (api_scope_id)`},
	{"idx_application_apis_api", `CREATE INDEX idx_application_apis_api ON application_apis (api_id)`},
	{"idx_application_api_scopes_scope", `CREATE INDEX idx_application_api_scopes_scope ON application_api_scopes (api_scope_id)`},
	{"idx_audit_logs_target", `CREATE INDEX idx_audit_logs_target ON audit_logs (target_type, target_id, created_at DESC)`},
	// Only the few entries that name an API are in it. The predicate is the
	// one store.APIAuditLog writes, which is how the planner knows it may use
	// the index, and it keeps any other metadata from ever being parsed.
	{"idx_audit_logs_api", `CREATE INDEX idx_audit_logs_api ON audit_logs (((metadata::jsonb) ->> 'api_id'))
		WHERE metadata LIKE '%"api_id"%'`},
}

// userSearch is the one expression store.Users searches, with a trigram
// index over it. pg_trgm ships with Postgres and is a trusted extension, so
// the database's owner can create it; where it cannot be installed, search
// still works — it reads the table, as it did before.
const userSearch = `CREATE INDEX idx_users_search ON users USING gin (
	LOWER(email || ' ' || COALESCE(first_name, '') || ' ' || COALESCE(last_name, '') || ' ' || COALESCE(data, ''))
	gin_trgm_ops)`

func upQueryIndexes(ctx context.Context, tx *sql.Tx) error {
	for _, index := range queryIndexes {
		if _, err := tx.ExecContext(ctx, index.create); err != nil {
			return fmt.Errorf("create %s: %w", index.name, err)
		}
	}

	// A savepoint, so an installation without the extension keeps the rest
	// of this migration rather than failing all of it.
	if _, err := tx.ExecContext(ctx, "SAVEPOINT trigram"); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, "CREATE EXTENSION IF NOT EXISTS pg_trgm"); err == nil {
		if _, err := tx.ExecContext(ctx, userSearch); err != nil {
			return fmt.Errorf("create idx_users_search: %w", err)
		}
	} else if _, err := tx.ExecContext(ctx, "ROLLBACK TO SAVEPOINT trigram"); err != nil {
		return err
	}

	// model.Base's index on deleted_at, on every table that has one.
	names, err := deletedAtIndexes(ctx, tx)
	if err != nil {
		return err
	}
	for _, name := range names {
		if _, err := tx.ExecContext(ctx, `DROP INDEX IF EXISTS "`+name+`"`); err != nil {
			return fmt.Errorf("drop %s: %w", name, err)
		}
	}
	if _, err := tx.ExecContext(ctx, "DROP INDEX IF EXISTS idx_audit_logs_target_type"); err != nil {
		return err
	}

	return nil
}

func downQueryIndexes(ctx context.Context, tx *sql.Tx) error {
	names := []string{"idx_users_search"}
	for _, index := range queryIndexes {
		names = append(names, index.name)
	}
	for _, name := range names {
		if _, err := tx.ExecContext(ctx, "DROP INDEX IF EXISTS "+name); err != nil {
			return fmt.Errorf("drop %s: %w", name, err)
		}
	}

	if _, err := tx.ExecContext(ctx, "CREATE INDEX IF NOT EXISTS idx_audit_logs_target_type ON audit_logs (target_type)"); err != nil {
		return err
	}

	rows, err := tx.QueryContext(ctx, `SELECT table_name FROM information_schema.columns
		WHERE table_schema = current_schema() AND column_name = 'deleted_at' ORDER BY table_name`)
	if err != nil {
		return err
	}
	var tables []string
	for rows.Next() {
		var table string
		if err := rows.Scan(&table); err != nil {
			rows.Close()
			return err
		}
		tables = append(tables, table)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return err
	}

	for _, table := range tables {
		create := fmt.Sprintf(`CREATE INDEX IF NOT EXISTS "idx_%s_deleted_at" ON "%s" (deleted_at)`, table, table)
		if _, err := tx.ExecContext(ctx, create); err != nil {
			return fmt.Errorf("restore the deleted_at index on %s: %w", table, err)
		}
	}

	return nil
}

// deletedAtIndexes names the indexes on deleted_at alone.
func deletedAtIndexes(ctx context.Context, tx *sql.Tx) ([]string, error) {
	rows, err := tx.QueryContext(ctx, `SELECT indexname FROM pg_indexes
		WHERE schemaname = current_schema() AND indexname LIKE 'idx\_%\_deleted\_at'
		  AND indexdef LIKE '%(deleted_at)' ORDER BY indexname`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var names []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		names = append(names, name)
	}
	return names, rows.Err()
}
