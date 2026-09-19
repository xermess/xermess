package roles

import (
	"xermess/internal/api/validate"
	"xermess/internal/model"
)

// The rule a role's name keeps, registered here so the rule and the thing it
// describes stay together.
func init() {
	validate.Register("rolename", model.RoleNamePattern.MatchString)
}

// validate checks the role's own columns. Whether the roles it inherits
// exist, and whether inheriting them would make a loop, needs the database
// and is checked by the handler.
func (r *roleRequest) validate() error {
	r.clean()

	return validate.Struct(r)
}
