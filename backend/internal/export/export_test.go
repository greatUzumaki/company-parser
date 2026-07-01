package export

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/xuri/excelize/v2"

	"github.com/parse-companies/backend/internal/domain"
)

func fixture() []domain.Company {
	return []domain.Company{
		{OSMType: "node", OSMID: "1", Name: "Alpha, Inc", Category: "shop", Subcat: "bakery",
			Phone: "+1", Addr: domain.Address{City: "Munich"}, Lat: 47.1, Lon: 8.1},
		{OSMType: "way", OSMID: "2", Name: "Beta", Category: "amenity", Subcat: "cafe",
			Website: "https://b.test", Instagram: "@b", Lat: 47.2, Lon: 8.2},
	}
}

func TestForUnknownFormat(t *testing.T) {
	if _, err := For("pdf"); !errors.Is(err, ErrBadFormat) {
		t.Errorf("err = %v, want ErrBadFormat", err)
	}
}

func TestCSVEncode(t *testing.T) {
	enc, _ := For("csv")
	var buf bytes.Buffer
	if err := enc.Encode(&buf, fixture()); err != nil {
		t.Fatalf("encode: %v", err)
	}
	out := buf.String()
	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	if len(lines) != 3 {
		t.Fatalf("got %d lines, want 3 (header + 2)", len(lines))
	}
	if !strings.HasPrefix(lines[0], "osmType,osmId,name,") {
		t.Errorf("header wrong: %s", lines[0])
	}
	// Name with a comma must be quoted by encoding/csv.
	if !strings.Contains(out, `"Alpha, Inc"`) {
		t.Errorf("comma field not quoted: %s", out)
	}
}

func TestJSONEncode(t *testing.T) {
	enc, _ := For("json")
	var buf bytes.Buffer
	if err := enc.Encode(&buf, fixture()); err != nil {
		t.Fatalf("encode: %v", err)
	}
	var got []domain.Company
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if len(got) != 2 || got[1].Website != "https://b.test" {
		t.Errorf("json round-trip wrong: %+v", got)
	}
}

func TestXLSXEncode(t *testing.T) {
	enc, _ := For("xlsx")
	var buf bytes.Buffer
	if err := enc.Encode(&buf, fixture()); err != nil {
		t.Fatalf("encode: %v", err)
	}
	f, err := excelize.OpenReader(&buf)
	if err != nil {
		t.Fatalf("open xlsx: %v", err)
	}
	rows, err := f.GetRows("Sheet1")
	if err != nil {
		t.Fatalf("get rows: %v", err)
	}
	if len(rows) != 3 {
		t.Fatalf("got %d rows, want 3", len(rows))
	}
	if rows[0][0] != "osmType" {
		t.Errorf("header[0] = %q", rows[0][0])
	}
	if rows[1][2] != "Alpha, Inc" {
		t.Errorf("data cell = %q, want 'Alpha, Inc'", rows[1][2])
	}
}

func TestContentTypes(t *testing.T) {
	for _, tc := range []struct{ fmt, ct, ext string }{
		{"json", "application/json", "json"},
		{"csv", "text/csv", "csv"},
		{"xlsx", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", "xlsx"},
	} {
		enc, _ := For(tc.fmt)
		if enc.ContentType() != tc.ct || enc.Ext() != tc.ext {
			t.Errorf("%s: ct=%q ext=%q", tc.fmt, enc.ContentType(), enc.Ext())
		}
	}
}
