ALTER TABLE payment_allocations DROP CONSTRAINT IF EXISTS fk_payment_allocations_accounts_receivable;
ALTER TABLE payment_allocations DROP CONSTRAINT IF EXISTS fk_payments_allocations;
ALTER TABLE payments DROP CONSTRAINT IF EXISTS fk_payments_customer;
ALTER TABLE accounts_receivables DROP CONSTRAINT IF EXISTS fk_accounts_receivables_order;
ALTER TABLE accounts_receivables DROP CONSTRAINT IF EXISTS fk_accounts_receivables_customer;
DROP TABLE IF EXISTS payment_allocations;
DROP TABLE IF EXISTS payments;
DROP TABLE IF EXISTS accounts_receivables;
