package migrations

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/pressly/goose/v3"

	"loginer/internal/model"
)

func init() {
	goose.AddMigrationContext(upAdminAvatarAndLogIndexes, downAdminAvatarAndLogIndexes)
}

// logIndexes are what keep the logs page's filters a range read as the log
// grows: its pages are read newest first by (created_at, id), and narrowed by
// the kind of entry or by who acted.
var logIndexes = []struct{ name, columns string }{
	{"idx_audit_logs_created_id", "created_at DESC, id DESC"},
	{"idx_audit_logs_action_created", "action, created_at DESC"},
	{"idx_audit_logs_actor_created", "actor_email, created_at DESC"},
}

// upAdminAvatarAndLogIndexes gives administrators a picture
// (model.AdminUser.AvatarURL), empty for everyone who has one today, and the
// activity log the indexes its filters need.
func upAdminAvatarAndLogIndexes(_ context.Context, tx *sql.Tx) error {
	db, err := gormTx(tx)
	if err != nil {
		return err
	}

	if err := db.AutoMigrate(&model.AdminUser{}); err != nil {
		return fmt.Errorf("add the administrators' avatar: %w", err)
	}

	for _, index := range logIndexes {
		statement := fmt.Sprintf(`CREATE INDEX IF NOT EXISTS %s ON audit_logs (%s)`, index.name, index.columns)
		if _, err := tx.Exec(statement); err != nil {
			return fmt.Errorf("create %s: %w", index.name, err)
		}
	}

	return nil
}

// downAdminAvatarAndLogIndexes drops them again. What is lost is the
// pictures administrators chose; the log itself is untouched.
func downAdminAvatarAndLogIndexes(_ context.Context, tx *sql.Tx) error {
	db, err := gormTx(tx)
	if err != nil {
		return err
	}

	for _, index := range logIndexes {
		if _, err := tx.Exec(`DROP INDEX IF EXISTS ` + index.name); err != nil {
			return fmt.Errorf("drop %s: %w", index.name, err)
		}
	}

	return db.Migrator().DropColumn(&model.AdminUser{}, "avatar_url")
}
