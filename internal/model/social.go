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

// SocialProvider is an account somewhere else that users may sign in with:
// Google, Apple, Facebook, Yandex ID, VK, or any other OAuth 2.0 or OpenID
// Connect provider an administrator configures.
//
// The record holds what this server needs to be a client of that provider —
// which is to say, what their console gave out when the application was
// registered there. Nothing here is a secret the provider does not already
// share with us, except ClientSecret and PrivateKey, which are encrypted with
// the server's secret key and never leave the server.
//
// What each kind's endpoints and claims are is not stored: it is SocialKinds
// below, so a provider that changes an address is corrected in one place
// rather than in every installation's database.
type SocialProvider struct {
	Base

	Kind SocialKind `gorm:"type:varchar(32);not null" json:"kind"`

	// Slug names the provider in the addresses users are sent to and come
	// back on, so it cannot change once people are signing in with it:
	// /oauth2/social/<slug>/start and /callback are registered with the
	// provider as the redirect URI.
	Slug string `gorm:"size:64;not null;uniqueIndex" json:"slug"`

	// Name is what the button says: "Continue with <name>".
	Name string `gorm:"size:100;not null" json:"name"`

	ClientID string `gorm:"size:255;not null" json:"client_id"`

	// ClientSecret is sealed with the server's secret key. Apple has none —
	// see PrivateKey.
	ClientSecret []byte `gorm:"type:bytea" json:"-"`

	// Scopes is what to ask the provider for. The kind's defaults are used
	// when it is empty.
	Scopes StringList `gorm:"serializer:json" json:"scopes"`

	Enabled bool `gorm:"not null" json:"enabled"`

	// Position is where the button sits among the others.
	Position int `gorm:"not null" json:"position"`

	// LinkVerifiedEmails attaches a sign-in to the account that already has
	// the address, when the provider says the address is verified. Off, an
	// address that is already taken is refused, and the user has to sign in
	// and connect the provider from their account.
	LinkVerifiedEmails bool `gorm:"not null" json:"link_verified_emails"`

	// AllowRegistration lets someone the server has never seen make an
	// account by signing in with this provider.
	AllowRegistration bool `gorm:"not null" json:"allow_registration"`

	// The endpoints, for the kinds that have none built in. A kind that does
	// ignores them.
	AuthorizeURL string `gorm:"size:512" json:"authorize_url"`
	TokenURL     string `gorm:"size:512" json:"token_url"`
	UserInfoURL  string `gorm:"size:512" json:"userinfo_url"`

	// TokenAuth is how the secret is presented at the token endpoint, for the
	// kinds whose provider is not known in advance: RFC 6749 says a server
	// must accept HTTP Basic and may accept the body, and they differ over
	// which they prefer. Empty is the kind's own default.
	TokenAuth SocialTokenAuth `gorm:"type:varchar(32)" json:"token_auth"`

	// Apple signs its client secret rather than holding one: a JWT made per
	// request from a key registered with the team. PrivateKey is the .p8
	// file's contents, sealed like ClientSecret.
	TeamID     string `gorm:"size:64" json:"team_id"`
	KeyID      string `gorm:"size:64" json:"key_id"`
	PrivateKey []byte `gorm:"type:bytea" json:"-"`
}

// TableName pins the table name.
func (SocialProvider) TableName() string {
	return "social_providers"
}

// StringList is a list of strings stored as JSON. It is its own type so an
// empty list is stored as [] rather than null.
type StringList []string

