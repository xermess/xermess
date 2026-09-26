package oidc

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"loginer/internal/jose"
	"loginer/internal/model"
	"loginer/internal/store"
)

// How signing keys come and go.
//
// A key is made, and published in the JWKS, keyPublishLead before anything is
// signed with it: an API that caches the JWKS, even for a long time, already
// has the new key when the first token signed with it arrives. Once it signs,
// the key before it is retired; a retired key is still published for
// keyRetention, longer than any token lives, so tokens it signed keep
// verifying until they expire. After that it is deleted.
const (
	keyPublishLead = 24 * time.Hour
	keyRetention   = 48 * time.Hour
	// keyReloadInterval is the least time between two reloads of the keys
	// when a token names a key id this server does not know: another server
	// may have made it.
	keyReloadInterval = 30 * time.Second
	// keyMaintenanceInterval is how often a server checks whether keys are
	// due to rotate, retire or be deleted, and picks up what other servers
	// did.
	keyMaintenanceInterval = time.Minute
)

// storedKey is a key and when it was made and retired.
type storedKey struct {
	jose.Key
	CreatedAt time.Time
	RetiredAt *time.Time
}

// keySet holds the signing keys in memory.
type keySet struct {
	store  *store.Store
	sealer *jose.Sealer
	log    *slog.Logger

	// rotation is how long a key signs before the next one is made. Zero
	// never rotates on its own.
	rotation time.Duration

	now func() time.Time

	mu       sync.RWMutex
	all      []storedKey // newest first
	byID     map[string]jose.Key
	loadedAt time.Time
}

// loadKeys reads the stored keys and makes one for any algorithm that has none,
// so every algorithm an API may choose can be signed with from the first
// request.
func loadKeys(ctx context.Context, st *store.Store, sealer *jose.Sealer, rotation time.Duration, log *slog.Logger) (*keySet, error) {
	ks := &keySet{store: st, sealer: sealer, rotation: rotation, log: log, now: time.Now}

	if err := ks.maintain(ctx); err != nil {
		return nil, err
	}

	return ks, nil
}

// maintain brings the keys up to date: a first key for an algorithm with none,
// a next key for one whose key is due to rotate, the old key retired once the
// new one signs, and long-retired keys deleted. Every server runs it; the
// store's lock and its checks make running it twice harmless.
func (ks *keySet) maintain(ctx context.Context) error {
	if err := ks.reload(ctx); err != nil {
		return err
	}

	now := ks.now().Truncate(time.Microsecond)

	for _, alg := range jose.Algorithms {
		newest, signing := ks.newestAndSigning(alg, now)

		switch {
		case newest == nil:
			// No key at all: make one, usable at once.
			if err := ks.add(ctx, alg, now); err != nil {
				return err
			}
		case ks.rotation > 0 && signing != nil && newest.ID == signing.ID && now.Sub(signing.CreatedAt) >= ks.rotation:
			// The signing key is due, and no next key waits: make one.
			if err := ks.add(ctx, alg, now); err != nil {
				return err
			}
		}

		// Keys older than the one signing have been taken over from.
		if signing != nil {
			if err := ks.store.RetireSigningKeys(ctx, alg, signing.CreatedAt, now); err != nil {
				return err
			}
		}
	}

	if err := ks.store.DeleteSigningKeys(ctx, now.Add(-keyRetention)); err != nil {
		return err
	}

	return ks.reload(ctx)
}

// add makes a key for an algorithm, unless another server has just made one.
func (ks *keySet) add(ctx context.Context, alg string, now time.Time) error {
	private, err := jose.Generate(alg)
	if err != nil {
		return err
	}

	sealed, err := ks.sealer.Seal(private)
	if err != nil {
		return err
	}

	var suffix [6]byte
	if _, err := rand.Read(suffix[:]); err != nil {
		return err
	}

	key := &model.SigningKey{
		KID:       strings.ToLower(alg) + "-" + hex.EncodeToString(suffix[:]),
		Algorithm: alg,
		Sealed:    sealed,
		CreatedAt: now,
	}

	// "Due" means no unretired key made in the last rotation period — or,
	// for a first key, none at all.
	dueBefore := now.Add(-ks.rotation)
	if ks.rotation == 0 {
		dueBefore = time.Time{}
	}

	added, err := ks.store.AddSigningKeyIfDue(ctx, key, dueBefore)
	if err == nil && added && ks.log != nil {
		ks.log.Info("made a signing key", "kid", key.KID, "algorithm", alg)
	}

	return err
}

// Rotate makes a new key for every algorithm now. With `immediate`, the new
// keys sign from now on and the old ones are retired at once; otherwise they
// wait the usual lead. With `revoke`, the old keys are deleted and no longer
// published, so APIs stop accepting the tokens they signed — for a key that
// may have leaked.
func (ks *keySet) Rotate(ctx context.Context, immediate, revoke bool) error {
	// Postgres keeps microseconds; comparing a nanosecond time with the stored
	// one would find the new key older than itself and retire it.
	now := ks.now().Truncate(time.Microsecond)
	made := []string{}

	for _, alg := range jose.Algorithms {
		private, err := jose.Generate(alg)
		if err != nil {
			return err
		}

		sealed, err := ks.sealer.Seal(private)
		if err != nil {
			return err
		}

		var suffix [6]byte
		if _, err := rand.Read(suffix[:]); err != nil {
			return err
		}

		created := now
		if immediate || revoke {
			// Made "lead ago", so it signs straight away.
			created = now.Add(-keyPublishLead)
		}

		key := &model.SigningKey{
			KID:       strings.ToLower(alg) + "-" + hex.EncodeToString(suffix[:]),
			Algorithm: alg,
			Sealed:    sealed,
			CreatedAt: created,
		}
		if err := ks.store.CreateSigningKey(ctx, key); err != nil {
			return err
		}
		made = append(made, key.KID)

		if immediate || revoke {
			if err := ks.store.RetireSigningKeys(ctx, alg, created, now); err != nil {
				return err
			}
		}
	}

	if revoke {
		if err := ks.store.DeleteSigningKeysExcept(ctx, made); err != nil {
			return err
		}
	}

	return ks.reload(ctx)
}

