package migrations

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/pressly/goose/v3"

	"loginer/internal/model"
)

func init() {
	goose.AddMigrationContext(upMailAndOTP, downMailAndOTP)
}

// mailAndOTP are the tables the Mail and One-time codes pages own: how this
// installation sends email (model.MailSettings), how the codes it emails
// behave (model.OTPSettings), and the sign-ins waiting for one
// (model.LoginCode).
//
// The two settings tables are seeded on the next start rather than here —
// store.EnsureMailSettings reads LOGINER_SMTP_* and EnsureOTPSettings the
// defaults — so a database stepped forward gets the same row a fresh one
// gets, from the configuration it is actually running with.
var mailAndOTP = []any{
	&model.MailSettings{},
	&model.OTPSettings{},
	&model.LoginCode{},
}

// upMailAndOTP adds them.
func upMailAndOTP(_ context.Context, tx *sql.Tx) error {
	db, err := gormTx(tx)
	if err != nil {
		return err
	}

	if err := db.AutoMigrate(mailAndOTP...); err != nil {
		return fmt.Errorf("create the mail and one-time code tables: %w", err)
	}

	return nil
}

// downMailAndOTP drops them. The mail server's address and its sealed
// password go with them, and so does any sign-in waiting for a code — which
// is no loss: whoever was signing in signs in again.
func downMailAndOTP(_ context.Context, tx *sql.Tx) error {
	db, err := gormTx(tx)
	if err != nil {
		return err
	}

	return db.Migrator().DropTable(mailAndOTP...)
}