// UserIdentity is one account at a provider, and the user it signs in.
//
// The subject is the provider's own id for the person, which is the only
// thing that identifies them reliably: an address can change hands, and some
// providers do not give one at all.
type UserIdentity struct {
	Base

	UserID     uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"`
	ProviderID uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_user_identities_subject,priority:1" json:"provider_id"`
	Subject    string    `gorm:"size:255;not null;uniqueIndex:idx_user_identities_subject,priority:2" json:"subject"`

	// Email is what the provider said the address was when the account was
	// connected, kept so the panel and the user's own page can show which
	// account at the provider this is.
	Email       string     `gorm:"size:255" json:"email"`
	LastLoginAt *time.Time `json:"last_login_at"`

	Provider *SocialProvider `gorm:"constraint:OnDelete:CASCADE" json:"-"`
	User     *User           `gorm:"constraint:OnDelete:CASCADE" json:"-"`
}

// TableName pins the table name.
func (UserIdentity) TableName() string {
	return "user_identities"
}

// SocialAccount is one provider a user signs in with, as their record shows
// it: enough to say which account at the provider it is, and to disconnect it.
type SocialAccount struct {
	// ID is the identity's, which is what disconnecting it names.
	ID       uuid.UUID  `json:"id"`
	Provider string     `json:"provider"`
	Slug     string     `json:"slug"`
	Kind     SocialKind `json:"kind"`

	Email       string     `json:"email"`
	ConnectedAt time.Time  `json:"connected_at"`
	LastLoginAt *time.Time `json:"last_login_at"`
}

// SocialLogin is a sign-in that has been sent to a provider and not yet come
// back: the `state` we gave them, hashed as every other secret is, and what
// the answer has to be matched against.
//
// It is a row rather than a cookie because it has to be single use and
// short-lived, and because the browser that comes back from the provider is
// not always the one that left — Apple posts the answer, and a phone may
// finish in a different browser than it started in.
type SocialLogin struct {
	Base

	StateHash  string    `gorm:"size:64;not null;uniqueIndex" json:"-"`
	ProviderID uuid.UUID `gorm:"type:uuid;not null;index" json:"provider_id"`

	// Verifier is the PKCE code verifier, for the providers that take one.
	Verifier string `gorm:"size:128" json:"-"`

	// Request is the sign-in under way this belongs to, so the user can be
	// sent back to the application that asked. Empty when someone is signing
	// in to their own account page.
	Request string `gorm:"size:64" json:"-"`

	// Next is where to send the browser when there is no application waiting.
	Next string `gorm:"size:512" json:"-"`

	ExpiresAt time.Time `gorm:"not null;index" json:"-"`
}

// TableName pins the table name.
func (SocialLogin) TableName() string {
	return "social_logins"
}

// SocialLoginLifetime is how long a sign-in may take at the provider before
// the answer is no longer accepted.
const SocialLoginLifetime = 15 * time.Minute

// SocialKind is which provider a record is for.
type SocialKind string

const (
	SocialGoogle   SocialKind = "google"
	SocialApple    SocialKind = "apple"
	SocialFacebook SocialKind = "facebook"
	SocialYandex   SocialKind = "yandex"
	SocialVK       SocialKind = "vk"
	// SocialOIDC is any other OpenID Connect provider: Keycloak, Okta,
	// Microsoft Entra, another xermess.
	SocialOIDC SocialKind = "oidc"
	// SocialOAuth2 is a provider that is not OpenID Connect but hands out
	// access tokens and has somewhere to read a profile from.
	SocialOAuth2 SocialKind = "oauth2"
)

// SocialTokenAuth is how a client proves itself at a token endpoint.
type SocialTokenAuth string

const (
	// SocialTokenAuthBasic sends the credentials as HTTP Basic, which RFC
	// 6749 section 2.3.1 says every authorization server must accept.
	SocialTokenAuthBasic SocialTokenAuth = "basic"
	// SocialTokenAuthPost sends them in the request body, which is what the
	// big providers document.
	SocialTokenAuthPost SocialTokenAuth = "post"
)

// SocialTokenAuths lists both, for validation and for the panel's picker.
var SocialTokenAuths = []SocialTokenAuth{SocialTokenAuthBasic, SocialTokenAuthPost}

