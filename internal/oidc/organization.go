package oidc

import (
	"context"

	"xermess/internal/model"
)

// PublicOrganization is what the sign-in pages show about the organisation
// this server signs users in for: who they are dealing with, how to reach
// somebody, and the agreements they are accepting.
//
// It is built by hand, like PublicApplication, because these pages are served
// to anyone who can reach a sign-in: only what a stranger may see is in it.
type PublicOrganization struct {
	Name         string `json:"name"`
	LogoURL      string `json:"logo_url"`
	Domain       string `json:"domain"`
	SupportEmail string `json:"support_email"`
	SupportPhone string `json:"support_phone"`
	TermsURL     string `json:"terms_url"`
	PrivacyURL   string `json:"privacy_url"`
}

// PublicOrg returns those fields of an organisation.
func PublicOrg(organization *model.Organization) PublicOrganization {
	return PublicOrganization{
		Name:         organization.Name,
		LogoURL:      organization.LogoURL,
		Domain:       organization.Domain,
		SupportEmail: organization.SupportEmail,
		SupportPhone: organization.SupportPhone,
		TermsURL:     organization.TermsURL,
		PrivacyURL:   organization.PrivacyURL,
	}
}

// Organization is what the sign-in pages ask for when they draw themselves.
func (s *Service) Organization(ctx context.Context) (PublicOrganization, error) {
	organization, err := s.store.Organization(ctx)
	if err != nil {
		return PublicOrganization{}, err
	}

	return PublicOrg(organization), nil
}
