package organization

import (
	"time"

	"loginer/internal/model"
)

// organizationResponse is the organisation as the panel sees it. It is built
// by hand rather than returning the model so the row's own bookkeeping — its
// id, when it was last written — stays out of a record of settings, and the
// one date worth showing is named for what it means.
type organizationResponse struct {
	Name         string    `json:"name"`
	Slug         string    `json:"slug"`
	LogoURL      string    `json:"logo_url"`
	SupportEmail string    `json:"support_email"`
	SupportPhone string    `json:"support_phone"`
	TermsURL     string    `json:"terms_url"`
	PrivacyURL   string    `json:"privacy_url"`
	Timezone     string    `json:"timezone"`
	CreatedAt    time.Time `json:"created_at"`
}

// response is what both endpoints answer with.
type response struct {
	Organization organizationResponse `json:"organization"`
}

func newResponse(organization model.Organization) response {
	return response{
		Organization: organizationResponse{
			Name:         organization.Name,
			Slug:         organization.Slug,
			LogoURL:      organization.LogoURL,
			SupportEmail: organization.SupportEmail,
			SupportPhone: organization.SupportPhone,
			TermsURL:     organization.TermsURL,
			PrivacyURL:   organization.PrivacyURL,
			Timezone:     organization.Timezone,
			CreatedAt:    organization.CreatedAt,
		},
	}
}
