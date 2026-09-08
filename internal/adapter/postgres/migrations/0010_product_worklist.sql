-- 0010: the worklist both screens live on.
--
-- The staff need "what is still unfinished, and which of the four steps are
-- left"; the customer needs "where is the thing I asked for". Neither question
-- is answerable from an aggregate — catalog knows a product's state only as the
-- reason Publish said no — and joining five contexts to answer it is exactly
-- what the outbox exists to avoid.
--
-- Written only by events, like every projection here. Nothing in it is an
-- invariant: a row can only be stale, and the cure for stale is to wait.

CREATE TABLE product_worklist (
    product      uuid PRIMARY KEY,
    merchant     uuid        NOT NULL,
    category     text        NOT NULL,
    name         text        NOT NULL,
    source_url   text        NOT NULL,        -- the page a buyer has to open
    price_minor  bigint      NOT NULL,
    price_currency text      NOT NULL,
    sourced_by   text        NOT NULL,        -- customer | operator | feed
    requested_by uuid,                        -- NULL: nobody is waiting

    -- The steps, each set by its own event. Descriptive, never an invariant:
    -- this table protects nothing, it only remembers. "has a variant" is not a
    -- column, because it is just "variants is not empty" and two places to
    -- store one fact is two places to disagree.
    listing_confirmed boolean NOT NULL DEFAULT false,
    measured          boolean NOT NULL DEFAULT false,
    published         boolean NOT NULL DEFAULT false,

    -- Every size announced so far, as [{"id":…,"label":"US 9 · black"}]. The
    -- customer has to PICK one to order, so the id has to be here; the label
    -- is the shop's own words, carried along so no screen has to join.
    --
    -- jsonb rather than a child table because this is a projection: nothing
    -- reads one variant on its own, and a row is always rewritten whole.
    variants jsonb NOT NULL DEFAULT '[]',

    added_at   timestamptz NOT NULL,
    updated_at timestamptz NOT NULL
);

-- The staff screen asks for unfinished work, oldest first.
CREATE INDEX product_worklist_open_idx ON product_worklist (published, added_at);
-- The customer screen asks for their own rows.
CREATE INDEX product_worklist_requested_by_idx ON product_worklist (requested_by);
