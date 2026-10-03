package cache

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"
)

// cooldown is how long Redis is skipped after a failure, so an outage costs one
// probe per cooldown instead of a failed dial per request.
const cooldown = 5 * time.Second

// errUnavailable is what a call answers while the cache is leaving Redis
// alone.
var errUnavailable = errors.New("redis is not answering")

// available reports whether the cache may talk to Redis: it is not in the
// cooldown that follows a failure.
func (c *Cache) available() bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	return time.Now().After(c.downUntil)
}

// failed notes a Redis that did not answer and starts a cooldown. A context
// that ended belongs to the caller, says nothing about Redis, and starts none.
func (c *Cache) failed(err error) {
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return
	}

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

// quiet passes go-redis's own log lines on at debug level.
type quiet struct{ log *slog.Logger }

func (q quiet) Printf(ctx context.Context, format string, v ...any) {
	q.log.DebugContext(ctx, "redis: "+fmt.Sprintf(format, v...))
}
