package social

import (
	"net/http"
	"strings"

	"loginer/internal/api/respond"
	"loginer/internal/api/validate"
	"loginer/internal/jose"
	"loginer/internal/model"
)

// applyTo checks the request and copies it onto a provider. Kind and slug are
// read only when `creating`, since they form the registered callback address.
// Everything else changes only where the request mentions it.
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
	provider.UserinfoURL = validate.Text(r.UserInfoURL, provider.UserinfoURL)
	provider.TokenAuth = model.SocialTokenAuth(lower((*string)(r.TokenAuth), string(provider.TokenAuth)))

	provider.IsEnabled = validate.Flag(r.IsEnabled, provider.IsEnabled)
	provider.LinkVerifiedEmails = validate.Flag(r.LinkVerifiedEmails, provider.LinkVerifiedEmails)
	provider.AllowRegistration = validate.Flag(r.AllowRegistration, provider.AllowRegistration)

	if r.Scopes != nil {
		provider.Scopes = cleanScopes(*r.Scopes)
	}

	if r.Position != nil {
		provider.Position = *r.Position
	}

	// A sent secret replaces the stored one; an absent one keeps it; "" clears
	// it.
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

// lower is validate.Lower except that an empty value falls back to `current`,
// the field's default, rather than clearing it.
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
