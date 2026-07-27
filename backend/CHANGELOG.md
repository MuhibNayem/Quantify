# Changelog

All notable changes to the Quantify backend are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).

## [Unreleased]

### Changed

- **Per-order tax amount** `Order` now persists `Subtotal` (pre-tax, post-promo-discount) and `TaxAmount` (the tax portion captured at checkout) alongside `TotalAmount` (post-tax, post-loyalty-discount). `applyTax` returns the post-tax total and the tax amount separately so the order row can store both. Downstream reporting no longer has to derive tax from `total / (1 + rate)`. The discount audit hook now uses the pre-tax subtotal instead of the post-tax total, which previously caused the recorded pre-discount subtotal to include tax. (`internal/domain/orders.go`, `internal/services/sales_service.go`)
- **Basket analysis response** `GET /reports/basket-analysis` now accepts `page` and `limit` query parameters and returns the pagination envelope `{items, total, page, limit}`. Default limit is 20. This replaces the legacy flat-array response and is a breaking change for existing consumers. (`internal/repository/reporting_repository.go`, `internal/services/reporting_service.go`, `internal/handlers/reports.go`)

### Breaking

- **WebSocket auth: `?token` query parameter removed (first-message auth only).** The `GET /ws` endpoint no longer accepts a JWT via the `?token=` query parameter. Any `?token=` value is **ignored entirely** (a `logrus.Warn` is emitted server-side so operators can spot un-updated clients). After the WebSocket upgrade, the client MUST send a first message with the shape `{"type":"auth","token":"<JWT>"}`; the server validates the JWT, resolves the user's role/permissions, and replies with `{"type":"auth_ok"}` once auth succeeds. If the first message is missing, malformed, has the wrong `type`, or carries an invalid token, the server closes the connection without registering the client. Additionally, if the client does not send a valid first-message auth within **5 seconds** of the upgrade, the connection is closed. **The frontend must be updated to send first-message auth before this change is deployed**, otherwise live notifications will silently stop working. (`internal/handlers/websocket.go`)

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

## [Phase 2] 2026-XX-XX — Data integrity

### Changed

- **P2.T1** Stock transfer now updates batch quantities inside a single transaction: FEFO deduction from the source batches with `FOR UPDATE` row lock, increment-or-create at the destination, plus the existing audit and transfer-record rows. The previous implementation wrote audit rows but never touched `batches`, so aggregate stock was unchanged after a transfer. (`internal/services/stock_transfer_service.go`, `internal/handlers/inventory.go`)
- **P2.T2** Refund computation now uses additive tax on the subtotal: `(unitPrice × qty) + (unitPrice × qty × taxRate)`, rounded to cents. Replaces the multiplicative `× (1 + taxRate/100)` formula. (`internal/handlers/return_handler.go`)
- **P2.T3** Loyalty point reversal on a return is now proportional to `(returnedQty / orderedQty)` rather than `int(refundAmount × earningRate)`. A `PointsRedeemed` of 50 on a 10-qty line reverses 5 points on a 1-qty return, not 5 × earningRate × refund. (`internal/handlers/return_handler.go`)
- **P2.T5** PO receive now inserts a `StockAdjustment(Type=STOCK_IN, ReasonCode=PURCHASE)` row per received item, inside the same transaction as the Batch creation. Previously only a `Batch` was written, leaving the audit trail incomplete. (`internal/services/purchase_order_service.go`)
- **P2.T6** `CreateFromSuggestion` re-fetches the product with `FOR UPDATE` inside the same transaction and uses the current `PurchasePrice` (not the suggestion's snapshot). When the price has drifted since the suggestion was generated, a `logrus.Warn` records the delta. `ReorderSuggestion.SnapshotPrice` is captured at suggestion creation for the audit trail. (`internal/services/replenishment_service.go`, `internal/services/purchase_order_service.go`, `internal/domain/models.go`)
- **P2.T7** Vendor returns now validate that every `BatchID` in the request belongs to the claimed `ProductID`. A mismatched BatchID returns an error before any state mutation. (`internal/services/vendor_return_service.go`)
- **P2.T10** Stock adjustment reads (`currentQuantity` sum and `batchesToUpdate`) now use `SELECT … FOR UPDATE` row locks inside the transaction. A concurrency test fires 50 simultaneous -5 adjustments against a single batch and asserts the final quantity is exactly 750. (`internal/handlers/stock.go`)
- **P2.T11** `UpdatePurchaseOrder` now diffs and reconciles items in place inside a transaction, preserving `ReceivedQuantity` for already-received lines. The previous implementation deleted all items and recreated them, silently losing received quantities. (`internal/services/purchase_order_service.go`)
- **P2.T12** Cron alert scan now wraps each product's processing in a `recover()` block. A panic for one product no longer blocks subsequent products. Failures are logged via `logrus.WithField("product_id", …).Errorf`. (`internal/handlers/alerts.go`)
- **P2.T13** `GetCustomer` now returns a uniform `404 {"error":"customer not found"}` for every missing identifier (id, email, phone, username). The previous behavior stopped at the first match and varied the response shape per lookup path, enabling account enumeration via phone-number probes. (`internal/handlers/crm.go`)
- **P2.T14** `GetChurnRisk` now accepts email, phone, or username in the `:identifier` path param, not just numeric IDs. (`internal/handlers/crm.go`)
- **P2.T15** `PUT /settings` now validates the value against a typed descriptor map (`tax_rate_percentage` must be 0–100, `currency` must be a 3-letter code, etc.) and rejects unknown setting keys. Previously any string was accepted, even for numeric settings. (`internal/handlers/settings.go`)
- **P2.T16** Dead first pass in `GetProductPerformanceAnalytics` removed (had a typo that would have made the function fail if it had been reached). The second pass is now correct — the missing `suppliers.id as supplier_id` column was restored to its `SELECT` and `GROUP BY` clauses. (`internal/repository/reporting_repository.go`)

### Added

- **P2.T4** Vendor return state machine: `Create` defaults to `PENDING` (no stock movement); `Approve` records `ApprovedBy`/`ApprovedAt`; `Complete` transitions PENDING → COMPLETED with stock deduction and `STOCK_OUT` audit; `Cancel` transitions PENDING → CANCELLED without stock changes. Invalid transitions return 409. New routes: `POST /replenishment/returns/:id/{approve,complete,cancel}`. (`internal/services/vendor_return_service.go`, `internal/handlers/replenishment.go`)
- **P2.T8** `ResolvePromotion()` extracted from `calculateItemDiscount` into `internal/services/promotion_resolver.go`. Pinned by 4 property tests: product scope beats subcategory beats category; ties broken by `Priority`, then `DiscountValue`. Empty list returns `nil`. (`internal/services/promotion_resolver.go`, `internal/services/sales_service.go`)
- **P2.T9** `CheckoutRequest.RegisterLocationID *uint` field added. When set, stock is deducted from that location's batches instead of the product's default location. FEFO allocation and `FOR UPDATE` locking now apply per-location. Falls back to the product's `LocationID` when unset. (`internal/requests/stock_request.go`, `internal/services/sales_service.go`)

### Migration notes

- The new `ReorderSuggestion.SnapshotPrice` column is added by `AutoMigrate` on next boot (Phase 4 will switch to versioned SQL migrations).
- New vendor-return state requires existing rows to have a valid status. If `PurchaseReturn.Status` is `COMPLETED` from before this change, the new transitions still work (COMPLETED is a valid terminal state).
