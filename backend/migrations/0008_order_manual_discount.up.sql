-- Manual discount fields on orders (admin opt-in employee discount at checkout).

ALTER TABLE orders ADD COLUMN manual_discount_type text DEFAULT '';
ALTER TABLE orders ADD COLUMN manual_discount_value numeric DEFAULT 0;
ALTER TABLE orders ADD COLUMN manual_discount_amount numeric(12,2) DEFAULT 0;
ALTER TABLE orders ADD COLUMN manual_discount_reason text;
ALTER TABLE orders ADD COLUMN manual_discount_by bigint;
