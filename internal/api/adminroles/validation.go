package adminroles

import (
	"net/http"

	"loginer/internal/api/respond"
	"loginer/internal/api/validate"
	"loginer/internal/model"
)

// The rule an admin role's name keeps: the same shape as a user role's.
func init() {
	validate.Register("adminrole", model.RoleNamePattern.MatchString)
}

// applyTo checks the request and copies it onto a role. The permissions are
// stored once each, in catalog order, and a name the catalog does not have is
// refused rather than stored to grant nothing.
func (r *roleRequest) applyTo(role *model.Role) error {
	r.clean()

	if err := validate.Struct(r); err != nil {
		return err
	}

	if r.Name == model.RoleSuperAdmin {
		return respond.Fault{Status: http.StatusConflict, Message: "super_admin is built in: choose another name"}
	}

	asked := make(map[string]bool, len(r.Permissions))
	for _, name := range r.Permissions {
		if !model.IsAdminPermission(name) {
			return respond.Fault{
				Status:  http.StatusBadRequest,
				Message: "permissions: " + name + " is not a permission",
			}
		}
		asked[name] = true
	}

	permissions := []string{}
	for _, name := range model.AdminPermissionNames() {
		if asked[name] {
			permissions = append(permissions, name)
		}
	}

	role.Name = r.Name
	role.Description = r.Description
	role.Permissions = permissions

	return nil
}
