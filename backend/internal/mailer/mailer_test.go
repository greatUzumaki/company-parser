package mailer

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSplitAddress(t *testing.T) {
	tests := []struct{ in, wantName, wantEmail string }{
		{"Acme <hello@acme.com>", "Acme", "hello@acme.com"},
		{"hello@acme.com", "", "hello@acme.com"},
		{"  Bob  <b@x.io> ", "Bob", "b@x.io"},
	}
	for _, tt := range tests {
		name, email := splitAddress(tt.in)
		if name != tt.wantName || email != tt.wantEmail {
			t.Errorf("splitAddress(%q) = %q,%q want %q,%q", tt.in, name, email, tt.wantName, tt.wantEmail)
		}
	}
}

func TestPostSuccess(t *testing.T) {
	var gotAuth, gotBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		b := make([]byte, r.ContentLength)
		r.Body.Read(b)
		gotBody = string(b)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	m := &httpMailer{client: httpClient(), provider: providerResend, apiKey: "k"}
	err := m.post(context.Background(), "resend", srv.URL,
		map[string]any{"subject": "hi"},
		func(r *http.Request) { r.Header.Set("Authorization", "Bearer "+m.apiKey) })
	if err != nil {
		t.Fatalf("post: %v", err)
	}
	if gotAuth != "Bearer k" {
		t.Errorf("auth header = %q", gotAuth)
	}
	if !strings.Contains(gotBody, `"subject":"hi"`) {
		t.Errorf("body = %q", gotBody)
	}
}

func TestPostErrorStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"message":"invalid api key"}`))
	}))
	defer srv.Close()

	m := &httpMailer{client: httpClient(), provider: providerResend, apiKey: "bad"}
	err := m.post(context.Background(), "resend", srv.URL, map[string]any{}, func(*http.Request) {})
	if err == nil || !strings.Contains(err.Error(), "401") {
		t.Fatalf("expected 401 error, got %v", err)
	}
}

func TestNewSelectsProvider(t *testing.T) {
	// resend/brevo need a key; without one, disabled (nil).
	if m, _ := New(Config{Provider: "resend"}); m != nil {
		t.Error("resend without key should be nil")
	}
	if m, _ := New(Config{Provider: "resend", APIKey: "k"}); m == nil {
		t.Error("resend with key should be non-nil")
	}
	if m, _ := New(Config{Provider: "brevo", APIKey: "k"}); m == nil {
		t.Error("brevo with key should be non-nil")
	}
	// smtp without host -> nil (not a non-nil interface wrapping a nil pointer).
	if m, _ := New(Config{Provider: "smtp"}); m != nil {
		t.Error("smtp without host should be nil interface")
	}
}
