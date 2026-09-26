package oidc

import (
	"context"
	"encoding/json"
	"errors"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"

	"loginer/internal/jose"
	"loginer/internal/model"
	"loginer/internal/store"
)

// idTokenAlgorithm signs every ID token. RS256 is the one algorithm every
// OpenID Connect client must accept; access tokens follow their API's choice.
const idTokenAlgorithm = jose.RS256

// ClientAuth is how a client identified itself at an endpoint: its id, its
// secret if it sent one, and which way it sent them.
type ClientAuth struct {
	ID     string
	Secret string
	Method model.AuthMethod
}

// authenticate finds the application a client claims to be and holds it to
// the method it was registered with: a confidential client proves itself with
// its secret, the way it said it would; a public client sends no secret at all.
func (s *Service) authenticate(ctx context.Context, auth ClientAuth) (*model.Application, error) {
	if auth.ID == "" {
		return nil, oauthError(ErrInvalidClient, "client authentication is required")
	}

	app, err := s.store.ApplicationByClientID(ctx, auth.ID)
	if errors.Is(err, store.ErrNotFound) {
		return nil, oauthError(ErrInvalidClient, "unknown client")
	}
	if err != nil {
		return nil, err
	}

	registered := app.TokenEndpointAuthMethod
	switch {
	case registered == model.AuthNone && auth.Method != model.AuthNone:
		return nil, oauthError(ErrInvalidClient, "this is a public client: send client_id alone, with no secret")
	case registered != model.AuthNone && auth.Method != registered:
		return nil, oauthError(ErrInvalidClient, "this client authenticates with "+string(registered))
	case registered != model.AuthNone && !app.CheckSecret(auth.Secret):
		return nil, oauthError(ErrInvalidClient, "client authentication failed")
	}

	if !app.Enabled {
		return nil, oauthError(ErrUnauthorizedClient, "the application is disabled")
	}

	return app, nil
}

// TokenParams are the parameters of a token request.
type TokenParams struct {
	Client       ClientAuth
	GrantType    string
	Code         string
	RedirectURI  string
	CodeVerifier string
	RefreshToken string
	Scope        string
	Audience     string
}

// TokenResponse is a successful token response (RFC 6749 section 5.1).
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
	RefreshToken string `json:"refresh_token,omitempty"`
	IDToken      string `json:"id_token,omitempty"`
	Scope        string `json:"scope"`
}

// Token handles a token request. An *Error is an OAuth error to answer with;
// anything else went wrong on the server.
func (s *Service) Token(ctx context.Context, p TokenParams) (*TokenResponse, error) {
	app, err := s.authenticate(ctx, p.Client)
	if err != nil {
		return nil, err
	}

	if !slices.Contains(model.GrantTypes, p.GrantType) {
		return nil, oauthError(ErrUnsupportedGrantType, "grant_type must be one of: "+strings.Join(model.GrantTypes, ", "))
	}
	if !slices.Contains(app.GrantTypes, p.GrantType) {
		return nil, oauthError(ErrUnauthorizedClient, "the application does not have the "+p.GrantType+" grant")
	}

	switch p.GrantType {
	case model.GrantAuthorizationCode:
		return s.exchangeCode(ctx, app, p)
	case model.GrantRefreshToken:
		return s.refresh(ctx, app, p)
	default:
		return s.clientCredentials(ctx, app, p)
	}
}

