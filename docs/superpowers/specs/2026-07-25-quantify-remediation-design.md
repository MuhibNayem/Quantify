# Quantify Backend Remediation — Design Spec

- **Date**: 2026-07-25
- **Author**: opencode (from audit at `/Users/a.k.mmuhibullahnayem/.local/share/opencode/tool-output/tool_f9a11f308001rJbWdJn1ErDGas`)
- **Status**: Draft, awaiting user review
- **Scope**: `backend/` (Go 1.24.3, Gin, GORM, Postgres 15, Redis cluster, RabbitMQ, MinIO)

## Goals

Fix all correctness, security, integrity, architectural, operational, testing, and documentation issues surfaced by the backend audit. Produce a single reviewable design spec that drives 7 ordered phase-chapters, each shipped as its own branch + PR, with verification gates between phases.

## Non-goals

- New features (Stripe, Shopify, QuickBooks, RFID, 2FA/MFA, multi-tenancy).
- Frontend / client changes.
- Schema redesign beyond minimal additions required to fix correctness (e.g., adding `Order.TaxAmount`, not redesigning the order domain).
- Production deploy/rollback runbook (deferred to Phase 7 doc work, but not a deployment plan).

## Architecture context

The backend is a Go service (`inventory/backend`, Go 1.24.3) using:

- Gin HTTP framework + CORS middleware
- GORM ORM against PostgreSQL 15 (single instance; DSN `sslmode=disable`)
- Redis (cluster of 3) for cache, sessions, rate-limit-ish lists, RBAC cache
- RabbitMQ 3 for async work (bulk import, alerts, reporting)
- MinIO for object storage (reports, bulk files)
- WebSocket Hub for realtime updates
- Prometheus + Grafana for observability (Go runtime only; no app metrics today)
- `logrus` for structured logging (with `fmt.Print` leftovers in 8 spots)

Layered as **Handler → Service → Repository**, with `repository.DB` exposed as a package-level singleton, `internal/router/router.go` as a single 442-line setup function, and `bootstrap/app.go` orchestrating DB / Redis / RabbitMQ / Hub / cron / consumers at startup.

## Phase structure (high level)

| Phase | Title | Tasks | Type |
|---|---|---|---|
| 1 | Security hardening | ~14 | Hotfix |
| 2 | Data integrity | ~16 | Hotfix |
| 3 | Correctness fixes | ~13 | Hotfix |
| 4 | Architectural overhaul | ~18 | Refactor |
| 5 | Operations & observability | ~10 | Infra |
| 6 | Test coverage | ~8 | Quality |
| 7 | Documentation & API spec | ~6 | Docs |

**Total**: ~85 tasks. Phase 1 ships first as a hotfix PR; subsequent phases land in sequence with verification gates between them.

## Cross-cutting conventions

Applied to every task:

- **Citations**: each task lists `file:line` ranges so reviewers can jump directly.
- **TDD**: each fix lands with a regression test (unit or integration). Failing test first, then fix, then green.
- **No new secrets in source**: any new credential is generated, stored in DB, or pulled from env.
- **Migration files**: any schema change ships as a versioned SQL migration under `migrations/` (added in Phase 4).
- **CHANGELOG**: each phase appends an entry to `backend/CHANGELOG.md`.
- **Backwards compatibility**: where existing data is involved (e.g., AuditLog backfill, AI agent user), the migration path is documented in the task.

---

# Phase 1 — Security hardening

**Goal**: Eliminate the security issues that allow unauthorised access, data exfiltration, or service compromise.

**Files in scope**:
- `backend/internal/middleware/auth.go`, `csrf.go`, `permission.go`, `error_handler.go`
- `backend/internal/auth/jwt.go`, `csrf.go`
- `backend/internal/services/payment_service.go`
- `backend/internal/handlers/user.go`, `alerts.go`, `notifications.go`, `websocket.go`
- `backend/internal/consumers/alert_consumer.go`
- `backend/internal/repository/search.go`, `redis.go`
- `backend/internal/migrations/seed_ai_user.go`, `seed_roles.go`
- `backend/internal/config/config.go`
- `backend/gen_hash.go` (root)
- `backend/.env`, `backend/.env.example`, `backend/.gitignore`

## Tasks

### P1.T1 — Remove hard-coded AI agent password
- **Files**: `backend/internal/migrations/seed_ai_user.go:31`
- **Current**: bcrypt hash of literal password `rZQ$4Rs!6{QHaR{5Sra{]z_%n` is committed and seeded; user is `IsActive=true` with `Admin` role.
- **Target**: Generate a random password on first run, print it to operator logs once, store only the bcrypt hash. Mark user `IsActive=false` until operator approves.
- **Acceptance**: Grep repo for the literal hash returns no matches. New boot logs the generated password exactly once, then never again. User must be activated manually via `/users/:id/approve`.
- **Depends on**: none.

### P1.T2 — SSLCommerz IPN signature verification
- **Files**: `backend/internal/services/payment_service.go:289-315`
- **Current**: `ValidateSSLCommerzIPN` only checks `status[0] == "VALID"`. No HMAC verification.
- **Target**: Implement SSLCommerz IPN validation per their docs — recompute HMAC-SHA512 over sorted params using the store password; compare to provided signature. Reject on mismatch with 400. Add positive integration test with a known-good signed payload.
- **Acceptance**: Test `TestSSLCommerzIPN_RejectsBadSignature` passes; tampered payload returns 400.
- **Depends on**: none.

### P1.T3 — WebSocket origin allowlist
- **Files**: `backend/internal/handlers/websocket.go:19-22`; `backend/internal/config/config.go` (add `WSAllowedOrigins []string`)
- **Current**: `CheckOrigin` returns `true` for every origin.
- **Target**: Read allowlist from `WS_ALLOWED_ORIGINS` env (CSV), default to `CORS_ALLOWED_ORIGINS` value. Reject unmatched origins with 403.
- **Acceptance**: Test `TestWebSocket_RejectsUnknownOrigin` returns 403; allowed origin upgrades successfully.
- **Depends on**: none.

