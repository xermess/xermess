package cache

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

// The kinds of session the session database holds, each under its own key.
const (
	AdminSession = "admin"
	UserSession  = "user"
)

// SessionTTL is the longest a session is kept in Redis before it is read from
// the database again. Every write that changes a session puts the new one
// here itself, so this only bounds what an outage at the wrong moment could
// leave behind.
const SessionTTL = 10 * time.Minute

// A session is found by the hash of the token in its cookie, which no write
// knows in advance, so sessions are not kept in generations the way the
// groups are. The rule instead is who may overwrite what:
//
//   - A reader that missed and loaded the row from the database fills the
//     key only if it is still empty (FillSession, SET NX).
//   - A write puts the session as it has just committed it (PutSession, SET),
//     over whatever is there — a revoked session included, which stays in
//     Redis as revoked.
//
// So a reader that loaded a session a moment before it was revoked cannot
// put it back: either its fill lands first and the revocation overwrites it,
// or the revocation lands first and the fill finds the key taken.

// Session reads the session a token hash names into `dst`, and reports
// whether it was there.
func (c *Cache) Session(ctx context.Context, kind, hash string, dst any) bool {
	if c == nil || !c.available() || !c.settle(ctx) {
		return false
	}

	raw, err := c.client.Get(ctx, c.sessionKey(kind, hash)).Bytes()
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

// FillSession keeps a session a reader has just loaded from the database,
// unless a write has put one there meanwhile.
func (c *Cache) FillSession(ctx context.Context, kind, hash string, session any, expires time.Time) {
	if c == nil || !c.available() || !c.settle(ctx) {
		return
	}

	raw, err := json.Marshal(session)
	if err != nil {
		c.log.Error("cache: a session would not encode", "kind", kind, "error", err)
		return
	}

	if err := c.client.SetNX(ctx, c.sessionKey(kind, hash), raw, sessionTTL(expires)).Err(); err != nil {
		c.failed(err)
	}
}

// PutSession replaces a session with what a write has just committed. It is
// called after the commit; when Redis does not answer, the key is removed as
// soon as it does, and until then this process reads nothing from Redis.
func (c *Cache) PutSession(ctx context.Context, kind, hash string, session any, expires time.Time) {
	if c == nil || hash == "" {
		return
	}

	raw, err := json.Marshal(session)
	if err != nil {
		c.log.Error("cache: a session would not encode", "kind", kind, "error", err)
		c.forgetKeys(err, c.sessionKey(kind, hash))
		return
	}

	err = errUnavailable
	if c.available() && c.settle(ctx) {
		err = c.client.Set(ctx, c.sessionKey(kind, hash), raw, sessionTTL(expires)).Err()
	}

	if err != nil {
		c.failed(err)
		c.forgetKeys(err, c.sessionKey(kind, hash))
	}
}

// forgetKeys remembers keys a write could not bring up to date, to remove them
// once Redis answers again.
func (c *Cache) forgetKeys(err error, keys ...string) {
	c.mu.Lock()
	if c.pendingKeys == nil {
		c.pendingKeys = map[string]bool{}
	}
	for _, key := range keys {
		c.pendingKeys[key] = true
	}
	c.mu.Unlock()

	c.log.Error("cache: could not update sessions; not using redis until they are removed", "sessions", len(keys), "error", err)
}

// sessionKey is where the session a token hash names is kept.
func (c *Cache) sessionKey(kind, hash string) string {
	return c.key("session", kind, hash)
}

// sessionTTL is how long a session is kept: SessionTTL, or less when it ends
// sooner. A session that has already ended is kept a second, which is as
// good as not keeping it — the reader checks the expiry itself.
func sessionTTL(expires time.Time) time.Duration {
	return max(time.Second, min(SessionTTL, time.Until(expires)))
}

// DropSessions removes sessions that no longer exist in the database — the
// sweep's, once it has deleted them. Nothing can put one back but a reader
// that loaded it before the delete, and the sweep deletes only sessions that
// have expired, which that reader turns away itself.
func (c *Cache) DropSessions(ctx context.Context, kind string, hashes ...string) {
	if c == nil || len(hashes) == 0 {
		return
	}

	keys := make([]string, len(hashes))
	for i, hash := range hashes {
		keys[i] = c.sessionKey(kind, hash)
	}

	err := errUnavailable
	if c.available() {
		err = c.client.Unlink(ctx, keys...).Err()
	}

	if err != nil {
		c.failed(err)
		c.forgetKeys(err, keys...)
	}
}
