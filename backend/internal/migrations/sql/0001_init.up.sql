--
-- PostgreSQL database dump
--

-- Dumped from database version 14.22 (Homebrew)
-- Dumped by pg_dump version 14.22 (Homebrew)

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', 'public', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: alert_role_subscriptions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.alert_role_subscriptions (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    alert_type text NOT NULL,
    role text NOT NULL
);


--
-- Name: alert_role_subscriptions_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.alert_role_subscriptions_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: alert_role_subscriptions_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.alert_role_subscriptions_id_seq OWNED BY public.alert_role_subscriptions.id;


--
-- Name: alerts; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.alerts (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    product_id bigint NOT NULL,
    type text NOT NULL,
    message text NOT NULL,
    triggered_at timestamp with time zone,
    status text DEFAULT 'ACTIVE'::text,
    batch_id bigint
);


--
-- Name: alerts_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.alerts_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: alerts_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.alerts_id_seq OWNED BY public.alerts.id;


--
-- Name: audit_logs; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.audit_logs (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    action text NOT NULL,
    entity text NOT NULL,
    entity_id text NOT NULL,
    user_id bigint NOT NULL,
    changes text,
    ip_address text,
    user_agent text,
    "timestamp" timestamp with time zone
);


--
-- Name: audit_logs_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.audit_logs_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: audit_logs_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.audit_logs_id_seq OWNED BY public.audit_logs.id;


--
-- Name: batches; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.batches (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    product_id bigint NOT NULL,
    location_id bigint,
    batch_number text NOT NULL,
    quantity bigint NOT NULL,
    expiry_date timestamp with time zone
);


--
-- Name: batches_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.batches_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: batches_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.batches_id_seq OWNED BY public.batches.id;


--
-- Name: cash_drawer_sessions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.cash_drawer_sessions (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    user_id bigint NOT NULL,
    location_id bigint NOT NULL,
    start_time timestamp with time zone NOT NULL,
    end_time timestamp with time zone,
    starting_cash numeric NOT NULL,
    ending_cash numeric,
    system_cash numeric,
    variance numeric,
    notes text,
    status text DEFAULT 'OPEN'::text,
    total_sales numeric,
    total_refunds numeric,
    total_drops numeric
);


--
-- Name: cash_drawer_sessions_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.cash_drawer_sessions_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: cash_drawer_sessions_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.cash_drawer_sessions_id_seq OWNED BY public.cash_drawer_sessions.id;


--
-- Name: cash_drops; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.cash_drops (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    session_id bigint NOT NULL,
    amount numeric NOT NULL,
    reason text,
    dropped_by bigint NOT NULL,
    dropped_at timestamp with time zone
);


--
-- Name: cash_drops_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.cash_drops_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: cash_drops_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.cash_drops_id_seq OWNED BY public.cash_drops.id;


--
-- Name: categories; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.categories (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    name text NOT NULL
);


--
-- Name: categories_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.categories_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: categories_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.categories_id_seq OWNED BY public.categories.id;


--
-- Name: demand_forecasts; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.demand_forecasts (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    product_id bigint NOT NULL,
    forecast_period text NOT NULL,
    predicted_demand bigint NOT NULL,
    confidence_score numeric,
    reasoning text,
    generated_at timestamp with time zone
);


--
-- Name: demand_forecasts_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.demand_forecasts_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: demand_forecasts_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.demand_forecasts_id_seq OWNED BY public.demand_forecasts.id;


--
-- Name: global_search_entries; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.global_search_entries (
    id bigint NOT NULL,
    entity_type text,
    entity_id bigint,
    content text,
    search_vector tsvector,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);


--
-- Name: global_search_entries_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.global_search_entries_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: global_search_entries_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.global_search_entries_id_seq OWNED BY public.global_search_entries.id;


--
-- Name: jobs; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.jobs (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    type text,
    status text,
    payload text,
    result text,
    last_error text,
    retry_count bigint,
    max_retries bigint,
    last_attempt_at timestamp with time zone
);


--
-- Name: jobs_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.jobs_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: jobs_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.jobs_id_seq OWNED BY public.jobs.id;


--
-- Name: locations; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.locations (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    name text NOT NULL,
    address text
);


--
-- Name: locations_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.locations_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: locations_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.locations_id_seq OWNED BY public.locations.id;


--
-- Name: loyalty_accounts; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.loyalty_accounts (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    user_id bigint NOT NULL,
    points bigint DEFAULT 0,
    tier text DEFAULT 'Bronze'::text
);


--
-- Name: loyalty_accounts_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.loyalty_accounts_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: loyalty_accounts_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.loyalty_accounts_id_seq OWNED BY public.loyalty_accounts.id;


--
-- Name: notifications; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.notifications (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    user_id bigint NOT NULL,
    type text NOT NULL,
    title text NOT NULL,
    message text NOT NULL,
    payload text,
    is_read boolean DEFAULT false,
    read_at timestamp with time zone,
    triggered_at timestamp with time zone
);


--
-- Name: notifications_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.notifications_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: notifications_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.notifications_id_seq OWNED BY public.notifications.id;


--
-- Name: order_items; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.order_items (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    order_id bigint NOT NULL,
    product_id bigint NOT NULL,
    quantity bigint NOT NULL,
    unit_price numeric NOT NULL,
    total_price numeric NOT NULL,
    is_returned boolean DEFAULT false,
    returned_qty bigint DEFAULT 0
);


--
-- Name: order_items_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.order_items_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: order_items_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.order_items_id_seq OWNED BY public.order_items.id;


--
-- Name: orders; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.orders (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    order_number text NOT NULL,
    user_id bigint,
    customer_id bigint,
    subtotal numeric(12,2) DEFAULT 0,
    tax_amount numeric(10,2) DEFAULT 0,
    total_amount numeric NOT NULL,
    status text DEFAULT 'COMPLETED'::text,
    payment_method text NOT NULL,
    order_date timestamp with time zone,
    points_redeemed bigint DEFAULT 0,
    discount_amount numeric DEFAULT 0
);


--
-- Name: orders_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.orders_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: orders_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.orders_id_seq OWNED BY public.orders.id;


