package overpass

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/parse-companies/backend/internal/domain"
	"github.com/parse-companies/backend/internal/provider"
)

const sampleJSON = `{"elements":[
  {"type":"node","id":1,"lat":47.1,"lon":8.1,"tags":{"shop":"bakery","name":"A","website":"https://a.test"}},
  {"type":"node","id":2,"lat":47.2,"lon":8.2,"tags":{"shop":"bakery","name":"B"}},
  {"type":"way","id":3,"center":{"lat":47.3,"lon":8.3},"tags":{"amenity":"cafe","name":"C","contact:instagram":"@c"}}
]}`

func TestClientSearch(t *testing.T) {
	t.Run("maps and returns all when no filter", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Write([]byte(sampleJSON))
		}))
		defer srv.Close()

		got, err := New(srv.Client(), srv.URL).Search(context.Background(), domain.Region{OSMAreaID: 1}, domain.Filter{})
		if err != nil {
			t.Fatalf("Search: %v", err)
		}
		if len(got) != 3 {
			t.Fatalf("got %d companies, want 3", len(got))
		}
	})

	t.Run("applies noWebsite filter", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Write([]byte(sampleJSON))
		}))
		defer srv.Close()

		got, err := New(srv.Client(), srv.URL).Search(context.Background(), domain.Region{OSMAreaID: 1}, domain.Filter{NoWebsite: true})
		if err != nil {
			t.Fatalf("Search: %v", err)
		}
		// A has a website -> excluded. B and C have none -> kept.
		if len(got) != 2 {
			t.Fatalf("got %d companies, want 2", len(got))
		}
		for _, c := range got {
			if c.Website != "" {
				t.Errorf("company %s should have been filtered out", c.Name)
			}
		}
	})

	t.Run("429 maps to ErrUpstreamBusy", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusTooManyRequests)
		}))
		defer srv.Close()

		_, err := New(srv.Client(), srv.URL).Search(context.Background(), domain.Region{OSMAreaID: 1}, domain.Filter{})
		if !errors.Is(err, provider.ErrUpstreamBusy) {
			t.Fatalf("err = %v, want ErrUpstreamBusy", err)
		}
	})

	t.Run("satisfies provider.Provider", func(t *testing.T) {
		var _ provider.Provider = New(nil, "http://x")
	})
}
