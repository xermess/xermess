package model

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"net/url"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Application is an app or service that signs its users in through this
// server with OAuth 2.0 and OpenID Connect: a client, in OAuth's words.
//
// What is stored follows the client metadata of OAuth 2.0 Dynamic Client
// Registration (RFC 7591) and the rules of the OAuth 2.0 Security Best
// Current Practice (RFC 9700): redirect URIs are matched exactly, the implicit
// and password grants are not offered, and a client that cannot keep a
// secret has to use PKCE.
//
// An application also owns roles. A role means something only inside the
// application it belongs to, so "admin" in one app says nothing about another,
// and a token issued for an app carries that app's roles alone.
type Application struct {
	Base

	Name        string          `gorm:"size:100;not null" json:"name"`
	Description string          `gorm:"size:255" json:"description"`
	Type        ApplicationType `gorm:"type:varchar(16);not null;index" json:"type"`

	// LogoURI and ClientURI are shown on the sign-in pages. There is no
	// consent page: see the note on asking in README.md.
	LogoURI   string `gorm:"size:512" json:"logo_uri"`
	ClientURI string `gorm:"size:512" json:"client_uri"`

	// PolicyURI and TosURI are the application's privacy policy and terms
	// of service, linked from its sign-in pages and agreed to on
	// registration. The names are RFC 7591's.
	PolicyURI string `gorm:"size:512" json:"policy_uri"`
	TosURI    string `gorm:"size:512" json:"tos_uri"`

	// AllowRegistration offers "Create an account" on the application's
	// sign-in page. Users made that way get the default roles, as users an
	// administrator makes do.
	AllowRegistration bool `gorm:"not null" json:"allow_registration"`

	// ClientID is what the application identifies itself with. It is made
	// here and never changes.
	ClientID string `gorm:"size:64;uniqueIndex;not null" json:"client_id"`

	// ClientSecretHash is a SHA-256 of the secret, which is only ever shown
	// once, when it is made. The secret is long and random, so a fast hash is
	// right here, as it is for session tokens. SecretHint is its last four
	// characters, so an administrator can tell which secret an app is using.
	ClientSecretHash string     `gorm:"size:64" json:"-"`
	SecretHint       string     `gorm:"size:8" json:"secret_hint"`
	SecretCreatedAt  *time.Time `json:"secret_created_at"`

	TokenEndpointAuthMethod AuthMethod `gorm:"type:varchar(32);not null" json:"token_endpoint_auth_method"`

	GrantTypes             []string `gorm:"type:text;serializer:json" json:"grant_types"`
	RedirectURIs           []string `gorm:"type:text;serializer:json" json:"redirect_uris"`
	PostLogoutRedirectURIs []string `gorm:"type:text;serializer:json" json:"post_logout_redirect_uris"`
	Scopes                 []string `gorm:"type:text;serializer:json" json:"scopes"`

	// RequirePKCE refuses an authorization request without a code challenge.
	// It is always on for a public client.
	RequirePKCE bool `gorm:"not null" json:"require_pkce"`

	// Token lifetimes, in seconds.
	AccessTokenLifetime  int `gorm:"not null" json:"access_token_lifetime"`
	IDTokenLifetime      int `gorm:"not null" json:"id_token_lifetime"`
	RefreshTokenLifetime int `gorm:"not null" json:"refresh_token_lifetime"`

	// AssertRoles puts the user's roles in this application into its tokens
	// and the userinfo response. RequireRoleAssignment lets only users who
	// hold at least one of its roles sign in to it. Both are Zitadel's
	// project settings of the same names.
	AssertRoles           bool `gorm:"not null" json:"assert_roles"`
	RequireRoleAssignment bool `gorm:"not null;default:false" json:"require_role_assignment"`

	// LoginFlowID is the login flow this application signs people in with.
	// Nil means the default flow, which is also what an application holding
	// a flow that has since been turned off falls back to — so a flow can be
	// withdrawn without taking the applications using it down with it.
	LoginFlowID *uuid.UUID `gorm:"type:uuid;index" json:"login_flow_id"`

	// Enabled is false for an application that may not sign anyone in.
	Enabled bool `gorm:"not null;index" json:"enabled"`
}

