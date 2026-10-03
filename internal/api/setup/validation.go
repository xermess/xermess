package setup

import (
	"strings"

	"loginer/internal/api/validate"
)

// validate checks the setup form. A super admin's password must be at least ten
// characters; only length is required.
func (r *setupRequest) validate() error {
	r.Email = strings.ToLower(strings.TrimSpace(r.Email))
	r.FirstName = strings.TrimSpace(r.FirstName)
	r.LastName = strings.TrimSpace(r.LastName)

	return validate.Struct(r)
}
