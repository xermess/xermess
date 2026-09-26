package api

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"

	"loginer/i18n"
	"loginer/internal/brand"
	"loginer/internal/cache"
	"loginer/internal/cache/cachetest"
	"loginer/internal/config"
	"loginer/internal/database"
	"loginer/internal/mail"
	"loginer/internal/model"
	"loginer/internal/oidc"
	"loginer/internal/store"
)

// These tests run the whole server — routes, permission checks, store and
// migrations — against a real Postgres. They need a database server to make
// throwaway databases on, named by LOGINER_TEST_DB_DSN:
//
//	LOGINER_TEST_DB_DSN=postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable go test ./internal/api/
//
// Without it they are skipped, so `go test ./...` still passes anywhere.

// testDSNEnv names the database server the integration tests may use.
const testDSNEnv = "LOGINER_TEST_DB_DSN"

// liveServer is the server running on a database of its own.
type liveServer struct {
	t   *testing.T
	url string
	// root is the public server's address: the issuer of its tokens.
	root string
	// adminRoot is the admin server's.
	adminRoot string
	mail      *mailbox
	// store is the server's own, for a test that checks what it keeps.
	store *store.Store
	// cache is the Redis the server reads through, or nil when the tests
	// have none.
	cache *cache.Cache
}

// testAccountURL is where the provider sends browsers to sign in. Nothing is
// served there in the tests: they read the redirect and call the account
// endpoints the page would.
const testAccountURL = "http://account.test"

// mailbox keeps what the server would have emailed.
type mailbox struct {
	mu       sync.Mutex
	messages []mail.Message
}

func (m *mailbox) Send(_ context.Context, msg mail.Message) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.messages = append(m.messages, msg)
	return nil
}

// wait returns the first message sent to `to`, waiting briefly: the server
// sends in the background.
func (m *mailbox) wait(t *testing.T, to string) mail.Message {
	t.Helper()

	for deadline := time.Now().Add(3 * time.Second); time.Now().Before(deadline); time.Sleep(20 * time.Millisecond) {
		m.mu.Lock()
		for _, msg := range m.messages {
			if strings.EqualFold(msg.To, to) {
				m.mu.Unlock()
				return msg
			}
		}
		m.mu.Unlock()
	}

	t.Fatalf("no email to %s", to)
	return mail.Message{}
}

// linkIn is the link a message carries to one of the sign-in pages —
// "/verify-email", "/reset-password" — parsed. The body is plain text, so the
// link runs to the first space or newline after it.
func linkIn(t *testing.T, body, page string) *url.URL {
	t.Helper()

	at := strings.Index(body, testAccountURL+page)
	if at < 0 {
		t.Fatalf("no %s link in the message:\n%s", page, body)
	}

	link, err := url.Parse(strings.Fields(body[at:])[0])
	if err != nil {
		t.Fatal(err)
	}

	return link
}

// tokenIn is that link's token, which is what an endpoint is given to use it.
func tokenIn(t *testing.T, body, page string) string {
	t.Helper()

	token := linkIn(t, body, page).Query().Get("token")
	if token == "" {
		t.Fatalf("the %s link carries no token:\n%s", page, body)
	}

	return token
}

// newLiveServer makes an empty database, migrates it, and serves the API on
// it, without a rate limit. The database is dropped when the test ends.
func newLiveServer(t *testing.T) *liveServer {
	t.Helper()
	return newLiveServerLimited(t, 0)
}

// newLiveServerLimited is newLiveServer with a rate limit, per minute.
func newLiveServerLimited(t *testing.T, limit int) *liveServer {
	t.Helper()
	return newLiveServerWith(t, func(cfg *config.Config) { cfg.RateLimit = limit })
}

