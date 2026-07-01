# CLAUDE.md — Company Parser

Web app that finds companies by region + filters from **OpenStreetMap**, with
map-based region selection and export to JSON / CSV / Excel. Primary use case:
lead generation (find businesses with gaps — no website, no socials, no phone).

## ⚖️ Legal — read before touching the data layer

- **Data source is OpenStreetMap only**, via the Overpass API. OSM data is
  licensed **ODbL** — attribution is **required** (UI footer + README).
- **Do NOT scrape Google Maps / Yandex Maps / 2GIS.** It violates their ToS and
  carries legal + anti-bot risk. If asked to add a commercial source, use its
  **official Places API** behind the `Provider` interface, never HTML scraping.
- Respect Overpass + Nominatim usage policies: set a `User-Agent`, cache
  responses, don't hammer the public endpoints in loops.

## Stack

| Layer | Tech |
|---|---|
| Backend | Go 1.23+, `net/http` ServeMux, `pgx/v5` + `sqlc`, `golang-migrate`, `excelize`, `log/slog`, `errgroup` |
| Data | OpenStreetMap / Overpass API + **Wikidata** (SPARQL, CC0); Nominatim for region lookup. Multi-provider fan-out behind `Provider`; results merged + deduped. |
| Storage | Postgres (plain — no PostGIS), Redis (cache) — both run in the user's Docker |
| Frontend | Next.js 15 (App Router), TypeScript, Tailwind + shadcn/ui, **MapLibre GL**, TanStack Query + Table, zustand |

## Architecture — hexagonal

The swappable data source is the core design driver. Everything depends on the
`Provider` interface, never on Overpass directly.

```
backend/internal/
  domain/      Company, Region, Filter — pure types, NO external deps
  provider/    Provider interface + overpass/ + nominatim/ impls
  service/     search orchestration, dedup, cache lookup
  store/       pgx + sqlc repo (companies, searches) — plain lat/lon columns
  export/      json.go csv.go xlsx.go — all stream to io.Writer
  api/         http handlers, DTOs, error→status mapping
  cache/       redis wrapper
  config/      env-based config
```

**Dependency rule:** `domain` imports nothing. `provider`/`store`/`export`/`cache`
import only `domain`. `service` orchestrates them. `api` depends on `service`.
`cmd/server` wires concretes together.

## Commands

```
make dev          # run API server (reads .env)
make test         # go test ./...
make build        # compile to bin/server
make lint         # go vet + gofmt -l
make migrate-up   # apply migrations
make sqlc         # regenerate DB code
make frontend-dev # next.js dev server
```

Copy `.env.example` → `.env` first.

## Conventions

- **TDD**: red → green → refactor. Table-driven tests. Commit per task.
- **sqlc** for all DB access — no hand-written SQL string concatenation in Go.
- **Error wrapping**: `fmt.Errorf("doing X: %w", err)`. Define typed sentinel
  errors (`provider.ErrUpstreamBusy`, `export.ErrBadFormat`) for control flow.
- **No silent fallbacks**: a partial failure fails the request with a clear
  message; never return partial data labeled as complete. Empty result
  (`count: 0`) is distinct from an error.
- **Streaming exports**: encoders write row-by-row to `io.Writer`. Never build
  the whole file in memory.
- **Focused files**: one responsibility per file. If a file grows unwieldy, split it.

## Filters → OSM tags

- no website = `website` AND `contact:website` absent
- no socials = `contact:instagram|facebook|vk|telegram` all absent
- no phone = `phone` AND `contact:phone` absent
- category = whitelist of `shop|amenity|office|craft|tourism` keys

## Subagents & skills

Project agents in `.claude/agents/`: `overpass-expert`, `go-backend-architect`,
`nextjs-map-frontend`, `export-specialist`.
Project skills in `.claude/skills/`: `overpass-query`, `add-data-provider`,
`db-migration`.

## Specs & plans

- Design: `docs/superpowers/specs/2026-07-01-company-parser-design.md`
- Plan: `docs/superpowers/plans/2026-07-01-company-parser.md`
