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

// FromStore is the Source a running server sends with: the Mail page's row with
// its password unsealed. A password that cannot be unsealed is an error, so the
// log says which thing failed.
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
		Enabled:     stored.IsEnabled,
		Host:        stored.Host,
		Port:        stored.Port,
		Encryption:  Encryption(stored.Encryption),
		Username:    stored.Username,
		Password:    password,
		FromAddress: stored.FromAddress,
		FromName:    stored.FromName,
	}
}

// implicitTLSPort is the port that speaks TLS from the first byte.
const implicitTLSPort = 465

// Seed is the row a fresh installation starts with, from LOGINER_SMTP_*. Naming
// a host turns sending on. The password is returned beside the record because
// the caller seals it.
func Seed(cfg config.Mail) (model.MailSettings, string) {
	settings := model.DefaultMailSettings()
	settings.IsEnabled = cfg.Host != ""
	settings.Host = cfg.Host
	settings.Username = cfg.Username

	if cfg.Port > 0 {
		settings.Port = cfg.Port
	}
	if settings.Port == implicitTLSPort {
		settings.Encryption = model.MailTLS
	}

	// LOGINER_SMTP_FROM is "Name <address>"; an unparseable value is left for
	// the Mail page to fix.
	if parsed, err := mail.ParseAddress(cfg.From); err == nil {
		settings.FromAddress = parsed.Address
		settings.FromName = parsed.Name
	}

	return settings, cfg.Password
}
