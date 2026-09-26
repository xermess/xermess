package oidc

import (
	"context"
	"errors"
	"net/url"
	"slices"
	"strconv"
	"strings"
	"time"

	"loginer/internal/model"
	"loginer/internal/store"
)

// AuthorizeParams are the parameters of an authorization request.
type AuthorizeParams struct {
	ClientID            string
	RedirectURI         string
	ResponseType        string
	Scope               string
	State               string
	Nonce               string
	CodeChallenge       string
	CodeChallengeMethod string
	Audience            string
	Prompt              string
	MaxAge              string
	LoginHint           string
	ResponseMode        string
}

// Authorize handles an authorization request and returns where to send the
// browser: back to the application with a code or an error, to the sign-in
// page, or — when the request cannot be trusted with a redirect — to the
// sign-in app's error page.
//
// `session` is the signed-in user's session in this browser, or nil. With one,
// and unless the request asks to sign in again, the user is not asked for
// their password a second time.
func (s *Service) Authorize(ctx context.Context, p AuthorizeParams, session *Session) (string, error) {
	app, err := s.store.ApplicationByClientID(ctx, p.ClientID)
	switch {
	case p.ClientID == "":
		return s.errorPage(ErrInvalidRequest, "client_id is required"), nil
	case errors.Is(err, store.ErrNotFound):
		return s.errorPage(ErrInvalidRequest, "no application has this client_id"), nil
	case err != nil:
		return "", err
	}

	// Until the redirect URI is known to be the application's, no error may
	// be sent to it.
	if p.RedirectURI == "" {
		return s.errorPage(ErrInvalidRequest, "redirect_uri is required"), nil
	}
	if !slices.Contains(app.RedirectURIs, p.RedirectURI) {
		return s.errorPage(ErrInvalidRequest, "redirect_uri is not registered for this application; it has to match exactly"), nil
	}

	back := func(code, description string) (string, error) {
		return s.redirectError(p.RedirectURI, p.State, code, description), nil
	}

	if !app.Enabled {
		return back(ErrUnauthorizedClient, "the application is disabled")
	}
	if p.ResponseType != "code" {
		return back(ErrUnsupportedResponseType, "response_type must be code")
	}
	if p.ResponseMode != "" && p.ResponseMode != "query" {
		return back(ErrInvalidRequest, "response_mode must be query")
	}
	if !slices.Contains(app.GrantTypes, model.GrantAuthorizationCode) {
		return back(ErrUnauthorizedClient, "the application does not have the authorization_code grant")
	}

	switch {
	case p.CodeChallenge == "" && app.RequirePKCE:
		return back(ErrInvalidRequest, "code_challenge is required: this application must use PKCE")
	case p.CodeChallenge != "" && p.CodeChallengeMethod != model.PKCES256:
		return back(ErrInvalidRequest, "code_challenge_method must be S256")
	case p.CodeChallenge != "" && (len(p.CodeChallenge) != 43):
		return back(ErrInvalidRequest, "code_challenge must be the base64url SHA-256 of the code verifier")
	}

	if strings.TrimSpace(p.Scope) == "" {
		return back(ErrInvalidScope, "scope is required")
	}

	if p.Audience != "" {
		if _, err := s.store.AudienceFor(ctx, app.ID, p.Audience); errors.Is(err, store.ErrNotFound) {
			return back(ErrInvalidRequest, "audience: no API has this identifier")
		} else if err != nil {
			return "", err
		}
	}

	prompts := strings.Fields(p.Prompt)
	if slices.Contains(prompts, "none") && len(prompts) > 1 {
		return back(ErrInvalidRequest, "prompt=none cannot be combined with other values")
	}

	maxAge := -1
	if p.MaxAge != "" {
		n, err := strconv.Atoi(p.MaxAge)
		if err != nil || n < 0 {
			return back(ErrInvalidRequest, "max_age must be a number of seconds")
		}
		maxAge = n
	}

	now := s.now()

	// A session counts only when it is recent enough for max_age and the
	// application did not ask for the password again.
	if session != nil {
		tooOld := maxAge >= 0 && now.Sub(session.Record.AuthTime) > time.Duration(maxAge)*time.Second
		if tooOld || slices.Contains(prompts, "login") {
			session = nil
		}
	}

	if session == nil && slices.Contains(prompts, "none") {
		return back(ErrLoginRequired, "the user is not signed in")
	}

	handle, hash, err := model.NewSecret()
	if err != nil {
		return "", err
	}

	req := &model.AuthorizationRequest{
		HandleHash:          hash,
		ApplicationID:       app.ID,
		RedirectURI:         p.RedirectURI,
		Scope:               strings.Join(strings.Fields(p.Scope), " "),
		State:               p.State,
		Nonce:               p.Nonce,
		CodeChallenge:       p.CodeChallenge,
		CodeChallengeMethod: p.CodeChallengeMethod,
		Audience:            p.Audience,
		LoginHint:           p.LoginHint,
		ExpiresAt:           now.Add(model.AuthorizationRequestLifetime),
	}
	req.Application = app

	if err := s.store.CreateAuthorizationRequest(ctx, req); err != nil {
		return "", err
	}

	if session != nil {
		return s.finish(ctx, req, session.User, &session.Record)
	}

	return withQuery(s.accountURL+PageLogin, url.Values{"request": {handle}}), nil
}

