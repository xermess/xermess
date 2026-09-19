package account

// loginRequest is the body of POST /account/login. Request is the handle of
// the sign-in under way, when there is one.
type loginRequest struct {
	Request  string `json:"request" validate:"max=64"`
	Email    string `json:"email" validate:"required,max=255"`
	Password string `json:"password" validate:"required,max=128"`
}

// registerRequest is the body of POST /account/register.
type registerRequest struct {
	Request     string `json:"request" validate:"required,max=64"`
	Email       string `json:"email" validate:"required,email,max=255"`
	Password    string `json:"password" validate:"required,max=72"`
	FirstName   string `json:"first_name" validate:"max=100"`
	LastName    string `json:"last_name" validate:"max=100"`
	AcceptTerms bool   `json:"accept_terms"`
}

// discoverRequest is the address "Sign in with SSO" looks a connection up by.
type discoverRequest struct {
	Email string `json:"email" validate:"required,email,max=255"`
}

// forgotRequest is the body of POST /account/forgot-password.
type forgotRequest struct {
	Request string `json:"request" validate:"max=64"`
	Email   string `json:"email" validate:"required,email,max=255"`
	// Language is the one the page was shown in, which the email is written
	// in. One that is not offered, or none, is the default.
	Language string `json:"language" validate:"max=16"`
}

// resetRequest is the body of POST /account/reset-password.
type resetRequest struct {
	Token    string `json:"token" validate:"required,max=64"`
	Password string `json:"password" validate:"required,max=72"`
}

// profileRequest is the body of PATCH /account/me.
type profileRequest struct {
	FirstName string `json:"first_name" validate:"max=100"`
	LastName  string `json:"last_name" validate:"max=100"`
}

// passwordRequest is the body of POST /account/password.
type passwordRequest struct {
	CurrentPassword string `json:"current_password" validate:"required,max=128"`
	NewPassword     string `json:"new_password" validate:"required,max=72"`
}
