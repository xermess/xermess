package cache

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

// What a super admin can see of Redis and do to it, for the panel's cache
// page. Everything here stays under this server's prefix: another
// installation sharing the Redis is neither listed nor touched.
//
// Two things are refused whatever the caller. A generation counter is never
// removed: a group whose counter starts again at zero would read the values of
// generation zero again, however old. And nothing in the session database is
// written by hand: a session or an administrator edited there would sign
// somebody in, or give them roles, that the database never did. Removing
// anything there is safe — the next request reads the database again.

// Kinds of key, by what the name says it is.
const (
	KindEntry      = "entry"      // a cached value of a group's current generation
	KindStale      = "stale"      // a cached value of an older generation, left to expire
	KindGeneration = "generation" // the counter that says which generation is current
	KindSession    = "session"    // a session, found by its token's hash
	KindRateLimit  = "ratelimit"  // a rate limit's bucket for one address
	KindOther      = "other"      // under the prefix, but nothing this server writes
)

// Errors the management calls answer with.
var (
	ErrNoRedis       = errors.New("redis is not configured")
	ErrNoKey         = errors.New("no such key")
	ErrGenerationKey = errors.New("a generation counter cannot be removed; clear its group instead")
	ErrNotEditable   = errors.New("only a cached value in the cache database can be edited")
)

// Groups lists the groups each database holds, in the order the panel shows
// them.
var Groups = map[string][]string{
	CacheDatabase:   {Languages, Organization, LoginFlows, SocialButtons, SSOButtons, OTPSettings},
	SessionDatabase: {Admins, Clients, AdminSecurity},
}

// Key is one key, described.
type Key struct {
	// Name is the key without this server's prefix.
	Name string
	Kind string
	// Group is the group, the kind of session, or the rate limit's scope;
	// Entry is the entry, the token's hash, or the address.
	Group string
	Entry string
	// TTL is how long the key has left, or -1 for a key that does not expire.
	TTL time.Duration
	// Size is what Redis says the key takes in memory, in bytes, or 0 when it
	// would not say.
	Size int64
	// Editable says the value may be replaced by hand.
	Editable bool
}

// Value is a key with what it holds: JSON for a cached value or a session,
// the fields of a rate limit's bucket.
type Value struct {
	Key
	Type  string
	Value json.RawMessage
}

// GroupStats is one group of a database.
type GroupStats struct {
	Name       string
	Generation int64
	Entries    int64
	Stale      int64
}

// Stats is one database, summed up.
type Stats struct {
	Name      string
	Number    int
	Available bool
	// Keys counts the keys under this server's prefix, by kind.
	Keys     int64
	Kinds    map[string]int64
	Groups   []GroupStats
	Sessions map[string]int64
	// Memory and Version describe the Redis server, which both databases
	// share.
	Memory  int64
	Version string
}

// Stats sums a database up. It walks every key under the prefix, which on a
// Redis holding millions of sessions takes a moment: it is for a person
// looking at a page, not for a request path.
func (c *Cache) Stats(ctx context.Context) (Stats, error) {
	if c == nil {
		return Stats{}, ErrNoRedis
	}

	stats := Stats{Name: c.name, Number: c.number, Kinds: map[string]int64{}, Sessions: map[string]int64{}}

	generations, err := c.generations(ctx, Groups[c.name])
	if err != nil {
		return stats, err
	}
	stats.Available = true

	groups := map[string]*GroupStats{}
	for _, name := range Groups[c.name] {
		groups[name] = &GroupStats{Name: name, Generation: generations[name]}
	}

	iter := c.client.Scan(ctx, 0, c.prefix+"*", 1000).Iterator()
	for iter.Next(ctx) {
		key := c.describe(strings.TrimPrefix(iter.Val(), c.prefix), generations)
		stats.Keys++
		stats.Kinds[key.Kind]++

		switch key.Kind {
		case KindEntry, KindStale:
			group, ok := groups[key.Group]
			if !ok {
				group = &GroupStats{Name: key.Group}
				groups[key.Group] = group
			}
			if key.Kind == KindEntry {
				group.Entries++
			} else {
				group.Stale++
			}
		case KindSession:
			stats.Sessions[key.Group]++
		}
	}
	if err := iter.Err(); err != nil {
		return stats, err
	}

	for _, name := range Groups[c.name] {
		stats.Groups = append(stats.Groups, *groups[name])
		delete(groups, name)
	}
	// A group this server no longer reads, still holding values from before.
	for _, group := range groups {
		stats.Groups = append(stats.Groups, *group)
	}

	stats.Memory, stats.Version = c.server(ctx)

	return stats, nil
}

// KeyQuery narrows a listing of keys.
type KeyQuery struct {
	// Kind and Group keep the keys of one kind, and of one group within it.
	Kind  string
	Group string
	// Search keeps the keys whose name contains it.
	Search string
	// Cursor is where the previous page stopped, 0 for the first page.
	Cursor uint64
	Limit  int
}

