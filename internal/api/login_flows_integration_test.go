package api

import (
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"
)

// A login flow decides how people sign in, not only what the page offers: a
// flow without a password refuses one, a flow that requires a verified
// address sends a link instead of a session, and a session lasts as long as
// the flow says.
func TestLiveLoginFlowsDecideSignIn(t *testing.T) {
	s := newLiveServer(t)
	super := s.superAdmin()

	var flows struct {
		Flows []struct {
			ID        string `json:"id"`
			IsDefault bool   `json:"is_default"`
		} `json:"flows"`
	}
	super.must(http.StatusOK, http.MethodGet, "/login-flows", nil, &flows)
	var flow string
	for _, f := range flows.Flows {
		if f.IsDefault {
			flow = f.ID
		}
	}
	if flow == "" {
		t.Fatal("there is no default flow")
	}
	change := func(body map[string]any) {
		super.must(http.StatusOK, http.MethodPatch, "/login-flows/"+flow, body, nil)
	}

	super.must(http.StatusCreated, http.MethodPost, "/users", map[string]any{
		"email": "ada@example.com", "password": "ada-password-1", "confirm_password": "ada-password-1",
	}, nil)
	signIn := func(b *browser) (int, problemBody) {
		var out problemBody
		status := b.account(http.MethodPost, "/login", map[string]string{"email": "ada@example.com", "password": "ada-password-1"}, &out)
		return status, out
	}

	// Without a password step, a password is refused before it is checked,
	// and nobody is sent a reset link.
	change(map[string]any{"steps": []string{"identifier", "social"}})
	if status, out := signIn(s.browser()); status != http.StatusForbidden || out.Code != "password_not_offered" {
		t.Errorf("a password at a flow without one = %d %+v, want password_not_offered", status, out)
	}
	s.browser().account(http.MethodPost, "/forgot-password", map[string]string{"email": "ada@example.com"}, nil)

	// A flow that requires a verified address sends a link, not a session.
	change(map[string]any{"steps": []string{"identifier", "password"}, "require_verified_email": true, "session_lifetime_hours": 2})
	b := s.browser()
	if status, out := signIn(b); status != http.StatusForbidden || out.Code != "email_not_verified" {
		t.Fatalf("an unverified address = %d %+v, want email_not_verified", status, out)
	}

	sent := s.mail.wait(t, "ada@example.com")
	if strings.Contains(sent.Subject, "password") {
		t.Fatalf("the first email to Ada is %q: the password-less flow sent a reset link", sent.Subject)
	}
	token := tokenIn(t, sent.Body, "/verify-email")

	if status := b.account(http.MethodPost, "/verify-email", map[string]string{"token": token}, nil); status != http.StatusOK {
		t.Fatalf("verifying the address = %d, want 200", status)
	}
	var reused problemBody
	if status := b.account(http.MethodPost, "/verify-email", map[string]string{"token": token}, &reused); status != http.StatusGone || reused.Code != "verification_invalid" {
		t.Errorf("a verification link used twice = %d %+v, want verification_invalid", status, reused)
	}

	// Verified, Ada signs in — for as long as the flow says.
	if status, out := signIn(b); status != http.StatusOK {
		t.Fatalf("signing in once verified = %d %+v", status, out)
	}

	var sessions struct {
		Sessions []struct {
			SignedInAt time.Time `json:"signed_in_at"`
			ExpiresAt  time.Time `json:"expires_at"`
		} `json:"sessions"`
	}
	super.must(http.StatusOK, http.MethodGet, "/user-sessions?search=ada", nil, &sessions)
	if len(sessions.Sessions) != 1 {
		t.Fatalf("Ada's sessions = %+v, want one", sessions.Sessions)
	}
	if lasts := sessions.Sessions[0].ExpiresAt.Sub(sessions.Sessions[0].SignedInAt); lasts != 2*time.Hour {
		t.Errorf("the session lasts %v, want the flow's two hours", lasts)
	}

	// A flow cannot be saved that nobody could get through today.
	var refused problemBody
	if status := super.do(http.MethodPatch, "/login-flows/"+flow, map[string]any{"steps": []string{"identifier", "totp"}}, &refused); status != http.StatusBadRequest {
		t.Errorf("a flow of steps not run yet = %d %+v, want 400", status, refused)
	}
}

