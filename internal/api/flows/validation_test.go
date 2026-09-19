package flows

import (
	"errors"
	"reflect"
	"testing"

	"xermess/internal/api/respond"
	"xermess/internal/model"
)

func ptr[T any](value T) *T {
	return &value
}

func TestFlowRequestApplyToCreating(t *testing.T) {
	tests := []struct {
		name    string
		request flowRequest
		want    string // the message, or "" when the request is fine
		check   func(model.LoginFlow) error
	}{
		{
			name:    "a flow named and nothing else",
			request: flowRequest{Name: ptr("  Staff sign-in  ")},
			check: func(f model.LoginFlow) error {
				switch {
				case f.Name != "Staff sign-in":
					return errors.New("name = " + f.Name)
				case f.Slug != "staff-sign-in":
					return errors.New("slug = " + f.Slug + ", want one made from the name")
				case f.IsDefault:
					return errors.New("a new flow made itself the default")
				case f.Enabled:
					return errors.New("a new flow offered itself before anyone turned it on")
				}
				return nil
			},
		},
		{
			name:    "an identifier of its own",
			request: flowRequest{Name: ptr("Staff sign-in"), Slug: ptr("  STAFF  ")},
			check: func(f model.LoginFlow) error {
				if f.Slug != "staff" {
					return errors.New("slug = " + f.Slug)
				}
				return nil
			},
		},
		{
			name: "steps with the blanks a form sends",
			request: flowRequest{
				Name:  ptr("Codes and a password"),
				Steps: ptr([]model.LoginStep{"Identifier", "", " email_code ", "password"}),
			},
			check: func(f model.LoginFlow) error {
				want := model.StepList{model.StepIdentifier, model.StepEmailCode, model.StepPassword}
				if !reflect.DeepEqual(f.Steps, want) {
					return errors.New("the steps were not tidied")
				}
				return nil
			},
		},
		{
			name:    "making it the default also turns it on",
			request: flowRequest{Name: ptr("Staff"), IsDefault: ptr(true), Enabled: ptr(false)},
			check: func(f model.LoginFlow) error {
				if !f.Enabled {
					return errors.New("the default flow was left turned off")
				}
				return nil
			},
		},
		{
			name:    "no name",
			request: flowRequest{Slug: ptr("staff")},
			want:    "name is required",
		},
		{
			name:    "a flow that asks for nothing",
			request: flowRequest{Name: ptr("Open door"), Steps: ptr([]model.LoginStep{"identifier"})},
			want:    "a flow must ask for at least one of: password, social",
		},
		{
			name:    "a session that lasts a year",
			request: flowRequest{Name: ptr("Staff"), SessionLifetimeHours: ptr(24 * 365)},
			want:    "session_lifetime_hours must be between 1 and 2160",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := tt.request

			flow := model.DefaultLoginFlow()
			flow.IsDefault, flow.Enabled = false, false
			flow.Name, flow.Slug, flow.Description = "", "", ""

			err := request.applyTo(&flow, true)

			if tt.want != "" {
				var fault respond.Fault
				if !errors.As(err, &fault) {
					t.Fatalf("applyTo() = %v, want a Fault", err)
				}
				if fault.Message != tt.want {
					t.Errorf("message = %q, want %q", fault.Message, tt.want)
				}
				return
			}

			if err != nil {
				t.Fatalf("applyTo() = %v, want nothing", err)
			}
			if tt.check != nil {
				if err := tt.check(flow); err != nil {
					t.Error(err)
				}
			}
		})
	}
}

// An update says what it changes and nothing else: turning a flow off must
// not empty the steps it was built out of.
func TestFlowRequestUpdatesOnlyWhatItMentions(t *testing.T) {
	flow := model.DefaultLoginFlow()
	flow.IsDefault = false
	was := flow

	request := flowRequest{Enabled: ptr(false)}

	if err := request.applyTo(&flow, false); err != nil {
		t.Fatalf("applyTo() = %v, want nothing", err)
	}

	switch {
	case flow.Enabled:
		t.Error("the flow was not turned off")
	case flow.Name != was.Name || flow.Description != was.Description:
		t.Errorf("the name or the description was emptied: %+v", flow)
	case !reflect.DeepEqual(flow.Steps, was.Steps):
		t.Errorf("the steps were emptied: %v", flow.Steps)
	case flow.SessionLifetimeHours != was.SessionLifetimeHours:
		t.Error("a setting nobody mentioned was changed")
	case !flow.AllowRegistration || !flow.AllowPasswordReset:
		t.Error("a switch nobody mentioned was thrown")
	}
}

// The identifier is what an application stores to name its flow, so an update
// may not move it.
func TestFlowRequestKeepsTheSlug(t *testing.T) {
	flow := model.DefaultLoginFlow()

	request := flowRequest{Slug: ptr("something-else"), Name: ptr("Renamed")}
	if err := request.applyTo(&flow, false); err != nil {
		t.Fatalf("applyTo() = %v", err)
	}

	switch {
	case flow.Slug != "default":
		t.Errorf("an update changed the identifier: %s", flow.Slug)
	case flow.Name != "Renamed":
		t.Errorf("the name did not change: %s", flow.Name)
	}
}

func TestSlugify(t *testing.T) {
	tests := []struct {
		name string
		from string
		want string
	}{
		{name: "a name with spaces", from: "Staff sign-in", want: "staff-sign-in"},
		{name: "a name with punctuation", from: "Staff (2024)!", want: "staff-2024"},
		{name: "a name that is already one", from: "staff", want: "staff"},
		{name: "a name of nothing but punctuation", from: "!!!", want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := slugify(nil, &tt.from); got != tt.want {
				t.Errorf("slugify(%q) = %q, want %q", tt.from, got, tt.want)
			}
		})
	}
}
