// Package store is the only place in the server that writes queries.
//
// Everything above it — the HTTP handlers, the auth service — asks the store
// for what it needs and gets models back. That keeps the queries in one place
// to read and change, keeps GORM out of the handlers, and means the errors
// the rest of the code handles are this package's own rather than the
// driver's.
package store

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"gorm.io/gorm"

	"xermess/internal/cache"
)

// Store holds the database connection every query runs on, and the cache in
// front of it.
type Store struct {
	db    *gorm.DB
	cache *cache.Cache
}

// New returns a store backed by the given connection, with no cache.
func New(db *gorm.DB) *Store {
	return &Store{db: db}
}

// WithCache puts a cache in front of the reads that every page asks for. A
// nil cache is no cache.
//
// Only a few reads go through it — what the sign-in pages and the panel ask
// for on every render and nobody changes often: the languages and their
// text, the organisation, the login flows and the sign-in buttons. Each
// method that writes one of those forgets its group once the write has
// committed. Nothing else is cached, so nothing else can be stale.
func (s *Store) WithCache(c *cache.Cache) *Store {
	s.cache = c
	return s
}

// cached reads one value through the cache: from Redis when it is there, and
// otherwise from `load`, whose answer is then kept for the next reader. An
// error from `load` is returned and nothing is kept.
func cached[T any](ctx context.Context, s *Store, group, field string, load func() (T, error)) (T, error) {
	var value T
	if s.cache.Get(ctx, group, field, &value) {
		return value, nil
	}

	value, err := load()
	if err != nil {
		return value, err
	}

	s.cache.Set(ctx, group, field, value)

	return value, nil
}

// forget drops what a write changed from the cache. It is called only once
// the write has committed: forgetting earlier would let a reader put the old
// row back before the new one is there.
func (s *Store) forget(ctx context.Context, groups ...string) {
	s.cache.Forget(ctx, groups...)
}

// ErrNotFound is returned when a row that was asked for is not there. It
// stands in for gorm.ErrRecordNotFound so callers do not import GORM to
// answer a 404.
var ErrNotFound = errors.New("not found")

// ErrDuplicate is returned when a write would repeat a value a unique index
// forbids — the same email twice, say. It is the writer's to fix, which is
// why it is told apart from anything else that can go wrong.
var ErrDuplicate = errors.New("already exists")

// translate turns what the driver returns into this package's errors.
// Anything it does not recognise is passed through unchanged.
func translate(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, gorm.ErrRecordNotFound):
		return ErrNotFound
	case strings.Contains(strings.ToLower(err.Error()), "duplicate"):
		// Wrapped rather than replaced: errors.Is still finds the sentinel,
		// and what the database said is still there for DuplicateField to
		// read the column out of.
		return fmt.Errorf("%w: %s", ErrDuplicate, err)
	default:
		return err
	}
}

// likeEscaper escapes what LIKE treats specially, so a search for "50%" or
// "first_name" finds those characters rather than matching anything.
var likeEscaper = strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)

// contains is the LIKE pattern for a case-insensitive search: the value, lower
// cased and escaped, anywhere in the column. Compare it against LOWER(column).
func contains(search string) string {
	return "%" + likeEscaper.Replace(strings.ToLower(search)) + "%"
}
