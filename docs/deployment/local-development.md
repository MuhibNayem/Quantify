# Local Development

Run the full Quantify stack on your machine with Docker Compose.

---

## Prerequisites

| Tool | Version | Purpose |
| :--- | :--- | :--- |
| Docker Desktop | ≥ 4.x | Containers + Compose |
| Make | ≥ 4.x | Task runner |
| Go | ≥ 1.24 | Backend builds, migrations |
| Python | ≥ 3.11 | AI service (optional, runs in container) |
| Git | ≥ 2.x | Source control |

Docker Desktop must be running with at least **8 GB RAM** and **4 CPUs**
allocated (the full stack runs ~15 containers).

---

## First-time setup

```bash
# 1. Clone and enter the repo
git clone <repo-url> && cd Quantify

# 2. Create environment files
cp .env.example .env
cp backend/.env.example backend/.env
# Create ai-service/.env with your API keys if needed

# 3. One-command setup (pulls images, starts stack, runs migrations)
./scripts/setup.sh
```

Or step by step:

```bash
make dev        # docker compose up -d --build
make migrate    # apply schema migrations
make seed       # insert sample users/data
```

---

## Starting services

```bash
make dev        # start (or rebuild) all services
make dev-logs   # follow logs for all services
```

The override file `docker-compose.override.yml` is applied automatically and
enables:

- **Hot reload** for the backend (`go run`) and AI service (`uvicorn --reload`)
- **Debug ports**: backend Delve `:2345`, AI service debugpy `:5678`
- Anonymous Grafana admin access

---

## Running migrations

```bash
make migrate
```

This runs the Go migrator against the Compose database:

```bash
docker compose exec backend go run ./cmd/migrator \
  -dsn "postgres://user:password@db:5432/inventory_db?sslmode=disable" \
  -cmd up
```

Check the current version:

```bash
docker compose exec backend go run ./cmd/migrator \
  -dsn "postgres://user:password@db:5432/inventory_db?sslmode=disable" \
  -cmd version
```

Revert one migration:

```bash
docker compose exec backend go run ./cmd/migrator \
  -dsn "postgres://user:password@db:5432/inventory_db?sslmode=disable" \
  -cmd down
```

---

## Seeding data

```bash
make seed       # runs scripts/seed-db.sh
```

Inserts two users (idempotent, `ON CONFLICT DO NOTHING`):

| Email | Password | Role |
| :--- | :--- | :--- |
| `admin@quantify.local` | `admin123` | admin |
| `test@quantify.local` | `test123` | user |

---

## Accessing services

| Service | URL | Default credentials |
| :--- | :--- | :--- |
| Backend API | http://localhost:8080 | — |
| AI Service | http://localhost:8001 | — |
| Swagger UI | http://localhost:8080/swagger | — |
| Grafana | http://localhost:3001 | admin / password |
| Prometheus | http://localhost:9092 | — |
| Alertmanager | http://localhost:9093 | — |
| Loki | http://localhost:3100 | — |
| pgAdmin | http://localhost:5050 | admin@quantify.local / password |
| RabbitMQ Management | http://localhost:15672 | user / password |
| MinIO Console | http://localhost:9001 | minioadmin4321 / minioadmin4321 |
| Mailhog | http://localhost:8025 | — |
| PostgreSQL | localhost:5432 | user / password |
| PostgreSQL Replica | localhost:5433 | user / password |
| Redis Cluster | localhost:6379–6381 | redispass |

All ports are configurable via `.env` (see `.env.example`).

---

## Stopping services

```bash
make dev-down   # stop containers, keep volumes
make clean      # stop + remove volumes + prune images (destructive)
```

---

## Running tests

```bash
make test               # unit tests (Go + Python)
make test-integration   # integration tests (requires Docker)
make lint               # golangci-lint + ruff
```

---

## Troubleshooting

### Port already in use

```bash
lsof -i :8080   # find the process
```

Change the port in `.env` (e.g. `BACKEND_PORT=8081`) and restart.

### Redis cluster won't form

```bash
docker compose up -d redis-init   # re-run the cluster initializer
docker compose exec redis1 redis-cli -a redispass cluster info
```

Look for `cluster_state:ok`.

### Backend won't start — config validation failure

The backend **fails fast** if critical secrets are missing or set to known
defaults (`my_secret_key`, `password`, `minioadmin4321`, etc.). Update
`backend/.env` with non-default values for:

- `JWT_SECRET`
- `REFRESH_TOKEN_SECRET`
- `MINIO_ACCESS_KEY_ID` / `MINIO_SECRET_ACCESS_KEY`
- `SSLCOMMERZ_STORE_PASSWORD`
- `RABBITMQ_URL`

Generate secrets: `openssl rand -base64 48`

### Database connection refused

```bash
docker compose ps db
docker compose logs --tail=50 db
docker compose exec db pg_isready -U user -d inventory_db
```

### Full reset

```bash
make clean      # removes all volumes — all data is lost
make dev
make migrate
make seed
```

---

## Related

- [Operations runbook](../runbook.md) — incident response procedures
- [Architecture](../architecture.md) — system design
- [Security](../security.md) — secret handling
