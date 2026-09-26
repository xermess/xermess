package oidc

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"maps"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"sync"
	"time"

	"loginer/internal/jose"
	"loginer/internal/model"
)

// The OpenID Connect half of enterprise single sign-on (sso.go): reading a
// provider's discovery document and keys, sending the browser there, and
// turning the code it comes back with into a verified id_token.

// oidcDiscovery is the part of a provider's discovery document this server
// reads.
type oidcDiscovery struct {
	Issuer                string   `json:"issuer"`
	AuthorizationEndpoint string   `json:"authorization_endpoint"`
	TokenEndpoint         string   `json:"token_endpoint"`
	UserInfoEndpoint      string   `json:"userinfo_endpoint"`
	JWKSURI               string   `json:"jwks_uri"`
	TokenAuthMethods      []string `json:"token_endpoint_auth_methods_supported"`
	ScopesSupported       []string `json:"scopes_supported"`
}

// ssoCache keeps each provider's discovery document and keys for a while,
// rather than asking for them on every sign-in.
type ssoCache struct {
	mu        sync.Mutex
	discovery map[string]cached[*oidcDiscovery]
	keys      map[string]cached[map[string]jose.Key]
}

type cached[T any] struct {
	value   T
	fetched time.Time
}

// ssoCacheLifetime is how long a discovery document or a key set is kept. A
// key the cache does not have is fetched at once, so rotating keys is not
// held up by it.
const ssoCacheLifetime = time.Hour

var ssoMemory = &ssoCache{
	discovery: map[string]cached[*oidcDiscovery]{},
	keys:      map[string]cached[map[string]jose.Key]{},
}

// Discover reads an issuer's discovery document, and checks it is the
// issuer's own. The panel's test uses it; a sign-in reads it from the cache.
func (s *Service) Discover(ctx context.Context, issuer string) (*oidcDiscovery, error) {
	issuer = strings.TrimRight(issuer, "/")

	ssoMemory.mu.Lock()
	if hit, ok := ssoMemory.discovery[issuer]; ok && s.now().Sub(hit.fetched) < ssoCacheLifetime {
		ssoMemory.mu.Unlock()
		return hit.value, nil
	}
	ssoMemory.mu.Unlock()

	var document oidcDiscovery
	if err := s.getJSON(ctx, issuer+"/.well-known/openid-configuration", &document); err != nil {
		return nil, err
	}

	// The document has to be the issuer's own (OpenID Connect Discovery
	// section 4.3): one that names another issuer would have its tokens
	// believed on this one's word.
	switch {
	case strings.TrimRight(document.Issuer, "/") != issuer:
		return nil, fmt.Errorf("the discovery document is for %q, not %q", document.Issuer, issuer)
	case document.AuthorizationEndpoint == "" || document.TokenEndpoint == "" || document.JWKSURI == "":
		return nil, fmt.Errorf("the discovery document is missing an endpoint")
	}

	ssoMemory.mu.Lock()
	ssoMemory.discovery[issuer] = cached[*oidcDiscovery]{value: &document, fetched: s.now()}
	ssoMemory.mu.Unlock()

	return &document, nil
}

func (s *Service) startOIDC(ctx context.Context, connection *model.SSOConnection, login *model.SSOLogin, state, loginHint string) (string, error) {
	document, err := s.Discover(ctx, connection.Issuer)
	if err != nil {
		return "", upstream("discovery", err)
	}

	verifier, _, err := model.NewSecret()
	if err != nil {
		return "", err
	}
	nonce, _, err := model.NewSecret()
	if err != nil {
		return "", err
	}
	login.Verifier, login.Nonce = verifier, nonce

	values := url.Values{
		"response_type":         {"code"},
		"client_id":             {connection.ClientID},
		"redirect_uri":          {connection.CallbackURL(s.issuer)},
		"scope":                 {strings.Join(connection.AskedScopes(), " ")},
		"state":                 {state},
		"nonce":                 {nonce},
		"code_challenge":        {challengeFor(verifier)},
		"code_challenge_method": {model.PKCES256},
	}
	if loginHint != "" {
		values.Set("login_hint", loginHint)
	}

	return withQuery(document.AuthorizationEndpoint, values), nil
}

