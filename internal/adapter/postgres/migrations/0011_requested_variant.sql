-- 0011: which size the customer asked for.
--
-- 0009 recorded WHO is waiting. This records WHAT they asked for, in their own
-- words: "US 9", "M 8 / W 9.5", "bản 256GB". Without it the person doing the
-- buying has to guess which size a paid order was for, and a guess there is a
-- wrong shoe in a box.
--
-- Free text, not a foreign key to product_variants: it is a wish typed before
-- anybody checked what the shop actually sells. Matching it to a real variant
-- is a person's job, and the mismatch is the useful part.

ALTER TABLE products ADD COLUMN requested_variant text NOT NULL DEFAULT '';

ALTER TABLE product_worklist ADD COLUMN requested_variant text NOT NULL DEFAULT '';
