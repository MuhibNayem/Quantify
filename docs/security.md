# Security

This document describes the Quantify backend's threat model, the controls that
mitigate each threat, and operational procedures for secrets and key rotation.
For how these controls are wired into the request path, see
[architecture.md §5](./architecture.md).

> Most of the controls below were introduced in the Phase 1 security hardening.
> The authoritative record is `backend/CHANGELOG.md` (Phase 1 section).

---

## 1. Threat model (STRIDE-lite)

The system is a single-tenant inventory/POS backend exposing a JSON API, a
WebSocket channel, payment-gateway callbacks, and webhook integrations. The
table maps each STRIDE category to the most relevant threats and the controls
in place.

| # | STRIDE | Threat | Control(s) | Reference |
| :-- | :----- | :----- | :--------- | :-------- |
| T1 | **Spoofing** | Forged/stolen bearer token used to impersonate a user. | JWT signature validation **plus** a Redis whitelist keyed by JTI (`access_token:<jti>`); logged-out tokens are rejected even before expiry. | `middleware/auth.go` |
| T2 | Spoofing | Weak/guessed secrets allow token forgery. | Boot fails fast if `JWT_SECRET` / `REFRESH_TOKEN_SECRET` / MinIO keys are empty or match known defaults. | `config/config.go` `validateCriticalConfig` |
| T3 | Spoofing | Cross-Site WebSocket hijacking. | WS requires an `Origin` header matching `WS_ALLOWED_ORIGINS`; empty origins rejected. | `handlers/websocket.go` |
| T4 | Spoofing | Forged payment gateway callback. | SSLCommerz IPN verifies the gateway **MD5 signature** using `SSLCOMMERZ_STORE_PASSWORD` before persisting any transaction. | `services/payment_service.go` |
| T5 | **Tampering** | CSRF: forged state-changing request from a victim's browser. | Per-user CSRF token required on `POST/PUT/PATCH/DELETE` (`X-CSRF-Token`); validated against the authenticated user. | `middleware/csrf.go` |
| T6 | Tampering | Privilege change leaves stale CSRF tokens valid. | All CSRF tokens for affected users are revoked on `AssignPermissions` / `UpdateRole`. | `auth/csrf.go`, `repository/role_repository*.go` |
| T7 | Tampering | SQL/TSQuery injection via search input. | Global search uses `websearch_to_tsquery('english', ?)` with bound parameters; report timezones are validated (`[A-Za-z0-9/_+-]`) and bound, not interpolated. | `repository/search.go`, `repository/reporting_repository.go` |
| T8 | Tampering | Race conditions corrupt stock/financial state. | Checkout, transfers, adjustments, and returns run in transactions with `SELECT … FOR UPDATE` row locks (FEFO batch allocation). | `services/checkout.go`, `stock_allocator.go`, `handlers/stock.go` |
| T9 | **Repudiation** | User denies a discount, settings change, or permission grant. | Best-effort `AuditLog` rows for discounts ≥ 10%, settings updates (`{old,new}`), and role permission grants; surfaced via `/reports/audit/voids`. | `repository/audit_log.go` |
| T10 | **Information Disclosure** | Account enumeration via CRM lookups. | `GetCustomer` returns a uniform `404 {"error":"customer not found"}` for every missing identifier (id/email/phone/username). | `handlers/crm.go` |
| T11 | Information Disclosure | Notification endpoints leak other users' data. | Notification handlers reject requests with no `user_id` in context (401); ownership-checked routes cannot be bypassed. | `handlers/notifications.go` |
| T12 | Information Disclosure | Committed secrets / default credentials. | `.gitignore` covers `.env`, `.env.local`, `.env.*.local`; `.env.example` is a template only; no hard-coded password hashes in source. | `backend/.gitignore`, P1.T1/T7/T8 |
| T13 | Information Disclosure | Clickjacking / MIME sniffing / data exfil via browser. | `SecurityHeaders()` sets HSTS, `X-Content-Type-Options: nosniff`, `X-Frame-Options: DENY`, `Referrer-Policy`, CSP `default-src 'self'`, `Permissions-Policy`. | `middleware/security_headers.go` |
| T14 | **Denial of Service** | Unbounded report/list queries exhaust memory. | Pagination on heavy endpoints (e.g. `ListAllOrders`, basket analysis); report responses cached in Redis with configurable TTLs. | `handlers/sales.go`, `repository/reporting_repository.go` |
| T15 | DoS | A panicking product blocks the alert sweep. | Cron alert scan wraps each product in `recover()`; failures are logged per product. | `handlers/alerts.go` |
| T16 | DoS | Open WebSocket to arbitrary origins floods the hub. | Origin allowlist (T3) + first-message auth within 5s or the connection is closed. | `handlers/websocket.go` |
| T17 | **Elevation of Privilege** | User accesses routes beyond their role. | `RequirePermission` per route; permissions cached 60s and invalidated on mutation. Admin bypass is explicit and intentional. | `middleware/permission.go` |
| T18 | Elevation of Privilege | Long-lived refresh tokens widen the attack window. | Refresh token TTL configurable via `REFRESH_TOKEN_TTL` (default `720h` = 30 days); boot fails on unparsable values. Replaces a prior 10-year TTL. | `handlers/user.go`, `config/config.go` |
| T19 | Elevation of Privilege | AI agent account enabled by default. | AI agent user is seeded with `IsActive=false` and a random password printed once; operator must `PUT /users/:id/approve`. | `migrations/seed_ai_user.go` |
| T20 | Elevation of Privilege | Stale permissions served from cache after a demotion. | Permission cache TTL is 60s (down from 1h) and `InvalidateRoleCache` runs on every role mutation. | `repository/role_repository*.go` |

