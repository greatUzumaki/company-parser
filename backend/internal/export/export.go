// Package export turns a list of companies into a downloadable file. Every
// encoder streams to an io.Writer so large result sets never buffer fully in
// memory.
package export

import (
	"errors"
	"io"
	"strconv"

	"github.com/parse-companies/backend/internal/domain"
)

// ErrBadFormat is returned for an unknown export format.
var ErrBadFormat = errors.New("export: unsupported format")

// Encoder writes companies in a specific file format.
type Encoder interface {
	Encode(w io.Writer, companies []domain.Company) error
	ContentType() string
	Ext() string
}

// For returns the encoder for "json", "csv", or "xlsx".
func For(format string) (Encoder, error) {
	switch format {
	case "json":
		return jsonEncoder{}, nil
	case "csv":
		return csvEncoder{}, nil
	case "xlsx":
		return xlsxEncoder{}, nil
	default:
		return nil, ErrBadFormat
	}
}

// columns defines the stable column order shared by CSV and XLSX (spec §7).
var columns = []string{
	"osmType", "osmId", "name", "category", "subcategory",
	"website", "phone", "email", "instagram", "facebook", "vk", "telegram", "whatsapp",
	"country", "city", "street", "housenumber", "postcode",
	"lat", "lon", "openingHours",
}

// row flattens a company into the column order above.
func row(c domain.Company) []string {
	return []string{
		c.OSMType, c.OSMID, c.Name, c.Category, c.Subcat,
		c.Website, c.Phone, c.Email, c.Instagram, c.Facebook, c.VK, c.Telegram, c.WhatsApp,
		c.Addr.Country, c.Addr.City, c.Addr.Street, c.Addr.Housenumber, c.Addr.Postcode,
		strconv.FormatFloat(c.Lat, 'f', -1, 64),
		strconv.FormatFloat(c.Lon, 'f', -1, 64),
		c.OpeningHours,
	}
}
