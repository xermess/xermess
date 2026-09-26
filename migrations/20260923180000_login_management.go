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
// it allows, beyond the steps it is made of.
var loginSwitches = []string{
	"allow_sign_in",
	"allow_remember_me",
	"verify_email_on_register",
	"allow_email_change",
}

// upLoginManagement adds them, and the address a verification link may be a
// change to (model.EmailVerification.NewEmail).
//
// The flows that already exist keep behaving exactly as they did, which for
// three of the four means off. The exception is allow_sign_in: every flow
// there is signs people in today, so it is turned on for all of them — the
// column is added with that as its default, which fills the rows that exist,
// and the default is then dropped so GORM cannot quietly store true for a
// flow somebody meant to close (model_test.TestNoBoolDefaultsToTrue).
func upLoginManagement(_ context.Context, tx *sql.Tx) error {
	db, err := gormTx(tx)
	if err != nil {
		return err
	}

	if err := db.AutoMigrate(&model.LoginFlow{}, &model.EmailVerification{}); err != nil {
		return fmt.Errorf("add the login flow switches: %w", err)
	}

	// AutoMigrate writes a not-null bool with no default, which Postgres
	// fills with false. Only the master switch has to say otherwise.
	if _, err := tx.Exec(`UPDATE login_flows SET allow_sign_in = true`); err != nil {
		return fmt.Errorf("open the flows that already exist: %w", err)
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

	for _, column := range loginSwitches {
		if err := db.Migrator().DropColumn(&model.LoginFlow{}, column); err != nil {
			return fmt.Errorf("drop %s: %w", column, err)
		}
	}

	return db.Migrator().DropColumn(&model.EmailVerification{}, "new_email")
}
