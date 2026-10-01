package oidc

import (
	"context"
	"errors"
	"net/url"
	"slices"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

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

// The most of each parameter an authorization request may carry. OAuth puts
// no size on any of them; these are generous for what each is for — a state
// is a nonce or a serialised return address, a scope a list of words — and
// small beside the megabyte a request body may be.
const (
	maxState     = 1024
	maxNonce     = 256
	maxScope     = 1024
	maxLoginHint = 255
	maxMethod    = 8
	maxAudience  = 255
)

// oversized says what is wrong with the size or encoding of the parameters
// that are stored, or "" when nothing is.
func (p AuthorizeParams) oversized() string {
	limits := []struct {
		name  string
		value string
		max   int
	}{
		{"state", p.State, maxState},
		{"nonce", p.Nonce, maxNonce},
		{"scope", p.Scope, maxScope},
		{"login_hint", p.LoginHint, maxLoginHint},
		{"code_challenge_method", p.CodeChallengeMethod, maxMethod},
		{"audience", p.Audience, maxAudience},
	}

	for _, limit := range limits {
		switch {
		case len(limit.value) > limit.max:
			return limit.name + " is longer than " + strconv.Itoa(limit.max) + " bytes"
		case !utf8.ValidString(limit.value) || strings.ContainsRune(limit.value, 0):
			return limit.name + " is not valid UTF-8"
		}
	}

	return ""
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

	// What is stored has a size. The columns do not — and a megabyte of
	// state per anonymous request is a way to fill the disk — so the sizes
	// are these, and what Postgres would refuse (a byte that is not UTF-8,
	// a NUL) is refused here as a bad request rather than a server error.
	// The error page rather than a redirect: an oversized state is not
	// echoed back to anyone.
	if reason := p.oversized(); reason != "" {
		return s.errorPage(ErrInvalidRequest, reason), nil
	}

	back := func(code, description string) (string, error) {
		return s.redirectError(p.RedirectURI, p.State, code, description), nil
	}

	if !app.IsEnabled {
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
		tooOld := maxAge >= 0 && now.Sub(session.Record.AuthenticatedAt) > time.Duration(maxAge)*time.Second
		if tooOld || slices.Contains(prompts, "login") {
			session = nil
		}
	}

	// A session also has to satisfy this application's flow. It was made by
	// whichever flow the person signed in through — the default one, when
	// they signed in with no application waiting — and that flow may be
	// looser than this one: the rules a flow puts on signing in are applied
	// where the session is made (startSession), so a stricter application
	// would otherwise be entered on a session that never met them. One that
	// does not meet them is sent to the sign-in page, which runs this flow.
	if session != nil {
		flow, err := s.store.EffectiveLoginFlow(ctx, app)
		if err != nil {
			return "", err
		}
		if !sessionSatisfies(flow, session, now) {
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

// sessionSatisfies reports whether a session, made by whichever flow signed
// the person in, would have been made by this flow: signing in is open, the
// address is verified where the flow insists on it, and the session is no
// older than the flow lets a session live.
//
// A session does not record which flow made it or which steps it passed, so
// a flow that holds sign-ins for an emailed code is taken not to have been
// met: the person signs in again, code and all, each time an application on
// such a flow sends them here. That gives up single sign-on between those
// applications for the certainty that the code was typed; recording the
// steps a session passed is how to have both.
func sessionSatisfies(flow *model.LoginFlow, session *Session, now time.Time) bool {
	switch {
	case !flow.AllowSignIn:
		return false
	case flow.RequireVerifiedEmail && !session.User.IsEmailVerified:
		return false
	case now.Sub(session.Record.AuthenticatedAt) > time.Duration(flow.SessionLifetimeHours)*time.Hour:
		return false
	case flow.Offers(model.StepEmailCode):
		return false
	}

	return true
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

	// The application's redirect URIs and PKCE rule are re-checked here, not
	// only when the request was made: an administrator may have removed the
	// redirect URI, or turned PKCE on, in between. A redirect URI that is no
	// longer registered cannot be redirected to at all, so the sign-in app's
	// error page says so; PKCE missing is sent back as an error, since that
	// redirect URI is still good.
	if !slices.Contains(app.RedirectURIs, req.RedirectURI) {
		return s.errorPage(ErrInvalidRequest, "redirect_uri is no longer registered for this application"), nil
	}
	if app.RequirePKCE && req.CodeChallenge == "" {
		return s.redirectError(req.RedirectURI, req.State, ErrInvalidRequest, "this application now requires PKCE; start the sign-in again"), nil
	}

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
		AuthenticatedAt:     session.AuthenticatedAt,
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