### P1.T4 — Refresh token TTL configurable, default 30 days
- **Files**: `backend/internal/handlers/user.go:217,304`; `backend/internal/config/config.go`
- **Current**: Refresh tokens stored in Redis with 10-year TTL.
- **Target**: Add `REFRESH_TOKEN_TTL` env, default `720h` (30 days). On successful refresh, slide the TTL forward if more than 50% consumed. Add test that confirms the new TTL is applied.
- **Acceptance**: `TestRefreshTokenTTL` asserts Redis `TTL` is within expected range after issuance.
- **Depends on**: none.

### P1.T5 — Permission cache invalidation on role/permission mutations
- **Files**: `backend/internal/repository/role_repository.go:59-138`; `backend/internal/repository/role_repository_ext.go:12-42`; `backend/internal/middleware/permission.go:13-55`
- **Current**: Cache `rbac:role:<name>:permissions` has 1-hour TTL; invalidated only by `AssignPermissions`, `UpdateRole`, `DeleteRole`. If those code paths miss, stale permissions persist for up to an hour.
- **Target**: Reduce TTL to 60 seconds and add explicit `InvalidateRoleCache(roleName)` called from every mutation path. Add integration test that asserts cache is cleared immediately after a permission grant.
- **Acceptance**: `TestPermissionCache_InvalidatedOnUpdate` green; latency unchanged on hot path.
- **Depends on**: none.

### P1.T6 — Deactivate AI agent user by default
- **Files**: `backend/internal/migrations/seed_ai_user.go`
- **Current**: Seeded user is `IsActive=true`.
- **Target**: Seed with `IsActive=false`. Operator must call `/users/:id/approve` to activate.
- **Acceptance**: Test boots fresh DB, asserts user row has `IsActive=false`.
- **Depends on**: P1.T1.

### P1.T7 — Remove `gen_hash.go` helper from source
- **Files**: `backend/gen_hash.go` (entire file)
- **Current**: Standalone `package main` that prints bcrypt of `Admin@1234`.
- **Target**: Delete the file. Document a CLI replacement in `docs/runbook.md` (Phase 7).
- **Acceptance**: `git grep "Admin@1234"` returns no matches.
- **Depends on**: none.

### P1.T8 — Scrub `.env` from repo history
- **Files**: `backend/.env`, `backend/.gitignore`
- **Current**: `backend/.env` is committed with real-looking secrets (`password`, `my_secret_key`, etc.).
- **Target**: Remove `.env` from working tree, add to `.gitignore`, run `git filter-repo` (or BFG) to strip from history. Add `backend/.env.example` as the canonical reference. Rotate any committed secret values.
- **Acceptance**: `git log --all -- backend/.env` returns nothing. `git status` shows `.env` ignored.
- **Depends on**: none.

### P1.T9 — Reject default-value secrets on boot
- **Files**: `backend/internal/config/config.go:145-174`
- **Current**: Fails fast on empty or default values like `my_secret_key` for JWT/MinIO. `minioadmin` is rejected, but `minioadmin4321` (the docker-compose default) is allowed.
- **Target**: Add an explicit list of disallowed secret strings (the docker-compose defaults included). Operators must set unique values.
- **Acceptance**: Test boots with disallowed secret and panics with clear error.
- **Depends on**: none.

### P1.T10 — CSRF token rotation on privilege change
- **Files**: `backend/internal/auth/csrf.go`; `backend/internal/handlers/user.go`
- **Current**: CSRF token generated at login + refresh, never rotated.
- **Target**: Rotate CSRF token on role/permission changes for that user. Add endpoint or internal call when role is updated.
- **Acceptance**: `TestCSRFRotation_OnRoleChange` green.
- **Depends on**: P1.T5.

### P1.T11 — Sanitize global search query
- **Files**: `backend/internal/repository/search.go:75-107`; `backend/internal/services/search_service.go`
- **Current**: `to_tsquery('english', q)` with words joined by `&`; no escaping of `:` or `&` in user input.
- **Target**: Use `websearch_to_tsquery('english', ?)` which handles user-input safely. Remove the manual `&` join.
- **Acceptance**: Test sends `q=foo:bar&baz` and returns results instead of SQL error.
- **Depends on**: none.

### P1.T12 — Notification ownership enforcement
- **Files**: `backend/internal/handlers/notifications.go:36-75`
- **Current**: Reads `:id` from URL; `c.GetUint("user_id")` returns 0 if missing (e.g., middleware was bypassed). Check `authUserID != uint(userID)` succeeds when both are 0, allowing access.
- **Target**: Reject when `authUserID == 0` with 401. Audit all handlers that read `user_id` from context to enforce the same rule.
- **Acceptance**: Test that simulates missing context returns 401.
- **Depends on**: none.

### P1.T13 — Fix alert consumer role query
- **Files**: `backend/internal/consumers/alert_consumer.go:124`
- **Current**: Queries `users WHERE role IN (?)` — but `users.role` was dropped by `fixRolePermissionsSchema` migration.
- **Target**: Query `users JOIN roles ON roles.id = users.role_id WHERE roles.name IN (?)`.
- **Acceptance**: `TestAlertConsumer_DeliversToSubscribedRoles` shows alerts delivered.
- **Depends on**: none.

### P1.T14 — Security headers middleware
- **Files**: `backend/internal/middleware/` (new file `security_headers.go`)
- **Current**: No HSTS, X-Content-Type-Options, X-Frame-Options, CSP, Referrer-Policy.
- **Target**: Add `SecurityHeadersMiddleware` that sets: `Strict-Transport-Security` (1 year, includeSubDomains), `X-Content-Type-Options: nosniff`, `X-Frame-Options: DENY`, `Referrer-Policy: strict-origin-when-cross-origin`, `Content-Security-Policy: default-src 'self'` (override-able).
- **Acceptance**: Test asserts headers present on `/health` response.
- **Depends on**: none.

## Phase 1 verification gate

- `gosec ./...` clean (no high-severity findings).
- All P1 unit + integration tests green.
- Manual security checklist signed off (changeset reviewed for unintended secret leakage).
- CHANGELOG entry appended.

---

# Phase 2 — Data integrity

**Goal**: Make every business operation atomic, accurate, and recoverable. Eliminate silent corruption.

