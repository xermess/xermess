package oidc

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"golang.org/x/crypto/bcrypt"

	"xermess/internal/mail"
	"xermess/internal/model"
	"xermess/internal/store"
	"xermess/locales"
)

// MinPasswordLength is the shortest password a user may choose, the same as
// an administrator may give them.
const MinPasswordLength = 8

// Session is a signed-in user in one browser.
type Session struct {
	Record model.UserSession
	User   *model.User
}

// SignInResult is what signing in or registering produced.
type SignInResult struct {
	Session *Session
	// Token is the session cookie's value. It is the only copy.
	Token string
	// ResetToken is set, instead of a session, when the user signed in with a
	// temporary password an administrator chose: they have to replace it
	// before they are let in, and this is the reset link that lets them.
	ResetToken string
}

// dummyHash is compared against for an unknown address, so it takes as long to
// refuse as a wrong password.
var dummyHash = sync.OnceValue(func() []byte {
	hash, err := bcrypt.GenerateFromPassword([]byte("not a password anyone has"), bcrypt.DefaultCost)
	if err != nil {
		panic(fmt.Sprintf("oidc: make the dummy hash: %v", err))
	}
	return hash
})

// SessionFor returns the session a cookie carries, or nil when it carries none
// that still signs a user in.
func (s *Service) SessionFor(ctx context.Context, token string) (*Session, error) {
	if token == "" {
		return nil, nil
	}

	record, err := s.store.UserSessionByHash(ctx, model.HashSecret(token))
	switch {
	case errors.Is(err, store.ErrNotFound):
		return nil, nil
	case err != nil:
		return nil, err
	}

	now := s.now()
	if !record.Active(now) {
		return nil, nil
	}

	user, err := s.store.User(ctx, record.UserID)
	switch {
	case errors.Is(err, store.ErrNotFound):
		return nil, nil
	case err != nil:
		return nil, err
	}

	if !user.CanSignIn(now) {
		return nil, nil
	}

	return &Session{Record: *record, User: user}, nil
}

// SignIn checks a user's address and password and starts a session, for the
// sign-in under way that `request` names, if any.
func (s *Service) SignIn(ctx context.Context, email, password, request string, client Client) (*SignInResult, error) {
	now := s.now()

	// A domain that has to sign in through its identity provider has no
	// password to try: it is refused before one is looked at, so the answer
	// says nothing about the account.
	if err := s.ssoRequiredFor(ctx, email); err != nil {
		return nil, err
	}

	// Neither does a flow without a password step: it is refused before the
	// password is looked at, for the same reason.
	flow, err := s.flowFor(ctx, request)
	if err != nil {
		return nil, err
	}
	if !flow.Offers(model.StepPassword) {
		return nil, ErrPasswordNotOffered
	}

	user, err := s.store.UserByEmail(ctx, email)
	switch {
	case errors.Is(err, store.ErrNotFound):
		_ = bcrypt.CompareHashAndPassword(dummyHash(), []byte(password))
		s.record(ctx, nil, email, "user.login_failed", client, map[string]any{"reason": "unknown email"})
		return nil, ErrInvalidCredentials
	case err != nil:
		return nil, err
	}

	// A user an administrator made without a password has none to match.
	if user.PasswordHash == "" || bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)) != nil {
		if user.PasswordHash == "" {
			_ = bcrypt.CompareHashAndPassword(dummyHash(), []byte(password))
		}

		locked, err := s.store.RecordUserFailedLogin(ctx, user, now, maxFailedLogins, lockoutDuration)
		if err != nil {
			return nil, err
		}

		reason := "wrong password"
		if locked {
			reason = "wrong password; locked after too many attempts"
		}
		s.record(ctx, user, user.Email, "user.login_failed", client, map[string]any{"reason": reason})

		return nil, ErrInvalidCredentials
	}

	if !user.CanSignIn(now) {
		reason := "inactive"
		if user.IsActive {
			reason = "locked"
		}
		s.record(ctx, user, user.Email, "user.login_blocked", client, map[string]any{"reason": reason})

		return nil, ErrInvalidCredentials
	}

	if user.IsTemporaryPassword {
		token, err := s.newReset(ctx, user)
		if err != nil {
			return nil, err
		}

		return &SignInResult{ResetToken: token}, nil
	}

	return s.startSession(ctx, user, flow, request, client, "user.login")
}

