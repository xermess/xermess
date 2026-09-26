// Package database opens the database connection.
package database

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"loginer/internal/config"
)

// Open connects to Postgres and checks the connection works, so a bad
// database stops the server at startup instead of on the first request.
func Open(cfg config.DB) (*gorm.DB, error) {
	if cfg.Driver != "postgres" {
		return nil, fmt.Errorf("unsupported database driver %q", cfg.Driver)
	}

	level := logger.Warn
	if cfg.LogQueries {
		level = logger.Info
	}

	// GORM's own default, but for one thing: a lookup that finds nothing is
	// not an error here. The store asks "is there one?" with First all over —
	// is this address taken, has this language been imported — and turns
	// the answer into store.ErrNotFound for the caller to act on. Logged, each
	// of those would read as a failure on a server that is working.
	db, err := gorm.Open(postgres.Open(dsn(cfg)), &gorm.Config{
		Logger: logger.New(log.New(os.Stdout, "\r\n", log.LstdFlags), logger.Config{
			SlowThreshold:             200 * time.Millisecond,
			LogLevel:                  level,
			IgnoreRecordNotFoundError: true,
			Colorful:                  true,
		}),
	})
	if err != nil {
		return nil, fmt.Errorf("connect to database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("ping database: %w", err)
	}

	// Past MaxConns a query waits for a free connection rather than opening
	// another; idle ones are kept, since a sign-in that has to dial first is
	// slower than one that does not. Connections are replaced every half hour
	// so a failover or a resized pool behind a proxy is picked up.
	// Unset — a tool or a test opening the database for itself — keeps Go's
	// own defaults.
	if cfg.MaxConns > 0 {
		sqlDB.SetMaxOpenConns(cfg.MaxConns)
		sqlDB.SetMaxIdleConns(cfg.MaxConns)
		sqlDB.SetConnMaxLifetime(30 * time.Minute)
		sqlDB.SetConnMaxIdleTime(5 * time.Minute)
	}

	return db, nil
}

// Close shuts the connection pool down.
func Close(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// dsn adds the time zone to the DSN when it does not already carry one. It is
// added by hand because escaping would turn the slash in "Asia/Bishkek" into
// %2F, which Postgres rejects.
func dsn(cfg config.DB) string {
	if strings.Contains(cfg.DSN, "TimeZone=") {
		return cfg.DSN
	}

	separator := "?"
	if strings.Contains(cfg.DSN, "?") {
		separator = "&"
	}

	return cfg.DSN + separator + "TimeZone=" + cfg.TimeZone
}
