-- Record the selling unit and base-unit equivalent on each order item so a
-- product sold in any unit (14 in, 1.2 ft, 1 coil) keeps an audit trail of
-- what was sold and how much base stock was deducted.

ALTER TABLE order_items ADD COLUMN IF NOT EXISTS unit_id bigint;
ALTER TABLE order_items ADD COLUMN IF NOT EXISTS base_quantity numeric(14,4);
