package oidc

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"testing"
	"time"

	"loginer/internal/model"
)

// unsignedIDToken is a JWT with these claims and nothing else worth reading:
// idTokenClaims does not check the signature, so a header and a payload are
// all it takes.
func unsignedIDToken(t *testing.T, claims map[string]any) string {
	t.Helper()

	payload, err := json.Marshal(claims)
	if err != nil {
		t.Fatal(err)
	}

	b64 := base64.RawURLEncoding

	return b64.EncodeToString([]byte(`{"alg":"none"}`)) + "." + b64.EncodeToString(payload) + ".not-checked"
}

// An id_token the identity is read out of is held to the client it was issued
// for, to the issuer its kind is known to have, and to the clock. The token
// endpoint answering over TLS says the answer came from the provider; it does
// not say the token inside it is one of theirs.
func TestIDTokenClaimsHoldsATokenToItsProvider(t *testing.T) {
	now := time.Date(2026, 9, 20, 12, 0, 0, 0, time.UTC)
	service := &Service{
		log: slog.New(slog.NewTextHandler(io.Discard, nil)),
		now: func() time.Time { return now },
	}

	apple := &model.SocialProvider{Kind: model.SocialApple, Slug: "apple", ClientID: "com.example.app"}
	inHouse := &model.SocialProvider{Kind: model.SocialOIDC, Slug: "in-house", ClientID: "in-house-app"}

	claims := func(issuer, audience string, expires time.Time) map[string]any {
		return map[string]any{"iss": issuer, "aud": audience, "sub": "000123", "exp": expires.Unix()}
	}

	tests := []struct {
		name     string
		provider *model.SocialProvider
		claims   map[string]any
		want     error
	}{
		{
			name:     "the provider's own token",
			provider: apple,
			claims:   claims("https://appleid.apple.com", "com.example.app", now.Add(time.Hour)),
		},
		{
			name:     "another issuer, with this client's audience",
			provider: apple,
			claims:   claims("https://not-apple.example", "com.example.app", now.Add(time.Hour)),
			want:     ErrSocialUpstream,
		},
		{
			name:     "no issuer at all",
			provider: apple,
			claims:   map[string]any{"aud": "com.example.app", "sub": "000123", "exp": now.Add(time.Hour).Unix()},
			want:     ErrSocialUpstream,
		},
		{
			name:     "another client's token",
			provider: apple,
			claims:   claims("https://appleid.apple.com", "com.somebody-else.app", now.Add(time.Hour)),
			want:     ErrSocialUpstream,
		},
		{
			name:     "a token that has expired",
			provider: apple,
			claims:   claims("https://appleid.apple.com", "com.example.app", now.Add(-time.Minute)),
			want:     ErrSocialExpired,
		},
		{
			name:     "a kind whose issuer nobody knows in advance",
			provider: inHouse,
			claims:   claims("https://whatever.example", "in-house-app", now.Add(time.Hour)),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := service.idTokenClaims(tt.provider, unsignedIDToken(t, tt.claims))
			if !errors.Is(err, tt.want) {
				t.Errorf("idTokenClaims = %v, want %v", err, tt.want)
			}
		})
	}
}

// Everything the provider fetches from elsewhere goes through one client, and a
// redirect may not take it out of https.
//
// Finding 24 in SECURITY-AUDIT-2.md: the client followed redirects with Go's
// default policy, so an https address could answer "fetch this http one
// instead" and this server would.
func TestKeepTheScheme(t *testing.T) {
	from := func(scheme string) []*http.Request {
		req, err := http.NewRequest(http.MethodGet, scheme+"://idp.example.com/metadata", nil)
		if err != nil {
			t.Fatal(err)
		}

		return []*http.Request{req}
	}

	to := func(raw string) *http.Request {
		req, err := http.NewRequest(http.MethodGet, raw, nil)
		if err != nil {
			t.Fatal(err)
		}

		return req
	}

	tests := []struct {
		name    string
		via     []*http.Request
		next    *http.Request
		allowed bool
	}{
		{
			name:    "https to https",
			via:     from("https"),
			next:    to("https://elsewhere.example.com/metadata"),
			allowed: true,
		},
		{
			name: "https to http",
			via:  from("https"),
			next: to("http://elsewhere.example.com/metadata"),
		},
		{
			name: "https to an address that only speaks http",
			via:  from("https"),
			next: to("http://169.254.169.254/latest/meta-data/"),
		},
		{
			// A provider on this machine, being tried out: it began in the
			// clear and there is nothing to step down from.
			name:    "http to http, having started that way",
			via:     from("http"),
			next:    to("http://localhost:8080/metadata"),
			allowed: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := keepTheScheme(tt.next, tt.via)

			if tt.allowed && err != nil {
				t.Errorf("keepTheScheme() = %v, want the redirect followed", err)
			}
			if !tt.allowed && err == nil {
				t.Error("keepTheScheme() = nil, want the redirect refused")
			}
		})
	}
}

// Replacing CheckRedirect replaces the limit that came with it, so the limit
// is still there.
func TestKeepTheSchemeStopsGoingRound(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet, "https://idp.example.com/metadata", nil)
	if err != nil {
		t.Fatal(err)
	}

	via := make([]*http.Request, maxRedirects)
	for i := range via {
		via[i] = req
	}

	if err := keepTheScheme(req, via); err == nil {
		t.Errorf("keepTheScheme() after %d hops = nil, want it stopped", maxRedirects)
	}
}

// Which addresses this server will go and read, for a provider somebody is
// setting up. The rule lives here because the two fetches apply it, not only
// the panel that takes the address (finding 24 in SECURITY-AUDIT-2.md).
func TestFetchable(t *testing.T) {
	tests := []struct {
		address string
		want    bool
	}{
		{address: "https://idp.example.com", want: true},
		{address: "https://idp.internal:8443/realms/acme", want: true},
		// An identity provider on an internal host is the ordinary case for
		// single sign-on, so https to one is allowed on purpose.
		{address: "https://10.0.0.7/metadata", want: true},
		{address: "http://localhost:8080/realms/acme", want: true},
		{address: "http://127.0.0.1/metadata", want: true},
		{address: "http://[::1]:7000/metadata", want: true},
		{address: "http://idp.example.com", want: false},
		{address: "http://169.254.169.254/latest/meta-data/", want: false},
		{address: "http://10.0.0.7/metadata", want: false},
		{address: "ftp://idp.example.com", want: false},
		{address: "file:///etc/passwd", want: false},
		{address: "idp.example.com", want: false},
		{address: "", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.address, func(t *testing.T) {
			if got := Fetchable(tt.address); got != tt.want {
				t.Errorf("Fetchable(%q) = %v, want %v", tt.address, got, tt.want)
			}
		})
	}
}
