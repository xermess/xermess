// Package cache is Redis: the cache in front of the database, and the shared
// counters the rate limit keeps.
//
// The cache holds what the sign-in pages and the panel ask for on every
// render and an administrator changes perhaps once a week — the languages
// and their text, the organisation, the login flows, the sign-in buttons. The
// store reads through it and forgets what it writes, so nothing above the
// store knows it is there.
//
// Every key is under the configured prefix, and then under what it is for,
// so a Redis browser shows a tree and `redis-cli --scan --pattern
// 'xermess:cache:languages:*'` finds one group:
//
//	xermess:cache:<group>:generation            the group's current generation
//	xermess:cache:<group>:v<generation>:<entry> one cached value, as JSON
//	xermess:ratelimit:<scope>:<address>         one rate-limit bucket
//
// Everything cached belongs to a group, and a write forgets the whole group
// at once by moving the group on to its next generation: forgetting is one
// INCR. That is what makes invalidation safe with several server processes
// and a reader in flight — a reader that loaded a row just before a write
// stores it under the old generation, which nobody reads again, rather than
// putting stale text back after the write cleared it. Old generations are
// left to expire.
//
// Redis is an optimisation, never a dependency of correctness. A nil *Cache
// is a cache that is always empty, which is what the server runs with when
// no Redis is configured; and a Redis that stops answering is logged and
// treated as a miss, so a page is slower rather than broken.
package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"

	"xermess/internal/config"
)

// TTL is how long a cached value lives if nothing forgets it first. Writes
// forget what they change, so this only bounds how long an abandoned
// generation takes up memory.
const TTL = time.Hour

// cooldown is how long the cache leaves Redis alone after it failed to
// answer. Without it every request of an outage would wait on a dial that is
// going to fail; with it, one request in each cooldown finds out whether Redis
// is back and the rest go straight to the database.
const cooldown = 5 * time.Second

// errUnavailable is what a call answers while the cache is leaving Redis
// alone.
var errUnavailable = errors.New("redis is not answering")

// Groups of cached values, each forgotten as one. The entries of each are
// named where the store reads them.
const (
	Languages     = "languages"
	Organization  = "organization"
	LoginFlows    = "login_flows"
	SocialButtons = "social_buttons"
)

// Cache is a Redis connection and the prefix every key starts with.
type Cache struct {
	client *redis.Client
	prefix string
	log    *slog.Logger

	// warned keeps an outage from writing a line per request: the first
	// failure is logged, and the next is logged once Redis has answered
	// again in between.
	mu        sync.Mutex
	warned    bool
	downUntil time.Time

	// pending are the groups a write could not forget because Redis did not
	// answer. Until they are forgotten this process reads none of the cache
	// and writes none of it, and every call tries the forget again first — so
	// a moment's outage during a save cannot leave the old text being served
	// once Redis is back.
	pending map[string]bool
}

// Open connects to the Redis in the configuration, and checks it answers so a
// wrong address stops the server at startup rather than quietly caching
// nothing. With no Redis configured it returns nil, which is a working cache
// that never holds anything.
func Open(ctx context.Context, cfg config.Redis, log *slog.Logger) (*Cache, error) {
	if !cfg.Enabled() {
		return nil, nil
	}

	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr(),
		Username: cfg.Username,
		Password: cfg.Password,
		DB:       cfg.DB,
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
		return nil, fmt.Errorf("redis at %s (database %d) did not answer: %w; start it, or leave XERMESS_REDIS_HOST empty to run without it", cfg.Addr(), cfg.DB, err)
	}

	// go-redis writes a line of its own for every failed dial, which in an
	// outage is a line per request. The cache says it once (failed), so the
	// library's lines go to debug.
	redis.SetLogger(quiet{log: log})

	return &Cache{client: client, prefix: cfg.Prefix, log: log}, nil
}

// quiet passes go-redis's own log lines on at debug level.
type quiet struct{ log *slog.Logger }

func (q quiet) Printf(ctx context.Context, format string, v ...any) {
	q.log.DebugContext(ctx, "redis: "+fmt.Sprintf(format, v...))
}

// Close shuts the connection pool down.
func (c *Cache) Close() error {
	if c == nil {
		return nil
	}

	return c.client.Close()
}

// Get reads one cached entry of a group into `dst`, and reports whether there
// was one. A miss, a value that no longer decodes and a Redis that did not
// answer are all "no": the caller reads the database instead.
func (c *Cache) Get(ctx context.Context, group, entry string, dst any) bool {
	if c == nil || !c.available() || !c.settle(ctx) {
		return false
	}

	key, err := c.entryKey(ctx, group, entry)
	if err != nil {
		c.failed(err)
		return false
	}

	raw, err := c.client.Get(ctx, key).Bytes()
	switch {
	case errors.Is(err, redis.Nil):
		c.recovered()
		return false
	case err != nil:
		c.failed(err)
		return false
	}

	c.recovered()

	return json.Unmarshal(raw, dst) == nil
}

