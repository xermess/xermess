package social

import (
	"net/http"
	"strings"

	"xermess/internal/api/respond"
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

	provider.Name = text(r.Name, provider.Name)
	provider.ClientID = text(r.ClientID, provider.ClientID)
	provider.TeamID = text(r.TeamID, provider.TeamID)
	provider.KeyID = text(r.KeyID, provider.KeyID)
	provider.AuthorizeURL = text(r.AuthorizeURL, provider.AuthorizeURL)
	provider.TokenURL = text(r.TokenURL, provider.TokenURL)
	provider.UserInfoURL = text(r.UserInfoURL, provider.UserInfoURL)
	provider.TokenAuth = model.SocialTokenAuth(lower((*string)(r.TokenAuth), string(provider.TokenAuth)))

	provider.Enabled = flag(r.Enabled, provider.Enabled)
	provider.LinkVerifiedEmails = flag(r.LinkVerifiedEmails, provider.LinkVerifiedEmails)
	provider.AllowRegistration = flag(r.AllowRegistration, provider.AllowRegistration)

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

// text reads a setting a request may leave out: what was sent, trimmed, or
// the stored value when nothing was.
func text(sent *string, current string) string {
	if sent == nil {
		return current
	}

	return strings.TrimSpace(*sent)
}

// lower is text for the settings that are compared or put in an address
// rather than read.
func lower(sent *string, current string) string {
	if value := strings.ToLower(text(sent, current)); value != "" {
		return value
	}

	return strings.ToLower(current)
}

// flag reads an on-or-off setting a request may leave out. A plain bool
// cannot tell "false" from "not sent", and a caller that omits a switch
// should not throw it.
func flag(sent *bool, current bool) bool {
	if sent == nil {
		return current
	}

	return *sent
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
