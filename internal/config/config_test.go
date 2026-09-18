package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// clearEnv unsets every XERMESS_ variable for the duration of the test, so a
// developer's own environment cannot change the result. Viper treats an empty
// variable as unset.
func clearEnv(t *testing.T) {
	t.Helper()

	for _, entry := range os.Environ() {
		if name, _, found := strings.Cut(entry, "="); found && strings.HasPrefix(name, "XERMESS_") {
			t.Setenv(name, "")
		}
	}
}

// writeEnv puts a .env in a temporary directory and makes it the working
// directory, because Load reads ./.env.
func writeEnv(t *testing.T, contents string) {
	t.Helper()

	clearEnv(t)

	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, ".env"), []byte(contents), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Chdir(dir)
}

func TestLoadReadsEnvFile(t *testing.T) {
	writeEnv(t, `
XERMESS_ADDR=:9000
XERMESS_CORS_ORIGINS=http://a.test, http://b.test
XERMESS_DB_DSN=postgres://user:pw@localhost:5432/mydb
XERMESS_SECRET_KEY=0123456789abcdef0123456789abcdef
XERMESS_ISSUER=https://auth.example.com/
XERMESS_DB_LOG_QUERIES=true
XERMESS_DB_MIGRATE=false
XERMESS_SECURE_COOKIES=true
XERMESS_TRUSTED_PROXIES=10.0.0.1, 192.168.0.0/16
`)

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}

	if !cfg.SecureUserCookies || !cfg.SecureAdminCookies {
		t.Error("XERMESS_SECURE_COOKIES=true did not turn both cookies secure")
	}
	if len(cfg.TrustedProxies) != 2 || cfg.TrustedProxies[1] != "192.168.0.0/16" {
		t.Errorf("TrustedProxies = %v, want both, trimmed", cfg.TrustedProxies)
	}

	if cfg.Issuer != "https://auth.example.com" {
		t.Errorf("Issuer = %q, want it without the trailing slash", cfg.Issuer)
	}

	if cfg.Addr != ":9000" {
		t.Errorf("Addr = %q, want :9000", cfg.Addr)
	}
	if cfg.DB.DSN != "postgres://user:pw@localhost:5432/mydb" {
		t.Errorf("DSN = %q", cfg.DB.DSN)
	}
	if !cfg.DB.LogQueries {
		t.Error("LogQueries = false, want true")
	}
	if cfg.DB.Migrate {
		t.Error("Migrate = true, want false")
	}

	want := []string{"http://a.test", "http://b.test"}
	if len(cfg.CORSOrigins) != len(want) {
		t.Fatalf("CORSOrigins = %v, want %v", cfg.CORSOrigins, want)
	}
	for i, origin := range want {
		if cfg.CORSOrigins[i] != origin {
			t.Errorf("CORSOrigins[%d] = %q, want %q (spaces should be trimmed)", i, cfg.CORSOrigins[i], origin)
		}
	}
}

func TestLoadUsesDefaults(t *testing.T) {
	writeEnv(t, "XERMESS_DB_DSN=postgres://u:p@localhost:5432/db\nXERMESS_SECRET_KEY=0123456789abcdef0123456789abcdef\n")

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}

	if cfg.Addr != ":8080" {
		t.Errorf("Addr = %q, want the default :8080", cfg.Addr)
	}
	if cfg.DB.Driver != "postgres" {
		t.Errorf("Driver = %q, want the default postgres", cfg.DB.Driver)
	}
	if cfg.DB.TimeZone != "Asia/Bishkek" {
		t.Errorf("TimeZone = %q, want the default Asia/Bishkek", cfg.DB.TimeZone)
	}
	if !cfg.DB.Migrate {
		t.Error("Migrate = false, want the default true")
	}
	if cfg.Issuer != "http://localhost:5173" || cfg.AccountURL != cfg.Issuer || cfg.AdminURL != "http://localhost:5174" {
		t.Errorf("Issuer, AccountURL, AdminURL = %q, %q, %q, want the local defaults", cfg.Issuer, cfg.AccountURL, cfg.AdminURL)
	}
	if cfg.AdminAddr != ":8081" || len(cfg.CORSOrigins) != 0 || cfg.RateLimit != 20 {
		t.Errorf("AdminAddr, CORSOrigins, RateLimit = %q, %v, %d, want :8081, none, 20", cfg.AdminAddr, cfg.CORSOrigins, cfg.RateLimit)
	}
	if cfg.SecureUserCookies || cfg.SecureAdminCookies {
		t.Error("cookies are secure on plain http localhost")
	}
	// A fresh installation does not make anybody set up an authenticator:
	// each administrator turns one on for themselves, and a super admin can
	// require them of everybody once there is somebody to ask.
	if cfg.AdminMFARequired {
		t.Error("AdminMFARequired = true, want two-factor sign-in optional by default")
	}
	if cfg.Mail.Host != "" {
		t.Errorf("Mail.Host = %q, want none: email is logged by default", cfg.Mail.Host)
	}
	if cfg.DB.MigrateDir != "./migrations" {
		t.Errorf("MigrateDir = %q, want the default ./migrations", cfg.DB.MigrateDir)
	}
}

