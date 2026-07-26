# Changelog

All notable changes to the Quantify backend are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).

## [Phase 1] 2026-XX-XX — Security hardening

### Changed

- **P1.T1** Random AI agent password generated at first boot; no hard-coded bcrypt hash in source. Password printed once via logrus at seed time. (`internal/migrations/seed_ai_user.go`)
- **P1.T2** SSLCommerz IPN callback now verifies the **MD5 signature** per the actual gateway protocol (with `store_passwd=<md5(store_password)>` as a signed parameter and `&` as separator). Replaces the previous "string status check" and the earlier wrong HMAC-SHA512 attempt. Transaction persistence restored. Single env var `SSLCOMMERZ_STORE_PASSWORD` is now used for both outgoing auth and IPN verification. (`internal/services/payment_service.go`, `internal/handlers/payment.go`)
- **P1.T3** WebSocket connections require an `Origin` header matching the configured allowlist (env `WS_ALLOWED_ORIGINS`, default `http://localhost:5173`). Empty origins are rejected. (`internal/handlers/websocket.go`, `internal/config/config.go`)
- **P1.T4** Refresh token TTL is now configurable via `REFRESH_TOKEN_TTL` env var (Go duration format, default `720h` = 30 days). Replaces the previous hard-coded 10-year TTL. Boot fails if the value is not parseable. (`internal/handlers/user.go`, `internal/config/config.go`)
- **P1.T5** Permission cache TTL reduced from 1 hour to 60 seconds, and `InvalidateRoleCache` is now called from every mutation path (`AssignPermissions`, `UpdateRole`). Affected users' CSRF tokens are also rotated. (`internal/repository/role_repository.go`, `internal/repository/role_repository_ext.go`, `internal/middleware/permission.go`)
- **P1.T6** AI agent user is created with `IsActive=false`. Operator must call `PUT /users/:id/approve` to enable. (`internal/migrations/seed_ai_user.go`)
- **P1.T8** `.gitignore` extended to cover `.env.local` and `.env.*.local`. `.env.example` documented as template only. (`backend/.gitignore`, `backend/.env.example`)
- **P1.T9** `LoadConfig` now fails fast on boot if any critical secret is empty or matches a known default value (`my_secret_key`, `my_refresh_secret_key`, `minioadmin`, `minioadmin4321`, `secret`, `changeme`, `password`, `12345`). (`internal/config/config.go`)
- **P1.T10** After `AssignPermissions` / `UpdateRole`, all CSRF tokens for affected users are revoked. The user must acquire a new CSRF token on their next request. (`internal/auth/csrf.go`, `internal/repository/role_repository.go`, `internal/repository/role_repository_ext.go`)
- **P1.T11** Global search now uses Postgres `websearch_to_tsquery('english', ?)`, replacing the previous unsafe `to_tsquery` with string concatenation. User input containing `:`, `&`, etc. no longer breaks the query. (`internal/repository/search.go`)
- **P1.T12** Notification handlers (`ListNotifications`, `GetUnreadNotificationCount`, `MarkNotificationAsRead`, `MarkAllNotificationsAsRead`) reject requests when `user_id` is missing from the Gin context (returns 401). Ownership-checked routes cannot be silently bypassed. (`internal/handlers/notifications.go`)
- **P1.T13** Alert consumer fetches subscribed users via `JOIN roles ON roles.id = users.role_id WHERE roles.name IN (?)`. The previous `users.role IN (?)` query referenced a column removed by an earlier migration and was returning 0 rows — no alerts were being delivered. (`internal/consumers/alert_consumer.go`)

### Added

- **P1.T14** `SecurityHeaders()` Gin middleware sets `Strict-Transport-Security`, `X-Content-Type-Options`, `X-Frame-Options`, `Referrer-Policy`, `Content-Security-Policy`, `Permissions-Policy` on every response. (`internal/middleware/security_headers.go`)

### Removed

- **P1.T7** `backend/gen_hash.go` helper (printed bcrypt of `Admin@1234`) deleted. Use a programmatic `bcrypt.GenerateFromPassword` instead.

### Security notes

- The hard-coded AI agent password (literal `rZQ$4Rs!6{QHaR{5Sra{]z_%n`) previously committed to source is no longer valid. The AI service (`ai-service/backend_client.py`) now reads the password from the `AI_AGENT_PASSWORD` env var, which the operator must set to the value logged on first backend boot. The value is **intentionally left empty** in `.env.example`.
- Operators upgrading an existing deployment must set all required secrets (`JWT_SECRET`, `REFRESH_TOKEN_SECRET`, `MINIO_ACCESS_KEY_ID`, `MINIO_SECRET_ACCESS_KEY`, `SSLCOMMERZ_STORE_PASSWORD`, `RABBITMQ_URL`) to non-default values, otherwise the backend will refuse to start.
- The committed `backend/server` and `backend/cmd/server/server` binaries still contain the old AI agent password hash and should be rebuilt (`go build ./cmd/server`) before deployment.