// SocialClaims says where the identity is in what a provider answers with.
// Each is a path into the JSON: dots step into objects, numbers into arrays,
// so VK's `response.0.id` is as easy to describe as Google's `sub`.
type SocialClaims struct {
	Subject       string `json:"subject"`
	Email         string `json:"email"`
	EmailVerified string `json:"email_verified"`
	FirstName     string `json:"first_name"`
	LastName      string `json:"last_name"`
	// FullName is used when there are no separate names to read.
	FullName string `json:"full_name"`
}

// SocialSpec is everything about a kind that is the same in every
// installation: where to send people, what to ask for, and how to read the
// answer. The catalog is in code rather than in a table for the same reason
// the admin permissions are — a provider that moves an endpoint is a change
// to this file, not to everybody's database.
type SocialSpec struct {
	Kind SocialKind `json:"kind"`
	// Label is what the panel calls it, and the name a new provider is given.
	Label string `json:"label"`

	// The endpoints. Empty in a kind whose provider is not known in advance,
	// where the record carries them instead.
	AuthorizeURL string `json:"authorize_url"`
	TokenURL     string `json:"token_url"`
	UserInfoURL  string `json:"userinfo_url"`

	// Scopes is what to ask for unless the record says otherwise.
	Scopes []string `json:"scopes"`

	// Claims is where to read the identity from.
	Claims SocialClaims `json:"-"`

	// Custom says the endpoints come from the record, and the panel asks for
	// them.
	Custom bool `json:"custom"`

	// IdentityInIDToken reads the identity out of the id_token that came with
	// the access token, for a provider with no userinfo endpoint (Apple).
	IdentityInIDToken bool `json:"-"`

	// IDTokenIssuer is the iss such a token has to name. It is filled in only
	// for the kinds whose identity is read out of an id_token and whose
	// provider is known in advance; a kind configured per installation has no
	// issuer to be held to, and is not.
	IDTokenIssuer string `json:"-"`

	// EmailInTokenResponse reads the address from the token response beside
	// the access token, where the profile endpoint does not give one (VK).
	EmailInTokenResponse bool `json:"-"`

	// UserInfoScheme is how the access token is presented to the profile
	// endpoint: "Bearer" for nearly everyone, "OAuth" for Yandex.
	UserInfoScheme string `json:"-"`

	// AuthorizeParams are added to every authorization request, for the
	// providers that need one: Apple's response_mode, VK's API version.
	AuthorizeParams map[string]string `json:"-"`

	// UserInfoParams are added to the profile request, for an API that is
	// versioned in the query (VK).
	UserInfoParams map[string]string `json:"-"`

	// UserInfoTokenParam passes the access token as a query parameter of that
	// name, for an API that reads it there rather than from a header (VK).
	UserInfoTokenParam string `json:"-"`

	// PKCE sends a code challenge, which every provider here supports except
	// the ones where it is known to break.
	PKCE bool `json:"-"`

	// SignedSecret says the client secret is a JWT this server signs with a
	// registered key, rather than a secret the provider gave out (Apple).
	SignedSecret bool `json:"signed_secret"`

	// TokenAuth is how this kind's provider wants the secret presented. A
	// record of a custom kind may say otherwise.
	TokenAuth SocialTokenAuth `json:"token_auth"`

	// AssumeEmailVerified is for a provider that only ever hands out
	// addresses it has verified itself, and says so nowhere in the answer.
	AssumeEmailVerified bool `json:"-"`

	// Docs is where an administrator registers this server with the provider.
	Docs string `json:"docs"`
}

