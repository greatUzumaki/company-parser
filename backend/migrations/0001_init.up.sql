CREATE TABLE companies (
    id           BIGSERIAL PRIMARY KEY,
    osm_type     TEXT        NOT NULL,
    osm_id       TEXT        NOT NULL,
    name         TEXT        NOT NULL DEFAULT '',
    category     TEXT        NOT NULL DEFAULT '',
    subcategory  TEXT        NOT NULL DEFAULT '',
    website      TEXT        NOT NULL DEFAULT '',
    phone        TEXT        NOT NULL DEFAULT '',
    email        TEXT        NOT NULL DEFAULT '',
    instagram    TEXT        NOT NULL DEFAULT '',
    facebook     TEXT        NOT NULL DEFAULT '',
    vk           TEXT        NOT NULL DEFAULT '',
    telegram     TEXT        NOT NULL DEFAULT '',
    addr_country TEXT        NOT NULL DEFAULT '',
    addr_city    TEXT        NOT NULL DEFAULT '',
    addr_street  TEXT        NOT NULL DEFAULT '',
    addr_housenumber TEXT    NOT NULL DEFAULT '',
    addr_postcode TEXT       NOT NULL DEFAULT '',
    opening_hours TEXT       NOT NULL DEFAULT '',
    tags         JSONB       NOT NULL DEFAULT '{}'::jsonb,
    lat          DOUBLE PRECISION NOT NULL DEFAULT 0,
    lon          DOUBLE PRECISION NOT NULL DEFAULT 0,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (osm_type, osm_id)
);

CREATE INDEX companies_tags_idx ON companies USING GIN (tags);
CREATE INDEX companies_category_idx ON companies (category);
CREATE INDEX companies_latlon_idx ON companies (lat, lon);

CREATE TABLE searches (
    id             BIGSERIAL PRIMARY KEY,
    region_name    TEXT        NOT NULL,
    region_area_id BIGINT      NOT NULL,
    filters        JSONB       NOT NULL DEFAULT '{}'::jsonb,
    result_count   INT         NOT NULL DEFAULT 0,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE search_results (
    search_id  BIGINT NOT NULL REFERENCES searches (id) ON DELETE CASCADE,
    company_id BIGINT NOT NULL REFERENCES companies (id) ON DELETE CASCADE,
    PRIMARY KEY (search_id, company_id)
);