**Files in scope**:
- `backend/internal/handlers/inventory.go`, `return_handler.go`, `crm.go`, `settings.go`, `alerts.go`
- `backend/internal/services/sales_service.go`, `purchase_order_service.go`, `vendor_return_service.go`, `replenishment_service.go`
- `backend/internal/repository/replenishment.go`, `product.go`, `category.go`
- `backend/internal/domain/models.go` (add fields where required)

## Tasks

### P2.T1 — Stock transfer updates batches
- **Files**: `backend/internal/handlers/inventory.go:28-105`; `backend/internal/services/stock_transfer_service.go` (new)
- **Current**: Stock transfer writes `StockTransfer` record + two `StockAdjustment` rows (one `STOCK_OUT`, one `STOCK_IN`) but never touches the `batches` table.
- **Target**: Open transaction. Decrement source batches FEFO (matching the checkout logic), increment destination batches (existing or new). Audit row created for the transfer itself. Verify source has enough stock before commit.
- **Acceptance**: Test `TestStockTransfer_UpdatesBatches` confirms source batch quantities decrease, destination batch quantities increase, total stock unchanged.
- **Depends on**: none.

### P2.T2 — Return refund math fix (additive tax)
- **Files**: `backend/internal/handlers/return_handler.go:55-216`
- **Current**: `RefundAmount = UnitPrice * Qty * (1 + tax_rate/100)` — tax applied multiplicatively (compounded), not added on top.
- **Target**: `RefundAmount = (UnitPrice * Qty) + ((UnitPrice * Qty) * tax_rate/100)`. Add test asserting the new formula matches expected values for several sample inputs.
- **Acceptance**: `TestReturnRefund_AdditiveTax` green.
- **Depends on**: none.

### P2.T3 — Loyalty point reversal prorates by OrderItem.PointsRedeemed
- **Files**: `backend/internal/handlers/return_handler.go:382-399`
- **Current**: Reverses `int(RefundAmount * earning_rate)` regardless of how the order was paid.
- **Target**: Compute reversal as `int(OrderItem.PointsRedeemed * (ReturnedQty / OrderItem.Quantity))`. Refuse negative balances. Add test for proportional reversal.
- **Acceptance**: `TestLoyaltyReversal_Proportional` green.
- **Depends on**: none.

### P2.T4 — Vendor return status workflow
- **Files**: `backend/internal/services/vendor_return_service.go:46`
- **Current**: Status hard-coded to `COMPLETED` at creation.
- **Target**: Default `PENDING`. Add explicit `Approve` and `Complete` service methods with state-machine guards. Handler routes for transitions. Schema already supports the enum.
- **Acceptance**: Test walks through `PENDING → APPROVED → COMPLETED`; invalid transitions return 409.
- **Depends on**: none.

### P2.T5 — PO receive writes StockAdjustment
- **Files**: `backend/internal/services/purchase_order_service.go:190-265`
- **Current**: Creates `Batch` per item but no `StockAdjustment` audit row.
- **Target**: Insert `StockAdjustment(Type=STOCK_IN, ReasonCode=PURCHASE, ProductID, BatchID, PreviousQuantity, NewQuantity, AdjustedBy, AdjustedAt)` per received item.
- **Acceptance**: Test `TestPOReceive_CreatesStockAdjustment` asserts row inserted.
- **Depends on**: none.

### P2.T6 — Product.PurchasePrice freshness on PO
- **Files**: `backend/internal/services/purchase_order_service.go:67`
- **Current**: Reads `Suggestion.Product.PurchasePrice` at create-PO time; if the product price has changed since the suggestion, the PO uses a stale price silently.
- **Target**: Inside the create-PO transaction, re-fetch the product with `FOR UPDATE`, use its current `PurchasePrice`. If different from the suggestion snapshot, log a warning and persist the new price into the PO line.
- **Acceptance**: Test updates product price after suggestion, verifies PO uses latest price.
- **Depends on**: none.

### P2.T7 — Vendor return BatchID-ProductID validation
- **Files**: `backend/internal/services/vendor_return_service.go`
- **Current**: BatchID is required but not validated against the product.
- **Target**: Reject if `Batch.ProductID != Item.ProductID`. Reject if batch belongs to a different supplier's PO when `StrictSupplier` setting is enabled.
- **Acceptance**: Test asserts mismatched BatchID returns 400.
- **Depends on**: none.

### P2.T8 — Promotion priority test coverage
- **Files**: `backend/internal/services/sales_service.go:202` (calculateItemDiscount); `backend/internal/services/promotion_resolver_test.go` (new)
- **Current**: Logic exists but no test pinning the rule.
- **Target**: Property-based test asserting product > subcategory > category with `Priority` tiebreak.
- **Acceptance**: `TestPromotionResolution_Hierarchy` covers overlapping promotions.
- **Depends on**: none.

### P2.T9 — Checkout uses register location
- **Files**: `backend/internal/handlers/sales.go:43-84`; `backend/internal/services/sales_service.go`; `backend/internal/requests/checkout_request.go`
- **Current**: Deducts stock from `product.LocationID` regardless of where the register is.
- **Target**: Accept optional `locationId` in checkout request. If provided, deduct from that location's batches. If absent, fall back to `product.LocationID`. Add validation that the location exists and is active.
- **Acceptance**: Test with explicit location asserts deduction from correct batches.
- **Depends on**: P2.T1 (so transfer logic also uses this convention).

### P2.T10 — Stock adjustment transactional lock verification
- **Files**: `backend/internal/handlers/stock.go:146-286`
- **Current**: Reads `currentQuantity` inside the transaction; uses `tx.Model(...).Select(...).Row().Scan(&currentQuantity)`. Believed correct but not tested.
- **Target**: Add a concurrency test that fires N concurrent adjustments on the same product and asserts no lost updates.
- **Acceptance**: `TestStockAdjustment_NoLostUpdates` green with `-race` flag.
- **Depends on**: none.

### P2.T11 — UpdatePurchaseOrder preserves ReceivedQuantity
- **Files**: `backend/internal/services/purchase_order_service.go:174-185`
- **Current**: Deletes all items and recreates them; loses `ReceivedQuantity` for already-received lines.
- **Target**: Diff existing items vs new items; update in place; only delete items removed from the request. Block updates if PO is not `DRAFT`.
- **Acceptance**: Test updates a DRAFT PO that had partial receive simulated and confirms `ReceivedQuantity` preserved.
- **Depends on**: none.

