package export

import (
	"encoding/json"
	"io"

	"github.com/parse-companies/backend/internal/domain"
)

type jsonEncoder struct{}

// Encode streams the companies as a JSON array. domain.Company already carries
// the public json tags, so it serializes to the API shape.
func (jsonEncoder) Encode(w io.Writer, companies []domain.Company) error {
	if companies == nil {
		companies = []domain.Company{}
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(companies)
}

func (jsonEncoder) ContentType() string { return "application/json" }
func (jsonEncoder) Ext() string         { return "json" }
