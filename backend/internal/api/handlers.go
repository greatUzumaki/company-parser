package api

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/parse-companies/backend/internal/campaign"
	"github.com/parse-companies/backend/internal/domain"
	"github.com/parse-companies/backend/internal/export"
	"github.com/parse-companies/backend/internal/service"
	"github.com/parse-companies/backend/internal/store"
)

// Searcher runs and persists a search, both in one shot and as a live stream.
// Implemented by *service.Service.
type Searcher interface {
	Search(ctx context.Context, r domain.Region, f domain.Filter) (int64, []domain.Company, error)
	SearchStream(ctx context.Context, r domain.Region, f domain.Filter, force bool, emit func(service.Event) error) error
}

// Repo reads search history. Implemented by *store.Store.
type Repo interface {
	ListSearches(ctx context.Context, limit, offset int) ([]store.SearchSummary, error)
	GetSearchResults(ctx context.Context, searchID int64) (store.SearchSummary, []domain.Company, error)
}

// RegionFinder resolves region names. Implemented by *nominatim.Client.
type RegionFinder interface {
	Search(ctx context.Context, query string) ([]domain.Region, error)
}

// CampaignSender sends an email campaign. Implemented by *campaign.Service.
type CampaignSender interface {
	Enabled() bool
	Send(ctx context.Context, subject, body string, recipients []campaign.Recipient, dryRun bool, emit func(campaign.Event) error) error
}

// Server holds handler dependencies.
type Server struct {
	search   Searcher
	repo     Repo
	regions  RegionFinder
	campaign CampaignSender
	logger   *slog.Logger
}

// NewServer builds the API server.
func NewServer(s Searcher, repo Repo, regions RegionFinder, camp CampaignSender, logger *slog.Logger) *Server {
	return &Server{search: s, repo: repo, regions: regions, campaign: camp, logger: logger}
}

func (s *Server) handleSearch(w http.ResponseWriter, r *http.Request) {
	var req searchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}
	if req.Region.OSMAreaID == 0 && req.Region.BBox == ([4]float64{}) && !req.Region.HasPolygon() {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "region required"})
		return
	}

	id, companies, err := s.search.Search(r.Context(), req.Region, req.Filters)
	if err != nil {
		writeError(w, s.logger, err)
		return
	}
	writeJSON(w, http.StatusOK, searchResponse{SearchID: id, Count: len(companies), Results: companies})
}

// handleSearchStream runs the search and streams newline-delimited JSON events
// (one per line) so the client can plot companies and show progress live.
func (s *Server) handleSearchStream(w http.ResponseWriter, r *http.Request) {
	var req searchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}
	if req.Region.OSMAreaID == 0 && req.Region.BBox == ([4]float64{}) && !req.Region.HasPolygon() {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "region required"})
		return
	}

	w.Header().Set("Content-Type", "application/x-ndjson")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("X-Accel-Buffering", "no") // disable proxy buffering

	flusher, _ := w.(http.Flusher)
	enc := json.NewEncoder(w)

	emit := func(e service.Event) error {
		if err := enc.Encode(e); err != nil { // Encode writes a trailing newline
			return err
		}
		if flusher != nil {
			flusher.Flush()
		}
		return nil
	}

	if err := s.search.SearchStream(r.Context(), req.Region, req.Filters, req.Force, emit); err != nil {
		// Headers/body already streaming; the error event was emitted by the
		// service. Just log here.
		s.logger.Error("search stream", "err", err)
	}
}

func (s *Server) handleListSearches(w http.ResponseWriter, r *http.Request) {
	limit := atoiDefault(r.URL.Query().Get("limit"), 50)
	offset := atoiDefault(r.URL.Query().Get("offset"), 0)
	list, err := s.repo.ListSearches(r.Context(), limit, offset)
	if err != nil {
		writeError(w, s.logger, err)
		return
	}
	writeJSON(w, http.StatusOK, list)
}

func (s *Server) handleGetSearch(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}
	summary, companies, err := s.repo.GetSearchResults(r.Context(), id)
	if err != nil {
		writeError(w, s.logger, err)
		return
	}
	writeJSON(w, http.StatusOK, searchDetailResponse{Search: summary, Results: companies})
}

func (s *Server) handleExport(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}
	enc, err := export.For(r.URL.Query().Get("format"))
	if err != nil {
		writeError(w, s.logger, err)
		return
	}
	_, companies, err := s.repo.GetSearchResults(r.Context(), id)
	if err != nil {
		writeError(w, s.logger, err)
		return
	}

	w.Header().Set("Content-Type", enc.ContentType())
	w.Header().Set("Content-Disposition", "attachment; filename=companies-"+strconv.FormatInt(id, 10)+"."+enc.Ext())
	if err := enc.Encode(w, companies); err != nil {
		// Headers are already sent; log and stop. Cannot change status now.
		s.logger.Error("export encode", "err", err)
	}
}

func (s *Server) handleRegions(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	if q == "" {
		writeJSON(w, http.StatusOK, []domain.Region{})
		return
	}
	regions, err := s.regions.Search(r.Context(), q)
	if err != nil {
		writeError(w, s.logger, err)
		return
	}
	writeJSON(w, http.StatusOK, regions)
}

func (s *Server) handleCategories(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, categories)
}

// handleCampaignStatus tells the UI whether email campaigns are available.
func (s *Server) handleCampaignStatus(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]bool{"enabled": s.campaign.Enabled()})
}

// handleCampaignSend streams the progress of an email campaign as NDJSON.
// The caller MUST set confirm=true (consent acknowledgement); dryRun previews
// without sending.
func (s *Server) handleCampaignSend(w http.ResponseWriter, r *http.Request) {
	var req campaignRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}
	if !s.campaign.Enabled() {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "email sending is not configured"})
		return
	}
	if !req.Confirm {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "consent confirmation required"})
		return
	}
	if len(req.Recipients) == 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "no recipients"})
		return
	}

	w.Header().Set("Content-Type", "application/x-ndjson")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("X-Accel-Buffering", "no")

	flusher, _ := w.(http.Flusher)
	enc := json.NewEncoder(w)
	emit := func(e campaign.Event) error {
		if err := enc.Encode(e); err != nil {
			return err
		}
		if flusher != nil {
			flusher.Flush()
		}
		return nil
	}

	if err := s.campaign.Send(r.Context(), req.Subject, req.Body, req.Recipients, req.DryRun, emit); err != nil {
		s.logger.Error("campaign send", "err", err)
		_ = emit(campaign.Event{Type: "error", Message: "campaign failed"})
	}
}

func atoiDefault(s string, def int) int {
	if v, err := strconv.Atoi(s); err == nil && v >= 0 {
		return v
	}
	return def
}
