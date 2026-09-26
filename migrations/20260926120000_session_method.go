package migrations

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/pressly/goose/v3"

	"loginer/internal/model"
)

func init() {
	goose.AddMigrationContext(upSessionMethod, downSessionMethod)
}

// upSessionMethod records on a user session the way in that made it
// (model.UserSession.Method), which is what lets an application's login flow
// say whether a session made through some other application's flow satisfies
// it (model.LoginFlow.Accepts).
//
// The column is added by hand rather than by AutoMigrate for the reason the
// login switches were: Postgres refuses a not-null column with no default on a
// table that has rows, and user_sessions holds everybody signed in right now.
//
// Those rows are given "password". It is a guess — the server did not record
// this before — and the conservative one of the four: it is what the sign-in
// page's own way in writes, so the sessions it is right about go on working,
// and where it is wrong, which is the sessions a provider or an organisation's
// identity provider made, a flow without the social step refuses a session it
// would have taken and whoever holds it signs in once more. The guess is never
// the stronger claim: it cannot satisfy a flow that asks for an emailed code,
// which is the hole this closes.
//
// The default is dropped again so that a session written without a method
// fails instead of quietly becoming a password one.
func upSessionMethod(ctx context.Context, tx *sql.Tx) error {
	add := fmt.Sprintf(
		`ALTER TABLE user_sessions ADD COLUMN IF NOT EXISTS method varchar(16) NOT NULL DEFAULT '%s'`,
		string(model.MethodPassword),
	)
	if _, err := tx.ExecContext(ctx, add); err != nil {
		return fmt.Errorf("add method: %w", err)
	}

	drop := `ALTER TABLE user_sessions ALTER COLUMN method DROP DEFAULT`
	if _, err := tx.ExecContext(ctx, drop); err != nil {
		return fmt.Errorf("drop the default on method: %w", err)
	}

	// What the struct says, now the column it could not add itself is there:
	// the table and model.UserSession cannot drift apart over the type.
	db, err := gormTx(tx)
	if err != nil {
		return err
	}
	if err := db.AutoMigrate(&model.UserSession{}); err != nil {
		return fmt.Errorf("bring user_sessions up to the model: %w", err)
	}

	return nil
}

// downSessionMethod drops the column again. What is lost is what proved each
// session: the sessions stay and go on signing their users in, and an
// application whose flow asks for an emailed code takes any of them again, as
// it did before the column existed.
func downSessionMethod(_ context.Context, tx *sql.Tx) error {
	db, err := gormTx(tx)
	if err != nil {
		return err
	}

	return db.Migrator().DropColumn(&model.UserSession{}, "method")
}
