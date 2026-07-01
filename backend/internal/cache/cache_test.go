package cache

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"

	"github.com/parse-companies/backend/internal/domain"
)

func TestKeyStable(t *testing.T) {
	r := domain.Region{OSMAreaID: 42}
	a := Key(r, domain.Filter{Categories: []string{"shop", "amenity"}})
	b := Key(r, domain.Filter{Categories: []string{"amenity", "shop"}})
	if a != b {
		t.Errorf("key not order-independent: %s vs %s", a, b)
	}
	c := Key(r, domain.Filter{NoWebsite: true, Categories: []string{"shop", "amenity"}})
	if a == c {
		t.Error("key should change when a filter flag changes")
	}
}

func TestRedisRoundTrip(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("miniredis: %v", err)
	}
	defer mr.Close()

	c := NewRedisFromClient(redis.NewClient(&redis.Options{Addr: mr.Addr()}))
	ctx := context.Background()
	key := "search:test"

	// Miss before set.
	if _, ok, err := c.Get(ctx, key); err != nil || ok {
		t.Fatalf("expected miss, got ok=%v err=%v", ok, err)
	}

	want := []domain.Company{{OSMType: "node", OSMID: "1", Name: "A", Website: "https://a.test"}}
	if err := c.Set(ctx, key, want, time.Minute); err != nil {
		t.Fatalf("set: %v", err)
	}

	got, ok, err := c.Get(ctx, key)
	if err != nil || !ok {
		t.Fatalf("expected hit, got ok=%v err=%v", ok, err)
	}
	if len(got.Companies) != 1 || got.Companies[0].Name != "A" || got.Companies[0].Website != "https://a.test" {
		t.Errorf("round-trip mismatch: %+v", got)
	}
	if got.FetchedAt.IsZero() {
		t.Error("FetchedAt not stamped on Set")
	}
}
