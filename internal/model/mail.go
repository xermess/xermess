package model

import (
	"fmt"
	"net/mail"
	"strings"

	"loginer/internal/brand"
)

// MailSettings is how this installation sends email: the server it hands a
// message to, and the address that message comes from.
//
// There is one row, like the organisation's and admin_security's. It starts
// as the configuration says (LOGINER_SMTP_*) and is a super admin's to change
// afterwards, and it is read when a message is sent rather than held in a
// field from startup — so correcting a password that was typed wrong takes
// effect on the next email instead of the next restart.
//
// Password is sealed with the server's secret key, like a social provider's
// client secret: the database should never hold it in the clear, and it never
// leaves the server. Nothing reads it back; a password that is forgotten is
// typed again.
type MailSettings struct {
	Base

	// Enabled is whether messages are sent at all. Off, each one is written
	// to the log instead — which is how a developer follows a reset link
	// without a mail server, and how an installation that has not been given
	// one yet behaves rather than failing every sign-up.
	Enabled bool `gorm:"not null" json:"enabled"`

	Host string `gorm:"size:255" json:"host"`
	Port int    `gorm:"not null" json:"port"`

	// Encryption is how the connection is protected; see MailEncryption.
	Encryption MailEncryption `gorm:"type:varchar(16);not null" json:"encryption"`

	// Username is empty for a server that takes no credentials — a relay on
	// the same host, usually. Password is then ignored.
	Username string `gorm:"size:255" json:"username"`
	Password []byte `gorm:"type:bytea" json:"-"`

	// FromAddress is what every message is sent from, and FromName what a
	// mail client shows instead of it. They are apart rather than one
	// "Name <address>" string because the panel asks for them apart, and
	// because an address has to be checked and a name does not.
	FromAddress string `gorm:"size:255" json:"from_address"`
	FromName    string `gorm:"size:100" json:"from_name"`
}

// TableName pins the table name.
func (MailSettings) TableName() string {
	return "mail_settings"
}

// MailEncryption is how the connection to the mail server is protected.
type MailEncryption string

const (
	// MailStartTLS connects in the clear and upgrades with STARTTLS. It is
	// what port 587 expects, and what almost every provider asks for.
	MailStartTLS MailEncryption = "starttls"

	// MailTLS connects with TLS from the first byte, which port 465 expects.
	MailTLS MailEncryption = "tls"

	// MailNone is no encryption at all: a relay on localhost, and nothing
	// that crosses a network. The panel says so.
	MailNone MailEncryption = "none"
)

// MailEncryptions is the catalog, in the order the panel offers them.
var MailEncryptions = []MailEncryption{MailStartTLS, MailTLS, MailNone}

// IsMailEncryption reports whether a value is one of them.
func IsMailEncryption(value MailEncryption) bool {
	for _, known := range MailEncryptions {
		if known == value {
			return true
		}
	}

	return false
}

// DefaultMailSettings is what an installation starts with when the
// configuration names no mail server: nothing sent, and the port and form a
// server is most likely to want once one is given.
//
// Off is the default because the alternative is worse than not sending: a
// server configured with a host that does not answer makes every sign-up wait
// for a connection to time out, where one that logs its messages lets the
// installation be tried.
func DefaultMailSettings() MailSettings {
	return MailSettings{
		Enabled:     false,
		Port:        587,
		Encryption:  MailStartTLS,
		FromAddress: "no-reply@localhost",
		FromName:    brand.Name,
	}
}

// Validate reports the first thing wrong with the settings.
//
// A mail server that is off is checked as loosely as it is used: nothing is
// sent, so a half-filled form is somebody's work in progress rather than a
// mistake. Turning it on is what makes the host and the address required.
func (m MailSettings) Validate() error {
	if !IsMailEncryption(m.Encryption) {
		return fmt.Errorf("encryption must be one of: %s", joinEncryptions())
	}

	if m.Port < 1 || m.Port > 65535 {
		return fmt.Errorf("port must be between 1 and 65535")
	}

	if len(m.Host) > 255 {
		return fmt.Errorf("host must be at most 255 characters")
	}
	if strings.ContainsAny(m.Host, " /:") {
		return fmt.Errorf("host must be a host name on its own, such as smtp.example.com")
	}

	if len(m.Username) > 255 {
		return fmt.Errorf("username must be at most 255 characters")
	}
	if len(m.FromName) > 100 {
		return fmt.Errorf("from_name must be at most 100 characters")
	}

	if m.FromAddress != "" {
		if _, err := mail.ParseAddress(m.FromAddress); err != nil {
			return fmt.Errorf("from_address must be an email address")
		}
	}

	if !m.Enabled {
		return nil
	}

	switch {
	case m.Host == "":
		return fmt.Errorf("host is required to send email")
	case m.FromAddress == "":
		return fmt.Errorf("from_address is required to send email")
	}

	return nil
}