// startSession signs a user in under a login flow's rules: an address it
// requires verified is sent a link to verify it instead, and the session lasts
// as long as the flow says. Every way in ends here, so the rules hold however
// somebody arrived.
func (s *Service) startSession(
	ctx context.Context,
	user *model.User,
	flow *model.LoginFlow,
	request string,
	client Client,
	action string,
) (*SignInResult, error) {
	now := s.now()

	if flow.RequireVerifiedEmail && !user.EmailVerified {
		if err := s.sendVerification(ctx, user, request, client); err != nil {
			return nil, err
		}
		s.record(ctx, user, user.Email, "user.login_blocked", client, map[string]any{"reason": "email not verified"})

		return nil, ErrEmailNotVerified
	}

	token, hash, err := model.NewSecret()
	if err != nil {
		return nil, err
	}

	record := model.UserSession{
		TokenHash: hash,
		UserID:    user.ID,
		AuthTime:  now,
		ExpiresAt: now.Add(time.Duration(flow.SessionLifetimeHours) * time.Hour),
		IP:        client.IP,
		UserAgent: truncate(client.UserAgent, 255),
	}
	if err := s.store.CreateUserSession(ctx, &record); err != nil {
		return nil, err
	}

	if err := s.store.MarkUserSignedIn(ctx, user, now); err != nil {
		return nil, err
	}

	s.record(ctx, user, user.Email, action, client, nil)

	return &SignInResult{Session: &Session{Record: record, User: user}, Token: token}, nil
}

// SignOut ends the session a cookie carries. Signing out twice is not an error.
func (s *Service) SignOut(ctx context.Context, token string, client Client) error {
	session, err := s.SessionFor(ctx, token)
	if err != nil || session == nil {
		return err
	}

	if err := s.store.RevokeUserSession(ctx, session.Record.ID, s.now()); err != nil {
		return err
	}

	s.record(ctx, session.User, session.User.Email, "user.logout", client, nil)

	return nil
}

// Registration is what someone creating an account types.
type Registration struct {
	// Request is the sign-in under way the account is made for. Registering
	// is offered per application, so there is no registering without one.
	Request   string
	Email     string
	Password  string
	FirstName string
	LastName  string
	// AcceptedTerms says the person agreed to the application's terms and
	// privacy policy. It is required when the application links either.
	AcceptedTerms bool
}

// Register makes an account for a sign-in under way, and signs it in.
func (s *Service) Register(ctx context.Context, r Registration, client Client) (*SignInResult, error) {
	// Accounts at a domain its identity provider owns are made by signing in
	// through it.
	if err := s.ssoRequiredFor(ctx, r.Email); err != nil {
		return nil, err
	}

	req, err := s.pending(ctx, r.Request)
	if err != nil {
		return nil, err
	}

	app := req.Application
	if !app.AllowRegistration || !app.Enabled {
		return nil, ErrRegistrationClosed
	}

	// The login flow has to allow it too. The application says whether it
	// wants new accounts; the flow says whether this installation takes them
	// at all, so a flow with registration turned off closes the door rather
	// than only hiding the link to it.
	// A flow without a password step makes accounts through the providers it
	// offers, not with a password typed here.
	flow, err := s.store.EffectiveLoginFlow(ctx, app)
	if err != nil {
		return nil, err
	}
	if !flow.AllowRegistration || !flow.Offers(model.StepPassword) {
		return nil, ErrRegistrationClosed
	}

	if (app.TosURI != "" || app.PolicyURI != "") && !r.AcceptedTerms {
		return nil, ErrTermsRequired
	}

	// Required additional fields cannot be filled in on a sign-in page, so
	// an installation that has them cannot take registrations.
	fields, err := s.store.UserFields(ctx)
	if err != nil {
		return nil, err
	}
	for _, field := range fields {
		if _, err := field.Normalise(nil); err != nil {
			return nil, ErrDetailsRequired
		}
	}

	user := &model.User{
		Email:     strings.ToLower(strings.TrimSpace(r.Email)),
		FirstName: strings.TrimSpace(r.FirstName),
		LastName:  strings.TrimSpace(r.LastName),
		IsActive:  true,
		Data:      map[string]any{},
	}

	if err := user.SetPassword(r.Password); errors.Is(err, model.ErrPasswordTooLong) {
		return nil, ErrPasswordTooLong
	} else if err != nil {
		return nil, err
	}

	roles, err := s.store.DefaultUserRoles(ctx)
	if err != nil {
		return nil, err
	}
	user.Roles = roles

	if err := s.store.CreateUser(ctx, user); errors.Is(err, store.ErrDuplicate) {
		return nil, ErrEmailTaken
	} else if err != nil {
		return nil, err
	}

	s.record(ctx, user, user.Email, "user.registered", client, map[string]any{"application": app.Name})

	return s.startSession(ctx, user, flow, r.Request, client, "user.login")
}