### P2.T12 — Cron alert scan error tolerance
- **Files**: `backend/internal/handlers/alerts.go:253-263`; `bootstrap/app.go:100-110`
- **Current**: Synchronous; one product's failure blocks subsequent products and the cron tick.
- **Target**: Wrap each product iteration in a recover-and-log block; emit structured `logrus.WithError(err).Error()` per failure; do not stop iteration. P5.T3 will later add a Prometheus counter; here we only ensure the cron keeps ticking.
- **Acceptance**: Test injects a panic for one product and asserts scan completes for others.
- **Depends on**: none.

### P2.T13 — CRM customer identifier privacy
- **Files**: `backend/internal/handlers/crm.go:68-114`
- **Current**: Tries ID, email, phone, username in order, stops at first hit. A 404 vs a 200 reveals whether the user exists.
- **Target**: Return the same response shape (404 with generic message) for all misses. Optionally, hash identifiers and look up by hash to avoid leaking which field matched.
- **Acceptance**: Test asserts identical response timing and shape for missing email/phone/username/ID.
- **Depends on**: none.

### P2.T14 — CRM churn-risk route accepts non-numeric identifiers
- **Files**: `backend/internal/handlers/crm.go:249`; `backend/internal/router/router.go` (route declaration)
- **Current**: Handler parses identifier as `uint`; email/username → 400.
- **Target**: Normalize identifier: try ID, then email, then phone, then username. Pass the resolved user ID to the AI call.
- **Acceptance**: Test with email/username/phone identifier returns churn risk JSON.
- **Depends on**: P2.T13 (consistent resolver).

### P2.T15 — Settings type validation
- **Files**: `backend/internal/handlers/settings.go:28-44`
- **Current**: Persists raw stringified value, bypasses type validation enforced by seeded rows.
- **Target**: Define a `SettingDescriptor` map per key (type, allowed values/range, validation function). Reject PUT with mismatched value. Add admin-only PUT with audit log entry.
- **Acceptance**: Test attempts to set `tax_rate_percentage` to "abc" and asserts 400.
- **Depends on**: none.

### P2.T16 — Remove dead code in `GetProductPerformanceAnalytics`
- **Files**: `backend/internal/repository/reporting_repository.go:882-919`
- **Current**: First `for rows.Next()` block has a typo (`suppliers.id AS supplier_id` scanned into `&p.SupplierName`); would fail if executed. The function never returns from this block.
- **Target**: Delete the dead block; keep the working second pass. Add a comment explaining the report's contract.
- **Acceptance**: Test loads product performance report successfully without touching dead code.
- **Depends on**: none.

## Phase 2 verification gate

- Data-integrity integration tests green (P6.T2-T4 land later but a smoke subset here).
- `go test -race ./...` green.
- CHANGELOG entry.

---

# Phase 3 — Correctness fixes

**Goal**: Fix bugs that don't fit into security or integrity categories — UX defects, broken flows, missing features called out by the audit.

**Files in scope**:
- `backend/internal/consumers/bulk_consumer.go`, `reporting_consumer.go`
- `backend/internal/services/reporting_service.go`
- `backend/internal/repository/reporting_repository.go`, `forecasting.go`
- `backend/internal/handlers/sales.go`, `reports.go`, `notifications.go`
- `backend/internal/domain/models.go` (add `Order.TaxAmount`)
- `backend/internal/handlers/websocket.go`

## Tasks

### P3.T1 — Stop bulk import auto-finalize
- **Files**: `backend/internal/consumers/bulk_consumer.go:213-221`
- **Current**: Consumer calls `finalizeImport` immediately after persisting the validation result, breaking the documented preview/confirm UX.
- **Target**: Set job status to `PENDING_CONFIRMATION` and stop. Finalize only when `/bulk/products/import/:jobId/confirm` is called.
- **Acceptance**: Test confirms job sits at `PENDING_CONFIRMATION` until explicit confirm.
- **Depends on**: none.

### P3.T2 — Re-confirm idempotency
- **Files**: `backend/internal/consumers/bulk_consumer.go:223-321`; `backend/internal/handlers/bulk.go:147-174`
- **Current**: Re-confirming a job re-runs `processImportConfirmDelivery` and re-inserts already-created products → unique-constraint errors on `SKU`.
- **Target**: Guard against re-running finalize for a job already in `COMPLETED`. Make `CreateInBatches` use `ON CONFLICT (sku) DO NOTHING` for idempotent inserts.
- **Acceptance**: Test confirms a second confirm call is a no-op.
- **Depends on**: P3.T1.

### P3.T3 — Timezone SQL injection
- **Files**: `backend/internal/repository/reporting_repository.go:453-465`
- **Current**: Builds raw SQL with `fmt.Sprintf` interpolating the timezone from `SystemSetting("timezone")`. SQL-injection vector.
- **Target**: Use parameterized query against `system_settings`; pull the value into Go and bind as a parameter. Alternative: validate timezone against a known list (`pg_timezone_names`).
- **Acceptance**: Test with malicious timezone string returns safe error or empty result.
- **Depends on**: none.

### P3.T4 — AuditLog writers
- **Files**: `backend/internal/handlers/sales.go` (checkout with discount), `backend/internal/repository/audit_log.go` (new)
- **Current**: Nothing writes `AuditLog` rows; `GetVoidDiscountAuditReport` is always empty.
- **Target**: Emit audit row on checkout discount > X%, on void, on permission grant/revoke, on setting change. Backfill from existing data where recoverable (look for orphan discounts in orders).
- **Acceptance**: Test asserts audit row written on checkout with >10% discount.
- **Depends on**: P2.T15 (settings types).

### P3.T5 — Basket analysis pagination
- **Files**: `backend/internal/repository/reporting_repository.go:742`
- **Current**: `LIMIT 20` with no offset.
- **Target**: Add `page` and `limit` query params (default 1, 20; max 100). Response includes `total` count.
- **Acceptance**: Test paginates through 50 baskets.
- **Depends on**: none.

