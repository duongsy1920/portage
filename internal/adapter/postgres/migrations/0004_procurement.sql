-- 0004: the procurement context.
--
-- purchase_tasks mirrors TaskSnapshot; procurement_shops and procurement_items
-- are procurement's projections of catalog (which shop a product belongs to,
-- what the shop bills in). No REFERENCES into other contexts' tables.

CREATE TABLE procurement_shops (
    merchant   uuid  PRIMARY KEY,
    name       text  NOT NULL,
    site       text  NOT NULL,
    currency   text  NOT NULL
);

CREATE TABLE procurement_items (
    product    uuid  PRIMARY KEY,
    merchant   uuid  NOT NULL,
    name       text  NOT NULL
);

CREATE TABLE purchase_tasks (
    id          uuid        PRIMARY KEY,
    "order"     uuid        NOT NULL UNIQUE,        -- one deposited order, one task (TaskRepository.ByOrder)
    product     uuid        NOT NULL,
    variant     uuid        NOT NULL,
    currency    text        NOT NULL,               -- the shop's; what paid_minor is in
    status      text        NOT NULL,               -- open | confirmed | failed
    reference   text        NOT NULL DEFAULT '',    -- the shop's order number, once confirmed
    paid_minor  bigint,                             -- what we actually paid, once confirmed (Quote vs Actual)
    paid_by     uuid,
    reason      text        NOT NULL DEFAULT '',    -- once failed
    opened_at   timestamptz NOT NULL,
    closed_at   timestamptz
);

CREATE INDEX purchase_tasks_open_idx ON purchase_tasks (opened_at) WHERE status = 'open';   -- the buyer's to-do list
