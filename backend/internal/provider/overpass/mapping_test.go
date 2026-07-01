package overpass

import "testing"

func TestMapElement(t *testing.T) {
	t.Run("contact:website fills website when website absent", func(t *testing.T) {
		c := mapElement(element{
			Type: "node", ID: 1,
			Tags: map[string]string{"contact:website": "https://x.test"},
		})
		if c.Website != "https://x.test" {
			t.Errorf("website = %q, want fallback from contact:website", c.Website)
		}
	})

	t.Run("plain website wins over contact:website", func(t *testing.T) {
		c := mapElement(element{
			Type: "node", ID: 1,
			Tags: map[string]string{"website": "a", "contact:website": "b"},
		})
		if c.Website != "a" {
			t.Errorf("website = %q, want primary tag", c.Website)
		}
	})

	t.Run("way uses center coordinates", func(t *testing.T) {
		c := mapElement(element{
			Type: "way", ID: 2,
			Center: &center{Lat: 47.5, Lon: 8.5},
			Tags:   map[string]string{"shop": "bakery"},
		})
		if c.Lat != 47.5 || c.Lon != 8.5 {
			t.Errorf("coords = (%v,%v), want center", c.Lat, c.Lon)
		}
	})

	t.Run("category and subcategory from first whitelist key", func(t *testing.T) {
		c := mapElement(element{
			Type: "node", ID: 3,
			Lat: 1, Lon: 2,
			Tags: map[string]string{"shop": "bakery", "name": "Brot"},
		})
		if c.Category != "shop" || c.Subcat != "bakery" {
			t.Errorf("category/subcat = %q/%q, want shop/bakery", c.Category, c.Subcat)
		}
		if c.Name != "Brot" {
			t.Errorf("name = %q", c.Name)
		}
	})

	t.Run("osm id formatted as string", func(t *testing.T) {
		c := mapElement(element{Type: "node", ID: 123, Tags: map[string]string{}})
		if c.OSMID != "123" {
			t.Errorf("osmId = %q, want 123", c.OSMID)
		}
	})

	t.Run("address mapped from addr tags", func(t *testing.T) {
		c := mapElement(element{Type: "node", ID: 1, Tags: map[string]string{
			"addr:city": "Munich", "addr:street": "Main", "addr:postcode": "80331",
		}})
		if c.Addr.City != "Munich" || c.Addr.Street != "Main" || c.Addr.Postcode != "80331" {
			t.Errorf("addr = %+v", c.Addr)
		}
	})
}
