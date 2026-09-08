-- The READ side (DDD.md §24). One flat row per order, written only by the
-- projector in internal/app/reporting from the events of five contexts.
--
-- Two things this table deliberately does NOT have:
--
--   * FOREIGN KEYS. Not to orders, not to products, not to parcels. A read
--     model that referenced the write side would make the two impossible to
--     deploy, migrate or shard apart — and would fail on the retry that
--     delivers an event before the row it "references" exists.
--   * NOT NULL on anything but the id. Any event may create the row (the
--     FRAME), so a half-filled row is a normal, expected state, not corruption.
CREATE TABLE order_summaries (
    "order"        uuid PRIMARY KEY,
    customer       uuid,
    product        uuid,
    variant        uuid,
    quote          uuid,

    product_name   text        NOT NULL DEFAULT '',
    status         text        NOT NULL DEFAULT '',
    tracking       text        NOT NULL DEFAULT 'none',

    total_minor    bigint,
    deposit_minor  bigint,
    refund_minor   bigint,
    currency       text,

    deposit_paid   boolean     NOT NULL DEFAULT false,
    balance_paid   boolean     NOT NULL DEFAULT false,
    forfeited      boolean     NOT NULL DEFAULT false,

    shop_reference text        NOT NULL DEFAULT '',

    placed_at      timestamptz,
    delivered_at   timestamptz,
    cancelled_at   timestamptz,
    updated_at     timestamptz NOT NULL DEFAULT now()
);

-- The two screens this table exists for, one index each.
CREATE INDEX order_summaries_customer_idx ON order_summaries (customer, placed_at DESC);
CREATE INDEX order_summaries_status_idx   ON order_summaries (status, placed_at DESC);
-- Only for the late-name backfill (a product published after an order on it).
CREATE INDEX order_summaries_product_idx  ON order_summaries (product);

-- product id → name, so a summary shows "Air Trainer 90" and not a UUID.
CREATE TABLE product_names (
    product uuid PRIMARY KEY,
    name    text NOT NULL
);