### P3.T6 — Customer insights net of returns
- **Files**: `backend/internal/repository/reporting_repository.go:791-809`
- **Current**: Filters `o.status='COMPLETED'` but ignores returns/refunds.
- **Target**: Subtract refunded amounts (`LEFT JOIN returns + return_items`). Include `NetRevenue` in addition to gross.
- **Acceptance**: Test with a customer who has $50k sales and $40k refunds asserts `NetRevenue ≈ $10k`.
- **Depends on**: none.

### P3.T7 — Per-order TaxAmount field
- **Files**: `backend/internal/domain/models.go` (Order struct); `backend/internal/services/sales_service.go`; `backend/internal/migrations/` (migration SQL)
- **Current**: Tax is computed at report time as `total / (1 + rate)`. Per-order rate isn't preserved.
- **Target**: Add `Order.TaxAmount decimal(10,2)` column. Populate at checkout. Migration adds the column nullable, then a backfill computes `TaxAmount = total - (total / (1 + rate))` for existing rows.
- **Acceptance**: Migration runs cleanly on existing DB; new orders write TaxAmount.
- **Depends on**: none.

### P3.T8 — Sales trends weekly groupBy
- **Files**: `backend/internal/repository/reporting_repository.go:33-101`; `backend/internal/handlers/reports.go`
- **Current**: `groupBy=weekly` maps to `date_trunc('week', ...)` which is ISO week.
- **Target**: Document the behavior; allow `groupBy=iso_week` (default) and `groupBy=calendar_week` (Sunday-start). Use `date_trunc('week', ...)` for ISO; compute calendar week in Go from local-tz timestamp.
- **Acceptance**: Test asserts both groupBy modes produce distinct buckets for cross-week data.
- **Depends on**: none.

### P3.T9 — `ListAllOrders` pagination
- **Files**: `backend/internal/handlers/sales.go:259-335`
- **Current**: Reads entire `orders` table into memory.
- **Target**: Add `page`, `limit` params; default `limit=50`, max `200`. Return `total` count.
- **Acceptance**: Test paginates through 200 orders.
- **Depends on**: none.

### P3.T10 — `User.AfterFind` safer preloading
- **Files**: `backend/internal/domain/models.go:171-176`; `backend/internal/handlers/user.go` (preload everywhere)
- **Current**: AfterFind sets `RoleName` only if `Role.ID != 0`. If a caller forgets to preload, response shows empty `RoleName`.
- **Target**: Change AfterFind to set `RoleName = ""` explicitly if Role is zero. Add lint/test that catches missing preloads via grep on handler functions.
- **Acceptance**: Test asserts `RoleName == ""` when Role not preloaded.
- **Depends on**: none.

### P3.T11 — WebSocket auth via first message
- **Files**: `backend/internal/handlers/websocket.go`; `backend/internal/websocket/client.go`
- **Current**: JWT in query string is logged by reverse proxies.
- **Target**: Accept a first message from the client containing `{type: "auth", token: "..."}`. Validate, then proceed. Reject upgrade requests without subsequent auth message within 5s. Optionally accept cookie too.
- **Acceptance**: Test confirms token in URL fails or never reaches logs; first-message auth succeeds.
- **Depends on**: P1.T3 (origin allowlist).

### P3.T12 — User-facing notification subscriptions
- **Files**: `backend/internal/handlers/notifications.go`; `backend/internal/repository/notification_repository.go`
- **Current**: `AlertRoleSubscription` only manageable by admins. No per-user opt-out.
- **Target**: Add `/users/:id/notifications/subscriptions` endpoints (GET/PUT) so a user can opt in/out per type. Keep admin-only role-wide subs as override.
- **Acceptance**: Test opts out and asserts no alert delivered.
- **Depends on**: none.

### P3.T13 — GetCustomer identifier leak fix (UX layer)
- **Files**: `backend/internal/handlers/crm.go:68-114`
- **Current**: 404 reveals user existence (covered in P2.T13 from a security angle). Here: UX correctness.
- **Target**: Return `200 OK` with a "not found" payload for missing identifier (consistent with the privacy fix in P2.T13). Document this as the official behavior.
- **Acceptance**: Same as P2.T13.
- **Depends on**: P2.T13.

## Phase 3 verification gate

- All P3 unit + integration tests green.
- OpenAPI spec regenerated; 100% match against router.
- CHANGELOG entry.

---

# Phase 4 — Architectural overhaul

**Goal**: Replace structural debt that blocks future work: monolithic router, package-level DB singleton, in-code migrations, oversized files, missing interfaces.

**Files in scope** (most touched):
- `backend/internal/router/router.go` → split into `internal/router/groups/*.go`
- `backend/internal/repository/database.go` (introduce interfaces, remove singleton)
- New: `backend/internal/app/container.go`
- New: `backend/migrations/` (golang-migrate SQL files)
- `backend/internal/repository/reporting_repository.go` → split
- `backend/internal/services/sales_service.go` → split
- `backend/internal/handlers/return_handler.go` → split
- `backend/internal/message_broker/rabbitmq.go` → extract topology

## Tasks

### P4.T1 — Router split into per-domain groups
- **Files**: `backend/internal/router/router.go:23-442`
- **Target**: New structure:
  ```
  internal/router/
    router.go           # bootstrap, Engine construction, common middleware
    groups/
      public.go         # /health, /metrics, /swagger, /webhooks
      payment.go        # /payment/*
      websocket.go      # /ws
      auth.go           # /api/v1/users/register, /login, /refresh-token
      users.go          # /api/v1/users
      products.go
      sales.go
      reports.go
      ... etc
  ```
  Each group is a `func(g *gin.RouterGroup, c *app.Container)` registering its routes.
- **Acceptance**: `router.go` < 100 lines; each group file < 300 lines; total LOC roughly unchanged.

### P4.T2 — DI container
- **Files**: new `backend/internal/app/container.go`; modified `router/router.go`
- **Target**: `app.Container` holds repos, services, handlers. Built once in `bootstrap/app.go`. `SetupRouter(container)` consumes it. No global state.
- **Acceptance**: No file outside `bootstrap/app.go` constructs a repository/service.

### P4.T3 — Repository interface extraction
- **Files**: new `backend/internal/repository/interfaces.go`; modify every consumer
- **Target**: Define interfaces for each repository (e.g., `ProductRepository`, `OrderRepository`). Services depend on interfaces. Bind concrete types in `app.Container`.
- **Acceptance**: `go test` mocks can replace any repository in any service test.

