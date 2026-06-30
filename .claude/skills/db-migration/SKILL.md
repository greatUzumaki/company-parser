---
name: db-migration
description: >
  Use when changing the database schema — creating, applying, or rolling back
  golang-migrate migrations, or working with the PostGIS geometry column and
  sqlc regeneration. Triggers on "add a migration", "alter the schema",
  "migrate up/down", "new table/column", "PostGIS", "regenerate sqlc".
---

# Database migrations (golang-migrate + sqlc + PostGIS)

Migrations live in `backend/migrations/`, paired `*.up.sql` / `*.down.sql`,
zero-padded sequential prefixes (`0001_init`, `0002_...`).

## Create a migration

```bash
migrate create -ext sql -dir backend/migrations -seq <name>
# -> 000X_<name>.up.sql and .down.sql
```

Write the forward change in `.up.sql` and the exact inverse in `.down.sql`.
**Always** provide a real down migration — a reviewer can reject otherwise.

## Apply / roll back

```bash
make migrate-up           # apply all pending
make migrate-down         # roll back the most recent one
```

`DATABASE_URL` comes from `.env`.

## PostGIS notes

- The extension must exist before `0001`:
  `CREATE EXTENSION IF NOT EXISTS postgis;` (first line of the init up-migration).
  If the user's Postgres image lacks PostGIS, switch it to `postgis/postgis`.
- Geometry column: `geom geometry(Point, 4326)`. Index it:
  `CREATE INDEX ... USING GIST (geom);`.
- Insert points with `ST_SetSRID(ST_MakePoint(lon, lat), 4326)`.
- Region filtering: `ST_Within(geom, ST_GeomFromGeoJSON($1))` or bbox via
  `ST_MakeEnvelope(minLon, minLat, maxLon, maxLat, 4326)`.

## After a schema change

1. Update `backend/internal/store/queries.sql`.
2. `make sqlc` to regenerate type-safe code.
3. Update the `store.go` wrapper + its integration test.
4. Run `make test`.

## Identity / dedup

`companies` has `UNIQUE (osm_type, osm_id)`. Upserts use
`ON CONFLICT (osm_type, osm_id) DO UPDATE` so re-running a search dedups across runs.
