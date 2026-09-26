package migrations

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/pressly/goose/v3"

	"loginer/internal/model"
)

func init() {
	goose.AddMigrationContext(upOrganizationTimezone, downOrganizationTimezone)
}

// upOrganizationTimezone gives the organisation the zone its dates are said
// in (model.Organization.Timezone), and takes away its domain, which nothing
// read.
//
// The column is added with UTC as its default, which fills the row that
// exists: that is the zone the pages were already saying dates in whenever
// the server ran in it, and the panel is where somebody corrects it. A
// database built after the domain went never had the column, so it is only
// dropped where it is.
func upOrganizationTimezone(_ context.Context, tx *sql.Tx) error {
	db, err := gormTx(tx)
	if err != nil {
		return err
	}

	if err := db.AutoMigrate(&model.Organization{}); err != nil {
		return fmt.Errorf("add the organization's timezone: %w", err)
	}

	if db.Migrator().HasColumn(&model.Organization{}, "domain") {
		if err := db.Migrator().DropColumn(&model.Organization{}, "domain"); err != nil {
			return fmt.Errorf("drop the organization's domain: %w", err)
		}
	}

	return nil
}

// downOrganizationTimezone puts the domain column back, empty, and drops the
// timezone. What is lost is the zone that was chosen, and whatever domain
// was written before the column went.
func downOrganizationTimezone(_ context.Context, tx *sql.Tx) error {
	db, err := gormTx(tx)
	if err != nil {
		return err
	}

	if _, err := tx.Exec(`ALTER TABLE organizations ADD COLUMN IF NOT EXISTS domain varchar(253)`); err != nil {
		return fmt.Errorf("restore the organization's domain: %w", err)
	}

	return db.Migrator().DropColumn(&model.Organization{}, "timezone")
}
