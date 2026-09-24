package oidc

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"xermess/i18n"
	"xermess/internal/mail"
	"xermess/internal/model"
	"xermess/internal/store"
)

// The emailed code step (model.StepEmailCode): a flow that names it holds a
// sign-in back once the password has been accepted, emails a code to the
// address on the account, and makes the session only when that code is typed
// back.
//
// It is the address being proved, not the password being doubted, which is
// why the ways in that a provider has already proved the address for — a
// social sign-in, an organisation's identity provider — go straight through.
// Asking those for a code would prove nothing that was not proved already,
// and would mean sending somebody back to a page they had left.
//
// The handle is the secret here. A six-digit code is guessable in a way
// nothing else this server hands out is, so the code alone is no use: it has
// to be typed into the browser that asked for it, which is the only thing
// holding the handle.

// CodeChallenge is a sign-in waiting for its emailed code, as the page shows
// it: what to come back with, where the message went, and what it may do
// while it waits.
type CodeChallenge struct {
	// Handle names the sign-in. The page keeps it and sends it back.
	Handle string `json:"handle"`

	// Email is the address the message went to, with the middle of the local
	// part hidden: enough for somebody to recognise their own address, not
	// enough to learn one they did not already know.
	Email string `json:"email"`

	ExpiresAt time.Time `json:"expires_at"`

	// ResendAfter is how many seconds until another message may be asked
	// for, so the page can count down rather than offering a button that
	// refuses.
	ResendAfter int `json:"resend_after"`

	// AttemptsLeft is how many more codes may be typed before this sign-in
	// is spent and has to be started again.
	AttemptsLeft int `json:"attempts_left"`
}

// codeContext is what one request needs to work on a waiting sign-in: the row
// and the settings of the moment, which are read together every time.
type codeContext struct {
	code     *model.LoginCode
	settings *model.OTPSettings
}

// sendLoginCode holds a sign-in back for a code and emails one.
//
// The row replaces any the same user was already waiting on
// (store.CreateLoginCode), so starting a sign-in twice leaves one code alive
// — the one in the message that arrived last.
func (s *Service) sendLoginCode(
	ctx context.Context,
	user *model.User,
	request string,
	remember bool,
	client Client,
) (*SignInResult, error) {
	settings, err := s.store.OTPSettings(ctx)
	if err != nil {
		return nil, err
	}

	handle, handleHash, err := model.NewSecret()
	if err != nil {
		return nil, err
	}

	code, codeHash, err := model.NewOTP(settings.Length)
	if err != nil {
		return nil, err
	}

	now := s.now()
	waiting := model.LoginCode{
		HandleHash: handleHash,
		CodeHash:   codeHash,
		UserID:     user.ID,
		Request:    request,
		Remember:   remember,
		SentAt:     now,
		ExpiresAt:  now.Add(settings.Lifetime()),
	}
	if err := s.store.CreateLoginCode(ctx, &waiting); err != nil {
		return nil, err
	}

	s.record(ctx, user, user.Email, "user.login_code_sent", client, nil)

	s.mailCode(ctx, user, code, request, client, *settings)

	answer := challenge(handle, waiting, *settings, user.Email, now)

	return &SignInResult{Code: &answer}, nil
}

// SubmitLoginCode finishes a sign-in that was waiting for a code.
//
// A wrong code is counted before it is refused, and the count is the row's
// rather than the address's: whoever is guessing has to hold the handle, and
// the sign-in they hold it for is the one that runs out of guesses.
func (s *Service) SubmitLoginCode(ctx context.Context, handle, typed string, client Client) (*SignInResult, error) {
	waiting, err := s.waitingCode(ctx, handle)
	if err != nil {
		return nil, err
	}

	user := waiting.code.User
	if user == nil || !user.CanSignIn(s.now()) {
		return nil, ErrInvalidCredentials
	}

	if !waiting.code.MatchesOTP(typed) {
		attempts, err := s.store.RecordLoginCodeAttempt(ctx, waiting.code)
		if err != nil {
			return nil, err
		}

		s.record(ctx, user, user.Email, "user.login_code_failed", client, map[string]any{
			"attempts": attempts, "of": waiting.settings.MaxAttempts,
		})

		if attempts >= waiting.settings.MaxAttempts {
			return nil, ErrCodeAttemptsUsed
		}

		return nil, ErrCodeInvalid
	}

	// The code is right, so it is spent — whether or not what follows works,
	// and whoever else is typing it at the same moment.
	used, err := s.store.ConsumeLoginCode(ctx, waiting.code, s.now())
	if err != nil {
		return nil, err
	}
	if !used {
		return nil, ErrCodeExpired
	}

	flow, err := s.flowFor(ctx, waiting.code.Request)
	if err != nil {
		return nil, err
	}

	return s.startSession(ctx, user, flow, waiting.code.Request, waiting.code.Remember, client, "user.login")
}