func (s *Service) exchangeCode(ctx context.Context, app *model.Application, p TokenParams) (*TokenResponse, error) {
	if p.Code == "" {
		return nil, oauthError(ErrInvalidRequest, "code is required")
	}

	now := s.now()
	code, err := s.store.ClaimAuthorizationCode(ctx, model.HashSecret(p.Code), app.ID, now)
	switch {
	case errors.Is(err, store.ErrNotFound):
		return nil, oauthError(ErrInvalidGrant, "the authorization code is not valid")
	case errors.Is(err, store.ErrAlreadyUsed):
		// The code was found but not claimed. Either it belongs to another
		// client — and it is left as it is, so the client it was issued to can
		// still spend it — or this client has already used it, and only this
		// client should ever have held it: whoever sent it again got it some
		// other way, so what the first exchange issued goes.
		if code.ApplicationID != app.ID {
			return nil, oauthError(ErrInvalidGrant, "the authorization code was issued to another client")
		}
		if err := s.store.RevokeRefreshTokensForCode(ctx, code.ID, now); err != nil {
			return nil, err
		}
		return nil, oauthError(ErrInvalidGrant, "the authorization code has already been used")
	case err != nil:
		return nil, err
	}

	// The code is this application's: ClaimAuthorizationCode would not have
	// claimed it otherwise.
	switch {
	case !now.Before(code.ExpiresAt):
		return nil, oauthError(ErrInvalidGrant, "the authorization code has expired")
	case p.RedirectURI != code.RedirectURI:
		return nil, oauthError(ErrInvalidGrant, "redirect_uri does not match the authorization request")
	case code.CodeChallenge != "" && !model.VerifyPKCE(code.CodeChallenge, p.CodeVerifier):
		return nil, oauthError(ErrInvalidGrant, "code_verifier does not match the code_challenge")
	case code.CodeChallenge == "" && p.CodeVerifier != "":
		return nil, oauthError(ErrInvalidGrant, "code_verifier was sent, but the authorization request had no code_challenge")
	}

	user, err := s.store.User(ctx, code.UserID)
	if errors.Is(err, store.ErrNotFound) {
		return nil, oauthError(ErrInvalidGrant, "the user no longer exists")
	}
	if err != nil {
		return nil, err
	}

	return s.issue(ctx, grant{
		app:      app,
		user:     user,
		scopes:   strings.Fields(code.Scope),
		audience: code.Audience,
		nonce:    code.Nonce,
		authTime: code.AuthTime,
		session:  code.SessionID,
		codeID:   &code.ID,
		refusal:  ErrInvalidGrant,
	})
}

func (s *Service) refresh(ctx context.Context, app *model.Application, p TokenParams) (*TokenResponse, error) {
	if p.RefreshToken == "" {
		return nil, oauthError(ErrInvalidRequest, "refresh_token is required")
	}

	now := s.now()
	old, err := s.store.RefreshTokenByHash(ctx, model.HashSecret(p.RefreshToken))
	switch {
	case errors.Is(err, store.ErrNotFound):
		return nil, oauthError(ErrInvalidGrant, "the refresh token is not valid")
	case err != nil:
		return nil, err
	}

	if old.ApplicationID != app.ID {
		return nil, oauthError(ErrInvalidGrant, "the refresh token was issued to another client")
	}

	if old.RevokedAt != nil {
		// A replaced token presented again: one of the two presenters is not
		// the client. The whole family goes, so the thief's copy dies too.
		if err := s.store.RevokeRefreshFamily(ctx, old.FamilyID, now); err != nil {
			return nil, err
		}
		return nil, oauthError(ErrInvalidGrant, "the refresh token has been revoked")
	}

	if !old.Usable(now) {
		return nil, oauthError(ErrInvalidGrant, "the refresh token has expired")
	}

	original := strings.Fields(old.Scope)
	scopes := original
	if p.Scope != "" {
		scopes = strings.Fields(p.Scope)
		for _, scope := range scopes {
			if !slices.Contains(original, scope) {
				return nil, oauthError(ErrInvalidScope, "a refresh can only narrow the original scope; "+scope+" was not granted")
			}
		}
	}

	user, err := s.store.User(ctx, old.UserID)
	if errors.Is(err, store.ErrNotFound) {
		return nil, oauthError(ErrInvalidGrant, "the user no longer exists")
	}
	if err != nil {
		return nil, err
	}

	// Roles, scopes and the user's state are evaluated again, so a role taken
	// away or a user deactivated since the sign-in takes effect at the next
	// refresh, not when the refresh token expires.
	return s.issue(ctx, grant{
		app:      app,
		user:     user,
		scopes:   scopes,
		audience: old.Audience,
		authTime: old.AuthTime,
		replaces: old,
		refusal:  ErrInvalidGrant,
	})
}

