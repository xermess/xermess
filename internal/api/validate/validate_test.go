package validate

import (
	"errors"
	"net/http"
	"testing"

	"xermess/internal/api/respond"
)

// fault is the fault an error carries, and whether it carried one at all.
func fault(err error) (respond.Fault, bool) {
	var f respond.Fault
	ok := errors.As(err, &f)

	return f, ok
}

type probe struct {
	Email string `json:"email" validate:"required,email"`
	Name  string `json:"name" validate:"max=4"`
	Kind  string `json:"kind" validate:"omitempty,oneof=one two"`
}

// TestStruct covers what a handler sees when a rule is broken: a 400, and a
// sentence that names the field the way the request named it.
func TestStruct(t *testing.T) {
	tests := []struct {
		name     string
		value    probe
		want     string // the message, or "" when the value is fine
		wantCode string
		// wantParams are what an app fills its own sentence with.
		wantParams map[string]any
	}{
		{
			name:  "everything in order",
			value: probe{Email: "a@b.com", Name: "Mira", Kind: "one"},
		},
		{
			name:       "a missing field",
			value:      probe{Name: "Mira"},
			want:       "email is required.",
			wantCode:   "validation.required",
			wantParams: map[string]any{"field": "email"},
		},
		{
			name:       "something that is not an address",
			value:      probe{Email: "not-an-address"},
			want:       "email must be an email address.",
			wantCode:   "validation.email",
			wantParams: map[string]any{"field": "email"},
		},
		{
			name:       "too long",
			value:      probe{Email: "a@b.com", Name: "Mirabel"},
			want:       "name must be at most 4 characters.",
			wantCode:   "validation.max",
			wantParams: map[string]any{"field": "name", "max": "4"},
		},
		{
			name:       "not one of the values allowed",
			value:      probe{Email: "a@b.com", Kind: "three"},
			want:       "kind must be one of: one, two.",
			wantCode:   "validation.oneof",
			wantParams: map[string]any{"field": "kind", "values": "one, two"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Struct(tt.value)

			if tt.want == "" {
				if err != nil {
					t.Fatalf("Struct() = %v, want nothing", err)
				}
				return
			}

			broken, ok := fault(err)
			if !ok {
				t.Fatalf("Struct() = %v, want a respond.Fault", err)
			}

			if broken.Status != http.StatusBadRequest {
				t.Errorf("status = %d, want 400", broken.Status)
			}
			if broken.Message != tt.want {
				t.Errorf("message = %q, want %q", broken.Message, tt.want)
			}
			if broken.Code != tt.wantCode {
				t.Errorf("code = %q, want %q", broken.Code, tt.wantCode)
			}
			for name, want := range tt.wantParams {
				if broken.Params[name] != want {
					t.Errorf("params[%s] = %v, want %v", name, broken.Params[name], want)
				}
			}
		})
	}
}

// A field with no json tag is named by the field itself, rather than by
// nothing at all.
func TestStructNamesAFieldWithoutATag(t *testing.T) {
	type untagged struct {
		Reference string `validate:"required"`
	}

	broken, ok := fault(Struct(untagged{}))
	if !ok {
		t.Fatal("want a respond.Fault")
	}

	if broken.Message != "Reference is required." {
		t.Errorf("message = %q, want %q", broken.Message, "Reference is required.")
	}
}

// A rule of our own is asked about the value, and breaking it is a problem of
// its own, `validation.<tag>`, which the panel's catalog says.
func TestRegister(t *testing.T) {
	Register("shouty", func(value string) bool {
		return value == "LOUD"
	})

	type probe struct {
		Word string `json:"word" validate:"shouty"`
	}

	if err := Struct(probe{Word: "LOUD"}); err != nil {
		t.Fatalf("Struct() = %v, want it accepted", err)
	}

	broken, ok := fault(Struct(probe{Word: "quiet"}))
	if !ok {
		t.Fatal("want a respond.Fault")
	}

	if broken.Code != "validation.shouty" || broken.Params["field"] != "word" {
		t.Errorf("fault = %q %v, want validation.shouty naming word", broken.Code, broken.Params)
	}
}

// Something that is not a struct is our own mistake, and must not come back
// as a 400 blaming whoever called the endpoint.
func TestStructRefusesToBlameTheCaller(t *testing.T) {
	err := Struct("not a struct")

	if err == nil {
		t.Fatal("Struct() = nothing, want an error")
	}

	if _, ok := fault(err); ok {
		t.Errorf("Struct() = %v, want it not to be a Fault", err)
	}
}

func TestFlag(t *testing.T) {
	yes, no := true, false

	if !Flag(nil, true) || Flag(nil, false) {
		t.Error("Flag(nil, current) should be current")
	}
	if !Flag(&yes, false) || Flag(&no, true) {
		t.Error("Flag(sent, current) should be what was sent")
	}
}
