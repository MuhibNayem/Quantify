#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(dirname "$SCRIPT_DIR")"

echo "=== Quantify Setup ==="

if ! command -v docker &>/dev/null; then
  echo "Error: docker is not installed" >&2
  exit 1
fi

if ! docker info &>/dev/null; then
  echo "Error: docker daemon is not running" >&2
  exit 1
fi

if [ ! -f "$ROOT_DIR/.env" ]; then
  echo "Creating .env from .env.example..."
  cp "$ROOT_DIR/.env.example" "$ROOT_DIR/.env"
fi

if [ ! -f "$ROOT_DIR/backend/.env" ]; then
  echo "Creating backend/.env from backend/.env.example..."
  cp "$ROOT_DIR/backend/.env.example" "$ROOT_DIR/backend/.env"
fi

if [ ! -f "$ROOT_DIR/ai-service/.env" ]; then
  echo "Warning: ai-service/.env does not exist. Create it with your API keys."
fi

echo "Pulling Docker images..."
docker compose pull

echo "Starting services..."
docker compose up -d --build

echo "Waiting for services to be ready..."
"$SCRIPT_DIR/wait-for-it.sh" localhost "${POSTGRES_PORT:-5432}" -t 60
"$SCRIPT_DIR/wait-for-it.sh" localhost "${BACKEND_PORT:-8080}" -t 90

echo "Running migrations..."
make migrate

echo ""
echo "=== Setup Complete ==="
echo "Backend:    http://localhost:${BACKEND_PORT:-8080}"
echo "AI Service: http://localhost:${AI_SERVICE_PORT:-8001}"
echo "Grafana:    http://localhost:${GRAFANA_PORT:-3001}"
echo "pgAdmin:    http://localhost:${PGADMIN_PORT:-5050}"
echo "Mailhog:    http://localhost:${MAILHOG_UI_PORT:-8025}"
echo "MinIO:      http://localhost:${MINIO_CONSOLE_PORT:-9001}"
echo "RabbitMQ:   http://localhost:${RABBITMQ_MANAGEMENT_PORT:-15672}"
