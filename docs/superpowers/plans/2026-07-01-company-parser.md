# Company Parser Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Web app to find companies by region + filters from OpenStreetMap, select region on a map, export to JSON/CSV/Excel.

**Architecture:** Hexagonal Go backend behind a `Provider` interface (Overpass impl), Postgres+PostGIS persistence, Redis cache, Next.js 15 + MapLibre frontend. Region = clicked admin boundary; filters = OSM tag absence checks.

**Tech Stack:** Go 1.23+, pgx/v5, sqlc, golang-migrate, excelize, slog, errgroup; Next.js 15, TypeScript, Tailwind, shadcn/ui, MapLibre GL, TanStack Query+Table, zustand.

## Global Constraints

- Data source: **OpenStreetMap / Overpass API only**. No Google/Yandex/2GIS scraping. ODbL attribution required in UI footer + README.
- Go module path: `github.com/parse-companies/backend`. Go 1.23+.
- All exports stream through `io.Writer` — no full in-memory buffering of result sets.
- No silent fallbacks: upstream/DB errors propagate as typed errors → mapped HTTP codes.
- DB + Redis already run in the user's Docker; backend connects via env (`DATABASE_URL`, `REDIS_URL`). **Verify Postgres has the PostGIS extension before Phase 3.**
- Tests: table-driven, TDD red→green→refactor, commit per task.

---

## Phase 1 — Scaffold + Claude config

### Task 1: Repo skeleton + Makefile + env

**Files:**
- Create: `backend/go.mod`, `backend/cmd/server/main.go` (stub), `Makefile`, `.env.example`, `.gitignore`, `README.md`
- Create: `docker-compose.yml` (app services only; references existing PG/Redis via env or external network)

**Steps:**
- [ ] **Step 1:** `cd backend && go mod init github.com/parse-companies/backend`
- [ ] **Step 2:** Write `cmd/server/main.go` stub that reads config and logs "starting", exits 0.
- [ ] **Step 3:** `.env.example`:
```
DATABASE_URL=postgres://user:pass@localhost:5432/parse_companies?sslmode=disable
REDIS_URL=redis://localhost:6379/0
OVERPASS_URL=https://overpass-api.de/api/interpreter
NOMINATIM_URL=https://nominatim.openstreetmap.org
HTTP_ADDR=:8080
```
- [ ] **Step 4:** `Makefile` targets: `dev`, `build`, `test`, `lint`, `migrate-up`, `migrate-down`, `sqlc`.
- [ ] **Step 5:** `go build ./...` succeeds. Commit `chore: scaffold backend module + tooling`.

### Task 2: CLAUDE.md

**Files:** Create `CLAUDE.md`

Content: project one-liner; stack; make targets; architecture map (the `backend/internal/*` tree and what each package owns); conventions (sqlc, table-driven tests, error wrapping with `%w`, no silent fallbacks, files focused/small); **legal note** (OSM = ODbL, attribution required, no map scraping).

- [ ] **Step 1:** Write `CLAUDE.md` per above.
- [ ] **Step 2:** Commit `docs: add CLAUDE.md`.

### Task 3: Subagents

**Files:** Create `.claude/agents/overpass-expert.md`, `go-backend-architect.md`, `nextjs-map-frontend.md`, `export-specialist.md`

Each: frontmatter (`name`, `description` with trigger examples, `tools`) + system prompt scoped to its domain. (Full content in Phase 1 execution.)

- [ ] **Step 1:** Write 4 agent files.
- [ ] **Step 2:** Commit `chore: add project subagents`.

### Task 4: Skills

**Files:** Create `.claude/skills/overpass-query/SKILL.md`, `add-data-provider/SKILL.md`, `db-migration/SKILL.md`

- [ ] **Step 1:** Write 3 skill files (frontmatter `name`+`description`, body = recipe).
- [ ] **Step 2:** Commit `chore: add project skills`.

---

## Phase 2 — Backend core (domain + Overpass provider)

### Task 5: Domain types

**Files:** Create `internal/domain/company.go`, `region.go`, `filter.go`; Test `internal/domain/filter_test.go`

**Produces:**
```go
type Company struct {
    OSMType, OSMID            string
    Name, Category, Subcat    string
    Website, Phone, Email     string
    Instagram, Facebook, VK, Telegram string
    Addr                      Address
    Lat, Lon                  float64
    OpeningHours              string
    Tags                      map[string]string
}
type Address struct{ Country, City, Street, Housenumber, Postcode string }
type Region struct{ Name string; OSMAreaID int64; BBox [4]float64 } // minLon,minLat,maxLon,maxLat
type Filter struct{ NoWebsite, NoSocials, NoPhone bool; Categories []string }
func (f Filter) Match(c Company) bool
```

