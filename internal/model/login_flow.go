package model

import (
	"fmt"
	"strings"
)

// LoginFlow is the steps someone is taken through when signing in, and what
// they may do along the way. One is the default; an application may name
// another through LoginFlowID.
//
// Every way in follows the flow of its application. Steps not implemented yet
// can be planned, but a flow must include one that is, so nothing saved can
// lock everybody out.
type LoginFlow struct {
	Base

	// Name is what the flow is called in the panel.
	Name string `gorm:"size:100;not null" json:"name"`

	// Slug is the short name an application or an export refers to it by.
	Slug string `gorm:"size:64;not null;uniqueIndex" json:"slug"`

	// Description is what this flow is for, in the administrator's words.
	Description string `gorm:"size:500" json:"description"`

	// IsDefault marks the flow for applications that name none; exactly one has
	// it, and it cannot be disabled or removed.
	IsDefault bool `gorm:"not null" json:"is_default"`

	// IsEnabled allows applications to use this flow; applications on a
	// disabled flow fall back to the default.
	IsEnabled bool `gorm:"not null" json:"is_enabled"`

	// Steps are the steps in the order they are taken.
	Steps StepList `gorm:"type:text;serializer:json;not null" json:"steps"`

	// AllowSignIn closes every way in through this flow, registration included.
	// Unlike Enabled, it does not fall back to the default flow.
	AllowSignIn bool `gorm:"not null" json:"allow_sign_in"`

	// AllowRegistration offers "create an account"; off, only administrators
	// create accounts.
	AllowRegistration bool `gorm:"not null" json:"allow_registration"`

	// AllowPasswordReset offers the "forgotten your password" link.
	AllowPasswordReset bool `gorm:"not null" json:"allow_password_reset"`

	// AllowRememberMe offers "stay signed in". Unticked, the cookie ends with
	// the browser window.
	AllowRememberMe bool `gorm:"not null" json:"allow_remember_me"`

	// VerifyEmailOnRegister sends a new account a confirmation link without
	// holding it back; RequireVerifiedEmail decides whether an unconfirmed
	// address may sign in.
	VerifyEmailOnRegister bool `gorm:"not null" json:"verify_email_on_register"`

	// RequireVerifiedEmail refuses a sign-in until the address has been
	// confirmed, rather than letting an unconfirmed account in and nagging.
	RequireVerifiedEmail bool `gorm:"not null" json:"require_verified_email"`

	// AllowEmailChange lets users change their address from the account page,
	// confirmed by a link to the new address.
	AllowEmailChange bool `gorm:"not null" json:"allow_email_change"`

	// SessionLifetimeHours is how long a session made by this flow lasts.
	SessionLifetimeHours int `gorm:"not null" json:"session_lifetime_hours"`
}

// TableName pins the table name.
func (LoginFlow) TableName() string {
	return "login_flows"
}

// StepList is a flow's steps, stored as JSON. It is its own type so an empty
// flow is stored as [] rather than null.
type StepList []LoginStep

// LoginStep is one thing a person does on their way in.
type LoginStep string

// The steps a flow can be made of.
const (
	// StepIdentifier asks who is signing in. Every flow starts with it.
	StepIdentifier LoginStep = "identifier"

	// StepPassword asks for the account's password.
	StepPassword LoginStep = "password"

	// StepSocial offers the buttons for the providers registered on the
	// Social page, beside the first step rather than after it.
	StepSocial LoginStep = "social"

	// StepEmailCode sends a one-time code to the address and asks for it.
	StepEmailCode LoginStep = "email_code"

	// StepTOTP asks for a code from an authenticator app.
	StepTOTP LoginStep = "totp"

	// StepTerms asks the person to accept the organisation's agreements
	// before their account is made.
	StepTerms LoginStep = "terms"

	// StepConsent shows what the application is asking for and waits to be
	// allowed.
	StepConsent LoginStep = "consent"
)

// LoginStepSpec describes one step for the panel: what it is called, what it
// does, and whether this server runs it yet.
type LoginStepSpec struct {
	Step        LoginStep `json:"step"`
	Label       string    `json:"label"`
	Description string    `json:"description"`

	// Fixed steps cannot be removed or moved: a flow that does not ask who
	// is signing in has nobody to sign in.
	Fixed bool `json:"fixed"`

	// Implemented says the sign-in pages run this step today; others can be
	// planned and are marked in the panel.
	Implemented bool `json:"implemented"`
}

// LoginStepSpecs is the catalog, in the order the panel offers them.
var LoginStepSpecs = []LoginStepSpec{
	{
		Step:        StepIdentifier,
		Label:       "Identify",
		Description: "Ask for the email address the account is under.",
		Fixed:       true,
		Implemented: true,
	},
	{
		Step:        StepPassword,
		Label:       "Password",
		Description: "Ask for the account's password.",
		Implemented: true,
	},
	{
		Step:        StepSocial,
		Label:       "Another account",
		Description: "Offer the providers on the Social page beside the first step.",
		Implemented: true,
	},
	{
		Step:        StepEmailCode,
		Label:       "Emailed code",
		Description: "Send a one-time code to the address and ask for it.",
		Implemented: true,
	},
	{
		Step:        StepTOTP,
		Label:       "Authenticator code",
		Description: "Ask for a code from an authenticator app.",
	},
	{
		Step:        StepTerms,
		Label:       "Agreements",
		Description: "Ask a new account to accept the organisation's terms and privacy policy.",
	},
	{
		Step:        StepConsent,
		Label:       "Consent",
		Description: "Show what the application is asking for and wait to be allowed.",
	},
}

