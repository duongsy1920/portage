-- 0005: the logistics context, and pricing's Quote-vs-Actual table.
--
-- No REFERENCES into other contexts. Batch items and allocations are child
-- rows of the batch aggregate: rewritten with their parent (like product
-- variants), never addressed on their own.

CREATE TABLE lane_rules (                       -- logistics' projection of pricing.lane_defined
    code      text   PRIMARY KEY,
    divisor   bigint NOT NULL,
    step_g    bigint NOT NULL
);

CREATE TABLE parcels (
    id                uuid        PRIMARY KEY,
    "order"           uuid        NOT NULL UNIQUE,   -- one order, one parcel (v1)
    reference         text        NOT NULL,          -- the shop's order number printed on the box
    status            text        NOT NULL,          -- expected | received | batched | shipped
    actual_weight_g   bigint,                        -- all NULL until received
    actual_length_mm  bigint,
    actual_width_mm   bigint,
    actual_height_mm  bigint,
    received_by       uuid,
    received_at       timestamptz,
    batch             uuid,                          -- once batched
    expected_at       timestamptz NOT NULL
);

CREATE INDEX parcels_pending_idx ON parcels (expected_at) WHERE status <> 'shipped';   -- the warehouse screen

CREATE TABLE batches (
    id                uuid        PRIMARY KEY,
    lane              text        NOT NULL,
    status            text        NOT NULL,          -- open | closed | shipped
    freight_minor     bigint,                        -- the carrier's invoice, once shipped
    freight_currency  text,
    opened_at         timestamptz NOT NULL,
    closed_at         timestamptz,
    shipped_at        timestamptz
);

CREATE TABLE batch_items (
    batch      uuid   NOT NULL REFERENCES batches(id) ON DELETE CASCADE,
    position   int    NOT NULL,
    parcel     uuid   NOT NULL,
    "order"    uuid   NOT NULL,
    weight_g   bigint NOT NULL,
    length_mm  bigint NOT NULL,
    width_mm   bigint NOT NULL,
    height_mm  bigint NOT NULL,
    PRIMARY KEY (batch, position)
);

CREATE TABLE batch_allocations (                -- frozen when the batch ships: each order's share
    batch          uuid   NOT NULL REFERENCES batches(id) ON DELETE CASCADE,
    position       int    NOT NULL,
    parcel         uuid   NOT NULL,
    "order"        uuid   NOT NULL,
    chargeable_g   bigint NOT NULL,
    freight_minor  bigint NOT NULL,
    PRIMARY KEY (batch, position)
);

-- pricing: Quote vs Actual per order (DDD.md §28). Quoted side from the quote,
-- actual side from procurement's receipt and logistics' invoice. Lane currency.
CREATE TABLE reconciliations (
    "order"                 uuid   PRIMARY KEY,
    quote                   uuid,
    currency                text,                    -- of every amount on the row, once any is known
    quoted_goods_minor      bigint,
    quoted_freight_minor    bigint,
    quoted_chargeable_g     bigint,
    actual_goods_minor      bigint,
    actual_freight_minor    bigint,
    actual_chargeable_g     bigint
);
