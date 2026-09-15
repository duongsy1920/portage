-- 0012: WHO placed an order, when it was not the customer.
--
-- Attribution belongs to the order, not the quote (P10-PLAN.md §2a): an order
-- is a commitment with money attached, an accepted quote is just a price.
--
-- NULL when the customer placed it with their own token — every order before
-- this migration is exactly that, so there is nothing to backfill. NULL is
-- not the same as a uuid of all zeros, which would read back as a real
-- operator who does not exist.

ALTER TABLE orders ADD COLUMN placed_by uuid;
