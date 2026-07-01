package campaign

import (
	"context"
	"errors"
	"strings"
	"testing"
)

type sentMail struct{ From, To, Subject, Body string }

type fakeMailer struct {
	sent []sentMail
	err  error
}

func (m *fakeMailer) Send(_ context.Context, from, to, subject, body string) error {
	if m.err != nil {
		return m.err
	}
	m.sent = append(m.sent, sentMail{from, to, subject, body})
	return nil
}

func sampleRecipients() []Recipient {
	return []Recipient{
		{Name: "Alpha", Email: "alpha@test.com"},
		{Name: "Beta", Email: "beta@test.com"},
		{Name: "Dup", Email: "ALPHA@test.com"}, // duplicate (case-insensitive) -> skipped
		{Name: "NoEmail", Email: ""},           // no email -> skipped
		{Name: "Bad", Email: "not-an-email"},   // invalid -> skipped
	}
}

func newSvc(m *fakeMailer) *Service {
	return New(m, "Sender <from@test.com>", 0, 100, nil)
}

func drain(t *testing.T, run func(emit func(Event) error) error) []Event {
	t.Helper()
	var events []Event
	if err := run(func(e Event) error { events = append(events, e); return nil }); err != nil {
		t.Fatalf("send: %v", err)
	}
	return events
}

func TestSendDedupsValidatesPersonalizes(t *testing.T) {
	m := &fakeMailer{}
	svc := newSvc(m)

	drain(t, func(emit func(Event) error) error {
		return svc.Send(context.Background(), "Hi {{name}}", "Hello {{name}},\nline2", sampleRecipients(), false, emit)
	})

	if len(m.sent) != 2 {
		t.Fatalf("sent %d mails, want 2 (dedup + invalid skipped)", len(m.sent))
	}
	// Personalization in subject (plain) and body (HTML with <br>).
	if m.sent[0].Subject != "Hi Alpha" {
		t.Errorf("subject = %q", m.sent[0].Subject)
	}
	if !strings.Contains(m.sent[0].Body, "Hello Alpha,<br>") {
		t.Errorf("body not personalized/newline-converted: %q", m.sent[0].Body)
	}
}

func TestSendEscapesNameInHTMLBody(t *testing.T) {
	m := &fakeMailer{}
	svc := newSvc(m)
	recipients := []Recipient{{Name: `<script>x</script>`, Email: "x@test.com"}}
	drain(t, func(emit func(Event) error) error {
		return svc.Send(context.Background(), "s", "Hi {{name}}", recipients, false, emit)
	})
	if strings.Contains(m.sent[0].Body, "<script>") {
		t.Errorf("name not escaped in HTML body: %q", m.sent[0].Body)
	}
	if !strings.Contains(m.sent[0].Body, "&lt;script&gt;") {
		t.Errorf("expected escaped name, got: %q", m.sent[0].Body)
	}
}

func TestDryRunDoesNotSend(t *testing.T) {
	m := &fakeMailer{}
	svc := newSvc(m)
	events := drain(t, func(emit func(Event) error) error {
		return svc.Send(context.Background(), "s", "b", sampleRecipients(), true, emit)
	})
	if len(m.sent) != 0 {
		t.Errorf("dry run sent %d mails, want 0", len(m.sent))
	}
	var done Event
	for _, e := range events {
		if e.Type == "done" {
			done = e
		}
	}
	if done.Sent != 2 || !done.DryRun {
		t.Errorf("done event = %+v", done)
	}
}

func TestSendReportsFailures(t *testing.T) {
	m := &fakeMailer{err: errors.New("smtp down")}
	svc := newSvc(m)
	events := drain(t, func(emit func(Event) error) error {
		return svc.Send(context.Background(), "s", "b", sampleRecipients(), false, emit)
	})
	var done Event
	for _, e := range events {
		if e.Type == "done" {
			done = e
		}
	}
	if done.Failed != 2 || done.Sent != 0 {
		t.Errorf("done event = %+v, want 2 failed", done)
	}
}

func TestDisabledWithoutMailer(t *testing.T) {
	svc := New(nil, "from@test.com", 0, 100, nil)
	if svc.Enabled() {
		t.Error("should be disabled with nil mailer")
	}
	err := svc.Send(context.Background(), "s", "b", sampleRecipients(), false, func(Event) error { return nil })
	if !errors.Is(err, ErrDisabled) {
		t.Errorf("err = %v, want ErrDisabled", err)
	}
}

func TestRequiresSubjectAndBody(t *testing.T) {
	svc := newSvc(&fakeMailer{})
	if err := svc.Send(context.Background(), "", "b", sampleRecipients(), false, func(Event) error { return nil }); err == nil {
		t.Error("expected error for empty subject")
	}
}
