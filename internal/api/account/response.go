package account

import (
	"net/http"
	"time"

	"github.com/google/uuid"

	"xermess/internal/api/respond"
	"xermess/internal/model"
	"xermess/internal/oidc"
)

// The problems these endpoints answer with that are their own.
var (
	applicationNotFound = respond.Define(http.StatusNotFound, "application_not_found", respond.Public)
	passwordTooShort    = respond.Define(http.StatusBadRequest, "password_too_short", respond.Public)
	noSSOConnection     = respond.Define(http.StatusNotFound, "no_sso_connection", respond.Public)
)

// provided are the provider's problems (oidc.Problems) as answers, by code:
// every one the provider can refuse something with is a problem of the public
// API, so a new one needs no line here — only its sentence in the sign-in
// pages' catalog, which the tests ask for. What differs is the status.
var provided = func() map[string]respond.Problem {
	statuses := map[string]int{
		oidc.ErrInvalidCredentials.Code: http.StatusUnauthorized,
		oidc.ErrRequestExpired.Code:     http.StatusGone,
		oidc.ErrResetInvalid.Code:       http.StatusGone,
		oidc.ErrRegistrationClosed.Code: http.StatusForbidden,
		oidc.ErrEmailTaken.Code:         http.StatusConflict,
		oidc.ErrNotYours.Code:           http.StatusNotFound,
		oidc.ErrSSORequired.Code:        http.StatusForbidden,
		oidc.ErrSSOUnknown.Code:         http.StatusNotFound,
	}

	out := map[string]respond.Problem{}
	for _, refused := range oidc.Problems() {
		status, ok := statuses[refused.Code]
		if !ok {
			status = http.StatusBadRequest
		}

		out[refused.Code] = respond.Define(status, refused.Code, respond.Public)
	}

	return out
}()

// requestResponse is a sign-in under way, as the sign-in page shows it.
type requestResponse struct {
	Application oidc.PublicApplication `json:"application"`
	LoginHint   string                 `json:"login_hint"`
	ExpiresAt   time.Time              `json:"expires_at"`
}

func newRequestResponse(pending *oidc.PendingRequest) requestResponse {
	return requestResponse{
		Application: oidc.Public(pending.Application),
		LoginHint:   pending.LoginHint,
		ExpiresAt:   pending.ExpiresAt,
	}
}

// signedInResponse says what the page does next: follow RedirectTo back to
// the application, or — for a temporary password — send the user to choose a
// new one with ResetToken. With neither, the user is signed in and there is
// nowhere to go.
type signedInResponse struct {
	RedirectTo             string `json:"redirect_to,omitempty"`
	PasswordChangeRequired bool   `json:"password_change_required,omitempty"`
	ResetToken             string `json:"reset_token,omitempty"`
}

// userResponse is a user as they see themselves. It is built by hand, so a
// column added to users later is not published to the user by accident — the
// roles and the additional fields an organisation keeps are not theirs to read
// here.
type userResponse struct {
	ID            uuid.UUID  `json:"id"`
	Email         string     `json:"email"`
	EmailVerified bool       `json:"email_verified"`
	FirstName     string     `json:"first_name"`
	LastName      string     `json:"last_name"`
	CreatedAt     time.Time  `json:"created_at"`
	LastLoginAt   *time.Time `json:"last_login_at"`
}

func newUserResponse(user *model.User) userResponse {
	return userResponse{
		ID:            user.ID,
		Email:         user.Email,
		EmailVerified: user.EmailVerified,
		FirstName:     user.FirstName,
		LastName:      user.LastName,
		CreatedAt:     user.CreatedAt,
		LastLoginAt:   user.LastLoginAt,
	}
}