// TableName pins the table name.
func (Application) TableName() string {
	return "applications"
}

// ApplicationType is what kind of client an application is, which decides
// whether it can keep a secret and which grants make sense for it.
type ApplicationType string

const (
	// AppWeb is a server-rendered app with a backend that keeps a secret.
	AppWeb ApplicationType = "web"
	// AppSPA is a single-page app running entirely in the browser.
	AppSPA ApplicationType = "spa"
	// AppNative is a mobile or desktop app installed on a device.
	AppNative ApplicationType = "native"
	// AppM2M is a service calling APIs as itself, with no user present.
	AppM2M ApplicationType = "m2m"
)

// ApplicationTypes lists every type, in the order the panel offers them.
var ApplicationTypes = []ApplicationType{AppWeb, AppSPA, AppNative, AppM2M}

// Valid reports whether t is a known type.
func (t ApplicationType) Valid() bool {
	return slices.Contains(ApplicationTypes, t)
}

// Public reports whether the type is a public client: one whose code ships to
// the user's device, and so cannot keep a secret.
func (t ApplicationType) Public() bool {
	return t == AppSPA || t == AppNative
}

// AuthMethod is how a client proves who it is at the token endpoint.
type AuthMethod string

const (
	AuthClientSecretBasic AuthMethod = "client_secret_basic"
	AuthClientSecretPost  AuthMethod = "client_secret_post"
	// AuthNone is a public client's: it has no secret, and PKCE stands in.
	AuthNone AuthMethod = "none"
)

// The grants an application may use. The implicit and resource owner password
// grants are left out on purpose: RFC 9700 says not to use them.
const (
	GrantAuthorizationCode = "authorization_code"
	GrantRefreshToken      = "refresh_token"
	GrantClientCredentials = "client_credentials"
)

// GrantTypes lists every grant, in the order the panel offers them.
var GrantTypes = []string{GrantAuthorizationCode, GrantRefreshToken, GrantClientCredentials}

// The scopes an application may request. "roles" asks for the user's roles
// in the application.
const (
	ScopeOpenID        = "openid"
	ScopeProfile       = "profile"
	ScopeEmail         = "email"
	ScopeOfflineAccess = "offline_access"
	ScopeRoles         = "roles"
)

// Scopes lists every scope, in the order the panel offers them.
var Scopes = []string{ScopeOpenID, ScopeProfile, ScopeEmail, ScopeOfflineAccess, ScopeRoles}

// Default token lifetimes, in seconds.
const (
	DefaultAccessTokenLifetime  = 60 * 60
	DefaultIDTokenLifetime      = 60 * 60
	DefaultRefreshTokenLifetime = 30 * 24 * 60 * 60
)

// Bounds on the lifetimes an administrator may set.
const (
	minTokenLifetime        = 60
	maxAccessTokenLifetime  = 24 * 60 * 60
	maxRefreshTokenLifetime = 365 * 24 * 60 * 60
)

// ResponseTypes is what the application may ask the authorization endpoint
// for. It follows from the grants (RFC 7591 section 2.1), so it is not stored.
func (a Application) ResponseTypes() []string {
	if slices.Contains(a.GrantTypes, GrantAuthorizationCode) {
		return []string{"code"}
	}

	return []string{}
}

// HasSecret reports whether the application authenticates with a secret.
func (a Application) HasSecret() bool {
	return a.TokenEndpointAuthMethod != AuthNone
}

