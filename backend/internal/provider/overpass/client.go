// Package overpass implements provider.Provider against the OpenStreetMap
// Overpass API.
package overpass

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/parse-companies/backend/internal/domain"
	"github.com/parse-companies/backend/internal/provider"
)

// userAgent identifies this client per the Overpass usage policy.
const userAgent = "parse-companies/0.1 (+https://github.com/parse-companies)"

// Client talks to an Overpass endpoint.
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

type response struct {
	Elements []element `json:"elements"`
}

// Search builds an area-scoped query, fetches elements, maps them to companies,
// and applies the filter's absence checks. It satisfies provider.Provider.
func (c *Client) Search(ctx context.Context, r domain.Region, f domain.Filter) ([]domain.Company, error) {
	query := BuildQuery(r, f)

	body := url.Values{"data": {query}}.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint, strings.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("overpass: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", userAgent)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("overpass: do request: %w", err)
	}
	defer resp.Body.Close()

	switch {
	case resp.StatusCode == http.StatusOK:
		// continue
	case resp.StatusCode == http.StatusTooManyRequests, resp.StatusCode >= 500:
		return nil, fmt.Errorf("overpass: status %d: %w", resp.StatusCode, provider.ErrUpstreamBusy)
	default:
		return nil, fmt.Errorf("overpass: unexpected status %d", resp.StatusCode)
	}

	var out response
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("overpass: decode response: %w", err)
	}

	companies := make([]domain.Company, 0, len(out.Elements))
	for _, el := range out.Elements {
		company := mapElement(el)
		if f.Match(company) && r.Accept(company.Lat, company.Lon) {
			companies = append(companies, company)
		}
	}
	return companies, nil
}
