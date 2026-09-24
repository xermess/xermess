// Package mail sends the emails this server sends: a password reset link, a
// link that confirms an address, a one-time code.
//
// Which server they go through is not read from the configuration at startup
// but asked for per message (Settings), because the mail settings are the
// panel's after the first start: a password corrected on the Mail page takes
// effect on the next email rather than the next restart. With sending turned
// off, a message is written to the log instead, so a developer can follow a
// reset link without a mail server.
package mail

import (
	"context"
	"crypto/tls"
	"fmt"
	"log/slog"
	"mime"
	"net"
	"net/mail"
	"net/smtp"
	"strconv"
	"strings"
	"time"
)

// Message is one plain-text email.
type Message struct {
	To      string
	Subject string
	Body    string
}

// Sender sends a message. The provider holds one; tests hand in their own to
// read what would have been sent.
type Sender interface {
	Send(ctx context.Context, msg Message) error
}

// Encryption is how the connection to the mail server is protected. The
// values are model.MailEncryption's, which is where they are described; the
// package does not import the models for three strings.
type Encryption string

const (
	StartTLS Encryption = "starttls"
	TLS      Encryption = "tls"
	None     Encryption = "none"
)

// Settings is a mail server, as sending one message needs it.
type Settings struct {
	// Enabled is whether anything is sent at all. Off, the message is
	// logged.
	Enabled bool

	Host       string
	Port       int
	Encryption Encryption
	Username   string
	// Password is the plain one: whoever hands these over has unsealed it.
	Password string

	FromAddress string
	FromName    string
}

// Address is the host and port to connect to.
func (s Settings) Address() string {
	return net.JoinHostPort(s.Host, strconv.Itoa(s.Port))
}

// From is the address messages are sent from, with the name a mail client
// shows in front of it.
func (s Settings) From() *mail.Address {
	return &mail.Address{Name: s.FromName, Address: s.FromAddress}
}

// Source gives the settings to send the next message with. It is asked once
// per message, so a change in the panel is picked up without anything being
// told about it.
type Source func(ctx context.Context) (Settings, error)

// New returns the sender the settings of the moment ask for.
func New(source Source, log *slog.Logger) Sender {
	return &Mailer{source: source, log: log}
}

// Mailer sends through whichever server its source names when a message goes
// out, and logs the message when none is configured.
type Mailer struct {
	source Source
	log    *slog.Logger
}

// Send sends one message.
func (m *Mailer) Send(ctx context.Context, msg Message) error {
	settings, err := m.source(ctx)
	if err != nil {
		return fmt.Errorf("mail: read the mail settings: %w", err)
	}

	if !settings.Enabled || settings.Host == "" {
		// Body included: this is only reached where no mail server is
		// configured, which is to say while developing.
		m.log.Info("email not sent: sending is off on the Mail page",
			"to", msg.To, "subject", msg.Subject, "body", msg.Body)

		return nil
	}

	return Deliver(ctx, settings, msg)
}

// Deliver hands one message to one mail server. It is what Send does once it
// knows where to send, and what the panel's "send a test email" calls with
// the settings on the form — which may not be the stored ones, since the
// point of the test is to try them before they are saved.
func Deliver(ctx context.Context, settings Settings, msg Message) error {
	from := settings.From()
	if _, err := mail.ParseAddress(from.Address); err != nil {
		return fmt.Errorf("mail: the address messages are sent from: %w", err)
	}

	to, err := mail.ParseAddress(msg.To)
	if err != nil {
		return fmt.Errorf("mail: recipient: %w", err)
	}

	// net/smtp has no context, so the whole exchange runs beside the caller
	// and the context cancels the waiting rather than the sending.
	done := make(chan error, 1)
	go func() {
		done <- deliver(settings, from, to, compose(from, to, msg))
	}()

	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

// dialTimeout is how long connecting to the mail server is given. Without it
// a host that accepts connections and says nothing holds the send — and, for
// a verification email, the request that asked for it — until the server's
// own write timeout.
const dialTimeout = 15 * time.Second

// deliver opens the connection the settings ask for and posts the message.
//
// smtp.SendMail would do this in a line, but only one of the three ways: it
// connects in the clear and upgrades if the server offers it. A server on 465
// expects TLS from the first byte and answers nothing to a plaintext hello,
// so the connection is made here and handed to smtp.NewClient.
func deliver(settings Settings, from, to *mail.Address, body []byte) error {
	conn, err := dial(settings)
	if err != nil {
		return fmt.Errorf("mail: connect to %s: %w", settings.Address(), err)
	}

	client, err := smtp.NewClient(conn, settings.Host)
	if err != nil {
		conn.Close()
		return fmt.Errorf("mail: %s: %w", settings.Address(), err)
	}
	defer client.Close()

	if settings.Encryption == StartTLS {
		if ok, _ := client.Extension("STARTTLS"); !ok {
			return fmt.Errorf("mail: %s does not offer STARTTLS", settings.Address())
		}
		if err := client.StartTLS(&tls.Config{ServerName: settings.Host}); err != nil {
			return fmt.Errorf("mail: STARTTLS: %w", err)
		}
	}

	// A server that takes no credentials is given none: a relay on the same
	// host is the usual reason, and offering it an empty password is not the
	// same as offering it nothing.
	if settings.Username != "" {
		auth := smtp.PlainAuth("", settings.Username, settings.Password, settings.Host)
		if err := client.Auth(auth); err != nil {
			return fmt.Errorf("mail: sign in to %s: %w", settings.Address(), err)
		}
	}

	if err := client.Mail(from.Address); err != nil {
		return fmt.Errorf("mail: from %s: %w", from.Address, err)
	}
	if err := client.Rcpt(to.Address); err != nil {
		return fmt.Errorf("mail: to %s: %w", to.Address, err)
	}

	writer, err := client.Data()
	if err != nil {
		return fmt.Errorf("mail: %w", err)
	}
	if _, err := writer.Write(body); err != nil {
		return fmt.Errorf("mail: %w", err)
	}
	if err := writer.Close(); err != nil {
		return fmt.Errorf("mail: %w", err)
	}

	return client.Quit()
}

// dial opens the connection, with TLS from the start where that is what the
// server expects.
func dial(settings Settings) (net.Conn, error) {
	dialer := &net.Dialer{Timeout: dialTimeout}

	if settings.Encryption == TLS {
		return tls.DialWithDialer(dialer, "tcp", settings.Address(), &tls.Config{ServerName: settings.Host})
	}

	return dialer.Dial("tcp", settings.Address())
}

// compose writes the message with the headers a mail client expects. The
// subject is encoded, since it may carry an application's name in any script.
func compose(from, to *mail.Address, msg Message) []byte {
	var b strings.Builder

	headers := [][2]string{
		{"From", from.String()},
		{"To", to.String()},
		{"Subject", mime.QEncoding.Encode("utf-8", msg.Subject)},
		{"Date", time.Now().Format(time.RFC1123Z)},
		{"MIME-Version", "1.0"},
		{"Content-Type", `text/plain; charset="utf-8"`},
		{"Content-Transfer-Encoding", "8bit"},
	}
	for _, header := range headers {
		fmt.Fprintf(&b, "%s: %s\r\n", header[0], header[1])
	}

	b.WriteString("\r\n")
	b.WriteString(strings.ReplaceAll(msg.Body, "\n", "\r\n"))

	return []byte(b.String())
}