func (s *Service) clientCredentials(ctx context.Context, app *model.Application, p TokenParams) (*TokenResponse, error) {
	if app.TokenEndpointAuthMethod == model.AuthNone {
		return nil, oauthError(ErrUnauthorizedClient, "a public client cannot use client_credentials")
	}

	return s.issue(ctx, grant{
		app:      app,
		scopes:   strings.Fields(p.Scope),
		audience: p.Audience,
		refusal:  ErrUnauthorizedClient,
	})
}

// grant is everything issue needs about one token request.
type grant struct {
	app      *model.Application
	user     *model.User
	scopes   []string
	audience string
	nonce    string
	authTime time.Time
	session  *uuid.UUID
	codeID   *uuid.UUID
	// replaces is the refresh token being rotated, when refreshing.
	replaces *model.RefreshToken
	// refusal is the error code for a request EvaluateToken refuses.
	refusal string
}

// issue evaluates a grant and signs what it amounts to.
func (s *Service) issue(ctx context.Context, g grant) (*TokenResponse, error) {
	if g.audience != "" {
		if _, err := s.store.AudienceFor(ctx, g.app.ID, g.audience); errors.Is(err, store.ErrNotFound) {
			return nil, oauthError(ErrInvalidRequest, "audience: no API has this identifier")
		}
	}

	evaluated, err := s.evaluate(ctx, g.app, g.user, g.scopes, g.audience)
	if err != nil {
		return nil, err
	}
	if !evaluated.Issued {
		return nil, oauthError(g.refusal, evaluated.Reason)
	}

	jti, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}

	access := evaluated.AccessToken
	access["jti"] = jti.String()
	if g.user != nil {
		access["auth_time"] = g.authTime.Unix()
	}

	algorithm, _ := evaluated.AccessTokenHeader["alg"].(string)
	accessKey, ok := s.keys.signing(algorithm)
	if !ok {
		return nil, errors.New("oidc: no signing key for " + algorithm)
	}

	accessToken, err := jose.Sign(accessKey, "at+jwt", access)
	if err != nil {
		return nil, err
	}

	granted, _ := access["scope"].(string)
	iat, _ := access["iat"].(int64)
	exp, _ := access["exp"].(int64)

	response := &TokenResponse{
		AccessToken: accessToken,
		TokenType:   "Bearer",
		ExpiresIn:   exp - iat,
		Scope:       granted,
	}

	if evaluated.IDToken != nil {
		id := evaluated.IDToken
		id["auth_time"] = g.authTime.Unix()
		id["at_hash"] = jose.HalfHash(accessToken)
		id["azp"] = g.app.ClientID
		if g.nonce != "" {
			id["nonce"] = g.nonce
		}
		if g.session != nil {
			id["sid"] = g.session.String()
		}

		idKey, ok := s.keys.signing(idTokenAlgorithm)
		if !ok {
			return nil, errors.New("oidc: no signing key for " + idTokenAlgorithm)
		}

		if response.IDToken, err = jose.Sign(idKey, "JWT", id); err != nil {
			return nil, err
		}
	}

	// A refresh rotates whatever it was given, whether or not offline_access
	// survived the request's scope: the token presented has been spent, and
	// leaving it usable would let a stolen one be presented again and again —
	// each time for a fresh access token, and never once reaching the reuse
	// detection in refresh, which is what the family is for. A grant that is
	// not a refresh gets a refresh token only when offline_access was granted.
	if g.user != nil && (g.replaces != nil || slices.Contains(strings.Fields(granted), model.ScopeOfflineAccess)) {
		token, err := s.newRefreshToken(ctx, g, granted)
		if err != nil {
			return nil, err
		}
		response.RefreshToken = token
	}

	return response, nil
}