// SocialSpecs is every kind, in the order the panel offers them.
//
// The endpoints and claim names follow each provider's own documentation;
// the ones that need explaining are explained beside them.
var SocialSpecs = []SocialSpec{
	{
		Kind:         SocialGoogle,
		Label:        "Google",
		AuthorizeURL: "https://accounts.google.com/o/oauth2/v2/auth",
		TokenURL:     "https://oauth2.googleapis.com/token",
		UserInfoURL:  "https://openidconnect.googleapis.com/v1/userinfo",
		Scopes:       []string{"openid", "email", "profile"},
		Claims: SocialClaims{
			Subject:       "sub",
			Email:         "email",
			EmailVerified: "email_verified",
			FirstName:     "given_name",
			LastName:      "family_name",
			FullName:      "name",
		},
		UserInfoScheme: "Bearer",
		PKCE:           true,
		TokenAuth:      SocialTokenAuthPost,
		Docs:           "https://console.cloud.google.com/apis/credentials",
	},
	{
		Kind:  SocialApple,
		Label: "Apple",
		// Apple has no userinfo endpoint: everything it will say about the
		// person is in the id_token, and the name only the first time.
		AuthorizeURL: "https://appleid.apple.com/auth/authorize",
		TokenURL:     "https://appleid.apple.com/auth/token",
		Scopes:       []string{"name", "email"},
		Claims: SocialClaims{
			Subject:       "sub",
			Email:         "email",
			EmailVerified: "email_verified",
		},
		IdentityInIDToken: true,
		IDTokenIssuer:     "https://appleid.apple.com",
		// Asking for a name or an address makes Apple post the answer back
		// instead of redirecting with it in the query.
		AuthorizeParams: map[string]string{"response_mode": "form_post"},
		PKCE:            true,
		SignedSecret:    true,
		TokenAuth:       SocialTokenAuthPost,
		Docs:            "https://developer.apple.com/account/resources/identifiers",
	},
	{
		Kind:         SocialFacebook,
		Label:        "Facebook",
		AuthorizeURL: "https://www.facebook.com/v21.0/dialog/oauth",
		TokenURL:     "https://graph.facebook.com/v21.0/oauth/access_token",
		// Graph returns only the fields that are asked for.
		UserInfoURL: "https://graph.facebook.com/v21.0/me?fields=id,email,first_name,last_name,name",
		Scopes:      []string{"email", "public_profile"},
		Claims: SocialClaims{
			Subject:   "id",
			Email:     "email",
			FirstName: "first_name",
			LastName:  "last_name",
			FullName:  "name",
		},
		UserInfoScheme: "Bearer",
		PKCE:           true,
		// Facebook hands out an address only once it has confirmed it, and
		// says nothing about it in the answer.
		AssumeEmailVerified: true,
		TokenAuth:           SocialTokenAuthPost,
		Docs:                "https://developers.facebook.com/apps",
	},
	{
		Kind:         SocialYandex,
		Label:        "Yandex ID",
		AuthorizeURL: "https://oauth.yandex.ru/authorize",
		TokenURL:     "https://oauth.yandex.ru/token",
		UserInfoURL:  "https://login.yandex.ru/info?format=json",
		Scopes:       []string{"login:email", "login:info"},
		Claims: SocialClaims{
			Subject:   "id",
			Email:     "default_email",
			FirstName: "first_name",
			LastName:  "last_name",
			FullName:  "real_name",
		},
		// Yandex reads its own scheme rather than Bearer.
		UserInfoScheme:      "OAuth",
		PKCE:                true,
		AssumeEmailVerified: true,
		TokenAuth:           SocialTokenAuthPost,
		Docs:                "https://oauth.yandex.ru/client/new",
	},
	{
		Kind:         SocialVK,
		Label:        "VK ID",
		AuthorizeURL: "https://oauth.vk.com/authorize",
		TokenURL:     "https://oauth.vk.com/access_token",
		UserInfoURL:  "https://api.vk.com/method/users.get",
		Scopes:       []string{"email"},
		Claims: SocialClaims{
			// users.get answers with a list under "response".
			Subject:   "response.0.id",
			FirstName: "response.0.first_name",
			LastName:  "response.0.last_name",
		},
		// VK's API reads the token and its version from the query.
		UserInfoParams:     map[string]string{"v": "5.131"},
		UserInfoTokenParam: "access_token",
		// VK sends the address back with the access token, not in the profile.
		EmailInTokenResponse: true,
		AuthorizeParams:      map[string]string{"v": "5.131"},
		AssumeEmailVerified:  true,
		TokenAuth:            SocialTokenAuthPost,
		Docs:                 "https://vk.com/editapp?act=create",
	},
	{
		Kind:  SocialOIDC,
		Label: "OpenID Connect",
		Scopes: []string{
			ScopeOpenID, ScopeEmail, ScopeProfile,
		},
		Claims: SocialClaims{
			Subject:       "sub",
			Email:         "email",
			EmailVerified: "email_verified",
			FirstName:     "given_name",
			LastName:      "family_name",
			FullName:      "name",
		},
		Custom:         true,
		UserInfoScheme: "Bearer",
		PKCE:           true,
		TokenAuth:      SocialTokenAuthBasic,
		Docs:           "https://openid.net/developers/how-connect-works/",
	},
	{
		Kind:   SocialOAuth2,
		Label:  "OAuth 2.0",
		Scopes: []string{},
		Claims: SocialClaims{
			Subject:   "id",
			Email:     "email",
			FirstName: "first_name",
			LastName:  "last_name",
			FullName:  "name",
		},
		Custom:         true,
		UserInfoScheme: "Bearer",
		PKCE:           true,
		TokenAuth:      SocialTokenAuthPost,
		Docs:           "https://datatracker.ietf.org/doc/html/rfc6749",
	},
}

