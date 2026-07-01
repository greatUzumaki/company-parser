package export

import (
	"encoding/csv"
	"io"

	"github.com/parse-companies/backend/internal/domain"
)

type csvEncoder struct{}

// Encode streams a header row followed by one row per company. encoding/csv
// handles quoting/escaping of commas, quotes, and newlines for us.
func (csvEncoder) Encode(w io.Writer, companies []domain.Company) error {
	cw := csv.NewWriter(w)
	if err := cw.Write(columns); err != nil {
		return err
	}
	for _, c := range companies {
		if err := cw.Write(row(c)); err != nil {
			return err
		}
	}
	cw.Flush()
	return cw.Error()
}

func (csvEncoder) ContentType() string { return "text/csv" }
func (csvEncoder) Ext() string         { return "csv" }
