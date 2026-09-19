package oidc

import (
	"context"
	"crypto"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"xermess/internal/jose"
	"xermess/internal/model"
	"xermess/internal/store"
	"xermess/locales"
)

// Signing in with an account somewhere else.
//
// This server is the client here, not the provider: it sends the browser to
// Google or Yandex, they send it back with a code, and this file turns that
// code into a session — the same session a password would have started, so
// everything after it is the same.
//
// The paths are the two halves of that, and they carry the provider's slug so
// one server can offer several. The callback is registered with the provider,
// so it may not move once anybody is signing in with it.
const (
	PathSocialStart    = "/oauth2/social/:slug/start"
	PathSocialCallback = "/oauth2/social/:slug/callback"
)

// socialTimeout is how long the provider has to answer each of the two
// requests this server makes to it.
const socialTimeout = 15 * time.Second

// SocialButton is a provider as a sign-in page shows it.
type SocialButton struct {
	Slug string `json:"slug"`
	Name string `json:"name"`
	Kind string `json:"kind"`
}

// SocialButtons is what the sign-in pages offer, in the order they are shown.
func (s *Service) SocialButtons(ctx context.Context) ([]SocialButton, error) {
	stored, err := s.store.SocialButtons(ctx)
	if err != nil {
		return nil, err
	}

	buttons := make([]SocialButton, 0, len(stored))
	for _, button := range stored {
		buttons = append(buttons, SocialButton(button))
	}

	return buttons, nil
}

// StartSocial is where to send the browser to sign in with a provider.
//
// `request` is the sign-in under way, if the person came from an application,
// and `next` is where to put them afterwards when they did not. Both are kept
// here rather than in the address the provider is given, which comes back
// only as far as `state`.
func (s *Service) StartSocial(ctx context.Context, slug, request, next string) (string, error) {
	provider, err := s.socialProvider(ctx, slug)
	if err != nil {
		return "", err
	}
	spec := provider.Spec()

	state, stateHash, err := model.NewSecret()
	if err != nil {
		return "", err
	}

	values := url.Values{
		"response_type": {"code"},
		"client_id":     {provider.ClientID},
		"redirect_uri":  {provider.CallbackURL(s.issuer)},
		"state":         {state},
	}

	if scopes := provider.AskedScopes(); len(scopes) > 0 {
		values.Set("scope", strings.Join(scopes, " "))
	}

	// PKCE binds the code to this sign-in, so a code taken out of a redirect
	// cannot be spent by anyone else (RFC 9700 section 2.1.1).
	var verifier string
	if spec.PKCE {
		if verifier, _, err = model.NewSecret(); err != nil {
			return "", err
		}

		values.Set("code_challenge", challengeFor(verifier))
		values.Set("code_challenge_method", model.PKCES256)
	}

	for key, value := range spec.AuthorizeParams {
		values.Set(key, value)
	}

	login := model.SocialLogin{
		StateHash:  stateHash,
		ProviderID: provider.ID,
		Verifier:   verifier,
		Request:    request,
		Next:       next,
		ExpiresAt:  s.now().Add(model.SocialLoginLifetime),
	}
	if err := s.store.CreateSocialLogin(ctx, &login); err != nil {
		return "", err
	}

	authorize, _, _ := provider.Endpoints()

	return withQuery(authorize, values), nil
}

// SocialResult is a finished sign-in at a provider: the session it started,
// and where the person was going when they left.
type SocialResult struct {
	SignIn *SignInResult
	// Request is the sign-in under way to continue, if there was one.
	Request string
	// Next is where to send the browser when there is no application waiting.
	Next string
	// Provider is the name to say in a message, such as "Signed in with
	// Google".
	Provider string
}

