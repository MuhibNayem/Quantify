-- Accounts payable: money we owe suppliers from received purchase orders, and payments made.

CREATE TABLE accounts_payables (
    id bigserial PRIMARY KEY,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    supplier_id bigint NOT NULL,
    purchase_order_id bigint,
    amount numeric(19,4) NOT NULL,
    paid_amount numeric(19,4) DEFAULT 0,
    status text DEFAULT 'UNPAID',
    due_date timestamp with time zone,
    journal_entry_id bigint
);
CREATE INDEX idx_accounts_payables_supplier_id ON accounts_payables (supplier_id);
CREATE INDEX idx_accounts_payables_purchase_order_id ON accounts_payables (purchase_order_id);
CREATE INDEX idx_accounts_payables_status ON accounts_payables (status);
CREATE INDEX idx_accounts_payables_deleted_at ON accounts_payables (deleted_at);

CREATE TABLE supplier_payments (
    id bigserial PRIMARY KEY,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    supplier_id bigint NOT NULL,
    amount numeric(19,4) NOT NULL,
    method text,
    paid_at timestamp with time zone,
    paid_by bigint,
    notes text,
    journal_entry_id bigint
);
CREATE INDEX idx_supplier_payments_supplier_id ON supplier_payments (supplier_id);
CREATE INDEX idx_supplier_payments_deleted_at ON supplier_payments (deleted_at);

CREATE TABLE supplier_payment_allocations (
    id bigserial PRIMARY KEY,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    supplier_payment_id bigint NOT NULL,
    accounts_payable_id bigint NOT NULL,
    amount numeric(19,4) NOT NULL
);
CREATE INDEX idx_supplier_payment_allocations_supplier_payment_id ON supplier_payment_allocations (supplier_payment_id);
CREATE INDEX idx_supplier_payment_allocations_accounts_payable_id ON supplier_payment_allocations (accounts_payable_id);
CREATE INDEX idx_supplier_payment_allocations_deleted_at ON supplier_payment_allocations (deleted_at);

-- Foreign keys.
ALTER TABLE accounts_payables ADD CONSTRAINT fk_accounts_payables_supplier FOREIGN KEY (supplier_id) REFERENCES suppliers(id);
ALTER TABLE accounts_payables ADD CONSTRAINT fk_accounts_payables_purchase_order FOREIGN KEY (purchase_order_id) REFERENCES purchase_orders(id);
ALTER TABLE supplier_payments ADD CONSTRAINT fk_supplier_payments_supplier FOREIGN KEY (supplier_id) REFERENCES suppliers(id);
ALTER TABLE supplier_payment_allocations ADD CONSTRAINT fk_supplier_payments_allocations FOREIGN KEY (supplier_payment_id) REFERENCES supplier_payments(id);
ALTER TABLE supplier_payment_allocations ADD CONSTRAINT fk_supplier_payment_allocations_accounts_payable FOREIGN KEY (accounts_payable_id) REFERENCES accounts_payables(id);
