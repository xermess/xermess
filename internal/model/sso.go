package model

import (
	"fmt"
	"net/url"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
)

// domainPattern is a host name and nothing else: labels joined by dots, with
// no scheme, no port and no path.
var domainPattern = regexp.MustCompile(`^[a-z0-9]([a-z0-9-]*[a-z0-9])?(\.[a-z0-9]([a-z0-9-]*[a-z0-9])?)+$`)

// SSOConnection is an organisation's own identity provider — Okta, Microsoft
// Entra ID, Google Workspace, ADFS, Keycloak — that its people sign in
// through, over OpenID Connect or SAML 2.0. It is what authentik calls a
// source and Keycloak an identity provider, set up for a company rather than
// for the public: where Social offers "Continue with Google" to anybody, a
// connection usually owns email domains, and the people at those domains sign
// in through it.
//
// What a connection decides, in the order a sign-in meets it:
//
//   - which addresses it may sign in (Domains): an identity provider says who
//     somebody is, and with domains this server believes it for those and for
//     no other, so a misconfigured provider cannot sign in as anyone elsewhere.
//     Without any it is trusted for every address, as authentik and Keycloak
//     trust a source, and people reach it only through its button;
//   - whether those domains have to use it (EnforceDomains): their password
//     sign-in, registration and reset are refused, and the sign-in page sends
//     them to their provider instead;
//   - what happens to someone this server already has (Matching) or has never
//     seen (CreateUsers — just-in-time provisioning);
//   - what their record says afterwards (SyncProfile), and which roles the
//     groups the provider puts them in give them (RoleMappings, SyncRoles).
//
// The secrets — the OIDC client secret and the key SAML requests are signed
// with — are sealed with the server's secret key and never leave the server.
type SSOConnection struct {
	Base

	// Slug names the connection in the addresses the provider is given —
	// /oauth2/sso/<slug>/callback, /acs, /metadata — so it cannot change once
	// the provider has them.
	Slug string `gorm:"size:64;not null;uniqueIndex" json:"slug"`

	// Name is what the sign-in page calls it: "Continue with <name>".
	Name     string      `gorm:"size:100;not null" json:"name"`
	Protocol SSOProtocol `gorm:"type:varchar(8);not null" json:"protocol"`
	Enabled  bool        `gorm:"not null" json:"enabled"`

	// Domains are the email domains the connection signs people in for, lower
	// case. A domain belongs to one connection at most. None is every address,
	// and then the button on the sign-in page is the only way to the
	// connection, since no address leads to it.
	Domains StringList `gorm:"type:jsonb;serializer:json;not null" json:"domains"`

	// EnforceDomains makes the connection the only way in for its domains.
	EnforceDomains bool `gorm:"not null" json:"enforce_domains"`

	// ShowOnLogin puts a button for the connection on the sign-in page. Off,
	// people reach it by typing an address at one of its domains.
	ShowOnLogin bool `gorm:"not null" json:"show_on_login"`

	// OpenID Connect: the issuer, whose discovery document says everything
	// else, and the client this server is registered as there.
	Issuer       string     `gorm:"size:512" json:"issuer"`
	ClientID     string     `gorm:"size:255" json:"client_id"`
	ClientSecret []byte     `gorm:"type:bytea" json:"-"`
	Scopes       StringList `gorm:"serializer:json" json:"scopes"`

	// SAML 2.0: the provider's metadata — fetched from MetadataURL, or pasted
	// in when the provider only offers a file — which carries its entity ID,
	// its sign-in address and the certificate its assertions are signed with.
	MetadataURL string `gorm:"size:1024" json:"metadata_url"`
	Metadata    string `gorm:"type:text" json:"metadata"`
	// NameIDFormat is what to ask the provider to identify people by.
	NameIDFormat SSONameIDFormat `gorm:"type:varchar(16)" json:"name_id_format"`
	// SignRequests signs the authentication requests sent to the provider,
	// for the providers that require it.
	SignRequests bool `gorm:"not null" json:"sign_requests"`
	// SPKey is the private key this server signs requests with, a PEM, sealed;
	// SPCertificate its self-signed certificate, which the provider is given.
	// Both are made when the connection is.
	SPKey         []byte `gorm:"type:bytea" json:"-"`
	SPCertificate string `gorm:"type:text" json:"sp_certificate"`

	// Matching is what to do with an address this server already has.
	Matching SSOMatching `gorm:"type:varchar(16);not null" json:"matching"`
	// CreateUsers makes an account for someone signing in for the first time.
	CreateUsers bool `gorm:"not null" json:"create_users"`
	// SyncProfile rewrites the user's name from the provider on every sign-in:
	// the provider is where it is kept.
	SyncProfile bool `gorm:"not null" json:"sync_profile"`

	// The names of the claims or attributes to read. Empty is the protocol's
	// usual ones (SSOAttributeDefaults).
	EmailAttribute     string `gorm:"size:255" json:"email_attribute"`
	FirstNameAttribute string `gorm:"size:255" json:"first_name_attribute"`
	LastNameAttribute  string `gorm:"size:255" json:"last_name_attribute"`
	GroupsAttribute    string `gorm:"size:255" json:"groups_attribute"`

	// RoleMappings give a role to everybody the provider puts in a group.
	RoleMappings SSORoleMappings `gorm:"type:jsonb;serializer:json;not null" json:"role_mappings"`
	// SyncRoles also takes a mapped role away from someone no longer in its
	// group. Off, mapped roles are only ever added.
	SyncRoles bool `gorm:"not null" json:"sync_roles"`
}

