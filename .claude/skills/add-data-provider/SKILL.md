---
name: add-data-provider
description: >
  Use when adding a new company data source (e.g. an official Places API) behind
  the Provider interface without touching service/api/frontend. Triggers on "add
  a provider", "new data source", "integrate Google Places / 2GIS / Yandex",
  "swap the data source".
---

# Adding a data Provider

The whole backend depends on one interface. A new source is one new package that
implements it — nothing upstream changes.

```go
// backend/internal/provider/provider.go
type Provider interface {
    Search(ctx context.Context, r domain.Region, f domain.Filter) ([]domain.Company, error)
}
```

## Steps

1. **Create the package:** `backend/internal/provider/<source>/client.go`.
2. **Legal check first.** Only official APIs with a ToS that permits storage +
   export. **No HTML/map scraping** (Google/Yandex/2GIS map UIs are off-limits —
   see CLAUDE.md). If unsure, stop and ask.
3. **Implement `Search`:** call the source, map its response to `domain.Company`
   (mirror the tag-mapping approach in `overpass/mapping.go`), then apply
   `f.Match` so absence filters behave identically across sources.
4. **Errors:** map upstream 429/5xx to `provider.ErrUpstreamBusy`; wrap others
   with `%w`. No silent empty results.
5. **Config:** add the endpoint/api-key env vars in `internal/config/config.go`
   and `.env.example`.
6. **Tests:** `client_test.go` with `httptest.Server` and a recorded fixture —
   assert mapping + filtering + the busy-error path.
7. **Wire it:** in `cmd/server/main.go`, construct your provider and pass it to
   `service.New(...)`. To support multiple sources at once, add a small
   composite provider that fans out and merges by OSM/identity dedup — but keep
   that logic in `service`, not in `api`.

## Do NOT

- Don't add source-specific fields to `domain.Company` unless every provider can
  fill them. Keep `domain` source-agnostic; stash extras in `Tags`.
- Don't let `api` or the frontend learn which provider is active.
