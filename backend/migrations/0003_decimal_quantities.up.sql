-- Migrate inventory/sales quantity columns from integer (bigint) to numeric(14,4)
-- so the system can handle fractional quantities (e.g., sell 1.2 ft of wire, 14 in).
-- Only product-quantity columns are changed; IDs, points, and counts are untouched.

ALTER TABLE batches ALTER COLUMN quantity TYPE numeric(14,4) USING quantity::numeric;

ALTER TABLE stock_adjustments ALTER COLUMN quantity TYPE numeric(14,4) USING quantity::numeric;
ALTER TABLE stock_adjustments ALTER COLUMN previous_quantity TYPE numeric(14,4) USING previous_quantity::numeric;
ALTER TABLE stock_adjustments ALTER COLUMN new_quantity TYPE numeric(14,4) USING new_quantity::numeric;

ALTER TABLE order_items ALTER COLUMN quantity TYPE numeric(14,4) USING quantity::numeric;
ALTER TABLE order_items ALTER COLUMN returned_qty TYPE numeric(14,4) USING returned_qty::numeric;

ALTER TABLE return_items ALTER COLUMN quantity TYPE numeric(14,4) USING quantity::numeric;

ALTER TABLE purchase_order_items ALTER COLUMN ordered_quantity TYPE numeric(14,4) USING ordered_quantity::numeric;
ALTER TABLE purchase_order_items ALTER COLUMN received_quantity TYPE numeric(14,4) USING received_quantity::numeric;

ALTER TABLE purchase_return_items ALTER COLUMN quantity TYPE numeric(14,4) USING quantity::numeric;

ALTER TABLE reorder_suggestions ALTER COLUMN current_stock TYPE numeric(14,4) USING current_stock::numeric;
ALTER TABLE reorder_suggestions ALTER COLUMN predicted_demand TYPE numeric(14,4) USING predicted_demand::numeric;
ALTER TABLE reorder_suggestions ALTER COLUMN suggested_order_quantity TYPE numeric(14,4) USING suggested_order_quantity::numeric;

ALTER TABLE stock_transfers ALTER COLUMN quantity TYPE numeric(14,4) USING quantity::numeric;

ALTER TABLE demand_forecasts ALTER COLUMN predicted_demand TYPE numeric(14,4) USING predicted_demand::numeric;
