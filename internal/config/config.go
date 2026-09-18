// Package config reads the settings in .env.
package config

import (
	"errors"
	"fmt"
	"io/fs"
	"net/url"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config is every setting the server has.
//
// There is nothing here about the first administrator: that account is made
// on /admin/new-super-admin the first time the panel is opened, so no
// password is ever written in a file.
type Config struct {
	// Addr is the public listener: the OAuth 2.0 / OpenID Connect provider
	// and the account API the id app calls. It is meant to face the
	// internet, behind a reverse proxy.
	Addr string

	// AdminAddr is the admin listener: the admin API, and nothing else. It is
	// a separate listener so the admin API is never on the public one — bind
	// it to an internal interface, or leave its port unpublished, and only
	// the network the console runs on can reach it.
	AdminAddr string

	// Issuer is this server's public URL: the iss of every token it signs,
	// and the base of every URL in its discovery document. It is the address
	// clients reach the provider at — in the recommended setup, the id app's
	// own origin, which the reverse proxy routes /oauth2 and /.well-known
	// from.
	Issuer string

	// AccountURL is where the id app is served: the users' app, with the
	// sign-in pages and account management. The authorization endpoint sends
	// users there to sign in, and the password reset email links there. It
	// defaults to the issuer, since both are one origin.
	AccountURL string

	// AdminURL is where the console is served. Only requests from its origin
	// may change anything through the admin API.
	AdminURL string

	// CORSOrigins are extra browser origins allowed to call the cookie
	// authenticated APIs from script, with credentials. With each app served
	// on the same origin as the API it calls, none is needed, and none is the
	// default: every origin listed can act as whoever is signed in.
	CORSOrigins []string

	// SecureUserCookies and SecureAdminCookies set the Secure flag on each
	// session cookie. They follow the scheme of AccountURL and AdminURL, so a
	// deployment on https gets them without remembering to; the
	// XERMESS_SECURE_COOKIES setting overrides both.
	SecureUserCookies  bool
	SecureAdminCookies bool

	// KeyRotation is how long a token signing key signs before the next one
	// takes over. Zero never rotates on its own.
	KeyRotation time.Duration

	// AdminMFARequired makes every administrator sign in with a second factor.
	// It is what a fresh installation starts with — off, so the first
	// administrator can get in and turn it on deliberately — and after that
	// the setting lives in the database, where a super admin owns it.
	// An administrator without one is made to set it up at their next sign-in,
	// before they can do anything else.
	AdminMFARequired bool

	// RateLimit is how many attempts per minute one address may make at the
	// endpoints that take a password or send an email. Zero turns the limit
	// off.
	RateLimit int

	// TrustedProxies are the addresses of the proxies in front of the server,
	// whose X-Forwarded-For header is believed about who is calling. Empty
	// believes no one, which is right when nothing sits in front: otherwise
	// any caller could write whatever address it liked into the activity log.
	// Behind a proxy it has to be set, or every request — and so the rate
	// limit — counts as the proxy's.
	TrustedProxies []string

	// SecretKey encrypts the private keys tokens are signed with before they
	// are stored. Losing it makes those keys unreadable, which signs every
	// user out of every application; leaking it with the database lets
	// anyone sign tokens.
	SecretKey string

	Mail Mail

	DB DB
}

// Origin is the scheme and host of a URL, which is what a browser sends in
// the Origin header.
func Origin(raw string) string {
	parsed, err := url.Parse(raw)
	if err != nil {
		return ""
	}

	return parsed.Scheme + "://" + parsed.Host
}

// minSecretKeyLength is the shortest secret key accepted: 32 characters, which
// `openssl rand -base64 32` comfortably exceeds.
const minSecretKeyLength = 32

// Mail is how email is sent. With no host, nothing is sent: each message is
// written to the log instead, which is enough to follow a reset link while
// developing.
type Mail struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
}

// DB is the database connection and migration settings.
type DB struct {
	Driver     string
	DSN        string
	TimeZone   string
	LogQueries bool
	Migrate    bool
	MigrateDir string
}

// read builds the reader both loaders use: .env first, then the environment,
// which wins. A missing .env is fine: in a container there are only
// environment variables.
func read() (*viper.Viper, error) {
	v := viper.New()
	v.SetConfigFile(".env")
	v.SetConfigType("env")
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return nil, err
	}

	return v, nil
}

