package mail

import (
	"net/http"

	"xermess/internal/api/respond"
	"xermess/internal/api/validate"
	"xermess/internal/jose"
	"xermess/internal/mail"
	"xermess/internal/model"
)

// applyTo copies what a request sent onto the settings and returns what it
// changed, in the order listed here, so the activity log can say what an
// administrator touched.
//
// The rules are the model's: what a host name or a from address may be is the
// same question wherever it is asked. Nothing is written when any of them is
// broken, since the record is only saved once this returns.
func (r *settingsRequest) applyTo(settings *model.MailSettings, sealer *jose.Sealer) ([]string, error) {
	updated := *settings
	updated.Enabled = validate.Flag(r.Enabled, settings.Enabled)
	updated.Host = validate.Lower(r.Host, settings.Host)
	updated.Port = validate.Number(r.Port, settings.Port)
	updated.Encryption = model.MailEncryption(validate.Lower(r.Encryption, string(settings.Encryption)))
	updated.Username = validate.Text(r.Username, settings.Username)
	updated.FromAddress = validate.Lower(r.FromAddress, settings.FromAddress)
	updated.FromName = validate.Text(r.FromName, settings.FromName)

	// A password is only ever written, never read back, so the request says
	// one of three things: nothing, a new one, or an empty string meaning
	// the server takes none.
	passwordChanged := false
	if r.Password != nil {
		password := *r.Password

		switch {
		case password == "":
			passwordChanged = len(settings.Password) > 0
			updated.Password = nil
		default:
			sealed, err := sealer.SealBytes([]byte(password))
			if err != nil {
				return nil, err
			}
			passwordChanged = true
			updated.Password = sealed
		}
	}

	if err := updated.Validate(); err != nil {
		return nil, respond.Fault{Status: http.StatusBadRequest, Message: err.Error()}
	}

	changed := []string{}
	for _, field := range []struct {
		name  string
		moved bool
	}{
		{"enabled", updated.Enabled != settings.Enabled},
		{"host", updated.Host != settings.Host},
		{"port", updated.Port != settings.Port},
		{"encryption", updated.Encryption != settings.Encryption},
		{"username", updated.Username != settings.Username},
		{"password", passwordChanged},
		{"from_address", updated.FromAddress != settings.FromAddress},
		{"from_name", updated.FromName != settings.FromName},
	} {
		if field.moved {
			changed = append(changed, field.name)
		}
	}

	*settings = updated

	return changed, nil
}

// settings is what a test message is sent with: the stored record with
// whatever the form holds laid over it, and the password unsealed.
//
// The form's password is the one typed just now. Where none was typed the
// stored one is unsealed instead, so testing a saved server again does not
// mean typing its password each time.
func (r *testRequest) settings(stored *model.MailSettings, sealer *jose.Sealer) (mail.Settings, error) {
	trying := *stored

	// A test sends whatever is on the form, whether or not sending is
	// switched on: turning it on after the test has worked is the order
	// somebody would do this in.
	trying.Enabled = true
	trying.Host = validate.Lower(r.Host, stored.Host)
	trying.Port = validate.Number(r.Port, stored.Port)
	trying.Encryption = model.MailEncryption(validate.Lower(r.Encryption, string(stored.Encryption)))
	trying.Username = validate.Text(r.Username, stored.Username)
	trying.FromAddress = validate.Lower(r.FromAddress, stored.FromAddress)
	trying.FromName = validate.Text(r.FromName, stored.FromName)

	if err := trying.Validate(); err != nil {
		return mail.Settings{}, respond.Fault{Status: http.StatusBadRequest, Message: err.Error()}
	}

	if r.Password != nil {
		return mail.Of(trying, *r.Password), nil
	}

	password := ""
	if len(stored.Password) > 0 {
		plain, err := sealer.OpenBytes(stored.Password)
		if err != nil {
			return mail.Settings{}, err
		}
		password = string(plain)
	}

	return mail.Of(trying, password), nil
}

// validate holds the words of an email to the keys there are. A key that is
// not one of model.MailMessageSpecs' is refused: this page holds a handful of
// a language's text, and writing anything else through it would be a bug
// rather than an old file being imported.
func (r *contentRequest) validate() error {
	for key, text := range r.Messages {
		if !model.IsMailMessageKey(key) {
			return contentKeyUnknown.With("key", key)
		}

		if len(text) > model.MaxMessageLength {
			return respond.Fault{
				Status:  http.StatusBadRequest,
				Message: key + " is longer than a message may be",
			}
		}
	}

	return nil
}
