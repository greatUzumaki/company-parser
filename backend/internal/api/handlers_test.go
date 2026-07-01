package api

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"

	"github.com/parse-companies/backend/internal/campaign"
	"github.com/parse-companies/backend/internal/domain"
	"github.com/parse-companies/backend/internal/service"
	"github.com/parse-companies/backend/internal/store"
)

type fakeSearcher struct {
	id        int64
	companies []domain.Company
}

func (f fakeSearcher) Search(context.Context, domain.Region, domain.Filter) (int64, []domain.Company, error) {
	return f.id, f.companies, nil
}

func (f fakeSearcher) SearchStream(_ context.Context, _ domain.Region, _ domain.Filter, _ bool, emit func(service.Event) error) error {
	if err := emit(service.Event{Type: "source_start", Source: "Test", Total: 1}); err != nil {
		return err
	}
	if err := emit(service.Event{Type: "companies", Source: "Test", Companies: f.companies, Count: len(f.companies)}); err != nil {
		return err
	}
	return emit(service.Event{Type: "done", SearchID: f.id, Count: len(f.companies)})
}

type fakeRepo struct {
	results []domain.Company
	getErr  error
}

func (f fakeRepo) ListSearches(context.Context, int, int) ([]store.SearchSummary, error) {
	return []store.SearchSummary{{ID: 1, RegionName: "X", ResultCount: 2}}, nil
}
func (f fakeRepo) GetSearchResults(context.Context, int64) (store.SearchSummary, []domain.Company, error) {
	if f.getErr != nil {
		return store.SearchSummary{}, nil, f.getErr
	}
	return store.SearchSummary{ID: 7}, f.results, nil
}

type fakeRegions struct{}

func (fakeRegions) Search(context.Context, string) ([]domain.Region, error) {
	return []domain.Region{{Name: "Bavaria", OSMAreaID: 3600002145}}, nil
}

type fakeCampaign struct{ enabled bool }

func (f fakeCampaign) Enabled() bool { return f.enabled }
func (f fakeCampaign) Send(_ context.Context, _, _ string, recipients []campaign.Recipient, dryRun bool, emit func(campaign.Event) error) error {
	if err := emit(campaign.Event{Type: "start", Total: len(recipients), DryRun: dryRun}); err != nil {
		return err
	}
	return emit(campaign.Event{Type: "done", Sent: len(recipients), DryRun: dryRun})
}

func testServer(repo Repo) http.Handler {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	s := NewServer(
		fakeSearcher{id: 42, companies: []domain.Company{{OSMID: "1", Name: "A"}}},
		repo, fakeRegions{}, fakeCampaign{enabled: true}, logger,
	)
	return s.Routes("*")
}

func TestSearchHandler(t *testing.T) {
	srv := testServer(fakeRepo{})
	body := `{"region":{"name":"B","osmAreaId":3600002145},"filters":{"noWebsite":true}}`
	req := httptest.NewRequest("POST", "/api/v1/search", strings.NewReader(body))
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Fatalf("status = %d, body=%s", rec.Code, rec.Body.String())
	}
	var resp searchResponse
	json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp.SearchID != 42 || resp.Count != 1 {
		t.Errorf("resp = %+v", resp)
	}
}

func TestSearchStreamHandler(t *testing.T) {
	srv := testServer(fakeRepo{})
	body := `{"region":{"name":"B","osmAreaId":3600002145},"filters":{}}`
	req := httptest.NewRequest("POST", "/api/v1/search/stream", strings.NewReader(body))
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Fatalf("status = %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/x-ndjson" {
		t.Errorf("content-type = %q", ct)
	}
	lines := strings.Split(strings.TrimSpace(rec.Body.String()), "\n")
	if len(lines) != 3 {
		t.Fatalf("got %d ndjson lines, want 3:\n%s", len(lines), rec.Body.String())
	}
	var last service.Event
	if err := json.Unmarshal([]byte(lines[len(lines)-1]), &last); err != nil {
		t.Fatalf("last line not json: %v", err)
	}
	if last.Type != "done" || last.SearchID != 42 {
		t.Errorf("final event = %+v", last)
	}
}

func TestSearchHandlerRejectsEmptyRegion(t *testing.T) {
	srv := testServer(fakeRepo{})
	req := httptest.NewRequest("POST", "/api/v1/search", strings.NewReader(`{"filters":{}}`))
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)
	if rec.Code != 400 {
		t.Errorf("status = %d, want 400", rec.Code)
	}
}

