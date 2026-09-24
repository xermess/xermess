package auth

// loginRequest is the body of POST /admin/auth/login.
//
// The password has no rule beyond being there: what a password must look like
// is the account's business, and a sign-in form that explains the policy is a
// sign-in form that helps someone guessing.
type loginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

// profileRequest is the body of PATCH /admin/me: the administrator's own
// name and address. The current password is only needed when the address
// changes, since the address is what they sign in with.
type profileRequest struct {
	FirstName       string `json:"first_name" validate:"required,max=100"`
	LastName        string `json:"last_name" validate:"max=100"`
	Email           string `json:"email" validate:"required,email,max=255"`
	CurrentPassword string `json:"current_password"`
}

// passwordRequest is the body of POST /admin/me/password.
type passwordRequest struct {
	CurrentPassword string `json:"current_password" validate:"required"`
	NewPassword     string `json:"new_password" validate:"required"`
}

// codeRequest is a code from an authenticator app, or a recovery code.
type codeRequest struct {
	Code string `json:"code" validate:"required,max=32"`
}
