// Package brand is what this project calls itself.
//
// Every name that reaches a browser, an email, a database or someone's .env
// is written here and nowhere else, and the names that are built from the
// name — the cookies, the Redis prefix, the locks — are built here too, so
// they cannot drift apart. The two apps keep the same list in their own
// brand.ts, and brand_test.go reads those to check the two sides still agree.
//
// So renaming the project is these constants, the module path on the first
// line of go.mod, one brand.ts in each app, and a `git mv` of the command's own
// directory under cmd/. Nothing else in the repository spells the name — the
// test fails the build on one that does, which is what makes that list enough
// rather than merely short.
package brand

const (
	// Name is the project as a person reads it: the wordmark in the panel,
	// the suffix of a page title, the organization a fresh installation
	// starts with, the display name on a mail it sends.
	Name = "Loginer"

	// Slug is the same name as an identifier: what the Go module, the
	// environment variables, the cookies and the Redis keys are built from.
	// Lower case, because a cookie and a key are not read as prose.
	Slug = "loginer"

	// EnvPrefix starts every setting in .env, so an installation's own
	// variables cannot be mistaken for the server's.
	EnvPrefix = "LOGINER_"

	// DocsURL and GitHubURL are where the project itself lives, rather than
	// this installation of it: the documentation, and the source. The panel's
	// header links to both.
	DocsURL   = "https://loginer.org/docs"
	GitHubURL = "https://github.com/loginer/loginer"
)

// The cookies the server sets and the sign-in pages read back. Each is the
// slug with what it carries on the end.
//
// The first three are named here because the API writes them and a browser
// has to send them back; LanguageCookie is read by the provider as well. The
// two apps' own cookies — the panel's theme and sidebar, the account's theme
// — are nobody else's business and live in their brand.ts.
const (
	// AdminSessionCookie carries an administrator's session. HttpOnly, so no
	// script in the browser ever reads it.
	AdminSessionCookie = Slug + "_session"

	// UserSessionCookie carries a person's own session, which is what lets the
	// authorization endpoint sign them in to a second application without
	// asking for their password again. A different cookie from the
	// administrators' on purpose: signing in to an application must never sign
	// anyone in to the panel.
	UserSessionCookie = Slug + "_user_session"

	// SignInStateCookie carries the OAuth state out to a provider and back,
	// which nobody may replay.
	SignInStateCookie = Slug + "_sign_in_state"

	// LanguageCookie is the language the sign-in pages were shown in, so the
	// provider can write its email in it. The pages and the provider are one
	// origin, so even the callback from a social provider, which has no body
	// to say it in, carries it.
	LanguageCookie = Slug + "-account-language"
)

// RedisPrefix starts every key this server writes, so one Redis can serve
// several installations without their keys meeting.
const RedisPrefix = Slug + ":"

// Realm is what an unauthorized answer names as the thing being asked for, in
// a WWW-Authenticate header.
const Realm = Name

// FlowSchema is the $schema an exported login flow claims. It is written into
// the files the panel exports, so changing it makes every file this version
// exported look like it belongs to another schema.
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