--
-- Name: permissions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.permissions (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    name text NOT NULL,
    description text,
    "group" text
);


--
-- Name: permissions_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.permissions_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: permissions_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.permissions_id_seq OWNED BY public.permissions.id;


--
-- Name: product_alert_settings; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.product_alert_settings (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    product_id bigint NOT NULL,
    low_stock_level bigint,
    over_stock_level bigint,
    expiry_alert_days bigint
);


--
-- Name: product_alert_settings_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.product_alert_settings_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: product_alert_settings_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.product_alert_settings_id_seq OWNED BY public.product_alert_settings.id;


--
-- Name: products; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.products (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    sku text NOT NULL,
    name text NOT NULL,
    description text,
    category_id bigint,
    sub_category_id bigint,
    supplier_id bigint,
    brand text,
    purchase_price numeric,
    selling_price numeric,
    barcode_upc text,
    image_urls text,
    status text DEFAULT 'Active'::text,
    location_id bigint
);


--
-- Name: products_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.products_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: products_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.products_id_seq OWNED BY public.products.id;


--
-- Name: promotions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.promotions (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    name text NOT NULL,
    description text,
    discount_type text NOT NULL,
    discount_value numeric NOT NULL,
    start_date timestamp with time zone NOT NULL,
    end_date timestamp with time zone NOT NULL,
    is_active boolean DEFAULT true,
    priority bigint DEFAULT 0,
    product_id bigint,
    category_id bigint,
    sub_category_id bigint
);


--
-- Name: promotions_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.promotions_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: promotions_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.promotions_id_seq OWNED BY public.promotions.id;


--
-- Name: purchase_order_items; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.purchase_order_items (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    purchase_order_id bigint NOT NULL,
    product_id bigint NOT NULL,
    ordered_quantity bigint NOT NULL,
    received_quantity bigint DEFAULT 0,
    unit_price numeric NOT NULL
);


--
-- Name: purchase_order_items_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.purchase_order_items_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: purchase_order_items_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.purchase_order_items_id_seq OWNED BY public.purchase_order_items.id;


--
-- Name: purchase_orders; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.purchase_orders (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    supplier_id bigint NOT NULL,
    status text DEFAULT 'DRAFT'::text,
    order_date timestamp with time zone,
    expected_delivery_date timestamp with time zone,
    actual_delivery_date timestamp with time zone,
    created_by bigint,
    approved_by bigint,
    approved_at timestamp with time zone
);


--
-- Name: purchase_orders_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.purchase_orders_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: purchase_orders_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.purchase_orders_id_seq OWNED BY public.purchase_orders.id;


--
-- Name: purchase_return_items; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.purchase_return_items (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    purchase_return_id bigint NOT NULL,
    product_id bigint NOT NULL,
    quantity bigint NOT NULL,
    reason text NOT NULL,
    batch_id bigint
);


--
-- Name: purchase_return_items_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.purchase_return_items_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: purchase_return_items_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.purchase_return_items_id_seq OWNED BY public.purchase_return_items.id;


--
-- Name: purchase_returns; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.purchase_returns (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    purchase_order_id bigint,
    supplier_id bigint NOT NULL,
    status text DEFAULT 'PENDING'::text,
    reason text,
    refund_amount numeric,
    returned_by bigint,
    returned_at timestamp with time zone,
    approved_by bigint,
    approved_at timestamp with time zone,
    completed_at timestamp with time zone,
    cancelled_at timestamp with time zone
);


--
-- Name: purchase_returns_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.purchase_returns_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: purchase_returns_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.purchase_returns_id_seq OWNED BY public.purchase_returns.id;


--
-- Name: reorder_suggestions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.reorder_suggestions (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    product_id bigint NOT NULL,
    supplier_id bigint NOT NULL,
    current_stock bigint,
    predicted_demand bigint,
    suggested_order_quantity bigint NOT NULL,
    lead_time_days bigint,
    snapshot_price numeric(12,2) DEFAULT 0,
    status text DEFAULT 'PENDING'::text,
    suggested_at timestamp with time zone
);


--
-- Name: reorder_suggestions_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.reorder_suggestions_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: reorder_suggestions_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.reorder_suggestions_id_seq OWNED BY public.reorder_suggestions.id;


--
-- Name: return_items; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.return_items (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    return_id bigint NOT NULL,
    order_item_id bigint NOT NULL,
    product_id bigint NOT NULL,
    quantity bigint NOT NULL,
    condition text DEFAULT 'GOOD'::text,
    reason text NOT NULL
);


--
-- Name: return_items_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.return_items_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: return_items_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.return_items_id_seq OWNED BY public.return_items.id;


--
-- Name: returns; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.returns (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    order_id bigint NOT NULL,
    user_id bigint NOT NULL,
    status text DEFAULT 'PENDING'::text,
    reason text NOT NULL,
    refund_amount numeric NOT NULL,
    approved_by bigint,
    approved_at timestamp with time zone
);


--
-- Name: returns_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.returns_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: returns_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.returns_id_seq OWNED BY public.returns.id;


--
-- Name: role_permissions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.role_permissions (
    role_id bigint NOT NULL,
    permission_id bigint NOT NULL
);


--
-- Name: roles; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.roles (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    name text NOT NULL,
    description text,
    is_system boolean DEFAULT false
);


--
-- Name: roles_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.roles_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: roles_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.roles_id_seq OWNED BY public.roles.id;


--
-- Name: stock_adjustments; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.stock_adjustments (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    product_id bigint NOT NULL,
    location_id bigint,
    type text NOT NULL,
    quantity bigint NOT NULL,
    reason_code text NOT NULL,
    notes text,
    adjusted_by bigint,
    adjusted_at timestamp with time zone,
    previous_quantity bigint,
    new_quantity bigint
);


--
-- Name: stock_adjustments_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.stock_adjustments_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: stock_adjustments_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.stock_adjustments_id_seq OWNED BY public.stock_adjustments.id;


