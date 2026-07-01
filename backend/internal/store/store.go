// Package store persists companies and searches in Postgres + PostGIS.
// It adapts between domain types and the sqlc-generated db package; nothing
// outside this package sees db.* types.
package store

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/parse-companies/backend/internal/domain"
	"github.com/parse-companies/backend/internal/store/db"
)

// Store is the database gateway.
type Store struct {
	pool *pgxpool.Pool
	q    *db.Queries
}

// SearchSummary is one row of search history.
type SearchSummary struct {
	ID          int64         `json:"id"`
	RegionName  string        `json:"regionName"`
	RegionArea  int64         `json:"regionAreaId"`
	Filter      domain.Filter `json:"filters"`
	ResultCount int           `json:"resultCount"`
	CreatedAt   time.Time     `json:"createdAt"`
}

// New opens a connection pool and returns a Store.
func New(ctx context.Context, dsn string) (*Store, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("store: connect: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("store: ping: %w", err)
	}
	return &Store{pool: pool, q: db.New(pool)}, nil
}

// Close releases the pool.
func (s *Store) Close() { s.pool.Close() }

// UpsertCompanies inserts or updates companies by OSM identity and returns their
// database ids in the same order as the input. Re-running dedups across runs.
func (s *Store) UpsertCompanies(ctx context.Context, companies []domain.Company) ([]int64, error) {
	if len(companies) == 0 {
		return nil, nil
	}

	params := make([]db.UpsertCompanyParams, len(companies))
	for i, c := range companies {
		tags, err := json.Marshal(c.Tags)
		if err != nil {
			return nil, fmt.Errorf("store: marshal tags: %w", err)
		}
		params[i] = db.UpsertCompanyParams{
			OsmType: c.OSMType, OsmID: c.OSMID, Name: c.Name,
			Category: c.Category, Subcategory: c.Subcat,
			Website: c.Website, Phone: c.Phone, Email: c.Email,
			Instagram: c.Instagram, Facebook: c.Facebook, Vk: c.VK, Telegram: c.Telegram,
			Whatsapp:    c.WhatsApp,
			AddrCountry: c.Addr.Country, AddrCity: c.Addr.City, AddrStreet: c.Addr.Street,
			AddrHousenumber: c.Addr.Housenumber, AddrPostcode: c.Addr.Postcode,
			OpeningHours: c.OpeningHours, Tags: tags,
			Lon: c.Lon, Lat: c.Lat,
		}
	}

	ids := make([]int64, len(companies))
	var batchErr error
	results := s.q.UpsertCompany(ctx, params)
	results.QueryRow(func(i int, id int64, err error) {
		if err != nil && batchErr == nil {
			batchErr = err
		}
		ids[i] = id
	})
	if err := results.Close(); err != nil {
		return nil, fmt.Errorf("store: upsert batch: %w", err)
	}
	if batchErr != nil {
		return nil, fmt.Errorf("store: upsert company: %w", batchErr)
	}
	return ids, nil
}

// CreateSearch records a search run and links it to the given company ids.
func (s *Store) CreateSearch(ctx context.Context, r domain.Region, f domain.Filter, companyIDs []int64) (int64, error) {
	filters, err := json.Marshal(f)
	if err != nil {
		return 0, fmt.Errorf("store: marshal filters: %w", err)
	}

	searchID, err := s.q.CreateSearch(ctx, db.CreateSearchParams{
		RegionName:   r.Name,
		RegionAreaID: r.OSMAreaID,
		Filters:      filters,
		ResultCount:  int32(len(companyIDs)),
	})
	if err != nil {
		return 0, fmt.Errorf("store: create search: %w", err)
	}

	if len(companyIDs) > 0 {
		links := make([]db.LinkSearchResultParams, len(companyIDs))
		for i, id := range companyIDs {
			links[i] = db.LinkSearchResultParams{SearchID: searchID, CompanyID: id}
		}
		var linkErr error
		res := s.q.LinkSearchResult(ctx, links)
		res.Exec(func(_ int, err error) {
			if err != nil && linkErr == nil {
				linkErr = err
			}
		})
		if err := res.Close(); err != nil {
			return 0, fmt.Errorf("store: link batch: %w", err)
		}
		if linkErr != nil {
			return 0, fmt.Errorf("store: link search result: %w", linkErr)
		}
	}
	return searchID, nil
}

// ListSearches returns search history, newest first.
func (s *Store) ListSearches(ctx context.Context, limit, offset int) ([]SearchSummary, error) {
	rows, err := s.q.ListSearches(ctx, db.ListSearchesParams{Lim: int32(limit), Off: int32(offset)})
	if err != nil {
		return nil, fmt.Errorf("store: list searches: %w", err)
	}
	out := make([]SearchSummary, len(rows))
	for i, r := range rows {
		out[i] = toSummary(r.ID, r.RegionName, r.RegionAreaID, r.Filters, int(r.ResultCount), r.CreatedAt.Time)
	}
	return out, nil
}

// GetSearchResults returns one search and its companies.
func (s *Store) GetSearchResults(ctx context.Context, searchID int64) (SearchSummary, []domain.Company, error) {
	srch, err := s.q.GetSearch(ctx, searchID)
	if err != nil {
		return SearchSummary{}, nil, fmt.Errorf("store: get search: %w", err)
	}
	rows, err := s.q.GetSearchCompanies(ctx, searchID)
	if err != nil {
		return SearchSummary{}, nil, fmt.Errorf("store: get search companies: %w", err)
	}

	companies := make([]domain.Company, len(rows))
	for i, r := range rows {
		companies[i] = domain.Company{
			OSMType: r.OsmType, OSMID: r.OsmID, Name: r.Name,
			Category: r.Category, Subcat: r.Subcategory,
			Website: r.Website, Phone: r.Phone, Email: r.Email,
			Instagram: r.Instagram, Facebook: r.Facebook, VK: r.Vk, Telegram: r.Telegram,
			WhatsApp: r.Whatsapp,
			Addr: domain.Address{
				Country: r.AddrCountry, City: r.AddrCity, Street: r.AddrStreet,
				Housenumber: r.AddrHousenumber, Postcode: r.AddrPostcode,
			},
			OpeningHours: r.OpeningHours, Lat: r.Lat, Lon: r.Lon,
		}
	}
	summary := toSummary(srch.ID, srch.RegionName, srch.RegionAreaID, srch.Filters, int(srch.ResultCount), srch.CreatedAt.Time)
	return summary, companies, nil
}

func toSummary(id int64, name string, area int64, filters []byte, count int, created time.Time) SearchSummary {
	var f domain.Filter
	_ = json.Unmarshal(filters, &f)
	return SearchSummary{
		ID: id, RegionName: name, RegionArea: area,
		Filter: f, ResultCount: count, CreatedAt: created,
	}
}