func TestExportUnknownSearchReturns404(t *testing.T) {
	srv := testServer(fakeRepo{getErr: pgx.ErrNoRows})
	req := httptest.NewRequest("GET", "/api/v1/searches/999/export?format=csv", nil)
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)
	if rec.Code != 404 {
		t.Errorf("status = %d, want 404", rec.Code)
	}
}

func TestExportBadFormatReturns400(t *testing.T) {
	srv := testServer(fakeRepo{})
	req := httptest.NewRequest("GET", "/api/v1/searches/7/export?format=pdf", nil)
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)
	if rec.Code != 400 {
		t.Errorf("status = %d, want 400", rec.Code)
	}
}

func TestExportCSVStreamsHeader(t *testing.T) {
	srv := testServer(fakeRepo{results: []domain.Company{{OSMID: "1", Name: "A"}}})
	req := httptest.NewRequest("GET", "/api/v1/searches/7/export?format=csv", nil)
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Fatalf("status = %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "text/csv" {
		t.Errorf("content-type = %q", ct)
	}
	if !bytes.HasPrefix(rec.Body.Bytes(), []byte("osmType,osmId,name,")) {
		t.Errorf("missing csv header: %s", rec.Body.String())
	}
}

func TestCategoriesAndRegions(t *testing.T) {
	srv := testServer(fakeRepo{})

	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, httptest.NewRequest("GET", "/api/v1/categories", nil))
	if rec.Code != 200 || !strings.Contains(rec.Body.String(), "shop") {
		t.Errorf("categories: %d %s", rec.Code, rec.Body.String())
	}

	rec = httptest.NewRecorder()
	srv.ServeHTTP(rec, httptest.NewRequest("GET", "/api/v1/regions?q=bav", nil))
	if rec.Code != 200 || !strings.Contains(rec.Body.String(), "Bavaria") {
		t.Errorf("regions: %d %s", rec.Code, rec.Body.String())
	}
}

func TestCampaignStatus(t *testing.T) {
	srv := testServer(fakeRepo{})
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, httptest.NewRequest("GET", "/api/v1/campaign/status", nil))
	if rec.Code != 200 || !strings.Contains(rec.Body.String(), `"enabled":true`) {
		t.Errorf("status: %d %s", rec.Code, rec.Body.String())
	}
}

func TestCampaignRequiresConfirm(t *testing.T) {
	srv := testServer(fakeRepo{})
	body := `{"recipients":[{"email":"a@b.com","name":"A"}],"subject":"hi","body":"hello","confirm":false}`
	req := httptest.NewRequest("POST", "/api/v1/campaign/send", strings.NewReader(body))
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)
	if rec.Code != 400 {
		t.Errorf("without confirm status = %d, want 400", rec.Code)
	}
}

func TestCampaignSendStreams(t *testing.T) {
	srv := testServer(fakeRepo{})
	body := `{"recipients":[{"email":"a@b.com","name":"A"}],"subject":"hi","body":"hello {{name}}","dryRun":true,"confirm":true}`
	req := httptest.NewRequest("POST", "/api/v1/campaign/send", strings.NewReader(body))
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Fatalf("status = %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/x-ndjson" {
		t.Errorf("content-type = %q", ct)
	}
	if !strings.Contains(rec.Body.String(), `"type":"done"`) {
		t.Errorf("missing done event: %s", rec.Body.String())
	}
}

func TestListSearches(t *testing.T) {
	srv := testServer(fakeRepo{})
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, httptest.NewRequest("GET", "/api/v1/searches", nil))
	if rec.Code != 200 || !strings.Contains(rec.Body.String(), `"regionName":"X"`) {
		t.Errorf("list: %d %s", rec.Code, rec.Body.String())
	}
}
