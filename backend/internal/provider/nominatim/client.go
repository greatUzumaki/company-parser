// Package nominatim resolves human region names to OSM areas for the map's
// region search box. It uses the public Nominatim API per its usage policy
// (identifying User-Agent, low volume, results cached upstream by the caller).
package nominatim

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/parse-companies/backend/internal/domain"
)

const userAgent = "parse-companies/0.1 (+https://github.com/parse-companies)"

// Client queries a Nominatim endpoint.
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

type nominatimResult struct {
	DisplayName string    `json:"display_name"`
	OSMType     string    `json:"osm_type"` // relation | way | node
	OSMID       int64     `json:"osm_id"`
	BoundingBox [4]string `json:"boundingbox"` // [minLat, maxLat, minLon, maxLon]
}

// Search returns regions matching the query, as domain.Region values. Only
// relations (administrative areas) get a usable Overpass area id; others return
// area id 0 and rely on the bbox.
func (c *Client) Search(ctx context.Context, query string) ([]domain.Region, error) {
	u, err := url.Parse(c.endpoint + "/search")
	if err != nil {
		return nil, fmt.Errorf("nominatim: parse endpoint: %w", err)
	}
	q := u.Query()
	q.Set("q", query)
	q.Set("format", "jsonv2")
	q.Set("limit", "8")
	q.Set("featureType", "settlement")
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("nominatim: build request: %w", err)
	}
	req.Header.Set("User-Agent", userAgent)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("nominatim: do request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("nominatim: status %d", resp.StatusCode)
	}

	var results []nominatimResult
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return nil, fmt.Errorf("nominatim: decode: %w", err)
	}

	regions := make([]domain.Region, 0, len(results))
	for _, r := range results {
		regions = append(regions, domain.Region{
			Name:      r.DisplayName,
			OSMAreaID: areaID(r.OSMType, r.OSMID),
			BBox:      bbox(r.BoundingBox),
		})
	}
	return regions, nil
}

// areaID converts a relation id into an Overpass area id; non-relations have no
// area and return 0.
func areaID(osmType string, id int64) int64 {
	if osmType == "relation" {
		return 3_600_000_000 + id
	}
	return 0
}

// bbox converts Nominatim's [minLat,maxLat,minLon,maxLon] strings into
// [minLon,minLat,maxLon,maxLat] floats.
func bbox(bb [4]string) [4]float64 {
	minLat := parseF(bb[0])
	maxLat := parseF(bb[1])
	minLon := parseF(bb[2])
	maxLon := parseF(bb[3])
	return [4]float64{minLon, minLat, maxLon, maxLat}
}

func parseF(s string) float64 {
	v, _ := strconv.ParseFloat(s, 64)
	return v
}
