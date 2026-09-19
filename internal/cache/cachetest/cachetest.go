// Package cachetest opens a Redis for a test that wants one.
//
// The Redis is named by XERMESS_TEST_REDIS, as host:port, and signed in to
// with XERMESS_REDIS_USERNAME and XERMESS_REDIS_PASSWORD when those are set —
// `make test-integration` fills all three in from .env. Each test gets a key
// prefix of its own, and its keys are removed when it ends, so tests can share
// a Redis with a running server without either seeing the other's.
package cachetest

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"io"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/redis/go-redis/v9"

	"xermess/internal/cache"
	"xermess/internal/config"
)

// Env names the Redis the tests may use.
const Env = "XERMESS_TEST_REDIS"

// Open returns a cache on the test Redis, or nil when there is none — which
// is a cache that never holds anything, and a server has to work with that
// too.
func Open(t *testing.T) *cache.Cache {
	t.Helper()

	addr := os.Getenv(Env)
	if addr == "" {
		return nil
	}

	host, port, _ := strings.Cut(addr, ":")
	number, err := strconv.Atoi(port)
	if err != nil {
		t.Fatalf("%s = %q, want host:port", Env, addr)
	}

	suffix := make([]byte, 6)
	if _, err := rand.Read(suffix); err != nil {
		t.Fatal(err)
	}

	c, err := cache.Open(context.Background(), config.Redis{
		Host:     host,
		Port:     number,
		Username: os.Getenv("XERMESS_REDIS_USERNAME"),
		Password: os.Getenv("XERMESS_REDIS_PASSWORD"),
		Prefix:   "xermess_test_" + hex.EncodeToString(suffix) + ":",
	}, slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		// A test that closed the cache on purpose — to see what the server
		// does without Redis — could not have written anything after.
		if err := c.Flush(context.Background()); err != nil && !errors.Is(err, redis.ErrClosed) {
			t.Errorf("remove the test's keys: %v", err)
		}
		_ = c.Close()
	})

	return c
}

// Require is Open for a test that means nothing without a Redis: it skips
// when there is none.
func Require(t *testing.T) *cache.Cache {
	t.Helper()

	c := Open(t)
	if c == nil {
		t.Skipf("%s is not set", Env)
	}

	return c
}