--
-- Name: stock_transfers; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.stock_transfers (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    product_id bigint NOT NULL,
    source_location_id bigint NOT NULL,
    dest_location_id bigint NOT NULL,
    quantity bigint NOT NULL,
    status text DEFAULT 'PENDING'::text,
    transferred_by bigint,
    transferred_at timestamp with time zone
);


--
-- Name: stock_transfers_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.stock_transfers_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: stock_transfers_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.stock_transfers_id_seq OWNED BY public.stock_transfers.id;


--
-- Name: sub_categories; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.sub_categories (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    name text NOT NULL,
    category_id bigint NOT NULL
);


--
-- Name: sub_categories_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.sub_categories_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: sub_categories_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.sub_categories_id_seq OWNED BY public.sub_categories.id;


--
-- Name: suppliers; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.suppliers (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    name text NOT NULL,
    contact_person text,
    email text,
    phone text,
    address text
);


--
-- Name: suppliers_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.suppliers_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: suppliers_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.suppliers_id_seq OWNED BY public.suppliers.id;


--
-- Name: system_settings; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.system_settings (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    key text NOT NULL,
    value text,
    "group" text,
    type text NOT NULL,
    description text
);


--
-- Name: system_settings_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.system_settings_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: system_settings_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.system_settings_id_seq OWNED BY public.system_settings.id;


--
-- Name: time_clocks; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.time_clocks (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    user_id bigint NOT NULL,
    clock_in timestamp with time zone,
    clock_out timestamp with time zone,
    break_start timestamp with time zone,
    break_end timestamp with time zone,
    status text DEFAULT 'CLOCKED_IN'::text,
    notes text
);


--
-- Name: time_clocks_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.time_clocks_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: time_clocks_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.time_clocks_id_seq OWNED BY public.time_clocks.id;


--
-- Name: transactions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.transactions (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    order_id text NOT NULL,
    amount bigint NOT NULL,
    currency text NOT NULL,
    payment_method text NOT NULL,
    status text NOT NULL,
    gateway_transaction_id text NOT NULL
);


--
-- Name: transactions_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.transactions_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: transactions_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.transactions_id_seq OWNED BY public.transactions.id;


--
-- Name: user_notification_settings; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.user_notification_settings (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    user_id bigint NOT NULL,
    email_notifications_enabled boolean,
    sms_notifications_enabled boolean,
    email_address text,
    phone_number text
);


--
-- Name: user_notification_settings_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.user_notification_settings_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: user_notification_settings_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.user_notification_settings_id_seq OWNED BY public.user_notification_settings.id;


--
-- Name: user_notification_subscriptions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.user_notification_subscriptions (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    user_id bigint NOT NULL,
    type text NOT NULL,
    enabled boolean NOT NULL
);


--
-- Name: user_notification_subscriptions_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.user_notification_subscriptions_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: user_notification_subscriptions_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.user_notification_subscriptions_id_seq OWNED BY public.user_notification_subscriptions.id;


--
-- Name: users; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.users (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    username text NOT NULL,
    password text NOT NULL,
    role text,
    role_id bigint DEFAULT 0 NOT NULL,
    is_active boolean DEFAULT false,
    first_name text,
    last_name text,
    email text,
    phone_number text,
    address text
);


--
-- Name: users_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.users_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: users_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.users_id_seq OWNED BY public.users.id;


--
-- Name: alert_role_subscriptions id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.alert_role_subscriptions ALTER COLUMN id SET DEFAULT nextval('public.alert_role_subscriptions_id_seq'::regclass);


--
-- Name: alerts id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.alerts ALTER COLUMN id SET DEFAULT nextval('public.alerts_id_seq'::regclass);


--
-- Name: audit_logs id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.audit_logs ALTER COLUMN id SET DEFAULT nextval('public.audit_logs_id_seq'::regclass);


--
-- Name: batches id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.batches ALTER COLUMN id SET DEFAULT nextval('public.batches_id_seq'::regclass);


--
-- Name: cash_drawer_sessions id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.cash_drawer_sessions ALTER COLUMN id SET DEFAULT nextval('public.cash_drawer_sessions_id_seq'::regclass);


--
-- Name: cash_drops id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.cash_drops ALTER COLUMN id SET DEFAULT nextval('public.cash_drops_id_seq'::regclass);


--
-- Name: categories id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.categories ALTER COLUMN id SET DEFAULT nextval('public.categories_id_seq'::regclass);


--
-- Name: demand_forecasts id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.demand_forecasts ALTER COLUMN id SET DEFAULT nextval('public.demand_forecasts_id_seq'::regclass);


--
-- Name: global_search_entries id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.global_search_entries ALTER COLUMN id SET DEFAULT nextval('public.global_search_entries_id_seq'::regclass);


--
-- Name: jobs id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.jobs ALTER COLUMN id SET DEFAULT nextval('public.jobs_id_seq'::regclass);


--
-- Name: locations id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.locations ALTER COLUMN id SET DEFAULT nextval('public.locations_id_seq'::regclass);


--
-- Name: loyalty_accounts id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.loyalty_accounts ALTER COLUMN id SET DEFAULT nextval('public.loyalty_accounts_id_seq'::regclass);


--
-- Name: notifications id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.notifications ALTER COLUMN id SET DEFAULT nextval('public.notifications_id_seq'::regclass);


--
-- Name: order_items id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.order_items ALTER COLUMN id SET DEFAULT nextval('public.order_items_id_seq'::regclass);


--
-- Name: orders id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.orders ALTER COLUMN id SET DEFAULT nextval('public.orders_id_seq'::regclass);


--
-- Name: permissions id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.permissions ALTER COLUMN id SET DEFAULT nextval('public.permissions_id_seq'::regclass);


--
-- Name: product_alert_settings id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.product_alert_settings ALTER COLUMN id SET DEFAULT nextval('public.product_alert_settings_id_seq'::regclass);


--
-- Name: products id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.products ALTER COLUMN id SET DEFAULT nextval('public.products_id_seq'::regclass);


--
-- Name: promotions id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.promotions ALTER COLUMN id SET DEFAULT nextval('public.promotions_id_seq'::regclass);


