-- Accounts receivable: money customers owe from credit (unpaid) sales, and payments received.

CREATE TABLE accounts_receivables (
    id bigserial PRIMARY KEY,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    customer_id bigint NOT NULL,
    order_id bigint,
    amount numeric(19,4) NOT NULL,
    paid_amount numeric(19,4) DEFAULT 0,
    status text DEFAULT 'UNPAID',
    due_date timestamp with time zone,
    journal_entry_id bigint
);
CREATE INDEX idx_accounts_receivables_customer_id ON accounts_receivables (customer_id);
CREATE INDEX idx_accounts_receivables_order_id ON accounts_receivables (order_id);
CREATE INDEX idx_accounts_receivables_status ON accounts_receivables (status);
CREATE INDEX idx_accounts_receivables_deleted_at ON accounts_receivables (deleted_at);

CREATE TABLE payments (
    id bigserial PRIMARY KEY,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    customer_id bigint NOT NULL,
    amount numeric(19,4) NOT NULL,
    method text,
    received_at timestamp with time zone,
    received_by bigint,
    notes text,
    journal_entry_id bigint
);
CREATE INDEX idx_payments_customer_id ON payments (customer_id);
CREATE INDEX idx_payments_deleted_at ON payments (deleted_at);

CREATE TABLE payment_allocations (
    id bigserial PRIMARY KEY,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    payment_id bigint NOT NULL,
    accounts_receivable_id bigint NOT NULL,
    amount numeric(19,4) NOT NULL
);
CREATE INDEX idx_payment_allocations_payment_id ON payment_allocations (payment_id);
CREATE INDEX idx_payment_allocations_accounts_receivable_id ON payment_allocations (accounts_receivable_id);
CREATE INDEX idx_payment_allocations_deleted_at ON payment_allocations (deleted_at);

-- Foreign keys.
ALTER TABLE accounts_receivables ADD CONSTRAINT fk_accounts_receivables_customer FOREIGN KEY (customer_id) REFERENCES users(id);
ALTER TABLE accounts_receivables ADD CONSTRAINT fk_accounts_receivables_order FOREIGN KEY (order_id) REFERENCES orders(id);
ALTER TABLE payments ADD CONSTRAINT fk_payments_customer FOREIGN KEY (customer_id) REFERENCES users(id);
ALTER TABLE payment_allocations ADD CONSTRAINT fk_payments_allocations FOREIGN KEY (payment_id) REFERENCES payments(id);
ALTER TABLE payment_allocations ADD CONSTRAINT fk_payment_allocations_accounts_receivable FOREIGN KEY (accounts_receivable_id) REFERENCES accounts_receivables(id);
