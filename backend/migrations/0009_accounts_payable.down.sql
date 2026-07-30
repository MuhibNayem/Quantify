ALTER TABLE supplier_payment_allocations DROP CONSTRAINT IF EXISTS fk_supplier_payment_allocations_accounts_payable;
ALTER TABLE supplier_payment_allocations DROP CONSTRAINT IF EXISTS fk_supplier_payments_allocations;
ALTER TABLE supplier_payments DROP CONSTRAINT IF EXISTS fk_supplier_payments_supplier;
ALTER TABLE accounts_payables DROP CONSTRAINT IF EXISTS fk_accounts_payables_purchase_order;
ALTER TABLE accounts_payables DROP CONSTRAINT IF EXISTS fk_accounts_payables_supplier;
DROP TABLE IF EXISTS supplier_payment_allocations;
DROP TABLE IF EXISTS supplier_payments;
DROP TABLE IF EXISTS accounts_payables;