--
-- Name: purchase_order_items id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.purchase_order_items ALTER COLUMN id SET DEFAULT nextval('public.purchase_order_items_id_seq'::regclass);


--
-- Name: purchase_orders id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.purchase_orders ALTER COLUMN id SET DEFAULT nextval('public.purchase_orders_id_seq'::regclass);


--
-- Name: purchase_return_items id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.purchase_return_items ALTER COLUMN id SET DEFAULT nextval('public.purchase_return_items_id_seq'::regclass);


--
-- Name: purchase_returns id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.purchase_returns ALTER COLUMN id SET DEFAULT nextval('public.purchase_returns_id_seq'::regclass);


--
-- Name: reorder_suggestions id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.reorder_suggestions ALTER COLUMN id SET DEFAULT nextval('public.reorder_suggestions_id_seq'::regclass);


--
-- Name: return_items id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.return_items ALTER COLUMN id SET DEFAULT nextval('public.return_items_id_seq'::regclass);


--
-- Name: returns id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.returns ALTER COLUMN id SET DEFAULT nextval('public.returns_id_seq'::regclass);


--
-- Name: roles id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.roles ALTER COLUMN id SET DEFAULT nextval('public.roles_id_seq'::regclass);


--
-- Name: stock_adjustments id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.stock_adjustments ALTER COLUMN id SET DEFAULT nextval('public.stock_adjustments_id_seq'::regclass);


--
-- Name: stock_transfers id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.stock_transfers ALTER COLUMN id SET DEFAULT nextval('public.stock_transfers_id_seq'::regclass);


--
-- Name: sub_categories id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.sub_categories ALTER COLUMN id SET DEFAULT nextval('public.sub_categories_id_seq'::regclass);


--
-- Name: suppliers id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.suppliers ALTER COLUMN id SET DEFAULT nextval('public.suppliers_id_seq'::regclass);


--
-- Name: system_settings id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.system_settings ALTER COLUMN id SET DEFAULT nextval('public.system_settings_id_seq'::regclass);


--
-- Name: time_clocks id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.time_clocks ALTER COLUMN id SET DEFAULT nextval('public.time_clocks_id_seq'::regclass);


--
-- Name: transactions id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.transactions ALTER COLUMN id SET DEFAULT nextval('public.transactions_id_seq'::regclass);


--
-- Name: user_notification_settings id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_notification_settings ALTER COLUMN id SET DEFAULT nextval('public.user_notification_settings_id_seq'::regclass);


--
-- Name: user_notification_subscriptions id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_notification_subscriptions ALTER COLUMN id SET DEFAULT nextval('public.user_notification_subscriptions_id_seq'::regclass);


--
-- Name: users id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users ALTER COLUMN id SET DEFAULT nextval('public.users_id_seq'::regclass);


--
-- Name: alert_role_subscriptions alert_role_subscriptions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.alert_role_subscriptions
    ADD CONSTRAINT alert_role_subscriptions_pkey PRIMARY KEY (id);


--
-- Name: alerts alerts_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.alerts
    ADD CONSTRAINT alerts_pkey PRIMARY KEY (id);


--
-- Name: audit_logs audit_logs_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.audit_logs
    ADD CONSTRAINT audit_logs_pkey PRIMARY KEY (id);


--
-- Name: batches batches_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.batches
    ADD CONSTRAINT batches_pkey PRIMARY KEY (id);


--
-- Name: cash_drawer_sessions cash_drawer_sessions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.cash_drawer_sessions
    ADD CONSTRAINT cash_drawer_sessions_pkey PRIMARY KEY (id);


--
-- Name: cash_drops cash_drops_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.cash_drops
    ADD CONSTRAINT cash_drops_pkey PRIMARY KEY (id);


--
-- Name: categories categories_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.categories
    ADD CONSTRAINT categories_pkey PRIMARY KEY (id);


--
-- Name: demand_forecasts demand_forecasts_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.demand_forecasts
    ADD CONSTRAINT demand_forecasts_pkey PRIMARY KEY (id);


--
-- Name: global_search_entries global_search_entries_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.global_search_entries
    ADD CONSTRAINT global_search_entries_pkey PRIMARY KEY (id);


--
-- Name: jobs jobs_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.jobs
    ADD CONSTRAINT jobs_pkey PRIMARY KEY (id);


--
-- Name: locations locations_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.locations
    ADD CONSTRAINT locations_pkey PRIMARY KEY (id);


--
-- Name: loyalty_accounts loyalty_accounts_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.loyalty_accounts
    ADD CONSTRAINT loyalty_accounts_pkey PRIMARY KEY (id);


--
-- Name: notifications notifications_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.notifications
    ADD CONSTRAINT notifications_pkey PRIMARY KEY (id);


--
-- Name: order_items order_items_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.order_items
    ADD CONSTRAINT order_items_pkey PRIMARY KEY (id);


--
-- Name: orders orders_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.orders
    ADD CONSTRAINT orders_pkey PRIMARY KEY (id);


--
-- Name: permissions permissions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.permissions
    ADD CONSTRAINT permissions_pkey PRIMARY KEY (id);


--
-- Name: product_alert_settings product_alert_settings_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.product_alert_settings
    ADD CONSTRAINT product_alert_settings_pkey PRIMARY KEY (id);


--
-- Name: products products_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.products
    ADD CONSTRAINT products_pkey PRIMARY KEY (id);


--
-- Name: promotions promotions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.promotions
    ADD CONSTRAINT promotions_pkey PRIMARY KEY (id);


--
-- Name: purchase_order_items purchase_order_items_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.purchase_order_items
    ADD CONSTRAINT purchase_order_items_pkey PRIMARY KEY (id);


--
-- Name: purchase_orders purchase_orders_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.purchase_orders
    ADD CONSTRAINT purchase_orders_pkey PRIMARY KEY (id);


--
-- Name: purchase_return_items purchase_return_items_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.purchase_return_items
    ADD CONSTRAINT purchase_return_items_pkey PRIMARY KEY (id);


--
-- Name: purchase_returns purchase_returns_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.purchase_returns
    ADD CONSTRAINT purchase_returns_pkey PRIMARY KEY (id);


