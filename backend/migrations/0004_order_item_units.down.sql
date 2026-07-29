-- Reverse 0004: drop the unit-aware columns from order_items.

ALTER TABLE order_items DROP COLUMN IF EXISTS unit_id;
ALTER TABLE order_items DROP COLUMN IF EXISTS base_quantity;
