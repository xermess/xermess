package admins

import (
	"fmt"
	"net/http"
	"unicode/utf8"

	"xermess/internal/api/respond"
	"xermess/internal/api/validate"
	"xermess/internal/model"
)

// minPasswordLength is the shortest password an administrator may be given:
// the same floor the first super admin's password keeps, since an admin
// account opens the panel.
const minPasswordLength = model.MinAdminPasswordLength

// validate checks the form. `creating` says whether this makes a new
// administrator, which is the one time a password has to be given.
func (r *adminRequest) validate(creating bool) error {
	r.clean()

	if err := validate.Struct(r); err != nil {
		return err
	}

	switch {
	case r.Password == "" && creating:
		return badRequest("password is required")
	case r.Password != "" && utf8.RuneCountInString(r.Password) < minPasswordLength:
		return badRequest(fmt.Sprintf("password must be at least %d characters", minPasswordLength))
	}

	return nil
}

func badRequest(message string) error {
	return respond.Fault{Status: http.StatusBadRequest, Message: message}
}