--
-- Name: reorder_suggestions reorder_suggestions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.reorder_suggestions
    ADD CONSTRAINT reorder_suggestions_pkey PRIMARY KEY (id);


--
-- Name: return_items return_items_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.return_items
    ADD CONSTRAINT return_items_pkey PRIMARY KEY (id);


--
-- Name: returns returns_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.returns
    ADD CONSTRAINT returns_pkey PRIMARY KEY (id);


--
-- Name: role_permissions role_permissions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.role_permissions
    ADD CONSTRAINT role_permissions_pkey PRIMARY KEY (role_id, permission_id);


--
-- Name: roles roles_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.roles
    ADD CONSTRAINT roles_pkey PRIMARY KEY (id);


--
-- Name: stock_adjustments stock_adjustments_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.stock_adjustments
    ADD CONSTRAINT stock_adjustments_pkey PRIMARY KEY (id);


--
-- Name: stock_transfers stock_transfers_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.stock_transfers
    ADD CONSTRAINT stock_transfers_pkey PRIMARY KEY (id);


--
-- Name: sub_categories sub_categories_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.sub_categories
    ADD CONSTRAINT sub_categories_pkey PRIMARY KEY (id);


--
-- Name: suppliers suppliers_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.suppliers
    ADD CONSTRAINT suppliers_pkey PRIMARY KEY (id);


--
-- Name: system_settings system_settings_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.system_settings
    ADD CONSTRAINT system_settings_pkey PRIMARY KEY (id);


--
-- Name: time_clocks time_clocks_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.time_clocks
    ADD CONSTRAINT time_clocks_pkey PRIMARY KEY (id);


--
-- Name: transactions transactions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.transactions
    ADD CONSTRAINT transactions_pkey PRIMARY KEY (id);


--
-- Name: user_notification_settings user_notification_settings_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_notification_settings
    ADD CONSTRAINT user_notification_settings_pkey PRIMARY KEY (id);


--
-- Name: user_notification_subscriptions user_notification_subscriptions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_notification_subscriptions
    ADD CONSTRAINT user_notification_subscriptions_pkey PRIMARY KEY (id);


--
-- Name: users users_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);


--
-- Name: idx_alert_product_status; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_alert_product_status ON public.alerts USING btree (product_id, status);


--
-- Name: idx_alert_role; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_alert_role ON public.alert_role_subscriptions USING btree (alert_type, role);


--
-- Name: idx_alert_role_subscriptions_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_alert_role_subscriptions_deleted_at ON public.alert_role_subscriptions USING btree (deleted_at);


--
-- Name: idx_alerts_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_alerts_deleted_at ON public.alerts USING btree (deleted_at);


--
-- Name: idx_alerts_triggered_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_alerts_triggered_at ON public.alerts USING btree (triggered_at);


--
-- Name: idx_alerts_type; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_alerts_type ON public.alerts USING btree (type);


--
-- Name: idx_audit_logs_action; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_audit_logs_action ON public.audit_logs USING btree (action);


--
-- Name: idx_audit_logs_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_audit_logs_deleted_at ON public.audit_logs USING btree (deleted_at);


--
-- Name: idx_audit_logs_entity; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_audit_logs_entity ON public.audit_logs USING btree (entity);


--
-- Name: idx_audit_logs_entity_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_audit_logs_entity_id ON public.audit_logs USING btree (entity_id);


--
-- Name: idx_audit_logs_timestamp; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_audit_logs_timestamp ON public.audit_logs USING btree ("timestamp");


--
-- Name: idx_audit_logs_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_audit_logs_user_id ON public.audit_logs USING btree (user_id);


--
-- Name: idx_batches_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_batches_deleted_at ON public.batches USING btree (deleted_at);


--
-- Name: idx_cash_drawer_sessions_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_cash_drawer_sessions_deleted_at ON public.cash_drawer_sessions USING btree (deleted_at);


--
-- Name: idx_cash_drawer_sessions_location_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_cash_drawer_sessions_location_id ON public.cash_drawer_sessions USING btree (location_id);


--
-- Name: idx_cash_drawer_sessions_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_cash_drawer_sessions_user_id ON public.cash_drawer_sessions USING btree (user_id);


--
-- Name: idx_cash_drops_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_cash_drops_deleted_at ON public.cash_drops USING btree (deleted_at);


--
-- Name: idx_cash_drops_session_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_cash_drops_session_id ON public.cash_drops USING btree (session_id);


--
-- Name: idx_categories_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_categories_deleted_at ON public.categories USING btree (deleted_at);


--
-- Name: idx_categories_name; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_categories_name ON public.categories USING btree (name);


--
-- Name: idx_demand_forecasts_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_demand_forecasts_deleted_at ON public.demand_forecasts USING btree (deleted_at);


--
-- Name: idx_entity; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_entity ON public.global_search_entries USING btree (entity_type, entity_id);


--
-- Name: idx_global_search_entries_search_vector; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_global_search_entries_search_vector ON public.global_search_entries USING btree (search_vector);


--
-- Name: idx_jobs_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_jobs_deleted_at ON public.jobs USING btree (deleted_at);


--
-- Name: idx_locations_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_locations_deleted_at ON public.locations USING btree (deleted_at);


--
-- Name: idx_locations_name; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_locations_name ON public.locations USING btree (name);


--
-- Name: idx_loyalty_accounts_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_loyalty_accounts_deleted_at ON public.loyalty_accounts USING btree (deleted_at);


--
-- Name: idx_loyalty_accounts_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_loyalty_accounts_user_id ON public.loyalty_accounts USING btree (user_id);


--
-- Name: idx_notifications_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_notifications_deleted_at ON public.notifications USING btree (deleted_at);


--
-- Name: idx_notifications_triggered_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_notifications_triggered_at ON public.notifications USING btree (triggered_at);


--
-- Name: idx_notifications_type; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_notifications_type ON public.notifications USING btree (type);


--
-- Name: idx_order_items_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_order_items_deleted_at ON public.order_items USING btree (deleted_at);


--
-- Name: idx_order_items_order_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_order_items_order_id ON public.order_items USING btree (order_id);


