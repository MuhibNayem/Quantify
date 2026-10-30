ALTER TABLE public.purchase_returns DROP COLUMN IF EXISTS journal_entry_id;
ALTER TABLE public.returns DROP COLUMN IF EXISTS journal_entry_id;
ALTER TABLE public.stock_adjustments DROP COLUMN IF EXISTS unit_cost;
ALTER TABLE public.order_items DROP COLUMN IF EXISTS unit_cost;
ALTER TABLE public.batches DROP COLUMN IF EXISTS unit_cost;

-- Reference accounts are left in place: journal lines may reference them, and
-- deleting a used account would violate the ledger's referential integrity.
