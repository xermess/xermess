package oidc

import (
	"context"

	"errors"
	"net/url"
	"slices"

	"loginer/internal/jose"
	"loginer/internal/model"
	"loginer/internal/store"
)

// LogoutParams are the parameters of an RP-initiated logout (OpenID Connect
// RP-Initiated Logout 1.0).
type LogoutParams struct {
	IDTokenHint           string
	ClientID              string
	PostLogoutRedirectURI string
	State                 string
}

// Logout ends the user's session at this server and returns where to send the
// browser, and whether it ended anything.
//
// Where to send them is the application's registered post-logout redirect URI
// when it named one, otherwise the sign-in app's signed-out page.
//
// `signedOut` is false for a request that was refused before anything was
// ended, and for one whose `id_token_hint` names somebody other than whoever
// this browser is signed in as. The caller clears the session cookie only when
// it is true: this endpoint is a plain GET that anyone can put in a link, so
// a request that is not about this browser's session must leave it as it was —
// otherwise a link with a stranger's hint, or with no valid hint at all, would
// sign people out as they followed it.
//
// It does not revoke the refresh tokens applications hold: signing out of the
// provider is not signing out of every application, and each application
// ends its own session.
func (s *Service) Logout(ctx context.Context, p LogoutParams, sessionToken string, client Client) (location string, signedOut bool, err error) {
	clientID := p.ClientID

	// subject is who the application says it is ending the session for, from
	// the hint, or "" when it sent none.
	var subject string

	if p.IDTokenHint != "" {
		// The hint names who asked; it may well have expired, which is fine.
		// Its signature and issuer still have to be this server's.
		var claims struct {
			Issuer   string `json:"iss"`
			Subject  string `json:"sub"`
			Audience any    `json:"aud"`
		}
		if _, err := jose.Verify(p.IDTokenHint, s.keys.lookup(ctx), &claims); err != nil || claims.Issuer != s.issuer {
			return s.errorPage(ErrInvalidRequest, "id_token_hint is not an ID token from this server"), false, nil
		}

		audience, _ := claims.Audience.(string)
		if clientID != "" && clientID != audience {
			return s.errorPage(ErrInvalidRequest, "client_id does not match the id_token_hint"), false, nil
		}
		clientID = audience
		subject = claims.Subject
	}

	var app *model.Application
	if clientID != "" {
		found, err := s.store.ApplicationByClientID(ctx, clientID)
		switch {
		case errors.Is(err, store.ErrNotFound):
			return s.errorPage(ErrInvalidRequest, "no application has this client_id"), false, nil
		case err != nil:
			return "", false, err
		}
		app = found
	}

	// Only a URI the application registered is followed, so a logout link
	// cannot be used to send someone anywhere.
	if p.PostLogoutRedirectURI != "" {
		if app == nil {
			return s.errorPage(ErrInvalidRequest, "post_logout_redirect_uri needs id_token_hint or client_id"), false, nil
		}
		if !slices.Contains(app.PostLogoutRedirectURIs, p.PostLogoutRedirectURI) {
			return s.errorPage(ErrInvalidRequest, "post_logout_redirect_uri is not registered for this application"), false, nil
		}
	}

	signedOut, err = s.endSession(ctx, sessionToken, subject, client)
	if err != nil {
		return "", false, err
	}

	if p.PostLogoutRedirectURI != "" {
		return withQuery(p.PostLogoutRedirectURI, url.Values{"state": {p.State}}), signedOut, nil
	}

	values := url.Values{}
	if app != nil {
		values.Set("client_id", app.ClientID)
	}

	return withQuery(s.accountURL+PageLoggedOut, values), signedOut, nil
}

// endSession signs this browser out, unless a hint said the request was about
// somebody else's session. It reports whether the request was about this
// browser at all — which is not the same as having found a session to end: a
// cookie whose session has already expired is still this browser's, and the
// caller may clear it.
func (s *Service) endSession(ctx context.Context, sessionToken, subject string, client Client) (bool, error) {
	if subject != "" {
		session, err := s.SessionFor(ctx, sessionToken)
		if err != nil {
			return false, err
		}
		if session != nil && session.User.ID.String() != subject {
			s.log.Info("a logout named another user than the one signed in here",
				"hint", subject, "session", session.User.ID)
			return false, nil
		}
	}

	return true, s.SignOut(ctx, sessionToken, client)
}

// PublicApplication is what anyone may know about an application: what its
// sign-in pages show.
type PublicApplication struct {
	Name              string `json:"name"`
	LogoURI           string `json:"logo_uri"`
	ClientURI         string `json:"client_uri"`
	PolicyURI         string `json:"policy_uri"`
	TosURI            string `json:"tos_uri"`
	AllowRegistration bool   `json:"allow_registration"`
}

// Public returns what an application's sign-in pages show about it.
func Public(app *model.Application) PublicApplication {
	return PublicApplication{
		Name:              app.Name,
		LogoURI:           app.LogoURI,
		ClientURI:         app.ClientURI,
		PolicyURI:         app.PolicyURI,
		TosURI:            app.TosURI,
		AllowRegistration: app.AllowRegistration && app.Enabled,
	}
}

// ApplicationByClientID returns what the signed-out page shows about an
// application, found by its client id.
func (s *Service) ApplicationByClientID(ctx context.Context, clientID string) (*PublicApplication, error) {
	app, err := s.store.ApplicationByClientID(ctx, clientID)
	if err != nil {
		return nil, err
	}

	public := Public(app)
	return &public, nil
}
