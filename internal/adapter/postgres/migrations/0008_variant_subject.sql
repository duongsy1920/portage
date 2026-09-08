-- 0008: which size to buy.
--
-- An order names the variant only by id, and procurement may not read
-- catalog's tables to translate it (guard 7). So catalog announces
-- catalog.variant_added and procurement keeps a dictionary of it, then copies
-- the words onto the task when it opens — the same way it already copies the
-- shop's currency. Before this, the buyer's screen showed a uuid.

CREATE TABLE procurement_variants (              -- projection of catalog.variant_added
    variant       uuid PRIMARY KEY,
    product       uuid NOT NULL,
    -- Free text, in the shop's own words: "M 8 / W 9.5", "42 EU", "XS".
    -- No taxonomy here on purpose (see contracts.VariantAddedV1).
    size          text NOT NULL DEFAULT '',
    color         text NOT NULL DEFAULT '',
    merchant_ref  text NOT NULL DEFAULT ''       -- the shop's own code, when it has one
);

CREATE INDEX procurement_variants_product_idx ON procurement_variants (product);

-- The page to buy from, announced by catalog.product_published.
ALTER TABLE procurement_items ADD COLUMN source text NOT NULL DEFAULT '';

-- What to buy, in words, frozen when the task opened (procurement.Subject).
-- Descriptive, never an invariant: a one-size product has no label at all.
ALTER TABLE purchase_tasks ADD COLUMN product_name  text NOT NULL DEFAULT '';
ALTER TABLE purchase_tasks ADD COLUMN variant_label text NOT NULL DEFAULT '';
ALTER TABLE purchase_tasks ADD COLUMN variant_ref   text NOT NULL DEFAULT '';
ALTER TABLE purchase_tasks ADD COLUMN source        text NOT NULL DEFAULT '';

-- Ordering keeps the thinnest possible copy: existence and owner. It has no
-- rule that needs the words, and a projection carrying unread data goes stale
-- without anybody noticing (ordering.Variant).
CREATE TABLE ordering_variants (
    variant uuid PRIMARY KEY,
    product uuid NOT NULL
);
