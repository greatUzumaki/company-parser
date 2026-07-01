package wikidata

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/parse-companies/backend/internal/domain"
	"github.com/parse-companies/backend/internal/provider"
)

const sampleSPARQL = `{"results":{"bindings":[
  {"item":{"value":"http://www.wikidata.org/entity/Q42"},"itemLabel":{"value":"Acme Corp"},"loc":{"value":"Point(7.42 43.73)"},"website":{"value":"https://acme.test"}},
  {"item":{"value":"http://www.wikidata.org/entity/Q99"},"itemLabel":{"value":"Q99"},"loc":{"value":"Point(7.40 43.70)"}}
]}}`

func TestSearchMapsBindings(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte(sampleSPARQL))
	}))
	defer srv.Close()

	region := domain.Region{BBox: [4]float64{7.4, 43.7, 7.44, 43.75}}
	got, err := New(srv.Client(), srv.URL).Search(context.Background(), region, domain.Filter{})
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("got %d, want 2", len(got))
	}

	acme := got[0]
	if acme.OSMType != "wikidata" || acme.OSMID != "Q42" {
		t.Errorf("identity wrong: %+v", acme)
	}
	if acme.Name != "Acme Corp" || acme.Website != "https://acme.test" {
		t.Errorf("fields wrong: %+v", acme)
	}
	if acme.Lat != 43.73 || acme.Lon != 7.42 {
		t.Errorf("coords wrong: lat=%v lon=%v", acme.Lat, acme.Lon)
	}
	// The second item's label equals its QID -> treated as no name.
	if got[1].Name != "" {
		t.Errorf("QID label should map to empty name, got %q", got[1].Name)
	}
}

func TestSearchAppliesFilter(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte(sampleSPARQL))
	}))
	defer srv.Close()

	region := domain.Region{BBox: [4]float64{7.4, 43.7, 7.44, 43.75}}
	got, err := New(srv.Client(), srv.URL).Search(context.Background(), region, domain.Filter{NoWebsite: true})
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	// Acme has a website -> excluded; Q99 has none -> kept.
	if len(got) != 1 || got[0].OSMID != "Q99" {
		t.Fatalf("filter wrong: %+v", got)
	}
}

func TestSearchNoBBoxReturnsNil(t *testing.T) {
	got, err := New(nil, "http://x").Search(context.Background(), domain.Region{}, domain.Filter{})
	if err != nil || got != nil {
		t.Errorf("expected nil,nil for empty bbox; got %v, %v", got, err)
	}
}

func TestSearch429IsUpstreamBusy(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer srv.Close()

	region := domain.Region{BBox: [4]float64{7.4, 43.7, 7.44, 43.75}}
	_, err := New(srv.Client(), srv.URL).Search(context.Background(), region, domain.Filter{})
	if !errors.Is(err, provider.ErrUpstreamBusy) {
		t.Fatalf("err = %v, want ErrUpstreamBusy", err)
	}
}

func TestSatisfiesProvider(t *testing.T) {
	var _ provider.Provider = New(nil, "http://x")
	// silence unused import if assertion form changes
	_ = strings.TrimSpace
}
