// Package brand is what this project calls itself, and every name built from
// it: cookies, the Redis prefix, locks. Each app keeps the same list in its
// brand.ts, and brand_test.go checks the two sides agree.
//
// Renaming the project is these constants, the module path in go.mod, each
// app's brand.ts, and a `git mv` of cmd/<name>. A test fails on the name
// spelled anywhere else.
package brand

const (
	// Name is the project as people read it.
	Name = "Loginer"

	// Slug is the name as an identifier: module, environment variables,
	// cookies, Redis keys.
	Slug = "loginer"

	// EnvPrefix starts every setting in .env, so an installation's own
	// variables cannot be mistaken for the server's.
	EnvPrefix = "LOGINER_"

	// DocsURL and GitHubURL are where the project itself lives; the panel's
	// header links to both.
	DocsURL   = "https://loginer.org/docs"
	GitHubURL = "https://github.com/loginer/loginer"
)

// The cookies the server sets and the apps read back. The apps' own cookies
// (theme, sidebar) live in their brand.ts.
const (
	// AdminSessionCookie carries an administrator's session. HttpOnly, so no
	// script in the browser ever reads it.
	AdminSessionCookie = Slug + "_session"

	// UserSessionCookie carries a person's own session for single sign-on
	// between applications. It is deliberately not the administrators' cookie:
	// signing in to an application must never sign anyone in to the panel.
	UserSessionCookie = Slug + "_user_session"

	// SignInStateCookie carries the OAuth state out to a provider and back,
	// which nobody may replay.
	SignInStateCookie = Slug + "_sign_in_state"

	// LanguageCookie is the sign-in pages' language, so the provider writes
	// emails in it, including on a social callback.
	LanguageCookie = Slug + "-account-language"
)

// The audiences of the admin and account APIs. They are URNs so they never
// change with an installation's addresses.
const (
	AdminAPIIdentifier   = "urn:" + Slug + ":admin-api"
	AccountAPIIdentifier = "urn:" + Slug + ":account-api"
)

// RedisPrefix starts every key this server writes, so one Redis can serve
// several installations without their keys meeting.
const RedisPrefix = Slug + ":"

// Realm is what an unauthorized answer names as the thing being asked for, in
// a WWW-Authenticate header.
const Realm = Name

// FlowSchema is the $schema of exported login flows; changing it orphans every
// exported file.
const FlowSchema = Slug + ".login-flow/1"

// The two advisory locks, so that two servers racing each other do not both
// make the first administrator or both rotate the signing keys.
const (
	// AdminLock is any number nothing else locks, and the letters of the slug
	// are as good as any: pg_advisory_lock takes a bigint, not a name.
	AdminLock int64 = 0x6c6f67696e6572 // "loginer"

	// SigningKeysLock is named rather than numbered, because this one is worth
	// being able to read in pg_locks.
	SigningKeysLock = Slug + "_signing_keys"
)
