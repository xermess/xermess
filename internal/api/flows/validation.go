package flows

import (
	"net/http"
	"strings"

	"loginer/internal/api/respond"
	"loginer/internal/api/validate"
	"loginer/internal/model"
)

// applyTo checks the request and copies it onto a flow.
//
// `creating` says whether the identifier is read from it. A flow's slug is
// what an application stores to name it, so it is set once and then stays.
//
// Everything else is copied only where the request mentioned it, so turning a
// flow off says nothing about its steps.
func (r *flowRequest) applyTo(flow *model.LoginFlow, creating bool) error {
	if creating {
		flow.Slug = slugify(r.Slug, r.Name)
	}

	flow.Name = validate.Text(r.Name, flow.Name)
	flow.Description = validate.Text(r.Description, flow.Description)

	flow.IsDefault = validate.Flag(r.IsDefault, flow.IsDefault)
	flow.Enabled = validate.Flag(r.Enabled, flow.Enabled)

	flow.AllowSignIn = validate.Flag(r.AllowSignIn, flow.AllowSignIn)
	flow.AllowRegistration = validate.Flag(r.AllowRegistration, flow.AllowRegistration)
	flow.AllowPasswordReset = validate.Flag(r.AllowPasswordReset, flow.AllowPasswordReset)
	flow.AllowRememberMe = validate.Flag(r.AllowRememberMe, flow.AllowRememberMe)
	flow.VerifyEmailOnRegister = validate.Flag(r.VerifyEmailOnRegister, flow.VerifyEmailOnRegister)
	flow.RequireVerifiedEmail = validate.Flag(r.RequireVerifiedEmail, flow.RequireVerifiedEmail)
	flow.AllowEmailChange = validate.Flag(r.AllowEmailChange, flow.AllowEmailChange)

	if r.SessionLifetimeHours != nil {
		flow.SessionLifetimeHours = *r.SessionLifetimeHours
	}

	if r.Steps != nil {
		flow.Steps = cleanSteps(*r.Steps)
	}

	// Marking a flow as the default is also turning it on: it is what every
	// application without one of its own falls back to, and the model refuses
	// a default that is off. Saying so here saves an administrator a second
	// save to find that out.
	if flow.IsDefault {
		flow.Enabled = true
	}

	// What a flow has to be is the model's to say.
	if err := flow.Validate(); err != nil {
		return badRequest(err.Error())
	}

	return nil
}

// cleanSteps tidies a list of steps: trimmed, lower case, and with the blanks
// a form sends dropped. What is left is the model's to judge.
func cleanSteps(sent []model.LoginStep) model.StepList {
	steps := make(model.StepList, 0, len(sent))

	for _, step := range sent {
		tidied := model.LoginStep(strings.ToLower(strings.TrimSpace(string(step))))
		if tidied != "" {
			steps = append(steps, tidied)
		}
	}

	return steps
}

// slugify is the identifier a new flow gets: the one that was sent, or one
// made from its name — "Staff sign-in" becomes "staff-sign-in", so the
// common case needs no second field filled in.
func slugify(sent, name *string) string {
	if sent != nil && strings.TrimSpace(*sent) != "" {
		return strings.ToLower(strings.TrimSpace(*sent))
	}

	if name == nil {
		return ""
	}

	var b strings.Builder
	dash := false

	for _, r := range strings.ToLower(strings.TrimSpace(*name)) {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
			dash = false
		case b.Len() > 0 && !dash:
			b.WriteRune('-')
			dash = true
		}
	}

	return strings.TrimSuffix(b.String(), "-")
}

// badRequest is a 400 carrying what was wrong with the request.
func badRequest(message string) error {
	return respond.Fault{Status: http.StatusBadRequest, Message: message}
}