// newLiveServerWith is newLiveServer with the configuration changed first.
func newLiveServerWith(t *testing.T, change func(*config.Config)) *liveServer {
	t.Helper()

	dsn := os.Getenv(testDSNEnv)
	if dsn == "" {
		t.Skipf("%s is not set", testDSNEnv)
	}

	log := slog.New(slog.NewTextHandler(io.Discard, nil))

	admin, err := database.Open(config.DB{Driver: "postgres", DSN: dsn, TimeZone: "UTC"})
	if err != nil {
		t.Fatalf("open the test database server: %v", err)
	}
	t.Cleanup(func() { _ = database.Close(admin) })

	name := brand.Slug + "_test_" + randomHex(t, 6)
	if err := admin.Exec("CREATE DATABASE " + name).Error; err != nil {
		t.Fatalf("create %s: %v", name, err)
	}

	cfg := config.DB{Driver: "postgres", DSN: withDatabase(t, dsn, name), TimeZone: "UTC", MigrateDir: migrationsDir(t)}

	db, err := database.Open(cfg)
	if err != nil {
		t.Fatalf("open %s: %v", name, err)
	}

	t.Cleanup(func() {
		_ = database.Close(db)
		_ = admin.Exec("DROP DATABASE IF EXISTS " + name + " WITH (FORCE)").Error
	})

	if err := database.Migrate(db, cfg, log); err != nil {
		t.Fatalf("migrate %s: %v", name, err)
	}

	// The issuer has to be the address the server answers on, which is only
	// known once it has a listener: so the listeners come first, and the
	// routers are built to fit them. Like the real process, it is two
	// servers: the public one, and the admin one.
	var publicRouter, adminRouter http.Handler
	publicServer := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		publicRouter.ServeHTTP(w, r)
	}))
	adminServer := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		adminRouter.ServeHTTP(w, r)
	}))
	root := "http://" + publicServer.Listener.Addr().String()
	adminRoot := "http://" + adminServer.Listener.Addr().String()

	mailer := &mailbox{}
	serverCfg := config.Config{
		DB:         cfg,
		Issuer:     root,
		AccountURL: testAccountURL,
		AdminURL:   "http://admin.test",
		SecretKey:  "integration-test-secret-key-0123456789",
	}
	change(&serverCfg)

	// With LOGINER_TEST_REDIS the whole server runs with its cache, under a
	// prefix of its own, so every test here also proves that a write is seen
	// by the next read through the cache.
	shared := cachetest.Open(t)
	st := store.New(db).WithCache(shared)

	// What main does on a fresh installation: write the settings for
	// administrators' sign-ins from the configuration, so these servers
	// behave the way a started one does.
	if err := st.EnsureAdminSecurity(context.Background(), serverCfg.AdminMFARequired); err != nil {
		t.Fatal(err)
	}

	shipped, err := i18n.Shipped()
	if err != nil {
		t.Fatal(err)
	}
	if err := st.EnsureLanguages(context.Background(), shipped); err != nil {
		t.Fatal(err)
	}

	provider, err := oidc.New(context.Background(), serverCfg, st, mailer, log)
	if err != nil {
		t.Fatal(err)
	}

	publicEngine, err := NewPublic(serverCfg, log, provider, shared)
	if err != nil {
		t.Fatal(err)
	}
	adminEngine, err := NewAdmin(serverCfg, st, log, provider, shared)
	if err != nil {
		t.Fatal(err)
	}
	publicRouter, adminRouter = publicEngine, adminEngine

	publicServer.Start()
	adminServer.Start()
	t.Cleanup(publicServer.Close)
	t.Cleanup(adminServer.Close)

	return &liveServer{t: t, url: adminRoot + "/api/v1/admin", root: root, adminRoot: adminRoot, mail: mailer, store: st, cache: shared}
}

// client is one browser: it keeps its own session cookie.
type client struct {
	s    *liveServer
	http *http.Client
}

func (s *liveServer) client() *client {
	jar, err := cookiejar.New(nil)
	if err != nil {
		s.t.Fatal(err)
	}

	return &client{s: s, http: &http.Client{Jar: jar}}
}

