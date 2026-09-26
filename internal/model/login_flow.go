package model

import (
	"fmt"
	"strings"
)

// LoginFlow is what somebody is taken through when they sign in: the steps,
// in order, and what they are allowed to do along the way.
//
// There are as many as an installation wants, and one of them is the default
// — the flow used by every application that does not name another. An
// application points at one with its LoginFlowID, so a staff tool can ask for
// a code from an authenticator while the shop asks for a password alone,
// without either of them holding a copy of the rules.
//
// Every way in follows the flow of the application it is for: a flow without
// the password step refuses a password, one without the social step refuses
// a provider, an address the flow requires verified is sent a link instead
// of a session, and the session lasts as long as the flow says. The steps
// that are not run yet — Implemented false in the catalog — can be named as a
// plan, but a flow has to include one that is, so nothing the panel saves
// can leave everybody locked out.
type LoginFlow struct {
	Base

	// Name is what the flow is called in the panel.
	Name string `gorm:"size:100;not null" json:"name"`

	// Slug is the short name an application or an export refers to it by.
	Slug string `gorm:"size:64;not null;uniqueIndex" json:"slug"`

	// Description is what this flow is for, in the administrator's words.
	Description string `gorm:"size:500" json:"description"`

	// IsDefault marks the flow used when an application names none. Exactly
	// one flow has it; the store moves it rather than letting two rows claim
	// it, and the default flow cannot be turned off or removed.
	IsDefault bool `gorm:"not null" json:"is_default"`

	// Enabled says whether an application may be pointed at this flow. A
	// flow that is off stays as it is, and the applications holding it fall
	// back to the default.
	Enabled bool `gorm:"not null" json:"enabled"`

	// Steps are the steps in the order they are taken.
	Steps StepList `gorm:"type:text;serializer:json;not null" json:"steps"`

	// AllowSignIn is whether anybody may sign in through this flow at all.
	// Off, every way in is refused — a password, a provider, a code, and
	// registering, which ends in a session like the rest.
	//
	// It is not Enabled above: that says whether an application may be
	// pointed at this flow, and an application pointed at one that is off
	// falls back to the default and signs its users in as usual. This says
	// the sign-ins themselves are closed, which is what an installation
	// reaches for while it is being worked on.
	AllowSignIn bool `gorm:"not null" json:"allow_sign_in"`

	// AllowRegistration offers the "create an account" way out of the
	// sign-in page. Off, an account is made by an administrator or not at
	// all.
	AllowRegistration bool `gorm:"not null" json:"allow_registration"`

	// AllowPasswordReset offers the "forgotten your password" link.
	AllowPasswordReset bool `gorm:"not null" json:"allow_password_reset"`

	// AllowRememberMe offers "stay signed in" beside the password. Ticked,
	// the session lasts SessionLifetimeHours; left alone — or not offered at
	// all — the cookie ends with the browser, so a shared machine forgets
	// whoever used it last when its window closes.
	AllowRememberMe bool `gorm:"not null" json:"allow_remember_me"`

	// VerifyEmailOnRegister sends a new account a link to confirm its
	// address. It does not hold the account back: the person is signed in
	// and the link is waiting for them.
	//
	// RequireVerifiedEmail below is the other half of the same subject and a
	// different decision — whether an unconfirmed address may sign in at
	// all. A flow can send the link and let people in (this alone), refuse
	// until it is used (both), or neither.
	VerifyEmailOnRegister bool `gorm:"not null" json:"verify_email_on_register"`

	// RequireVerifiedEmail refuses a sign-in until the address has been
	// confirmed, rather than letting an unconfirmed account in and nagging.
	RequireVerifiedEmail bool `gorm:"not null" json:"require_verified_email"`

	// AllowEmailChange lets a user change the address they sign in with,
	// from their own account page. The new one is confirmed by a link before
	// it replaces the old, so nobody can take an account by typing an
	// address they do not read.
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

	// Implemented says the sign-in pages run this step today. A flow may
	// name one that is not, and the panel marks it: the flow is a plan the
	// server is being built towards, not a promise it already keeps.
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

// proofSteps are the steps that prove the person is who they say and that
// the sign-in pages run today. A flow with none of them would ask for an
// address and let nobody in — or, if the planned steps counted, let nobody in
// until they are built.
var proofSteps = []LoginStep{StepPassword, StepSocial}

// maxSessionLifetimeHours is ninety days: long enough for "stay signed in",
// short enough that a session left behind on a shared machine expires.
const maxSessionLifetimeHours = 24 * 90

// DefaultLoginFlow is the flow a fresh installation starts with, and the one
// the store falls back to: an address and a password, the provider buttons
// beside them, and an account anybody can make.
//
// It is what the sign-in pages do today, written down — so the first thing an
// administrator sees on the Login flows page is the server they already have.
func DefaultLoginFlow() LoginFlow {
	return LoginFlow{
		Name:                 "Default sign-in",
		Slug:                 "default",
		Description:          "An email address and a password, with the registered providers beside them.",
		IsDefault:            true,
		Enabled:              true,
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
	if f.IsDefault && !f.Enabled {
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
	// Steps are the steps that will actually happen, in order. A step the
	// server does not run yet is left out rather than sent along: to the
	// pages this is a list of what to put on the screen, and a step nobody
	// will be asked for has no business in it.
	Steps []LoginStep `json:"steps"`

	// AllowSignIn false is a closed door: the pages say so instead of asking
	// for anything.
	AllowSignIn        bool `json:"allow_sign_in"`
	AllowRegistration  bool `json:"allow_registration"`
	AllowPasswordReset bool `json:"allow_password_reset"`

	// AllowRememberMe puts "stay signed in" beside the password.
	AllowRememberMe bool `json:"allow_remember_me"`

	// AllowEmailChange offers "change" beside the address on the account
	// page. It is read there rather than on the sign-in page, but it is the
	// flow's to decide like the rest.
	AllowEmailChange bool `json:"allow_email_change"`
}

// PublicOptions returns those fields. The panel is told about the steps a flow
// only plans — that is what its Planned list is for — but the sign-in pages
// are told what happens.
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

// Accepts reports whether a flow signs somebody in on the strength of a
// session another sign-in already made.
//
// It is the same question the ways in answer when a session is being made, so
// that the two cannot drift: a flow that does not offer passwords does not
// take a session a password made (Service.SignIn refuses the password
// itself), one without the provider buttons does not take a session a
// provider made (Service.SocialStart refuses the button), and one that asks
// for an emailed code is not satisfied by the password that came before the
// code (Service.finishSignIn would send one).
//
// A method this flow knows nothing about is refused, which is also the answer
// for a session made before the server recorded methods at all: whoever holds
// it signs in once more, and the session that replaces it says what proved it.
func (f LoginFlow) Accepts(method SignInMethod) bool {
	switch method {
	case MethodPassword:
		return f.Offers(StepPassword) && !f.Offers(StepEmailCode)
	case MethodEmailCode:
		return f.Offers(StepPassword)
	case MethodSocial:
		return f.Offers(StepSocial)
	case MethodSSO:
		// An organisation's identity provider is not one of the steps: the
		// connection decides who reaches this server through it, and which
		// steps a flow names has never had a say (Service.signInThroughSSO).
		return true
	}

	return false
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
