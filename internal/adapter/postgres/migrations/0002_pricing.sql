-- 0002: the pricing context.
--
-- Same rules as 0001: columns mirror the domain's snapshots and value objects,
-- money is minor units + ISO code, weights are grams, sizes are millimetres.
-- Nothing here references a catalog table — pricing owns COPIES of what it
-- needs (listings, category_profiles), fed by catalog's events. That is the
-- boundary between two bounded contexts drawn in SQL (DDD.md §7).

-- A shipping lane = the carrier's price list, one row per lane, replaced whole.
CREATE TABLE lanes (
    code                     text    PRIMARY KEY,           -- natural key: us_forwarder
    name                     text    NOT NULL,
    divisor                  bigint  NOT NULL,              -- cm³ per kg: 5000, 6000
    step_g                   bigint  NOT NULL,              -- billing step in grams
    currency                 text    NOT NULL,              -- of every amount on this row
    rate_standard_minor      bigint  NOT NULL,              -- per kg, by goods class
    rate_branded_minor       bigint  NOT NULL,
    rate_electronics_minor   bigint  NOT NULL,
    rate_sensitive_minor     bigint  NOT NULL,
    surcharge_battery_minor  bigint,                        -- NULL = this lane has none
    duty_itemised            boolean NOT NULL DEFAULT false,
    duty_rate_ppm            bigint  NOT NULL DEFAULT 0     -- only meaningful when itemised
);

-- Today's rate per currency pair. A quote copies the rate it used (quotes.fx_rate).
CREATE TABLE fx_rates (
    from_code   text        NOT NULL,
    to_code     text        NOT NULL,
    rate        text        NOT NULL,                       -- canonical decimal text: "26000.5"
    updated_at  timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (from_code, to_code)
);

-- Pricing's PROJECTION of a product: written only by the projector that
-- consumes catalog.product_* events. product is a bare uuid on purpose —
-- no foreign key into catalog's tables.
CREATE TABLE listings (
    product           uuid    PRIMARY KEY,
    name              text    NOT NULL,
    category          text    NOT NULL,
    price_minor       bigint  NOT NULL,
    price_currency    text    NOT NULL,
    parcel_weight_g   bigint,                               -- all NULL until measured
    parcel_length_mm  bigint,
    parcel_width_mm   bigint,
    parcel_height_mm  bigint,
    measured          boolean NOT NULL DEFAULT false,
    active            boolean NOT NULL DEFAULT true
);

-- Pricing's projection of a category, plus its own goods class.
CREATE TABLE category_profiles (
    code           text    PRIMARY KEY,
    class          text    NOT NULL,                        -- standard | branded | electronics | sensitive
    est_weight_g   bigint  NOT NULL,
    est_length_mm  bigint  NOT NULL,
    est_width_mm   bigint  NOT NULL,
    est_height_mm  bigint  NOT NULL,
    restrictions   text[]  NOT NULL DEFAULT '{}'
);

-- A quote is a photograph: every input it was computed from, frozen. The
-- breakdown is stored line by line so reconciliation can read it in SQL.
CREATE TABLE quotes (
    id                    uuid        PRIMARY KEY,
    product               uuid        NOT NULL,
    lane                  text        NOT NULL,
    status                text        NOT NULL,             -- issued | accepted | expired
    issued_at             timestamptz NOT NULL,
    expires_at            timestamptz NOT NULL,
    -- breakdown, in the lane's currency
    class                 text        NOT NULL,
    chargeable_g          bigint      NOT NULL,
    estimated             boolean     NOT NULL,
    lane_currency         text        NOT NULL,
    item_minor            bigint      NOT NULL,
    tax_minor             bigint      NOT NULL,
    freight_minor         bigint      NOT NULL,
    surcharge_minor       bigint      NOT NULL,
    duty_minor            bigint      NOT NULL,
    subtotal_minor        bigint      NOT NULL,
    -- breakdown, in the home currency
    fx_rate               text        NOT NULL,             -- lane_currency → home_currency, as quoted
    home_currency         text        NOT NULL,
    subtotal_home_minor   bigint      NOT NULL,
    fee_minor             bigint      NOT NULL,
    total_minor           bigint      NOT NULL,
    deposit_minor         bigint      NOT NULL
);

CREATE INDEX quotes_product_idx ON quotes (product);
CREATE INDEX quotes_open_idx ON quotes (expires_at) WHERE status = 'issued';   -- the expiry sweep