### P4.T4 — Remove package-level `repository.DB` singleton
- **Files**: `backend/internal/repository/database.go:17`; every consumer that reads `repository.DB`
- **Target**: Replace `repository.DB` access with `repository.WithTx(ctx, func(tx) error { ... })` or constructor injection. The package exposes helper functions; underlying `*gorm.DB` is private.
- **Acceptance**: Grep `repository.DB` returns only the package's own definition.

### P4.T5 — Adopt golang-migrate
- **Files**: new `backend/migrations/` directory; new `cmd/migrator/main.go` replacement
- **Target**: Versioned up/down SQL files: `0001_init.up.sql`, `0001_init.down.sql`, etc. Convert current schema generation into migration files. Add `cmd/migrator` binary that runs them.
- **Acceptance**: Fresh DB + `migrator up` produces same schema as today's AutoMigrate.

### P4.T6 — Convert in-code seeding to migrations
- **Files**: `backend/internal/repository/database.go:151-368`; `backend/migrations/0002_seed.up.sql`
- **Target**: Extract permission list, role-permission grants, system settings, default alert subscriptions into a migration file (or a post-migrate SQL seed). Idempotent (`ON CONFLICT DO NOTHING`).
- **Acceptance**: Fresh DB boots with seeded roles + settings; running migrations twice is a no-op.
- **Depends on**: P4.T5 (migration tool must exist first).

### P4.T7 — Disable AutoMigrate outside dev
- **Files**: `backend/internal/bootstrap/app.go:48`; `backend/internal/repository/database.go`
- **Target**: Only run AutoMigrate when `APP_ENV=dev`. Production relies on `cmd/migrator`.
- **Acceptance**: Boot with `APP_ENV=production` and an unmigrated DB returns clear error.
- **Depends on**: P4.T5, P4.T6 (migrations must be the supported path).

### P4.T8 — Reporting repository split
- **Files**: `backend/internal/repository/reporting_repository.go` (1089 LOC)
- **Target**: Split into:
  - `sales_reports.go` — sales trends, hourly heatmap, basket, returns analysis
  - `inventory_reports.go` — turnover, dead stock, aging, GMROI
  - `financial_reports.go` — profit margin, supplier performance, tax, cash reconciliation, audit
  - `crm_reports.go` — customer insights, product performance, shrinkage
  - Each < 400 LOC, each with focused test file.
- **Acceptance**: `wc -l internal/repository/*_reports.go` confirms split.

### P4.T9 — Sales service split
- **Files**: `backend/internal/services/sales_service.go:27-363`
- **Target**: Extract into:
  - `pricing.go` — line-level pricing, tax, discounts
  - `stock_allocator.go` — FEFO batch selection, locking
  - `loyalty.go` — earn/redeem/tiers
  - `promotion_resolver.go` — promotion matching
  - `checkout.go` — orchestrator that composes the above
- **Acceptance**: Each file < 300 LOC, dedicated tests.

### P4.T10 — Return handler split
- **Files**: `backend/internal/handlers/return_handler.go:230-454`
- **Target**: `request_return.go` (creation flow), `process_return.go` (approve/reject flow).
- **Acceptance**: Each handler file < 250 LOC.

### P4.T11 — OpenAPI regeneration
- **Files**: every handler file (add `// @Summary`, `// @Param`, `// @Success`, `// @Failure` annotations); `backend/api/docs.go`, `swagger.json`, `swagger.yaml`, `openapi.yaml`
- **Target**: Annotate every route; run `swag init` (or equivalent) at build time; commit generated files. The current 47-path spec grows to 160.
- **Acceptance**: `/swagger/index.html` renders all 160 paths.

### P4.T12 — Real webhook integration factory
- **Files**: `backend/internal/integrations/integrations.go`; new `backend/internal/integrations/shopify.go`, `quickbooks.go` (interface stubs)
- **Target**: Convert to a real registry. Add Shopify + QuickBooks no-op stubs that document the integration points. Document how to add a new integration in `docs/architecture.md`.
- **Acceptance**: `/webhooks` returns 200 for registered providers, 404 for unknown.

### P4.T13 — Centralized config via Viper/koanf
- **Files**: `backend/internal/config/config.go:1-175`; new dependencies
- **Target**: Replace hand-rolled `getEnv` with `koanf` (lighter) or `viper`. Support file-based config + env overrides + defaults. Add config struct validation.
- **Acceptance**: `config.Load()` returns typed config struct; all `getEnv` calls removed.

### P4.T14 — Centralized error wrapping
- **Files**: 10 handlers that emit `gin.H{"error": "..."}` directly
- **Target**: Replace with `c.Error(appErr)` or `panic(appErr)` caught by `ErrorHandler` middleware. Audit every handler.
- **Acceptance**: Grep for `gin.H{"error"` returns zero matches.

### P4.T15 — RabbitMQ topology centralization
- **Files**: `backend/internal/message_broker/rabbitmq.go:20-319`
- **Target**: Move all queue/exchange/routing-key declarations into `internal/message_broker/topology.go`. Constants for every routing key. Subscribers iterate the topology spec.
- **Acceptance**: Adding a new queue requires only editing topology.go.

### P4.T16 — Consumer cancel-func wiring
- **Files**: `backend/internal/bootstrap/app.go:171-205`
- **Target**: `Consumer.Start(ctx)` returns a `CancelFunc`; `app.Stop()` calls all cancel funcs before closing connections. Add wait group so consumers finish in-flight work.
- **Acceptance**: `TestGracefulShutdown_CancelsConsumers` asserts all cancel funcs called within 5s.

### P4.T17 — Env-driven CORS
- **Files**: `backend/internal/router/router.go:26-34`
- **Target**: Read `CORS_ALLOWED_ORIGINS` env (CSV), default to `http://localhost:5173` for dev. Multiple origins supported.
- **Acceptance**: Boot with custom origins and assert CORS headers.

### P4.T18 — Loyalty tier thresholds via settings
- **Files**: `backend/internal/services/sales_service.go:257`; `backend/internal/handlers/settings.go`
- **Target**: Replace hard-coded `500/2500/10000` thresholds with `loyalty.tier_silver`, `loyalty.tier_gold`, `loyalty.tier_platinum` settings. Defaults preserved.
- **Acceptance**: Test changes tier threshold via PUT /settings and asserts tier promotion occurs at new threshold.