- [ ] **Step 1:** Write `filter_test.go` table-driven: company with website + `NoWebsite=true` → false; company without any web tag + `NoWebsite=true` → true; socials/phone analogues; category whitelist match.
- [ ] **Step 2:** Run `go test ./internal/domain/` → FAIL (Match undefined).
- [ ] **Step 3:** Implement types + `Match` (absence logic per spec §8).
- [ ] **Step 4:** Run tests → PASS.
- [ ] **Step 5:** Commit `feat(domain): company/region/filter types + filter logic`.

### Task 6: Provider interface + Overpass QL builder

**Files:** Create `internal/provider/provider.go`, `internal/provider/overpass/query.go`; Test `query_test.go`

**Produces:**
```go
// provider.go
type Provider interface {
    Search(ctx context.Context, r domain.Region, f domain.Filter) ([]domain.Company, error)
}
// overpass/query.go
func BuildQuery(r domain.Region, f domain.Filter) string
```
QL: `[out:json][timeout:180]; area(<3600000000+osmAreaId>)->.a; ( node["shop"](area.a); ... ); out center tags;` — one `node/way` clause per category key, area-scoped. (Absence filters applied post-fetch.)

- [ ] **Step 1:** Write `query_test.go`: assert output contains `area(3600002145)` for areaId `2145`, contains `node["shop"](area.a)` when categories include `shop`, contains `out center tags`.
- [ ] **Step 2:** Run → FAIL.
- [ ] **Step 3:** Implement `BuildQuery` (area id offset rule: Overpass area = `3600000000 + relationId`; if `OSMAreaID` already > 3.6e9 use as-is).
- [ ] **Step 4:** Tests PASS.
- [ ] **Step 5:** Commit `feat(overpass): area-scoped QL builder`.

### Task 7: OSM tag → Company mapping

**Files:** Create `internal/provider/overpass/mapping.go`; Test `mapping_test.go`

**Produces:** `func mapElement(el overpassElement) domain.Company` + `type overpassElement struct{ Type string; ID int64; Lat,Lon float64; Center *struct{Lat,Lon float64}; Tags map[string]string }`

Mapping rules: `Website` ← `website` || `contact:website`; `Phone` ← `phone` || `contact:phone`; socials ← `contact:instagram` etc; `Category` ← first present of whitelist keys, `Subcat` ← its value; `Addr.*` ← `addr:*`; `Lat/Lon` ← node coords or `center`.

- [ ] **Step 1:** `mapping_test.go`: element with `contact:website` only → `Website` set; way with `center` → coords from center; `shop=bakery` → Category `shop`, Subcat `bakery`.
- [ ] **Step 2:** Run → FAIL.
- [ ] **Step 3:** Implement `mapElement`.
- [ ] **Step 4:** PASS.
- [ ] **Step 5:** Commit `feat(overpass): OSM tag → Company mapping`.

### Task 8: Overpass HTTP client

**Files:** Create `internal/provider/overpass/client.go`; Test `client_test.go` (httptest)

**Produces:** `func New(httpClient *http.Client, endpoint string) *Client` implementing `provider.Provider`. Parses `{elements:[...]}`, maps each, returns companies. On HTTP 429/504 → `provider.ErrUpstreamBusy` (define in provider.go). Filters applied here via `f.Match`.

- [ ] **Step 1:** `client_test.go` with `httptest.Server` returning fixed JSON → assert N companies, filtered correctly; 429 → `ErrUpstreamBusy`.
- [ ] **Step 2:** Run → FAIL.
- [ ] **Step 3:** Implement client (build query, POST `data=`, decode, map, filter).
- [ ] **Step 4:** PASS.
- [ ] **Step 5:** Commit `feat(overpass): http client + provider impl`.

---

## Phase 3 — Persistence (PostGIS + sqlc)

> **Pre-task:** confirm PostGIS available: `psql $DATABASE_URL -c "CREATE EXTENSION IF NOT EXISTS postgis;"`. If it fails, switch the user's PG image to `postgis/postgis` (note in README) before continuing.

### Task 9: Migrations

