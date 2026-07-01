// Package mailer sends email through a user-configured SMTP server. The user
// owns the SMTP account and is the accountable sender; this package never
// stores or logs credentials beyond the process config.
package mailer

import (
	"context"
	"fmt"

	"github.com/wneessen/go-mail"
)

// Mailer sends one HTML email. Implemented by SMTPMailer; faked in tests.
type Mailer interface {
	Send(ctx context.Context, from, to, subject, htmlBody string) error
}

// Config holds SMTP connection settings.
type Config struct {
	Host     string
	Port     int
	Username string
	Password string
}

// SMTPMailer sends via SMTP with STARTTLS.
type SMTPMailer struct {
	client *mail.Client
}

// NewSMTP builds an SMTPMailer. Returns (nil, nil) when Host is empty so the
// caller can treat campaigns as disabled without an error.
func NewSMTP(cfg Config) (*SMTPMailer, error) {
	if cfg.Host == "" {
		return nil, nil
	}
	opts := []mail.Option{
		mail.WithPort(cfg.Port),
		mail.WithTLSPolicy(mail.TLSOpportunistic),
	}
	// Only authenticate when credentials are supplied (local relays like
	// MailHog accept anonymous mail).
	if cfg.Username != "" {
		opts = append(opts,
			mail.WithSMTPAuth(mail.SMTPAuthPlain),
			mail.WithUsername(cfg.Username),
			mail.WithPassword(cfg.Password),
		)
	}
	client, err := mail.NewClient(cfg.Host, opts...)
	if err != nil {
		return nil, fmt.Errorf("mailer: new client: %w", err)
	}
	return &SMTPMailer{client: client}, nil
}

// Send delivers one HTML message.
func (m *SMTPMailer) Send(ctx context.Context, from, to, subject, htmlBody string) error {
	msg := mail.NewMsg()
	if err := msg.From(from); err != nil {
		return fmt.Errorf("mailer: from %q: %w", from, err)
	}
	if err := msg.To(to); err != nil {
		return fmt.Errorf("mailer: to %q: %w", to, err)
	}
	msg.Subject(subject)
	msg.SetBodyString(mail.TypeTextHTML, htmlBody)
	if err := m.client.DialAndSendWithContext(ctx, msg); err != nil {
		return fmt.Errorf("mailer: send: %w", err)
	}
	return nil
}
