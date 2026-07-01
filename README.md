# Company Parser

Find companies by **region + filters** from open data, plot them on a map, export
to **JSON / CSV / Excel**, and run **email campaigns** to the ones with an address.
Built for lead generation — e.g. find businesses with **no website / no socials /
no phone**, then reach out.

> **Data:** © OpenStreetMap contributors ([ODbL](https://www.openstreetmap.org/copyright))
> + Wikidata (CC0). Uses the Overpass API, Nominatim, and the Wikidata Query
> Service — **no Google/Yandex/2GIS scraping**.

## Features

- **Map-first UI** (MapLibre GL) — full-screen map, floating collapsible panels.
- **Pick an area** three ways: click a country boundary, or **draw a zone**
  (circle / rectangle / polygon / freehand) and search inside it.
- **Multiple free sources** — OpenStreetMap (Overpass) + Wikidata, fanned out
  concurrently, merged and deduped. Add more behind the `Provider` interface.
- **Live streaming search** — results plot on the map and fill the table as they
  arrive, with per-source progress. Companies you hover in the list highlight on
  the map.
- **Stale-while-revalidate cache** — repeat searches are instant; a background
  refresh finds new/updated companies (manual refresh button forces a re-parse).
- **Contact extraction** — website, phone, email, Instagram/Facebook/VK/Telegram,
  WhatsApp. Filter/sort results (e.g. "has email"), click a marker for details.
- **Export** — streaming JSON / CSV / Excel.
- **Email campaigns** — send to the filtered recipients with an email, via your
  own SMTP, with `{{name}}` personalization, dry-run, consent gate, and rate limit.
- **i18n** — English / Russian.

## Stack

- **Backend:** Go 1.23+ — hexagonal, `Provider` interface, `pgx`/`sqlc`,
  `golang-migrate`, streaming NDJSON, Redis cache, `go-mail` for SMTP.
- **Frontend:** Next.js 16 (App Router) + TypeScript + Tailwind + MapLibre GL +
  Terra Draw + TanStack Query/Table + zustand + Framer Motion.
- **Storage:** Postgres (no PostGIS needed — plain lat/lon) + Redis. Both expected
  to already run in your Docker.

## Quick start

```bash
cp .env.example .env          # set DATABASE_URL / REDIS_URL to your containers
# create the app DB:
docker exec postgres psql -U postgres -c "CREATE DATABASE parse_companies;"
make migrate-up               # apply schema
make dev                      # backend on :8080
make frontend-dev             # frontend on :3000  ->  open http://localhost:3000
```

### Email campaigns (optional)

Disabled until SMTP is configured. Add your **own** SMTP creds to `.env`
(Gmail app-password, Brevo, SendGrid, SES…), then restart the backend:

```bash
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=you@gmail.com
SMTP_PASSWORD=your_app_password
SMTP_FROM=Your Name <you@gmail.com>
```

`GET /api/v1/campaign/status` → `{"enabled":true}` when ready. ⚠️ Only email
recipients who consented — unsolicited bulk email is regulated (CAN-SPAM / GDPR /
ФЗ «О рекламе»). `.env` is gitignored — keep credentials private.

## API

| Method | Path | Purpose |
|---|---|---|
| POST | `/api/v1/search` | one-shot search (cached) |
| POST | `/api/v1/search/stream` | streaming NDJSON search (`force` to re-parse) |
| GET | `/api/v1/searches` · `/{id}` | history · one search |
| GET | `/api/v1/searches/{id}/export?format=json\|csv\|xlsx` | export |
| GET | `/api/v1/regions?q=` · `/categories` | region autocomplete · categories |
| GET | `/api/v1/campaign/status` | is SMTP configured |
| POST | `/api/v1/campaign/send` | streaming campaign (recipients + consent) |

## Project docs

- Design spec: `docs/superpowers/specs/2026-07-01-company-parser-design.md`
- Implementation plan: `docs/superpowers/plans/2026-07-01-company-parser.md`
- Conventions & architecture: `CLAUDE.md`

## Make targets

`make dev | build | test | lint | migrate-up | migrate-down | sqlc | frontend-dev`
