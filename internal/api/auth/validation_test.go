package auth

import (
	"errors"
	"net/http"
	"testing"

	"xermess/internal/api/respond"
)

// TestLoginRequestValidate covers the sign-in form: both fields are wanted,
// and a username of nothing but spaces is nothing.
func TestLoginRequestValidate(t *testing.T) {
	tests := []struct {
		name    string
		request loginRequest
		want    string // the message, or "" when the form is fine
	}{
		{
			name:    "a filled in form",
			request: loginRequest{Username: "admin", Password: "hunter2"},
		},
		{
			name:    "the username is trimmed",
			request: loginRequest{Username: "  admin  ", Password: "hunter2"},
		},
		{
			name:    "no username",
			request: loginRequest{Password: "hunter2"},
			want:    "username is required.",
		},
		{
			name:    "a username of spaces",
			request: loginRequest{Username: "   ", Password: "hunter2"},
			want:    "username is required.",
		},
		{
			name:    "no password",
			request: loginRequest{Username: "admin"},
			want:    "password is required.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := tt.request
			err := request.validate()

			if tt.want == "" {
				if err != nil {
					t.Fatalf("validate() = %v, want nothing", err)
				}
				if request.Username != "admin" {
					t.Errorf("username = %q, want it trimmed", request.Username)
				}
				return
			}

			var fault respond.Fault
			if !errors.As(err, &fault) {
				t.Fatalf("validate() = %v, want a respond.Fault", err)
			}

			if fault.Status != http.StatusBadRequest {
				t.Errorf("status = %d, want 400", fault.Status)
			}
			if fault.Message != tt.want {
				t.Errorf("message = %q, want %q", fault.Message, tt.want)
			}
		})
	}
}