--
-- Name: idx_order_items_product_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_order_items_product_id ON public.order_items USING btree (product_id);


--
-- Name: idx_orders_customer_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_orders_customer_id ON public.orders USING btree (customer_id);


--
-- Name: idx_orders_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_orders_deleted_at ON public.orders USING btree (deleted_at);


--
-- Name: idx_orders_order_date; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_orders_order_date ON public.orders USING btree (order_date);


--
-- Name: idx_orders_order_number; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_orders_order_number ON public.orders USING btree (order_number);


--
-- Name: idx_orders_status; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_orders_status ON public.orders USING btree (status);


--
-- Name: idx_orders_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_orders_user_id ON public.orders USING btree (user_id);


--
-- Name: idx_permissions_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_permissions_deleted_at ON public.permissions USING btree (deleted_at);


--
-- Name: idx_permissions_name; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_permissions_name ON public.permissions USING btree (name);


--
-- Name: idx_product_alert_settings_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_product_alert_settings_deleted_at ON public.product_alert_settings USING btree (deleted_at);


--
-- Name: idx_product_alert_settings_product_id; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_product_alert_settings_product_id ON public.product_alert_settings USING btree (product_id);


--
-- Name: idx_product_location; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_product_location ON public.batches USING btree (product_id, location_id);


--
-- Name: idx_products_barcode_upc; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_products_barcode_upc ON public.products USING btree (barcode_upc);


--
-- Name: idx_products_barcode_upc_not_null; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_products_barcode_upc_not_null ON public.products USING btree (barcode_upc) WHERE ((barcode_upc IS NOT NULL) AND (barcode_upc <> ''::text));


--
-- Name: idx_products_brand; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_products_brand ON public.products USING btree (brand);


--
-- Name: idx_products_category_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_products_category_id ON public.products USING btree (category_id);


--
-- Name: idx_products_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_products_deleted_at ON public.products USING btree (deleted_at);


--
-- Name: idx_products_location_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_products_location_id ON public.products USING btree (location_id);


--
-- Name: idx_products_name; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_products_name ON public.products USING btree (name);


--
-- Name: idx_products_sku; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_products_sku ON public.products USING btree (sku);


--
-- Name: idx_products_status; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_products_status ON public.products USING btree (status);


--
-- Name: idx_products_sub_category_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_products_sub_category_id ON public.products USING btree (sub_category_id);


--
-- Name: idx_products_supplier_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_products_supplier_id ON public.products USING btree (supplier_id);


--
-- Name: idx_promotions_category_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_promotions_category_id ON public.promotions USING btree (category_id);


--
-- Name: idx_promotions_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_promotions_deleted_at ON public.promotions USING btree (deleted_at);


--
-- Name: idx_promotions_product_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_promotions_product_id ON public.promotions USING btree (product_id);


--
-- Name: idx_promotions_sub_category_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_promotions_sub_category_id ON public.promotions USING btree (sub_category_id);


--
-- Name: idx_purchase_order_items_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_purchase_order_items_deleted_at ON public.purchase_order_items USING btree (deleted_at);


--
-- Name: idx_purchase_orders_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_purchase_orders_deleted_at ON public.purchase_orders USING btree (deleted_at);


--
-- Name: idx_purchase_return_items_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_purchase_return_items_deleted_at ON public.purchase_return_items USING btree (deleted_at);


--
-- Name: idx_purchase_return_items_purchase_return_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_purchase_return_items_purchase_return_id ON public.purchase_return_items USING btree (purchase_return_id);


--
-- Name: idx_purchase_returns_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_purchase_returns_deleted_at ON public.purchase_returns USING btree (deleted_at);


--
-- Name: idx_purchase_returns_purchase_order_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_purchase_returns_purchase_order_id ON public.purchase_returns USING btree (purchase_order_id);


--
-- Name: idx_purchase_returns_supplier_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_purchase_returns_supplier_id ON public.purchase_returns USING btree (supplier_id);


--
-- Name: idx_reorder_suggestions_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_reorder_suggestions_deleted_at ON public.reorder_suggestions USING btree (deleted_at);


--
-- Name: idx_return_items_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_return_items_deleted_at ON public.return_items USING btree (deleted_at);


--
-- Name: idx_return_items_order_item_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_return_items_order_item_id ON public.return_items USING btree (order_item_id);


--
-- Name: idx_return_items_return_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_return_items_return_id ON public.return_items USING btree (return_id);


--
-- Name: idx_returns_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_returns_deleted_at ON public.returns USING btree (deleted_at);


--
-- Name: idx_returns_order_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_returns_order_id ON public.returns USING btree (order_id);


--
-- Name: idx_returns_status; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_returns_status ON public.returns USING btree (status);


--
-- Name: idx_returns_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_returns_user_id ON public.returns USING btree (user_id);


--
-- Name: idx_roles_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_roles_deleted_at ON public.roles USING btree (deleted_at);


--
-- Name: idx_roles_name; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_roles_name ON public.roles USING btree (name);


--
-- Name: idx_stock_adjustments_adjusted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_stock_adjustments_adjusted_at ON public.stock_adjustments USING btree (adjusted_at);


--
-- Name: idx_stock_adjustments_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_stock_adjustments_deleted_at ON public.stock_adjustments USING btree (deleted_at);


--
-- Name: idx_stock_adjustments_product_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_stock_adjustments_product_id ON public.stock_adjustments USING btree (product_id);


--
-- Name: idx_stock_transfers_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_stock_transfers_deleted_at ON public.stock_transfers USING btree (deleted_at);


--
-- Name: idx_sub_categories_category_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_sub_categories_category_id ON public.sub_categories USING btree (category_id);


--
-- Name: idx_sub_categories_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_sub_categories_deleted_at ON public.sub_categories USING btree (deleted_at);


--
-- Name: idx_sub_categories_name; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_sub_categories_name ON public.sub_categories USING btree (name);


--
-- Name: idx_suppliers_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_suppliers_deleted_at ON public.suppliers USING btree (deleted_at);


--
-- Name: idx_suppliers_email; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_suppliers_email ON public.suppliers USING btree (email);


