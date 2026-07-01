package domain

import "testing"

func TestRegionContains(t *testing.T) {
	// A square [lon,lat] from (0,0) to (10,10).
	r := Region{Polygon: [][2]float64{{0, 0}, {10, 0}, {10, 10}, {0, 10}}}

	if !r.HasPolygon() {
		t.Fatal("expected HasPolygon")
	}
	// Contains(lat, lon)
	if !r.Contains(5, 5) {
		t.Error("point (5,5) should be inside")
	}
	if r.Contains(20, 5) {
		t.Error("point (20,5) should be outside")
	}
	if r.Contains(5, -1) {
		t.Error("point (5,-1) should be outside")
	}
}

func TestRegionAccept(t *testing.T) {
	none := Region{}
	if !none.Accept(99, 99) {
		t.Error("region without polygon should accept any point")
	}
	square := Region{Polygon: [][2]float64{{0, 0}, {10, 0}, {10, 10}, {0, 10}}}
	if !square.Accept(5, 5) || square.Accept(50, 50) {
		t.Error("polygon Accept should mirror Contains")
	}
}
