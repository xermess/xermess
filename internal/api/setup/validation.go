package setup

import (
	"strings"

	"loginer/internal/api/validate"
)

// validate checks the form.
//
// The password has a length here, unlike the sign-in form: this is where one
// is chosen, and a super admin's password is the whole panel. Ten characters
// is the floor, and length is all that is asked for — a rule about symbols
// mostly produces one symbol on the end.
func (r *setupRequest) validate() error {
	r.Email = strings.ToLower(strings.TrimSpace(r.Email))
	r.FirstName = strings.TrimSpace(r.FirstName)
	r.LastName = strings.TrimSpace(r.LastName)

	return validate.Struct(r)
}
