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

// SSOConnection is an organisation's identity provider (Okta, Entra ID, Google
// Workspace, ADFS, Keycloak) over OpenID Connect or SAML 2.0.
//
// What it decides:
//
//   - Domains: the addresses it may sign in. Without any it is trusted for
//     every address and reached only through its button;
//   - EnforceDomains: those domains must use it, and password sign-in,
//     registration and reset are refused;
//   - Matching and CreateUsers: what happens to known and new people;
//   - SyncProfile, RoleMappings and SyncRoles: what their record and roles
//     become.
//
// The OIDC client secret and SAML signing key are sealed with the server's
// secret key.
type SSOConnection struct {
	Base

	// Slug is part of the addresses given to the provider (callback, ACS,
	// metadata), so it never changes.
	Slug string `gorm:"size:64;not null;uniqueIndex" json:"slug"`

	// Name is what the sign-in page calls it: "Continue with <name>".
	Name      string      `gorm:"size:100;not null" json:"name"`
	Protocol  SSOProtocol `gorm:"type:varchar(8);not null" json:"protocol"`
	IsEnabled bool        `gorm:"not null" json:"is_enabled"`

	// Domains the connection signs people in for, lower case, each owned by one
	// connection at most. None means every address, reachable only through its
	// button.
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

	// SAML 2.0: the provider's metadata (from MetadataURL or pasted) with its
	// entity ID, sign-in address and signing certificate.
	MetadataURL string `gorm:"size:1024" json:"metadata_url"`
	Metadata    string `gorm:"type:text" json:"metadata"`
	// NameIDFormat is what to ask the provider to identify people by.
	NameIDFormat SSONameIDFormat `gorm:"type:varchar(16)" json:"name_id_format"`
	// SignRequests signs the authentication requests sent to the provider,
	// for the providers that require it.
	SignRequests bool `gorm:"not null" json:"sign_requests"`
	// SPKey is our sealed PEM signing key and SPCertificate its self-signed
	// certificate, both made with the connection.
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

// SSOMatching decides what happens when an address already has an account
// (authentik's user matching modes).
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

// For returns the roles mapped to the given groups, compared
// case-insensitively.
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

// SSOAttributeDefaults are where common providers put each attribute, tried in
// order when a connection names none.
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

// Attribute is where to look for "email", "first_name", "last_name" or
// "groups": the connection's name or the protocol's defaults.
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

// AskedScopes are the connection's scopes with openid always included.
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

// Addresses given to the provider, all under the issuer: the OIDC redirect URI,
// the SAML ACS, our metadata, and the entity ID (the metadata address).
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

// SSOIdentity links a provider's subject (OIDC sub or SAML NameID) to a user;
// the address is kept for display.
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

// SSOLogin is a sign-in sent to a connection and not yet back: the hashed
// (relay) state and what the answer must match (PKCE verifier and nonce, or the
// SAML request ID). Like SocialLogin it is single use.
type SSOLogin struct {
	Base

	StateHash    string    `gorm:"size:64;not null;uniqueIndex" json:"-"`
	ConnectionID uuid.UUID `gorm:"type:uuid;not null;index" json:"connection_id"`

	Verifier      string `gorm:"size:128" json:"-"`
	Nonce         string `gorm:"size:128" json:"-"`
	SAMLRequestID string `gorm:"size:128" json:"-"`

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