// ForgotPassword sends a reset link to the address, if it has an account that
// may sign in. It says nothing either way, so the page cannot be used to find
// out which addresses have accounts; the email is sent in the background for
// the same reason, since sending takes a noticeable time.
//
// `language` is the one the page was shown in, and the email is written in
// it: somebody who asked in Uzbek is answered in Uzbek.
func (s *Service) ForgotPassword(ctx context.Context, email, request, language string, client Client) error {
	// A domain its identity provider owns keeps its passwords there.
	if err := s.ssoRequiredFor(ctx, email); err != nil {
		return err
	}

	// A flow that does not offer password resets does not send one. It says
	// nothing about it either: the answer is the same whatever happened here,
	// which is what keeps this page from telling anybody which addresses have
	// accounts.
	var app *model.Application
	if pending, err := s.pending(ctx, request); err == nil {
		app = pending.Application
	}

	flow, err := s.store.EffectiveLoginFlow(ctx, app)
	if err != nil {
		return err
	}
	if !flow.AllowPasswordReset || !flow.Offers(model.StepPassword) {
		return nil
	}

	user, err := s.store.UserByEmail(ctx, strings.TrimSpace(email))
	if errors.Is(err, store.ErrNotFound) {
		return nil
	}
	if err != nil {
		return err
	}

	if !user.IsActive {
		return nil
	}

	token, err := s.newReset(ctx, user)
	if err != nil {
		return err
	}

	text := s.textIn(ctx, language)

	name := text["email.reset.your_account"]
	if pending, err := s.pending(ctx, request); err == nil {
		name = pending.Application.Name
	} else {
		request = ""
	}

	link := withQuery(s.accountURL+PageReset, url.Values{"token": {token}, "request": {request}})
	params := map[string]any{
		"app":     name,
		"email":   user.Email,
		"link":    link,
		"minutes": int(model.PasswordResetLifetime.Minutes()),
	}

	msg := mail.Message{
		To:      user.Email,
		Subject: locales.Fill(text["email.reset.subject"], params),
		Body:    locales.Fill(text["email.reset.body"], params),
	}

	s.record(ctx, user, user.Email, "user.password_reset_requested", client, nil)

	go func() {
		// The request that asked may be long gone by the time this sends.
		sendCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 30*time.Second)
		defer cancel()

		if err := s.mail.Send(sendCtx, msg); err != nil {
			s.log.Error("sending a password reset email failed", "error", err)
		}
	}()

	return nil
}

func (s *Service) newReset(ctx context.Context, user *model.User) (string, error) {
	token, hash, err := model.NewSecret()
	if err != nil {
		return "", err
	}

	err = s.store.CreatePasswordReset(ctx, &model.PasswordReset{
		TokenHash: hash,
		UserID:    user.ID,
		ExpiresAt: s.now().Add(model.PasswordResetLifetime),
	})

	return token, err
}

// CheckReset reports whether a reset link still works, so the page can say so
// before someone types a new password into it.
func (s *Service) CheckReset(ctx context.Context, token string) error {
	_, err := s.reset(ctx, token)
	return err
}

func (s *Service) reset(ctx context.Context, token string) (*model.PasswordReset, error) {
	if token == "" {
		return nil, ErrResetInvalid
	}

	reset, err := s.store.PasswordResetByHash(ctx, model.HashSecret(token))
	switch {
	case errors.Is(err, store.ErrNotFound):
		return nil, ErrResetInvalid
	case err != nil:
		return nil, err
	}

	if !reset.Usable(s.now()) {
		return nil, ErrResetInvalid
	}

	return reset, nil
}

