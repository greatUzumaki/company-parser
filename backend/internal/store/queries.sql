-- name: UpsertCompany :batchone
-- Insert a company or update it on OSM identity conflict; returns its id.
INSERT INTO companies (
    osm_type, osm_id, name, category, subcategory,
    website, phone, email, instagram, facebook, vk, telegram, whatsapp,
    addr_country, addr_city, addr_street, addr_housenumber, addr_postcode,
    opening_hours, tags, lat, lon
) VALUES (
    @osm_type, @osm_id, @name, @category, @subcategory,
    @website, @phone, @email, @instagram, @facebook, @vk, @telegram, @whatsapp,
    @addr_country, @addr_city, @addr_street, @addr_housenumber, @addr_postcode,
    @opening_hours, @tags, @lat, @lon
)
ON CONFLICT (osm_type, osm_id) DO UPDATE SET
    name = EXCLUDED.name,
    category = EXCLUDED.category,
    subcategory = EXCLUDED.subcategory,
    website = EXCLUDED.website,
    phone = EXCLUDED.phone,
    email = EXCLUDED.email,
    instagram = EXCLUDED.instagram,
    facebook = EXCLUDED.facebook,
    vk = EXCLUDED.vk,
    telegram = EXCLUDED.telegram,
    whatsapp = EXCLUDED.whatsapp,
    addr_country = EXCLUDED.addr_country,
    addr_city = EXCLUDED.addr_city,
    addr_street = EXCLUDED.addr_street,
    addr_housenumber = EXCLUDED.addr_housenumber,
    addr_postcode = EXCLUDED.addr_postcode,
    opening_hours = EXCLUDED.opening_hours,
    tags = EXCLUDED.tags,
    lat = EXCLUDED.lat,
    lon = EXCLUDED.lon,
    updated_at = now()
RETURNING id;

-- name: CreateSearch :one
INSERT INTO searches (region_name, region_area_id, filters, result_count)
VALUES (@region_name, @region_area_id, @filters, @result_count)
RETURNING id;

-- name: LinkSearchResult :batchexec
INSERT INTO search_results (search_id, company_id)
VALUES (@search_id, @company_id)
ON CONFLICT DO NOTHING;

-- name: ListSearches :many
SELECT id, region_name, region_area_id, filters, result_count, created_at
FROM searches
ORDER BY created_at DESC
LIMIT @lim OFFSET @off;

-- name: GetSearch :one
SELECT id, region_name, region_area_id, filters, result_count, created_at
FROM searches
WHERE id = @id;

-- name: GetSearchCompanies :many
SELECT
    c.osm_type, c.osm_id, c.name, c.category, c.subcategory,
    c.website, c.phone, c.email, c.instagram, c.facebook, c.vk, c.telegram, c.whatsapp,
    c.addr_country, c.addr_city, c.addr_street, c.addr_housenumber, c.addr_postcode,
    c.opening_hours, c.lat, c.lon
FROM search_results sr
JOIN companies c ON c.id = sr.company_id
WHERE sr.search_id = @search_id
ORDER BY c.name;
