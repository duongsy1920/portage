-- 0003: the ordering context.
--
-- Same rules as before: columns mirror OrderSnapshot, money is minor units +
-- ISO code, ids are uuid v7. No REFERENCES into pricing or catalog — quote,
-- product and variant are ids ordering was TOLD, not rows it may join.

-- Ordering's projection of pricing.quote_accepted: the numbers an order is
-- placed on. Written only by the projector.
CREATE TABLE accepted_quotes (
    quote           uuid    PRIMARY KEY,
    product         uuid    NOT NULL,
    total_minor     bigint  NOT NULL,
    deposit_minor   bigint  NOT NULL,
    currency        text    NOT NULL
);

CREATE TABLE orders (
    id              uuid        PRIMARY KEY,
    quote           uuid        NOT NULL UNIQUE,        -- one quote, one order (OrderRepository.ByQuote)
    product         uuid        NOT NULL,
    variant         uuid        NOT NULL,
    customer        uuid        NOT NULL,
    total_minor     bigint      NOT NULL,
    deposit_minor   bigint      NOT NULL,
    currency        text        NOT NULL,
    status          text        NOT NULL,               -- awaiting_deposit | deposited | purchased | purchase_failed | in_transit | delivered | cancelled
    balance_paid    boolean     NOT NULL DEFAULT false,
    refund_minor    bigint,                             -- only when cancelled
    forfeited       boolean     NOT NULL DEFAULT false,
    cancel_reason   text        NOT NULL DEFAULT '',
    placed_at       timestamptz NOT NULL
);

CREATE INDEX orders_customer_idx ON orders (customer);
CREATE INDEX orders_status_idx ON orders (status);
