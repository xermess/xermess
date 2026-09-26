package sso

import (
	"errors"
	"testing"

	"loginer/internal/api/respond"
)

// Trying a provider sends this server to fetch an address somebody typed, so
// which addresses those may be is a rule of its own.
//
// Finding 24 in SECURITY-AUDIT-2.md: nothing checked them at all. A test named
// an address and the server read it, whatever it was — including plain http,
// which is how an address that only speaks http is reached.
func TestTestRequestValidate(t *testing.T) {
	tests := []struct {
		name    string
		request testRequest
		want    string // the problem's code, or "" when the test may go ahead
	}{
		{
			name:    "an issuer over https",
			request: testRequest{Protocol: "oidc", Issuer: "https://idp.example.com"},
		},
		{
			name:    "an issuer on this machine, being tried out",
			request: testRequest{Protocol: "oidc", Issuer: "http://localhost:8080/realms/acme"},
		},
		{
			name:    "an issuer over plain http",
			request: testRequest{Protocol: "oidc", Issuer: "http://idp.example.com"},
			want:    "sso_issuer_invalid",
		},
		{
			name:    "an address that only speaks http",
			request: testRequest{Protocol: "oidc", Issuer: "http://169.254.169.254"},
			want:    "sso_issuer_invalid",
		},
		{
			name:    "no issuer at all",
			request: testRequest{Protocol: "oidc"},
			want:    "sso_issuer_invalid",
		},
		{
			name:    "something that is not an address",
			request: testRequest{Protocol: "oidc", Issuer: "idp.example.com"},
			want:    "sso_issuer_invalid",
		},
		{
			name:    "metadata over https",
			request: testRequest{Protocol: "saml", MetadataURL: "https://idp.example.com/metadata"},
		},
		{
			name:    "metadata over plain http",
			request: testRequest{Protocol: "saml", MetadataURL: "http://idp.example.com/metadata"},
			want:    "validation.invalid",
		},
		{
			name:    "metadata pasted rather than fetched",
			request: testRequest{Protocol: "saml", Metadata: "<EntityDescriptor/>"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.request.validate()

			if tt.want == "" {
				if err != nil {
					t.Fatalf("validate() = %v, want the test to go ahead", err)
				}
				return
			}

			var fault respond.Fault
			if !errors.As(err, &fault) {
				t.Fatalf("validate() = %v, want a Fault", err)
			}
			if fault.Code != tt.want {
				t.Errorf("validate() = %q, want %q", fault.Code, tt.want)
			}
		})
	}
}

// The same rule holds for a metadata address that is saved, because it is
// fetched again on every refresh.
func TestSavedMetadataURLIsHeldToTheSameRule(t *testing.T) {
	for _, tt := range []struct {
		name string
		url  string
		want string
	}{
		{name: "https", url: "https://idp.example.com/metadata"},
		{name: "on this machine", url: "http://127.0.0.1:7000/metadata"},
		{name: "plain http", url: "http://idp.example.com/metadata", want: "validation.invalid"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if fetchable(tt.url) != (tt.want == "") {
				t.Fatalf("fetchable(%q) = %v", tt.url, !fetchable(tt.url))
			}
		})
	}
}
