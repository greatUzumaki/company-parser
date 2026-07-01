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

## Coordinates (plain Postgres — no PostGIS)

The deployment Postgres has no PostGIS, so coordinates are plain columns:
`lat DOUBLE PRECISION, lon DOUBLE PRECISION`, with a btree index
`(lat, lon)`. Region scoping happens server-side in the Overpass query (by
admin area); the DB stores coordinates for display/export, not spatial joins.

If you ever need true spatial queries, do NOT alter the shared Postgres
container — stand up a dedicated `postgis/postgis` instance and add `geom
geometry(Point,4326)` there. Until then, bbox filtering is a plain
`lat BETWEEN ... AND lon BETWEEN ...`.

## After a schema change

1. Update `backend/internal/store/queries.sql`.
2. `make sqlc` to regenerate type-safe code.
3. Update the `store.go` wrapper + its integration test.
4. Run `make test`.

## Identity / dedup

`companies` has `UNIQUE (osm_type, osm_id)`. Upserts use
`ON CONFLICT (osm_type, osm_id) DO UPDATE` so re-running a search dedups across runs.