// do sends a JSON request and decodes the JSON answer into `out`, when given.
func (c *client) do(method, path string, body any, out any) int {
	c.s.t.Helper()

	var reader io.Reader
	if body != nil {
		encoded, err := json.Marshal(body)
		if err != nil {
			c.s.t.Fatal(err)
		}
		reader = bytes.NewReader(encoded)
	}

	req, err := http.NewRequest(method, c.s.url+path, reader)
	if err != nil {
		c.s.t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	res, err := c.http.Do(req)
	if err != nil {
		c.s.t.Fatal(err)
	}
	defer res.Body.Close()

	if out != nil {
		if err := json.NewDecoder(res.Body).Decode(out); err != nil && err != io.EOF {
			c.s.t.Fatalf("%s %s: decode: %v", method, path, err)
		}
	}

	return res.StatusCode
}

// must sends a request that has to answer with `want`.
func (c *client) must(want int, method, path string, body any, out any) {
	c.s.t.Helper()

	var raw json.RawMessage
	status := c.do(method, path, body, &raw)
	if status != want {
		c.s.t.Fatalf("%s %s = %d %s, want %d", method, path, status, raw, want)
	}

	if out != nil && len(raw) > 0 {
		if err := json.Unmarshal(raw, out); err != nil {
			c.s.t.Fatal(err)
		}
	}
}

func (c *client) login(email, password string) int {
	return c.do(http.MethodPost, "/auth/login", map[string]string{"username": email, "password": password}, nil)
}

const superEmail, superPassword = "root@example.com", "root-password-1"

// superAdmin sets the panel up and signs its first administrator in.
func (s *liveServer) superAdmin() *client {
	c := s.client()
	c.must(http.StatusCreated, http.MethodPost, "/setup", map[string]string{
		"email": superEmail, "password": superPassword, "first_name": "Root",
	}, nil)

	if status := c.login(superEmail, superPassword); status != http.StatusOK {
		s.t.Fatalf("super admin login = %d", status)
	}

	return c
}

type idOnly struct {
	ID string `json:"id"`
}

// application registers a web application and returns its id.
func (c *client) application(name string) string {
	var out struct {
		Application idOnly `json:"application"`
	}
	c.must(http.StatusCreated, http.MethodPost, "/applications", map[string]any{
		"name": name, "type": "web",
		"grant_types":   []string{"authorization_code"},
		"redirect_uris": []string{"https://" + name + ".example.com/callback"},
		"scopes":        []string{"openid"},
	}, &out)

	return out.Application.ID
}

// role creates a user role — global when app is "" — and returns its id.
func (c *client) role(name, app string, inherits ...string) string {
	body := map[string]any{"name": name, "inherits": inherits}
	if app != "" {
		body["application_id"] = app
	}

	var out struct {
		Role idOnly `json:"role"`
	}
	c.must(http.StatusCreated, http.MethodPost, "/user-roles", body, &out)

	return out.Role.ID
}

// adminRoleID finds a seeded admin role by name.
func (c *client) adminRoleID(name string) string {
	var out struct {
		Roles []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"roles"`
	}
	c.must(http.StatusOK, http.MethodGet, "/admin-roles", nil, &out)

	for _, role := range out.Roles {
		if role.Name == name {
			return role.ID
		}
	}

	c.s.t.Fatalf("no admin role %q", name)
	return ""
}

// appManager creates an administrator holding app_manager for one application
// and signs them in.
func (s *liveServer) appManager(super *client, app string) *client {
	const email, password = "manager@example.com", "manager-password-1"

	super.must(http.StatusCreated, http.MethodPost, "/admins", map[string]any{
		"email": email, "first_name": "Manager", "status": "active",
		"password": password, "confirm_password": password,
		"assignments": []map[string]any{{"role_id": super.adminRoleID("app_manager"), "application_id": app}},
	}, nil)

	c := s.client()
	if status := c.login(email, password); status != http.StatusOK {
		s.t.Fatalf("app manager login = %d", status)
	}

	return c
}

func TestLiveFirstAdminOnlyOnce(t *testing.T) {
	s := newLiveServer(t)
	s.superAdmin()

	status := s.client().do(http.MethodPost, "/setup", map[string]string{
		"email": "second@example.com", "password": "second-password", "first_name": "Second",
	}, nil)
	if status != http.StatusConflict {
		t.Errorf("second setup = %d, want 409", status)
	}
}

func TestLiveLoginLockout(t *testing.T) {
	s := newLiveServer(t)
	s.superAdmin()

	c := s.client()
	for range 5 {
		if status := c.login(superEmail, "wrong-password"); status != http.StatusUnauthorized {
			t.Fatalf("wrong password = %d, want 401", status)
		}
	}

	if status := c.login(superEmail, superPassword); status != http.StatusUnauthorized {
		t.Errorf("right password on a locked account = %d, want 401", status)
	}
}

func TestLiveLoginResetsFailures(t *testing.T) {
	s := newLiveServer(t)
	s.superAdmin()

	c := s.client()
	for range 4 {
		c.login(superEmail, "wrong-password")
	}
	if status := c.login(superEmail, superPassword); status != http.StatusOK {
		t.Fatalf("login after four failures = %d, want 200", status)
	}

	// The count started again, so four more are not a lock either.
	for range 4 {
		c.login(superEmail, "wrong-password")
	}
	if status := c.login(superEmail, superPassword); status != http.StatusOK {
		t.Errorf("login after four more failures = %d, want 200", status)
	}
}

// An administrator changes their own account from the profile: a name
// freely, a new address and a new password only with the one they have. A
// new password signs every other browser out and keeps the one it came from.
func TestLiveOwnAccount(t *testing.T) {
	s := newLiveServer(t)
	here := s.superAdmin()

	elsewhere := s.client()
	if status := elsewhere.login(superEmail, superPassword); status != http.StatusOK {
		t.Fatalf("second sign-in = %d", status)
	}

	var answer struct {
		Admin struct {
			Email     string `json:"email"`
			FirstName string `json:"first_name"`
			LastName  string `json:"last_name"`
			FullName  string `json:"full_name"`
		} `json:"admin"`
	}
	here.must(http.StatusOK, http.MethodPatch, "/me", map[string]string{
		"first_name": "Ada", "last_name": "Lovelace", "email": superEmail,
	}, &answer)
	if answer.Admin.FullName != "Ada Lovelace" || answer.Admin.LastName != "Lovelace" {
		t.Errorf("after renaming = %+v, want Ada Lovelace", answer.Admin)
	}

	const moved = "ada@example.com"
	for _, tt := range []struct {
		name     string
		password string
		want     int
	}{
		{"a new address without the password", "", http.StatusBadRequest},
		{"a new address with the wrong password", "not-the-password", http.StatusBadRequest},
		{"a new address with the password", superPassword, http.StatusOK},
	} {
		status := here.do(http.MethodPatch, "/me", map[string]string{
			"first_name": "Ada", "email": moved, "current_password": tt.password,
		}, nil)
		if status != tt.want {
			t.Errorf("%s = %d, want %d", tt.name, status, tt.want)
		}
	}

	if status := s.client().login(moved, superPassword); status != http.StatusOK {
		t.Errorf("sign-in with the new address = %d, want 200", status)
	}

	const next = "a-brand-new-password"
	if status := here.do(http.MethodPost, "/me/password", map[string]string{
		"current_password": "not-the-password", "new_password": next,
	}, nil); status != http.StatusBadRequest {
		t.Errorf("new password with the wrong current one = %d, want 400", status)
	}
	if status := here.do(http.MethodPost, "/me/password", map[string]string{
		"current_password": superPassword, "new_password": "short",
	}, nil); status != http.StatusBadRequest {
		t.Errorf("a new password that is too short = %d, want 400", status)
	}
	here.must(http.StatusOK, http.MethodPost, "/me/password", map[string]string{
		"current_password": superPassword, "new_password": next,
	}, nil)

	if status := here.do(http.MethodGet, "/me", nil, nil); status != http.StatusOK {
		t.Errorf("the session that changed the password = %d, want still signed in", status)
	}
	if status := elsewhere.do(http.MethodGet, "/me", nil, nil); status != http.StatusUnauthorized {
		t.Errorf("another session after a new password = %d, want 401", status)
	}
	if status := s.client().login(moved, next); status != http.StatusOK {
		t.Errorf("sign-in with the new password = %d, want 200", status)
	}
}

// An administrator of one application manages its roles, but cannot reach
// the global roles or another application's through them.
func TestLiveScopedAdminStaysInTheirApplication(t *testing.T) {
	s := newLiveServer(t)
	super := s.superAdmin()

	shop := super.application("shop")
	blog := super.application("blog")
	employee := super.role("employee", "")
	blogWriter := super.role("writer", blog)

	manager := s.appManager(super, shop)

	// Their own application's roles are theirs to make.
	viewer := manager.role("viewer", shop)
	manager.role("editor", shop, viewer)

	// Including a global role would hand it to everyone holding the role.
	status := manager.do(http.MethodPost, "/user-roles", map[string]any{
		"name": "staff", "application_id": shop, "inherits": []string{employee},
	}, nil)
	if status != http.StatusForbidden {
		t.Errorf("app role including a global role = %d, want 403", status)
	}

	// Another application's roles are not even there.
	status = manager.do(http.MethodPost, "/user-roles", map[string]any{
		"name": "staff", "application_id": shop, "inherits": []string{blogWriter},
	}, nil)
	if status != http.StatusBadRequest {
		t.Errorf("app role including another app's role = %d, want 400", status)
	}

	if status := manager.do(http.MethodPost, "/user-roles", map[string]any{"name": "intruder", "application_id": blog}, nil); status != http.StatusBadRequest {
		t.Errorf("role in another application = %d, want 400", status)
	}

	if status := manager.do(http.MethodPost, "/user-roles", map[string]any{"name": "intruder"}, nil); status != http.StatusForbidden {
		t.Errorf("global role = %d, want 403", status)
	}

	if status := manager.do(http.MethodGet, "/applications/"+blog, nil, nil); status != http.StatusNotFound {
		t.Errorf("another application = %d, want 404", status)
	}

	if status := manager.do(http.MethodGet, "/admins", nil, nil); status != http.StatusForbidden {
		t.Errorf("administrators = %d, want 403", status)
	}

	// A super admin may still compose across scopes.
	super.role("staff", shop, employee)
}

// An inclusion someone else made stays when a scoped administrator edits the
// rest of the role.
func TestLiveScopedAdminKeepsExistingInclusions(t *testing.T) {
	s := newLiveServer(t)
	super := s.superAdmin()

	shop := super.application("shop")
	employee := super.role("employee", "")
	staff := super.role("staff", shop, employee)

	manager := s.appManager(super, shop)

	manager.must(http.StatusOK, http.MethodPatch, "/user-roles/"+staff, map[string]any{
		"name": "staff", "description": "Everyone at the shop", "inherits": []string{employee},
	}, nil)
}

type apiScope struct {
	ID          string `json:"id,omitempty"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Default     bool   `json:"default"`
}

type apiBody struct {
	ID     string     `json:"id"`
	Scopes []apiScope `json:"scopes"`
}

func TestLiveAPIScopesCanTradeNames(t *testing.T) {
	s := newLiveServer(t)
	super := s.superAdmin()

	var created struct {
		API apiBody `json:"api"`
	}
	super.must(http.StatusCreated, http.MethodPost, "/apis", map[string]any{
		"name": "Orders", "identifier": "https://api.example.com/orders",
		"scopes": []apiScope{{Name: "orders:read"}, {Name: "orders:write"}},
	}, &created)

	read, write := created.API.Scopes[0], created.API.Scopes[1]
	read.Name, write.Name = write.Name, read.Name

	var updated struct {
		API apiBody `json:"api"`
	}
	super.must(http.StatusOK, http.MethodPatch, "/apis/"+created.API.ID, map[string]any{
		"name": "Orders", "scopes": []apiScope{read, write},
	}, &updated)

	names := map[string]string{}
	for _, scope := range updated.API.Scopes {
		names[scope.ID] = scope.Name
	}
	if names[read.ID] != "orders:write" || names[write.ID] != "orders:read" {
		t.Errorf("scopes after trading names = %v, want the ids kept and the names swapped", names)
	}

	// A new scope may take the name of one removed in the same save: read is
	// called orders:write by now.
	super.must(http.StatusOK, http.MethodPatch, "/apis/"+created.API.ID, map[string]any{
		"name": "Orders", "scopes": []apiScope{write, {Name: "orders:write"}},
	}, nil)
}

// Deleting an application takes its roles with it, the API scopes they grant
// included.
func TestLiveDeleteApplicationWithRoleGrants(t *testing.T) {
	s := newLiveServer(t)
	super := s.superAdmin()

	var api struct {
		API apiBody `json:"api"`
	}
	super.must(http.StatusCreated, http.MethodPost, "/apis", map[string]any{
		"name": "Orders", "identifier": "https://api.example.com/orders",
		"scopes": []apiScope{{Name: "orders:read"}},
	}, &api)

	shop := super.application("shop")
	super.must(http.StatusCreated, http.MethodPost, "/user-roles", map[string]any{
		"name": "buyer", "application_id": shop, "api_scopes": []string{api.API.Scopes[0].ID},
	}, nil)
	super.must(http.StatusOK, http.MethodPut, "/applications/"+shop+"/apis/"+api.API.ID, map[string]any{
		"scopes": []string{api.API.Scopes[0].ID},
	}, nil)

	super.must(http.StatusNoContent, http.MethodDelete, "/applications/"+shop, nil, nil)

	var roles struct {
		Total int `json:"total"`
	}
	super.must(http.StatusOK, http.MethodGet, "/user-roles", nil, &roles)
	if roles.Total != 0 {
		t.Errorf("roles after deleting their application = %d, want 0", roles.Total)
	}

	super.must(http.StatusNoContent, http.MethodDelete, "/apis/"+api.API.ID, nil, nil)
}

func TestLiveSearchTreatsWildcardsLiterally(t *testing.T) {
	s := newLiveServer(t)
	super := s.superAdmin()

	super.role("first_line", "")
	super.role("firstaline", "")

	var out struct {
		Total int `json:"total"`
	}
	super.must(http.StatusOK, http.MethodGet, "/user-roles?search="+url.QueryEscape("first_"), nil, &out)
	if out.Total != 1 {
		t.Errorf("search for first_ = %d roles, want only first_line", out.Total)
	}
}

// migrationsDir is the migrations folder, found from this file rather than
// from wherever the tests run.
func migrationsDir(t *testing.T) string {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot find the migrations")
	}

	return filepath.Join(filepath.Dir(file), "..", "..", "migrations")
}

// withDatabase is the DSN with its database swapped for another.
func withDatabase(t *testing.T, dsn, name string) string {
	parsed, err := url.Parse(dsn)
	if err != nil {
		t.Fatalf("%s is not a URL: %v", testDSNEnv, err)
	}
	parsed.Path = "/" + name

	return parsed.String()
}

func randomHex(t *testing.T, n int) string {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		t.Fatal(err)
	}

	return hex.EncodeToString(b)
}

