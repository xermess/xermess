package organization

// targetType is what this row is called in the activity log.
const targetType = "organization"

// organizationRequest is what an update sends.
//
// Every field is a pointer because this is a PATCH: nil is "not sent", and
// the stored value stands. An empty string is sent on purpose and clears the
// setting — which only the ones that may be empty accept.
type organizationRequest struct {
	Name         *string `json:"name"`
	Slug         *string `json:"slug"`
	Domain       *string `json:"domain"`
	LogoURL      *string `json:"logo_url"`
	SupportEmail *string `json:"support_email"`
	SupportPhone *string `json:"support_phone"`
	TermsURL     *string `json:"terms_url"`
	PrivacyURL   *string `json:"privacy_url"`
}
