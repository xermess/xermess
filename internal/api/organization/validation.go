package organization

import (
	"net/http"

	"loginer/internal/api/respond"
	"loginer/internal/api/validate"
	"loginer/internal/model"
)

// applyTo copies what a request sent onto the organisation and returns the
// settings it changed, in the order they are listed here, so the activity log
// can say what an administrator touched.
//
// The rules are the model's: what a name, a timezone or a logo may be is the
// same question wherever it is asked, and asking it in one place is what
// keeps an answer from drifting. Nothing is written when any of them is
// broken, since the record is only saved once this returns.
func (r *organizationRequest) applyTo(organization *model.Organization) ([]string, error) {
	// A short name is compared rather than read, so it is stored lower case:
	// the panel should not hold two spellings of one identifier.
	updated := *organization
	updated.Name = validate.Text(r.Name, organization.Name)
	updated.Slug = validate.Lower(r.Slug, organization.Slug)
	updated.LogoURL = validate.Text(r.LogoURL, organization.LogoURL)
	updated.SupportEmail = validate.Text(r.SupportEmail, organization.SupportEmail)
	updated.SupportPhone = validate.Text(r.SupportPhone, organization.SupportPhone)
	updated.TermsURL = validate.Text(r.TermsURL, organization.TermsURL)
	updated.PrivacyURL = validate.Text(r.PrivacyURL, organization.PrivacyURL)
	updated.Timezone = validate.Text(r.Timezone, organization.Timezone)

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
		{"logo_url", organization.LogoURL, updated.LogoURL},
		{"support_email", organization.SupportEmail, updated.SupportEmail},
		{"support_phone", organization.SupportPhone, updated.SupportPhone},
		{"terms_url", organization.TermsURL, updated.TermsURL},
		{"privacy_url", organization.PrivacyURL, updated.PrivacyURL},
		{"timezone", organization.Timezone, updated.Timezone},
	} {
		if field.was != field.became {
			changed = append(changed, field.name)
		}
	}

	*organization = updated

	return changed, nil
}
