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

// WithCache puts Redis in front of the reads every request makes. A nil
// Redis is no cache.
//
// Only a few reads go through it. In the cache database: what the sign-in
// pages and the panel ask for on every render and nobody changes often — the
// languages and their text, the organisation, the login flows, the sign-in
// buttons, how one-time codes work. In the session database: what every
// signed-in request reads to decide who is asking and what they may do — the
// sessions behind the cookies, the administrators with their roles, the
// applications by client id, the administrators' sign-in settings. Each
// method that writes one of those forgets it once the write has committed.
// Nothing else is cached, so nothing else can be stale: not users, not
// tokens or codes, which are spent once and have to be spent in the database.
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

// cached reads one value through the cache: from Redis when it is there, and
// otherwise from `load`, whose answer is then kept for the next reader. An
// error from `load` is returned and nothing is kept.
func cached[T any](ctx context.Context, s *Store, group, field string, load func() (T, error)) (T, error) {
	var value T
	// The generation read here is the one the value is written back into, so a
	// write that forgets the group while `load` runs moves the group on and
	// leaves this value in a generation nobody reads — rather than putting a
	// stale principal back after the write cleared it.
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

// forget drops what a write changed from the cache. It is called only once
// the write has committed: forgetting earlier would let a reader put the old
// row back before the new one is there.
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

// afterCommit is the context for telling the cache what a write changed. The
// write has committed, so the news has to reach Redis whether or not the
// request that made it is still there: a caller that hung up a moment after
// the commit would otherwise leave every other process reading the old row
// — the old client secret, the roles that were taken away, the session that
// was revoked — until the cache's TTL ran out. The request's values stay; its
// cancellation does not, and a short timeout of its own stands in for it.
func afterCommit(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.WithoutCancel(ctx), afterCommitTimeout)
}

// forgetting is forget for a write that may have failed: it forgets once the
// write has succeeded, and answers the write's error either way — so a write
// reads `return s.forgetting(ctx, write(), groups...)`.
func (s *Store) forgetting(ctx context.Context, err error, groups ...string) error {
	if err == nil {
		s.forget(ctx, groups...)
	}

	return err
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