--
-- Name: idx_suppliers_name; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_suppliers_name ON public.suppliers USING btree (name);


--
-- Name: idx_suppliers_phone; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_suppliers_phone ON public.suppliers USING btree (phone);


--
-- Name: idx_system_settings_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_system_settings_deleted_at ON public.system_settings USING btree (deleted_at);


--
-- Name: idx_system_settings_group; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_system_settings_group ON public.system_settings USING btree ("group");


--
-- Name: idx_system_settings_key; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_system_settings_key ON public.system_settings USING btree (key);


--
-- Name: idx_time_clocks_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_time_clocks_deleted_at ON public.time_clocks USING btree (deleted_at);


--
-- Name: idx_transactions_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_transactions_deleted_at ON public.transactions USING btree (deleted_at);


--
-- Name: idx_transactions_gateway_transaction_id; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_transactions_gateway_transaction_id ON public.transactions USING btree (gateway_transaction_id);


--
-- Name: idx_user_notification_settings_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_user_notification_settings_deleted_at ON public.user_notification_settings USING btree (deleted_at);


--
-- Name: idx_user_notification_settings_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_user_notification_settings_user_id ON public.user_notification_settings USING btree (user_id);


--
-- Name: idx_user_notification_subscriptions_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_user_notification_subscriptions_deleted_at ON public.user_notification_subscriptions USING btree (deleted_at);


--
-- Name: idx_user_read; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_user_read ON public.notifications USING btree (user_id, is_read);


--
-- Name: idx_user_type; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_user_type ON public.user_notification_subscriptions USING btree (user_id, type);


--
-- Name: idx_users_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_users_deleted_at ON public.users USING btree (deleted_at);


--
-- Name: idx_users_email; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_users_email ON public.users USING btree (email);


--
-- Name: idx_users_phone_number; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_users_phone_number ON public.users USING btree (phone_number);


--
-- Name: idx_users_role_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_users_role_id ON public.users USING btree (role_id);


--
-- Name: idx_users_username; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_users_username ON public.users USING btree (username);


--
-- Name: alerts fk_alerts_batch; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.alerts
    ADD CONSTRAINT fk_alerts_batch FOREIGN KEY (batch_id) REFERENCES public.batches(id);


--
-- Name: alerts fk_alerts_product; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.alerts
    ADD CONSTRAINT fk_alerts_product FOREIGN KEY (product_id) REFERENCES public.products(id);


--
-- Name: audit_logs fk_audit_logs_user; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.audit_logs
    ADD CONSTRAINT fk_audit_logs_user FOREIGN KEY (user_id) REFERENCES public.users(id);


--
-- Name: batches fk_batches_location; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.batches
    ADD CONSTRAINT fk_batches_location FOREIGN KEY (location_id) REFERENCES public.locations(id);


--
-- Name: batches fk_batches_product; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.batches
    ADD CONSTRAINT fk_batches_product FOREIGN KEY (product_id) REFERENCES public.products(id);


--
-- Name: cash_drawer_sessions fk_cash_drawer_sessions_location; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.cash_drawer_sessions
    ADD CONSTRAINT fk_cash_drawer_sessions_location FOREIGN KEY (location_id) REFERENCES public.locations(id);


--
-- Name: cash_drawer_sessions fk_cash_drawer_sessions_user; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.cash_drawer_sessions
    ADD CONSTRAINT fk_cash_drawer_sessions_user FOREIGN KEY (user_id) REFERENCES public.users(id);


--
-- Name: cash_drops fk_cash_drops_session; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.cash_drops
    ADD CONSTRAINT fk_cash_drops_session FOREIGN KEY (session_id) REFERENCES public.cash_drawer_sessions(id);


--
-- Name: sub_categories fk_categories_sub_categories; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.sub_categories
    ADD CONSTRAINT fk_categories_sub_categories FOREIGN KEY (category_id) REFERENCES public.categories(id);


--
-- Name: demand_forecasts fk_demand_forecasts_product; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.demand_forecasts
    ADD CONSTRAINT fk_demand_forecasts_product FOREIGN KEY (product_id) REFERENCES public.products(id);


--
-- Name: notifications fk_notifications_user; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.notifications
    ADD CONSTRAINT fk_notifications_user FOREIGN KEY (user_id) REFERENCES public.users(id);


--
-- Name: order_items fk_order_items_product; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.order_items
    ADD CONSTRAINT fk_order_items_product FOREIGN KEY (product_id) REFERENCES public.products(id);


--
-- Name: orders fk_orders_customer; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.orders
    ADD CONSTRAINT fk_orders_customer FOREIGN KEY (customer_id) REFERENCES public.users(id);


--
-- Name: order_items fk_orders_order_items; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.order_items
    ADD CONSTRAINT fk_orders_order_items FOREIGN KEY (order_id) REFERENCES public.orders(id);


--
-- Name: returns fk_orders_returns; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.returns
    ADD CONSTRAINT fk_orders_returns FOREIGN KEY (order_id) REFERENCES public.orders(id);


--
-- Name: orders fk_orders_user; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.orders
    ADD CONSTRAINT fk_orders_user FOREIGN KEY (user_id) REFERENCES public.users(id);


--
-- Name: product_alert_settings fk_product_alert_settings_product; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.product_alert_settings
    ADD CONSTRAINT fk_product_alert_settings_product FOREIGN KEY (product_id) REFERENCES public.products(id);


--
-- Name: products fk_products_category; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.products
    ADD CONSTRAINT fk_products_category FOREIGN KEY (category_id) REFERENCES public.categories(id);


--
-- Name: products fk_products_location; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.products
    ADD CONSTRAINT fk_products_location FOREIGN KEY (location_id) REFERENCES public.locations(id);


--
-- Name: products fk_products_sub_category; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.products
    ADD CONSTRAINT fk_products_sub_category FOREIGN KEY (sub_category_id) REFERENCES public.sub_categories(id);


--
-- Name: products fk_products_supplier; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.products
    ADD CONSTRAINT fk_products_supplier FOREIGN KEY (supplier_id) REFERENCES public.suppliers(id);


