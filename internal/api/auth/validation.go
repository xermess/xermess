package auth

import (
	"strings"
	"unicode/utf8"

	"loginer/internal/api/validate"
	"loginer/internal/model"
)

// validate checks the sign-in form before anything is looked up.
func (r *loginRequest) validate() error {
	r.Username = strings.TrimSpace(r.Username)

	return validate.Struct(r)
}

// validate tidies the profile form and checks it. The address is kept
// trimmed and lower case, the way every address is, since it is the
// username too.
func (r *profileRequest) validate() error {
	r.FirstName = strings.TrimSpace(r.FirstName)
	r.LastName = strings.TrimSpace(r.LastName)
	r.Email = model.NormalizeEmail(r.Email)
	r.AvatarURL = strings.TrimSpace(r.AvatarURL)

	if err := validate.Struct(r); err != nil {
		return err
	}

	// The picture is put in an img tag on every page of the panel, so it is
	// held to the rule every such link is: a full http(s) address, or none.
	if model.ValidLink("avatar_url", r.AvatarURL) != nil {
		return avatarInvalid.With()
	}

	return nil
}

// validate checks a password change. The new password is taken as it was
// typed; only its length is a rule here.
func (r *passwordRequest) validate() error {
	if err := validate.Struct(r); err != nil {
		return err
	}

	if utf8.RuneCountInString(r.NewPassword) < model.MinAdminPasswordLength {
		return passwordTooShort.With("min", model.MinAdminPasswordLength)
	}

	return nil
}
