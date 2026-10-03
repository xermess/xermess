// Package mail sends this server's emails. The SMTP settings are read per
// message, so a change on the Mail page applies to the next email. With sending
// off, only the subject is logged.
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

// Encryption is how the SMTP connection is protected; values mirror
// model.MailEncryption.
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

// Source gives the settings for the next message, asked per message so panel
// changes apply at once.
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
		// Only the subject: the body holds a live reset link or sign-in code,
		// and the address is personal data.
		m.log.Info("email not sent: sending is off on the Mail page", "subject", msg.Subject)

		return nil
	}

	return Deliver(ctx, settings, msg)
}

// Deliver sends one message with the given settings: Send's last step, and the
// panel's test send with unsaved settings.
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

// dialTimeout bounds connecting, so a silent host cannot hold the request that
// is sending.
const dialTimeout = 15 * time.Second

// sendTimeout bounds the whole SMTP exchange, so a relay that stalls after
// connecting cannot hold a goroutine forever.
const sendTimeout = 60 * time.Second

// deliver connects itself rather than using smtp.SendMail, which cannot speak
// implicit TLS on port 465.
func deliver(settings Settings, from, to *mail.Address, body []byte) error {
	conn, err := dial(settings)
	if err != nil {
		return fmt.Errorf("mail: connect to %s: %w", settings.Address(), err)
	}

	// Every read and write that follows shares one deadline, so no step of the
	// exchange can hang on a server that goes quiet.
	_ = conn.SetDeadline(time.Now().Add(sendTimeout))

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

	// No credentials configured means no AUTH at all, which is not the same as
	// an empty password.
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
