# Company Parser

Find companies by **region + filters** from OpenStreetMap, pick the region on a
map, and export to **JSON / CSV / Excel**. Built for lead generation — e.g. find
businesses with **no website**, **no socials**, or **no phone**.

> **Data:** © OpenStreetMap contributors, licensed under
> [ODbL](https://www.openstreetmap.org/copyright). This project uses the Overpass
> API and Nominatim — **no Google/Yandex/2GIS scraping**.

## Stack

- **Backend:** Go 1.23+ (hexagonal, `Provider` interface, pgx/sqlc, PostGIS, Redis cache)
- **Frontend:** Next.js 15 + TypeScript + Tailwind/shadcn + MapLibre GL
- **Data:** OpenStreetMap via Overpass API; Nominatim for region lookup

## Quick start

Postgres (+PostGIS) and Redis are expected to already run in your Docker.

```bash
cp .env.example .env          # adjust DATABASE_URL / REDIS_URL
make migrate-up               # apply schema
make dev                      # backend on :8080
make frontend-dev             # frontend on :3000
```

> PostGIS must be available. Check:
> `psql "$DATABASE_URL" -c "CREATE EXTENSION IF NOT EXISTS postgis;"`
> If that fails, use a `postgis/postgis` image for Postgres.

## Project docs

- Design spec: `docs/superpowers/specs/2026-07-01-company-parser-design.md`
- Implementation plan: `docs/superpowers/plans/2026-07-01-company-parser.md`
- Conventions & architecture: `CLAUDE.md`

## Make targets

`make dev | build | test | lint | migrate-up | migrate-down | sqlc | frontend-dev`
