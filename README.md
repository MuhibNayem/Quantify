# Quantify

Quantify is a comprehensive inventory management and point-of-sale (POS) system
designed to streamline and automate your inventory processes. It provides a
robust set of features to manage products, track stock levels, handle orders,
and gain insights into your business performance.

## Core Features

- **Product Management:** Create, update, and manage your product catalog, including categories, sub-categories, and suppliers.
- **Stock Control:** Keep inventory accurate with batch-level tracking (FEFO allocation), stock adjustments, history, and stock transfers between locations.
- **Order Management:** Manage procurement with purchase orders (create, approve, send, receive) and vendor returns.
- **Point of Sale:** Checkout with promotions, tax, and loyalty; customer returns with approval workflow.
- **User Management:** Secure user access with registration, login, and granular role-based permissions (RBAC).
- **Insightful Reporting:** Sales trends, inventory turnover, profit margins, GMROI, shrinkage, basket analysis, and more — with cached and async (job-based) reports.
- **Automated Alerts:** Low stock, out-of-stock, overstock, and expiry alerts with per-role and per-user subscriptions.
- **Bulk Operations:** Async, confirmation-gated bulk import and export for product data.
- **Barcode Support:** 1D barcode lookup and generation for quick product access.
- **Demand Forecasting:** AI-assisted demand forecasting and reorder suggestions (via the `ai-service`).
- **Real-time Notifications:** WebSocket notifications on order status changes, returns, and alerts.

## Technical Overview

The backend is built with Go and utilizes the following technologies:

- **Framework:** Gin Gonic
- **Database:** PostgreSQL (GORM; versioned migrations via golang-migrate)
- **Cache & Token Store:** Redis cluster
- **Message Broker:** RabbitMQ for asynchronous task processing (bulk jobs, reports, alerts)
- **Object Storage:** MinIO for import/export and report files
- **Authentication:** httpOnly cookie sessions (JWT access + rotating refresh) with a Redis
  token whitelist and double-submit CSRF; Bearer tokens remain supported for API clients
- **Rate limiting:** Redis-backed per-IP limits (global + stricter auth endpoints)
- **Observability:** Prometheus metrics (`/metrics`) and optional OpenTelemetry tracing
- **API:** A versioned RESTful API (`/api/v1`)

## Documentation

Specification and operations docs live in [`docs/`](./docs/README.md):

| Document | Contents |
| :--- | :--- |
| [docs/README.md](./docs/README.md) | Documentation index |
| [docs/architecture.md](./docs/architecture.md) | Layering, ADRs, deployment topology, request lifecycle |
| [docs/business_logic.md](./docs/business_logic.md) | RBAC, checkout, returns, bulk import, loyalty |
| [docs/security.md](./docs/security.md) | Threat model (STRIDE-lite), controls, key rotation, secret handling |
| [docs/runbook.md](./docs/runbook.md) | Incident response, deploy, and rollback procedures |
| [docs/access_control.md](./docs/access_control.md) | Permission-to-route mapping |
| [docs/reports_user_guide.md](./docs/reports_user_guide.md) | Reporting reference |

The API is documented with the OpenAPI specification in
[`backend/api/openapi.yaml`](./backend/api/openapi.yaml). A Swagger UI is served
at `/swagger/` when the backend is running.

## Getting Started

You will need Go (1.25+), Docker, and Docker Compose installed.

1. **Clone the repository:**
   ```bash
   git clone https://github.com/your-username/quantify.git
   cd quantify
   ```

2. **Configure the environment:**
   ```bash
   cd backend
   cp .env.example .env
   ```
   Edit `.env` and set every required secret to a strong, non-default value.
   The backend **fails fast** on boot if `JWT_SECRET`, `REFRESH_TOKEN_SECRET`,
   `MINIO_ACCESS_KEY_ID`, `MINIO_SECRET_ACCESS_KEY`, or
   `SSLCOMMERZ_STORE_PASSWORD` are missing or set to known defaults. See
   [docs/security.md](./docs/security.md) for details.

3. **Run the stack:**
   ```bash
   # From the repository root:
   docker-compose up --build
   ```
   The backend will be available at `http://localhost:8080`.

### Dev quickstart (run the backend locally against Dockerized infra)

