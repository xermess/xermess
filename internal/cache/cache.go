// Package cache is Redis: two databases on one server.
//
// The cache database holds what anybody may see and every page reads
// (languages, the organisation, login flows, sign-in buttons). The session
// database holds what decides who is signed in and what they may do (sessions,
// administrators, client credentials, grants) and the rate limit's counts, so
// flushing the cache never signs anybody out.
//
// Keys, under the configured prefix:
//
//	<prefix>cache:<group>:generation            the group's current generation
//	<prefix>cache:<group>:v<generation>:<entry> one cached value, as JSON
//	<prefix>session:<kind>:<token hash>         one session, as JSON
//	<prefix>ratelimit:<scope>:<address>         one rate-limit bucket
//
// A write forgets a whole group with one INCR of its generation, so a reader
// that loaded a row just before the write stores it under a generation nobody
// reads again. Sessions are handled differently; see sessions.go.
//
// Redis is an optimisation, never needed for correctness: a nil *Cache is
// always empty, and a Redis that stops answering is treated as a miss.
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

// Cache groups, each forgotten as a whole. The first live in the cache
// database, the rest in the session database.
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
	// Grants is what decides what a token carries: the role graph, and each
	// application's access to each API.
	Grants = "grants"
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

// KeepSettling retries, until ctx ends, the invalidations a write could not
// deliver while Redis was unreachable, so an idle process does not leave others
// serving a stale value.
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

// BumpAuthorityGroups moves the session-database groups on by one generation at
// startup, so an invalidation lost when a process died cannot outlive a
// restart.
func (r *Redis) BumpAuthorityGroups(ctx context.Context) error {
	if r == nil || r.Sessions == nil {
		return nil
	}

	return r.Sessions.incr(ctx, Groups[SessionDatabase])
}

// Cache is one Redis database, and the prefix every key starts with.
type Cache struct {
	client *redis.Client
	prefix string
	name   string
	number int
	log    *slog.Logger

	// warned logs an outage once, and again only after Redis has recovered in
	// between.
	mu        sync.Mutex
	warned    bool
	downUntil time.Time

	// pending is what a write could not forget because Redis did not answer.
	// Until it is delivered this process neither reads nor writes the database,
	// so an outage during a save cannot leave the old value served.
	pending     map[string]bool
	pendingKeys map[string]bool
}

// Open connects to both Redis databases and pings them, so a wrong address
// fails startup. With no Redis configured it returns nil: caches that never
// hold anything.
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

	// The cache logs outages once (failed), so go-redis's per-dial lines go to
	// debug.
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
		// A cache slower than the database is useless, so it fails fast and
		// cools down instead of retrying per request.
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
