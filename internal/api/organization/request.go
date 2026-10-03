package organization

// targetType is what this row is called in the activity log.
const targetType = "organization"

// organizationRequest is a PATCH: nil leaves a setting alone, an empty string
// clears one that may be empty.
type organizationRequest struct {
	Name         *string `json:"name"`
	Slug         *string `json:"slug"`
	LogoURL      *string `json:"logo_url"`
	SupportEmail *string `json:"support_email"`
	SupportPhone *string `json:"support_phone"`
	TermsURL     *string `json:"terms_url"`
	PrivacyURL   *string `json:"privacy_url"`
	Timezone     *string `json:"timezone"`
}
