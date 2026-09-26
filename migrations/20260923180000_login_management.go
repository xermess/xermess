package migrations

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/pressly/goose/v3"

	"loginer/internal/model"
)

func init() {
	goose.AddMigrationContext(upLoginManagement, downLoginManagement)
}

// loginSwitches are the columns a login flow gained: what a sign-in through
// it allows, beyond the steps it is made of, and what the flows that already
// exist get for it.
//
// Those flows keep behaving exactly as they did, which for three of the four
// means off. The exception is allow_sign_in: every flow there is signs people
// in today.
var loginSwitches = []struct{ column, existing string }{
	{"allow_sign_in", "true"},
	{"allow_remember_me", "false"},
	{"verify_email_on_register", "false"},
	{"allow_email_change", "false"},
}

// upLoginManagement adds them, and the address a verification link may be a
// change to (model.EmailVerification.NewEmail).
//
// The switches are added by hand rather than by AutoMigrate: Postgres refuses
// a not-null column with no default on a table that has rows, and login_flows
// has held the default flow since the first migration. Each column is added
// saying what those rows get, and its default is dropped again so GORM cannot
// quietly store true for a flow somebody meant to close
// (model_test.TestNoBoolDefaultsToTrue).
func upLoginManagement(_ context.Context, tx *sql.Tx) error {
	db, err := gormTx(tx)
	if err != nil {
		return err
	}

	for _, switched := range loginSwitches {
		add := fmt.Sprintf(
			`ALTER TABLE login_flows ADD COLUMN IF NOT EXISTS %s boolean NOT NULL DEFAULT %s`,
			switched.column, switched.existing,
		)
		if _, err := tx.Exec(add); err != nil {
			return fmt.Errorf("add %s: %w", switched.column, err)
		}

		drop := fmt.Sprintf(
			`ALTER TABLE login_flows ALTER COLUMN %s DROP DEFAULT`,
			switched.column,
		)
		if _, err := tx.Exec(drop); err != nil {
			return fmt.Errorf("drop the default on %s: %w", switched.column, err)
		}
	}

	// The rest of what the two structs say, which is new_email on the
	// verifications now the switches are in place.
	if err := db.AutoMigrate(&model.LoginFlow{}, &model.EmailVerification{}); err != nil {
		return fmt.Errorf("add the login flow switches: %w", err)
	}

	return nil
}

// downLoginManagement drops them again. What is lost is which flows were
// closed and what each allowed — the flows themselves stay, and behave as
// they did before the columns existed.
func downLoginManagement(_ context.Context, tx *sql.Tx) error {
	db, err := gormTx(tx)
	if err != nil {
		return err
	}

	for _, switched := range loginSwitches {
		if err := db.Migrator().DropColumn(&model.LoginFlow{}, switched.column); err != nil {
			return fmt.Errorf("drop %s: %w", switched.column, err)
		}
	}

	return db.Migrator().DropColumn(&model.EmailVerification{}, "new_email")
}
