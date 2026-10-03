// Package store is the only place that writes queries: everything above it gets
// models and this package's errors.
package store

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"gorm.io/gorm"

	"loginer/internal/cache"
)

// Store holds the database connection every query runs on, and the two Redis
// databases in front of it.
type Store struct {
	db       *gorm.DB
	cache    *cache.Cache
	sessions *cache.Cache
}

// New returns a store backed by the given connection, with no cache.
func New(db *gorm.DB) *Store {
	return &Store{db: db}
}

// WithCache puts Redis in front of the hot reads. A nil Redis is no cache.
//
// Cached: languages, the organisation, login flows, sign-in buttons and OTP
// settings (cache database); sessions, administrators, applications by client
// id, admin security and grants (session database). Each writer forgets what it
// changed after commit. Users, tokens and codes are never cached.
func (s *Store) WithCache(r *cache.Redis) *Store {
	s.cache = r.CacheDB()
	s.sessions = r.SessionDB()
	return s
}

// in is the database a group is kept in.
func (s *Store) in(group string) *cache.Cache {
	if slices.Contains(cache.Groups[cache.SessionDatabase], group) {
		return s.sessions
	}
	return s.cache
}

// cached reads a value through the cache, loading and storing it on a miss. A
// load error is returned and nothing is stored.
func cached[T any](ctx context.Context, s *Store, group, field string, load func() (T, error)) (T, error) {
	var value T
	// Write back into the generation read, so a write that forgets the group
	// during `load` leaves this value unread.
	generation, ok := s.in(group).GetAt(ctx, group, field, &value)
	if ok {
		return value, nil
	}

	value, err := load()
	if err != nil {
		return value, err
	}

	s.in(group).SetAt(ctx, group, field, value, generation)

	return value, nil
}

// forget is called only after commit; forgetting earlier would let a reader
// cache the old row again.
func (s *Store) forget(ctx context.Context, groups ...string) {
	ctx, cancel := afterCommit(ctx)
	defer cancel()

	for _, group := range groups {
		s.in(group).Forget(ctx, group)
	}
}

// afterCommitTimeout is how long telling the cache about a committed write
// may take.
const afterCommitTimeout = 2 * time.Second

// afterCommit is the context for invalidating after a commit: it keeps the
// request's values but not its cancellation, with a short timeout, so a caller
// hanging up cannot leave other processes serving stale data.
func afterCommit(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.WithoutCancel(ctx), afterCommitTimeout)
}

// forgetting forgets the groups if the write succeeded and returns its error:
// `return s.forgetting(ctx, write(), groups...)`.
func (s *Store) forgetting(ctx context.Context, err error, groups ...string) error {
	if err == nil {
		s.forget(ctx, groups...)
	}

	return err
}

// ErrNotFound stands in for gorm.ErrRecordNotFound so callers do not import
// GORM.
var ErrNotFound = errors.New("not found")

// ErrDuplicate is a unique-index violation, such as a taken email: the caller
// can fix it.
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
		// Wrapped so errors.Is still matches and DuplicateField can read the
		// column.
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
