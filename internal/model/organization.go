package model

import (
	"fmt"
	"net/mail"
	"net/url"
	"regexp"
	"strings"
	"time"

	"loginer/internal/brand"
)

// Organization is the single row describing who the installation belongs to.
// Its terms and privacy links are published in discovery (op_tos_uri,
// op_policy_uri), and the sign-in pages fall back to its name, logo and
// contact. Only a super admin may change it.
type Organization struct {
	Base

	// Name is the organisation as people refer to it, and what the sign-in
	// pages say when the application has no name to show.
	Name string `gorm:"size:100;not null" json:"name"`

	// Slug is the short name it is known by where a name with spaces will not
	// do — a subdomain, a directory, an export.
	Slug string `gorm:"size:64;not null;uniqueIndex" json:"slug"`

	// LogoURL is an absolute http(s) URL to the organisation's logo.
	LogoURL string `gorm:"size:512" json:"logo_url"`

	// SupportEmail and SupportPhone are shown on the sign-in pages for people
	// who cannot get in.
	SupportEmail string `gorm:"size:255" json:"support_email"`
	SupportPhone string `gorm:"size:32" json:"support_phone"`

	// TermsURL and PrivacyURL are accepted on registration; an application's
	// own links replace them.
	TermsURL   string `gorm:"size:512" json:"terms_url"`
	PrivacyURL string `gorm:"size:512" json:"privacy_url"`
	// Timezone is the IANA zone account pages show dates in, so server and
	// browser agree on the day.
	Timezone string `gorm:"size:64;not null;default:UTC" json:"timezone"`
}

// TableName pins the table name.
func (Organization) TableName() string {
	return "organizations"
}

// DefaultOrganization is the row a fresh installation starts with, seeded by
// the migration and used as the store's fallback.
func DefaultOrganization() Organization {
	return Organization{Name: brand.Name, Slug: brand.Slug, Timezone: "UTC"}
}

// slugPattern is what a short name may look like: lower case letters, numbers
// and dashes, starting and ending with one of the first two.
var slugPattern = regexp.MustCompile(`^[a-z0-9]([a-z0-9-]*[a-z0-9])?$`)

// phonePattern is deliberately loose: digits with common separators, optionally
// international.
var phonePattern = regexp.MustCompile(`^\+?[0-9][0-9 ()./-]{4,30}$`)

// Validate reports the first thing wrong with an organisation.
func (o Organization) Validate() error {
	switch {
	case strings.TrimSpace(o.Name) == "":
		return fmt.Errorf("name is required")
	case len(o.Name) > 100:
		return fmt.Errorf("name must be at most 100 characters")
	}

	switch {
	case o.Slug == "":
		return fmt.Errorf("slug is required")
	case len(o.Slug) > 64:
		return fmt.Errorf("slug must be at most 64 characters")
	case !slugPattern.MatchString(o.Slug):
		return fmt.Errorf("slug must be lower case letters, numbers and dashes, such as acme-inc")
	}

	if err := validTimezone(o.Timezone); err != nil {
		return err
	}

	if o.SupportEmail != "" {
		if _, err := mail.ParseAddress(o.SupportEmail); err != nil {
			return fmt.Errorf("support_email must be an email address")
		}
	}

	if o.SupportPhone != "" {
		if len(o.SupportPhone) > 32 || !phonePattern.MatchString(o.SupportPhone) {
			return fmt.Errorf("support_phone must be a number someone can dial, such as +996 555 123456")
		}
	}

	for _, link := range []struct {
		field string
		value string
	}{
		{"logo_url", o.LogoURL},
		{"terms_url", o.TermsURL},
		{"privacy_url", o.PrivacyURL},
	} {
		if err := ValidLink(link.field, link.value); err != nil {
			return err
		}
	}

	return nil
}

// validTimezone requires an IANA zone; "Local" is refused because it depends on
// the host.
func validTimezone(zone string) error {
	if zone == "" {
		return fmt.Errorf("timezone is required")
	}

	if _, err := time.LoadLocation(zone); err != nil || zone == "Local" || len(zone) > 64 {
		return fmt.Errorf("timezone must be a zone such as Asia/Bishkek or UTC")
	}

	return nil
}

// ValidLink accepts an empty value or an absolute http(s) URL; these are
// rendered in pages, where a relative path or javascript: URL would be wrong or
// dangerous.
func ValidLink(field, value string) error {
	if value == "" {
		return nil
	}

	if len(value) > 512 {
		return fmt.Errorf("%s must be at most 512 characters", field)
	}

	parsed, err := url.Parse(value)
	if err != nil || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return fmt.Errorf("%s must be a full address starting with http:// or https://", field)
	}

	return nil
}
