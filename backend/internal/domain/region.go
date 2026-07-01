package domain

// Region is a search area selected on the map. OSMAreaID is the Overpass area
// id (3600000000 + relation id) when known; BBox is minLon,minLat,maxLon,maxLat
// and is used for tiling large areas and as a fallback scope.
type Region struct {
	Name      string     `json:"name"`
	OSMAreaID int64      `json:"osmAreaId"`
	BBox      [4]float64 `json:"bbox"`
}