// Set caches one entry of a group.
func (c *Cache) Set(ctx context.Context, group, entry string, value any) {
	if c == nil || !c.available() || !c.settle(ctx) {
		return
	}

	raw, err := json.Marshal(value)
	if err != nil {
		c.log.Error("cache: a value would not encode", "group", group, "entry", entry, "error", err)
		return
	}

	key, err := c.entryKey(ctx, group, entry)
	if err != nil {
		c.failed(err)
		return
	}

	if err := c.client.Set(ctx, key, raw, TTL).Err(); err != nil {
		c.failed(err)
	}
}

// Forget drops everything cached in the given groups, for every server
// process at once.
//
// It is called after a write has committed. When Redis cannot be reached the
// groups are remembered as pending (see Cache.pending) and forgotten as soon
// as it answers again.
func (c *Cache) Forget(ctx context.Context, groups ...string) {
	if c == nil {
		return
	}

	err := errUnavailable
	if c.available() {
		err = c.incr(ctx, groups)
	}

	if err != nil {
		c.failed(err)

		c.mu.Lock()
		if c.pending == nil {
			c.pending = map[string]bool{}
		}
		for _, group := range groups {
			c.pending[group] = true
		}
		c.mu.Unlock()

		c.log.Error("cache: could not forget a group; not using the cache until it is forgotten",
			"groups", groups, "error", err)
	}
}

// settle forgets the pending groups, and reports whether the cache may be
// used: there are none left, or they have just been forgotten.
func (c *Cache) settle(ctx context.Context) bool {
	c.mu.Lock()
	groups := make([]string, 0, len(c.pending))
	for group := range c.pending {
		groups = append(groups, group)
	}
	c.mu.Unlock()

	if len(groups) == 0 {
		return true
	}

	if err := c.incr(ctx, groups); err != nil {
		c.failed(err)
		return false
	}

	c.mu.Lock()
	for _, group := range groups {
		delete(c.pending, group)
	}
	c.mu.Unlock()

	c.log.Info("cache: forgot what a write changed while redis was away", "groups", groups)

	return true
}

// incr moves each group on to its next generation.
func (c *Cache) incr(ctx context.Context, groups []string) error {
	pipe := c.client.Pipeline()
	for _, group := range groups {
		pipe.Incr(ctx, c.generationKey(group))
	}

	_, err := pipe.Exec(ctx)

	return err
}

// Flush removes every key under this server's prefix — the cached values,
// the generations and the rate limit's buckets — and leaves anything else in
// the Redis alone. It is SCAN rather than KEYS, so a large Redis is not
// blocked while it looks.
func (c *Cache) Flush(ctx context.Context) error {
	if c == nil {
		return nil
	}

	iter := c.client.Scan(ctx, 0, c.prefix+"*", 500).Iterator()

	var batch []string
	for iter.Next(ctx) {
		batch = append(batch, iter.Val())

		if len(batch) == 500 {
			if err := c.client.Unlink(ctx, batch...).Err(); err != nil {
				return err
			}
			batch = batch[:0]
		}
	}
	if err := iter.Err(); err != nil {
		return err
	}

	if len(batch) > 0 {
		return c.client.Unlink(ctx, batch...).Err()
	}

	return nil
}

// key is a key under this server's prefix: its parts joined with colons.
func (c *Cache) key(parts ...string) string {
	return c.prefix + strings.Join(parts, ":")
}

// generationKey is the counter that says which generation of a group is
// current. It never expires, so memory pressure evicts cached values and
// never what says which of them may be read.
func (c *Cache) generationKey(group string) string {
	return c.key("cache", group, "generation")
}

// entryKey is where an entry of a group lives in the current generation.
func (c *Cache) entryKey(ctx context.Context, group, entry string) (string, error) {
	generation, err := c.client.Get(ctx, c.generationKey(group)).Int64()
	if err != nil && !errors.Is(err, redis.Nil) {
		return "", err
	}

	return c.key("cache", group, "v"+strconv.FormatInt(generation, 10), entry), nil
}

// available reports whether the cache may talk to Redis: it is not in the
// cooldown that follows a failure.
func (c *Cache) available() bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	return time.Now().After(c.downUntil)
}

// failed notes a Redis that did not answer: the cache leaves it alone for the
// cooldown, and says so once per outage.
func (c *Cache) failed(err error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.downUntil = time.Now().Add(cooldown)

	if !c.warned {
		c.log.Warn("cache: redis did not answer; reading the database until it does", "error", err)
		c.warned = true
	}
}

// recovered notes that Redis answered, so the next outage is logged again.
func (c *Cache) recovered() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.warned {
		c.log.Info("cache: redis is answering again")
		c.warned = false
	}
}
