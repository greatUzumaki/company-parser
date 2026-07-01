package overpass

import (
	"strconv"

	"github.com/parse-companies/backend/internal/domain"
)

// element is one item from an Overpass `{elements:[...]}` response.
type element struct {
	Type   string            `json:"type"` // node | way | relation
	ID     int64             `json:"id"`
	Lat    float64           `json:"lat"`
	Lon    float64           `json:"lon"`
	Center *center           `json:"center"`
	Tags   map[string]string `json:"tags"`
}

type center struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}

// categoryKeys is the ordered whitelist; the first present key wins as Category.
var categoryKeys = []string{"shop", "amenity", "office", "craft", "tourism"}

// mapElement converts a raw OSM element into a domain.Company.
func mapElement(el element) domain.Company {
	t := el.Tags
	c := domain.Company{
		OSMType:      el.Type,
		OSMID:        strconv.FormatInt(el.ID, 10),
		Name:         t["name"],
		Website:      firstNonEmpty(t["website"], t["contact:website"]),
		Phone:        firstNonEmpty(t["phone"], t["contact:phone"]),
		Email:        firstNonEmpty(t["email"], t["contact:email"]),
		Instagram:    firstNonEmpty(t["contact:instagram"], t["instagram"]),
		Facebook:     firstNonEmpty(t["contact:facebook"], t["facebook"]),
		VK:           firstNonEmpty(t["contact:vk"], t["vk"]),
		Telegram:     firstNonEmpty(t["contact:telegram"], t["telegram"]),
		WhatsApp:     firstNonEmpty(t["contact:whatsapp"], t["whatsapp"]),
		OpeningHours: t["opening_hours"],
		Addr: domain.Address{
			Country:     t["addr:country"],
			City:        t["addr:city"],
			Street:      t["addr:street"],
			Housenumber: t["addr:housenumber"],
			Postcode:    t["addr:postcode"],
		},
		Tags: t,
	}

	for _, key := range categoryKeys {
		if v, ok := t[key]; ok {
			c.Category = key
			c.Subcat = v
			break
		}
	}

	// Nodes carry lat/lon directly; ways/relations use the computed center.
	if el.Center != nil {
		c.Lat, c.Lon = el.Center.Lat, el.Center.Lon
	} else {
		c.Lat, c.Lon = el.Lat, el.Lon
	}

	return c
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}