// CompleteSocial turns the code a provider sent the browser back with into a
// session.
func (s *Service) CompleteSocial(ctx context.Context, slug, code, state string, client Client) (*SocialResult, error) {
	provider, err := s.socialProvider(ctx, slug)
	if err != nil {
		return nil, err
	}

	if code == "" || state == "" {
		return nil, ErrSocialExpired
	}

	login, err := s.store.TakeSocialLogin(ctx, model.HashSecret(state), s.now())
	switch {
	case errors.Is(err, store.ErrNotFound):
		return nil, ErrSocialExpired
	case err != nil:
		return nil, err
	}

	// The state was made for a sign-in at this provider and no other.
	if login.ProviderID != provider.ID {
		return nil, ErrSocialExpired
	}

	token, err := s.exchangeSocialCode(ctx, provider, code, login.Verifier)
	if err != nil {
		return nil, err
	}

	identity, err := s.socialIdentity(ctx, provider, token)
	if err != nil {
		return nil, err
	}

	result, err := s.signInWithIdentity(ctx, provider, identity, login.Request, client)
	if err != nil {
		return nil, err
	}

	return &SocialResult{
		SignIn:   result,
		Request:  login.Request,
		Next:     login.Next,
		Provider: provider.Name,
	}, nil
}

// socialProvider is a provider by slug, if it is one users may sign in with.
func (s *Service) socialProvider(ctx context.Context, slug string) (*model.SocialProvider, error) {
	provider, err := s.store.SocialProviderBySlug(ctx, slug)
	switch {
	case errors.Is(err, store.ErrNotFound):
		return nil, ErrSocialUnknown
	case err != nil:
		return nil, err
	case !provider.Enabled:
		return nil, ErrSocialUnknown
	}

	return provider, nil
}

// challengeFor is the S256 code challenge of a verifier.
func challengeFor(verifier string) string {
	sum := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}

// socialToken is what a provider's token endpoint answered with. The fields
// that are read by name are here; the rest stays in Raw, because a provider
// may put the address there (VK) and nowhere else.
type socialToken struct {
	AccessToken string
	IDToken     string
	Raw         map[string]any
}

// exchangeSocialCode trades the code for a token, as RFC 6749 section 4.1.3.
func (s *Service) exchangeSocialCode(ctx context.Context, provider *model.SocialProvider, code, verifier string) (*socialToken, error) {
	_, tokenURL, _ := provider.Endpoints()

	secret, err := s.socialClientSecret(provider)
	if err != nil {
		return nil, err
	}

	form := url.Values{
		"grant_type":   {model.GrantAuthorizationCode},
		"code":         {code},
		"redirect_uri": {provider.CallbackURL(s.issuer)},
		"client_id":    {provider.ClientID},
	}
	if verifier != "" {
		form.Set("code_verifier", verifier)
	}

	// Only ever one way at a time: a server that is sent both may refuse them
	// both (RFC 6749 section 2.3).
	basic := provider.TokenAuthMethod() == model.SocialTokenAuthBasic
	if !basic {
		form.Set("client_secret", secret)
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, tokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Set("Accept", "application/json")

	if basic {
		request.SetBasicAuth(url.QueryEscape(provider.ClientID), url.QueryEscape(secret))
	}

	body, err := s.callProvider(request, provider, "token")
	if err != nil {
		return nil, err
	}

	var answer map[string]any
	if err := json.Unmarshal(body, &answer); err != nil {
		s.log.Error("a social provider's token answer was not JSON", "provider", provider.Slug, "error", err)
		return nil, ErrSocialUpstream
	}

	token := &socialToken{
		AccessToken: stringClaim(answer, "access_token"),
		IDToken:     stringClaim(answer, "id_token"),
		Raw:         answer,
	}

	if token.AccessToken == "" && token.IDToken == "" {
		s.log.Error("a social provider gave no token",
			"provider", provider.Slug, "error", stringClaim(answer, "error"))
		return nil, ErrSocialUpstream
	}

	return token, nil
}

// socialClientSecret is what proves this server to the provider: the secret
// they gave out, or — for Apple — a JWT signed with the key they registered.
func (s *Service) socialClientSecret(provider *model.SocialProvider) (string, error) {
	if !provider.Spec().SignedSecret {
		secret, err := s.sealer.OpenBytes(provider.ClientSecret)
		if err != nil {
			return "", fmt.Errorf("unseal the client secret of %s: %w", provider.Slug, err)
		}

		return string(secret), nil
	}

	return s.appleClientSecret(provider)
}

// appleClientSecret is the JWT Apple takes in place of a client secret: five
// minutes long, signed with the .p8 key registered to the team.
func (s *Service) appleClientSecret(provider *model.SocialProvider) (string, error) {
	pemBytes, err := s.sealer.OpenBytes(provider.PrivateKey)
	if err != nil {
		return "", fmt.Errorf("unseal the signing key of %s: %w", provider.Slug, err)
	}

	block, _ := pem.Decode(pemBytes)
	if block == nil {
		return "", fmt.Errorf("the signing key of %s is not PEM", provider.Slug)
	}

	parsed, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return "", fmt.Errorf("parse the signing key of %s: %w", provider.Slug, err)
	}

	signer, ok := parsed.(crypto.Signer)
	if !ok {
		return "", fmt.Errorf("the signing key of %s cannot sign", provider.Slug)
	}

	now := s.now()

	return jose.Sign(
		jose.Key{ID: provider.KeyID, Algorithm: jose.ES256, Private: signer},
		"JWT",
		map[string]any{
			"iss": provider.TeamID,
			"sub": provider.ClientID,
			"aud": "https://appleid.apple.com",
			"iat": now.Unix(),
			"exp": now.Add(5 * time.Minute).Unix(),
		},
	)
}