--
-- Name: promotions fk_promotions_category; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.promotions
    ADD CONSTRAINT fk_promotions_category FOREIGN KEY (category_id) REFERENCES public.categories(id);


--
-- Name: promotions fk_promotions_product; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.promotions
    ADD CONSTRAINT fk_promotions_product FOREIGN KEY (product_id) REFERENCES public.products(id);


--
-- Name: promotions fk_promotions_sub_category; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.promotions
    ADD CONSTRAINT fk_promotions_sub_category FOREIGN KEY (sub_category_id) REFERENCES public.sub_categories(id);


--
-- Name: purchase_order_items fk_purchase_order_items_product; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.purchase_order_items
    ADD CONSTRAINT fk_purchase_order_items_product FOREIGN KEY (product_id) REFERENCES public.products(id);


--
-- Name: purchase_order_items fk_purchase_orders_purchase_order_items; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.purchase_order_items
    ADD CONSTRAINT fk_purchase_orders_purchase_order_items FOREIGN KEY (purchase_order_id) REFERENCES public.purchase_orders(id);


--
-- Name: purchase_orders fk_purchase_orders_supplier; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.purchase_orders
    ADD CONSTRAINT fk_purchase_orders_supplier FOREIGN KEY (supplier_id) REFERENCES public.suppliers(id);


--
-- Name: purchase_return_items fk_purchase_return_items_batch; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.purchase_return_items
    ADD CONSTRAINT fk_purchase_return_items_batch FOREIGN KEY (batch_id) REFERENCES public.batches(id);


--
-- Name: purchase_return_items fk_purchase_return_items_product; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.purchase_return_items
    ADD CONSTRAINT fk_purchase_return_items_product FOREIGN KEY (product_id) REFERENCES public.products(id);


--
-- Name: purchase_returns fk_purchase_returns_purchase_order; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.purchase_returns
    ADD CONSTRAINT fk_purchase_returns_purchase_order FOREIGN KEY (purchase_order_id) REFERENCES public.purchase_orders(id);


--
-- Name: purchase_return_items fk_purchase_returns_purchase_return_items; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.purchase_return_items
    ADD CONSTRAINT fk_purchase_returns_purchase_return_items FOREIGN KEY (purchase_return_id) REFERENCES public.purchase_returns(id);


--
-- Name: purchase_returns fk_purchase_returns_supplier; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.purchase_returns
    ADD CONSTRAINT fk_purchase_returns_supplier FOREIGN KEY (supplier_id) REFERENCES public.suppliers(id);


--
-- Name: reorder_suggestions fk_reorder_suggestions_product; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.reorder_suggestions
    ADD CONSTRAINT fk_reorder_suggestions_product FOREIGN KEY (product_id) REFERENCES public.products(id);


--
-- Name: reorder_suggestions fk_reorder_suggestions_supplier; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.reorder_suggestions
    ADD CONSTRAINT fk_reorder_suggestions_supplier FOREIGN KEY (supplier_id) REFERENCES public.suppliers(id);


--
-- Name: return_items fk_return_items_order_item; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.return_items
    ADD CONSTRAINT fk_return_items_order_item FOREIGN KEY (order_item_id) REFERENCES public.order_items(id);


--
-- Name: return_items fk_return_items_product; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.return_items
    ADD CONSTRAINT fk_return_items_product FOREIGN KEY (product_id) REFERENCES public.products(id);


--
-- Name: return_items fk_returns_return_items; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.return_items
    ADD CONSTRAINT fk_returns_return_items FOREIGN KEY (return_id) REFERENCES public.returns(id);


--
-- Name: returns fk_returns_user; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.returns
    ADD CONSTRAINT fk_returns_user FOREIGN KEY (user_id) REFERENCES public.users(id);


--
-- Name: role_permissions fk_role_permissions_permission; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.role_permissions
    ADD CONSTRAINT fk_role_permissions_permission FOREIGN KEY (permission_id) REFERENCES public.permissions(id);


--
-- Name: role_permissions fk_role_permissions_role; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.role_permissions
    ADD CONSTRAINT fk_role_permissions_role FOREIGN KEY (role_id) REFERENCES public.roles(id);


--
-- Name: stock_adjustments fk_stock_adjustments_location; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.stock_adjustments
    ADD CONSTRAINT fk_stock_adjustments_location FOREIGN KEY (location_id) REFERENCES public.locations(id);


--
-- Name: stock_adjustments fk_stock_adjustments_product; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.stock_adjustments
    ADD CONSTRAINT fk_stock_adjustments_product FOREIGN KEY (product_id) REFERENCES public.products(id);


--
-- Name: stock_transfers fk_stock_transfers_dest_location; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.stock_transfers
    ADD CONSTRAINT fk_stock_transfers_dest_location FOREIGN KEY (dest_location_id) REFERENCES public.locations(id);


--
-- Name: stock_transfers fk_stock_transfers_product; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.stock_transfers
    ADD CONSTRAINT fk_stock_transfers_product FOREIGN KEY (product_id) REFERENCES public.products(id);


--
-- Name: stock_transfers fk_stock_transfers_source_location; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.stock_transfers
    ADD CONSTRAINT fk_stock_transfers_source_location FOREIGN KEY (source_location_id) REFERENCES public.locations(id);


--
-- Name: time_clocks fk_time_clocks_user; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.time_clocks
    ADD CONSTRAINT fk_time_clocks_user FOREIGN KEY (user_id) REFERENCES public.users(id);


--
-- Name: user_notification_settings fk_user_notification_settings_user; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_notification_settings
    ADD CONSTRAINT fk_user_notification_settings_user FOREIGN KEY (user_id) REFERENCES public.users(id);


--
-- Name: loyalty_accounts fk_users_loyalty_account; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.loyalty_accounts
    ADD CONSTRAINT fk_users_loyalty_account FOREIGN KEY (user_id) REFERENCES public.users(id);


--
-- Name: users fk_users_role; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT fk_users_role FOREIGN KEY (role_id) REFERENCES public.roles(id);


--
-- PostgreSQL database dump complete
--

