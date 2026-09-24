package oidc

import (
	"net/http"
)

// The error codes of RFC 6749 section 4.1.2.1 and 5.2, RFC 6750 section 3.1
// and OpenID Connect Core section 3.1.2.6 that this server answers with.
const (
	ErrInvalidRequest          = "invalid_request"
	ErrInvalidClient           = "invalid_client"
	ErrInvalidGrant            = "invalid_grant"
	ErrUnauthorizedClient      = "unauthorized_client"
	ErrUnsupportedGrantType    = "unsupported_grant_type"
	ErrUnsupportedResponseType = "unsupported_response_type"
	ErrInvalidScope            = "invalid_scope"
	ErrAccessDenied            = "access_denied"
	ErrServerError             = "server_error"
	ErrLoginRequired           = "login_required"
	ErrInvalidToken            = "invalid_token"
	ErrInsufficientScope       = "insufficient_scope"
)

// Error is an OAuth error: a code a client can act on, a sentence a developer
// can read, and the HTTP status the endpoints answer it with.
type Error struct {
	Code        string
	Description string
	Status      int
}

func (e *Error) Error() string {
	return e.Code + ": " + e.Description
}

func oauthError(code, description string) *Error {
	status := http.StatusBadRequest
	switch code {
	case ErrInvalidClient, ErrInvalidToken:
		status = http.StatusUnauthorized
	case ErrInsufficientScope:
		status = http.StatusForbidden
	}

	return &Error{Code: code, Description: description, Status: status}
}

// Problem is something a person did that this service refuses: a wrong
// password, a link that has expired. Its code is what the sign-in pages show
// the reader, in their language — the API answers with it, and the page looks
// up `error.<code>` — so the sentence is the pages' to say. The one here is
// for the log.
//
// Every problem is a value of its own, declared below, so a caller asks
// errors.Is(err, ErrEmailTaken) and the API layer finds the code with
// errors.As. Problems lists them all, which is how the tests hold the API's
// definitions and the catalogs to this list.
type Problem struct {
	Code    string
	message string
}

func (p *Problem) Error() string {
	return p.message
}

var problems []*Problem

func problem(code, message string) *Problem {
	p := &Problem{Code: code, message: message}
	problems = append(problems, p)

	return p
}

// Problems is every problem this service can refuse something with.
func Problems() []*Problem {
	return problems
}

// The errors of signing a user in. They are what the sign-in pages show, so
// they say what went wrong without saying which accounts exist.
var (
	// ErrInvalidCredentials covers an unknown address, a wrong password, and
	// an account that may not sign in.
	ErrInvalidCredentials = problem("invalid_credentials", "wrong email or password")
	// ErrRequestExpired is a sign-in handle that has expired, been used, or
	// never existed.
	ErrRequestExpired = problem("request_expired", "the sign-in handle has expired")
	// ErrRegistrationClosed is registering where the application does not
	// allow it.
	ErrRegistrationClosed = problem("registration_closed", "the application does not allow registration")
	// ErrEmailTaken is registering an address that already has an account.
	ErrEmailTaken = problem("email_taken", "an account with this email already exists")
	// ErrResetInvalid is a reset link that has expired or been used.
	ErrResetInvalid = problem("reset_invalid", "the reset link has expired or been used")
	// ErrTermsRequired is registering without accepting the application's
	// terms and privacy policy.
	ErrTermsRequired = problem("terms_required", "the terms were not accepted")
	// ErrDetailsRequired is registering where accounts need fields a sign-in
	// page cannot ask for.
	ErrDetailsRequired = problem("details_required", "accounts need fields the page cannot ask for")
	// ErrPasswordTooLong is a password longer than bcrypt reads.
	ErrPasswordTooLong = problem("password_too_long", "the password is too long")
	// ErrPasswordNotOffered is a password where the login flow does not take
	// one — a flow that signs people in through their other accounts alone.
	ErrPasswordNotOffered = problem("password_not_offered", "this sign-in does not take a password")
	// ErrEmailNotVerified is an account whose address is unproved, at a flow
	// that requires it proved. A link to prove it has just been sent.
	ErrEmailNotVerified = problem("email_not_verified", "the address is not verified; a link to verify it was sent")
	// ErrVerificationInvalid is a verification link that has expired or been
	// used.
	ErrVerificationInvalid = problem("verification_invalid", "the verification link has expired or been used")
	// ErrSignInClosed is a login flow that is not signing anybody in.
	ErrSignInClosed = problem("sign_in_closed", "sign-ins are closed for this application")
	// ErrEmailChangeNotOffered is changing the sign-in address where the
	// login flow does not allow it.
	ErrEmailChangeNotOffered = problem("email_change_not_offered", "this sign-in does not allow changing the address")

	// The errors of the emailed code step. A code that is simply wrong is
	// told apart from a sign-in that is over, because the page does
	// different things with them: the first is typed again, the second is
	// started again.

	// ErrCodeInvalid is a wrong code, with guesses left.
	ErrCodeInvalid = problem("code_invalid", "that code is not the one that was sent")
	// ErrCodeExpired is a sign-in that is no longer waiting for a code: it
	// has expired, been finished, or never existed.
	ErrCodeExpired = problem("code_expired", "the code has expired; sign in again")
	// ErrCodeAttemptsUsed is the last guess being used up. The sign-in is
	// over, and the page says so rather than offering another try.
	ErrCodeAttemptsUsed = problem("code_attempts_used", "too many wrong codes; sign in again")
	// ErrCodeTooSoon is asking for another message before the wait is over.
	ErrCodeTooSoon = problem("code_too_soon", "another code cannot be sent yet")

	// The errors of signing in with an account somewhere else. They are shown
	// on the sign-in pages, so each says what the person can do about it.

	// ErrSocialUnknown is a provider that is not configured, or not enabled.
	ErrSocialUnknown = problem("social_unknown", "the provider is not available")
	// ErrSocialExpired is a sign-in that took too long at the provider, came
	// back twice, or was not started here at all.
	ErrSocialExpired = problem("social_expired", "the social sign-in expired or was used")
	// ErrSocialUpstream is the provider refusing or failing to answer. What
	// it said is logged; the person is only told it did not work.
	ErrSocialUpstream = problem("social_upstream", "the provider did not complete the sign-in")
	// ErrSocialNoEmail is a provider that gave no address, leaving nothing to
	// make an account with.
	ErrSocialNoEmail = problem("social_no_email", "the provider shared no email address")
	// ErrSocialLinkRefused is an address that already has an account the
	// provider has not proved belongs to this person.
	ErrSocialLinkRefused = problem("social_link_refused", "the address has an account the provider does not prove")
	// ErrSocialRegistrationClosed is a provider, or an application, that does
	// not make accounts this way.
	ErrSocialRegistrationClosed = problem("social_registration_closed", "the provider cannot create accounts here")
	// ErrSocialBlocked is an account found through a provider that may not
	// sign in.
	ErrSocialBlocked = problem("social_blocked", "the account may not sign in")
	// ErrSocialNotOffered is a provider where the login flow does not offer
	// signing in with another account.
	ErrSocialNotOffered = problem("social_not_offered", "this sign-in does not offer other accounts")
)
