# Backend Remediation Plan (Code-Derived)

This plan is built only from the current backend codebase (handlers, services, middleware, repositories, config). It replaces prior analyses and focuses on actionable, production-readiness work.

## 1) Authentication & Authorization
- **1.1 Harden secrets & token policies**
  - Require strong JWT/refresh secrets via env validation; fail fast on defaults.
  - Shorten refresh token TTL; rotate on login/refresh; store jti for revocation.
  - Add device/session identifiers to token storage for targeted logout.
- **1.2 Access token storage logic**
  - Ensure Redis availability and error handling around token checks; add fallback/metrics.
  - Enforce token audience/issuer; consider key rotation support.
- **1.3 CSRF enforcement**
  - Validate CSRF token binding to user/refresh token; add rotation endpoint; log failures.
- **1.4 RBAC performance & cache**
  - Cache role permissions in Redis with invalidation on role update; add local in-memory fallback.
- **1.5 Ownership/IDOR checks**
  - Audit all user/CRM/alerts routes for ownership validation (only admins bypass).
  - Add helper to assert `user_id` context matches path params where applicable.
- **1.6 Password & login policies**
  - Enforce password complexity and lockout/backoff on repeated failures.
  - Add optional MFA flag for high-privilege roles.

## 2) Messaging & Background Workers
- **2.1 RabbitMQ robustness**
  - Refactor manager to remove race conditions; guard channel reuse; handle backoff/jitter.
  - Add context-aware shutdown for consumers; drain/ack inflight messages on SIGTERM.
  - Health checks and metrics for connection/channel state.
- **2.2 Consumer concurrency & idempotency**
  - Parameterize worker counts per queue; add idempotency keys for bulk/report jobs.
  - Define dead-letter queues and retry strategy with max attempts and poison-message parking.
- **2.3 Event contracts**
  - Document payload schemas for stock/product/alert/report events; validate before publish.

## 3) HTTP Server, Middleware, and Routing
- **3.1 Graceful shutdown**
  - Ensure HTTP, cron, and MQ consumers honor context cancellation and timeouts.
- **3.2 Error handling consistency**
  - Standardize AppError responses (status/code/message); avoid raw errors to clients.
  - Add request-scoped logging with correlation IDs.
- **3.3 CORS & security headers**
  - Make origins configurable; add security headers (HSTS, CSP skeleton, frame-opts, referrer policy).

## 4) Data Layer & Transactions
- **4.1 Repository discipline**
  - Remove handler use of global `repository.DB`; inject interfaces for testability.
  - Add interfaces for transactional units of work.
- **4.2 Transaction coverage**
  - Wrap multi-step flows: stock adjustments/transfers, PO approve/send/receive, CRM updates, user updates (role changes), payments.
- **4.3 Query performance**
  - Audit N+1s on list endpoints; add selective preloads and pagination defaults/limits.
  - Index review: ensure filters (status, foreign keys, dates) have supporting indexes.
- **4.4 Data integrity**
  - Enforce enum constraints (status/type) via validation and DB check constraints where possible.

## 5) Stock, Inventory, and Replenishment
- **5.1 Stock adjustment fidelity**
  - Implement FEFO/FIFO configurable depletion; handle partial batch and zero-quantity cleanup.
  - Validate quantities > 0; prevent stock-out below zero; record audit trail with actor and reason.
- **5.2 Stock transfer atomicity**
  - Ensure source/dest adjustments occur in one transaction with invariant checks.
- **5.3 Alert correctness**
  - Debounce duplicate alerts per product/batch; add resolution timestamps and actor; avoid re-trigger loops.

## 6) Reporting & Caching
- **6.1 Cache strategy**
  - Centralize TTL selection; add cache versioning and invalidation on related data changes.
  - Add negative-cache guardrails to avoid stale errors.
- **6.2 Report generation**
  - Validate job payloads; enforce allowed report types; add progress/error metadata.
  - Stream large exports; set size limits; redact PII in exports where not needed.

