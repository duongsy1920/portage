-- 0013: WHO placed an order, on the read model too (P10).
--
-- ordering.orders already has this (0012); order_summaries is a separate
-- projection with its own copy, filled by the projector from
-- ordering.order_placed — never joined back to the write side (see 0007).
--
-- NULL when the customer placed it themselves — same rule as 0012, same
-- reason: a uuid of all zeros would read back as a real operator.

ALTER TABLE order_summaries ADD COLUMN placed_by uuid;