// Normalise fills in what a type implies and tidies the lists, before
// Validate checks what is left. A public client gets no secret method and
// always requires PKCE; a machine-to-machine one has no user, so nothing to
// redirect to and no OpenID scopes.
func (a *Application) Normalise() {
	switch {
	case a.Type.Public():
		a.TokenEndpointAuthMethod = AuthNone
		a.RequirePKCE = true
	case a.TokenEndpointAuthMethod == "" || a.TokenEndpointAuthMethod == AuthNone:
		a.TokenEndpointAuthMethod = AuthClientSecretBasic
	}

	if a.Type == AppM2M {
		a.AllowRegistration = false
		a.GrantTypes = []string{GrantClientCredentials}
		a.RedirectURIs = []string{}
		a.PostLogoutRedirectURIs = []string{}
		a.Scopes = []string{}
		a.RequirePKCE = false
		a.RequireRoleAssignment = false
	}

	a.GrantTypes = ordered(a.GrantTypes, GrantTypes)
	a.Scopes = ordered(a.Scopes, Scopes)
	a.RedirectURIs = tidyURIs(a.RedirectURIs)
	a.PostLogoutRedirectURIs = tidyURIs(a.PostLogoutRedirectURIs)
}

// Validate reports the first thing wrong with an application's settings, as
// a sentence naming the field. Call Normalise first.
func (a Application) Validate() error {
	if !a.Type.Valid() {
		return fmt.Errorf("type must be one of: web, spa, native, m2m")
	}

	switch a.TokenEndpointAuthMethod {
	case AuthClientSecretBasic, AuthClientSecretPost, AuthNone:
	default:
		return fmt.Errorf("token_endpoint_auth_method must be one of: client_secret_basic, client_secret_post, none")
	}

	for _, grant := range a.GrantTypes {
		if !slices.Contains(GrantTypes, grant) {
			return fmt.Errorf("grant_types: %q is not offered (implicit and password grants are not supported)", grant)
		}
	}

	for _, scope := range a.Scopes {
		if !slices.Contains(Scopes, scope) {
			return fmt.Errorf("scopes: %q is not a scope", scope)
		}
	}

	code := slices.Contains(a.GrantTypes, GrantAuthorizationCode)

	switch a.Type {
	case AppM2M:
		// Normalise has already settled everything for this type.
	case AppWeb, AppSPA, AppNative:
		if !code {
			return fmt.Errorf("grant_types: a %s application signs users in, so it needs authorization_code", a.Type)
		}
		if a.Type.Public() && slices.Contains(a.GrantTypes, GrantClientCredentials) {
			return fmt.Errorf("grant_types: a public client has no secret, so it cannot use client_credentials")
		}
	}

	if slices.Contains(a.GrantTypes, GrantRefreshToken) && !code {
		return fmt.Errorf("grant_types: refresh_token needs authorization_code")
	}

	if code && len(a.RedirectURIs) == 0 {
		return fmt.Errorf("redirect_uris: authorization_code needs at least one redirect URI")
	}

	for _, uri := range a.RedirectURIs {
		if err := checkRedirectURI(uri, a.Type); err != nil {
			return fmt.Errorf("redirect_uris: %w", err)
		}
	}

	for _, uri := range a.PostLogoutRedirectURIs {
		if err := checkRedirectURI(uri, a.Type); err != nil {
			return fmt.Errorf("post_logout_redirect_uris: %w", err)
		}
	}

	// client_uri, policy_uri and tos_uri are only links, so plain http is
	// fine, as it is for an app running on a developer's machine. logo_uri
	// is loaded as an image on the sign-in page, which is served over https,
	// and a browser blocks an http image there.
	links := []struct{ name, value string }{
		{"client_uri", a.ClientURI},
		{"policy_uri", a.PolicyURI},
		{"tos_uri", a.TosURI},
	}
	for _, link := range links {
		if link.value == "" {
			continue
		}
		if parsed, err := url.Parse(link.value); err != nil || (parsed.Scheme != "https" && parsed.Scheme != "http") || parsed.Host == "" {
			return fmt.Errorf("%s must be an http or https URL", link.name)
		}
	}

	if a.LogoURI != "" {
		if parsed, err := url.Parse(a.LogoURI); err != nil || parsed.Scheme != "https" || parsed.Host == "" {
			return fmt.Errorf("logo_uri must be an https URL")
		}
	}

	lifetimes := []struct {
		name  string
		value int
		max   int
	}{
		{"access_token_lifetime", a.AccessTokenLifetime, maxAccessTokenLifetime},
		{"id_token_lifetime", a.IDTokenLifetime, maxAccessTokenLifetime},
		{"refresh_token_lifetime", a.RefreshTokenLifetime, maxRefreshTokenLifetime},
	}
	for _, lifetime := range lifetimes {
		if lifetime.value < minTokenLifetime || lifetime.value > lifetime.max {
			return fmt.Errorf("%s must be between %d and %d seconds", lifetime.name, minTokenLifetime, lifetime.max)
		}
	}

	return nil
}