// Keys lists keys under the prefix, a page at a time: it answers the keys and
// the cursor to pass for the next page, which is 0 after the last.
//
// Redis walks a database in no particular order, so a page is whatever the
// walk reached, and a key written meanwhile may or may not be on it.
func (c *Cache) Keys(ctx context.Context, q KeyQuery) ([]Key, uint64, error) {
	if c == nil {
		return nil, 0, ErrNoRedis
	}

	limit := q.Limit
	if limit <= 0 || limit > 500 {
		limit = 100
	}

	generations, err := c.generations(ctx, Groups[c.name])
	if err != nil {
		return nil, 0, err
	}

	pattern := c.prefix + keyPattern(q.Kind, q.Group)

	var (
		keys   []Key
		cursor = q.Cursor
	)
	// A bounded number of steps: a pattern that matches little in a large
	// database would otherwise walk all of it for one page.
	for step := 0; step < 50; step++ {
		names, next, err := c.client.Scan(ctx, cursor, pattern, 500).Result()
		if err != nil {
			return nil, 0, err
		}

		for _, name := range names {
			key := c.describe(strings.TrimPrefix(name, c.prefix), generations)
			if q.Kind != "" && key.Kind != q.Kind && !(q.Kind == KindEntry && key.Kind == KindStale) {
				continue
			}
			if q.Search != "" && !strings.Contains(strings.ToLower(key.Name), strings.ToLower(q.Search)) {
				continue
			}
			keys = append(keys, key)
		}

		cursor = next
		if cursor == 0 || len(keys) >= limit {
			break
		}
	}

	if err := c.measure(ctx, keys); err != nil {
		return nil, 0, err
	}

	return keys, cursor, nil
}

// Read answers one key and what it holds.
func (c *Cache) Read(ctx context.Context, name string) (Value, error) {
	if c == nil {
		return Value{}, ErrNoRedis
	}

	generations, err := c.generations(ctx, Groups[c.name])
	if err != nil {
		return Value{}, err
	}

	keys := []Key{c.describe(name, generations)}
	if err := c.measure(ctx, keys); err != nil {
		return Value{}, err
	}

	full := c.prefix + name
	kind, err := c.client.Type(ctx, full).Result()
	if err != nil {
		return Value{}, err
	}

	value := Value{Key: keys[0], Type: kind}

	switch kind {
	case "none":
		return Value{}, ErrNoKey
	case "string":
		raw, err := c.client.Get(ctx, full).Bytes()
		if err != nil {
			return Value{}, err
		}
		if !json.Valid(raw) {
			// A counter or a value this server did not write as JSON.
			raw, _ = json.Marshal(string(raw))
		}
		value.Value = raw
	case "hash":
		fields, err := c.client.HGetAll(ctx, full).Result()
		if err != nil {
			return Value{}, err
		}
		value.Value, _ = json.Marshal(fields)
	default:
		value.Value, _ = json.Marshal("a " + kind + " this server does not write")
	}

	return value, nil
}

// Write replaces a cached value by hand. `value` has to be JSON, and is what
// the next reader decodes: one that no longer decodes into what the store
// expects is a miss, and the database is read instead. A zero `ttl` keeps
// the time the value had left.
func (c *Cache) Write(ctx context.Context, name string, value json.RawMessage, ttl time.Duration) error {
	if c == nil {
		return ErrNoRedis
	}

	generations, err := c.generations(ctx, Groups[c.name])
	if err != nil {
		return err
	}

	if key := c.describe(name, generations); !key.Editable {
		return ErrNotEditable
	}

	full := c.prefix + name
	exists, err := c.client.Exists(ctx, full).Result()
	if err != nil {
		return err
	}
	if exists == 0 {
		return ErrNoKey
	}

	if ttl > 0 {
		return c.client.Set(ctx, full, []byte(value), ttl).Err()
	}

	return c.client.SetArgs(ctx, full, []byte(value), redis.SetArgs{KeepTTL: true}).Err()
}

// Delete removes one key, and reports whether it was there.
func (c *Cache) Delete(ctx context.Context, name string) (bool, error) {
	if c == nil {
		return false, ErrNoRedis
	}

	if c.describe(name, nil).Kind == KindGeneration {
		return false, ErrGenerationKey
	}

	removed, err := c.client.Unlink(ctx, c.prefix+name).Result()

	return removed > 0, err
}

// Clear forgets a group now, and says when it could not — unlike Forget,
// which a write calls and which cannot fail.
func (c *Cache) Clear(ctx context.Context, group string) error {
	if c == nil {
		return ErrNoRedis
	}

	return c.incr(ctx, []string{group})
}

// Flush removes every key under this server's prefix in this database — the
// cached values and the generations, or the sessions and the rate limit's
// buckets — and leaves anything else in the Redis alone. It is SCAN rather
// than KEYS, so a large Redis is not blocked while it looks.
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

