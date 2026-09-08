-- 0001: the catalog context and the outbox.
--
-- Columns mirror the snapshots one to one (internal/domain/catalog/snapshot.go):
-- money is minor units + ISO code, weights are grams, sizes are millimetres,
-- ids are UUID v7, times are timestamptz. Nothing here is derived or denormalised.

CREATE TABLE merchants (
    id                         uuid        PRIMARY KEY,
    name                       text        NOT NULL,
    site                       text        NOT NULL UNIQUE,
    currency                   text        NOT NULL,
    free_ship_kind             text        NOT NULL,          -- never | over | always
    free_ship_threshold_minor  bigint,                        -- only when kind = over
    sourcing                   text[]      NOT NULL DEFAULT '{}',
    status                     text        NOT NULL,          -- active | suspended
    added_at                   timestamptz NOT NULL
);

CREATE TABLE categories (
    code            text    PRIMARY KEY,                      -- natural key: footwear, apparel, ...
    est_weight_g    bigint  NOT NULL,
    est_length_mm   bigint  NOT NULL,
    est_width_mm    bigint  NOT NULL,
    est_height_mm   bigint  NOT NULL,
    restrictions    text[]  NOT NULL DEFAULT '{}'
);

CREATE TABLE products (
    id                      uuid        PRIMARY KEY,
    merchant_id             uuid        NOT NULL REFERENCES merchants(id),
    category                text        NOT NULL REFERENCES categories(code),
    name                    text        NOT NULL,
    source_url              text        NOT NULL,
    source_host             text        NOT NULL,
    -- three provenances, one per attribute group (CATALOG.md §4)
    listing_source          text        NOT NULL,
    listing_at              timestamptz NOT NULL,
    listing_by              uuid,
    price_minor             bigint      NOT NULL,
    price_currency          text        NOT NULL,
    price_source            text        NOT NULL,
    price_at                timestamptz NOT NULL,
    price_by                uuid,
    -- parcel: all NULL until measured, all set after — never half
    parcel_weight_g         bigint,
    parcel_length_mm        bigint,
    parcel_width_mm         bigint,
    parcel_height_mm        bigint,
    parcel_source           text,
    parcel_at               timestamptz,
    parcel_by               uuid,
    suspected_duplicate_of  uuid,
    dismissed_duplicates    text[]      NOT NULL DEFAULT '{}',
    status                  text        NOT NULL,             -- draft | published | retired
    added_at                timestamptz NOT NULL
);

CREATE INDEX products_source_url_idx ON products (source_url);   -- ProductRepository.BySource

CREATE TABLE product_variants (
    id            uuid        PRIMARY KEY,
    product_id    uuid        NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    position      int         NOT NULL,                       -- insertion order, kept on reload
    size          text        NOT NULL DEFAULT '',
    color         text        NOT NULL DEFAULT '',
    merchant_ref  text        NOT NULL DEFAULT '',
    added_at      timestamptz NOT NULL
);

-- The outbox: one row per domain event, written in the SAME transaction as the
-- aggregate (DDD.md §27). A worker reads sent_at IS NULL in id order, publishes,
-- sets sent_at. At-least-once: a crash between publish and update re-sends.
CREATE TABLE outbox (
    id          bigserial   PRIMARY KEY,
    event_name  text        NOT NULL,
    occurred_at timestamptz NOT NULL,
    payload     jsonb       NOT NULL,
    created_at  timestamptz NOT NULL DEFAULT now(),
    sent_at     timestamptz
);

CREATE INDEX outbox_pending_idx ON outbox (id) WHERE sent_at IS NULL;
