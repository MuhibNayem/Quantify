-- Reverse 0003: restore product-quantity columns to bigint (their original type),
-- rounding any fractional values to the nearest whole number.

ALTER TABLE batches ALTER COLUMN quantity TYPE bigint USING round(quantity)::bigint;

ALTER TABLE stock_adjustments ALTER COLUMN quantity TYPE bigint USING round(quantity)::bigint;
ALTER TABLE stock_adjustments ALTER COLUMN previous_quantity TYPE bigint USING round(previous_quantity)::bigint;
ALTER TABLE stock_adjustments ALTER COLUMN new_quantity TYPE bigint USING round(new_quantity)::bigint;

ALTER TABLE order_items ALTER COLUMN quantity TYPE bigint USING round(quantity)::bigint;
ALTER TABLE order_items ALTER COLUMN returned_qty TYPE bigint USING round(returned_qty)::bigint;

ALTER TABLE return_items ALTER COLUMN quantity TYPE bigint USING round(quantity)::bigint;

ALTER TABLE purchase_order_items ALTER COLUMN ordered_quantity TYPE bigint USING round(ordered_quantity)::bigint;
ALTER TABLE purchase_order_items ALTER COLUMN received_quantity TYPE bigint USING round(received_quantity)::bigint;

ALTER TABLE purchase_return_items ALTER COLUMN quantity TYPE bigint USING round(quantity)::bigint;

ALTER TABLE reorder_suggestions ALTER COLUMN current_stock TYPE bigint USING round(current_stock)::bigint;
ALTER TABLE reorder_suggestions ALTER COLUMN predicted_demand TYPE bigint USING round(predicted_demand)::bigint;
ALTER TABLE reorder_suggestions ALTER COLUMN suggested_order_quantity TYPE bigint USING round(suggested_order_quantity)::bigint;

ALTER TABLE stock_transfers ALTER COLUMN quantity TYPE bigint USING round(quantity)::bigint;

ALTER TABLE demand_forecasts ALTER COLUMN predicted_demand TYPE bigint USING round(predicted_demand)::bigint;
