package sso

import (
	"crypto/x509"
	"encoding/base64"
	"strings"
	"time"

	"github.com/crewjam/saml"

	"loginer/internal/model"
	"loginer/internal/oidc"
)

// connectionResponse is a connection as the panel shows it: what is stored,
// never the secrets; how many people sign in through it; what to give the
// identity provider; and, for SAML, what its metadata says.
type connectionResponse struct {
	model.SSOConnection

	HasClientSecret bool  `json:"has_client_secret"`
	Users           int64 `json:"users"`

	ServiceProvider  serviceProvider   `json:"service_provider"`
	IdentityProvider *identityProvider `json:"identity_provider,omitempty"`
}

// serviceProvider is what the identity provider is set up with: for OpenID
// Connect the redirect URI, for SAML the assertion consumer service, the
// entity ID, the metadata that says both, and the certificate requests are
// signed with.
type serviceProvider struct {
	CallbackURL string `json:"callback_url,omitempty"`
	ACSURL      string `json:"acs_url,omitempty"`
	EntityID    string `json:"entity_id,omitempty"`
	MetadataURL string `json:"metadata_url,omitempty"`
	Certificate string `json:"certificate,omitempty"`
}

// identityProvider is what a SAML provider's metadata says about it, so the
// panel can show it is the right one — and when its certificate runs out,
// which is the usual way a working connection stops working.
type identityProvider struct {
	EntityID           string     `json:"entity_id"`
	SSOURL             string     `json:"sso_url"`
	Certificates       int        `json:"certificates"`
	CertificateExpires *time.Time `json:"certificate_expires,omitempty"`
}

type listResponse struct {
	Connections []connectionResponse `json:"connections"`
}

type response struct {
	Connection connectionResponse `json:"connection"`
}

// testResponse is what trying a provider found.
type testResponse struct {
	// For OpenID Connect, the endpoints its discovery document gave.
	Issuer                string `json:"issuer,omitempty"`
	AuthorizationEndpoint string `json:"authorization_endpoint,omitempty"`
	TokenEndpoint         string `json:"token_endpoint,omitempty"`
	JWKSURI               string `json:"jwks_uri,omitempty"`

	// UnsupportedScopes are the scopes the connection would ask for that the
	// provider does not list. Some providers refuse the whole sign-in over
	// one — Keycloak answers invalid_scope — so the panel warns before
	// anybody tries.
	UnsupportedScopes []string `json:"unsupported_scopes,omitempty"`

	// For SAML, what its metadata said, and the metadata itself when it was
	// fetched, so the panel can keep it.
	IdentityProvider *identityProvider `json:"identity_provider,omitempty"`
	Metadata         string            `json:"metadata,omitempty"`
}

func newConnectionResponse(connection model.SSOConnection, users int64, issuer string) connectionResponse {
	out := connectionResponse{
		SSOConnection:   connection,
		HasClientSecret: len(connection.ClientSecret) > 0,
		Users:           users,
	}

	switch connection.Protocol {
	case model.SSOProtocolOIDC:
		out.ServiceProvider = serviceProvider{CallbackURL: connection.CallbackURL(issuer)}
	case model.SSOProtocolSAML:
		out.ServiceProvider = serviceProvider{
			ACSURL:      connection.ACSURL(issuer),
			EntityID:    connection.EntityID(issuer),
			MetadataURL: connection.MetadataURLFor(issuer),
			Certificate: connection.SPCertificate,
		}
		if descriptor, err := oidc.ParseSAMLMetadata([]byte(connection.Metadata)); err == nil {
			out.IdentityProvider = describeIdentityProvider(descriptor)
		}
	}

	return out
}

// describeIdentityProvider is what a provider's metadata says, briefly.
func describeIdentityProvider(descriptor *saml.EntityDescriptor) *identityProvider {
	idp := descriptor.IDPSSODescriptors[0]
	out := &identityProvider{EntityID: descriptor.EntityID}

	for _, endpoint := range idp.SingleSignOnServices {
		if endpoint.Binding == saml.HTTPRedirectBinding {
			out.SSOURL = endpoint.Location
		}
	}

	for _, key := range idp.KeyDescriptors {
		if key.Use != "" && key.Use != "signing" {
			continue
		}
		for _, certificate := range key.KeyInfo.X509Data.X509Certificates {
			out.Certificates++
			if expires, ok := certificateExpiry(certificate.Data); ok &&
				(out.CertificateExpires == nil || expires.Before(*out.CertificateExpires)) {
				out.CertificateExpires = &expires
			}
		}
	}

	return out
}

// certificateExpiry is when a certificate in metadata — base64 DER, as SAML
// carries it, often broken over lines — stops being valid.
func certificateExpiry(data string) (time.Time, bool) {
	der, err := base64.StdEncoding.DecodeString(strings.Join(strings.Fields(data), ""))
	if err != nil {
		return time.Time{}, false
	}

	certificate, err := x509.ParseCertificate(der)
	if err != nil {
		return time.Time{}, false
	}

	return certificate.NotAfter, true
}
