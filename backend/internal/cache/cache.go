// Package cache memoizes provider results in Redis so repeated searches for the
// same region+filter don't re-hit the upstream API.
package cache

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/parse-companies/backend/internal/domain"
)

// Entry is a cached search result plus when it was fetched, so callers can
// decide whether it is fresh enough or should be refreshed.
type Entry struct {
	Companies []domain.Company `json:"companies"`
	FetchedAt time.Time        `json:"fetchedAt"`
}

// Cache stores and retrieves company lists by a region+filter key.
type Cache interface {
	Get(ctx context.Context, key string) (Entry, bool, error)
	Set(ctx context.Context, key string, companies []domain.Company, ttl time.Duration) error
}

// Key derives a stable cache key from a region and filter. Category order is
// normalized so the same selection always hashes identically. The area id,
// bounding box, and drawn polygon all feed the key so distinct zones (which
// share osmAreaId 0) never collide.
func Key(r domain.Region, f domain.Filter) string {
	cats := append([]string(nil), f.Categories...)
	sort.Strings(cats)
	var poly strings.Builder
	for _, p := range r.Polygon {
		fmt.Fprintf(&poly, "%.5f,%.5f;", p[0], p[1])
	}
	raw := fmt.Sprintf("%d|%.5f,%.5f,%.5f,%.5f|%s|%t|%t|%t|%s",
		r.OSMAreaID, r.BBox[0], r.BBox[1], r.BBox[2], r.BBox[3], poly.String(),
		f.NoWebsite, f.NoSocials, f.NoPhone, strings.Join(cats, ","))
	sum := sha256.Sum256([]byte(raw))
	return "search:" + hex.EncodeToString(sum[:])
}

// encode/decode keep the wire format in one place.
func encode(e Entry) ([]byte, error) { return json.Marshal(e) }

func decode(b []byte) (Entry, error) {
	var e Entry
	if err := json.Unmarshal(b, &e); err != nil {
		return Entry{}, err
	}
	return e, nil
}
