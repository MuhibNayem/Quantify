-- General ledger: chart of accounts, journal entries, and journal lines.

CREATE TABLE accounts (
    id bigserial PRIMARY KEY,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    code text NOT NULL,
    name text NOT NULL,
    type text NOT NULL,
    description text,
    is_active boolean DEFAULT true,
    system boolean DEFAULT false
);
CREATE UNIQUE INDEX idx_accounts_code ON accounts (code);
CREATE INDEX idx_accounts_type ON accounts (type);
CREATE INDEX idx_accounts_deleted_at ON accounts (deleted_at);

CREATE TABLE journal_entries (
    id bigserial PRIMARY KEY,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    date timestamp with time zone NOT NULL,
    reference text,
    description text,
    source_type text,
    source_id bigint,
    posted_by bigint
);
CREATE INDEX idx_journal_entries_date ON journal_entries (date);
CREATE INDEX idx_journal_entries_reference ON journal_entries (reference);
CREATE INDEX idx_journal_entries_source_type ON journal_entries (source_type);
CREATE INDEX idx_journal_entries_deleted_at ON journal_entries (deleted_at);

CREATE TABLE journal_lines (
    id bigserial PRIMARY KEY,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    journal_entry_id bigint NOT NULL,
    account_id bigint NOT NULL,
    debit numeric(19,4) DEFAULT 0,
    credit numeric(19,4) DEFAULT 0,
    description text
);
CREATE INDEX idx_journal_lines_journal_entry_id ON journal_lines (journal_entry_id);
CREATE INDEX idx_journal_lines_account_id ON journal_lines (account_id);
CREATE INDEX idx_journal_lines_deleted_at ON journal_lines (deleted_at);

-- Foreign keys (both on journal_lines).
ALTER TABLE journal_lines ADD CONSTRAINT fk_journal_entries_lines FOREIGN KEY (journal_entry_id) REFERENCES journal_entries(id);
ALTER TABLE journal_lines ADD CONSTRAINT fk_journal_lines_account FOREIGN KEY (account_id) REFERENCES accounts(id);
