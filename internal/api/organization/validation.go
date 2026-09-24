package organization

import (
	"net/http"

	"xermess/internal/api/respond"
	"xermess/internal/api/validate"
	"xermess/internal/model"
)

// applyTo copies what a request sent onto the organisation and returns the
// settings it changed, in the order they are listed here, so the activity log
// can say what an administrator touched.
//
// The rules are the model's: what a name, a domain or a logo may be is the
// same question wherever it is asked, and asking it in one place is what
// keeps an answer from drifting. Nothing is written when any of them is
// broken, since the record is only saved once this returns.
func (r *organizationRequest) applyTo(organization *model.Organization) ([]string, error) {
	// A host name and a short name are compared rather than read, so they are
	// stored lower case: "Example.com" and "example.com" are one domain, and
	// the panel should not hold two spellings of it.
	updated := *organization
	updated.Name = validate.Text(r.Name, organization.Name)
	updated.Slug = validate.Lower(r.Slug, organization.Slug)
	updated.Domain = validate.Lower(r.Domain, organization.Domain)
	updated.LogoURL = validate.Text(r.LogoURL, organization.LogoURL)
	updated.SupportEmail = validate.Text(r.SupportEmail, organization.SupportEmail)
	updated.SupportPhone = validate.Text(r.SupportPhone, organization.SupportPhone)
	updated.TermsURL = validate.Text(r.TermsURL, organization.TermsURL)
	updated.PrivacyURL = validate.Text(r.PrivacyURL, organization.PrivacyURL)

	if err := updated.Validate(); err != nil {
		return nil, respond.Fault{Status: http.StatusBadRequest, Message: err.Error()}
	}

	changed := []string{}
	for _, field := range []struct {
		name   string
		was    string
		became string
	}{
		{"name", organization.Name, updated.Name},
		{"slug", organization.Slug, updated.Slug},
		{"domain", organization.Domain, updated.Domain},
		{"logo_url", organization.LogoURL, updated.LogoURL},
		{"support_email", organization.SupportEmail, updated.SupportEmail},
		{"support_phone", organization.SupportPhone, updated.SupportPhone},
		{"terms_url", organization.TermsURL, updated.TermsURL},
		{"privacy_url", organization.PrivacyURL, updated.PrivacyURL},
	} {
		if field.was != field.became {
			changed = append(changed, field.name)
		}
	}

	*organization = updated

	return changed, nil
}