## Phase 4 verification gate

- `go test ./...` green.
- Grep for `repository.DB` returns zero non-package references.
- Grep for `gin.H{"error"` returns zero matches.
- OpenAPI renders all 160 paths.
- CHANGELOG entry.

---

# Phase 5 — Operations & observability

**Goal**: Make the service operable in production. Health, metrics, logs, traces, CI.

**Files in scope**:
- `backend/internal/handlers/health.go` (split into `health.go`, `ready.go`)
- `backend/internal/middleware/` (new `request_id.go`)
- `backend/cmd/server/main.go`
- `backend/Dockerfile`
- New: `.github/workflows/ci.yml`
- `grafana/dashboards/` (new)
- `backend/internal/handlers/*.go`, `services/*.go`, `repository/*.go` (8 fmt.Print replacements)

## Tasks

### P5.T1 — Add `/ready` endpoint
- **Files**: new `backend/internal/handlers/ready.go`
- **Target**: `GET /ready` pings DB (`db.Ping`), Redis (`PING`), RabbitMQ (channel open check), MinIO (bucket exists). Returns 200 only if all four succeed; 503 otherwise with details.
- **Acceptance**: Test with one dependency down returns 503.

### P5.T2 — Liveness vs readiness split
- **Files**: `backend/internal/router/router.go:100`
- **Target**: `/health` (existing, unconditional `UP`) is liveness; `/ready` (new) is readiness. Document in `docs/runbook.md` (Phase 7).
- **Acceptance**: Kubernetes probes configured; liveness never fails unless process is broken.

### P5.T3 — Custom Prometheus collectors
- **Files**: new `backend/internal/metrics/` package; modified handlers/services
- **Target**: Add counters and histograms:
  - `http_requests_total{method,path,status}`
  - `http_request_duration_seconds{method,path}`
  - `cache_hits_total{kind}` / `cache_misses_total{kind}`
  - `jobs_processed_total{type,status}`
  - `payments_processed_total{method,status}`
  - `alerts_fired_total{type}` / `alerts_delivered_total{type}`
- **Acceptance**: `/metrics` exposes new collectors; Grafana queries return non-zero after a synthetic load test.

### P5.T4 — Grafana dashboards
- **Files**: new `grafana/dashboards/{overview,api,jobs,payments,alerts}.json`
- **Target**: JSON dashboards provisioned via `grafana/provisioning/dashboards/dashboards.yml`. Cover: request rate/latency, error rate, job queue depth, payment success ratio, alert volume by type.
- **Acceptance**: Dashboards render in Grafana with sample data.

### P5.T5 — Replace `fmt.Print` with `logrus`
- **Files**: known list — `handlers/notifications.go:46-48`, `handlers/websocket.go:56`, `services/purchase_order_service.go:110`, `repository/reporting_repository.go:745,787`, plus grep finds.
- **Target**: Replace every `fmt.Print/Printf/Println` with `logrus.WithFields(...).Info/Error/etc`.
- **Acceptance**: Grep for `fmt.Print` returns zero matches in non-test files.

### P5.T6 — Request ID middleware
- **Files**: new `backend/internal/middleware/request_id.go`; modified `cmd/server/main.go` (logrus hook)
- **Target**: Generate UUID per request (header `X-Request-ID` if present). Inject into Gin context. Add logrus hook that pulls the ID and includes it in every log line within that request scope.
- **Acceptance**: Test sends request with `X-Request-ID: foo`; logs contain `request_id=foo`.

### P5.T7 — OpenTelemetry tracing
- **Files**: new `backend/internal/telemetry/` package; modified middleware/services/repos/message_broker
- **Target**: Gin → service → repo spans. Propagate via AMQP headers (`traceparent`) and Redis (W3C trace context). Exporter: OTLP (configurable endpoint).
- **Acceptance**: Synthetic request trace shows full span tree in Jaeger/Tempo.

### P5.T8 — CI workflow
- **Files**: new `.github/workflows/ci.yml`
- **Target**: On push/PR: `go vet`, `go test ./...`, `golangci-lint run`, `gosec ./...`, build Docker image. Matrix: Go 1.24.
- **Acceptance**: PR check required; CI blocks merges on failure.

### P5.T9 — Container hardening
- **Files**: `backend/Dockerfile`
- **Target**: Add non-root user, `HEALTHCHECK --interval=30s --timeout=5s CMD wget -qO- http://localhost:8080/health || exit 1`, pin base image digests, multi-stage cleanup of build cache, drop capabilities.
- **Acceptance**: `docker inspect` shows `User: "nonroot"`; `docker run --cap-drop=ALL` works.

### P5.T10 — Log shipping example config
- **Files**: new `docs/log-shipping.md`
- **Target**: Provide Vector + Loki and Fluent Bit configurations for stdout JSON shipping. No code changes.
- **Acceptance**: Doc reviewed; sample config tested against local Loki.

## Phase 5 verification gate

- `/metrics` exposes new collectors.
- Grafana dashboards render.
- CI workflow runs green on a synthetic PR.
- CHANGELOG entry.

---

# Phase 6 — Test coverage

**Goal**: Establish reliable automated tests covering the riskiest paths and regressions from earlier phases.

**Files in scope**:
- New: `backend/test/integration/` (Postgres testcontainers)
- Modified: every handler/service touched in Phases 1-3

## Tasks

### P6.T1 — Test database via testcontainers
- **Files**: new `backend/test/integration/db.go`
- **Target**: Helper that spins up a Postgres testcontainer, runs migrations, returns `*gorm.DB`. Used by every integration test in the phase.
- **Acceptance**: Helper used by ≥ 3 test files; total test runtime < 5 minutes.

### P6.T2 — Checkout → return → refund integration test
- **Files**: rewrite `backend/internal/handlers/return_test.go:50-169` (current SQLite version)
- **Target**: Postgres-aware test: create product + batch + user, run checkout, request return, approve return, assert stock restored + refund transaction + loyalty reversal.
- **Acceptance**: Test green; covers all of P2.T1, P2.T2, P2.T3.

