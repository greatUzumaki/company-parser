// Package campaign sends an email campaign to the addresses collected in a
// search. It is deliberately conservative: recipients are validated and
// deduped, sends are rate-limited, and a dry run lets the user preview without
// sending. The caller must have obtained consent — this tool does not.
package campaign

import (
	"context"
	"errors"
	"fmt"
	"html"
	"net/mail"
	"strings"
	"time"

	mailerpkg "github.com/parse-companies/backend/internal/mailer"
)

// ErrDisabled means no SMTP is configured.
var ErrDisabled = errors.New("campaign: email sending is not configured")

// Recipient is one target of a campaign. Name personalizes {{name}}.
type Recipient struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}

// Event is one message in a streaming campaign run.
type Event struct {
	Type    string `json:"type"` // start | sent | failed | done | error
	Email   string `json:"email,omitempty"`
	Name    string `json:"name,omitempty"`
	Sent    int    `json:"sent,omitempty"`
	Failed  int    `json:"failed,omitempty"`
	Total   int    `json:"total,omitempty"`
	DryRun  bool   `json:"dryRun,omitempty"`
	Message string `json:"message,omitempty"`
}

// Service runs campaigns.
type Service struct {
	mailer mailerpkg.Mailer
	from   string
	delay  time.Duration
	max    int
}

// New wires the mailer and sending policy. A nil mailer or empty from disables
// campaigns (Enabled reports false).
func New(m mailerpkg.Mailer, from string, delayMS, max int) *Service {
	return &Service{
		mailer: m,
		from:   from,
		delay:  time.Duration(delayMS) * time.Millisecond,
		max:    max,
	}
}

// Enabled reports whether campaigns can be sent.
func (s *Service) Enabled() bool {
	return s.mailer != nil && s.from != ""
}

// Send delivers a campaign to every valid, unique email in the given recipient
// list. When dryRun is true it streams progress without contacting SMTP.
func (s *Service) Send(ctx context.Context, subject, body string, in []Recipient, dryRun bool, emit func(Event) error) error {
	if !s.Enabled() {
		return ErrDisabled
	}
	if strings.TrimSpace(subject) == "" || strings.TrimSpace(body) == "" {
		return fmt.Errorf("campaign: subject and body are required")
	}

	recipients := collect(in, s.max)

	if err := emit(Event{Type: "start", Total: len(recipients), DryRun: dryRun}); err != nil {
		return err
	}

	sent, failed := 0, 0
	for i, r := range recipients {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		// Throttle between real sends to avoid tripping spam controls.
		if i > 0 && !dryRun {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(s.delay):
			}
		}

		subj := personalize(subject, r.Name, false)
		htmlBody := personalize(body, r.Name, true)

		if dryRun {
			sent++
			if err := emit(Event{Type: "sent", Email: r.Email, Name: r.Name, Sent: sent, Failed: failed, DryRun: true}); err != nil {
				return err
			}
			continue
		}

		if err := s.mailer.Send(ctx, s.from, r.Email, subj, htmlBody); err != nil {
			failed++
			if e := emit(Event{Type: "failed", Email: r.Email, Name: r.Name, Sent: sent, Failed: failed, Message: "send failed"}); e != nil {
				return e
			}
			continue
		}
		sent++
		if err := emit(Event{Type: "sent", Email: r.Email, Name: r.Name, Sent: sent, Failed: failed}); err != nil {
			return err
		}
	}

	return emit(Event{Type: "done", Sent: sent, Failed: failed, Total: len(recipients), DryRun: dryRun})
}

// collect returns unique, syntactically valid recipients, capped at max.
func collect(in []Recipient, max int) []Recipient {
	seen := make(map[string]struct{})
	var out []Recipient
	for _, r := range in {
		email := strings.TrimSpace(strings.ToLower(r.Email))
		if email == "" {
			continue
		}
		if _, err := mail.ParseAddress(email); err != nil {
			continue
		}
		if _, dup := seen[email]; dup {
			continue
		}
		seen[email] = struct{}{}
		out = append(out, Recipient{Email: email, Name: r.Name})
		if max > 0 && len(out) >= max {
			break
		}
	}
	return out
}

// personalize substitutes {{name}} with the recipient's name (or "there").
// For HTML bodies the name is escaped and newlines become <br>, since the name
// originates from publicly editable data and must not inject markup.
func personalize(template, name string, asHTML bool) string {
	display := strings.TrimSpace(name)
	if display == "" {
		display = "there"
	}
	if asHTML {
		out := strings.ReplaceAll(template, "{{name}}", html.EscapeString(display))
		return strings.ReplaceAll(out, "\n", "<br>\n")
	}
	return strings.ReplaceAll(template, "{{name}}", display)
}