// Load reads every setting the server needs.
func Load() (Config, error) {
	v, err := read()
	if err != nil {
		return Config{}, err
	}

	v.SetDefault("XERMESS_ADDR", ":8080")
	v.SetDefault("XERMESS_ADMIN_ADDR", ":8081")
	v.SetDefault("XERMESS_CORS_ORIGINS", "")
	v.SetDefault("XERMESS_TRUSTED_PROXIES", "")
	v.SetDefault("XERMESS_RATE_LIMIT", 20)
	v.SetDefault("XERMESS_ADMIN_MFA", "optional")
	v.SetDefault("XERMESS_KEY_ROTATION_DAYS", 90)
	// In development the id app serves the provider on its own origin,
	// through Vite's proxy, the same way the reverse proxy does in
	// production.
	v.SetDefault("XERMESS_ISSUER", "http://localhost:5173")
	v.SetDefault("XERMESS_ADMIN_URL", "http://localhost:5174")
	v.SetDefault("XERMESS_SMTP_PORT", 587)
	v.SetDefault("XERMESS_SMTP_FROM", "xermess <no-reply@localhost>")
	v.SetDefault("XERMESS_DB_DRIVER", "postgres")
	v.SetDefault("XERMESS_DB_TIMEZONE", "Asia/Bishkek")
	v.SetDefault("XERMESS_DB_LOG_QUERIES", false)
	v.SetDefault("XERMESS_DB_MIGRATE", true)
	v.SetDefault("XERMESS_DB_MIGRATE_DIR", "./migrations")

	issuer := strings.TrimRight(v.GetString("XERMESS_ISSUER"), "/")
	accountURL := strings.TrimRight(v.GetString("XERMESS_ACCOUNT_URL"), "/")
	if accountURL == "" {
		accountURL = issuer
	}
	adminURL := strings.TrimRight(v.GetString("XERMESS_ADMIN_URL"), "/")

	cfg := Config{
		Addr:               v.GetString("XERMESS_ADDR"),
		AdminAddr:          v.GetString("XERMESS_ADMIN_ADDR"),
		Issuer:             issuer,
		AccountURL:         accountURL,
		AdminURL:           adminURL,
		CORSOrigins:        splitList(v.GetString("XERMESS_CORS_ORIGINS")),
		SecureUserCookies:  strings.HasPrefix(accountURL, "https://"),
		SecureAdminCookies: strings.HasPrefix(adminURL, "https://"),
		AdminMFARequired:   v.GetString("XERMESS_ADMIN_MFA") == "required",
		KeyRotation:        time.Duration(v.GetInt("XERMESS_KEY_ROTATION_DAYS")) * 24 * time.Hour,
		RateLimit:          v.GetInt("XERMESS_RATE_LIMIT"),
		TrustedProxies:     splitList(v.GetString("XERMESS_TRUSTED_PROXIES")),
		SecretKey:          v.GetString("XERMESS_SECRET_KEY"),
		Mail: Mail{
			Host:     v.GetString("XERMESS_SMTP_HOST"),
			Port:     v.GetInt("XERMESS_SMTP_PORT"),
			Username: v.GetString("XERMESS_SMTP_USERNAME"),
			Password: v.GetString("XERMESS_SMTP_PASSWORD"),
			From:     v.GetString("XERMESS_SMTP_FROM"),
		},
		DB: DB{
			Driver:     v.GetString("XERMESS_DB_DRIVER"),
			DSN:        v.GetString("XERMESS_DB_DSN"),
			TimeZone:   v.GetString("XERMESS_DB_TIMEZONE"),
			LogQueries: v.GetBool("XERMESS_DB_LOG_QUERIES"),
			Migrate:    v.GetBool("XERMESS_DB_MIGRATE"),
			MigrateDir: v.GetString("XERMESS_DB_MIGRATE_DIR"),
		},
	}

	if v.IsSet("XERMESS_SECURE_COOKIES") && v.GetString("XERMESS_SECURE_COOKIES") != "" {
		cfg.SecureUserCookies = v.GetBool("XERMESS_SECURE_COOKIES")
		cfg.SecureAdminCookies = cfg.SecureUserCookies
	}

	// No default for the DSN: it names the database and carries the password,
	// so a missing one should stop the server, not quietly connect somewhere.
	if cfg.DB.DSN == "" {
		return Config{}, errors.New("XERMESS_DB_DSN is not set")
	}

	// No default for the secret key either: a default would be the same key
	// on every installation, which is no secret at all.
	if len(cfg.SecretKey) < minSecretKeyLength {
		return Config{}, fmt.Errorf("XERMESS_SECRET_KEY must be at least %d characters; make one with: openssl rand -base64 32", minSecretKeyLength)
	}

	urls := map[string]string{
		"XERMESS_ISSUER":      cfg.Issuer,
		"XERMESS_ACCOUNT_URL": cfg.AccountURL,
		"XERMESS_ADMIN_URL":   cfg.AdminURL,
	}
	for name, value := range urls {
		if parsed, err := url.Parse(value); err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" {
			return Config{}, fmt.Errorf("%s must be an http or https URL, got %q", name, value)
		}
	}

	// One listener for both would put the admin API back on the public one,
	// which is exactly what the second listener is there to prevent.
	if cfg.AdminAddr == "" || cfg.AdminAddr == cfg.Addr {
		return Config{}, errors.New("XERMESS_ADMIN_ADDR must be set, and differ from XERMESS_ADDR: the admin API has its own listener")
	}

	if mode := v.GetString("XERMESS_ADMIN_MFA"); mode != "required" && mode != "optional" {
		return Config{}, fmt.Errorf("XERMESS_ADMIN_MFA must be required or optional, got %q", mode)
	}

	if cfg.KeyRotation < 0 {
		return Config{}, errors.New("XERMESS_KEY_ROTATION_DAYS must be zero or more")
	}

	if cfg.RateLimit < 0 {
		return Config{}, errors.New("XERMESS_RATE_LIMIT must be zero or more")
	}

	return cfg, nil
}

// splitList reads "a,b,c" into a slice.
func splitList(s string) []string {
	var out []string
	for _, part := range strings.Split(s, ",") {
		if p := strings.TrimSpace(part); p != "" {
			out = append(out, p)
		}
	}
	return out
}
