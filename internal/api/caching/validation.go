package caching

import (
	"bytes"
	"encoding/json"
	"net/http"
	"slices"
	"time"

	"loginer/internal/api/respond"
	"loginer/internal/cache"
)

// The problems these endpoints answer with.
var (
	notConfigured    = respond.Define(http.StatusServiceUnavailable, "redis_not_configured", respond.Admin)
	databaseNotFound = respond.Define(http.StatusNotFound, "cache_database_not_found", respond.Admin)
	groupNotFound    = respond.Define(http.StatusNotFound, "cache_group_not_found", respond.Admin)
	keyNotFound      = respond.Define(http.StatusNotFound, "cache_key_not_found", respond.Admin)
	keyNotEditable   = respond.Define(http.StatusConflict, "cache_key_not_editable", respond.Admin)
	generationKey    = respond.Define(http.StatusConflict, "cache_generation_key", respond.Admin)
	invalidKeyName   = respond.Define(http.StatusBadRequest, "cache_key_invalid", respond.Admin)
	invalidKind      = respond.Define(http.StatusBadRequest, "cache_kind_invalid", respond.Admin)
	invalidCursor    = respond.Define(http.StatusBadRequest, "cache_cursor_invalid", respond.Admin)
	invalidValue     = respond.Define(http.StatusBadRequest, "cache_value_invalid", respond.Admin)
	invalidTTL       = respond.Define(http.StatusBadRequest, "cache_ttl_invalid", respond.Admin)
)

// maxTTL is the longest a value written by hand may be given: a day, well
// past the hour the store keeps a value itself, and short enough that a
// mistake does not outlive the day it was made.
const maxTTL = 24 * time.Hour

// kinds are the kinds of key a listing may be narrowed to.
var kinds = []string{
	cache.KindEntry, cache.KindStale, cache.KindGeneration,
	cache.KindSession, cache.KindRateLimit, cache.KindOther,
}

// checkKind accepts no kind, or one of the kinds.
func checkKind(kind string) error {
	if kind != "" && !slices.Contains(kinds, kind) {
		return invalidKind.With()
	}
	return nil
}

// checkName accepts a key name as a listing gave it: not empty, and short
// enough to be one of this server's.
func checkName(name string) error {
	if name == "" || len(name) > 1024 {
		return invalidKeyName.With()
	}
	return nil
}

// check reads the new value and how long it should live.
func (r writeRequest) check() (json.RawMessage, time.Duration, error) {
	// null is JSON, but it decodes into an empty record without complaint, so
	// the pages would show nothing where the value was.
	if len(r.Value) == 0 || !json.Valid(r.Value) || bytes.Equal(bytes.TrimSpace(r.Value), []byte("null")) {
		return nil, 0, invalidValue.With()
	}

	ttl := time.Duration(r.TTLSeconds) * time.Second
	if ttl < 0 || ttl > maxTTL {
		return nil, 0, invalidTTL.With("max", int(maxTTL.Seconds()))
	}

	return r.Value, ttl, nil
}