// socialIdentity is who the provider says signed in.
type socialIdentity struct {
	Subject       string
	Email         string
	EmailVerified bool
	FirstName     string
	LastName      string
}

// socialIdentity reads the identity out of what the provider answered with:
// the id_token that came back with the token, or the profile endpoint.
func (s *Service) socialIdentity(ctx context.Context, provider *model.SocialProvider, token *socialToken) (*socialIdentity, error) {
	spec := provider.Spec()
	_, _, userInfoURL := provider.Endpoints()

	// A provider with no profile endpoint says everything it is going to say
	// in the id_token. So does one that has an endpoint we were not given.
	var claims map[string]any
	if spec.IdentityInIDToken || userInfoURL == "" {
		payload, err := s.idTokenClaims(provider, token.IDToken)
		if err != nil {
			return nil, err
		}
		claims = payload
	} else {
		payload, err := s.fetchSocialProfile(ctx, provider, token)
		if err != nil {
			return nil, err
		}
		claims = payload
	}

	identity := &socialIdentity{
		Subject:   claimString(claims, spec.Claims.Subject),
		Email:     claimString(claims, spec.Claims.Email),
		FirstName: claimString(claims, spec.Claims.FirstName),
		LastName:  claimString(claims, spec.Claims.LastName),
	}

	// A provider that sends the address with the token rather than in the
	// profile (VK), and one that only ever hands out addresses it has checked
	// itself and says so nowhere.
	if identity.Email == "" && spec.EmailInTokenResponse {
		identity.Email = stringClaim(token.Raw, "email")
	}
	identity.EmailVerified = spec.AssumeEmailVerified || claimBool(claims, spec.Claims.EmailVerified)

	if identity.FirstName == "" && identity.LastName == "" {
		identity.FirstName, identity.LastName = splitName(claimString(claims, spec.Claims.FullName))
	}

	if identity.Subject == "" {
		s.log.Error("a social provider gave no subject", "provider", provider.Slug)
		return nil, ErrSocialUpstream
	}

	return identity, nil
}

// idTokenClaims reads an id_token that came straight back from the token
// endpoint.
//
// Its signature is not checked, and does not have to be: it arrived over TLS
// from the provider's own token endpoint, in answer to a request carrying
// this client's secret, which OpenID Connect Core section 3.1.3.7 accepts in
// place of checking the signature. What is checked is that the token is for
// this client and has not expired.
func (s *Service) idTokenClaims(provider *model.SocialProvider, idToken string) (map[string]any, error) {
	if idToken == "" {
		s.log.Error("a social provider gave no id_token", "provider", provider.Slug)
		return nil, ErrSocialUpstream
	}

	parts := strings.Split(idToken, ".")
	if len(parts) != 3 {
		return nil, ErrSocialUpstream
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, ErrSocialUpstream
	}

	var claims map[string]any
	if err := json.Unmarshal(payload, &claims); err != nil {
		return nil, ErrSocialUpstream
	}

	if !audienceHas(claims["aud"], provider.ClientID) {
		s.log.Error("a social provider's id_token was for another client", "provider", provider.Slug)
		return nil, ErrSocialUpstream
	}

	if exp, ok := numberClaim(claims["exp"]); ok && s.now().After(time.Unix(int64(exp), 0)) {
		return nil, ErrSocialExpired
	}

	return claims, nil
}

