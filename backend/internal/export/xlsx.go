package export

import (
	"fmt"
	"io"

	"github.com/xuri/excelize/v2"

	"github.com/parse-companies/backend/internal/domain"
)

type xlsxEncoder struct{}

// Encode writes an xlsx workbook via excelize's StreamWriter, which flushes
// rows incrementally instead of holding the whole sheet in memory.
func (xlsxEncoder) Encode(w io.Writer, companies []domain.Company) error {
	f := excelize.NewFile()
	defer f.Close()

	sw, err := f.NewStreamWriter("Sheet1")
	if err != nil {
		return fmt.Errorf("export: stream writer: %w", err)
	}

	header := make([]interface{}, len(columns))
	for i, c := range columns {
		header[i] = c
	}
	if err := sw.SetRow("A1", header); err != nil {
		return fmt.Errorf("export: header row: %w", err)
	}

	for i, c := range companies {
		cell, err := excelize.CoordinatesToCellName(1, i+2)
		if err != nil {
			return fmt.Errorf("export: cell name: %w", err)
		}
		vals := row(c)
		cells := make([]interface{}, len(vals))
		for j, v := range vals {
			cells[j] = v
		}
		if err := sw.SetRow(cell, cells); err != nil {
			return fmt.Errorf("export: data row: %w", err)
		}
	}

	if err := sw.Flush(); err != nil {
		return fmt.Errorf("export: flush: %w", err)
	}
	if err := f.Write(w); err != nil {
		return fmt.Errorf("export: write workbook: %w", err)
	}
	return nil
}

func (xlsxEncoder) ContentType() string {
	return "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
}
func (xlsxEncoder) Ext() string { return "xlsx" }
