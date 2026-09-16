package organization

import (
	"errors"
	"reflect"
	"testing"

	"xermess/internal/api/respond"
	"xermess/internal/model"
)

// stored is the organisation a request in these tests is applied to.
func stored() model.Organization {
	return model.Organization{
		Name:         "Acme",
		Slug:         "acme",
		Domain:       "acme.example",
		LogoURL:      "https://acme.example/logo.svg",
		SupportEmail: "support@acme.example",
		SupportPhone: "+996 555 123456",
		TermsURL:     "https://acme.example/terms",
		PrivacyURL:   "https://acme.example/privacy",
	}
}

func ptr(value string) *string {
	return &value
}

func TestOrganizationRequestApplyTo(t *testing.T) {
	tests := []struct {
		name    string
		request organizationRequest
		want    string   // the message, or "" when the request is fine
		changed []string // the settings it should report as changed
		check   func(model.Organization) error
	}{
		{
			name:    "nothing sent changes nothing",
			request: organizationRequest{},
			changed: []string{},
		},
		{
			name:    "a new name, trimmed",
			request: organizationRequest{Name: ptr("  Acme Inc  ")},
			changed: []string{"name"},
			check: func(o model.Organization) error {
				if o.Name != "Acme Inc" {
					return errors.New("name = " + o.Name)
				}
				return nil
			},
		},
		{
			name:    "a domain is stored lower case",
			request: organizationRequest{Domain: ptr("Acme.Example.COM")},
			changed: []string{"domain"},
			check: func(o model.Organization) error {
				if o.Domain != "acme.example.com" {
					return errors.New("domain = " + o.Domain)
				}
				return nil
			},
		},
		{
			name:    "an empty string clears a setting that may be empty",
			request: organizationRequest{Domain: ptr(""), SupportPhone: ptr("")},
			changed: []string{"domain", "support_phone"},
		},
		{
			name:    "the same values again are no change",
			request: organizationRequest{Name: ptr("Acme"), Slug: ptr("acme")},
			changed: []string{},
		},
		{
			name:    "new agreements",
			request: organizationRequest{TermsURL: ptr("https://acme.example/legal/terms")},
			changed: []string{"terms_url"},
		},
		{
			name:    "a name cleared",
			request: organizationRequest{Name: ptr("   ")},
			want:    "name is required",
		},
		{
			name:    "a slug with spaces",
			request: organizationRequest{Slug: ptr("acme inc")},
			want:    "slug must be lower case letters, numbers and dashes, such as acme-inc",
		},
		{
			name:    "a domain with a scheme",
			request: organizationRequest{Domain: ptr("https://acme.example")},
			want:    "domain must be a host name on its own, such as example.com",
		},
		{
			name:    "an address that is not one",
			request: organizationRequest{SupportEmail: ptr("support at acme")},
			want:    "support_email must be an email address",
		},
		{
			name:    "a logo that is not a full address",
			request: organizationRequest{LogoURL: ptr("/logo.svg")},
			want:    "logo_url must be a full address starting with http:// or https://",
		},
		{
			name:    "a logo that could run something",
			request: organizationRequest{LogoURL: ptr("javascript:alert(1)")},
			want:    "logo_url must be a full address starting with http:// or https://",
		},
		{
			name:    "a support number nobody could dial",
			request: organizationRequest{SupportPhone: ptr("ring us")},
			want:    "support_phone must be a number someone can dial, such as +996 555 123456",
		},
		{
			name:    "terms that are not a full address",
			request: organizationRequest{TermsURL: ptr("example.com/terms")},
			want:    "terms_url must be a full address starting with http:// or https://",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := tt.request
			organization := stored()

			changed, err := request.applyTo(&organization)

			if tt.want != "" {
				var fault respond.Fault
				if !errors.As(err, &fault) {
					t.Fatalf("applyTo() = %v, want a Fault", err)
				}
				if fault.Message != tt.want {
					t.Errorf("message = %q, want %q", fault.Message, tt.want)
				}
				// Nothing may be written when a request is refused.
				if !reflect.DeepEqual(organization, stored()) {
					t.Errorf("the organization was changed by a refused request")
				}
				return
			}

			if err != nil {
				t.Fatalf("applyTo() = %v, want nothing", err)
			}
			if !reflect.DeepEqual(changed, tt.changed) {
				t.Errorf("changed = %v, want %v", changed, tt.changed)
			}
			if tt.check != nil {
				if err := tt.check(organization); err != nil {
					t.Error(err)
				}
			}
		})
	}
}
