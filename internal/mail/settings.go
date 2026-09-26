package mail

import (
	"context"
	"fmt"
	"net/mail"

	"loginer/internal/config"
	"loginer/internal/model"
)

// Store is the part of the store this package uses: the mail settings, read
// once per message.
type Store interface {
	MailSettings(ctx context.Context) (*model.MailSettings, error)
}

// Opener unseals what the secret key sealed — the stored password. It is
// jose.Sealer, named by what is needed of it so the tests can stand in.
type Opener interface {
	OpenBytes(sealed []byte) ([]byte, error)
}

// FromStore is the Source a running server sends with: the row on the Mail
// page, with its password unsealed for the one message about to go out.
//
// A password that cannot be unsealed is an error rather than an empty one:
// signing in to the mail server without it would fail anyway, and the log
// should say which of the two went wrong.
func FromStore(st Store, opener Opener) Source {
	return func(ctx context.Context) (Settings, error) {
		stored, err := st.MailSettings(ctx)
		if err != nil {
			return Settings{}, err
		}

		password := ""
		if len(stored.Password) > 0 {
			plain, err := opener.OpenBytes(stored.Password)
			if err != nil {
				return Settings{}, fmt.Errorf("unseal the mail password: %w", err)
			}
			password = string(plain)
		}

		return Of(*stored, password), nil
	}
}

// Of is a stored record as this package takes it, with the password already
// unsealed. The panel's test message uses it too, with what is on the form.
func Of(stored model.MailSettings, password string) Settings {
	return Settings{
		Enabled:     stored.Enabled,
		Host:        stored.Host,
		Port:        stored.Port,
		Encryption:  Encryption(stored.Encryption),
		Username:    stored.Username,
		Password:    password,
		FromAddress: stored.FromAddress,
		FromName:    stored.FromName,
	}
}

// implicitTLSPort is the port that expects TLS from the first byte. It is the
// one thing the configuration does not say, and the port is how every mail
// client has always guessed it.
const implicitTLSPort = 465

// Seed is the row a fresh installation starts with, from the configuration:
// LOGINER_SMTP_HOST and the rest, which are read once and then belong to the
// panel. A host named there turns sending on, since naming one is what an
// installation does to mean "send email".
//
// The password is not sealed here — the caller holds the secret key — so it
// is returned beside the record rather than in it.
func Seed(cfg config.Mail) (model.MailSettings, string) {
	settings := model.DefaultMailSettings()
	settings.Enabled = cfg.Host != ""
	settings.Host = cfg.Host
	settings.Username = cfg.Username

	if cfg.Port > 0 {
		settings.Port = cfg.Port
	}
	if settings.Port == implicitTLSPort {
		settings.Encryption = model.MailTLS
	}

	// LOGINER_SMTP_FROM is one string in the form a mail client shows,
	// "a name <no-reply@localhost>"; the panel asks for the two parts
	// apart. Anything that is not an address is left for the Mail page to
	// correct rather than stopping the server.
	if parsed, err := mail.ParseAddress(cfg.From); err == nil {
		settings.FromAddress = parsed.Address
		settings.FromName = parsed.Name
	}

	return settings, cfg.Password
}