func (s *Service) completeOIDC(ctx context.Context, connection *model.SSOConnection, login *model.SSOLogin, code string) (*ssoPerson, error) {
	document, err := s.Discover(ctx, connection.Issuer)
	if err != nil {
		return nil, upstream("discovery", err)
	}

	secret := ""
	if len(connection.ClientSecret) > 0 {
		plain, err := s.sealer.OpenBytes(connection.ClientSecret)
		if err != nil {
			return nil, err
		}
		secret = string(plain)
	}

	form := url.Values{
		"grant_type":    {"authorization_code"},
		"code":          {code},
		"redirect_uri":  {connection.CallbackURL(s.issuer)},
		"code_verifier": {login.Verifier},
	}

	// HTTP Basic, which every server has to accept (RFC 6749 section
	// 2.3.1), unless the provider says it only takes the body.
	basic := len(document.TokenAuthMethods) == 0 || slices.Contains(document.TokenAuthMethods, "client_secret_basic")
	if !basic {
		form.Set("client_id", connection.ClientID)
		form.Set("client_secret", secret)
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, document.TokenEndpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Set("Accept", "application/json")
	if basic {
		request.SetBasicAuth(url.QueryEscape(connection.ClientID), url.QueryEscape(secret))
	}

	var token struct {
		AccessToken string `json:"access_token"`
		IDToken     string `json:"id_token"`
	}
	if err := s.doJSON(request, &token); err != nil {
		return nil, upstream("token", err)
	}

	claims, err := s.verifyIDToken(ctx, connection, document, login, token.IDToken)
	if err != nil {
		return nil, upstream("id_token", err)
	}

	// Some providers keep the address or the groups to the userinfo endpoint.
	if claimFirst(claims, connection.Attribute("email")) == "" && document.UserInfoEndpoint != "" && token.AccessToken != "" {
		if extra, err := s.oidcUserInfo(ctx, document.UserInfoEndpoint, token.AccessToken); err == nil &&
			stringClaim(extra, "sub") == stringClaim(claims, "sub") {
			for name, value := range extra {
				if _, has := claims[name]; !has {
					claims[name] = value
				}
			}
		}
	}

	// A provider that says outright it has not verified the address is not
	// vouching for it.
	email := claimFirst(claims, connection.Attribute("email"))
	if verified, said := claims["email_verified"].(bool); said && !verified && email != "" {
		return nil, refused(ErrSSONoEmail, "the provider says it has not verified %s (email_verified is false)", email)
	}

	return &ssoPerson{
		Subject:   stringClaim(claims, "sub"),
		Email:     email,
		FirstName: claimFirst(claims, connection.Attribute("first_name")),
		LastName:  claimFirst(claims, connection.Attribute("last_name")),
		Groups:    claimList(claims, connection.Attribute("groups")),
		Sent:      slices.Sorted(maps.Keys(claims)),
	}, nil
}

// idTokenLeeway is how far apart two clocks may be.
const idTokenLeeway = time.Minute

// verifyIDToken checks an id_token is the provider's, for this client, in
// time, and for this sign-in (OpenID Connect Core section 3.1.3.7).
func (s *Service) verifyIDToken(
	ctx context.Context,
	connection *model.SSOConnection,
	document *oidcDiscovery,
	login *model.SSOLogin,
	idToken string,
) (map[string]any, error) {
	if idToken == "" {
		return nil, fmt.Errorf("no id_token")
	}

	keys, err := s.providerKeys(ctx, document.JWKSURI, false)
	if err != nil {
		return nil, err
	}

	lookup := func(keys map[string]jose.Key) func(string) (jose.Key, bool) {
		return func(kid string) (jose.Key, bool) {
			if key, ok := keys[kid]; ok {
				return key, true
			}
			// A token that names no key, from a provider with one.
			if kid == "" && len(keys) == 1 {
				for _, key := range keys {
					return key, true
				}
			}
			return jose.Key{}, false
		}
	}

	var claims map[string]any
	if _, err := jose.Verify(idToken, lookup(keys), &claims); err != nil {
		// Perhaps the provider has rotated to a key we have not seen yet.
		if keys, err = s.providerKeys(ctx, document.JWKSURI, true); err != nil {
			return nil, err
		}
		if _, err := jose.Verify(idToken, lookup(keys), &claims); err != nil {
			return nil, fmt.Errorf("signature: %w", err)
		}
	}

	now := s.now()

	switch {
	case stringClaim(claims, "iss") != document.Issuer:
		return nil, fmt.Errorf("issued by %q", stringClaim(claims, "iss"))
	case !audienceHas(claims["aud"], connection.ClientID):
		return nil, fmt.Errorf("for another client")
	case stringClaim(claims, "nonce") != login.Nonce:
		return nil, fmt.Errorf("for another sign-in")
	}

	if azp := stringClaim(claims, "azp"); azp != "" && azp != connection.ClientID {
		return nil, fmt.Errorf("authorised for another client")
	}

	exp, ok := numberClaim(claims["exp"])
	if !ok || now.After(time.Unix(int64(exp), 0).Add(idTokenLeeway)) {
		return nil, fmt.Errorf("expired")
	}
	if iat, ok := numberClaim(claims["iat"]); ok && time.Unix(int64(iat), 0).After(now.Add(idTokenLeeway)) {
		return nil, fmt.Errorf("issued in the future")
	}

	return claims, nil
}

// providerKeys are the keys a provider signs with, by id.
func (s *Service) providerKeys(ctx context.Context, jwksURI string, refresh bool) (map[string]jose.Key, error) {
	ssoMemory.mu.Lock()
	if hit, ok := ssoMemory.keys[jwksURI]; ok && !refresh && s.now().Sub(hit.fetched) < ssoCacheLifetime {
		ssoMemory.mu.Unlock()
		return hit.value, nil
	}
	ssoMemory.mu.Unlock()

	var set struct {
		Keys []jose.JWK `json:"keys"`
	}
	if err := s.getJSON(ctx, jwksURI, &set); err != nil {
		return nil, err
	}

	keys := map[string]jose.Key{}
	for _, jwk := range set.Keys {
		if jwk.Use != "" && jwk.Use != "sig" {
			continue
		}
		// A key of a kind this server cannot verify with is skipped, not an
		// error: the provider may publish others beside the one it uses.
		if key, err := jose.ParseJWK(jwk); err == nil {
			keys[key.ID] = key
		}
	}

	if len(keys) == 0 {
		return nil, fmt.Errorf("the provider publishes no key this server can verify with")
	}

	ssoMemory.mu.Lock()
	ssoMemory.keys[jwksURI] = cached[map[string]jose.Key]{value: keys, fetched: s.now()}
	ssoMemory.mu.Unlock()

	return keys, nil
}

func (s *Service) oidcUserInfo(ctx context.Context, endpoint, accessToken string) (map[string]any, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("Authorization", "Bearer "+accessToken)
	request.Header.Set("Accept", "application/json")

	var claims map[string]any
	return claims, s.doJSON(request, &claims)
}

// getJSON reads a JSON document from a provider.
func (s *Service) getJSON(ctx context.Context, address string, into any) error {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, address, nil)
	if err != nil {
		return err
	}
	request.Header.Set("Accept", "application/json")

	return s.doJSON(request, into)
}

