-- Drop foreign keys first (referencing units), then columns and tables.
ALTER TABLE order_items DROP CONSTRAINT IF EXISTS fk_order_items_unit;
ALTER TABLE product_units DROP CONSTRAINT IF EXISTS fk_products_product_units;
ALTER TABLE product_units DROP CONSTRAINT IF EXISTS fk_product_units_unit;
ALTER TABLE units DROP CONSTRAINT IF EXISTS fk_units_dimension;
ALTER TABLE products DROP CONSTRAINT IF EXISTS fk_products_base_unit;
DROP INDEX IF EXISTS idx_products_base_unit_id;
ALTER TABLE products DROP COLUMN IF EXISTS allow_fractional;
ALTER TABLE products DROP COLUMN IF EXISTS base_unit_id;
DROP TABLE IF EXISTS product_units;
DROP TABLE IF EXISTS units;
DROP TABLE IF EXISTS unit_dimensions;
