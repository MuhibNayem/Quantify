-- 0011_accounting_depth
-- Batch-level costing, tax/returns/adjustment accounts, and ledger links on
-- returns so every stock movement can be valued and posted.

ALTER TABLE public.batches
  ADD COLUMN IF NOT EXISTS unit_cost numeric(14,4) DEFAULT 0;

ALTER TABLE public.order_items
  ADD COLUMN IF NOT EXISTS unit_cost numeric(14,4) DEFAULT 0;

ALTER TABLE public.stock_adjustments
  ADD COLUMN IF NOT EXISTS unit_cost numeric(14,4) DEFAULT 0;

ALTER TABLE public.returns
  ADD COLUMN IF NOT EXISTS journal_entry_id bigint;

ALTER TABLE public.purchase_returns
  ADD COLUMN IF NOT EXISTS journal_entry_id bigint;

INSERT INTO public.accounts (code, name, type, description, is_active, system, created_at, updated_at)
VALUES
  ('2100', 'Sales Tax Payable', 'LIABILITY', 'Tax collected at checkout and owed to the tax authority', true, true, NOW(), NOW()),
  ('4200', 'Sales Returns', 'REVENUE', 'Customer refunds (contra-revenue)', true, true, NOW(), NOW()),
  ('4900', 'Inventory Adjustment Gain', 'REVENUE', 'Gains from positive stock adjustments', true, true, NOW(), NOW()),
  ('5100', 'Inventory Shrinkage', 'EXPENSE', 'Damage, loss, and write-off expense', true, true, NOW(), NOW())
ON CONFLICT (code) DO NOTHING;