// PendingRequest is what the sign-in page shows about a sign-in under way.
type PendingRequest struct {
	Application *model.Application
	LoginHint   string
	ExpiresAt   time.Time
}

// Pending returns a sign-in under way, by the handle the sign-in page was
// given. One that has expired, been used, or never existed is
// ErrRequestExpired.
func (s *Service) Pending(ctx context.Context, handle string) (*PendingRequest, error) {
	req, err := s.pending(ctx, handle)
	if err != nil {
		return nil, err
	}

	return &PendingRequest{Application: req.Application, LoginHint: req.LoginHint, ExpiresAt: req.ExpiresAt}, nil
}

func (s *Service) pending(ctx context.Context, handle string) (*model.AuthorizationRequest, error) {
	if handle == "" {
		return nil, ErrRequestExpired
	}

	req, err := s.store.AuthorizationRequestByHandle(ctx, model.HashSecret(handle))
	switch {
	case errors.Is(err, store.ErrNotFound):
		return nil, ErrRequestExpired
	case err != nil:
		return nil, err
	}

	if !req.Usable(s.now()) || req.Application == nil {
		return nil, ErrRequestExpired
	}

	return req, nil
}

// Continue finishes a sign-in under way for a user who has just signed in on
// the sign-in page, and returns where to send the browser.
func (s *Service) Continue(ctx context.Context, handle string, session *Session) (string, error) {
	req, err := s.pending(ctx, handle)
	if err != nil {
		return "", err
	}

	return s.finish(ctx, req, session.User, &session.Record)
}

// finish ends an authorization request for a signed-in user: a code back to
// the application, or an error when the user may not have tokens for it.
func (s *Service) finish(ctx context.Context, req *model.AuthorizationRequest, user *model.User, session *model.UserSession) (string, error) {
	app := req.Application

	// The same evaluation the token endpoint runs, now, so a user who may not
	// sign in to the application — inactive, or without a required role —
	// is turned away here rather than holding a code that will be refused.
	token, err := s.evaluate(ctx, app, user, strings.Fields(req.Scope), req.Audience)
	if err != nil {
		return "", err
	}
	if !token.Issued {
		return s.redirectError(req.RedirectURI, req.State, ErrAccessDenied, token.Reason), nil
	}

	now := s.now()
	if err := s.store.CompleteAuthorizationRequest(ctx, req, now); errors.Is(err, store.ErrAlreadyUsed) {
		return "", ErrRequestExpired
	} else if err != nil {
		return "", err
	}

	code, hash, err := model.NewSecret()
	if err != nil {
		return "", err
	}

	err = s.store.CreateAuthorizationCode(ctx, &model.AuthorizationCode{
		CodeHash:            hash,
		ApplicationID:       app.ID,
		UserID:              user.ID,
		SessionID:           &session.ID,
		RedirectURI:         req.RedirectURI,
		Scope:               req.Scope,
		Nonce:               req.Nonce,
		CodeChallenge:       req.CodeChallenge,
		CodeChallengeMethod: req.CodeChallengeMethod,
		Audience:            req.Audience,
		AuthTime:            session.AuthTime,
		ExpiresAt:           now.Add(model.AuthorizationCodeLifetime),
	})
	if err != nil {
		return "", err
	}

	// iss tells the client which server answered, so a mix-up attack across
	// two providers is caught (RFC 9207).
	return withQuery(req.RedirectURI, url.Values{"code": {code}, "state": {req.State}, "iss": {s.issuer}}), nil
}

func (s *Service) redirectError(redirectURI, state, code, description string) string {
	return withQuery(redirectURI, url.Values{
		"error":             {code},
		"error_description": {description},
		"state":             {state},
		"iss":               {s.issuer},
	})
}

// evaluate loads what model.EvaluateToken needs and runs it. `user` is nil for
// a client credentials token.
func (s *Service) evaluate(ctx context.Context, app *model.Application, user *model.User, scopes []string, audience string) (model.TokenPreview, error) {
	request := model.TokenRequest{
		Application: *app,
		User:        user,
		Requested:   scopes,
		Issuer:      s.issuer,
		Now:         s.now(),
	}

	if user != nil {
		roles, err := s.store.EffectiveRoles(ctx, user)
		if err != nil {
			return model.TokenPreview{}, err
		}
		request.Roles = roles
	}

	if audience != "" {
		access, err := s.store.AudienceFor(ctx, app.ID, audience)
		if errors.Is(err, store.ErrNotFound) {
			return model.TokenPreview{Reason: "no API has the identifier " + audience}, nil
		}
		if err != nil {
			return model.TokenPreview{}, err
		}

		request.API = access.API
		request.Authorized = access.Authorized
		request.Allowed = access.Allowed
	}

	return model.EvaluateToken(request), nil
}