// fetchSocialProfile reads the profile endpoint with the access token.
func (s *Service) fetchSocialProfile(ctx context.Context, provider *model.SocialProvider, token *socialToken) (map[string]any, error) {
	spec := provider.Spec()
	_, _, userInfoURL := provider.Endpoints()

	values := url.Values{}
	for key, value := range spec.UserInfoParams {
		values.Set(key, value)
	}
	// An API that reads the token from the query rather than a header.
	if spec.UserInfoTokenParam != "" {
		values.Set(spec.UserInfoTokenParam, token.AccessToken)
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, withQuery(userInfoURL, values), nil)
	if err != nil {
		return nil, err
	}

	scheme := spec.UserInfoScheme
	if scheme == "" {
		scheme = "Bearer"
	}
	request.Header.Set("Authorization", scheme+" "+token.AccessToken)
	request.Header.Set("Accept", "application/json")

	body, err := s.callProvider(request, provider, "profile")
	if err != nil {
		return nil, err
	}

	var claims map[string]any
	if err := json.Unmarshal(body, &claims); err != nil {
		s.log.Error("a social provider's profile was not JSON", "provider", provider.Slug, "error", err)
		return nil, ErrSocialUpstream
	}

	return claims, nil
}

// callProvider makes one request to a provider and reads the answer. What
// went wrong is logged with the provider's own words; the caller gets one
// error, because none of it is the person signing in to act on.
func (s *Service) callProvider(request *http.Request, provider *model.SocialProvider, what string) ([]byte, error) {
	client := s.social
	if client == nil {
		client = &http.Client{Timeout: socialTimeout}
	}

	response, err := client.Do(request)
	if err != nil {
		s.log.Error("a social provider could not be reached",
			"provider", provider.Slug, "request", what, "error", err)
		return nil, ErrSocialUpstream
	}
	defer response.Body.Close()

	// Enough of the answer to work with, and no more: a provider answering
	// with a stream is not going to be read into memory.
	body, err := io.ReadAll(io.LimitReader(response.Body, 1<<20))
	if err != nil {
		return nil, ErrSocialUpstream
	}

	if response.StatusCode != http.StatusOK {
		s.log.Error("a social provider refused",
			"provider", provider.Slug, "request", what,
			"status", response.StatusCode, "body", truncate(string(body), 512))
		return nil, ErrSocialUpstream
	}

	return body, nil
}