// checkRedirectURI holds a redirect URI to RFC 9700: an absolute URI with no
// fragment and no wildcard, since it is matched exactly. It has to be https,
// except on a loopback address, and a native app may also use a private-use
// scheme such as com.example.app:/callback (RFC 8252).
func checkRedirectURI(raw string, kind ApplicationType) error {
	if strings.Contains(raw, "*") {
		return fmt.Errorf("%q: wildcards are not allowed, redirect URIs are matched exactly", raw)
	}

	parsed, err := url.Parse(raw)
	if err != nil || !parsed.IsAbs() {
		return fmt.Errorf("%q is not an absolute URI", raw)
	}

	if parsed.Fragment != "" || strings.Contains(raw, "#") {
		return fmt.Errorf("%q must not have a fragment", raw)
	}

	switch parsed.Scheme {
	case "https":
		if parsed.Host == "" {
			return fmt.Errorf("%q has no host", raw)
		}
		return nil
	case "http":
		if loopback(parsed.Hostname()) {
			return nil
		}
		return fmt.Errorf("%q must use https, unless it is on localhost", raw)
	default:
		// A private-use scheme is a reverse domain name, so it has a dot in
		// it and cannot be mistaken for javascript: or data:.
		if kind == AppNative && strings.Contains(parsed.Scheme, ".") {
			return nil
		}
		return fmt.Errorf("%q must use https", raw)
	}
}

func loopback(host string) bool {
	return host == "localhost" || host == "127.0.0.1" || host == "::1"
}

// NewClientID returns a client id: 128 random bits, hex encoded.
func NewClientID() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", fmt.Errorf("generate client id: %w", err)
	}

	return hex.EncodeToString(b[:]), nil
}

// ErrNoSecret is returned when a secret is asked of a public client.
var ErrNoSecret = errors.New("a public client has no secret")

// IssueSecret gives the application a new client secret, replacing any it had,
// and returns it. This is the only time the secret exists outside the caller:
// only its hash is kept.
func (a *Application) IssueSecret(now time.Time) (string, error) {
	if !a.HasSecret() {
		return "", ErrNoSecret
	}

	var b [32]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", fmt.Errorf("generate client secret: %w", err)
	}

	secret := base64.RawURLEncoding.EncodeToString(b[:])
	sum := sha256.Sum256([]byte(secret))

	a.ClientSecretHash = hex.EncodeToString(sum[:])
	a.SecretHint = secret[len(secret)-4:]
	a.SecretCreatedAt = &now

	return secret, nil
}

// CheckSecret reports whether a presented secret is the application's, in
// constant time.
func (a Application) CheckSecret(secret string) bool {
	if a.ClientSecretHash == "" {
		return false
	}

	sum := sha256.Sum256([]byte(secret))

	return subtle.ConstantTimeCompare([]byte(hex.EncodeToString(sum[:])), []byte(a.ClientSecretHash)) == 1
}

// ordered keeps the values that appear in `catalog`, once each, in catalog
// order. Anything else stays at the end, in the order given, for Validate to
// refuse.
func ordered(values, catalog []string) []string {
	out := []string{}
	for _, known := range catalog {
		if slices.Contains(values, known) {
			out = append(out, known)
		}
	}

	for _, value := range values {
		if !slices.Contains(out, value) {
			out = append(out, value)
		}
	}

	return out
}

// tidyURIs trims each URI and drops blanks and repeats.
func tidyURIs(uris []string) []string {
	out := []string{}
	for _, uri := range uris {
		uri = strings.TrimSpace(uri)
		if uri != "" && !slices.Contains(out, uri) {
			out = append(out, uri)
		}
	}

	return out
}