// The environment has to win over .env, or a container could not change a
// single setting without rewriting the file.
func TestEnvironmentOverridesEnvFile(t *testing.T) {
	writeEnv(t, "XERMESS_ADDR=:8080\nXERMESS_DB_DSN=postgres://u:p@localhost:5432/from_file\nXERMESS_SECRET_KEY=0123456789abcdef0123456789abcdef\n")
	t.Setenv("XERMESS_ADDR", ":7777")

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}

	if cfg.Addr != ":7777" {
		t.Errorf("Addr = %q, want :7777 from the environment", cfg.Addr)
	}
}

// The secret key encrypts the signing keys, so a missing or short one has to
// stop the server rather than fall back to something guessable.
func TestLoadFailsWithoutSecretKey(t *testing.T) {
	writeEnv(t, "XERMESS_DB_DSN=postgres://u:p@localhost:5432/db\n")
	if _, err := Load(); err == nil {
		t.Fatal("want an error when XERMESS_SECRET_KEY is missing, got none")
	}

	writeEnv(t, "XERMESS_DB_DSN=postgres://u:p@localhost:5432/db\nXERMESS_SECRET_KEY=short\n")
	if _, err := Load(); err == nil {
		t.Fatal("want an error when XERMESS_SECRET_KEY is too short, got none")
	}
}

// Deployed on https, the session cookies are Secure without being told, and
// the account app defaults to the issuer's origin.
func TestLoadDerivesFromPublicURLs(t *testing.T) {
	writeEnv(t, `
XERMESS_DB_DSN=postgres://u:p@localhost:5432/db
XERMESS_SECRET_KEY=0123456789abcdef0123456789abcdef
XERMESS_ISSUER=https://id.mywebsite.com
XERMESS_ADMIN_URL=https://admin-id.mywebsite.com
`)

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}

	if cfg.AccountURL != "https://id.mywebsite.com" {
		t.Errorf("AccountURL = %q, want the issuer", cfg.AccountURL)
	}
	if !cfg.SecureUserCookies || !cfg.SecureAdminCookies {
		t.Error("cookies are not secure on https")
	}
	if Origin("https://id.mywebsite.com/login?x=1") != "https://id.mywebsite.com" {
		t.Error("Origin did not keep only the scheme and host")
	}
}

// The admin API must never share the public listener.
func TestLoadRefusesOneListenerForBoth(t *testing.T) {
	writeEnv(t, "XERMESS_DB_DSN=postgres://u:p@localhost:5432/db\nXERMESS_SECRET_KEY=0123456789abcdef0123456789abcdef\nXERMESS_ADDR=:9000\nXERMESS_ADMIN_ADDR=:9000\n")

	if _, err := Load(); err == nil {
		t.Fatal("want an error when the admin listener is the public one")
	}
}

func TestLoadFailsWithoutDSN(t *testing.T) {
	writeEnv(t, "XERMESS_ADDR=:8080\n")

	if _, err := Load(); err == nil {
		t.Fatal("want an error when XERMESS_DB_DSN is missing, got none")
	}
}

// A missing .env is normal in a container, where only real environment
// variables are set.
func TestLoadWithoutEnvFile(t *testing.T) {
	clearEnv(t)
	t.Chdir(t.TempDir())
	t.Setenv("XERMESS_DB_DSN", "postgres://u:p@localhost:5432/db")
	t.Setenv("XERMESS_SECRET_KEY", "0123456789abcdef0123456789abcdef")

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Addr != ":8080" {
		t.Errorf("Addr = %q, want the default :8080", cfg.Addr)
	}
}
