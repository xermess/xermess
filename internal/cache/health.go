package cache

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"
)

// cooldown is how long the cache leaves Redis alone after it failed to
// answer. Without it every request of an outage would wait on a dial that is
// going to fail; with it, one request in each cooldown finds out whether Redis
// is back and the rest go straight to the database.
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

// failed notes a Redis that did not answer: the cache leaves it alone for the
// cooldown, and says so once per outage.
//
// A context that ended is the caller's — a browser that went away, a request
// out of time — and says nothing about Redis, so it starts no cooldown: one
// would send every process's reads to the database for five seconds, and
// leave any write in those seconds unable to tell the other processes what
// it changed.
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
