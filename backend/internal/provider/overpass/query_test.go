package overpass

import (
	"strings"
	"testing"

	"github.com/parse-companies/backend/internal/domain"
)

func TestBuildQuery(t *testing.T) {
	t.Run("relation id gets area offset", func(t *testing.T) {
		q := BuildQuery(domain.Region{OSMAreaID: 2145}, domain.Filter{Categories: []string{"shop"}})
		if !strings.Contains(q, "area(3600002145)->.a;") {
			t.Errorf("missing offset area id, got:\n%s", q)
		}
	})

	t.Run("already-area id passes through", func(t *testing.T) {
		q := BuildQuery(domain.Region{OSMAreaID: 3600002145}, domain.Filter{Categories: []string{"shop"}})
		if !strings.Contains(q, "area(3600002145)->.a;") {
			t.Errorf("area id altered, got:\n%s", q)
		}
	})

	t.Run("selected category emits node and way clauses", func(t *testing.T) {
		q := BuildQuery(domain.Region{OSMAreaID: 1}, domain.Filter{Categories: []string{"shop"}})
		if !strings.Contains(q, `node["shop"](area.a);`) {
			t.Errorf("missing node clause, got:\n%s", q)
		}
		if !strings.Contains(q, `way["shop"](area.a);`) {
			t.Errorf("missing way clause, got:\n%s", q)
		}
	})

	t.Run("empty categories fall back to default whitelist", func(t *testing.T) {
		q := BuildQuery(domain.Region{OSMAreaID: 1}, domain.Filter{})
		for _, c := range []string{"shop", "amenity", "office", "craft", "tourism"} {
			if !strings.Contains(q, `node["`+c+`"](area.a);`) {
				t.Errorf("missing default category %q, got:\n%s", c, q)
			}
		}
	})

	t.Run("ends with out center tags", func(t *testing.T) {
		q := BuildQuery(domain.Region{OSMAreaID: 1}, domain.Filter{})
		if !strings.Contains(q, "out center tags;") {
			t.Errorf("missing out statement, got:\n%s", q)
		}
	})

	t.Run("no area id falls back to bbox scoping", func(t *testing.T) {
		// BBox is [minLon, minLat, maxLon, maxLat]; Overpass wants (S,W,N,E).
		q := BuildQuery(
			domain.Region{OSMAreaID: 0, BBox: [4]float64{7.4, 43.7, 7.44, 43.75}},
			domain.Filter{Categories: []string{"shop"}},
		)
		if strings.Contains(q, "area(") {
			t.Errorf("should not use area scoping when id is 0, got:\n%s", q)
		}
		if !strings.Contains(q, `node["shop"](43.7,7.4,43.75,7.44);`) {
			t.Errorf("missing bbox-scoped clause (S,W,N,E), got:\n%s", q)
		}
	})
}
