ALTER TABLE journal_lines DROP CONSTRAINT IF EXISTS fk_journal_lines_account;
ALTER TABLE journal_lines DROP CONSTRAINT IF EXISTS fk_journal_entries_lines;
DROP TABLE IF EXISTS journal_lines;
DROP TABLE IF EXISTS journal_entries;
DROP TABLE IF EXISTS accounts;
