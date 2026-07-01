// Package mailer sends email through a user-configured provider: SMTP, or an
// HTTP API (Resend / Brevo). HTTP providers work over HTTPS, so they send even
// in networks that block outbound SMTP ports. The user owns the account and is
// the accountable sender; credentials live only in process config.
package mailer

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// Mailer sends one HTML email. Implemented by the SMTP and HTTP providers.
type Mailer interface {
	Send(ctx context.Context, from, to, subject, htmlBody string) error
}

// Config selects and configures a provider.
type Config struct {
	Provider string // "smtp" | "resend" | "brevo"
	APIKey   string // for resend/brevo

	Host     string // smtp
	Port     int
	Username string
	Password string
}

// New builds the configured Mailer, or (nil, nil) when nothing is configured
// (campaigns then report disabled).
func New(cfg Config) (Mailer, error) {
	switch strings.ToLower(strings.TrimSpace(cfg.Provider)) {
	case "resend":
		if cfg.APIKey == "" {
			return nil, nil
		}
		return &httpMailer{client: httpClient(), provider: providerResend, apiKey: cfg.APIKey}, nil
	case "brevo":
		if cfg.APIKey == "" {
			return nil, nil
		}
		return &httpMailer{client: httpClient(), provider: providerBrevo, apiKey: cfg.APIKey}, nil
	default: // smtp
		sm, err := NewSMTP(Config{Host: cfg.Host, Port: cfg.Port, Username: cfg.Username, Password: cfg.Password})
		if err != nil {
			return nil, err
		}
		if sm == nil {
			return nil, nil // avoid a non-nil interface wrapping a nil pointer
		}
		return sm, nil
	}
}

func httpClient() *http.Client { return &http.Client{Timeout: 30 * time.Second} }

// splitAddress parses "Name <email>" into its parts. When there is no name it
// returns the address for both.
func splitAddress(from string) (name, email string) {
	from = strings.TrimSpace(from)
	if i := strings.LastIndex(from, "<"); i >= 0 && strings.HasSuffix(from, ">") {
		return strings.TrimSpace(from[:i]), strings.TrimSpace(from[i+1 : len(from)-1])
	}
	return "", from
}

// httpError formats a non-2xx API response.
func httpError(provider string, status int, body string) error {
	body = strings.TrimSpace(body)
	if len(body) > 300 {
		body = body[:300]
	}
	return fmt.Errorf("mailer(%s): status %d: %s", provider, status, body)
}
