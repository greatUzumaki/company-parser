// Package domain holds the core types of the application. It imports nothing
// from the rest of the codebase — every other package depends on domain, never
// the reverse.
package domain

// Company is a business found in the data source, normalized across providers.
// Source-specific extras live in Tags so domain stays provider-agnostic.
type Company struct {
	OSMType      string            `json:"osmType"` // node | way | relation
	OSMID        string            `json:"osmId"`
	Name         string            `json:"name"`
	Category     string            `json:"category"`    // e.g. "shop"
	Subcat       string            `json:"subcategory"` // e.g. "bakery"
	Website      string            `json:"website"`
	Phone        string            `json:"phone"`
	Email        string            `json:"email"`
	Instagram    string            `json:"instagram"`
	Facebook     string            `json:"facebook"`
	VK           string            `json:"vk"`
	Telegram     string            `json:"telegram"`
	WhatsApp     string            `json:"whatsapp"`
	Addr         Address           `json:"addr"`
	Lat          float64           `json:"lat"`
	Lon          float64           `json:"lon"`
	OpeningHours string            `json:"openingHours"`
	Tags         map[string]string `json:"-"`
}

// Address is the structured postal address from addr:* tags.
type Address struct {
	Country     string `json:"country"`
	City        string `json:"city"`
	Street      string `json:"street"`
	Housenumber string `json:"housenumber"`
	Postcode    string `json:"postcode"`
}

// HasSocial reports whether any social-media contact is present.
func (c Company) HasSocial() bool {
	return c.Instagram != "" || c.Facebook != "" || c.VK != "" || c.Telegram != ""
}
