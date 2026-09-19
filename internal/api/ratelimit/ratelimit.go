// Package ratelimit slows down guessing: it limits how often one address may
// call the endpoints that take a password or send an email.
//
// The per-account lockout already stops guessing one account's password. This
// is the other half: one address trying many accounts, creating accounts in
// bulk, or filling someone's inbox with reset links.
//
// With Redis configured the count is kept there, so every server process
// shares one budget per address and a restart does not hand out a fresh one.
// Without it — or while Redis is not answering — each process counts in its
// own memory, which is a weaker limit rather than none.
package ratelimit

import (
	"context"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	"xermess/internal/api/respond"
	"xermess/internal/cache"
)

// RateLimited is what an address over the limit is told, with how many
// seconds until it may try again.
var RateLimited = respond.Define(http.StatusTooManyRequests, "rate_limited", respond.Both)

// Limiter is a token bucket per client address: `perMinute` requests a
// minute, refilled continuously, with the whole minute's worth available at
// once.
type Limiter struct {
	perMinute int
	rate      float64 // tokens per second
	burst     float64

	// shared is the Redis the buckets are kept in, and scope keeps this
	// limiter's buckets apart from another's there: the public and admin
	// servers each have their own.
	shared *cache.Cache
	scope  string

	mu      sync.Mutex
	buckets map[string]*bucket
	swept   time.Time

	now func() time.Time
}

type bucket struct {
	tokens float64
	seen   time.Time
}

// New returns a limiter allowing `perMinute` requests a minute per address.
// Zero or less allows everything.
func New(perMinute int) *Limiter {
	return &Limiter{
		perMinute: perMinute,
		rate:      float64(perMinute) / 60,
		burst:     float64(perMinute),
		buckets:   map[string]*bucket{},
		now:       time.Now,
	}
}

// Shared keeps the buckets in Redis, under `scope`, so every server process
// counts against the same budget. A nil cache leaves them in memory.
func (l *Limiter) Shared(c *cache.Cache, scope string) *Limiter {
	l.shared = c
	l.scope = scope
	return l
}

// Allow takes a token for `key`, and says whether there was one and, when not,
// how long until there is. It asks Redis when there is one, and its own
// memory when there is not or when Redis does not answer.
func (l *Limiter) Allow(ctx context.Context, key string) (bool, time.Duration) {
	if l.burst <= 0 {
		return true, 0
	}

	if l.shared != nil {
		ok, wait, err := l.shared.Take(ctx, l.scope+":"+key, l.perMinute)
		if err == nil {
			return ok, wait
		}
	}

	return l.allowLocally(key)
}

// allowLocally is Allow against this process's own buckets.
func (l *Limiter) allowLocally(key string) (bool, time.Duration) {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := l.now()
	l.sweep(now)

	b, ok := l.buckets[key]
	if !ok {
		b = &bucket{tokens: l.burst, seen: now}
		l.buckets[key] = b
	}

	b.tokens = min(l.burst, b.tokens+now.Sub(b.seen).Seconds()*l.rate)
	b.seen = now

	if b.tokens < 1 {
		return false, time.Duration((1 - b.tokens) / l.rate * float64(time.Second))
	}

	b.tokens--
	return true, 0
}

// sweep forgets addresses whose bucket has refilled, so the map does not grow
// with every address that ever called. It runs at most once a minute.
func (l *Limiter) sweep(now time.Time) {
	if now.Sub(l.swept) < time.Minute {
		return
	}
	l.swept = now

	full := time.Duration(l.burst / l.rate * float64(time.Second))
	for key, b := range l.buckets {
		if now.Sub(b.seen) > full {
			delete(l.buckets, key)
		}
	}
}

// Middleware refuses a request over the limit with 429 and Retry-After. The
// address is Gin's ClientIP, so behind a proxy XERMESS_TRUSTED_PROXIES has to
// name it — otherwise every client shares the proxy's one budget.
func (l *Limiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ok, wait := l.Allow(c.Request.Context(), c.ClientIP())
		if !ok {
			seconds := int(wait/time.Second) + 1
			c.Header("Retry-After", strconv.Itoa(seconds))
			respond.Abort(c, RateLimited, "seconds", seconds)
			return
		}

		c.Next()
	}
}
