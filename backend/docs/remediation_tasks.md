# Remediation Tasks

This document captures the concrete tasks and sub-tasks required to close all currently identified gaps. Use this as a checklist when resuming work in a new session.

---

## 1. RabbitMQ Subscription Nil Pointer Crash
**Files:** `internal/message_broker/rabbitmq.go`, `internal/consumers/bulk_consumer.go`

- [x] Audit every `message_broker.Subscribe` call and note whether a WebSocket hub is really needed.
- [x] Update the subscriber implementation to skip broadcasting (or to no-op) when `hub == nil`.
- [ ] Add unit/integration coverage for bulk import/export/report queues to confirm handlers still process messages when no hub is provided.

## 2. User Notification Settings Creates Dummy Accounts
**File:** `internal/handlers/alerts.go`

- [x] Remove the code path that creates placeholder users with plaintext passwords.
- [x] Replace it with a hard failure (`404`) when the requested user does not exist.
- [x] Ensure existing users’ identities are pulled from the DB and no passwords are manipulated in this handler.
- [ ] Add tests to cover valid updates and missing-user behavior. (optional - no need now)

## 3. Transaction Gateway ID Constraint Breaks Payments
**Files:** `internal/services/payment_service.go`, `internal/domain/models.go`, `internal/repository/payment.go`

- [x] Decide on an approach: generate a unique provisional `GatewayTransactionID` (UUID) at insert time.
- [x] Update the model/repository/service accordingly so every transaction row satisfies uniqueness.
- [x] Extend payment flows (cash, bKash, SSLCommerz) to ensure the gateway ID column is populated with the correct identifier once the gateway responds.
- [ ] Backfill or migrate existing rows if necessary.
- [ ] Add regression tests that attempt multiple payments sequentially to ensure no constraint violations occur. (optional - no need now)

## 4. Reporting Cache TTLs Must Be Configurable and Varied
**Files:** `internal/services/reporting_service.go`, `internal/config/config.go`

- [x] Enhance `Config` to include TTL values for each reporting endpoint (e.g., sales trends, inventory turnover, profit margin).
- [x] Update `ReportingService` to use `time.Duration` values (not integer seconds) and to respect the endpoint-specific TTL.
- [x] Provide sensible defaults (e.g., sales trends cache = 1h, inventory turnover = 6h) while allowing overrides via env vars.
- [ ] Add tests to assert TTL values are applied correctly.

## 5. Stock Adjustments Must Record the Authenticated User
**Files:** `internal/handlers/stock.go`, `internal/middleware/auth.go`

- [x] Update the stock adjustment handler to retrieve `user_id` from the Gin context (set by `AuthMiddleware`) and fail if it is missing.
- [x] Store that ID in `StockAdjustment.AdjustedBy` instead of the hard-coded `1`.
- [x] Optionally extend the response payloads/reporting to include the human-readable user.
- [ ] Add test coverage to verify the handler rejects requests without an authenticated user and correctly stamps the ID otherwise.

## 6. MinIO Must Support HTTP and HTTPS Endpoints Securely
**Files:** `internal/storage/minio.go`, `internal/config/config.go`

- [x] Add configuration that distinguishes between HTTP and HTTPS MinIO endpoints (e.g., `MINIO_USE_TLS` boolean or infer from endpoint scheme).
- [x] Update `NewMinIOUploader` to honor the TLS setting; default to TLS unless explicitly disabled to satisfy compliance requirements.
- [x] Document the new env vars in the README or deployment docs.
- [ ] Consider rotating existing keys if they may have traversed plaintext connections.

## 7. Distinct Cache Lifetimes for Reporting Endpoints
**Files:** `internal/services/reporting_service.go`, `internal/config/config.go`

- [x] While working on Task 4, ensure each report type can specify its own TTL (possibly reuse the same configuration fields).
- [x] Expose helper methods (e.g., `getCacheTTL(reportType string)`) so adding future reports is straightforward.
- [x] Document the recommended TTL values for each report type.

---

## 8. Harden Authorization Across Endpoints (High Priority)
**Files:** `internal/router/router.go`, `internal/middleware/auth.go`, `internal/handlers/*`

- [x] Catalogue every `/api/v1` route and decide the minimal role required (Admin, Manager, Staff). Document the mapping (`docs/api_authorization_matrix.md`).
- [x] Apply role-specific middleware to each route group (e.g., Admin-only for user management, Manager-or-above for reporting/bulk ops, Staff for day-to-day inventory).
- [x] Add ownership checks inside handlers where users access resources by ID, username, phone or emails (users, CRM customers, alerts) to prevent IDOR.
- [x] Verify that previously public but sensitive endpoints are now protected and update API docs accordingly (`docs/api_documentation.md`, `docs/api_authorization_matrix.md`).
- [ ] Add regression tests or manual verification ensuring unauthorized roles receive `403`.

## 9. Standardize Error Handling & Logging
**Files:** `internal/handlers/*`, `internal/services/*`

- [x] Replace ad-hoc `gin.H` error responses with consistent `AppError` usage.
- [x] Ensure services bubble up structured errors instead of swallowing them; log or return JSON marshal/cache failures.
- [x] Use structured logrus fields for every error log, capturing request context and identifiers.

## 10. Input Validation & CSRF Strategy
**Files:** `internal/requests/*`, `internal/handlers/*`

- [x] Audit all request structs and add validation tags/enums.
- [x] Validate query params for endpoints without structs (alerts filters, reporting params, etc.).
- [x] Decide and document CSRF mitigation (header-based token). Middleware enforces `X-CSRF-Token`, and login/refresh responses return the token.

## 11. Testing & Observability Foundations
**Files:** repo-wide

- [ ] Add foundational unit tests (auth middleware, product CRUD, stock adjustments) plus integration tests for role enforcement.
- [ ] Extend logging to structured patterns and introduce custom Prometheus metrics (alerts triggered, jobs started, payments processed).
- [ ] Plan distributed tracing (OpenTelemetry) and instrument high-value flows.

## 12. Graceful Shutdown & Consumer Scaling
**Files:** `cmd/server/main.go`, `internal/message_broker/rabbitmq.go`, `internal/consumers/*`

- [ ] Add OS signal handling for graceful HTTP/cron/RabbitMQ shutdown.
- [ ] Refactor RabbitMQ manager to avoid race conditions and support context cancellation.
- [ ] Allow multiple consumer instances/concurrency configuration for bulk/report workers.

## 13. Database Access Discipline
**Files:** `internal/handlers/*`, `internal/repository/*`

- [ ] Refactor handlers to rely on repository interfaces instead of `repository.DB`.
- [ ] Wrap multi-step operations (user flows, alerts, purchase orders) in transactions.
- [ ] Profile endpoints for N+1 queries and optimize with `Preload`/joins.

---

**Next Session Bootstrapping Checklist**
1. Re-read this document and open the referenced files.
2. Decide implementation order (e.g., fix crashing bug first, then security issues).
3. For each task, create a short branch-level TODO list and tick items off here as you progress.
