-- 0009: WHICH customer asked for a product.
--
-- Catalog knew "a customer pasted this" — that is provenance, and it is about
-- how much the data can be trusted. It never knew which person was waiting, so
-- nothing could ever tell them their item was ready.
--
-- NULL when an operator added the product on spec: legal, and not the same as
-- a uuid of all zeros, which would read back as a real customer who does not
-- exist.

ALTER TABLE products ADD COLUMN requested_by uuid;