**Files:** Create `migrations/0001_init.up.sql`, `0001_init.down.sql`

Schema per spec §9: `companies` (with `geom geometry(Point,4326)`, GIST + GIN indexes, unique `(osm_type,osm_id)`), `searches`, `search_results`.

- [ ] **Step 1:** Write up/down SQL.
- [ ] **Step 2:** `make migrate-up` succeeds against the dev DB.
- [ ] **Step 3:** Commit `feat(db): initial PostGIS schema`.

### Task 10: sqlc repository

**Files:** Create `sqlc.yaml`, `internal/store/queries.sql`, generated `internal/store/db/*`; wrapper `internal/store/store.go`; Test `internal/store/store_test.go` (integration)

**Produces:**
```go
type Store struct{ /* pgxpool */ }
func New(ctx, dsn string) (*Store, error)
func (s *Store) UpsertCompanies(ctx, []domain.Company) ([]int64, error) // upsert on (osm_type,osm_id)
func (s *Store) CreateSearch(ctx, region domain.Region, f domain.Filter, companyIDs []int64) (searchID int64, err error)
func (s *Store) ListSearches(ctx, limit, offset int) ([]SearchSummary, error)
func (s *Store) GetSearchResults(ctx, searchID int64) (SearchSummary, []domain.Company, error)
```

- [ ] **Step 1:** Write queries.sql (upsert with `ST_SetSRID(ST_MakePoint(lon,lat),4326)`, insert search, link results, list, fetch).
- [ ] **Step 2:** `sqlc generate`.
- [ ] **Step 3:** `store_test.go`: upsert 2 companies → 2 ids; upsert same again → same ids (dedup); create search → fetch returns both.
- [ ] **Step 4:** Run integration test → PASS.
- [ ] **Step 5:** Commit `feat(store): pgx/sqlc repository with dedup upsert`.

---

## Phase 4 — Service + API + cache

### Task 11: Redis cache wrapper

**Files:** Create `internal/cache/redis.go`; Test `redis_test.go` (miniredis)

**Produces:** `type Cache interface{ Get(ctx,key string)([]domain.Company,bool,error); Set(ctx,key string,v []domain.Company,ttl time.Duration) error }` + `redisCache` impl + `func Key(r domain.Region, f domain.Filter) string` (stable hash).

- [ ] Steps: failing test (set→get roundtrip, miss returns false) → implement → pass → commit `feat(cache): redis result cache`.

### Task 12: Search service

**Files:** Create `internal/service/search.go`; Test `search_test.go` (fake provider + fake cache + fake store)

**Produces:** `func New(p provider.Provider, s *store.Store, c cache.Cache) *Service` + `func (svc *Service) Search(ctx, r domain.Region, f domain.Filter) (searchID int64, companies []domain.Company, err error)`. Flow: cache check → provider.Search → dedup (already unique by OSM id) → store.UpsertCompanies → store.CreateSearch → cache.Set.

- [ ] Steps: failing test (cache hit skips provider; miss calls provider then persists) → implement → pass → commit `feat(service): search orchestration`.

### Task 13: HTTP API

**Files:** Create `internal/api/router.go`, `handlers.go`, `dto.go`, `errors.go`; Test `handlers_test.go`. Modify `cmd/server/main.go` (wire everything).

**Produces:** routes per spec §7. DTO mapping domain↔JSON. Error mapping: `ErrUpstreamBusy`→503, not-found→404, bad format→400, validation→400.

- [ ] Steps: failing handler test (`POST /api/v1/search` with fake service returns 200 + body; unknown export id → 404) → implement router+handlers+main wiring → pass → commit `feat(api): http handlers + server wiring`.

### Task 14: Regions + categories endpoints

**Files:** Modify `internal/api/handlers.go`; Create `internal/provider/nominatim/client.go`; Test.

**Produces:** `GET /api/v1/regions?q=` proxies Nominatim (respect usage policy: User-Agent header, cache in Redis); `GET /api/v1/categories` returns static whitelist with labels.

- [ ] Steps: failing test → implement → pass → commit `feat(api): regions autocomplete + categories`.

---

## Phase 5 — Export (streaming)

### Task 15: Export encoders

**Files:** Create `internal/export/export.go` (column order + `Encoder` interface), `json.go`, `csv.go`, `xlsx.go`; Test `export_test.go` (golden files in `testdata/`)

