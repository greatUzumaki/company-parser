// Package wikidata implements provider.Provider against the Wikidata Query
// Service (SPARQL). Wikidata is CC0 — free and legal to reuse. It contributes
// notable organizations/companies within a bounding box, complementing OSM.
package wikidata

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/parse-companies/backend/internal/domain"
	"github.com/parse-companies/backend/internal/provider"
)

const userAgent = "parse-companies/0.1 (+https://github.com/parse-companies)"

// Client queries the Wikidata SPARQL endpoint.
type Client struct {
	http     *http.Client
	endpoint string
}

// New returns a Client. A nil httpClient uses http.DefaultClient.
func New(httpClient *http.Client, endpoint string) *Client {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return &Client{http: httpClient, endpoint: endpoint}
}

// buildSPARQL scopes to the region's bounding box and pulls organizations with
// coordinates plus their contact info. Wikidata's box service is indexed, so
// the bbox narrows the candidate set before the type filter runs.
func buildSPARQL(r domain.Region) string {
	west, south := r.BBox[0], r.BBox[1]
	east, north := r.BBox[2], r.BBox[3]
	return fmt.Sprintf(`SELECT ?item ?itemLabel ?loc ?website ?phone ?email ?facebook ?instagram ?vk ?telegram WHERE {
  SERVICE wikibase:box {
    ?item wdt:P625 ?loc .
    bd:serviceParam wikibase:cornerWest "Point(%g %g)"^^geo:wktLiteral .
    bd:serviceParam wikibase:cornerEast "Point(%g %g)"^^geo:wktLiteral .
  }
  ?item wdt:P31/wdt:P279* wd:Q4830453 .
  OPTIONAL { ?item wdt:P856 ?website. }
  OPTIONAL { ?item wdt:P1329 ?phone. }
  OPTIONAL { ?item wdt:P968 ?email. }
  OPTIONAL { ?item wdt:P2013 ?facebook. }
  OPTIONAL { ?item wdt:P2003 ?instagram. }
  OPTIONAL { ?item wdt:P3185 ?vk. }
  OPTIONAL { ?item wdt:P3789 ?telegram. }
  SERVICE wikibase:label { bd:serviceParam wikibase:language "en,ru,de,fr,es". }
}
LIMIT 500`, west, south, east, north)
}

type sparqlResponse struct {
	Results struct {
		Bindings []map[string]struct {
			Value string `json:"value"`
		} `json:"bindings"`
	} `json:"results"`
}

var qidRe = regexp.MustCompile(`Q\d+$`)
var pointRe = regexp.MustCompile(`Point\(([-\d.]+) ([-\d.]+)\)`)

// Search runs the SPARQL query for the region, maps rows to companies, and
// applies the filter. It satisfies provider.Provider.
func (c *Client) Search(ctx context.Context, r domain.Region, f domain.Filter) ([]domain.Company, error) {
	// Wikidata scopes by bbox; without one there is nothing to query.
	if r.BBox == ([4]float64{}) {
		return nil, nil
	}

	q := url.Values{"query": {buildSPARQL(r)}, "format": {"json"}}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint, strings.NewReader(q.Encode()))
	if err != nil {
		return nil, fmt.Errorf("wikidata: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/sparql-results+json")
	req.Header.Set("User-Agent", userAgent)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("wikidata: do request: %w", err)
	}
	defer resp.Body.Close()

	switch {
	case resp.StatusCode == http.StatusOK:
	case resp.StatusCode == http.StatusTooManyRequests, resp.StatusCode >= 500:
		return nil, fmt.Errorf("wikidata: status %d: %w", resp.StatusCode, provider.ErrUpstreamBusy)
	default:
		return nil, fmt.Errorf("wikidata: unexpected status %d", resp.StatusCode)
	}

	var out sparqlResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("wikidata: decode: %w", err)
	}

	companies := make([]domain.Company, 0, len(out.Results.Bindings))
	for _, b := range out.Results.Bindings {
		company, ok := mapBinding(b)
		if !ok {
			continue
		}
		if f.Match(company) && r.Accept(company.Lat, company.Lon) {
			companies = append(companies, company)
		}
	}
	return companies, nil
}

// mapBinding converts one SPARQL row into a domain.Company. Returns false if the
// row has no usable coordinates.
func mapBinding(b map[string]struct {
	Value string `json:"value"`
}) (domain.Company, bool) {
	loc := b["loc"].Value
	m := pointRe.FindStringSubmatch(loc)
	if m == nil {
		return domain.Company{}, false
	}
	lon, _ := strconv.ParseFloat(m[1], 64)
	lat, _ := strconv.ParseFloat(m[2], 64)

	qid := qidRe.FindString(b["item"].Value)
	name := b["itemLabel"].Value
	// When there is no English/localized label, Wikidata returns the QID as the
	// label; treat that as no name.
	if name == qid {
		name = ""
	}

	return domain.Company{
		OSMType:   "wikidata",
		OSMID:     qid,
		Name:      name,
		Website:   b["website"].Value,
		Phone:     b["phone"].Value,
		Email:     strings.TrimPrefix(b["email"].Value, "mailto:"),
		Facebook:  b["facebook"].Value,
		Instagram: b["instagram"].Value,
		VK:        b["vk"].Value,
		Telegram:  b["telegram"].Value,
		Lat:       lat,
		Lon:       lon,
		Tags:      map[string]string{"source": "wikidata"},
	}, true
}