// LoginStepSpecOf returns what is known about a step.
func LoginStepSpecOf(step LoginStep) (LoginStepSpec, bool) {
	for _, spec := range LoginStepSpecs {
		if spec.Step == step {
			return spec, true
		}
	}

	return LoginStepSpec{}, false
}

// proofSteps are the implemented steps that prove identity. A flow needs one,
// or nobody could sign in.
var proofSteps = []LoginStep{StepPassword, StepSocial}

// maxSessionLifetimeHours is ninety days: long enough for "stay signed in",
// short enough that a session left behind on a shared machine expires.
const maxSessionLifetimeHours = 24 * 90

// DefaultLoginFlow is an address and password, the provider buttons, and open
// registration: what a fresh installation does.
func DefaultLoginFlow() LoginFlow {
	return LoginFlow{
		Name:                 "Default sign-in",
		Slug:                 "default",
		Description:          "An email address and a password, with the registered providers beside them.",
		IsDefault:            true,
		IsEnabled:            true,
		Steps:                StepList{StepIdentifier, StepPassword, StepSocial},
		AllowSignIn:          true,
		AllowRegistration:    true,
		AllowPasswordReset:   true,
		AllowRememberMe:      true,
		RequireVerifiedEmail: false,
		SessionLifetimeHours: 24 * 14,
	}
}

// Validate reports the first thing wrong with a flow.
func (f LoginFlow) Validate() error {
	switch {
	case strings.TrimSpace(f.Name) == "":
		return fmt.Errorf("name is required")
	case len(f.Name) > 100:
		return fmt.Errorf("name must be at most 100 characters")
	}

	switch {
	case f.Slug == "":
		return fmt.Errorf("slug is required")
	case len(f.Slug) > 64:
		return fmt.Errorf("slug must be at most 64 characters")
	case !slugPattern.MatchString(f.Slug):
		return fmt.Errorf("slug must be lower case letters, numbers and dashes, such as staff-sign-in")
	}

	if len(f.Description) > 500 {
		return fmt.Errorf("description must be at most 500 characters")
	}

	if err := f.validateSteps(); err != nil {
		return err
	}

	if f.SessionLifetimeHours < 1 || f.SessionLifetimeHours > maxSessionLifetimeHours {
		return fmt.Errorf("session_lifetime_hours must be between 1 and %d", maxSessionLifetimeHours)
	}

	// The default is what every application without a flow of its own falls
	// back to, so there has to be something there to fall back to.
	if f.IsDefault && !f.IsEnabled {
		return fmt.Errorf("the default flow cannot be turned off")
	}

	return nil
}

// validateSteps holds the shape of a flow: it starts by asking who is signing
// in, says each thing once, and asks for at least one of them to be proved.
func (f LoginFlow) validateSteps() error {
	if len(f.Steps) == 0 {
		return fmt.Errorf("steps is required")
	}

	if f.Steps[0] != StepIdentifier {
		return fmt.Errorf("the first step must be %s", StepIdentifier)
	}

	seen := make(map[LoginStep]bool, len(f.Steps))
	for _, step := range f.Steps {
		if _, known := LoginStepSpecOf(step); !known {
			return fmt.Errorf("step must be one of: %s", strings.Join(LoginStepNames(), ", "))
		}
		if seen[step] {
			return fmt.Errorf("%s is listed twice", step)
		}
		seen[step] = true
	}

	for _, step := range proofSteps {
		if seen[step] {
			return nil
		}
	}

	return fmt.Errorf("a flow must ask for at least one of: %s", joinSteps(proofSteps))
}

// LoginStepNames is every step in the catalog, in catalog order.
func LoginStepNames() []string {
	names := make([]string, 0, len(LoginStepSpecs))
	for _, spec := range LoginStepSpecs {
		names = append(names, string(spec.Step))
	}

	return names
}

func joinSteps(steps []LoginStep) string {
	names := make([]string, 0, len(steps))
	for _, step := range steps {
		names = append(names, string(step))
	}

	return strings.Join(names, ", ")
}

// PublicLoginOptions is the part of a flow the sign-in pages are told: what
// they may offer, and nothing about how the server checks any of it.
type PublicLoginOptions struct {
	// Steps are the steps that will actually run, in order; planned steps are
	// left out.
	Steps []LoginStep `json:"steps"`

	// AllowSignIn false is a closed door: the pages say so instead of asking
	// for anything.
	AllowSignIn        bool `json:"allow_sign_in"`
	AllowRegistration  bool `json:"allow_registration"`
	AllowPasswordReset bool `json:"allow_password_reset"`

	// AllowRememberMe puts "stay signed in" beside the password.
	AllowRememberMe bool `json:"allow_remember_me"`

	// AllowEmailChange offers "change" beside the address on the account page.
	AllowEmailChange bool `json:"allow_email_change"`
}

// PublicOptions returns what the sign-in pages are told: only what will
// actually happen.
func (f LoginFlow) PublicOptions() PublicLoginOptions {
	steps := make([]LoginStep, 0, len(f.Steps))
	for _, step := range f.Steps {
		if spec, known := LoginStepSpecOf(step); known && spec.Implemented {
			steps = append(steps, step)
		}
	}

	return PublicLoginOptions{
		Steps:              steps,
		AllowSignIn:        f.AllowSignIn,
		AllowRegistration:  f.AllowRegistration,
		AllowPasswordReset: f.AllowPasswordReset,
		AllowRememberMe:    f.AllowRememberMe,
		AllowEmailChange:   f.AllowEmailChange,
	}
}

// Offers reports whether a flow names a step.
func (f LoginFlow) Offers(step LoginStep) bool {
	for _, named := range f.Steps {
		if named == step {
			return true
		}
	}

	return false
}
