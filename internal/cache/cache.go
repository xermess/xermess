// Package cache is Redis: two databases on one server, each with its own job.
//
// The cache database (LOGINER_REDIS_CACHE_DB) holds what the sign-in pages
// and the panel read on every render and anybody may see — the languages and
// their text, the organisation, the login flows, the sign-in buttons, how a
// one-time code works. Losing it costs nothing but a slower page.
//
// The session database (LOGINER_REDIS_SESSION_DB) holds what decides who is
// signed in and what they may do: the sessions behind the cookies, the
// administrators those sessions belong to with the roles they hold, the
// applications' client credentials, the sign-in security settings — and the
// rate limit's counts. It can be given its own persistence and access, and
// flushing the cache database never signs anybody out.
//
// The store reads through both and forgets what it writes, so nothing above
// the store knows either is there. Every key is under the configured prefix,
// and then under what it is for, so a Redis browser shows a tree:
//
//	<prefix>cache:<group>:generation            the group's current generation
//	<prefix>cache:<group>:v<generation>:<entry> one cached value, as JSON
//	<prefix>session:<kind>:<token hash>         one session, as JSON
//	<prefix>ratelimit:<scope>:<address>         one rate-limit bucket
//
// Most of what is cached belongs to a group, and a write forgets the whole
// group at once by moving it on to its next generation: forgetting is one
// INCR. That is what makes invalidation safe with several server processes
// and a reader in flight — a reader that loaded a row just before a write
// stores it under the old generation, which nobody reads again, rather than
// putting stale text back after the write cleared it. Old generations are
// left to expire. Sessions are the exception, because they are found by
// their token rather than by anything a write knows in advance; see
// sessions.go.
//
// Redis is an optimisation, never a dependency of correctness. A nil *Cache
// is a cache that is always empty, which is what the server runs with when
// no Redis is configured; and a Redis that stops answering is logged and
// treated as a miss, so a page is slower rather than broken.
package cache

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"

	"loginer/internal/config"
)

// The two databases, by the name the panel and the log call them.
const (
	CacheDatabase   = "cache"
	SessionDatabase = "sessions"
)

// Groups of cached values, each forgotten as one. The entries of each are
// named where the store reads them. The first ones live in the cache
// database; the rest decide who may do what, and live in the session one.
const (
	Languages     = "languages"
	Organization  = "organization"
	LoginFlows    = "login_flows"
	SocialButtons = "social_buttons"
	SSOButtons    = "sso_buttons"
	OTPSettings   = "otp_settings"

	Admins        = "admins"
	Clients       = "clients"
	AdminSecurity = "admin_security"
)

// Redis is the two databases. Either is nil when no Redis is configured,
// and a nil *Redis has two nil databases.
type Redis struct {
	Cache    *Cache
	Sessions *Cache
}

// CacheDB is the cache database, or nil.
func (r *Redis) CacheDB() *Cache {
	if r == nil {
		return nil
	}
	return r.Cache
}

// SessionDB is the session database, or nil.
func (r *Redis) SessionDB() *Cache {
	if r == nil {
		return nil
	}
	return r.Sessions
}

// Databases lists the databases that are there, cache first.
func (r *Redis) Databases() []*Cache {
	if r == nil || r.Cache == nil {
		return nil
	}
	return []*Cache{r.Cache, r.Sessions}
}

// Close shuts both connection pools down.
func (r *Redis) Close() error {
	if r == nil {
		return nil
	}

	return errors.Join(r.Cache.Close(), r.Sessions.Close())
}

// settleRetryInterval is how often a process retries the invalidations it
// could not deliver while Redis was away.
const settleRetryInterval = 30 * time.Second

// KeepSettling retries, on a timer until ctx ends, the group and session
// invalidations a write could not deliver because Redis was unreachable at the
// moment it committed. Without it those sit in memory until the next request
// happens to touch the cache; a process that goes idle, or is about to be
// restarted, would otherwise leave other processes serving a stale value.
func (r *Redis) KeepSettling(ctx context.Context) {
	if r == nil {
		return
	}

	ticker := time.NewTicker(settleRetryInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			for _, c := range r.Databases() {
				if c != nil && c.available() {
					c.settle(ctx)
				}
			}
		}
	}
}