---

## 2. Authentication & authorization model

- **Transport**: assume TLS termination in front of the backend; HSTS is set
  unconditionally. CORS is restricted to `CORS_ALLOWED_ORIGINS` with
  credentials enabled.
- **Access tokens**: short-lived JWTs. Validity requires both a valid signature
  and presence in the Redis whitelist. Logout removes the token from Redis.
- **Refresh tokens**: separate secret (`REFRESH_TOKEN_SECRET`), configurable
  TTL, rotated on the refresh endpoint.
- **CSRF tokens**: issued alongside access/refresh tokens at login and refresh;
  required on unsafe methods; rotated on refresh; invalidated on logout and on
  role/permission changes.
- **Authorization**: permission-based RBAC with an explicit Admin bypass. See
  [business_logic.md §1](./business_logic.md) and
  [access_control.md](./access_control.md).
- **WebSocket auth**: after upgrade, the client must send
  `{"type":"auth","token":"<JWT>"}` as the first message within 5 seconds. The
  legacy `?token=` query parameter is ignored (a warning is logged so operators
  can spot un-updated clients).

---

## 3. Secret handling

### 3.1 Storage and injection

- Secrets are provided via **environment variables** (or an optional TOML file),
  never committed. Precedence: defaults → TOML → env (`config.LoadConfig`).
- `.env` and variants are git-ignored. `backend/.env.example` documents keys
  with placeholder values only.
- In Docker Compose the backend reads `backend/.env` via `env_file`.

### 3.2 Required secrets

The backend **refuses to start** if any of these is missing or set to a known
default (`my_secret_key`, `my_refresh_secret_key`, `minioadmin`,
`minioadmin4321`, `secret`, `changeme`, `password`, `12345`, empty):

| Variable | Purpose |
| :--- | :--- |
| `JWT_SECRET` | Access-token signing. |
| `REFRESH_TOKEN_SECRET` | Refresh-token signing. |
| `MINIO_ACCESS_KEY_ID` / `MINIO_SECRET_ACCESS_KEY` | Object storage credentials. |
| `SSLCOMMERZ_STORE_PASSWORD` | Gateway auth + IPN signature verification (required, validated). |
| `RABBITMQ_URL` | Broker connection (required). |

### 3.3 AI agent password

On first boot the backend generates a **random** password for the AI agent user
and prints it once via the logs. The operator must capture it and set
`AI_AGENT_PASSWORD` for the `ai-service`. It is intentionally left empty in
`.env.example`; do not use a placeholder.

### 3.4 General rules

- Generate secrets with a CSPRNG (e.g. `openssl rand -base64 48`).
- Never log secrets. The AI agent password is the single deliberate exception
  and is printed only at seed time.
- Rebuild binaries (`go build ./cmd/server`) before deploy; do not ship
  committed binaries, which may embed stale hashes.

---

## 4. Key rotation runbook

Rotating `JWT_SECRET` / `REFRESH_TOKEN_SECRET` invalidates **all** existing
tokens (signatures no longer verify). Plan for a forced re-login.

### 4.1 Prepare

1. Generate new secrets:
   ```bash
   openssl rand -base64 48   # JWT_SECRET
   openssl rand -base64 48   # REFRESH_TOKEN_SECRET
   ```
2. Update the secret store / `backend/.env` (or orchestrator secrets). Do **not**
   commit them.

### 4.2 Rotate (rolling)

1. Apply the new env to the backend deployment.
2. Restart/redeploy the backend. On boot it re-initializes JWT signing with the
   new secrets.
3. All existing access and refresh tokens are now invalid; clients receive 401
   and must log in again. CSRF tokens are likewise invalid and are re-issued at
   login.
4. Verify: `GET /ready` returns healthy; a fresh login succeeds; an old token
   returns 401.

### 4.3 Zero-downtime (optional, dual-secret)

If a forced re-login is unacceptable, support a brief dual-key window:

1. Deploy a build that verifies against `[new, old]` secrets but signs with
   `new`. (Not currently built in — requires a code change to `auth.InitializeJWT`.)
2. After the access-token TTL elapses, deploy the single-`new`-secret config.

> The current implementation supports a single secret pair, so the default
> procedure is the rolling rotation in §4.2.

### 4.4 MinIO / gateway credentials

- **MinIO**: rotate the access key in MinIO, update `MINIO_ACCESS_KEY_ID` /
  `MINIO_SECRET_ACCESS_KEY`, redeploy. Confirm `/ready` (which pings MinIO) is
  healthy.
- **SSLCommerz / bKash**: rotate via the provider console, update the env vars,
  redeploy. For SSLCommerz, the same `SSLCOMMERZ_STORE_PASSWORD` is used for
  outgoing auth and IPN verification, so both update together.

### 4.5 Post-rotation checklist

- [ ] Old tokens return 401.
- [ ] Fresh login + refresh works.
- [ ] `/ready` healthy (DB, Redis, AMQP, MinIO).
- [ ] AI service re-authenticates (if `AI_AGENT_PASSWORD` changed).
- [ ] No secrets appear in logs or VCS history.

---

## 5. CI security gates

`/.github/workflows/ci.yml` runs on every PR:

- `go vet ./...`
- `go test -race ./...`
- `golangci-lint`
- `gosec ./...` (static security analysis)
- Docker build smoke test

---

## 6. Related documents

- [architecture.md](./architecture.md)
- [business_logic.md](./business_logic.md)
- [runbook.md](./runbook.md)
- `backend/CHANGELOG.md` (Phase 1 security hardening)