```bash
# 1. Start infrastructure only (Postgres, Redis cluster, RabbitMQ, MinIO)
docker-compose up -d db redis1 redis2 redis3 redis-init rabbitmq minio

# 2. Apply database migrations
cd backend
export DATABASE_URL="postgres://user:password@localhost:5432/inventory_db?sslmode=disable"
go run ./cmd/migrator -cmd up -dsn "$DATABASE_URL"

# 3. Run the backend (APP_ENV=dev enables GORM AutoMigrate for convenience)
APP_ENV=dev SERVER_PORT=8080 \
  DB_HOST=localhost REDIS_CLUSTER_ADDRS=localhost:6379 \
  RABBITMQ_URL="amqp://user:password@localhost:5672/" \
  go run ./cmd/server
```

The server listens on `:8080`. Health and readiness are at `/health` and
`/ready`; `/ready` checks Postgres, Redis, RabbitMQ, and MinIO.

## Testing

The CI pipeline (`.github/workflows/ci.yml`) runs these checks. Run them
locally before pushing:

```bash
cd backend

go build ./...          # compile
go vet ./...            # static analysis
go test -race ./...     # unit + integration tests with the race detector

# Lint and security scans (as used in CI)
golangci-lint run       # golangci-lint
gosec ./...             # gosec static security analysis
```

Integration tests under `backend/test/integration/` expect a reachable
Postgres (see the `postgres` service in the CI workflow).

## Session & CSRF Protection

Login sets two `HttpOnly` cookies — `access_token` (8 h) and `refresh_token`
(rotating, scoped to `/api/v1/users`) — plus a readable `csrf_token` cookie.
The SPA never stores or reads JWTs. `POST /users/login` returns only the user,
permissions, and the non-secret CSRF token. Every state-changing request
(`POST`, `PUT`, `PATCH`, `DELETE`) must echo the CSRF cookie in an
`X-CSRF-Token` header; tokens rotate on refresh and are invalidated on logout
and on role/permission changes. Bearer tokens remain supported for API clients.

## WebSocket Authentication

Browser sessions authenticate from the `access_token` cookie sent during the
WebSocket handshake, then send `{"type":"auth"}` as the first message within
5 seconds. API clients may instead include a token in that message
(`{"type":"auth","token":"<JWT>"}`). The legacy `?token=` query parameter is
ignored. See [docs/architecture.md](./docs/architecture.md) for the full request
lifecycle.

## Environment Variables

Configuration is loaded by `backend/internal/config/config.go` in precedence
order: built-in defaults → optional TOML file → environment variables.

### Database

| Variable | Default | Description |
| :--- | :--- | :--- |
| `DB_HOST` | `localhost` | Postgres host. |
| `DB_PORT` | `5432` | Postgres port. |
| `DB_USER` | `user` | Postgres user. |
| `DB_PASSWORD` | `password` | Postgres password. |
| `DB_NAME` | `inventory_db` | Postgres database name. |

### Server & CORS

| Variable | Default | Description |
| :--- | :--- | :--- |
| `SERVER_PORT` | `8080` | HTTP listen port. |
| `CORS_ALLOWED_ORIGINS` | `http://localhost:5173` | CSV of allowed CORS origins. |
| `WS_ALLOWED_ORIGINS` | `http://localhost:5173` | CSV of allowed WebSocket `Origin` values. |
| `APP_ENV` | `dev` | Environment. GORM `AutoMigrate` runs only when `dev`. |
| `SERVICE_NAME` | `quantify-backend` | Service name for telemetry. |
| `OTLP_ENDPOINT` | _(empty)_ | OTel collector endpoint; tracing disabled when empty. |

### Redis

| Variable | Default | Description |
| :--- | :--- | :--- |
| `REDIS_ADDR` | _(empty)_ | Single-node Redis address (optional). |
| `REDIS_CLUSTER_ADDRS` | `redis1:6379,redis2:6379,redis3:6379` | CSV of Redis cluster node addresses. |
| `REDIS_PASSWORD` | `redispass` | Redis password. |

### Messaging & Async