// describe says what a key is from its name. `generations` are the current
// generations of the groups; a group not among them is taken to be at
// generation zero.
func (c *Cache) describe(name string, generations map[string]int64) Key {
	key := Key{Name: name, Kind: KindOther, TTL: -1}

	parts := strings.SplitN(name, ":", 4)
	switch {
	case len(parts) == 3 && parts[0] == "cache" && parts[2] == "generation":
		key.Kind, key.Group = KindGeneration, parts[1]
	case len(parts) == 4 && parts[0] == "cache" && strings.HasPrefix(parts[2], "v"):
		key.Group, key.Entry = parts[1], parts[3]
		generation, _ := strconv.ParseInt(strings.TrimPrefix(parts[2], "v"), 10, 64)
		if generation == generations[key.Group] {
			key.Kind = KindEntry
			key.Editable = c.name == CacheDatabase
		} else {
			key.Kind = KindStale
		}
	case len(parts) >= 3 && parts[0] == "session":
		key.Kind, key.Group, key.Entry = KindSession, parts[1], strings.Join(parts[2:], ":")
	case len(parts) >= 3 && parts[0] == "ratelimit":
		// An IPv6 address has colons of its own.
		key.Kind, key.Group, key.Entry = KindRateLimit, parts[1], strings.Join(parts[2:], ":")
	}

	return key
}

// measure fills in each key's time to live and size, in one round trip.
func (c *Cache) measure(ctx context.Context, keys []Key) error {
	if len(keys) == 0 {
		return nil
	}

	pipe := c.client.Pipeline()
	ttls := make([]*redis.DurationCmd, len(keys))
	sizes := make([]*redis.IntCmd, len(keys))
	for i, key := range keys {
		ttls[i] = pipe.PTTL(ctx, c.prefix+key.Name)
		sizes[i] = pipe.MemoryUsage(ctx, c.prefix+key.Name)
	}

	// MEMORY USAGE is missing from some servers that speak the protocol, and
	// answers nil for a key that went away meanwhile: neither is a failure
	// of the listing, so only the connection's own error is.
	if _, err := pipe.Exec(ctx); err != nil && !errors.Is(err, redis.Nil) && !isCommandError(err) {
		return err
	}

	for i := range keys {
		if ttl, err := ttls[i].Result(); err == nil && ttl >= 0 {
			keys[i].TTL = ttl
		}
		keys[i].Size, _ = sizes[i].Result()
	}

	return nil
}

// generations reads the current generation of each group in one round trip.
func (c *Cache) generations(ctx context.Context, groups []string) (map[string]int64, error) {
	current := make(map[string]int64, len(groups))
	if len(groups) == 0 {
		return current, nil
	}

	names := make([]string, len(groups))
	for i, group := range groups {
		names[i] = c.generationKey(group)
	}

	values, err := c.client.MGet(ctx, names...).Result()
	if err != nil {
		return nil, err
	}

	for i, value := range values {
		if text, ok := value.(string); ok {
			current[groups[i]], _ = strconv.ParseInt(text, 10, 64)
		}
	}

	return current, nil
}

// server reads how much memory the Redis server uses and which version it is.
// Either is left empty when the server will not say.
func (c *Cache) server(ctx context.Context) (memory int64, version string) {
	info, err := c.client.Info(ctx, "server", "memory").Result()
	if err != nil {
		return 0, ""
	}

	lines := bufio.NewScanner(strings.NewReader(info))
	for lines.Scan() {
		name, value, ok := strings.Cut(strings.TrimSpace(lines.Text()), ":")
		if !ok {
			continue
		}
		switch name {
		case "used_memory":
			memory, _ = strconv.ParseInt(value, 10, 64)
		case "redis_version":
			version = value
		}
	}

	return memory, version
}

// keyPattern is the SCAN pattern for a kind of key, and a group within it.
func keyPattern(kind, group string) string {
	switch kind {
	case KindEntry, KindStale:
		if group != "" {
			return "cache:" + escapeGlob(group) + ":v*"
		}
		return "cache:*:v*"
	case KindGeneration:
		return "cache:*:generation"
	case KindSession:
		if group != "" {
			return "session:" + escapeGlob(group) + ":*"
		}
		return "session:*"
	case KindRateLimit:
		if group != "" {
			return "ratelimit:" + escapeGlob(group) + ":*"
		}
		return "ratelimit:*"
	}

	// A group with no kind: everything cached in it, and its generation.
	if group != "" {
		return "cache:" + escapeGlob(group) + ":*"
	}

	return "*"
}

// escapeGlob escapes what a SCAN pattern treats specially.
func escapeGlob(text string) string {
	return strings.NewReplacer(`\`, `\\`, `*`, `\*`, `?`, `\?`, `[`, `\[`, `]`, `\]`).Replace(text)
}

// isCommandError reports whether Redis refused a command, rather than the
// connection failing.
func isCommandError(err error) bool {
	var redisErr redis.Error
	return errors.As(err, &redisErr)
}
