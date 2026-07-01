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

// Cache stores and retrieves company lists by a region+filter key.
type Cache interface {
	Get(ctx context.Context, key string) ([]domain.Company, bool, error)
	Set(ctx context.Context, key string, companies []domain.Company, ttl time.Duration) error
}

// Key derives a stable cache key from a region and filter. Category order is
// normalized so the same selection always hashes identically.
func Key(r domain.Region, f domain.Filter) string {
	cats := append([]string(nil), f.Categories...)
	sort.Strings(cats)
	raw := fmt.Sprintf("%d|%t|%t|%t|%s",
		r.OSMAreaID, f.NoWebsite, f.NoSocials, f.NoPhone, strings.Join(cats, ","))
	sum := sha256.Sum256([]byte(raw))
	return "search:" + hex.EncodeToString(sum[:])
}

// encode/decode keep the wire format in one place.
func encode(companies []domain.Company) ([]byte, error) { return json.Marshal(companies) }

func decode(b []byte) ([]domain.Company, error) {
	var cs []domain.Company
	if err := json.Unmarshal(b, &cs); err != nil {
		return nil, err
	}
	return cs, nil
}