// signInWithIdentity turns who the provider said it was into a session: the
// account that already holds this identity, the one with the same address, or
// a new one.
func (s *Service) signInWithIdentity(
	ctx context.Context,
	provider *model.SocialProvider,
	who *socialIdentity,
	request string,
	client Client,
) (*SignInResult, error) {
	now := s.now()
	email := strings.ToLower(strings.TrimSpace(who.Email))

	// Someone who has signed in with this provider before.
	identity, err := s.store.UserIdentity(ctx, provider.ID, who.Subject)
	switch {
	case err == nil:
		user, err := s.store.User(ctx, identity.UserID)
		if err != nil {
			return nil, err
		}
		if !user.CanSignIn(now) {
			s.record(ctx, user, user.Email, "user.login_blocked", client, map[string]any{
				"reason": "inactive", "provider": provider.Name,
			})
			return nil, ErrInvalidCredentials
		}

		if err := s.store.MarkIdentityUsed(ctx, identity, email, now); err != nil {
			return nil, err
		}

		return s.startSocialSession(ctx, user, provider, client)
	case !errors.Is(err, store.ErrNotFound):
		return nil, err
	}

	if email == "" {
		return nil, ErrSocialNoEmail
	}

	// An account with that address already: the provider has to have proved
	// the address belongs to whoever is signing in, and the provider has to
	// be one this installation trusts to prove it. The account has to have
	// proved it too — otherwise whoever registered the address before its
	// owner arrived would keep a password to the owner's account.
	user, err := s.store.UserByEmail(ctx, email)
	switch {
	case err == nil:
		if !who.EmailVerified || !provider.LinkVerifiedEmails || !user.EmailVerified {
			return nil, ErrSocialLinkRefused
		}
		if !user.CanSignIn(now) {
			return nil, ErrInvalidCredentials
		}

		if err := s.connectIdentity(ctx, user, provider, who, email, now); err != nil {
			return nil, err
		}
		s.record(ctx, user, user.Email, "user.identity_connected", client, map[string]any{
			"provider": provider.Name,
		})

		return s.startSocialSession(ctx, user, provider, client)
	case !errors.Is(err, store.ErrNotFound):
		return nil, err
	}

	// Nobody yet: make an account, if this provider and this application make
	// accounts at all.
	if !provider.AllowRegistration {
		return nil, ErrSocialRegistrationClosed
	}
	if err := s.socialRegistrationAllowed(ctx, request); err != nil {
		return nil, err
	}

	user = &model.User{
		Email:         email,
		EmailVerified: who.EmailVerified,
		FirstName:     truncate(who.FirstName, 100),
		LastName:      truncate(who.LastName, 100),
		IsActive:      true,
		Data:          map[string]any{},
	}

	roles, err := s.store.DefaultUserRoles(ctx)
	if err != nil {
		return nil, err
	}
	user.Roles = roles

	if err := s.store.CreateUser(ctx, user); errors.Is(err, store.ErrDuplicate) {
		// Somebody made the account between the two queries above.
		return nil, ErrSocialLinkRefused
	} else if err != nil {
		return nil, err
	}

	if err := s.connectIdentity(ctx, user, provider, who, email, now); err != nil {
		return nil, err
	}

	s.record(ctx, user, user.Email, "user.registered", client, map[string]any{"provider": provider.Name})

	return s.startSocialSession(ctx, user, provider, client)
}

// socialRegistrationAllowed refuses to make an account for a sign-in to an
// application that does not take registrations, so the rule is the same
// whether someone registers with a password or with a provider.
func (s *Service) socialRegistrationAllowed(ctx context.Context, request string) error {
	if request == "" {
		return nil
	}

	req, err := s.pending(ctx, request)
	if err != nil {
		// The handle expired while they were away; the sign-in itself is
		// still good, and they can be let in to their own account.
		return nil
	}

	if !req.Application.AllowRegistration || !req.Application.Enabled {
		return ErrSocialRegistrationClosed
	}

	return nil
}

func (s *Service) connectIdentity(
	ctx context.Context,
	user *model.User,
	provider *model.SocialProvider,
	who *socialIdentity,
	email string,
	now time.Time,
) error {
	identity := &model.UserIdentity{
		UserID:      user.ID,
		ProviderID:  provider.ID,
		Subject:     who.Subject,
		Email:       email,
		LastLoginAt: &now,
	}

	return s.store.CreateUserIdentity(ctx, identity)
}

func (s *Service) startSocialSession(
	ctx context.Context,
	user *model.User,
	provider *model.SocialProvider,
	client Client,
) (*SignInResult, error) {
	result, err := s.startSession(ctx, user, client, "user.login")
	if err != nil {
		return nil, err
	}

	s.log.Info("a user signed in with a social provider",
		"provider", provider.Slug, "user", user.ID)

	return result, nil
}

// claimString reads a string out of a payload by path: dots step into
// objects and numbers into arrays, so "response.0.id" is as easy to name as
// "sub". A number is read as a string, because a provider that gives a
// numeric id (VK, Facebook) still gives an identity.
func claimString(claims map[string]any, path string) string {
	switch value := claimAt(claims, path).(type) {
	case string:
		return strings.TrimSpace(value)
	case float64:
		return strconv.FormatFloat(value, 'f', -1, 64)
	case json.Number:
		return value.String()
	default:
		return ""
	}
}

// claimBool reads a flag that a provider may write as a bool or as a string.
func claimBool(claims map[string]any, path string) bool {
	switch value := claimAt(claims, path).(type) {
	case bool:
		return value
	case string:
		return value == "true"
	default:
		return false
	}
}