// A session belongs to a browser; a login flow belongs to an application. The
// cookie one flow made is not enough for an application whose flow asks for
// more, so authorize sends that browser to the sign-in page instead of handing
// out a code.
//
// Finding 16 in SECURITY-AUDIT-2.md: authorize took whatever session the
// browser carried. Signing in through the default flow — on the account pages,
// with no application involved at all — drew a code for an application that
// asks for a code emailed to the address, and for one closed to sign-ins.
func TestLiveASessionDoesNotCarryPastAStricterFlow(t *testing.T) {
	s := newLiveServer(t)
	super := s.superAdmin()

	// Four digits rather than six: shorter to read in the message.
	super.must(http.StatusOK, http.MethodPatch, "/otp", map[string]any{
		"length": 4, "max_attempts": 3, "resend_seconds": 0,
	}, nil)

	// The staff tool asks for a code emailed to the address. The shop names no
	// flow, so it falls back to the default one: a password and no more.
	var strict struct {
		Flow idOnly `json:"flow"`
	}
	super.must(http.StatusCreated, http.MethodPost, "/login-flows", map[string]any{
		"name": "Staff sign-in", "slug": "staff", "enabled": true,
		"steps": []string{"identifier", "password", "email_code"},
	}, &strict)

	const staffRedirect = "https://staff.example.com/callback"
	staff := flowApp(t, super, "Staff", staffRedirect, strict.Flow.ID)
	shop := flowApp(t, super, "Shop", shopRedirect, "")

	super.must(http.StatusCreated, http.MethodPost, "/users", map[string]any{
		"email": "ada@example.com", "password": "ada-password-1",
		"confirm_password": "ada-password-1", "email_verified": true,
	}, nil)

	// Ada signs in on the account pages. No application, so the default flow,
	// which the password alone satisfies.
	b := s.browser()
	if status := b.account(http.MethodPost, "/login", map[string]string{
		"email": "ada@example.com", "password": "ada-password-1",
	}, nil); status != http.StatusOK {
		t.Fatalf("signing in through the default flow = %d, want 200", status)
	}

	// The shop takes that session: the flow that made it is its own.
	if to := b.visit(authorizeAs(s, shop, shopRedirect)); to.Query().Get("code") == "" {
		t.Errorf("the shop sent the browser to %s, want a code", to)
	}

	// The staff tool does not, and says so the way it would to a browser with
	// no session at all: by asking it to sign in.
	to := b.visit(authorizeAs(s, staff, staffRedirect))
	if to.Query().Get("code") != "" {
		t.Fatalf("a password-only session drew a code for the staff tool: %s", to)
	}
	if !strings.HasPrefix(to.String(), testAccountURL+"/login?") {
		t.Fatalf("the staff tool sent the browser to %s, want the sign-in page", to)
	}

	// Through the flow, code and all, the staff tool is satisfied.
	var held struct {
		Code *struct {
			Handle string `json:"handle"`
		} `json:"code"`
	}
	if status := b.account(http.MethodPost, "/login", map[string]string{
		"request": to.Query().Get("request"), "email": "ada@example.com", "password": "ada-password-1",
	}, &held); status != http.StatusOK || held.Code == nil {
		t.Fatalf("signing in to the staff tool = %d %+v, want a code to type", status, held)
	}

	var finished struct {
		RedirectTo string `json:"redirect_to"`
	}
	code := codeIn(t, s.mail.wait(t, "ada@example.com").Body, 4)
	if status := b.account(http.MethodPost, "/login/code", map[string]string{
		"handle": held.Code.Handle, "code": code,
	}, &finished); status != http.StatusOK {
		t.Fatalf("typing the code = %d, want 200", status)
	}
	if !strings.Contains(finished.RedirectTo, "code=") {
		t.Fatalf("the code did not finish the sign-in: %s", finished.RedirectTo)
	}

	// That session carries back to the shop, which asks for less than it
	// proved.
	if to := b.visit(authorizeAs(s, shop, shopRedirect)); to.Query().Get("code") == "" {
		t.Errorf("the shop refused the session the staff flow made: %s", to)
	}

	// A flow closed to sign-ins lends none either, session in hand or not.
	super.must(http.StatusOK, http.MethodPatch, "/login-flows/"+defaultFlow(t, super), map[string]any{
		"allow_sign_in": false,
	}, nil)
	if to := b.visit(authorizeAs(s, shop, shopRedirect)); to.Query().Get("code") != "" {
		t.Errorf("a flow closed to sign-ins still drew a code for the shop: %s", to)
	}
}

// flowApp registers an application held to one login flow, or to the default
// one when `flow` is empty, and returns its client id.
func flowApp(t *testing.T, super *client, name, redirect, flow string) string {
	t.Helper()

	body := map[string]any{
		"name": name, "type": "web",
		"grant_types":   []string{"authorization_code"},
		"redirect_uris": []string{redirect},
		"scopes":        []string{"openid", "profile"},
	}
	if flow != "" {
		body["login_flow_id"] = flow
	}

	var out struct {
		Application struct {
			ClientID string `json:"client_id"`
		} `json:"application"`
	}
	super.must(http.StatusCreated, http.MethodPost, "/applications", body, &out)

	return out.Application.ClientID
}

// authorizeAs is an authorization request as one application's own library
// would build it.
func authorizeAs(s *liveServer, clientID, redirect string) string {
	_, challenge := pkce(s.t)

	return s.root + "/oauth2/authorize?" + url.Values{
		"response_type":         {"code"},
		"client_id":             {clientID},
		"redirect_uri":          {redirect},
		"scope":                 {"openid profile"},
		"state":                 {"a-state"},
		"code_challenge":        {challenge},
		"code_challenge_method": {"S256"},
	}.Encode()
}
