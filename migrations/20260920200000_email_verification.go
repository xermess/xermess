package migrations

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/pressly/goose/v3"

	"xermess/internal/model"
)

func init() {
	goose.AddMigrationContext(upEmailVerification, downEmailVerification)
}

// upEmailVerification adds the links a login flow that requires a verified
// address sends (model.EmailVerification).
func upEmailVerification(_ context.Context, tx *sql.Tx) error {
	db, err := gormTx(tx)
	if err != nil {
		return err
	}

	if err := db.AutoMigrate(&model.EmailVerification{}); err != nil {
		return fmt.Errorf("create email_verifications: %w", err)
	}

	return nil
}

func downEmailVerification(_ context.Context, tx *sql.Tx) error {
	db, err := gormTx(tx)
	if err != nil {
		return err
	}

	return db.Migrator().DropTable(&model.EmailVerification{})
}
