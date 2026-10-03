// Package cachetest opens the Redis named by LOGINER_TEST_REDIS (host:port) for
// a test. Each test gets its own key prefix, removed when it ends.
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

	"loginer/internal/brand"
	"loginer/internal/cache"
	"loginer/internal/config"
)

// Env names the Redis the tests may use.
const Env = "LOGINER_TEST_REDIS"

// Open returns the test Redis (databases 14 and 15) or nil when none is
// configured; the server must work either way.
func Open(t *testing.T) *cache.Redis {
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

	r, err := cache.Open(context.Background(), config.Redis{
		Host:      host,
		Port:      number,
		CacheDB:   14,
		SessionDB: 15,
		Username:  os.Getenv("LOGINER_REDIS_USERNAME"),
		Password:  os.Getenv("LOGINER_REDIS_PASSWORD"),
		Prefix:    brand.Slug + "_test_" + hex.EncodeToString(suffix) + ":",
	}, slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		// A test that closed the cache on purpose — to see what the server
		// does without Redis — could not have written anything after.
		for _, c := range r.Databases() {
			if err := c.Flush(context.Background()); err != nil && !errors.Is(err, redis.ErrClosed) {
				t.Errorf("remove the test's keys: %v", err)
			}
		}
		_ = r.Close()
	})

	return r
}

// Require is Open for a test that means nothing without a Redis: it skips
// when there is none.
func Require(t *testing.T) *cache.Redis {
	t.Helper()

	r := Open(t)
	if r == nil {
		t.Skipf("%s is not set", Env)
	}

	return r
}