func (ks *keySet) reload(ctx context.Context) error {
	stored, err := ks.store.SigningKeys(ctx)
	if err != nil {
		return fmt.Errorf("load signing keys: %w", err)
	}

	all := make([]storedKey, 0, len(stored))
	byID := make(map[string]jose.Key, len(stored))

	for _, row := range stored {
		private, err := ks.sealer.Open(row.Sealed)
		if err != nil {
			return err
		}

		key := storedKey{
			Key:       jose.Key{ID: row.KID, Algorithm: row.Algorithm, Private: private},
			CreatedAt: row.CreatedAt,
			RetiredAt: row.RetiredAt,
		}
		all = append(all, key)
		byID[key.ID] = key.Key
	}

	ks.mu.Lock()
	ks.all, ks.byID, ks.loadedAt = all, byID, ks.now()
	ks.mu.Unlock()

	return nil
}

// newestAndSigning finds, for an algorithm, the newest unretired key and the
// one to sign with: the newest that has been published for the lead — or,
// when none has, the oldest unretired one, which is what a fresh installation
// has.
func (ks *keySet) newestAndSigning(alg string, now time.Time) (newest, signing *storedKey) {
	ks.mu.RLock()
	defer ks.mu.RUnlock()

	var oldest *storedKey
	for i := range ks.all {
		key := &ks.all[i]
		if key.Algorithm != alg || key.RetiredAt != nil {
			continue
		}

		if newest == nil {
			newest = key
		}
		if signing == nil && now.Sub(key.CreatedAt) >= keyPublishLead {
			signing = key
		}
		oldest = key
	}

	if signing == nil {
		signing = oldest
	}

	return newest, signing
}

// signing returns the key to sign with for an algorithm.
func (ks *keySet) signing(alg string) (jose.Key, bool) {
	_, key := ks.newestAndSigning(alg, ks.now())
	if key == nil {
		return jose.Key{}, false
	}

	return key.Key, true
}

// lookup returns a key by id, for checking a token. An unknown id reloads the
// keys first, at most once per keyReloadInterval.
func (ks *keySet) lookup(ctx context.Context) func(kid string) (jose.Key, bool) {
	return func(kid string) (jose.Key, bool) {
		ks.mu.RLock()
		key, ok := ks.byID[kid]
		stale := ks.now().Sub(ks.loadedAt) > keyReloadInterval
		ks.mu.RUnlock()

		if ok || !stale {
			return key, ok
		}

		if err := ks.reload(ctx); err != nil {
			return jose.Key{}, false
		}

		ks.mu.RLock()
		defer ks.mu.RUnlock()
		key, ok = ks.byID[kid]
		return key, ok
	}
}

// public returns every published key — signing, waiting and retired — as a
// JWK, reloading first when the set is stale.
func (ks *keySet) public(ctx context.Context) ([]jose.JWK, error) {
	ks.mu.RLock()
	stale := ks.now().Sub(ks.loadedAt) > keyReloadInterval
	ks.mu.RUnlock()

	if stale {
		if err := ks.reload(ctx); err != nil {
			return nil, err
		}
	}

	ks.mu.RLock()
	defer ks.mu.RUnlock()

	out := make([]jose.JWK, 0, len(ks.all))
	for _, key := range ks.all {
		jwk, err := key.Public()
		if err != nil {
			return nil, err
		}
		out = append(out, jwk)
	}

	return out, nil
}

// KeyInfo describes one signing key, for administrators.
type KeyInfo struct {
	KID       string     `json:"kid"`
	Algorithm string     `json:"algorithm"`
	State     string     `json:"state"` // "signing", "next" or "retired"
	CreatedAt time.Time  `json:"created_at"`
	SignsFrom time.Time  `json:"signs_from"`
	RetiredAt *time.Time `json:"retired_at"`
	DeletedAt *time.Time `json:"deleted_at"`
}

// describe lists the keys, newest first.
func (ks *keySet) describe() []KeyInfo {
	now := ks.now()
	signingIDs := map[string]bool{}
	for _, alg := range jose.Algorithms {
		if _, key := ks.newestAndSigning(alg, now); key != nil {
			signingIDs[key.ID] = true
		}
	}

	ks.mu.RLock()
	defer ks.mu.RUnlock()

	out := make([]KeyInfo, 0, len(ks.all))
	for _, key := range ks.all {
		info := KeyInfo{
			KID:       key.ID,
			Algorithm: key.Algorithm,
			CreatedAt: key.CreatedAt,
			SignsFrom: key.CreatedAt.Add(keyPublishLead),
			RetiredAt: key.RetiredAt,
		}

		switch {
		case key.RetiredAt != nil:
			info.State = "retired"
			deleted := key.RetiredAt.Add(keyRetention)
			info.DeletedAt = &deleted
		case signingIDs[key.ID]:
			info.State = "signing"
		default:
			info.State = "next"
		}

		out = append(out, info)
	}

	return out
}
