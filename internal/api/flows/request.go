package flows

import "loginer/internal/model"

// targetType is what a flow is called in the activity log.
const targetType = "login_flow"

// flowRequest is what a create or an update sends.
//
// Every field is a pointer because an update is a PATCH: nil is "not sent",
// and the stored value stands. A create reads the same struct — what it
// leaves out is taken from the flow a new installation starts with, so the
// panel can save a flow that is a name and nothing else.
type flowRequest struct {
	Name        *string `json:"name"`
	Slug        *string `json:"slug"`
	Description *string `json:"description"`

	IsDefault *bool `json:"is_default"`
	Enabled   *bool `json:"enabled"`

	Steps *[]model.LoginStep `json:"steps"`

	AllowSignIn           *bool `json:"allow_sign_in"`
	AllowRegistration     *bool `json:"allow_registration"`
	AllowPasswordReset    *bool `json:"allow_password_reset"`
	AllowRememberMe       *bool `json:"allow_remember_me"`
	VerifyEmailOnRegister *bool `json:"verify_email_on_register"`
	RequireVerifiedEmail  *bool `json:"require_verified_email"`
	AllowEmailChange      *bool `json:"allow_email_change"`
	SessionLifetimeHours  *int  `json:"session_lifetime_hours"`
}
