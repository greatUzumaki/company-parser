package api

import (
	"github.com/parse-companies/backend/internal/campaign"
	"github.com/parse-companies/backend/internal/domain"
	"github.com/parse-companies/backend/internal/store"
)

// searchRequest is the POST /api/v1/search body. Force skips the cache and
// re-queries the providers (used by the manual refresh button).
type searchRequest struct {
	Region  domain.Region `json:"region"`
	Filters domain.Filter `json:"filters"`
	Force   bool          `json:"force"`
}

// searchResponse is returned after a search runs.
type searchResponse struct {
	SearchID int64            `json:"searchId"`
	Count    int              `json:"count"`
	Results  []domain.Company `json:"results"`
}

// searchDetailResponse is GET /api/v1/searches/{id}.
type searchDetailResponse struct {
	Search  store.SearchSummary `json:"search"`
	Results []domain.Company    `json:"results"`
}

// campaignRequest is the POST /api/v1/campaign/send body. Recipients are the
// client-filtered targets; Confirm is the consent acknowledgement; DryRun
// previews without sending.
type campaignRequest struct {
	Subject    string               `json:"subject"`
	Body       string               `json:"body"`
	Recipients []campaign.Recipient `json:"recipients"`
	DryRun     bool                 `json:"dryRun"`
	Confirm    bool                 `json:"confirm"`
}

// category is one selectable OSM business category.
type category struct {
	Key   string `json:"key"`
	Label string `json:"label"`
}

// categories is the fixed whitelist exposed to the UI.
var categories = []category{
	{"shop", "Shops & retail"},
	{"amenity", "Amenities (cafés, clinics, banks…)"},
	{"office", "Offices"},
	{"craft", "Craft & trades"},
	{"tourism", "Tourism (hotels, attractions…)"},
}
