package store

import (
	"context"
	"os"
	"testing"

	"github.com/parse-companies/backend/internal/domain"
)

// newTestStore connects to the DB in TEST_DATABASE_URL, skipping if unset, and
// truncates the tables so each test starts clean.
func newTestStore(t *testing.T) *Store {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL not set; skipping integration test")
	}
	s, err := New(context.Background(), dsn)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	_, err = s.pool.Exec(context.Background(),
		"TRUNCATE search_results, searches, companies RESTART IDENTITY CASCADE")
	if err != nil {
		t.Fatalf("truncate: %v", err)
	}
	t.Cleanup(s.Close)
	return s
}

func sampleCompanies() []domain.Company {
	return []domain.Company{
		{OSMType: "node", OSMID: "1", Name: "Alpha", Category: "shop", Lat: 47.1, Lon: 8.1,
			Tags: map[string]string{"shop": "bakery"}},
		{OSMType: "node", OSMID: "2", Name: "Beta", Category: "amenity", Lat: 47.2, Lon: 8.2,
			Website: "https://beta.test", Tags: map[string]string{"amenity": "cafe"}},
	}
}

func TestUpsertDedup(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	ids1, err := s.UpsertCompanies(ctx, sampleCompanies())
	if err != nil {
		t.Fatalf("first upsert: %v", err)
	}
	if len(ids1) != 2 {
		t.Fatalf("got %d ids, want 2", len(ids1))
	}

	// Re-upsert the same OSM identities -> same ids (dedup across runs).
	ids2, err := s.UpsertCompanies(ctx, sampleCompanies())
	if err != nil {
		t.Fatalf("second upsert: %v", err)
	}
	if ids1[0] != ids2[0] || ids1[1] != ids2[1] {
		t.Errorf("ids changed on re-upsert: %v vs %v", ids1, ids2)
	}
}

func TestCreateAndGetSearch(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	ids, err := s.UpsertCompanies(ctx, sampleCompanies())
	if err != nil {
		t.Fatalf("upsert: %v", err)
	}
	region := domain.Region{Name: "Testland", OSMAreaID: 3600000123}
	filter := domain.Filter{NoWebsite: true, Categories: []string{"shop"}}

	searchID, err := s.CreateSearch(ctx, region, filter, ids)
	if err != nil {
		t.Fatalf("create search: %v", err)
	}

	summary, companies, err := s.GetSearchResults(ctx, searchID)
	if err != nil {
		t.Fatalf("get results: %v", err)
	}
	if summary.RegionName != "Testland" || summary.ResultCount != 2 {
		t.Errorf("summary = %+v", summary)
	}
	if !summary.Filter.NoWebsite || len(summary.Filter.Categories) != 1 {
		t.Errorf("filter not round-tripped: %+v", summary.Filter)
	}
	if len(companies) != 2 {
		t.Fatalf("got %d companies, want 2", len(companies))
	}
	// Coordinates survive the PostGIS round-trip.
	var beta domain.Company
	for _, c := range companies {
		if c.Name == "Beta" {
			beta = c
		}
	}
	if beta.Lat < 47.19 || beta.Lat > 47.21 || beta.Website != "https://beta.test" {
		t.Errorf("Beta round-trip wrong: %+v", beta)
	}
}

func TestListSearches(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()
	ids, _ := s.UpsertCompanies(ctx, sampleCompanies())
	for i := 0; i < 3; i++ {
		if _, err := s.CreateSearch(ctx, domain.Region{Name: "R"}, domain.Filter{}, ids); err != nil {
			t.Fatalf("create search %d: %v", i, err)
		}
	}
	list, err := s.ListSearches(ctx, 10, 0)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(list) != 3 {
		t.Errorf("got %d searches, want 3", len(list))
	}
}