// SocialSpecFor returns what is known about a kind.
func SocialSpecFor(kind SocialKind) (SocialSpec, bool) {
	for _, spec := range SocialSpecs {
		if spec.Kind == kind {
			return spec, true
		}
	}

	return SocialSpec{}, false
}

// Spec is what is known about this provider's kind.
func (p SocialProvider) Spec() SocialSpec {
	spec, _ := SocialSpecFor(p.Kind)
	return spec
}

// Endpoints are the addresses to use for this provider: the kind's, or the
// record's own where the kind has none.
func (p SocialProvider) Endpoints() (authorize, token, userInfo string) {
	spec := p.Spec()
	if spec.Custom {
		return p.AuthorizeURL, p.TokenURL, p.UserInfoURL
	}

	return spec.AuthorizeURL, spec.TokenURL, spec.UserInfoURL
}

// TokenAuthMethod is how to present the secret at the token endpoint: what
// the record says, or the kind's own way.
func (p SocialProvider) TokenAuthMethod() SocialTokenAuth {
	if p.TokenAuth != "" {
		return p.TokenAuth
	}

	if auth := p.Spec().TokenAuth; auth != "" {
		return auth
	}

	return SocialTokenAuthBasic
}

// AskedScopes is what to ask the provider for: what the record says, or the
// kind's own defaults.
func (p SocialProvider) AskedScopes() []string {
	if len(p.Scopes) > 0 {
		return p.Scopes
	}

	return p.Spec().Scopes
}

// CallbackURL is where the provider sends the browser back to, which is also
// what has to be registered with them. It is built from the issuer so it is
// the same address tokens are issued under.
func (p SocialProvider) CallbackURL(issuer string) string {
	return strings.TrimRight(issuer, "/") + "/oauth2/social/" + p.Slug + "/callback"
}

// socialSlugPattern is what a provider may be called in an address: lower
// case letters, numbers and dashes.
var socialSlugPattern = regexp.MustCompile(`^[a-z0-9]([a-z0-9-]*[a-z0-9])?$`)