## 7) Payments
- **7.1 Gateway compliance**
  - Hash/verify SSLCommerz IPN signatures; log and reject invalid callbacks.
  - Securely store gateway credentials; avoid logging sensitive fields.
- **7.2 Transaction integrity**
  - Enforce unique OrderID+method constraints; make GatewayTransactionID non-empty, update on callback.
  - Add idempotent callback processing and status transitions with audit fields.
- **7.3 Error handling**
  - Timeouts/retries with backoff for gateway calls; surface clear errors to clients.

## 8) Notifications & WebSockets
- **8.1 Delivery semantics**
  - Persist notification send status and errors; retry policies for email.
  - Ensure WebSocket hub handles nil cases; add auth and per-user routing validation.
- **8.2 Preferences**
  - Enforce that notification settings are only set for existing users; validate emails/phones.

## 9) Configuration & Secrets
- **9.1 Env validation**
  - Fail fast on missing critical envs (DB, Redis, RabbitMQ, MinIO, JWT, SMTP, payment keys).
  - Type-safe parsing with clear defaults only for dev.
- **9.2 Profiles**
  - Provide dev/prod config templates; disable auto-migrate in prod; add migration CLI.

## 10) Observability & Operations
- **10.1 Metrics**
  - Custom Prometheus counters/histograms: requests (by route/status), MQ events, jobs, payments, alerts, cache hits/misses.
- **10.2 Logging**
  - Structured logging everywhere with request IDs, user IDs, job IDs, and correlation keys.
- **10.3 Tracing**
  - Introduce OpenTelemetry tracing across HTTP, DB, RabbitMQ, Redis.
- **10.4 Health & readiness**
  - Add /healthz and /ready probes checking DB, Redis, RabbitMQ connectivity.

## 11) Validation & API Contracts
- **11.1 Request validation**
  - Strengthen validation tags; add enums for types/status; validate query params with defaults.
- **11.2 Response shape**
  - Standardize pagination and error envelopes; remove debug prints (e.g., LoginUser).
- **11.3 Swagger/OpenAPI**
  - Regenerate specs from code annotations; ensure auth headers/CSRF documented.

## 12) Testing & Quality Gates
- **12.1 Unit tests**
  - Auth middleware, permission middleware with cached/uncached paths, CSRF validation.
  - Stock adjustments/transfers (happy path, insufficient stock, audit stamping).
  - Payment flows (success/fail/ipn) with mocked HTTP clients.
- **12.2 Integration tests**
  - API slices: products CRUD, roles/permissions enforcement, alerts, reporting job enqueue.
  - RabbitMQ consumer flow with in-memory or test broker.
- **12.3 Tooling**
  - Add CI pipeline with lint (go vet, staticcheck), tests, race detector for critical paths.

## 13) Frontend API Alignment (backend-impact)
- **13.1 Auth flows**
  - Ensure refresh/CSRF/token shapes match frontend expectations; document headers and cookie strategy if adopted.
- **13.2 Error surfaces**
  - Consistent error fields (`error`, `message`, `code`) to allow client to render meaningful toasts.

## 14) Deployment & Data Safety
- **14.1 Migrations**
  - Add migration runner; remove auto-migrate in prod; version schema changes (indexes, constraints, enums).
- **14.2 Backups & retention**
  - Define DB backup strategy; Redis persistence settings; MinIO bucket policies.
- **14.3 Rate limiting & abuse protection**
  - Add per-IP/user rate limits on auth and sensitive endpoints; captcha hook for login if needed.

## Suggested Execution Order (High-Level)
1) Stabilize platform: graceful shutdown, RabbitMQ fixes, auth/CSRF/RBAC caching, env validation.  
2) Data correctness: transactions around stock/PO/payment flows; validation enums; ownership checks.  
3) Security/compliance: password/MFA options, payment IPN verification, secret handling, rate limits.  
4) Observability & tests: metrics/logging/tracing; unit/integration tests; CI gate.  
5) Performance & UX: cache strategy, query tuning, API error/response consistency.  
6) Deployment hygiene: migration tooling, prod configs, backups.
