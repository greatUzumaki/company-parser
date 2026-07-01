package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/parse-companies/backend/internal/domain"
)

type fakeProvider struct {
	calls     int
	companies []domain.Company
	err       error
}

func (f *fakeProvider) Search(_ context.Context, _ domain.Region, _ domain.Filter) ([]domain.Company, error) {
	f.calls++
	return f.companies, f.err
}

type fakeRepo struct {
	upserted   []domain.Company
	createdFor []int64
}

func (r *fakeRepo) UpsertCompanies(_ context.Context, c []domain.Company) ([]int64, error) {
	r.upserted = c
	ids := make([]int64, len(c))
	for i := range c {
		ids[i] = int64(i + 1)
	}
	return ids, nil
}

func (r *fakeRepo) CreateSearch(_ context.Context, _ domain.Region, _ domain.Filter, ids []int64) (int64, error) {
	r.createdFor = ids
	return 99, nil
}

type fakeCache struct{ stored map[string][]domain.Company }

func newFakeCache() *fakeCache { return &fakeCache{stored: map[string][]domain.Company{}} }
func (c *fakeCache) Get(_ context.Context, key string) ([]domain.Company, bool, error) {
	v, ok := c.stored[key]
	return v, ok, nil
}
func (c *fakeCache) Set(_ context.Context, key string, v []domain.Company, _ time.Duration) error {
	c.stored[key] = v
	return nil
}

func providers(ps ...*fakeProvider) []NamedProvider {
	out := make([]NamedProvider, len(ps))
	for i, p := range ps {
		out[i] = NamedProvider{Name: "p" + string(rune('A'+i)), P: p}
	}
	return out
}

func TestSearchMergesProvidersAndPersists(t *testing.T) {
	a := &fakeProvider{companies: []domain.Company{{OSMType: "node", OSMID: "1", Name: "Alpha", Lat: 1, Lon: 1}}}
	b := &fakeProvider{companies: []domain.Company{{OSMType: "wikidata", OSMID: "Q2", Name: "Beta", Lat: 2, Lon: 2}}}
	repo := &fakeRepo{}
	svc := New(providers(a, b), repo, newFakeCache())

	id, companies, err := svc.Search(context.Background(), domain.Region{OSMAreaID: 1}, domain.Filter{})
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if id != 99 || len(companies) != 2 {
		t.Fatalf("id=%d count=%d", id, len(companies))
	}
	if a.calls != 1 || b.calls != 1 {
		t.Errorf("providers not both called: a=%d b=%d", a.calls, b.calls)
	}
}

func TestSearchDedupsAcrossSources(t *testing.T) {
	// Same name + same rounded coords from two sources → one merged company.
	a := &fakeProvider{companies: []domain.Company{{OSMType: "node", OSMID: "1", Name: "Cafe X", Lat: 47.1234, Lon: 8.1234}}}
	b := &fakeProvider{companies: []domain.Company{{OSMType: "wikidata", OSMID: "Q9", Name: "Cafe X", Lat: 47.1234, Lon: 8.1234, Website: "https://x.test"}}}
	svc := New(providers(a, b), &fakeRepo{}, newFakeCache())

	_, companies, err := svc.Search(context.Background(), domain.Region{OSMAreaID: 1}, domain.Filter{})
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if len(companies) != 1 {
		t.Fatalf("expected 1 merged company, got %d", len(companies))
	}
	if companies[0].Website != "https://x.test" {
		t.Errorf("merge did not fill website: %+v", companies[0])
	}
}

func TestSearchCacheHitSkipsProviders(t *testing.T) {
	a := &fakeProvider{companies: []domain.Company{{OSMID: "1"}}}
	c := newFakeCache()
	svc := New(providers(a), &fakeRepo{}, c)
	region := domain.Region{OSMAreaID: 7}

	if _, _, err := svc.Search(context.Background(), region, domain.Filter{}); err != nil {
		t.Fatalf("first: %v", err)
	}
	if _, _, err := svc.Search(context.Background(), region, domain.Filter{}); err != nil {
		t.Fatalf("second: %v", err)
	}
	if a.calls != 1 {
		t.Errorf("provider called %d times, want 1 (cache should serve second)", a.calls)
	}
}

func TestSearchSurvivesPartialProviderFailure(t *testing.T) {
	a := &fakeProvider{err: errors.New("boom")}
	b := &fakeProvider{companies: []domain.Company{{OSMID: "1", Name: "ok"}}}
	svc := New(providers(a, b), &fakeRepo{}, newFakeCache())

	_, companies, err := svc.Search(context.Background(), domain.Region{OSMAreaID: 1}, domain.Filter{})
	if err != nil {
		t.Fatalf("should tolerate one failure, got %v", err)
	}
	if len(companies) != 1 {
		t.Errorf("expected the surviving provider's result, got %d", len(companies))
	}
}

func TestSearchStreamEmitsProgressAndDone(t *testing.T) {
	a := &fakeProvider{companies: []domain.Company{{OSMType: "node", OSMID: "1", Name: "A", Lat: 1, Lon: 1}}}
	b := &fakeProvider{companies: []domain.Company{{OSMType: "wikidata", OSMID: "Q2", Name: "B", Lat: 2, Lon: 2}}}
	svc := New(providers(a, b), &fakeRepo{}, newFakeCache())

	var events []Event
	err := svc.SearchStream(context.Background(), domain.Region{OSMAreaID: 1}, domain.Filter{}, func(e Event) error {
		events = append(events, e)
		return nil
	})
	if err != nil {
		t.Fatalf("stream: %v", err)
	}

	var starts, companies, dones, doneFinal int
	var searchID int64
	for _, e := range events {
		switch e.Type {
		case "source_start":
			starts++
		case "companies":
			companies += len(e.Companies)
		case "source_done":
			dones++
		case "done":
			doneFinal++
			searchID = e.SearchID
		}
	}
	if starts != 2 {
		t.Errorf("source_start events = %d, want 2", starts)
	}
	if companies != 2 {
		t.Errorf("streamed companies = %d, want 2", companies)
	}
	if dones != 2 || doneFinal != 1 {
		t.Errorf("source_done=%d done=%d", dones, doneFinal)
	}
	if searchID != 99 {
		t.Errorf("final searchId = %d, want 99", searchID)
	}
}