// Validate reports the first thing wrong with a provider.
//
// It is about the record alone: whether the provider will accept these
// credentials is only known once somebody signs in with them.
func (p SocialProvider) Validate() error {
	spec, known := SocialSpecFor(p.Kind)
	if !known {
		return fmt.Errorf("kind must be one of: %s", strings.Join(socialKindNames(), ", "))
	}

	switch {
	case p.Slug == "":
		return fmt.Errorf("slug is required")
	case len(p.Slug) > 64 || !socialSlugPattern.MatchString(p.Slug):
		return fmt.Errorf("slug must be lower case letters, numbers and dashes, such as google-workspace")
	}

	switch {
	case strings.TrimSpace(p.Name) == "":
		return fmt.Errorf("name is required")
	case len(p.Name) > 100:
		return fmt.Errorf("name must be at most 100 characters")
	}

	switch {
	case strings.TrimSpace(p.ClientID) == "":
		return fmt.Errorf("client_id is required")
	case len(p.ClientID) > 255:
		return fmt.Errorf("client_id must be at most 255 characters")
	}

	for _, field := range []struct{ name, value string }{
		{"team_id", p.TeamID},
		{"key_id", p.KeyID},
	} {
		if len(field.value) > 64 {
			return fmt.Errorf("%s must be at most 64 characters", field.name)
		}
	}

	if spec.SignedSecret {
		switch {
		case strings.TrimSpace(p.TeamID) == "":
			return fmt.Errorf("team_id is required for %s", spec.Label)
		case strings.TrimSpace(p.KeyID) == "":
			return fmt.Errorf("key_id is required for %s", spec.Label)
		case len(p.PrivateKey) == 0:
			return fmt.Errorf("private_key is required for %s: the .p8 file from the developer account", spec.Label)
		}
	} else if len(p.ClientSecret) == 0 {
		return fmt.Errorf("client_secret is required")
	}

	if spec.Custom {
		for _, endpoint := range []struct{ field, value string }{
			{"authorize_url", p.AuthorizeURL},
			{"token_url", p.TokenURL},
		} {
			if err := socialEndpoint(endpoint.field, endpoint.value, true); err != nil {
				return err
			}
		}

		// OpenID Connect can answer with an id_token instead; plain OAuth 2.0
		// has nowhere else to read a profile from.
		if err := socialEndpoint("userinfo_url", p.UserInfoURL, p.Kind == SocialOAuth2); err != nil {
			return err
		}
	}

	if p.TokenAuth != "" && !slices.Contains(SocialTokenAuths, p.TokenAuth) {
		return fmt.Errorf("token_auth must be basic or post")
	}

	for _, scope := range p.Scopes {
		if strings.ContainsAny(scope, " \t\r\n") {
			return fmt.Errorf("scopes: %q must not contain spaces — list them one at a time", scope)
		}
	}

	return nil
}

// socialEndpoint checks one address. An endpoint is http(s) and absolute;
// http is allowed because a provider on the same network as this server in a
// test or on an internal deployment has no certificate of its own.
func socialEndpoint(field, value string, required bool) error {
	if value == "" {
		if required {
			return fmt.Errorf("%s is required for this kind of provider", field)
		}
		return nil
	}

	if len(value) > 512 {
		return fmt.Errorf("%s must be at most 512 characters", field)
	}

	parsed, err := url.Parse(value)
	if err != nil || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return fmt.Errorf("%s must be a full address starting with http:// or https://", field)
	}

	return nil
}

func socialKindNames() []string {
	names := make([]string, 0, len(SocialSpecs))
	for _, spec := range SocialSpecs {
		names = append(names, string(spec.Kind))
	}

	return names
}

// DefaultSocialProvider is a new provider of a kind, as the panel offers it
// before anything is typed in: the kind's name, a slug to match, and the
// settings an installation almost always wants.
func DefaultSocialProvider(kind SocialKind) SocialProvider {
	spec, _ := SocialSpecFor(kind)

	return SocialProvider{
		Kind:               kind,
		Slug:               string(kind),
		Name:               spec.Label,
		Scopes:             slices.Clone(spec.Scopes),
		Enabled:            true,
		LinkVerifiedEmails: true,
		AllowRegistration:  true,
	}
}
