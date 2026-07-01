package domain

// Region is a search area selected on the map. OSMAreaID is the Overpass area
// id (3600000000 + relation id) when known; BBox is minLon,minLat,maxLon,maxLat
// and is used for tiling large areas and as a fallback scope. Polygon, when set
// (a drawn circle/rectangle/freehand zone), is a ring of [lon,lat] points and
// takes precedence: searches are scoped to and filtered by it.
type Region struct {
	Name      string       `json:"name"`
	OSMAreaID int64        `json:"osmAreaId"`
	BBox      [4]float64   `json:"bbox"`
	Polygon   [][2]float64 `json:"polygon,omitempty"`
}

// HasPolygon reports whether the region is a drawn zone.
func (r Region) HasPolygon() bool { return len(r.Polygon) >= 3 }

// Contains reports whether a point is inside the region's polygon (ray casting).
// Points are [lon,lat]; the argument order is (lat, lon) to match Company.
func (r Region) Contains(lat, lon float64) bool {
	p := r.Polygon
	inside := false
	for i, j := 0, len(p)-1; i < len(p); j, i = i, i+1 {
		xi, yi := p[i][0], p[i][1] // lon, lat
		xj, yj := p[j][0], p[j][1]
		if (yi > lat) != (yj > lat) &&
			lon < (xj-xi)*(lat-yi)/(yj-yi)+xi {
			inside = !inside
		}
	}
	return inside
}

// Accept keeps a point when there is no polygon, or it lies inside one.
func (r Region) Accept(lat, lon float64) bool {
	return !r.HasPolygon() || r.Contains(lat, lon)
}