// TableName pins the table name.
func (SSOConnection) TableName() string {
	return "sso_connections"
}

// SSOProtocol is which protocol a connection speaks.
type SSOProtocol string

const (
	SSOProtocolOIDC SSOProtocol = "oidc"
	SSOProtocolSAML SSOProtocol = "saml"
)

// SSOMatching is what to do when somebody signs in through a connection with
// an address this server already has an account for — authentik's user
// matching modes, for the two that make sense when the provider owns the
// domain.
type SSOMatching string

const (
	// SSOMatchLink signs them in to that account and connects the provider to
	// it: the provider is trusted for the domain, so it is the same person.
	SSOMatchLink SSOMatching = "link"
	// SSOMatchDeny refuses, so an existing account is only ever reached the
	// way it was made.
	SSOMatchDeny SSOMatching = "deny"
)

// SSONameIDFormat is what a SAML provider is asked to identify people by.
type SSONameIDFormat string

const (
	SSONameIDEmail       SSONameIDFormat = "email"
	SSONameIDPersistent  SSONameIDFormat = "persistent"
	SSONameIDUnspecified SSONameIDFormat = "unspecified"
)

// URN is the format's name in SAML.
func (f SSONameIDFormat) URN() string {
	switch f {
	case SSONameIDEmail:
		return "urn:oasis:names:tc:SAML:1.1:nameid-format:emailAddress"
	case SSONameIDPersistent:
		return "urn:oasis:names:tc:SAML:2.0:nameid-format:persistent"
	default:
		return "urn:oasis:names:tc:SAML:1.1:nameid-format:unspecified"
	}
}

// SSORoleMapping gives a role to the members of one group at the provider.
type SSORoleMapping struct {
	Group  string    `json:"group"`
	RoleID uuid.UUID `json:"role_id"`
}

// SSORoleMappings is a list of them, stored as JSON; an empty one is [].
type SSORoleMappings []SSORoleMapping

// Roles are the roles the mappings can give, each once.
func (m SSORoleMappings) Roles() []uuid.UUID {
	var out []uuid.UUID
	for _, mapping := range m {
		if !slices.Contains(out, mapping.RoleID) {
			out = append(out, mapping.RoleID)
		}
	}

	return out
}

// For are the roles the mappings give someone in the given groups. Group
// names are compared without regard to case, as providers are inconsistent
// about it.
func (m SSORoleMappings) For(groups []string) []uuid.UUID {
	var out []uuid.UUID
	for _, mapping := range m {
		if slices.ContainsFunc(groups, func(group string) bool { return strings.EqualFold(group, mapping.Group) }) &&
			!slices.Contains(out, mapping.RoleID) {
			out = append(out, mapping.RoleID)
		}
	}

	return out
}