// BumpAuthorityGroups moves the session-database groups on by one generation.
// It runs at startup: a process that forgot one of these groups but died
// before Redis took the invalidation would otherwise leave every process
// serving a stale principal, client secret or security setting until the TTL
// ran out, and a restart would not clear it. Discarding a generation only ever
// throws cached values away, which is always safe, so this is cheap insurance.
func (r *Redis) BumpAuthorityGroups(ctx context.Context) error {
	if r == nil || r.Sessions == nil {
		return nil
	}

	return r.Sessions.incr(ctx, []string{Admins, Clients, AdminSecurity})
}

// Cache is one Redis database, and the prefix every key starts with.
type Cache struct {
	client *redis.Client
	prefix string
	name   string
	number int
	log    *slog.Logger

	// warned keeps an outage from writing a line per request: the first
	// failure is logged, and the next is logged once Redis has answered
	// again in between.
	mu        sync.Mutex
	warned    bool
	downUntil time.Time

	// pending is what a write could not forget because Redis did not
	// answer: groups to move on and keys to remove. Until they are done this
	// process reads none of the database and writes none of it, and every
	// call tries them again first — so a moment's outage during a save
	// cannot leave the old value being served once Redis is back.
	pending     map[string]bool
	pendingKeys map[string]bool
}

// Open connects to both databases of the Redis in the configuration, and
// checks each answers so a wrong address stops the server at startup rather
// than quietly caching nothing. With no Redis configured it returns nil,
// which is two working caches that never hold anything.
func Open(ctx context.Context, cfg config.Redis, log *slog.Logger) (*Redis, error) {
	if !cfg.Enabled() {
		return nil, nil
	}

	cacheDB, err := open(ctx, cfg, CacheDatabase, cfg.CacheDB, log)
	if err != nil {
		return nil, err
	}

	sessionDB, err := open(ctx, cfg, SessionDatabase, cfg.SessionDB, log)
	if err != nil {
		_ = cacheDB.Close()
		return nil, err
	}

	// go-redis writes a line of its own for every failed dial, which in an
	// outage is a line per request. The cache says it once (failed), so the
	// library's lines go to debug.
	redis.SetLogger(quiet{log: log})

	return &Redis{Cache: cacheDB, Sessions: sessionDB}, nil
}

// open connects to one database.
func open(ctx context.Context, cfg config.Redis, name string, number int, log *slog.Logger) (*Cache, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr(),
		Username: cfg.Username,
		Password: cfg.Password,
		DB:       number,
		// A cache that takes longer than this to answer is slower than the
		// database it stands in front of, so it fails fast and is left alone
		// for a while (cooldown) rather than retried on every request.
		DialTimeout:   time.Second,
		DialerRetries: 1,
		MaxRetries:    1,
		ReadTimeout:   500 * time.Millisecond,
		WriteTimeout:  500 * time.Millisecond,
	})

	ping, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	if err := client.Ping(ping).Err(); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("redis at %s (%s database %d) did not answer: %w; start it, or leave LOGINER_REDIS_HOST empty to run without it",
			cfg.Addr(), name, number, err)
	}

	return &Cache{client: client, prefix: cfg.Prefix, name: name, number: number, log: log.With("redis", name)}, nil
}

// Name is which of the two databases this is: CacheDatabase or
// SessionDatabase.
func (c *Cache) Name() string {
	if c == nil {
		return ""
	}
	return c.name
}

// Number is the Redis database number.
func (c *Cache) Number() int {
	if c == nil {
		return 0
	}
	return c.number
}

// Close shuts the connection pool down.
func (c *Cache) Close() error {
	if c == nil {
		return nil
	}

	return c.client.Close()
}

// key is a key under this server's prefix: its parts joined with colons.
func (c *Cache) key(parts ...string) string {
	return c.prefix + strings.Join(parts, ":")
}