// newRefreshToken stores a refresh token for a grant: a new family after a
// code exchange, the next one of the family when refreshing.
func (s *Service) newRefreshToken(ctx context.Context, g grant, scope string) (string, error) {
	now := s.now()

	token, hash, err := model.NewSecret()
	if err != nil {
		return "", err
	}

	next := &model.RefreshToken{
		TokenHash:     hash,
		ApplicationID: g.app.ID,
		UserID:        g.user.ID,
		CodeID:        g.codeID,
		Scope:         scope,
		Audience:      g.audience,
		AuthTime:      g.authTime,
		ExpiresAt:     now.Add(time.Duration(g.app.RefreshTokenLifetime) * time.Second),
	}

	if g.replaces == nil {
		if next.FamilyID, err = uuid.NewV7(); err != nil {
			return "", err
		}
		return token, s.store.CreateRefreshToken(ctx, next)
	}

	next.FamilyID = g.replaces.FamilyID
	next.CodeID = g.replaces.CodeID
	next.ExpiresAt = g.replaces.ExpiresAt

	err = s.store.RotateRefreshToken(ctx, g.replaces, next, now)
	if errors.Is(err, store.ErrAlreadyUsed) {
		// Another request rotated it first: the same replay as above.
		if err := s.store.RevokeRefreshFamily(ctx, g.replaces.FamilyID, now); err != nil {
			return "", err
		}
		return "", oauthError(ErrInvalidGrant, "the refresh token has been revoked")
	}

	return token, err
}

// accessClaims are the claims of an access token this server signed that the
// endpoints accepting one read.
type accessClaims struct {
	Issuer   string      `json:"iss"`
	Subject  string      `json:"sub"`
	Audience []string    `json:"aud"`
	ClientID string      `json:"client_id"`
	Scope    string      `json:"scope"`
	IssuedAt json.Number `json:"iat"`
	Expiry   json.Number `json:"exp"`
}

// verifyAccessToken checks an access token this server issued: signed by one of
// its keys, an access token and not an ID token, from this issuer, unexpired.
func (s *Service) verifyAccessToken(ctx context.Context, token string) (*accessClaims, map[string]any, error) {
	var raw map[string]any
	header, err := jose.Verify(token, s.keys.lookup(ctx), &raw)
	if err != nil || header.Type != "at+jwt" {
		return nil, nil, jose.ErrInvalid
	}

	encoded, err := json.Marshal(raw)
	if err != nil {
		return nil, nil, err
	}

	var claims accessClaims
	if err := json.Unmarshal(encoded, &claims); err != nil {
		return nil, nil, jose.ErrInvalid
	}

	exp, err := claims.Expiry.Int64()
	if err != nil || claims.Issuer != s.issuer || s.now().Unix() >= exp {
		return nil, nil, jose.ErrInvalid
	}

	return &claims, raw, nil
}

// UserInfo returns the claims about the user an access token was issued for,
// as the application may see them now.
func (s *Service) UserInfo(ctx context.Context, accessToken string) (map[string]any, error) {
	if accessToken == "" {
		return nil, oauthError(ErrInvalidToken, "an access token is required")
	}

	claims, _, err := s.verifyAccessToken(ctx, accessToken)
	if err != nil {
		return nil, oauthError(ErrInvalidToken, "the access token is invalid or has expired")
	}

	scopes := strings.Fields(claims.Scope)
	if !slices.Contains(scopes, model.ScopeOpenID) {
		return nil, oauthError(ErrInsufficientScope, "the access token was not granted the openid scope")
	}

	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return nil, oauthError(ErrInvalidToken, "the access token is not for a user")
	}

	app, err := s.store.ApplicationByClientID(ctx, claims.ClientID)
	if errors.Is(err, store.ErrNotFound) {
		return nil, oauthError(ErrInvalidToken, "the application no longer exists")
	}
	if err != nil {
		return nil, err
	}

	user, err := s.store.User(ctx, userID)
	if errors.Is(err, store.ErrNotFound) {
		return nil, oauthError(ErrInvalidToken, "the user no longer exists")
	}
	if err != nil {
		return nil, err
	}

	// The ID token the application would get now, for the scopes it was
	// granted: the same claims, filtered the same way, from the current
	// record. API scopes are irrelevant here, so no audience is loaded.
	evaluated, err := s.evaluate(ctx, app, user, scopes, "")
	if err != nil {
		return nil, err
	}
	if !evaluated.Issued || evaluated.IDToken == nil {
		return nil, oauthError(ErrInvalidToken, evaluated.Reason)
	}

	info := evaluated.IDToken
	for _, claim := range []string{"iss", "aud", "iat", "exp"} {
		delete(info, claim)
	}

	return info, nil
}