// SSOAttributeDefaults are where the usual providers put each thing, tried in
// order when a connection names nothing: OpenID Connect's standard claims,
// and for SAML the names Entra ID, ADFS, Okta, Google Workspace and the LDAP
// OIDs use.
var SSOAttributeDefaults = map[SSOProtocol]map[string][]string{
	SSOProtocolOIDC: {
		"email":      {"email"},
		"first_name": {"given_name"},
		"last_name":  {"family_name"},
		"groups":     {"groups"},
	},
	SSOProtocolSAML: {
		"email": {
			"email", "mail", "emailAddress", "Email",
			"http://schemas.xmlsoap.org/ws/2005/05/identity/claims/emailaddress",
			"urn:oid:0.9.2342.19200300.100.1.3",
		},
		"first_name": {
			"firstName", "givenName", "given_name", "FirstName",
			"http://schemas.xmlsoap.org/ws/2005/05/identity/claims/givenname",
			"urn:oid:2.5.4.42",
		},
		"last_name": {
			"lastName", "surname", "sn", "family_name", "LastName",
			"http://schemas.xmlsoap.org/ws/2005/05/identity/claims/surname",
			"urn:oid:2.5.4.4",
		},
		"groups": {
			"groups", "memberOf", "Groups",
			"http://schemas.microsoft.com/ws/2008/06/identity/claims/groups",
			"http://schemas.xmlsoap.org/claims/Group",
		},
	},
}

// Attribute is where to look for one thing — "email", "first_name",
// "last_name" or "groups": the name the connection gives, or the protocol's
// usual ones.
func (c SSOConnection) Attribute(what string) []string {
	named := map[string]string{
		"email":      c.EmailAttribute,
		"first_name": c.FirstNameAttribute,
		"last_name":  c.LastNameAttribute,
		"groups":     c.GroupsAttribute,
	}[what]

	if strings.TrimSpace(named) != "" {
		return []string{strings.TrimSpace(named)}
	}

	return SSOAttributeDefaults[c.Protocol][what]
}

// AskedScopes is what to ask an OpenID Connect provider for: the connection's
// scopes, with openid always among them, since without it there is no
// id_token to read.
func (c SSOConnection) AskedScopes() []string {
	scopes := []string{"openid", "email", "profile"}
	if len(c.Scopes) > 0 {
		scopes = append([]string{"openid"}, c.Scopes...)
	}

	return slices.Compact(slices.DeleteFunc(slices.Clone(scopes), func(scope string) bool {
		return strings.TrimSpace(scope) == ""
	}))
}

// OwnsEmail reports whether the connection may sign an address in: one at
// one of its domains, or any address when it has none.
func (c SSOConnection) OwnsEmail(email string) bool {
	_, domain, ok := strings.Cut(strings.ToLower(strings.TrimSpace(email)), "@")
	return ok && (len(c.Domains) == 0 || slices.Contains(c.Domains, domain))
}

// The addresses a connection is given to its provider at, all under the
// issuer: the OpenID Connect redirect URI, and the SAML assertion consumer
// service, the metadata that describes this server as a service provider,
// and the entity ID, which is the metadata's address as is usual.
func (c SSOConnection) CallbackURL(issuer string) string { return c.base(issuer) + "/callback" }
func (c SSOConnection) ACSURL(issuer string) string      { return c.base(issuer) + "/acs" }
func (c SSOConnection) MetadataURLFor(issuer string) string {
	return c.base(issuer) + "/metadata"
}
func (c SSOConnection) EntityID(issuer string) string { return c.MetadataURLFor(issuer) }

func (c SSOConnection) base(issuer string) string {
	return strings.TrimRight(issuer, "/") + "/oauth2/sso/" + c.Slug
}

// ValidDomain reports whether a domain looks like one: labels of letters,
// numbers and dashes, at least two of them.
func ValidDomain(domain string) bool {
	return domainPattern.MatchString(domain)
}

// NormalizeDomain is a domain as connections store it: lower case, without an
// "@" or a trailing dot.
func NormalizeDomain(domain string) string {
	return strings.TrimSuffix(strings.TrimPrefix(strings.ToLower(strings.TrimSpace(domain)), "@"), ".")
}

