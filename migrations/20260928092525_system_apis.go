package migrations

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/pressly/goose/v3"

	"loginer/internal/model"
)

func init() {
	goose.AddMigrationContext(upSystemAPIs, downSystemAPIs)
}

// upSystemAPIs adds apis.system, which marks the APIs this server is itself:
// the admin API and the account API, which the server makes at startup
// (store.EnsureSystemAPIs) rather than here, since their scopes follow the
// permission catalog from one release to the next.
func upSystemAPIs(_ context.Context, tx *sql.Tx) error {
	db, err := gormTx(tx)
	if err != nil {
		return err
	}

	if db.Migrator().HasColumn(&model.API{}, "System") {
		return nil
	}
	if err := db.Migrator().AddColumn(&model.API{}, "System"); err != nil {
		return fmt.Errorf("add apis.system: %w", err)
	}

	return nil
}

// downSystemAPIs removes the system APIs and the column that marked them.
// Their scopes, and every application's access to them, go with them.
func downSystemAPIs(_ context.Context, tx *sql.Tx) error {
	db, err := gormTx(tx)
	if err != nil {
		return err
	}

	for _, statement := range []string{
		`DELETE FROM user_role_api_scopes WHERE api_scope_id IN
			(SELECT s.id FROM api_scopes s JOIN apis a ON a.id = s.api_id WHERE a.system <> '')`,
		`DELETE FROM application_api_scopes WHERE api_scope_id IN
			(SELECT s.id FROM api_scopes s JOIN apis a ON a.id = s.api_id WHERE a.system <> '')`,
		`DELETE FROM application_apis WHERE api_id IN (SELECT id FROM apis WHERE system <> '')`,
		`DELETE FROM api_scopes WHERE api_id IN (SELECT id FROM apis WHERE system <> '')`,
		`DELETE FROM apis WHERE system <> ''`,
	} {
		if err := db.Exec(statement).Error; err != nil {
			return fmt.Errorf("remove the system apis: %w", err)
		}
	}

	return db.Migrator().DropColumn(&model.API{}, "System")
}