// ResetPassword sets a new password through a reset link. Every session and
// refresh token the user had ends: whoever made the reset necessary is signed
// out with everyone else.
func (s *Service) ResetPassword(ctx context.Context, token, password string, client Client) error {
	reset, err := s.reset(ctx, token)
	if err != nil {
		return err
	}

	user, err := s.store.User(ctx, reset.UserID)
	if errors.Is(err, store.ErrNotFound) {
		return ErrResetInvalid
	}
	if err != nil {
		return err
	}

	if err := user.SetPassword(password); errors.Is(err, model.ErrPasswordTooLong) {
		return ErrPasswordTooLong
	} else if err != nil {
		return err
	}

	if err := s.store.ResetPassword(ctx, reset, user, s.now()); errors.Is(err, store.ErrAlreadyUsed) {
		return ErrResetInvalid
	} else if err != nil {
		return err
	}

	s.record(ctx, user, user.Email, "user.password_reset", client, nil)

	return nil
}

// sendVerification sends a link that proves the address is the user's, in the
// language the pages were shown in. It is sent while the person waits, unlike
// a reset: they are told it has gone, and a link that failed to go should say
// so rather than leave them watching an empty inbox.
func (s *Service) sendVerification(ctx context.Context, user *model.User, request string, client Client) error {
	token, hash, err := model.NewSecret()
	if err != nil {
		return err
	}

	err = s.store.CreateEmailVerification(ctx, &model.EmailVerification{
		TokenHash: hash,
		UserID:    user.ID,
		ExpiresAt: s.now().Add(model.EmailVerificationLifetime),
	})
	if err != nil {
		return err
	}

	text := s.textIn(ctx, client.Language)

	name := text["email.reset.your_account"]
	if pending, err := s.pending(ctx, request); err == nil {
		name = pending.Application.Name
	} else {
		request = ""
	}

	params := map[string]any{
		"app":   name,
		"email": user.Email,
		"link":  withQuery(s.accountURL+PageVerify, url.Values{"token": {token}, "request": {request}}),
		"hours": int(model.EmailVerificationLifetime.Hours()),
	}

	sendCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	err = s.mail.Send(sendCtx, mail.Message{
		To:      user.Email,
		Subject: locales.Fill(text["email.verify.subject"], params),
		Body:    locales.Fill(text["email.verify.body"], params),
	})
	if err != nil {
		return fmt.Errorf("send the verification email: %w", err)
	}

	s.record(ctx, user, user.Email, "user.email_verification_sent", client, nil)

	return nil
}

// VerifyEmail uses a verification link: the address is the user's, and every
// other link sent to it stops working.
func (s *Service) VerifyEmail(ctx context.Context, token string, client Client) error {
	if token == "" {
		return ErrVerificationInvalid
	}

	verification, err := s.store.EmailVerificationByHash(ctx, model.HashSecret(token))
	switch {
	case errors.Is(err, store.ErrNotFound):
		return ErrVerificationInvalid
	case err != nil:
		return err
	case !verification.Usable(s.now()):
		return ErrVerificationInvalid
	}

	if err := s.store.VerifyEmail(ctx, verification, s.now()); errors.Is(err, store.ErrAlreadyUsed) {
		return ErrVerificationInvalid
	} else if err != nil {
		return err
	}

	user, err := s.store.User(ctx, verification.UserID)
	if err != nil {
		return err
	}
	s.record(ctx, user, user.Email, "user.email_verified", client, nil)

	return nil
}

// truncate is a value cut to fit a column, in bytes, and left as text a
// database will take.
//
// Two things have to be true of what comes out. It has to be valid UTF-8:
// what goes through here is a name a provider gave and a User-Agent header,
// neither of which this server writes, and Postgres refuses a string with a
// byte sequence that is not UTF-8 — which would fail the sign-in itself,
// rather than the name it was carrying. And the cut has to land between
// runes: cutting a 100-byte limit through the middle of a two-byte letter
// leaves exactly such a sequence, which is how a long enough name in Cyrillic
// or Japanese would otherwise stop somebody signing in at all.
//
// The limit is in bytes while the column counts characters, so this is
// conservative rather than exact: it can shorten more than it has to, never
// less.
func truncate(value string, max int) string {
	value = strings.ToValidUTF8(value, "")
	if len(value) <= max {
		return value
	}

	// Back up off a continuation byte to the start of the rune it belongs to.
	for max > 0 && !utf8.RuneStart(value[max]) {
		max--
	}

	return value[:max]
}