### P6.T3 — Bulk import end-to-end test
- **Files**: new `backend/internal/consumers/bulk_consumer_test.go`
- **Target**: Spin up MinIO container + Postgres, upload CSV, run consumer, assert products created and consumer respects PENDING_CONFIRMATION (P3.T1).
- **Acceptance**: Test green; covers P3.T1, P3.T2.

### P6.T4 — PO receive integration test
- **Files**: new `backend/internal/services/purchase_order_service_test.go`
- **Target**: Create PO, approve, send, receive partial, assert StockAdjustment created (P2.T5), product price freshness (P2.T6), UpdatePurchaseOrder preserves received qty (P2.T11).
- **Acceptance**: Test green.

### P6.T5 — Alert consumer test
- **Files**: new `backend/internal/consumers/alert_consumer_test.go`
- **Target**: Seed roles + alert subscriptions + product with low stock; trigger alert; assert delivery to subscribed users via the fixed query (P1.T13).
- **Acceptance**: Test green; demonstrates the audit fix.

### P6.T6 — Permission cache invalidation test
- **Files**: new `backend/internal/repository/role_repository_ext_test.go`
- **Target**: Assert cache is populated, mutation clears cache (P1.T5), next read shows fresh permissions.
- **Acceptance**: Test green.

### P6.T7 — Webhook integration factory test
- **Files**: new `backend/internal/integrations/integrations_test.go`
- **Target**: Register a fake integration; call `/webhooks`; assert it receives the payload. Assert unknown provider returns 404.
- **Acceptance**: Test green.

### P6.T8 — Property-based tests
- **Files**: new `backend/internal/services/promotion_resolver_test.go`, `refund_math_test.go`, `tax_calc_test.go`
- **Target**: Use `rapid` or `gopter` to assert properties (promotion hierarchy, refund positivity, tax round-trip).
- **Acceptance**: Property tests run for 1 minute each without counterexample.

## Phase 6 verification gate

- `go test ./...` green; coverage on `services/` and `handlers/` > 70%.
- No `-race` failures.
- CHANGELOG entry.

---

# Phase 7 — Documentation & API spec

**Goal**: Make the project understandable and operable by humans.

**Files in scope**:
- `backend/api/openapi.yaml`
- `docs/business_logic.md`
- New: `docs/architecture.md`, `docs/security.md`, `docs/runbook.md`
- `README.md`, `docs/README.md`

## Tasks

### P7.T1 — Regenerate `openapi.yaml` to 160 routes
- **Files**: `backend/api/openapi.yaml`
- **Target**: After P4.T11, the generated file lists every route. Add request/response examples for non-trivial endpoints (checkout, return, bulk import, churn).
- **Acceptance**: OpenAPI renders in Swagger UI; every route documented.

### P7.T2 — Update `docs/business_logic.md`
- **Files**: `docs/business_logic.md`
- **Target**: Remove multi-tenancy claims until implemented. Document actual RBAC model, return/refund flow, loyalty tiers (settings-driven after P4.T18), bulk import lifecycle.
- **Acceptance**: Doc accurately describes what the code does today.

### P7.T3 — Add `docs/architecture.md`
- **Files**: new `docs/architecture.md`
- **Target**: ADRs (router split, DI, migrations tool choice), layer diagram, deployment topology, request lifecycle diagram.
- **Acceptance**: Reviewed and merged.

### P7.T4 — Add `docs/security.md`
- **Files**: new `docs/security.md`
- **Target**: Threat model (STRIDE-lite), controls per threat, key rotation runbook, secret handling, vulnerability disclosure process.
- **Acceptance**: Reviewed by security-aware reviewer.

### P7.T5 — Add `docs/runbook.md`
- **Files**: new `docs/runbook.md`
- **Target**: Common incidents (DB down, Redis down, RabbitMQ down, MinIO down, alert backlog), recovery steps, on-call rotation template, deploy/rollback procedure.
- **Acceptance**: Reviewed.

### P7.T6 — Update README files
- **Files**: `README.md`, `docs/README.md`
- **Target**: Accurate env var table (matches `config.go`), ops quickstart (docker-compose up + run migrator), dev quickstart (run server + test client), link to spec docs.
- **Acceptance**: New developer can onboard using README alone.

## Phase 7 verification gate

- All docs reviewed.
- `README.md` env var table matches code.
- OpenAPI spec 100% match.
- Final CHANGELOG entry: "Remediation plan v1 complete".

---

# Risks & open questions

- **Production data**: AI agent user has hard-coded password in any environment that already booted. After P1.T1+T6, operators must reset password + delete `.env` artifacts. Document migration steps in `docs/runbook.md`.
- **Backwards compatibility**: schema additions (P3.T7) require migration + backfill; deployed instances must run before the next release.
- **Bulk import breaking change**: P3.T1 changes observable behavior (jobs no longer auto-complete). Frontend expectations may need updating; coordinate with client team.
- **Migration tool adoption (P4.T5-P4.T7)**: any external automation that relied on AutoMigrate running at boot must use `cmd/migrator` instead. Document in `docs/runbook.md`.
- **OpenTelemetry overhead (P5.T7)**: traces add ~5-15% latency. Validate in staging before production rollout.
- **CI resource cost (P5.T8)**: integration tests with testcontainers may exceed free-tier runner minutes. Consider self-hosted runners.
- **Refactor scope (Phase 4)**: large refactor touching every handler/service. Recommend doing Phase 4 in a long-lived branch with frequent merges back to main to avoid drift.

# Acceptance criteria

The plan is complete when:

1. All 85 tasks closed with passing tests and reviewer sign-off.
2. `go test ./...` green; coverage on services + handlers > 70%.
3. `golangci-lint` and `gosec` clean.
4. OpenAPI spec renders all 160 routes.
5. Grafana dashboards provisioned and live.
6. CI pipeline runs green on every PR.
7. All four documents (architecture, security, runbook, business_logic) reviewed.
8. CHANGELOG covers every phase.
9. No `repository.DB` direct access outside the package; no `fmt.Print`; no `gin.H{"error": ...}`.
10. Production-ready security headers; SSLCommerz IPN signed; permission cache correctly invalidated; alert consumer delivering alerts.