// ResendLoginCode sends another code for a sign-in that is still waiting,
// without starting it again: the handle stays, and so do the guesses already
// spent.
func (s *Service) ResendLoginCode(ctx context.Context, handle string, client Client) (*CodeChallenge, error) {
	waiting, err := s.waitingCode(ctx, handle)
	if err != nil {
		return nil, err
	}

	user := waiting.code.User
	if user == nil || !user.CanSignIn(s.now()) {
		return nil, ErrInvalidCredentials
	}

	now := s.now()
	if wait := waiting.code.SentAt.Add(waiting.settings.Resend()); now.Before(wait) {
		return nil, ErrCodeTooSoon
	}

	code, codeHash, err := model.NewOTP(waiting.settings.Length)
	if err != nil {
		return nil, err
	}

	// The sign-in keeps its own life rather than gaining another: a code
	// asked for again and again would otherwise never expire.
	expires := now.Add(waiting.settings.Lifetime())
	if err := s.store.ResendLoginCode(ctx, waiting.code, codeHash, now, expires); err != nil {
		return nil, err
	}

	s.record(ctx, user, user.Email, "user.login_code_sent", client, map[string]any{"resent": true})

	s.mailCode(ctx, user, code, waiting.code.Request, client, *waiting.settings)

	answer := challenge(handle, *waiting.code, *waiting.settings, user.Email, now)

	return &answer, nil
}

// waitingCode reads the sign-in a handle names, with the settings it is held
// to. Anything that cannot be used — a handle nobody has, a code that has
// expired, one that has been guessed at too often — is the same answer: the
// page starts the sign-in again either way, and which of them it was is not
// the page's business.
func (s *Service) waitingCode(ctx context.Context, handle string) (*codeContext, error) {
	if handle == "" {
		return nil, ErrCodeExpired
	}

	settings, err := s.store.OTPSettings(ctx)
	if err != nil {
		return nil, err
	}

	code, err := s.store.LoginCodeByHash(ctx, model.HashSecret(handle))
	switch {
	case errors.Is(err, store.ErrNotFound):
		return nil, ErrCodeExpired
	case err != nil:
		return nil, err
	}

	if !code.Usable(s.now(), settings.MaxAttempts) {
		return nil, ErrCodeExpired
	}

	return &codeContext{code: code, settings: settings}, nil
}

// challenge is a waiting sign-in as the page is told about it. The handle is
// passed in rather than read off the row: the row holds its hash, and the
// handle itself exists only where it was made.
func challenge(handle string, code model.LoginCode, settings model.OTPSettings, email string, now time.Time) CodeChallenge {
	wait := int(code.SentAt.Add(settings.Resend()).Sub(now).Round(time.Second).Seconds())
	if wait < 0 {
		wait = 0
	}

	return CodeChallenge{
		Handle:       handle,
		Email:        maskEmail(email),
		ExpiresAt:    code.ExpiresAt,
		ResendAfter:  wait,
		AttemptsLeft: max(settings.MaxAttempts-code.Attempts, 0),
	}
}

// mailCode sends the message, in the reader's language, without holding the
// request that asked up: the answer is the same whether the mail server is
// quick or slow, and a page waiting on one is a page that looks broken.
func (s *Service) mailCode(
	ctx context.Context,
	user *model.User,
	code, request string,
	client Client,
	settings model.OTPSettings,
) {
	text := s.textIn(ctx, client.Language)

	name := text["email.reset.your_account"]
	if pending, err := s.pending(ctx, request); err == nil {
		name = pending.Application.Name
	}

	params := map[string]any{
		"app":     name,
		"email":   user.Email,
		"code":    code,
		"minutes": settings.LifetimeMinutes,
	}

	msg := mail.Message{
		To:      user.Email,
		Subject: i18n.Fill(text["email.code.subject"], params),
		Body:    i18n.Fill(text["email.code.body"], params),
	}

	go func() {
		// The request that asked may be long gone by the time this sends.
		sendCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 30*time.Second)
		defer cancel()

		if err := s.mail.Send(sendCtx, msg); err != nil {
			s.log.Error("sending a one-time code failed", "error", err)
		}
	}()
}

// maskEmail hides the middle of the local part: "alexander@example.com"
// becomes "al•••••••r@example.com". It is shown to whoever asked for the code
// so they can tell which of their addresses the message went to, and it says
// nothing to somebody who did not know the address already.
func maskEmail(email string) string {
	at := strings.LastIndex(email, "@")
	if at < 1 {
		return email
	}

	local, domain := email[:at], email[at:]
	if len(local) <= 2 {
		return strings.Repeat("•", len(local)) + domain
	}

	return fmt.Sprintf("%c%s%c%s", local[0], strings.Repeat("•", len(local)-2), local[len(local)-1], domain)
}
