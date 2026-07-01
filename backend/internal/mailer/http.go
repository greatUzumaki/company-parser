package mailer

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type provider int

const (
	providerResend provider = iota
	providerBrevo
)

// httpMailer sends via a provider's HTTPS REST API (works where SMTP is blocked).
type httpMailer struct {
	client   *http.Client
	provider provider
	apiKey   string
}

func (m *httpMailer) Send(ctx context.Context, from, to, subject, htmlBody string) error {
	switch m.provider {
	case providerBrevo:
		return m.sendBrevo(ctx, from, to, subject, htmlBody)
	default:
		return m.sendResend(ctx, from, to, subject, htmlBody)
	}
}

// sendResend posts to https://api.resend.com/emails.
func (m *httpMailer) sendResend(ctx context.Context, from, to, subject, htmlBody string) error {
	payload := map[string]any{
		"from":    from,
		"to":      []string{to},
		"subject": subject,
		"html":    htmlBody,
	}
	return m.post(ctx, "resend", "https://api.resend.com/emails", payload, func(r *http.Request) {
		r.Header.Set("Authorization", "Bearer "+m.apiKey)
	})
}

// sendBrevo posts to https://api.brevo.com/v3/smtp/email.
func (m *httpMailer) sendBrevo(ctx context.Context, from, to, subject, htmlBody string) error {
	name, email := splitAddress(from)
	sender := map[string]string{"email": email}
	if name != "" {
		sender["name"] = name
	}
	payload := map[string]any{
		"sender":      sender,
		"to":          []map[string]string{{"email": to}},
		"subject":     subject,
		"htmlContent": htmlBody,
	}
	return m.post(ctx, "brevo", "https://api.brevo.com/v3/smtp/email", payload, func(r *http.Request) {
		r.Header.Set("api-key", m.apiKey)
	})
}

// post sends the JSON payload and treats any non-2xx status as an error.
func (m *httpMailer) post(ctx context.Context, name, url string, payload any, auth func(*http.Request)) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("mailer(%s): marshal: %w", name, err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("mailer(%s): build request: %w", name, err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	auth(req)

	resp, err := m.client.Do(req)
	if err != nil {
		return fmt.Errorf("mailer(%s): send: %w", name, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}
	b, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
	return httpError(name, resp.StatusCode, string(b))
}