// claimAt walks a path into decoded JSON.
func claimAt(claims map[string]any, path string) any {
	if path == "" {
		return nil
	}

	var current any = claims
	for _, step := range strings.Split(path, ".") {
		switch node := current.(type) {
		case map[string]any:
			current = node[step]
		case []any:
			index, err := strconv.Atoi(step)
			if err != nil || index < 0 || index >= len(node) {
				return nil
			}
			current = node[index]
		default:
			return nil
		}
	}

	return current
}

// stringClaim is claimString for a value at the top of a payload.
func stringClaim(claims map[string]any, name string) string {
	return claimString(claims, name)
}

// numberClaim reads a JSON number, which may have been decoded either way.
func numberClaim(value any) (float64, bool) {
	switch number := value.(type) {
	case float64:
		return number, true
	case json.Number:
		parsed, err := number.Float64()
		return parsed, err == nil
	default:
		return 0, false
	}
}

// audienceHas reports whether an id_token's aud names this client. It is one
// string or a list of them (RFC 7519 section 4.1.3).
func audienceHas(aud any, clientID string) bool {
	switch value := aud.(type) {
	case string:
		return value == clientID
	case []any:
		for _, one := range value {
			if name, ok := one.(string); ok && name == clientID {
				return true
			}
		}
	}

	return false
}

// splitName makes a first and last name out of the one name a provider gives.
func splitName(full string) (first, last string) {
	parts := strings.Fields(full)
	switch len(parts) {
	case 0:
		return "", ""
	case 1:
		return parts[0], ""
	default:
		return parts[0], strings.Join(parts[1:], " ")
	}
}

// SocialLanding is where to send the browser once a provider has signed
// someone in: back to the application that asked, or to their own account.
//
// A sign-in handle that expired while the person was away at the provider is
// not an error to show them — they are signed in, and their account page is
// somewhere to be.
func (s *Service) SocialLanding(ctx context.Context, result *SocialResult) string {
	if result.Request != "" {
		location, err := s.Continue(ctx, result.Request, result.SignIn.Session)
		if err == nil {
			return location
		}

		s.log.Info("a sign-in request expired while the user was at a provider",
			"provider", result.Provider, "error", err)
	}

	return s.accountLanding(result.Next)
}

// accountLanding is a path on the account app, checked the way the app checks
// its own `next`: anything else would make this server a way to send someone
// wherever the link said.
func (s *Service) accountLanding(next string) string {
	if strings.HasPrefix(next, "/") && !strings.HasPrefix(next, "//") && !strings.HasPrefix(next, `/\`) {
		return s.accountURL + next
	}

	return s.accountURL + "/"
}

// SocialErrorPage is the sign-in app's error page, saying what went wrong in
// words the person can act on.
func (s *Service) SocialErrorPage(err error) string {
	reason := ErrSocialUpstream

	var known *Problem
	switch {
	case errors.Is(err, ErrInvalidCredentials):
		reason = ErrSocialBlocked
	case errors.As(err, &known) && strings.HasPrefix(known.Code, "social_"):
		reason = known
	}

	// `reason` is what the page says, in the reader's language: its
	// `error.<reason>`. The description is the same sentence in English, for
	// whoever reads the address rather than the page.
	description, ok := locales.Text(locales.ID, "error."+reason.Code)
	if !ok {
		description = reason.message
	}

	return withQuery(s.accountURL+PageError, url.Values{
		"error":             {SocialErrorCode},
		"reason":            {reason.Code},
		"error_description": {description},
	})
}

// SocialErrorCode is what the error page is told a failed social sign-in was,
// so it can show the sentence rather than an OAuth code meant for developers.
const SocialErrorCode = "social_sign_in_failed"

// IsSocialFailure reports whether an error is one of the sign-in failures
// that are the person's to act on, rather than the server's to fix. The HTTP
// layer logs everything else.
func IsSocialFailure(err error) bool {
	for _, known := range []error{
		ErrSocialUnknown,
		ErrSocialExpired,
		ErrSocialUpstream,
		ErrSocialNoEmail,
		ErrSocialLinkRefused,
		ErrSocialRegistrationClosed,
		ErrInvalidCredentials,
	} {
		if errors.Is(err, known) {
			return true
		}
	}

	return false
}