**Produces:**
```go
type Encoder interface{ Encode(w io.Writer, companies []domain.Company) error; ContentType() string; Ext() string }
func For(format string) (Encoder, error) // json|csv|xlsx, else ErrBadFormat
```
Columns per spec §7 "Company record". CSV via `encoding/csv` row stream; JSON via `json.Encoder`; XLSX via `excelize` `StreamWriter`.

- [ ] **Step 1:** Golden tests: 2-company fixture → exact CSV bytes, valid JSON array, xlsx with header + 2 rows.
- [ ] **Step 2:** Run → FAIL.
- [ ] **Step 3:** Implement encoders.
- [ ] **Step 4:** PASS.
- [ ] **Step 5:** Commit `feat(export): streaming json/csv/xlsx encoders`.

### Task 16: Export endpoint wiring

**Files:** Modify `internal/api/handlers.go`; Test.

**Produces:** `GET /api/v1/searches/{id}/export?format=` → fetch results from store, pick `export.For(format)`, set `Content-Type` + `Content-Disposition`, stream `Encode(w, companies)`.

- [ ] Steps: failing test (csv export of seeded search → header row present) → implement → pass → commit `feat(api): export endpoint`.

---

## Phase 6 — Frontend (Next.js + MapLibre)

### Task 17: Next.js app scaffold

**Files:** `frontend/` via `npx create-next-app@latest frontend --ts --tailwind --app --eslint --src-dir --use-npm --yes`; init shadcn/ui; install `maplibre-gl @tanstack/react-query @tanstack/react-table zustand`.

- [ ] Steps: scaffold → `npm run build` ok → commit `chore(frontend): scaffold next.js + deps`.

### Task 18: API client + types

**Files:** Create `frontend/src/lib/api.ts`, `frontend/src/lib/types.ts`

**Produces:** typed `searchCompanies(region,filters)`, `listSearches()`, `exportUrl(id,format)`, `searchRegions(q)`, `getCategories()`. Types mirror backend DTOs.

- [ ] Steps: write client + types → typecheck → commit `feat(frontend): api client + types`.

### Task 19: Map with boundary click

**Files:** Create `frontend/src/components/RegionMap.tsx`, `frontend/src/store/useSearch.ts`

**Produces:** MapLibre map (OSM tiles, no token), admin-boundary GeoJSON source + fill/line layers, `click` on a feature → set selected region `{name, osmAreaId, bbox}` in zustand. Source: country-level GeoJSON bundled; region drill via Nominatim lookup on click fallback.

- [ ] Steps: render map, wire click → store update → commit `feat(frontend): region map with boundary selection`.

### Task 20: Filter panel + results table + export bar + history

**Files:** Create `frontend/src/components/FilterPanel.tsx`, `ResultsTable.tsx`, `ExportBar.tsx`, `HistoryList.tsx`; wire in `frontend/src/app/page.tsx`

**Produces:** filter toggles (no website/socials/phone) + category multiselect → bound to store; "Search" → TanStack Query mutation → ResultsTable (TanStack Table, paginated); ExportBar with JSON/CSV/Excel buttons → `window.location = exportUrl(...)`; HistoryList from `listSearches()`. ODbL attribution in footer.

- [ ] Steps: build components, wire page, manual smoke → commit `feat(frontend): filters, results, export, history`.

---

## Phase 7 — Integration + polish

### Task 21: End-to-end + compose + docs

**Files:** Modify `docker-compose.yml` (app + frontend, external PG/Redis), `README.md`

- [ ] **Step 1:** `docker compose up` brings backend+frontend; backend reaches existing PG/Redis.
- [ ] **Step 2:** Manual e2e: select region → search → results → export each format opens valid file.
- [ ] **Step 3:** README: setup, env, make targets, ODbL attribution, legal note.
- [ ] **Step 4:** Commit `docs: README + compose for full stack`.

---

## Self-Review

- **Spec coverage:** §2 scope → Tasks 5–21; §4 architecture → package layout in Tasks 5–15; §7 API → Tasks 13,14,16; §8 filters → Task 5; §9 schema → Task 9; §10 errors → Tasks 8,13; §11 testing → every task TDD; §12 Claude setup → Tasks 2–4. No gaps.
- **Placeholders:** none — each task names exact files, signatures, and test assertions.
- **Type consistency:** `Provider.Search`, `Filter.Match`, `Store.UpsertCompanies/CreateSearch/GetSearchResults`, `Cache.Get/Set`, `Encoder.Encode` used identically across Tasks 5–16.
