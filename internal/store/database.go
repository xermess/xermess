package store

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"time"

	"gorm.io/gorm/clause"
)

// This file is the database browser's half of the store: the panel's read-only
// view of the tables this server keeps its own records in.
//
// It is the one place in the project that reads a table by name rather than
// through a model, so two rules hold it in. A name is only ever used after it
// has been found in the list of tables the database actually has — a name that
// is not on that list is refused, and nothing a caller sends reaches a query.
// And the columns holding a password, a key or a token are never read at all:
// they are dropped from the SELECT rather than blanked afterwards, so a value
// that is not shown is a value that was never fetched.

// Table is one of the server's tables, as the panel lists it.
type Table struct {
	Name    string `json:"name"`
	Rows    int64  `json:"rows"`
	Columns int    `json:"columns"`
}

// Column is one column of a table.
type Column struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Nullable bool   `json:"nullable"`

	// PrimaryKey marks the column rows are identified by.
	PrimaryKey bool `json:"primary_key"`

	// Hidden marks a column this browser will not read: a password, a key or
	// a token. The panel shows the column and says it is hidden, which is
	// more honest than leaving it out and looking like the table has no such
	// thing.
	Hidden bool `json:"hidden"`
}

// Tables returns every table in the database, by name, with how big each is.
func (s *Store) Tables(ctx context.Context) ([]Table, error) {
	names, err := s.tableNames(ctx)
	if err != nil {
		return nil, err
	}

	tables := make([]Table, 0, len(names))
	for _, name := range names {
		columns, err := s.TableColumns(ctx, name)
		if err != nil {
			return nil, err
		}

		var rows int64
		if err := s.db.WithContext(ctx).Table(name).Count(&rows).Error; err != nil {
			return nil, fmt.Errorf("count %s: %w", name, err)
		}

		tables = append(tables, Table{Name: name, Rows: rows, Columns: len(columns)})
	}

	return tables, nil
}

// TableColumns returns the columns of one table, in the order the database
// holds them. It returns ErrNotFound for a name that is not a table.
func (s *Store) TableColumns(ctx context.Context, table string) ([]Column, error) {
	known, err := s.knownTable(ctx, table)
	if err != nil {
		return nil, err
	}

	types, err := s.db.WithContext(ctx).Migrator().ColumnTypes(known)
	if err != nil {
		return nil, fmt.Errorf("describe %s: %w", known, err)
	}

	columns := make([]Column, 0, len(types))
	for _, column := range types {
		nullable, _ := column.Nullable()
		primary, _ := column.PrimaryKey()

		columns = append(columns, Column{
			Name:       column.Name(),
			Type:       strings.ToLower(column.DatabaseTypeName()),
			Nullable:   nullable,
			PrimaryKey: primary,
			Hidden:     hiddenColumn(column.Name()),
		})
	}

	return columns, nil
}

// TableRows returns a page of one table's rows, and how many there are
// altogether. Hidden columns are not read.
//
// The newest rows come first where the table says when a row was made, which
// is every table a model describes; the rest are returned in whatever order
// the database holds them, which is the only order they have.
func (s *Store) TableRows(ctx context.Context, table string, limit, offset int) ([]map[string]any, int64, error) {
	known, err := s.knownTable(ctx, table)
	if err != nil {
		return nil, 0, err
	}

	columns, err := s.TableColumns(ctx, known)
	if err != nil {
		return nil, 0, err
	}

	var total int64
	if err := s.db.WithContext(ctx).Table(known).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count %s: %w", known, err)
	}

	// The columns are named as clause.Column rather than as text, so the
	// driver quotes each one. A table here may have a column called
	// "default" — api_scopes does — and an unquoted name like that is a
	// syntax error rather than a column.
	query := s.db.WithContext(ctx).Table(known).
		Clauses(clause.Select{Columns: readable(columns)}).
		Limit(limit).Offset(offset)

	if hasColumn(columns, "created_at") {
		query = query.Order(clause.OrderByColumn{Column: clause.Column{Name: "created_at"}, Desc: true})
	}

	var rows []map[string]any
	if err := query.Find(&rows).Error; err != nil {
		return nil, 0, fmt.Errorf("read %s: %w", known, err)
	}

	for _, row := range rows {
		tidyRow(row)
	}

	return rows, total, nil
}

// tableNames is what the database actually has, sorted. It is read each time
// rather than cached: a migration adds a table, and a panel that has to be
// restarted to see it would be lying about what is there.
func (s *Store) tableNames(ctx context.Context) ([]string, error) {
	names, err := s.db.WithContext(ctx).Migrator().GetTables()
	if err != nil {
		return nil, fmt.Errorf("list tables: %w", err)
	}

	slices.Sort(names)

	return names, nil
}

// knownTable returns the name as the database spells it, or ErrNotFound.
// Every query in this file goes through it: a name that reaches a statement
// is one the database listed, never one a request sent.
func (s *Store) knownTable(ctx context.Context, table string) (string, error) {
	names, err := s.tableNames(ctx)
	if err != nil {
		return "", err
	}

	wanted := strings.ToLower(strings.TrimSpace(table))
	for _, name := range names {
		if strings.ToLower(name) == wanted {
			return name, nil
		}
	}

	return "", ErrNotFound
}

// hiddenSuffixes and hiddenNames are how a column that holds something nobody
// should read back is recognised: the hash standing in for a token, and the
// secrets kept encrypted rather than hashed.
//
// The suffixes are the naming this project already follows, so a model added
// later is covered without anybody remembering this list. They are suffixes
// rather than anything a name merely contains, because that over-reaches:
// "allow_password_reset" is a switch and "secret_hint" is the four characters
// shown on the Applications page, and hiding either would say the table holds
// something it does not.
var (
	hiddenSuffixes = []string{"_hash", "_secret"}
	hiddenNames    = []string{"secret", "password", "private_key", "recovery_codes"}
)

func hiddenColumn(name string) bool {
	lowered := strings.ToLower(name)

	for _, suffix := range hiddenSuffixes {
		if strings.HasSuffix(lowered, suffix) {
			return true
		}
	}

	return slices.Contains(hiddenNames, lowered)
}

// readable is the columns a SELECT may name.
func readable(columns []Column) []clause.Column {
	names := make([]clause.Column, 0, len(columns))

	for _, column := range columns {
		if !column.Hidden {
			names = append(names, clause.Column{Name: column.Name})
		}
	}

	return names
}

func hasColumn(columns []Column, name string) bool {
	return slices.ContainsFunc(columns, func(column Column) bool {
		return strings.EqualFold(column.Name, name)
	})
}

// tidyRow turns what the driver hands back into something JSON can carry and
// a person can read: bytes become how many there are rather than a wall of
// base64, and times become the same strings every other endpoint uses.
func tidyRow(row map[string]any) {
	for key, value := range row {
		switch typed := value.(type) {
		case []byte:
			row[key] = fmt.Sprintf("%d bytes", len(typed))
		case time.Time:
			row[key] = typed.UTC().Format(time.RFC3339)
		}
	}
}