type overviewBody struct {
	Counts  map[string]int64 `json:"counts"`
	SignIns struct {
		Succeeded int64 `json:"succeeded"`
		Failed    int64 `json:"failed"`
		Blocked   int64 `json:"blocked"`
	} `json:"sign_ins"`
	Daily []struct {
		Day      string `json:"day"`
		Events   int64  `json:"events"`
		Failures int64  `json:"failures"`
	} `json:"daily"`
	TopActors []struct {
		Actor  string `json:"actor"`
		Events int64  `json:"events"`
	} `json:"top_actors"`
	Activity []struct {
		Action string `json:"action"`
		Detail string `json:"detail"`
		Target *struct {
			Type string `json:"type"`
			Name string `json:"name"`
		} `json:"target"`
	} `json:"activity"`
}

func TestLiveOverview(t *testing.T) {
	s := newLiveServer(t)
	super := s.superAdmin()

	shop := super.application("shop")
	super.role("viewer", shop)
	s.client().login(superEmail, "wrong-password")
	super.must(http.StatusCreated, http.MethodPost, "/users", map[string]any{
		"email": "user@example.com", "password": "user-password-1", "confirm_password": "user-password-1",
	}, nil)
	user := s.browser()
	if status := user.account(http.MethodPost, "/login", map[string]string{
		"email": "user@example.com", "password": "wrong-password",
	}, nil); status != http.StatusUnauthorized {
		t.Fatalf("user login with a wrong password = %d, want 401", status)
	}
	if status := user.account(http.MethodPost, "/login", map[string]string{
		"email": "user@example.com", "password": "user-password-1",
	}, nil); status != http.StatusOK {
		t.Fatalf("user login = %d, want 200", status)
	}

	var got overviewBody
	super.must(http.StatusOK, http.MethodGet, "/overview", nil, &got)

	if got.Counts["applications"] != 1 || got.Counts["enabled_applications"] != 1 || got.Counts["user_roles"] != 1 || got.Counts["admins"] != 1 {
		t.Errorf("counts = %v", got.Counts)
	}
	if got.Counts["active_sessions"] != 1 {
		t.Errorf("active_sessions = %d, want the super admin's one", got.Counts["active_sessions"])
	}

	if got.SignIns.Succeeded != 2 || got.SignIns.Failed != 2 || got.SignIns.Blocked != 0 {
		t.Errorf("sign-ins = %+v, want two successes, two failures and no blocked attempts", got.SignIns)
	}

	if len(got.Daily) != 14 {
		t.Fatalf("daily = %d days, want 14", len(got.Daily))
	}
	if today := got.Daily[13]; today.Events < 7 || today.Failures != 2 {
		t.Errorf("today = %+v, want user and admin events plus two failures", today)
	}

	if len(got.TopActors) != 1 || got.TopActors[0].Actor != superEmail || got.TopActors[0].Events != 3 {
		t.Errorf("top actors = %+v, want the super admin with three changes and no sign-ins", got.TopActors)
	}

	named := map[string]string{}
	for _, event := range got.Activity {
		if event.Target != nil {
			named[event.Action] = event.Target.Name
		}
		if event.Action == "admin.login_failed" && event.Detail != "wrong password" {
			t.Errorf("failed sign-in detail = %q, want the reason", event.Detail)
		}
	}
	if named["application.created"] != "shop" || named["user_role.created"] != "viewer" {
		t.Errorf("targets named = %v, want the application and the role", named)
	}
}