// Validate reports the first thing wrong with a connection. It checks what a
// connection is, not whether its provider answers — that is the panel's test.
func (c SSOConnection) Validate() error {
	switch {
	case !socialSlugPattern.MatchString(c.Slug) || len(c.Slug) > 64:
		return fmt.Errorf("slug must be lower case letters, numbers and dashes")
	case strings.TrimSpace(c.Name) == "" || len(c.Name) > 100:
		return fmt.Errorf("name is required, and at most 100 characters")
	case len(c.Domains) == 0 && c.EnforceDomains:
		return fmt.Errorf("only a connection with domains can be required")
	case len(c.Domains) == 0 && !c.ShowOnLogin:
		return fmt.Errorf("a connection without domains needs its button, or nobody can reach it")
	case c.Matching != SSOMatchLink && c.Matching != SSOMatchDeny:
		return fmt.Errorf("matching must be link or deny")
	}

	for _, domain := range c.Domains {
		if domain != NormalizeDomain(domain) || !domainPattern.MatchString(domain) {
			return fmt.Errorf("%q is not a domain", domain)
		}
	}

	for _, mapping := range c.RoleMappings {
		if strings.TrimSpace(mapping.Group) == "" || mapping.RoleID == uuid.Nil {
			return fmt.Errorf("every role mapping needs a group and a role")
		}
	}

	switch c.Protocol {
	case SSOProtocolOIDC:
		if parsed, err := url.Parse(c.Issuer); err != nil || parsed.Scheme != "https" && !isLocalhost(parsed) || parsed.Host == "" {
			return fmt.Errorf("issuer must be an https URL")
		}
		if strings.TrimSpace(c.ClientID) == "" {
			return fmt.Errorf("client ID is required")
		}
	case SSOProtocolSAML:
		if strings.TrimSpace(c.Metadata) == "" && strings.TrimSpace(c.MetadataURL) == "" {
			return fmt.Errorf("SAML needs the identity provider's metadata, as a URL or pasted in")
		}
		switch c.NameIDFormat {
		case SSONameIDEmail, SSONameIDPersistent, SSONameIDUnspecified:
		default:
			return fmt.Errorf("name ID format must be email, persistent or unspecified")
		}
	default:
		return fmt.Errorf("protocol must be oidc or saml")
	}

	return nil
}

// isLocalhost lets a provider on this machine be plain http, which is how one
// is tried out; anywhere else the identity it vouches for travels over TLS.
func isLocalhost(u *url.URL) bool {
	host := u.Hostname()
	return u.Scheme == "http" && (host == "localhost" || host == "127.0.0.1" || host == "::1")
}

// SSOIdentity is one person at a connection's provider, and the user it signs
// in. The subject is the provider's own id for them — the OIDC `sub`, the SAML
// NameID — which is what identifies them reliably; the address is kept to show
// which account at the provider it is.
type SSOIdentity struct {
	Base

	UserID       uuid.UUID  `gorm:"type:uuid;not null;index" json:"user_id"`
	ConnectionID uuid.UUID  `gorm:"type:uuid;not null;uniqueIndex:idx_sso_identities_subject,priority:1" json:"connection_id"`
	Subject      string     `gorm:"size:512;not null;uniqueIndex:idx_sso_identities_subject,priority:2" json:"subject"`
	Email        string     `gorm:"size:255" json:"email"`
	LastLoginAt  *time.Time `json:"last_login_at"`

	Connection *SSOConnection `gorm:"constraint:OnDelete:CASCADE" json:"-"`
	User       *User          `gorm:"constraint:OnDelete:CASCADE" json:"-"`
}

// TableName pins the table name.
func (SSOIdentity) TableName() string {
	return "sso_identities"
}

// SSOLogin is a sign-in sent to a connection's provider and not yet back: the
// state (OIDC) or relay state (SAML) we gave it, hashed, and what the answer
// has to be matched against — the PKCE verifier and the nonce, or the ID of
// the SAML request the response has to be in response to. Like SocialLogin,
// it is a row so it is single use and survives a different browser finishing.
type SSOLogin struct {
	Base

	StateHash    string    `gorm:"size:64;not null;uniqueIndex" json:"-"`
	ConnectionID uuid.UUID `gorm:"type:uuid;not null;index" json:"connection_id"`

	Verifier  string `gorm:"size:128" json:"-"`
	Nonce     string `gorm:"size:128" json:"-"`
	RequestID string `gorm:"size:128" json:"-"`

	// Request is the sign-in under way to continue, and Next where to go when
	// there is none.
	Request string `gorm:"size:64" json:"-"`
	Next    string `gorm:"size:512" json:"-"`

	ExpiresAt time.Time `gorm:"not null;index" json:"-"`

	Connection *SSOConnection `gorm:"constraint:OnDelete:CASCADE" json:"-"`
}

// TableName pins the table name.
func (SSOLogin) TableName() string {
	return "sso_logins"
}
