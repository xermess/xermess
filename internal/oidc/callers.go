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

// This server's admin and account APIs also accept its own access tokens: the
// admin API from a service (client credentials), the account API from an
// application acting for a user.

// ErrTokenRefused is a token that may not call this API: foreign, expired,
// wrong audience or kind, or from an application that lost access.
var ErrTokenRefused = errors.New("oidc: the access token is not accepted by this API")

// AdminCaller is software calling the admin API.
type AdminCaller struct {
	Application *model.Application
	// Permissions are the token's scopes the application is still allowed, so
	// revoking a scope applies on the next call.
	Permissions []string
}

// AdminCaller checks a token presented to the admin API. It must be for the
// admin API and a service's token; a user's token is refused whatever its
// scopes.
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

// AccountCaller checks a token for the account API: issued for a user who still
// exists and may sign in.
func (s *Service) AccountCaller(ctx context.Context, token string) (*AccountCaller, error) {
	claims, app, err := s.caller(ctx, token, brand.AccountAPIIdentifier)
	if err != nil {
		return nil, err
	}

	// Access and scopes are rechecked on every call, so revoking them applies
	// at once.
	audience, err := s.store.AudienceFor(ctx, app.ID, brand.AccountAPIIdentifier)
	if err != nil {
		return nil, err
	}
	if !audience.Authorized {
		return nil, ErrTokenRefused
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
	// Active, and not unlocked: a token was issued to a signed-in user, and
	// the lock after wrong passwords is on signing in (see SessionFor).
	if !user.IsActive {
		return nil, ErrTokenRefused
	}

	scopes := []string{}
	for _, scope := range strings.Fields(claims.Scope) {
		if (scope == model.ScopeAccountRead || scope == model.ScopeAccountWrite) && slices.Contains(audience.Allowed, scope) {
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
