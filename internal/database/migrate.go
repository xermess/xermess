package database

import (
	"database/sql"
	"fmt"
	"log/slog"
	"os"

	"github.com/pressly/goose/v3"
	"gorm.io/gorm"

	"loginer/internal/config"
	// Registers the migrations with goose. They are Go functions, so they
	// exist only in a binary that imports them.
	_ "loginer/migrations"
)

// Migrate applies the migrations that have not run yet. Goose records what it
// applied in the goose_db_version table, so running this on every start is
// safe.
func Migrate(db *gorm.DB, cfg config.DB, log *slog.Logger) error {
	sqlDB, err := prepareGoose(db)
	if err != nil {
		return err
	}

	before, err := goose.GetDBVersion(sqlDB)
	if err != nil {
		return err
	}

	if err := goose.Up(sqlDB, cfg.MigrateDir); err != nil {
		return err
	}

	after, err := goose.GetDBVersion(sqlDB)
	if err != nil {
		return err
	}

	if after == before {
		log.Info("database is up to date", "version", after)
	} else {
		log.Info("migrations applied", "from", before, "to", after)
	}

	return nil
}

// MigrateDown rolls the newest migration back, one step.
func MigrateDown(db *gorm.DB, cfg config.DB, log *slog.Logger) error {
	sqlDB, err := prepareGoose(db)
	if err != nil {
		return err
	}

	if err := goose.Down(sqlDB, cfg.MigrateDir); err != nil {
		return err
	}

	version, err := goose.GetDBVersion(sqlDB)
	if err != nil {
		return err
	}

	log.Info("rolled back", "version", version)

	return nil
}

// MigrateStatus prints which migrations have run and which have not.
func MigrateStatus(db *gorm.DB, cfg config.DB) error {
	sqlDB, err := prepareGoose(db)
	if err != nil {
		return err
	}

	// Status is a table meant to be read, so let goose print it itself.
	goose.SetLogger(printLogger{})

	return goose.Status(sqlDB, cfg.MigrateDir)
}

// printLogger lets goose write its own output, which is what makes the status
// table readable. Goose's Logger interface requires Fatalf, and goose only
// calls it on an error it cannot return.
type printLogger struct{}

func (printLogger) Printf(format string, v ...any) {
	fmt.Printf(format, v...)
}

func (printLogger) Fatalf(format string, v ...any) {
	fmt.Fprintf(os.Stderr, format, v...)
	os.Exit(1)
}

// prepareGoose hands goose the underlying connection and tells it which
// database it is talking to.
func prepareGoose(db *gorm.DB) (*sql.DB, error) {
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	if err := goose.SetDialect("postgres"); err != nil {
		return nil, err
	}

	// Goose prints its own progress; we log the outcome ourselves so the
	// format matches the rest of the server.
	goose.SetLogger(goose.NopLogger())

	return sqlDB, nil
}