// An administrator who may read the log but not a kind of record sees that
// something happened to it, without what it is called.
func TestLiveActivityHidesNamesAdministratorsCannotSee(t *testing.T) {
	s := newLiveServer(t)
	super := s.superAdmin()

	super.application("shop")

	const email, password = "auditor@example.com", "auditor-password-1"
	var created struct {
		Role idOnly `json:"role"`
	}
	super.must(http.StatusCreated, http.MethodPost, "/admin-roles", map[string]any{
		"name": "log_reader", "permissions": []string{"activity.read"},
	}, &created)
	super.must(http.StatusCreated, http.MethodPost, "/admins", map[string]any{
		"email": email, "first_name": "Log", "status": "active",
		"password": password, "confirm_password": password,
		"assignments": []map[string]any{{"role_id": created.Role.ID}},
	}, nil)

	reader := s.client()
	if status := reader.login(email, password); status != http.StatusOK {
		t.Fatalf("login = %d", status)
	}

	var got overviewBody
	reader.must(http.StatusOK, http.MethodGet, "/overview", nil, &got)

	for _, event := range got.Activity {
		if event.Target != nil && event.Target.Name != "" {
			t.Errorf("%s: target named %q, want no name for this administrator", event.Action, event.Target.Name)
		}
	}
}

