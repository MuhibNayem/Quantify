-- Units of measure: dimensions, units, and per-product selling units.

CREATE TABLE unit_dimensions (
    id bigserial PRIMARY KEY,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    code text NOT NULL,
    name text NOT NULL,
    base_unit_id bigint
);
CREATE UNIQUE INDEX idx_unit_dimensions_code ON unit_dimensions (code);
CREATE INDEX idx_unit_dimensions_deleted_at ON unit_dimensions (deleted_at);

CREATE TABLE units (
    id bigserial PRIMARY KEY,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    code text NOT NULL,
    name text NOT NULL,
    symbol text,
    dimension_id bigint NOT NULL,
    factor numeric DEFAULT 1 NOT NULL,
    is_base boolean DEFAULT false
);
CREATE UNIQUE INDEX idx_units_code ON units (code);
CREATE INDEX idx_units_dimension_id ON units (dimension_id);
CREATE INDEX idx_units_deleted_at ON units (deleted_at);

CREATE TABLE product_units (
    id bigserial PRIMARY KEY,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    product_id bigint NOT NULL,
    unit_id bigint NOT NULL,
    factor numeric DEFAULT 1 NOT NULL,
    price numeric DEFAULT 0,
    purchase_price numeric DEFAULT 0,
    barcode text,
    is_active boolean DEFAULT true
);
CREATE UNIQUE INDEX idx_product_unit ON product_units (product_id, unit_id);
CREATE INDEX idx_product_units_barcode ON product_units (barcode);
CREATE INDEX idx_product_units_deleted_at ON product_units (deleted_at);

-- Products gain a base stock unit and a fractional-allowed flag.
ALTER TABLE products ADD COLUMN base_unit_id bigint;
ALTER TABLE products ADD COLUMN allow_fractional boolean DEFAULT false;
CREATE INDEX idx_products_base_unit_id ON products (base_unit_id);
ALTER TABLE products ADD CONSTRAINT fk_products_base_unit FOREIGN KEY (base_unit_id) REFERENCES units(id);

-- Foreign keys (match GORM AutoMigrate referential integrity).
ALTER TABLE units ADD CONSTRAINT fk_units_dimension FOREIGN KEY (dimension_id) REFERENCES unit_dimensions(id);
ALTER TABLE product_units ADD CONSTRAINT fk_product_units_unit FOREIGN KEY (unit_id) REFERENCES units(id);
ALTER TABLE product_units ADD CONSTRAINT fk_products_product_units FOREIGN KEY (product_id) REFERENCES products(id);
ALTER TABLE order_items ADD CONSTRAINT fk_order_items_unit FOREIGN KEY (unit_id) REFERENCES units(id);