| Variable | Default | Description |
| :--- | :--- | :--- |
| `RABBITMQ_URL` | _(empty, **required**)_ | AMQP connection URL. |
| `CONSUMER_CONCURRENCY` | `1` | Number of bulk-consumer workers. |
| `AI_SERVICE_URL` | `http://localhost:8001` | Base URL of the Python AI service. |
| `RATE_LIMIT_REQUESTS` | `300` | Global requests allowed per client IP per window. |
| `RATE_LIMIT_WINDOW` | `1m` | Global rate-limit window. |
| `AUTH_RATE_LIMIT_REQUESTS` | `10` | Login/register/refresh requests per IP per window. |
| `AUTH_RATE_LIMIT_WINDOW` | `1m` | Auth rate-limit window. |
| `TRUSTED_PROXIES` | _(empty)_ | CSV of proxy CIDRs trusted for `X-Forwarded-For` (set this behind ingress). |

### Auth & Security (required — must not be defaults)

| Variable | Default | Description |
| :--- | :--- | :--- |
| `JWT_SECRET` | `my_secret_key` | Access-token signing secret. **Must be changed.** |
| `REFRESH_TOKEN_SECRET` | `my_refresh_secret_key` | Refresh-token signing secret. **Must be changed.** |
| `REFRESH_TOKEN_TTL` | `720h` | Refresh-token lifetime (Go duration). |

### Email (SMTP)

| Variable | Default | Description |
| :--- | :--- | :--- |
| `SMTP_HOST` | _(empty)_ | SMTP host. |
| `SMTP_PORT` | `587` | SMTP port. |
| `SMTP_USER` | _(empty)_ | SMTP username. |
| `SMTP_PASS` | _(empty)_ | SMTP password. |
| `SMTPSENDER` | _(empty)_ | Sender address. |

### Payment gateways

| Variable | Default | Description |
| :--- | :--- | :--- |
| `SSLCOMMERZ_STORE_ID` | _(empty)_ | SSLCommerz store ID. |
| `SSLCOMMERZ_STORE_PASSWORD` | _(empty, **required**)_ | Used for gateway auth and IPN signature verification. |
| `SSLCOMMERZ_API_HOST` | `https://sandbox.sslcommerz.com` | Gateway API host. |
| `SSLCOMMERZ_SUCCESS_URL` | `http://localhost:8080/payment/success` | Success callback. |
| `SSLCOMMERZ_FAIL_URL` | `http://localhost:8080/payment/fail` | Failure callback. |
| `SSLCOMMERZ_CANCEL_URL` | `http://localhost:8080/payment/cancel` | Cancel callback. |
| `SSLCOMMERZ_IPN_URL` | `http://localhost:8080/payment/ipn` | IPN callback. |
| `BKASH_API_KEY` | _(empty)_ | bKash API key. |
| `BKASH_API_SECRET` | _(empty)_ | bKash API secret. |
| `BKASH_USERNAME` | _(empty)_ | bKash username. |
| `BKASH_PASSWORD` | _(empty)_ | bKash password. |
| `BKASH_BASE_URL` | `https://sandbox.bkash.com/v1.2.0-beta` | bKash base URL. |

### MinIO object storage

| Variable | Default | Description |
| :--- | :--- | :--- |
| `MINIO_ENDPOINT` | `localhost:9000` | MinIO endpoint (host:port). |
| `MINIO_ACCESS_KEY_ID` | `minioadmin` | Access key. **Must be changed.** |
| `MINIO_SECRET_ACCESS_KEY` | `minioadmin` | Secret key. **Must be changed.** |
| `MINIO_BUCKET_NAME` | `reports` | Bucket for report files. |
| `MINIO_USE_TLS` | `false` | Set `true` for HTTPS endpoints. |

### Reporting cache TTLs (Go duration)

| Variable | Default | Description |
| :--- | :--- | :--- |
| `SALES_TRENDS_CACHE_TTL` | `5m` | Cache TTL for sales trends report. |
| `INVENTORY_TURNOVER_CACHE_TTL` | `5m` | Cache TTL for inventory turnover report. |
| `PROFIT_MARGIN_CACHE_TTL` | `5m` | Cache TTL for profit margin report. |

### Special-purpose variables (read directly, not via the config struct)

| Variable | Where | Description |
| :--- | :--- | :--- |
| `DATABASE_URL` | `cmd/migrator` | Postgres DSN for migrations. |
| `RUN_MIGRATIONS` | `bootstrap` | When `true`, populates the search index on startup. |
| `AI_AGENT_PASSWORD` | `ai-service` | Password for the seeded AI agent user (printed once at first backend boot). Leave empty in `.env.example`. |
