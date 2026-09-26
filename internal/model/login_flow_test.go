package model

import (
	"slices"
	"testing"
)

// Accepts is what an application asks of a session some other sign-in made.
// It has to answer as the ways in answer when a session is being made, or a
// flow's rules would hold only for the application somebody signed in
// through.
func TestLoginFlowAccepts(t *testing.T) {
	flow := func(steps ...LoginStep) LoginFlow {
		return LoginFlow{Steps: append(StepList{StepIdentifier}, steps...)}
	}

	tests := []struct {
		name   string
		flow   LoginFlow
		method SignInMethod
		want   bool
	}{
		{"a password, where the flow asks for one", flow(StepPassword), MethodPassword, true},
		{"a password, where the flow asks for a code as well", flow(StepPassword, StepEmailCode), MethodPassword, false},
		{"a password and a code, where the flow asks for both", flow(StepPassword, StepEmailCode), MethodEmailCode, true},
		{"a password and a code, where the flow asks only for the password", flow(StepPassword), MethodEmailCode, true},
		{"a provider, where the flow offers the buttons", flow(StepSocial), MethodSocial, true},
		{"a provider, where the flow does not offer them", flow(StepPassword), MethodSocial, false},
		{"a password, where the flow takes none", flow(StepSocial), MethodPassword, false},
		// A connection decides who signs in through it; the steps never had
		// a say, so a session it made satisfies any flow.
		{"an organisation's provider, whatever the steps are", flow(StepPassword, StepEmailCode), MethodSSO, true},
		// A session made before the server wrote down what proved it, and one
		// claiming something this server has never issued.
		{"a session with nothing recorded against it", flow(StepPassword), "", false},
		{"a way in this server knows nothing about", flow(StepPassword), "fingerprint", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.flow.Accepts(tt.method); got != tt.want {
				t.Errorf("Accepts(%q) = %v, want %v", tt.method, got, tt.want)
			}
		})
	}
}

func TestLoginFlowValidate(t *testing.T) {
	tests := []struct {
		name   string
		change func(*LoginFlow)
		want   string // the message, or "" when the flow is fine
	}{
		{name: "the flow a new installation starts with", change: func(*LoginFlow) {}},
		{
			name: "a flow that only offers other accounts",
			change: func(f *LoginFlow) {
				f.Steps = StepList{StepIdentifier, StepSocial}
			},
		},
		{
			name:   "a flow nothing has been written about",
			change: func(f *LoginFlow) { f.Description = "" },
		},
		{
			name:   "no name",
			change: func(f *LoginFlow) { f.Name = "  " },
			want:   "name is required",
		},
		{
			name:   "a slug with spaces in it",
			change: func(f *LoginFlow) { f.Slug = "staff sign in" },
			want:   "slug must be lower case letters, numbers and dashes, such as staff-sign-in",
		},
		{
			name:   "no steps at all",
			change: func(f *LoginFlow) { f.Steps = StepList{} },
			want:   "steps is required",
		},
		{
			name:   "a flow that never asks who is signing in",
			change: func(f *LoginFlow) { f.Steps = StepList{StepPassword} },
			want:   "the first step must be identifier",
		},
		{
			name:   "a step nobody offers",
			change: func(f *LoginFlow) { f.Steps = StepList{StepIdentifier, "fingerprint"} },
			want:   "step must be one of: identifier, password, social, email_code, totp, terms, consent",
		},
		{
			name: "the same step twice",
			change: func(f *LoginFlow) {
				f.Steps = StepList{StepIdentifier, StepPassword, StepPassword}
			},
			want: "password is listed twice",
		},
		{
			name:   "a flow that asks for an address and nothing else",
			change: func(f *LoginFlow) { f.Steps = StepList{StepIdentifier, StepTerms} },
			want:   "a flow must ask for at least one of: password, social",
		},
		{
			name:   "a session that never ends",
			change: func(f *LoginFlow) { f.SessionLifetimeHours = 0 },
			want:   "session_lifetime_hours must be between 1 and 2160",
		},
		{
			name:   "a session lasting a year",
			change: func(f *LoginFlow) { f.SessionLifetimeHours = 24 * 365 },
			want:   "session_lifetime_hours must be between 1 and 2160",
		},
		{
			name:   "turning off the flow everything falls back to",
			change: func(f *LoginFlow) { f.Enabled = false },
			want:   "the default flow cannot be turned off",
		},
		{
			name: "turning off a flow that is not the default",
			change: func(f *LoginFlow) {
				f.IsDefault = false
				f.Enabled = false
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flow := DefaultLoginFlow()
			tt.change(&flow)

			err := flow.Validate()

			switch {
			case tt.want == "" && err != nil:
				t.Fatalf("Validate() = %v, want nothing", err)
			case tt.want == "":
				return
			case err == nil:
				t.Fatalf("Validate() = nothing, want %q", tt.want)
			case err.Error() != tt.want:
				t.Errorf("Validate() = %q, want %q", err, tt.want)
			}
		})
	}
}

// Every step a flow can name has to be in the catalog the panel builds its
// step picker from, or it is a step nobody can add and nobody can explain.
func TestLoginStepSpecsCoverEveryStep(t *testing.T) {
	steps := []LoginStep{
		StepIdentifier, StepPassword, StepSocial, StepEmailCode, StepTOTP, StepTerms, StepConsent,
	}

	if got, want := len(LoginStepSpecs), len(steps); got != want {
		t.Fatalf("the catalog lists %d steps, want %d", got, want)
	}

	for _, step := range steps {
		spec, ok := LoginStepSpecOf(step)
		switch {
		case !ok:
			t.Errorf("%s is not in the catalog", step)
		case spec.Label == "" || spec.Description == "":
			t.Errorf("%s has nothing said about it", step)
		}
	}
}

// The default flow is what the store falls back to and what the migration
// seeds, so it has to pass the same rules as one an administrator writes.
func TestDefaultLoginFlowIsValid(t *testing.T) {
	flow := DefaultLoginFlow()

	if err := flow.Validate(); err != nil {
		t.Fatalf("DefaultLoginFlow() = %v", err)
	}
	if !flow.IsDefault || !flow.Enabled {
		t.Error("the default flow has to be the default, and on")
	}
	if !flow.Offers(StepPassword) {
		t.Error("the default flow is the one the server already runs, which asks for a password")
	}
}

// What the sign-in pages are told is what will happen to somebody signing in.
// A flow may name a step as a plan — the panel draws those dashed and says so
// — but the pages get a list to put on the screen, and a step nobody will be
// asked for does not belong in it.
func TestPublicOptionsLeaveOutTheStepsNobodyRunsYet(t *testing.T) {
	flow := LoginFlow{
		Steps:             StepList{StepIdentifier, StepPassword, StepTOTP, StepSocial, StepConsent},
		AllowRegistration: true,
	}

	options := flow.PublicOptions()

	want := []LoginStep{StepIdentifier, StepPassword, StepSocial}
	if !slices.Equal(options.Steps, want) {
		t.Errorf("steps = %v, want %v", options.Steps, want)
	}
	if !options.AllowRegistration {
		t.Error("the rest of the options did not come through")
	}

	// Every step left in is one the server runs, whatever the catalog gains.
	for _, step := range options.Steps {
		spec, known := LoginStepSpecOf(step)
		if !known || !spec.Implemented {
			t.Errorf("%s is offered to the sign-in pages but is not run", step)
		}
	}
}
