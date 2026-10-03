// Package ratelimit limits how often one address may call the endpoints that
// take a password or send email, complementing the per-account lockout. With
// Redis every process shares one budget per address; without it each counts in
// memory.
package ratelimit

import (
	"context"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	"loginer/internal/api/respond"
	"loginer/internal/cache"
)

// RateLimited is what an address over the limit is told, with how many
// seconds until it may try again.
var RateLimited = respond.Define(http.StatusTooManyRequests, "rate_limited", respond.Both)

// Limiter is a per-address token bucket: `perMinute` requests a minute,
// refilled continuously, with a full minute's burst.
type Limiter struct {
	perMinute int
	rate      float64 // tokens per second
	burst     float64

	// shared is the Redis buckets are kept in; scope separates this limiter's
	// buckets from others'.
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

// Allow takes a token for `key`, reporting whether there was one and how long
// until the next. It uses Redis, falling back to memory when Redis is absent or
// down.
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

// bucketFor groups IPv6 clients by /64, since one client can rotate through a
// whole /64.
func bucketFor(address string) string {
	ip := net.ParseIP(address)
	if ip == nil || ip.To4() != nil {
		return address
	}

	return ip.Mask(net.CIDRMask(64, 128)).String() + "/64"
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

// Middleware answers 429 with Retry-After over the limit. It keys on Gin's
// ClientIP, so behind a proxy LOGINER_TRUSTED_PROXIES must name it.
func (l *Limiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ok, wait := l.Allow(c.Request.Context(), bucketFor(c.ClientIP()))
		if !ok {
			seconds := int(wait/time.Second) + 1
			c.Header("Retry-After", strconv.Itoa(seconds))
			respond.Abort(c, RateLimited, "seconds", seconds)
			return
		}

		c.Next()
	}
}
