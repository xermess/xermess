package migrations

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/pressly/goose/v3"

	"xermess/internal/model"
)

func init() {
	goose.AddMigrationContext(upScale, downScale)
}

// sweptByExpiry are the tables the sweep (store.Sweep) deletes from by
// expires_at that had no index on it, so each sweep would have read the
// whole table to find what had expired.
var sweptByExpiry = []any{
	&model.AuthorizationCode{},
	&model.RefreshToken{},
	&model.UserSession{},
	&model.PasswordReset{},
}

// upScale makes a database with millions of users behave like one with a
// hundred: signing in finds the user by the unique index on their address
// rather than by reading every user, and sweeping expired records finds them
// by index.
//
// Addresses become lower case, as the model now stores them
// (model.NormalizeEmail). Two users whose addresses differ only in case
// would become one address the unique index refuses twice; the migration
// stops and names them rather than choosing which account to keep.
func upScale(_ context.Context, tx *sql.Tx) error {
	rows, err := tx.Query(`
		SELECT lower(trim(email)) FROM users
		GROUP BY 1 HAVING count(*) > 1
		ORDER BY 1 LIMIT 20`)
	if err != nil {
		return fmt.Errorf("look for addresses that differ only in case: %w", err)
	}

	var clashes []string
	for rows.Next() {
		var email string
		if err := rows.Scan(&email); err != nil {
			rows.Close()
			return err
		}
		clashes = append(clashes, email)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return err
	}

	if len(clashes) > 0 {
		return fmt.Errorf("these addresses belong to more than one user, differing only in case: %s; "+
			"merge or rename those users, then migrate again", strings.Join(clashes, ", "))
	}

	if _, err := tx.Exec(`UPDATE users SET email = lower(trim(email)) WHERE email <> lower(trim(email))`); err != nil {
		return fmt.Errorf("lower-case the users' addresses: %w", err)
	}

	db, err := gormTx(tx)
	if err != nil {
		return err
	}

	for _, table := range sweptByExpiry {
		if db.Migrator().HasIndex(table, "ExpiresAt") {
			continue
		}
		if err := db.Migrator().CreateIndex(table, "ExpiresAt"); err != nil {
			return fmt.Errorf("index expires_at: %w", err)
		}
	}

	return nil
}

// downScale drops the indexes again. The addresses stay lower case: which
// letters were capitals is not kept anywhere, and nothing depends on them.
func downScale(_ context.Context, tx *sql.Tx) error {
	db, err := gormTx(tx)
	if err != nil {
		return err
	}

	for _, table := range sweptByExpiry {
		if !db.Migrator().HasIndex(table, "ExpiresAt") {
			continue
		}
		if err := db.Migrator().DropIndex(table, "ExpiresAt"); err != nil {
			return fmt.Errorf("drop the expires_at index: %w", err)
		}
	}

	return nil
}
