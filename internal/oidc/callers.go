package oidc

import (
	"context"
	"errors"
	"slices"
	"strings"

	"github.com/google/uuid"

	"loginer/internal/brand"
	"loginer/internal/model"
	"loginer/internal/store"
)

// This server's own two APIs can be called with an access token it issued,
// as well as with a browser's session: the admin API by software acting as
// itself (admin-cli, or any machine-to-machine application an administrator
// authorizes), and the account API by an application acting for a signed-in
// user. What a token may do there is decided here, like everything else a
// token means.

// ErrTokenRefused is an access token that may not call the API it was
// presented to: not this server's, expired, for another audience, of the
// wrong kind, or from an application that has since lost access.
var ErrTokenRefused = errors.New("oidc: the access token is not accepted by this API")

// AdminCaller is software calling the admin API.
type AdminCaller struct {
	Application *model.Application
	// Permissions are the admin permissions the call may use: the token's
	// scopes that the application is still allowed, so taking a scope away
	// in the panel takes effect on the next call rather than at expiry.
	Permissions []string
}

// AdminCaller checks an access token presented to the admin API. It has to be
// for the admin API, and a service's token — from the client credentials
// grant, whose subject is the application itself. A token issued for a user
// is refused whatever its scopes: the admin API acts for administrators, and
// a person signing in to an application is not one.
func (s *Service) AdminCaller(ctx context.Context, token string) (*AdminCaller, error) {
	claims, app, err := s.caller(ctx, token, brand.AdminAPIIdentifier)
	if err != nil {
		return nil, err
	}
	if claims.Subject != claims.ClientID {
		return nil, ErrTokenRefused
	}

	audience, err := s.store.AudienceFor(ctx, app.ID, brand.AdminAPIIdentifier)
	if err != nil {
		return nil, err
	}
	if !audience.Authorized {
		return nil, ErrTokenRefused
	}

	catalog := model.AdminPermissionNames()
	permissions := []string{}
	for _, scope := range strings.Fields(claims.Scope) {
		if slices.Contains(audience.Allowed, scope) && slices.Contains(catalog, scope) {
			permissions = append(permissions, scope)
		}
	}

	return &AdminCaller{Application: app, Permissions: permissions}, nil
}

// AccountCaller is an application calling the account API for a user.
type AccountCaller struct {
	// Session is the user, with no browser session behind it: the account
	// functions treat none of the user's sessions as the current one.
	Session  *Session
	ClientID string
	// Scopes are the account API scopes the token carries.
	Scopes []string
}

// AccountCaller checks an access token presented to the account API. It has
// to be for the account API and issued for a user, who still exists and may
// still sign in.
func (s *Service) AccountCaller(ctx context.Context, token string) (*AccountCaller, error) {
	claims, _, err := s.caller(ctx, token, brand.AccountAPIIdentifier)
	if err != nil {
		return nil, err
	}

	id, err := uuid.Parse(claims.Subject)
	if err != nil || claims.Subject == claims.ClientID {
		return nil, ErrTokenRefused
	}

	user, err := s.store.User(ctx, id)
	if errors.Is(err, store.ErrNotFound) {
		return nil, ErrTokenRefused
	}
	if err != nil {
		return nil, err
	}
	if !user.CanSignIn(s.now()) {
		return nil, ErrTokenRefused
	}

	scopes := []string{}
	for _, scope := range strings.Fields(claims.Scope) {
		if scope == model.ScopeAccountRead || scope == model.ScopeAccountWrite {
			scopes = append(scopes, scope)
		}
	}

	return &AccountCaller{Session: &Session{User: user}, ClientID: claims.ClientID, Scopes: scopes}, nil
}

// caller checks what both APIs need of a token: this server's, unexpired,
// for this audience, and from an application that still exists and is on.
func (s *Service) caller(ctx context.Context, token, audience string) (*accessClaims, *model.Application, error) {
	claims, _, err := s.verifyAccessToken(ctx, token)
	if err != nil || !slices.Contains(claims.Audience, audience) {
		return nil, nil, ErrTokenRefused
	}

	app, err := s.store.ApplicationByClientID(ctx, claims.ClientID)
	if errors.Is(err, store.ErrNotFound) {
		return nil, nil, ErrTokenRefused
	}
	if err != nil {
		return nil, nil, err
	}
	if !app.IsEnabled {
		return nil, nil, ErrTokenRefused
	}

	return claims, app, nil
}
