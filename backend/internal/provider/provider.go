// Package provider defines the data-source abstraction. The rest of the backend
// depends on the Provider interface, never on a concrete source. Adding a new
// source (e.g. an official Places API) means implementing this one method.
package provider

import (
	"context"
	"errors"

	"github.com/parse-companies/backend/internal/domain"
)

// ErrUpstreamBusy signals the data source is rate-limited or temporarily
// unavailable (HTTP 429/5xx / timeout). The API maps it to 503 so callers can
// retry — it is never a silent empty result.
var ErrUpstreamBusy = errors.New("provider: upstream busy")

// Provider fetches companies for a region, already filtered.
type Provider interface {
	Search(ctx context.Context, r domain.Region, f domain.Filter) ([]domain.Company, error)
}