func joinEncryptions() string {
	names := make([]string, 0, len(MailEncryptions))
	for _, encryption := range MailEncryptions {
		names = append(names, string(encryption))
	}

	return strings.Join(names, ", ")
}

// MailMessageSpec is one email this server sends, for the panel to list: what
// it is called, when it goes out, and the keys its subject and body are
// written under.
//
// The text itself is not here. It is the sign-in pages' text, in the
// languages table, so a message goes out in the reader's language and is
// translated by whoever translates the rest — the Mail page edits those keys
// rather than holding a second copy of them.
type MailMessageSpec struct {
	Kind        string `json:"kind"`
	Label       string `json:"label"`
	Description string `json:"description"`

	// SubjectKey and BodyKey are the keys in the sign-in pages' catalog.
	SubjectKey string `json:"subject_key"`
	BodyKey    string `json:"body_key"`

	// Params are the {braces} the text may use, each with what it stands
	// for, so the panel can list them beside the editor and refuse one it
	// does not know.
	Params []MailParam `json:"params"`
}

// MailParam is one placeholder a message's text may use.
type MailParam struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// MailMessageSpecs is every email the server sends, in the order the panel
// lists them. A message added to the server is added here, or the Mail page
// will not know it exists.
var MailMessageSpecs = []MailMessageSpec{
	{
		Kind:        "reset",
		Label:       "Password reset",
		Description: "Sent when somebody asks for a new password, with the link that sets one.",
		SubjectKey:  "email.reset.subject",
		BodyKey:     "email.reset.body",
		Params: []MailParam{
			{Name: "app", Description: "The application being signed in to"},
			{Name: "email", Description: "The address the message went to"},
			{Name: "link", Description: "The reset link"},
			{Name: "minutes", Description: "How long the link works, in minutes"},
		},
	},
	{
		Kind:        "verify",
		Label:       "Confirm an address",
		Description: "Sent when a login flow requires a verified address and the account's is not.",
		SubjectKey:  "email.verify.subject",
		BodyKey:     "email.verify.body",
		Params: []MailParam{
			{Name: "app", Description: "The application being signed in to"},
			{Name: "email", Description: "The address the message went to"},
			{Name: "link", Description: "The verification link"},
			{Name: "hours", Description: "How long the link works, in hours"},
		},
	},
	{
		Kind:        "code",
		Label:       "One-time code",
		Description: "Sent by a login flow with the emailed code step, with the code to type back.",
		SubjectKey:  "email.code.subject",
		BodyKey:     "email.code.body",
		Params: []MailParam{
			{Name: "app", Description: "The application being signed in to"},
			{Name: "email", Description: "The address the message went to"},
			{Name: "code", Description: "The code itself"},
			{Name: "minutes", Description: "How long the code works, in minutes"},
		},
	},
}

// MailMessageKeys is every key the Mail page edits, in the order the messages
// are listed. It is what tells the page's reads and writes apart from the
// rest of a language's text.
func MailMessageKeys() []string {
	keys := make([]string, 0, len(MailMessageSpecs)*2)
	for _, spec := range MailMessageSpecs {
		keys = append(keys, spec.SubjectKey, spec.BodyKey)
	}

	return keys
}

// IsMailMessageKey reports whether a key is one of them, so a write to the
// Mail page cannot reach the rest of a language's text.
func IsMailMessageKey(key string) bool {
	for _, known := range MailMessageKeys() {
		if known == key {
			return true
		}
	}

	return false
}
