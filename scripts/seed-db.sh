#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(dirname "$SCRIPT_DIR")"

source "$ROOT_DIR/.env" 2>/dev/null || true

DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${POSTGRES_PORT:-5432}"
DB_USER="${POSTGRES_USER:-user}"
DB_PASS="${POSTGRES_PASSWORD:-password}"
DB_NAME="${POSTGRES_DB:-inventory_db}"

export PGPASSWORD="$DB_PASS"

echo "Seeding database $DB_NAME on $DB_HOST:$DB_PORT..."

"$SCRIPT_DIR/wait-for-it.sh" "$DB_HOST" "$DB_PORT" -t 30

docker compose exec db psql -U "$DB_USER" -d "$DB_NAME" -c "
  INSERT INTO users (id, name, email, password, role, created_at, updated_at)
  VALUES
    (gen_random_uuid(), 'Admin User', 'admin@quantify.local', crypt('admin123', gen_salt('bf')), 'admin', NOW(), NOW()),
    (gen_random_uuid(), 'Test User', 'test@quantify.local', crypt('test123', gen_salt('bf')), 'user', NOW(), NOW())
  ON CONFLICT (email) DO NOTHING;
" 2>/dev/null && echo "Seed data inserted." || echo "Seed skipped (table may not exist yet or data already present)."
