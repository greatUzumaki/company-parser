package overpass

import (
	"fmt"
	"strings"

	"github.com/parse-companies/backend/internal/domain"
)

// areaIDOffset is Overpass's mapping from an OSM relation id to an area id.
const areaIDOffset = 3_600_000_000

// defaultCategories is the business-tag whitelist used when the filter selects
// no specific category.
var defaultCategories = []string{"shop", "amenity", "office", "craft", "tourism"}

// BuildQuery renders an Overpass QL query for the region and filter.
//
// When the region has an OSM area id it scopes by admin area (precise, efficient
// for large regions). Otherwise it falls back to the region's bounding box —
// this is what map clicks on the bundled countries layer use, since those carry
// a bbox but no area id.
//
// It narrows by area/bbox + category only; absence filters (no
// website/socials/phone) are applied in Go after the fetch, because pure-QL
// absence queries are slow.
func BuildQuery(r domain.Region, f domain.Filter) string {
	cats := f.Categories
	if len(cats) == 0 {
		cats = defaultCategories
	}

	var b strings.Builder
	b.WriteString("[out:json][timeout:180];\n")

	if r.OSMAreaID != 0 {
		fmt.Fprintf(&b, "area(%d)->.a;\n", areaID(r.OSMAreaID))
		b.WriteString("(\n")
		for _, c := range cats {
			fmt.Fprintf(&b, "  node[%q](area.a);\n", c)
			fmt.Fprintf(&b, "  way[%q](area.a);\n", c)
		}
	} else {
		// Overpass bbox order is (south,west,north,east) = (minLat,minLon,maxLat,maxLon).
		south, west, north, east := r.BBox[1], r.BBox[0], r.BBox[3], r.BBox[2]
		b.WriteString("(\n")
		for _, c := range cats {
			fmt.Fprintf(&b, "  node[%q](%g,%g,%g,%g);\n", c, south, west, north, east)
			fmt.Fprintf(&b, "  way[%q](%g,%g,%g,%g);\n", c, south, west, north, east)
		}
	}

	b.WriteString(");\n")
	b.WriteString("out center tags;")
	return b.String()
}

// areaID returns a valid Overpass area id. Values already in area space
// (> offset) pass through; smaller values are treated as relation ids.
func areaID(id int64) int64 {
	if id >= areaIDOffset {
		return id
	}
	return areaIDOffset + id
}
