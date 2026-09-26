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

// Organization is who this installation belongs to: what it is called, how
// users reach it, and the agreements they sign up to.
//
// There is one row. An installation is one organisation's — the users, the
// applications and the administrators all belong to it — so this is a record
// of settings rather than a list of things: the panel reads it and writes it
// back, and never makes another. The row is created by the migration, and by
// the store if it ever finds none.
//
// It is the organisation's half of what the server shows the outside world.
// Its terms and privacy links are published in the discovery document as
// op_tos_uri and op_policy_uri, and the sign-in pages fall back to them, and
// to its name, logo and support contact, for an application that gives none
// of its own. Only a super admin may change any of it.
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

	// SupportEmail and SupportPhone are how someone who cannot get in asks
	// for help. The sign-in pages show them, so an account locked out of
	// every application still has somewhere to turn.
	SupportEmail string `gorm:"size:255" json:"support_email"`
	SupportPhone string `gorm:"size:32" json:"support_phone"`

	// TermsURL and PrivacyURL are the agreements a user accepts by making an
	// account: the organisation's own, which an application's links replace
	// where it has them.
	TermsURL   string `gorm:"size:512" json:"terms_url"`
	PrivacyURL string `gorm:"size:512" json:"privacy_url"`
	// Timezone is the IANA zone the organisation keeps its calendar in,
	// "Asia/Bishkek". A user's account pages say dates in it, so the page the
	// server renders and the one the browser redraws name the same day.
	Timezone string `gorm:"size:64;not null;default:UTC" json:"timezone"`
}

// TableName pins the table name.
func (Organization) TableName() string {
	return "organizations"
}

// DefaultOrganization is the row a fresh installation starts with: named
// after the server, with nothing filled in that would be a guess, for whoever
// installed it to correct in the panel.
//
// The migration seeds this, and the store falls back to it, so what a new
// installation holds is written down once.
func DefaultOrganization() Organization {
	return Organization{Name: brand.Name, Slug: brand.Slug, Timezone: "UTC"}
}

// slugPattern is what a short name may look like: lower case letters, numbers
// and dashes, starting and ending with one of the first two.
var slugPattern = regexp.MustCompile(`^[a-z0-9]([a-z0-9-]*[a-z0-9])?$`)

// phonePattern is a number someone can dial: digits, in any of the ways
// people write them apart, optionally in international form. It is deliberately
// loose — numbering plans differ by country, and a number refused here is a
// number the organisation cannot publish.
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
		if err := validLink(link.field, link.value); err != nil {
			return err
		}
	}

	return nil
}

// validTimezone holds a zone the IANA database names. "Local" is refused
// although Go loads it: it is whatever zone the server happens to run in,
// which is not a setting anybody chose.
func validTimezone(zone string) error {
	if zone == "" {
		return fmt.Errorf("timezone is required")
	}

	if _, err := time.LoadLocation(zone); err != nil || zone == "Local" || len(zone) > 64 {
		return fmt.Errorf("timezone must be a zone such as Asia/Bishkek or UTC")
	}

	return nil
}

// validLink holds a link to an absolute http(s) address: these are put in an
// img tag and in anchors on pages served by this server and by the
// applications reading the discovery document, where a relative path would
// point at whichever of them rendered it and a javascript: URL would be a way
// in. An empty value is no link at all, which is allowed.
func validLink(field, value string) error {
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
