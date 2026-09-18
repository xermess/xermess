package social

// targetType is what these rows are called in the activity log.
const targetType = "social_provider"

// providerRequest is what a create or an update sends.
//
// Every settable field is a pointer, because this is a PATCH: a request that
// does not mention a setting leaves it as it is, the way the organisation's
// endpoint works. It matters more here than anywhere else — a caller turning
// a provider off sends `{"enabled": false}`, and that must not quietly empty
// the endpoints and keys around it.
//
// The secrets are pointers for a second reason: they cannot be sent back, so
// only a request carrying a new value replaces one.
type providerRequest struct {
	// Kind and Slug are read when a provider is registered and never again:
	// the slug is in the address registered with the provider, and the kind
	// decides which addresses those are.
	Kind string `json:"kind"`
	Slug string `json:"slug"`

	Name     *string `json:"name"`
	ClientID *string `json:"client_id"`

	ClientSecret *string `json:"client_secret"`
	PrivateKey   *string `json:"private_key"`

	TeamID *string `json:"team_id"`
	KeyID  *string `json:"key_id"`

	Scopes *[]string `json:"scopes"`

	// TokenAuth is how the secret is presented at the token endpoint. An
	// empty string leaves it to the kind's own default.
	TokenAuth *string `json:"token_auth"`

	AuthorizeURL *string `json:"authorize_url"`
	TokenURL     *string `json:"token_url"`
	UserInfoURL  *string `json:"userinfo_url"`

	Enabled            *bool `json:"enabled"`
	LinkVerifiedEmails *bool `json:"link_verified_emails"`
	AllowRegistration  *bool `json:"allow_registration"`

	// Position is where the button sits. Left out, a new provider goes last
	// and an existing one stays where it is.
	Position *int `json:"position"`
}
