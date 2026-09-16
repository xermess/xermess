// Package oidc is the OAuth 2.0 authorization server and OpenID Connect
// provider: what happens behind the authorization, token, userinfo, logout,
// revocation and introspection endpoints, and behind the sign-in pages users
// reach from an application.
//
// It knows nothing of HTTP. The handlers in internal/api/oauth and
// internal/api/account read requests, call a method here, and write what it
// returns; the rules live here, in one place, and what a token carries is
// still decided by model.EvaluateToken alone — the same function the panel's
// token preview runs.
package oidc

import (
	"context"
	"log/slog"
	"net/url"
	"time"

	"xermess/internal/config"
	"xermess/internal/jose"
	"xermess/internal/mail"
	"xermess/internal/model"
	"xermess/internal/store"
)

// The paths the provider serves, relative to the issuer. They are here rather
// than only in the route table because the discovery document and the
// redirects name them too.
const (
	PathDiscovery  = "/.well-known/openid-configuration"
	PathJWKS       = "/.well-known/jwks.json"
	PathAuthorize  = "/oauth2/authorize"
	PathToken      = "/oauth2/token"
	PathUserInfo   = "/oauth2/userinfo"
	PathLogout     = "/oauth2/logout"
	PathRevoke     = "/oauth2/revoke"
	PathIntrospect = "/oauth2/introspect"
)

// The pages of the id app the provider sends browsers to, relative to its
// URL.
const (
	PageLogin     = "/login"
	PageReset     = "/reset-password"
	PageError     = "/error"
	PageLoggedOut = "/logged-out"
)

// Lockout for users, the same as for administrators.
const (
	maxFailedLogins = 5
	lockoutDuration = 15 * time.Minute
)

// Service is the provider.
type Service struct {
	store      *store.Store
	keys       *keySet
	mail       mail.Sender
	log        *slog.Logger
	issuer     string
	accountURL string

	// now is time.Now, and a test's clock when it needs one.
	now func() time.Time
}

// New builds the provider, loading its signing keys and making any that are
// missing.
func New(ctx context.Context, cfg config.Config, st *store.Store, mailer mail.Sender, log *slog.Logger) (*Service, error) {
	sealer, err := jose.NewSealer(cfg.SecretKey)
	if err != nil {
		return nil, err
	}

	keys, err := loadKeys(ctx, st, sealer, cfg.KeyRotation, log)
	if err != nil {
		return nil, err
	}

	return &Service{
		store:      st,
		keys:       keys,
		mail:       mailer,
		log:        log,
		issuer:     cfg.Issuer,
		accountURL: cfg.AccountURL,
		now:        time.Now,
	}, nil
}

// MaintainKeys keeps the signing keys rotating until ctx ends: it makes the
// next key when one is due, retires and deletes old ones, and picks up what
// other servers changed. main runs it for as long as the server serves.
func (s *Service) MaintainKeys(ctx context.Context) {
	ticker := time.NewTicker(keyMaintenanceInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := s.keys.maintain(ctx); err != nil && ctx.Err() == nil {
				s.log.Error("maintaining signing keys failed", "error", err)
			}
		}
	}
}

// SigningKeys describes every published key, newest first.
func (s *Service) SigningKeys() []KeyInfo {
	return s.keys.describe()
}

// RotateKeys makes new signing keys now; see keySet.Rotate.
func (s *Service) RotateKeys(ctx context.Context, immediate, revoke bool) error {
	return s.keys.Rotate(ctx, immediate, revoke)
}

// Issuer is the iss of every token.
func (s *Service) Issuer() string {
	return s.issuer
}

// Discovery is the OpenID Connect Discovery 1.0 document.
type Discovery struct {
	Issuer                                     string   `json:"issuer"`
	AuthorizationEndpoint                      string   `json:"authorization_endpoint"`
	TokenEndpoint                              string   `json:"token_endpoint"`
	UserInfoEndpoint                           string   `json:"userinfo_endpoint"`
	EndSessionEndpoint                         string   `json:"end_session_endpoint"`
	RevocationEndpoint                         string   `json:"revocation_endpoint"`
	IntrospectionEndpoint                      string   `json:"introspection_endpoint"`
	JWKSURI                                    string   `json:"jwks_uri"`
	ResponseTypesSupported                     []string `json:"response_types_supported"`
	ResponseModesSupported                     []string `json:"response_modes_supported"`
	GrantTypesSupported                        []string `json:"grant_types_supported"`
	SubjectTypesSupported                      []string `json:"subject_types_supported"`
	IDTokenSigningAlgValuesSupported           []string `json:"id_token_signing_alg_values_supported"`
	ScopesSupported                            []string `json:"scopes_supported"`
	ClaimsSupported                            []string `json:"claims_supported"`
	TokenEndpointAuthMethodsSupported          []string `json:"token_endpoint_auth_methods_supported"`
	RevocationEndpointAuthMethodsSupported     []string `json:"revocation_endpoint_auth_methods_supported"`
	IntrospectionEndpointAuthMethodsSupported  []string `json:"introspection_endpoint_auth_methods_supported"`
	CodeChallengeMethodsSupported              []string `json:"code_challenge_methods_supported"`
	PromptValuesSupported                      []string `json:"prompt_values_supported"`
	AuthorizationResponseIssParameterSupported bool     `json:"authorization_response_iss_parameter_supported"`

	// The organisation's agreements, which OpenID Provider Metadata names
	// op_tos_uri and op_policy_uri: what a client's users accept by signing
	// in here. Left out when the organisation has published neither.
	OpTosURI    string `json:"op_tos_uri,omitempty"`
	OpPolicyURI string `json:"op_policy_uri,omitempty"`

	ClaimsParameterSupported     bool `json:"claims_parameter_supported"`
	RequestParameterSupported    bool `json:"request_parameter_supported"`
	RequestURIParameterSupported bool `json:"request_uri_parameter_supported"`
}

