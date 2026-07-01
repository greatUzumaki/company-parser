// Package service orchestrates a search across one or more data providers:
// cache lookup, concurrent provider fetch, cross-source dedup, persistence, and
// history. It depends only on interfaces so each collaborator can be faked.
package service

import (
	"context"
	"fmt"
	"time"

	"github.com/parse-companies/backend/internal/cache"
	"github.com/parse-companies/backend/internal/domain"
	"github.com/parse-companies/backend/internal/provider"
)

// Repository is the persistence surface the service needs.
type Repository interface {
	UpsertCompanies(ctx context.Context, companies []domain.Company) ([]int64, error)
	CreateSearch(ctx context.Context, r domain.Region, f domain.Filter, companyIDs []int64) (int64, error)
}

// NamedProvider pairs a provider with a stable name shown in progress events.
type NamedProvider struct {
	Name string
	P    provider.Provider
}

// Event is one message in a streaming search.
type Event struct {
	Type      string           `json:"type"` // source_start | source_done | companies | done | error
	Source    string           `json:"source,omitempty"`
	Companies []domain.Company `json:"companies,omitempty"`
	Count     int              `json:"count,omitempty"`    // running total after dedup
	Done      int              `json:"done,omitempty"`     // providers finished
	Total     int              `json:"total,omitempty"`    // providers total
	SearchID  int64            `json:"searchId,omitempty"` // set on "done"
	Message   string           `json:"message,omitempty"`
}

const cacheTTL = 24 * time.Hour

// Service runs searches across its providers.
type Service struct {
	providers []NamedProvider
	repo      Repository
	cache     cache.Cache
}

// New wires the providers and collaborators.
func New(providers []NamedProvider, repo Repository, c cache.Cache) *Service {
	return &Service{providers: providers, repo: repo, cache: c}
}

// sourceResult carries one provider's outcome off a goroutine.
type sourceResult struct {
	name      string
	companies []domain.Company
	err       error
}

// fanOut launches every provider concurrently and returns a channel of results.
func (s *Service) fanOut(ctx context.Context, r domain.Region, f domain.Filter) <-chan sourceResult {
	ch := make(chan sourceResult, len(s.providers))
	for _, np := range s.providers {
		go func(np NamedProvider) {
			companies, err := np.P.Search(ctx, r, f)
			ch <- sourceResult{name: np.Name, companies: companies, err: err}
		}(np)
	}
	return ch
}

// Search runs all providers, merges + dedups, persists, and returns the result.
// A cache hit skips the providers but still records history so it is exportable.
func (s *Service) Search(ctx context.Context, r domain.Region, f domain.Filter) (int64, []domain.Company, error) {
	key := cache.Key(r, f)

	companies, hit, err := s.cache.Get(ctx, key)
	if err != nil {
		return 0, nil, fmt.Errorf("service: cache get: %w", err)
	}
	if !hit {
		acc := newAccumulator()
		ch := s.fanOut(ctx, r, f)
		var firstErr error
		for range s.providers {
			res := <-ch
			if res.err != nil {
				if firstErr == nil {
					firstErr = res.err
				}
				continue
			}
			acc.add(res.companies)
		}
		companies = acc.all()
		// Fail only if every provider failed; a partial set is still useful but
		// a total upstream outage must surface, never masquerade as 0 results.
		if len(companies) == 0 && firstErr != nil {
			return 0, nil, fmt.Errorf("service: all providers failed: %w", firstErr)
		}
		if err := s.cache.Set(ctx, key, companies, cacheTTL); err != nil {
			return 0, nil, fmt.Errorf("service: cache set: %w", err)
		}
	}

	searchID, err := s.persist(ctx, r, f, companies)
	if err != nil {
		return 0, nil, err
	}
	return searchID, companies, nil
}

// SearchStream runs all providers concurrently and emits events as each source
// completes: per-source start/done progress plus incremental batches of newly
// found companies (already deduped against earlier sources). It persists the
// merged set and emits a final "done" event with the search id.
func (s *Service) SearchStream(ctx context.Context, r domain.Region, f domain.Filter, emit func(Event) error) error {
	total := len(s.providers)
	if err := emitAll(emit, sourceStarts(s.providers, total)); err != nil {
		return err
	}

	acc := newAccumulator()
	ch := s.fanOut(ctx, r, f)
	var firstErr error
	done := 0
	for range s.providers {
		res := <-ch
		done++
		if res.err != nil {
			if firstErr == nil {
				firstErr = res.err
			}
			if err := emit(Event{Type: "source_done", Source: res.name, Done: done, Total: total, Message: "failed"}); err != nil {
				return err
			}
			continue
		}
		fresh := acc.add(res.companies)
		if err := emit(Event{Type: "companies", Source: res.name, Companies: fresh, Count: len(acc.all())}); err != nil {
			return err
		}
		if err := emit(Event{Type: "source_done", Source: res.name, Done: done, Total: total, Count: len(acc.all())}); err != nil {
			return err
		}
	}

	companies := acc.all()
	if len(companies) == 0 && firstErr != nil {
		_ = emit(Event{Type: "error", Message: "all providers failed"})
		return fmt.Errorf("service: all providers failed: %w", firstErr)
	}

	_ = s.cache.Set(ctx, cache.Key(r, f), companies, cacheTTL)

	searchID, err := s.persist(ctx, r, f, companies)
	if err != nil {
		_ = emit(Event{Type: "error", Message: "persist failed"})
		return err
	}
	return emit(Event{Type: "done", SearchID: searchID, Count: len(companies)})
}

// persist upserts the companies and records the search run.
func (s *Service) persist(ctx context.Context, r domain.Region, f domain.Filter, companies []domain.Company) (int64, error) {
	ids, err := s.repo.UpsertCompanies(ctx, companies)
	if err != nil {
		return 0, fmt.Errorf("service: persist companies: %w", err)
	}
	searchID, err := s.repo.CreateSearch(ctx, r, f, ids)
	if err != nil {
		return 0, fmt.Errorf("service: record search: %w", err)
	}
	return searchID, nil
}

func sourceStarts(providers []NamedProvider, total int) []Event {
	events := make([]Event, len(providers))
	for i, np := range providers {
		events[i] = Event{Type: "source_start", Source: np.Name, Total: total}
	}
	return events
}

func emitAll(emit func(Event) error, events []Event) error {
	for _, e := range events {
		if err := emit(e); err != nil {
			return err
		}
	}
	return nil
}