// A panel with nothing done this week still answers with lists, not null,
// which the dashboard counts the length of.
func TestLiveOverviewListsAreNeverNull(t *testing.T) {
	s := newLiveServer(t)
	super := s.superAdmin()

	var raw map[string]json.RawMessage
	super.must(http.StatusOK, http.MethodGet, "/overview", nil, &raw)

	for _, key := range []string{"daily", "top_actors", "activity"} {
		if value := string(raw[key]); value == "null" || value == "" || value[0] != '[' {
			t.Errorf("%s = %s, want a list", key, value)
		}
	}
}

// The organisation is one record of settings every installation starts with:
// an update changes what it names and leaves the rest as it was, and the log
// says which settings moved.
// A stale organisation in the cache — cached before the database was reset,
// or by another server sharing the Redis — is not what a save writes back:
// the save reads the row itself, and succeeds.
func TestLiveOrganizationSaveIgnoresAStaleCache(t *testing.T) {
	s := newLiveServer(t)
	if s.cache == nil {
		t.Skip("no test Redis: set " + cachetest.Env)
	}
	super := s.superAdmin()

	// Something reads the organisation first, so the group is in use.
	super.must(http.StatusOK, http.MethodGet, "/organization", nil, nil)

	stale := model.DefaultOrganization()
	stale.ID = uuid.New()
	stale.Name = "stale"
	s.cache.Set(context.Background(), cache.Organization, "settings", stale)

	var answer struct {
		Organization struct {
			Name     string `json:"name"`
			TermsURL string `json:"terms_url"`
		} `json:"organization"`
	}
	super.must(http.StatusOK, http.MethodPatch, "/organization", map[string]any{
		"terms_url": "https://acme.example.com/terms",
	}, &answer)

	if answer.Organization.TermsURL != "https://acme.example.com/terms" || answer.Organization.Name != brand.Name {
		t.Errorf("after the save = %+v, want the stored organization with its terms", answer.Organization)
	}

	stored, err := s.store.OrganizationForUpdate(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if stored.TermsURL != "https://acme.example.com/terms" {
		t.Errorf("stored terms = %q, want the saved link", stored.TermsURL)
	}
}

func TestLiveOrganizationSettings(t *testing.T) {
	s := newLiveServer(t)
	super := s.superAdmin()

	type organizationBody struct {
		Organization struct {
			Name         string `json:"name"`
			Slug         string `json:"slug"`
			Domain       string `json:"domain"`
			SupportEmail string `json:"support_email"`
			SupportPhone string `json:"support_phone"`
			TermsURL     string `json:"terms_url"`
			PrivacyURL   string `json:"privacy_url"`
		} `json:"organization"`
	}

	// The migration seeds the row, so the page has something to open on.
	var seeded organizationBody
	super.must(http.StatusOK, http.MethodGet, "/organization", nil, &seeded)

	if seeded.Organization.Name != brand.Name || seeded.Organization.Slug != brand.Slug {
		t.Errorf("seeded = %+v, want the default organization", seeded.Organization)
	}

	var updated organizationBody
	super.must(http.StatusOK, http.MethodPatch, "/organization", map[string]any{
		"name":          "  Acme Inc  ",
		"slug":          "acme",
		"domain":        "Acme.Example.COM",
		"support_email": "support@acme.example.com",
		"support_phone": "+996 555 123456",
		"terms_url":     "https://acme.example.com/terms",
		"privacy_url":   "https://acme.example.com/privacy",
	}, &updated)

	if got := updated.Organization; got.Name != "Acme Inc" || got.Slug != "acme" || got.Domain != "acme.example.com" {
		t.Errorf("updated = %+v, want the values trimmed and lower cased", got)
	}

	// A second update says nothing about the domain, which therefore stays.
	var kept organizationBody
	super.must(http.StatusOK, http.MethodPatch, "/organization", map[string]any{
		"support_phone": "+996 555 999888",
	}, &kept)

	if kept.Organization.Domain != "acme.example.com" || kept.Organization.SupportPhone != "+996 555 999888" {
		t.Errorf("kept = %+v, want the domain left alone and the new number", kept.Organization)
	}

	// A refused change writes nothing.
	super.must(http.StatusBadRequest, http.MethodPatch, "/organization", map[string]any{
		"slug": "acme inc",
	}, nil)

	var reread organizationBody
	super.must(http.StatusOK, http.MethodGet, "/organization", nil, &reread)
	if reread.Organization.Slug != "acme" {
		t.Errorf("slug = %q, want the stored one after a refused change", reread.Organization.Slug)
	}

	// The log says which settings moved, newest first.
	var logs struct {
		Logs []struct {
			Action string `json:"action"`
			Target *struct {
				Type string `json:"type"`
				ID   string `json:"id"`
			} `json:"target"`
			Detail string `json:"detail"`
		} `json:"logs"`
	}
	super.must(http.StatusOK, http.MethodGet, "/logs?limit=50", nil, &logs)

	var changed []string
	for _, entry := range logs.Logs {
		if entry.Action == "organization.updated" {
			if entry.Target == nil || entry.Target.Type != "organization" || entry.Target.ID != "acme" {
				t.Errorf("target = %+v, want the organization by its slug", entry.Target)
			}
			changed = append(changed, entry.Detail)
		}
	}
	if len(changed) != 2 {
		t.Fatalf("organization.updated entries = %d, want one per change that took", len(changed))
	}
	if changed[0] != "support_phone" {
		t.Errorf("the last change recorded %q, want the number alone", changed[0])
	}

	// A role scoped to one application grants nothing here: these are the
	// whole installation's settings.
	manager := s.appManager(super, super.application("shop"))
	manager.must(http.StatusForbidden, http.MethodGet, "/organization", nil, nil)
	manager.must(http.StatusForbidden, http.MethodPatch, "/organization", map[string]any{"name": "Theirs"}, nil)

	// What an administrator saved is what the outside world is told. The
	// agreements are the provider's own metadata, and the sign-in pages read
	// the rest from the public API, which takes no session.
	discovery := map[string]any{}
	getJSON(t, s.root+"/.well-known/openid-configuration", &discovery)

	if discovery["op_tos_uri"] != "https://acme.example.com/terms" {
		t.Errorf("op_tos_uri = %v, want the organization's terms", discovery["op_tos_uri"])
	}
	if discovery["op_policy_uri"] != "https://acme.example.com/privacy" {
		t.Errorf("op_policy_uri = %v, want the organization's privacy policy", discovery["op_policy_uri"])
	}

	var public struct {
		Organization struct {
			Name         string `json:"name"`
			SupportEmail string `json:"support_email"`
			SupportPhone string `json:"support_phone"`
			TermsURL     string `json:"terms_url"`
		} `json:"organization"`
	}
	getJSON(t, s.root+"/api/v1/account/organization", &public)

	if got := public.Organization; got.Name != "Acme Inc" || got.SupportEmail != "support@acme.example.com" {
		t.Errorf("the sign-in pages are told %+v, want the saved name and address", got)
	}
	if public.Organization.SupportPhone != "+996 555 999888" {
		t.Errorf("support_phone = %q, want the number saved last", public.Organization.SupportPhone)
	}
}

// getJSON reads a public endpoint, which needs no session and no cookie.
func getJSON(t *testing.T, url string, out any) {
	t.Helper()

	res, err := http.Get(url)
	if err != nil {
		t.Fatalf("GET %s: %v", url, err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Fatalf("GET %s = %d, want 200", url, res.StatusCode)
	}
	if err := json.NewDecoder(res.Body).Decode(out); err != nil {
		t.Fatalf("GET %s: decode: %v", url, err)
	}
}

// defaultFlow is the id of the flow every application without one of its own
// falls back to.
func defaultFlow(t *testing.T, super *client) string {
	t.Helper()

	var flows struct {
		Flows []struct {
			ID        string `json:"id"`
			IsDefault bool   `json:"is_default"`
		} `json:"flows"`
	}
	super.must(http.StatusOK, http.MethodGet, "/login-flows", nil, &flows)

	for _, flow := range flows.Flows {
		if flow.IsDefault {
			return flow.ID
		}
	}

	t.Fatal("there is no default flow")

	return ""
}
