package social

import (
	"net/http"
	"strings"

	"xermess/internal/api/respond"
	"xermess/internal/api/validate"
	"xermess/internal/jose"
	"xermess/internal/model"
)

// applyTo checks the request and copies it onto a provider.
//
// `creating` says whether the kind and the slug are read from it: neither may
// change afterwards. The slug is half of the address registered with the
// provider, and the kind decides what the other half is — changing either
// would break every sign-in with it, silently, until somebody noticed.
//
// Everything else is copied only where the request mentioned it, so a change
// to one setting is a change to one setting.
func (r *providerRequest) applyTo(provider *model.SocialProvider, sealer *jose.Sealer, creating bool) error {
	if creating {
		kind := model.SocialKind(strings.ToLower(strings.TrimSpace(r.Kind)))
		if _, known := model.SocialSpecFor(kind); !known {
			return badRequest("kind must be one of: " + strings.Join(kindNames(), ", "))
		}

		provider.Kind = kind
		provider.Slug = lower(&r.Slug, string(kind))
	}

	provider.Name = validate.Text(r.Name, provider.Name)
	provider.ClientID = validate.Text(r.ClientID, provider.ClientID)
	provider.TeamID = validate.Text(r.TeamID, provider.TeamID)
	provider.KeyID = validate.Text(r.KeyID, provider.KeyID)
	provider.AuthorizeURL = validate.Text(r.AuthorizeURL, provider.AuthorizeURL)
	provider.TokenURL = validate.Text(r.TokenURL, provider.TokenURL)
	provider.UserInfoURL = validate.Text(r.UserInfoURL, provider.UserInfoURL)
	provider.TokenAuth = model.SocialTokenAuth(lower((*string)(r.TokenAuth), string(provider.TokenAuth)))

	provider.Enabled = validate.Flag(r.Enabled, provider.Enabled)
	provider.LinkVerifiedEmails = validate.Flag(r.LinkVerifiedEmails, provider.LinkVerifiedEmails)
	provider.AllowRegistration = validate.Flag(r.AllowRegistration, provider.AllowRegistration)

	if r.Scopes != nil {
		provider.Scopes = cleanScopes(*r.Scopes)
	}

	if r.Position != nil {
		provider.Position = *r.Position
	}

	// A secret that was sent replaces the stored one; one that was not leaves
	// it alone. An empty string is how a secret is cleared, for a provider
	// that is changing kind of credential.
	if err := seal(sealer, r.ClientSecret, &provider.ClientSecret); err != nil {
		return err
	}
	if err := seal(sealer, r.PrivateKey, &provider.PrivateKey); err != nil {
		return err
	}

	// What a provider of this kind has to carry is the model's to say.
	if err := provider.Validate(); err != nil {
		return badRequest(err.Error())
	}

	return nil
}

// lower is validate.Lower with one difference, which is why it is here: a
// value sent empty falls back to `current` rather than clearing the field.
//
// Both fields it reads have a standing default — the slug is the kind's name,
// the token auth the kind's own — and both are passed in as `current`. Sending
// either as "" means "use the default", not "have none": a provider with no
// slug has no address for users to come back on.
func lower(sent *string, current string) string {
	if value := strings.ToLower(validate.Text(sent, current)); value != "" {
		return value
	}

	return strings.ToLower(current)
}

// cleanScopes drops the empty ones and the spaces around each.
func cleanScopes(sent []string) model.StringList {
	scopes := make(model.StringList, 0, len(sent))
	for _, scope := range sent {
		if scope = strings.TrimSpace(scope); scope != "" {
			scopes = append(scopes, scope)
		}
	}

	return scopes
}

// seal encrypts a secret that was sent, and leaves the stored one alone when
// none was.
func seal(sealer *jose.Sealer, sent *string, stored *[]byte) error {
	if sent == nil {
		return nil
	}

	value := strings.TrimSpace(*sent)
	if value == "" {
		*stored = nil
		return nil
	}

	sealed, err := sealer.SealBytes([]byte(value))
	if err != nil {
		return err
	}

	*stored = sealed

	return nil
}

func kindNames() []string {
	names := make([]string, 0, len(model.SocialSpecs))
	for _, spec := range model.SocialSpecs {
		names = append(names, string(spec.Kind))
	}

	return names
}

func badRequest(message string) error {
	return respond.Fault{Status: http.StatusBadRequest, Message: message}
}