// Revoke revokes a refresh token, and the rest of its family (RFC 7009). An
// access token is a signed JWT that APIs check on their own, so it cannot be
// revoked and simply expires; asking to revoke one, or a token that does not
// exist, is still a success, as the RFC requires.
func (s *Service) Revoke(ctx context.Context, client ClientAuth, token string) error {
	app, err := s.authenticate(ctx, client)
	if err != nil {
		return err
	}

	if token == "" {
		return oauthError(ErrInvalidRequest, "token is required")
	}

	refresh, err := s.store.RefreshTokenByHash(ctx, model.HashSecret(token))
	if errors.Is(err, store.ErrNotFound) {
		return nil
	}
	if err != nil {
		return err
	}

	// Another client's token is left alone, and the answer does not say it
	// exists.
	if refresh.ApplicationID != app.ID {
		return nil
	}

	return s.store.RevokeRefreshFamily(ctx, refresh.FamilyID, s.now())
}

// Introspect says whether a token is active and what it carries (RFC 7662).
// Only confidential clients may ask: introspection tells whoever asks about a
// user's tokens, so the caller has to prove who it is — and it is only ever
// told about its own. A token issued to another client reads as inactive,
// access token and refresh token alike.
func (s *Service) Introspect(ctx context.Context, client ClientAuth, token string) (map[string]any, error) {
	app, err := s.authenticate(ctx, client)
	if err != nil {
		return nil, err
	}
	if app.TokenEndpointAuthMethod == model.AuthNone {
		return nil, oauthError(ErrInvalidClient, "introspection needs a confidential client")
	}

	inactive := map[string]any{"active": false}
	if token == "" {
		return inactive, nil
	}

	if claims, raw, err := s.verifyAccessToken(ctx, token); err == nil {
		// Whose token it is, the same question the refresh token below is
		// held to. RFC 7662 section 2.1 leaves the server to decide that a
		// token is the caller's to ask about, and the answer here carries the
		// subject, the scope and the user's roles — which is not something
		// any client that happens to hold a secret may read out of another
		// client's token.
		if claims.ClientID != app.ClientID {
			return inactive, nil
		}

		out := map[string]any{"active": true, "token_type": "Bearer"}
		for key, value := range raw {
			out[key] = value
		}
		return out, nil
	}

	refresh, err := s.store.RefreshTokenByHash(ctx, model.HashSecret(token))
	if errors.Is(err, store.ErrNotFound) {
		return inactive, nil
	}
	if err != nil {
		return nil, err
	}

	// A refresh token is only ever described to the client that holds it.
	if refresh.ApplicationID != app.ID || !refresh.Usable(s.now()) {
		return inactive, nil
	}

	return map[string]any{
		"active":     true,
		"token_type": "refresh_token",
		"client_id":  app.ClientID,
		"sub":        refresh.UserID.String(),
		"scope":      refresh.Scope,
		"exp":        refresh.ExpiresAt.Unix(),
		"iat":        refresh.CreatedAt.Unix(),
		"iss":        s.issuer,
	}, nil
}