// Discovery describes the provider. Every list is read from the model, so it
// cannot claim a grant, scope or method the server does not offer.
//
// The organisation's terms and privacy links are read from the database: they
// are settings an administrator changes in the panel, and the document has to
// say what they are now rather than what they were when the server started.
// A database that cannot be reached still yields a document — the rest of it
// is what a client needs to reach the endpoints at all — with those two left
// out and the failure logged.
func (s *Service) Discovery(ctx context.Context) Discovery {
	secretMethods := []string{string(model.AuthClientSecretBasic), string(model.AuthClientSecretPost)}

	var tos, policy string
	if organization, err := s.store.Organization(ctx); err == nil {
		tos, policy = organization.TermsURL, organization.PrivacyURL
	} else {
		s.log.Error("reading the organization for the discovery document failed", "error", err)
	}

	return Discovery{
		Issuer:                                     s.issuer,
		AuthorizationEndpoint:                      s.issuer + PathAuthorize,
		TokenEndpoint:                              s.issuer + PathToken,
		UserInfoEndpoint:                           s.issuer + PathUserInfo,
		EndSessionEndpoint:                         s.issuer + PathLogout,
		RevocationEndpoint:                         s.issuer + PathRevoke,
		IntrospectionEndpoint:                      s.issuer + PathIntrospect,
		JWKSURI:                                    s.issuer + PathJWKS,
		ResponseTypesSupported:                     []string{"code"},
		ResponseModesSupported:                     []string{"query"},
		GrantTypesSupported:                        model.GrantTypes,
		SubjectTypesSupported:                      []string{"public"},
		IDTokenSigningAlgValuesSupported:           []string{idTokenAlgorithm},
		ScopesSupported:                            model.Scopes,
		ClaimsSupported:                            claimsSupported,
		TokenEndpointAuthMethodsSupported:          append(secretMethods, string(model.AuthNone)),
		RevocationEndpointAuthMethodsSupported:     append(append([]string{}, secretMethods...), string(model.AuthNone)),
		IntrospectionEndpointAuthMethodsSupported:  secretMethods,
		CodeChallengeMethodsSupported:              []string{model.PKCES256},
		PromptValuesSupported:                      []string{"none", "login"},
		AuthorizationResponseIssParameterSupported: true,
		OpTosURI:    tos,
		OpPolicyURI: policy,
	}
}

var claimsSupported = []string{
	"iss", "sub", "aud", "exp", "iat", "auth_time", "nonce", "at_hash", "sid",
	"azp", "client_id", "scope", "jti",
	"name", "given_name", "family_name", "email", "email_verified",
	model.ClaimApplicationRoles, model.ClaimGlobalRoles,
}

// JWKS is the JSON Web Key Set.
type JWKS struct {
	Keys []jose.JWK `json:"keys"`
}

// JWKS returns every public key tokens may be signed with.
func (s *Service) JWKS(ctx context.Context) (JWKS, error) {
	keys, err := s.keys.public(ctx)
	return JWKS{Keys: keys}, err
}

// Client is where a browser request came from, for the session and the log.
type Client struct {
	IP        string
	UserAgent string
}

// withQuery appends parameters to a URL that may already have a query.
func withQuery(base string, values url.Values) string {
	u, err := url.Parse(base)
	if err != nil {
		return base
	}

	query := u.Query()
	for key, vals := range values {
		for _, v := range vals {
			if v != "" {
				query.Add(key, v)
			}
		}
	}
	u.RawQuery = query.Encode()

	return u.String()
}

// errorPage is the sign-in app's page for an authorization error that cannot
// be sent back to the application: an unknown client, or a redirect URI that
// is not registered, where redirecting would hand the error — and the user —
// to whoever wrote the link.
func (s *Service) errorPage(code, description string) string {
	return withQuery(s.accountURL+PageError, url.Values{"error": {code}, "error_description": {description}})
}

// record writes a user's sign-in activity to the log. As for administrators,
// failing to write it does not fail what it records.
func (s *Service) record(ctx context.Context, user *model.User, email, action string, client Client, metadata map[string]any) {
	entry := model.AuditLog{
		ActorEmail: email,
		Action:     action,
		TargetType: "user",
		IP:         client.IP,
		UserAgent:  client.UserAgent,
		Metadata:   metadata,
	}
	if user != nil {
		entry.TargetID = user.ID.String()
	}

	if err := s.store.WriteAudit(ctx, &entry); err != nil {
		s.log.Error("writing the activity log failed", "error", err, "action", action)
	}
}
