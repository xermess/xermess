package migrations

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/pressly/goose/v3"

	"xermess/internal/model"
)

func init() {
	goose.AddMigrationContext(upOrganization, downOrganization)
}

// upOrganization adds the organizations table and the one row every
// installation has: what the organisation is called, how users reach it, and
// the agreements they accept.
//
// The row is seeded rather than left to the first request: the panel's
// settings page edits an organisation, and a page with nothing to edit would
// have to invent one. What it holds is model.DefaultOrganization, so the
// database a fresh installation starts with and the fallback in the store
// cannot drift apart.
func upOrganization(_ context.Context, tx *sql.Tx) error {
	db, err := gormTx(tx)
	if err != nil {
		return err
	}

	if err := db.AutoMigrate(&model.Organization{}); err != nil {
		return fmt.Errorf("create organizations table: %w", err)
	}

	var count int64
	if err := db.Model(&model.Organization{}).Count(&count).Error; err != nil {
		return fmt.Errorf("count organizations: %w", err)
	}
	if count > 0 {
		return nil
	}

	organization := model.DefaultOrganization()

	return db.Create(&organization).Error
}

// downOrganization drops the table again. Nothing else refers to it, so the
// settings it held are all that is lost.
func downOrganization(_ context.Context, tx *sql.Tx) error {
	db, err := gormTx(tx)
	if err != nil {
		return err
	}

	return db.Migrator().DropTable(&model.Organization{})
}