// doJSON makes a request to a provider and decodes its JSON answer, within the
// time a provider is given and no bigger than a provider's answer should be.
func (s *Service) doJSON(request *http.Request, into any) error {
	ctx, cancel := context.WithTimeout(request.Context(), socialTimeout)
	defer cancel()

	response, err := s.social.Do(request.WithContext(ctx))
	if err != nil {
		return err
	}
	defer response.Body.Close()

	body, err := io.ReadAll(io.LimitReader(response.Body, 1<<20))
	if err != nil {
		return err
	}
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("%s answered %d: %.200s", request.URL.Host, response.StatusCode, body)
	}

	return json.Unmarshal(body, into)
}

// claimFirst is the first of several claims that has a value.
func claimFirst(claims map[string]any, names []string) string {
	for _, name := range names {
		if value := claimString(claims, name); value != "" {
			return value
		}
	}

	return ""
}

// claimList is a claim that is a list of strings — or one string, which some
// providers send for a single group.
func claimList(claims map[string]any, names []string) []string {
	for _, name := range names {
		switch value := claimAt(claims, name).(type) {
		case []any:
			var out []string
			for _, item := range value {
				if text, ok := item.(string); ok && text != "" {
					out = append(out, text)
				}
			}
			return out
		case string:
			if value != "" {
				return []string{value}
			}
		}
	}

	return nil
}
