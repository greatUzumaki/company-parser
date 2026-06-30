---
name: overpass-query
description: >
  Use when building or modifying an Overpass QL query for this project — turning
  a selected region + filters into the query string, choosing tags, or debugging
  timeouts/empty results. Triggers on "overpass query", "OSM query", "fetch
  companies for region", "area scoping", "out center".
---

# Building an Overpass query (this project)

## Recipe: region + filter → QL

1. **Scope to the region by area.** Overpass area id = `3600000000 + osm_relation_id`.
   The frontend sends `osmAreaId` already in this form when it can; if you only
   have a relation id, add the offset.

   ```
   area(3600002145)->.a;
   ```

2. **One clause per whitelisted category key**, all `node`/`way`, scoped to `.a`:

   ```overpassql
   [out:json][timeout:180];
   area(3600002145)->.a;
   (
     node["shop"](area.a);
     way["shop"](area.a);
     node["amenity"](area.a);
     way["amenity"](area.a);
   );
   out center tags;
   ```

   Use the categories the user selected. `out center` gives ways a coordinate.

3. **Do NOT encode absence filters in QL.** "No website / no socials / no phone"
   are applied in Go (`domain.Filter.Match`) after fetch — pure-QL absence
   (`[!website]`) is slow and composes badly. QL only narrows by area + category.

## Tag → field cheat sheet

| Field | Tags (first present wins) |
|---|---|
| website | `website`, `contact:website` |
| phone | `phone`, `contact:phone` |
| email | `email`, `contact:email` |
| instagram/facebook/vk/telegram | `contact:instagram` / `contact:facebook` / `contact:vk` / `contact:telegram` |
| category / subcategory | first whitelist key present / its value |
| address | `addr:country/city/street/housenumber/postcode` |

## Debugging

- **Timeout / 504 / 429** → region too big or endpoint busy. Split into bbox
  tiles, lower per-query scope, back off and retry. Map to `ErrUpstreamBusy`.
- **Empty result** → check the area id offset and that at least one category
  clause is present. Verify with a tiny bbox first:
  `node["shop"](47.0,8.0,47.1,8.1);out;`.
- Always send a `User-Agent` header and cache the response in Redis.

## Where the code lives

- `backend/internal/provider/overpass/query.go` — `BuildQuery(region, filter)`
- `mapping.go` — `mapElement` (tags → `domain.Company`)
- `client.go` — HTTP, decode, filter
