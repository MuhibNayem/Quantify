# Business Logic

This document describes the core business rules implemented in the Quantify
backend. It is the source of truth for *what* the system does; see
[architecture.md](./architecture.md) for *how* the code is organized and
[security.md](./security.md) for the threat model.

> Scope note: Quantify is a single-tenant inventory management and POS system.
> There is no multi-tenancy and no RFID integration. Barcode support is limited
> to 1D barcode lookup/generation via the `/barcode` endpoints.

---

## 1. Role-Based Access Control (RBAC)

Access control is permission-based, not hard-coded to role names. A user has
exactly one `Role`; a `Role` holds many `Permission`s through a many-to-many
join (`role_permissions`).

### 1.1 Roles

Four roles are seeded as **system roles** (`is_system = true`). System roles
cannot be deleted and cannot be renamed (`internal/services/role_service.go`).

| Role | Description | Permission set |
| :--- | :--- | :--- |
| **Admin** | Full system access | All permissions. Also bypasses every `RequirePermission` check in middleware. |
| **Manager** | Store management | Product/category/supplier/location management, replenishment, CRM + loyalty, POS, orders, returns (request + manage), all reports, alerts, bulk import/export, user management, settings view, dashboard. |
| **Staff** | Standard employee | Read-only product/category/location/supplier, barcode, customer read, POS access, order read, return request, alert view, dashboard, notifications. |
| **Customer** | Customer portal | Seeded as a system role; permissions are assigned per deployment. Customers request returns on their own orders. |

Custom (non-system) roles can be created freely via `POST /roles` and assigned
any subset of permissions via `PUT /roles/:id/permissions`.

### 1.2 Enforcement

- `middleware.AuthMiddleware` validates the JWT and loads `user_id` + `role`
  into the Gin context. The access token must also exist in the Redis whitelist
  (keyed by JTI), so logged-out tokens are rejected even before expiry.
- `middleware.RequirePermission(roleRepo, "perm.name")` guards each route.
  - **Admin bypass**: the `Admin` role short-circuits the check and is always
    allowed (`internal/middleware/permission.go`).
  - Other roles have their permission set fetched via
    `roleRepo.GetPermissionsByRoleName`, cached in Redis with a **60s TTL** and
    invalidated on every role mutation.
- A few routes still use coarse role gates (`AuthorizeMiddleware("Admin",
  "Manager")`, `AdminOnly()`) — e.g. job cancellation and alert subscriptions.

### 1.3 Permission catalog

Seeded in `migrations/0002_seed.up.sql` (mirrored by
`internal/repository/database.go`). Grouped by domain:

| Group | Permissions |
| :--- | :--- |
| Product Management | `products.read`, `products.write`, `products.delete`, `categories.read`, `categories.write` |
| Inventory | `locations.read`, `locations.write`, `suppliers.read`, `suppliers.write`, `barcode.read`, `replenishment.read`, `replenishment.write`, `alerts.view`, `alerts.manage`, `inventory.view`, `inventory.read`, `inventory.write` |
| CRM | `customers.read`, `customers.write`, `loyalty.read`, `loyalty.write`, `crm.view`, `crm.read`, `crm.write` |
| POS | `pos.access`, `returns.request`, `returns.manage`, `pos.view` |
| Orders | `orders.read`, `orders.manage` |
| Reports | `reports.sales`, `reports.inventory`, `reports.financial`, `reports.view` |
| System | `bulk.import`, `bulk.export`, `dashboard.view`, `notifications.read`, `notifications.write` |
| Settings | `settings.view`, `settings.manage` |
| Access Control | `users.view`, `users.manage`, `roles.view`, `roles.manage` |

See [access_control.md](./access_control.md) for the mapping of permissions to
UI routes and sidebar items.

---

## 2. Checkout Flow

Entry point: `POST /api/v1/sales/checkout` (`pos.access`), implemented by
`SalesService.ProcessCheckout` (`internal/services/checkout.go`). The entire
operation runs inside a single database transaction; any error rolls back all
stock, order, loyalty, and audit writes.

### 2.1 Steps

1. **Lock products.** `fetchProducts` selects all requested products with
   `SELECT … FOR UPDATE` to serialize concurrent checkouts touching the same
   products.
2. **Lock batches (FEFO).** `fetchBatches` selects batches with
   `quantity > 0`, locked `FOR UPDATE`, ordered
   `expiry_date ASC NULLS LAST, created_at ASC`. This is **First-Expired,
   First-Out**: perishable stock expiring soonest is sold first; non-perishable
   (`NULL` expiry) is sold last, oldest batch first.
3. **Load active promotions.** Promotions where `is_active = true` and
   `start_date <= now <= end_date`, with product/category/sub-category
   associations preloaded.
4. **Per item — allocate stock.** `resolveDeductionLocation` picks the stock
   location: the request's `RegisterLocationID` if set, otherwise the product's
   default `LocationID`. When a register location is set, batches are re-fetched
   (locked) for that specific location. `allocateStock` verifies available
   quantity, deducts across batches in FEFO order, and writes a
   `StockAdjustment(Type=STOCK_OUT, ReasonCode=SALE)` audit row.
5. **Per item — price & promote.** `calculateItemDiscount` runs
   `ResolvePromotion` to pick the best promotion and adjusts the unit price.
6. **Tax.** `applyTax` reads the `tax_rate_percentage` system setting and
   computes `taxAmount = subtotal × rate` and `postTaxTotal = subtotal + tax`.
7. **Loyalty.** `handleLoyalty` redeems points (if a customer is attached and
   `pointsToRedeem > 0`), then accrues new points on the net amount and
   recomputes the tier.
8. **Persist.** `createOrderAndTransaction` writes the `Order` (with
   `Subtotal`, `TaxAmount`, `TotalAmount`), its `OrderItem`s, and a
   `Transaction` row, then emits a discount audit entry if significant.

### 2.2 Promotion resolution

`ResolvePromotion` (`internal/services/promotion_resolver.go`) picks the single
most specific applicable promotion per product:

- **Specificity hierarchy**: product-scoped (rank 3) beats sub-category (rank 2)
  beats category (rank 1).
- **Tie-break**: higher `Priority` wins; if still tied, higher `DiscountValue`.
- Discount types: `PERCENTAGE` (off the unit price) and `FIXED_AMOUNT`. The unit
  price is floored at zero.

### 2.3 Order amount fields (P3.T7)

`Order` persists three monetary fields so reports never have to re-derive tax:

| Field | Meaning |
| :--- | :--- |
| `Subtotal` | Pre-tax, **post**-promotion-discount sum of line totals. |
| `TaxAmount` | Tax portion captured at checkout (`subtotal × rate`). |
| `TotalAmount` | Post-tax, **post**-loyalty-discount final charge (`subtotal + tax − loyaltyDiscount`). |

`DiscountAmount` stores the combined loyalty + promotion discount. Because tax
is captured at checkout, later changes to `tax_rate_percentage` do not
retroactively alter historical orders.

### 2.4 Discount audit

If the combined discount is **≥ 10%** of the pre-discount subtotal
(`discountAuditThresholdPct`), an `AuditLog(Action=DISCOUNT, Entity=Order)` row
is written with a JSON breakdown of subtotal, total discount, discount percent,
promotion discount, and loyalty discount. This feeds the
`/reports/audit/voids` report.

---

## 3. Return Flow

Customer returns are a two-phase workflow: a customer **requests** a return,
then staff **approve or reject** it. Implemented in
`internal/handlers/request_return.go` and `process_return.go`.

### 3.1 Request (`POST /returns/request`, `returns.request`)

Runs in a transaction:

1. Load the order by `order_number`, scoped to the requesting user
   (`user_id`) — customers can only return their own orders.
2. Reject if a `PENDING` return already exists for the order.
3. Create a `Return(Status=PENDING)`.
4. For each requested item, validate the return quantity does not exceed
   `quantity − returned_qty` on the order item.
5. Compute the refund per line with **additive tax**:
   `round2((unitPrice × qty) + (unitPrice × qty × taxRate))`, using
   `tax_rate_percentage`. Sum into `Return.RefundAmount`.
6. Persist `ReturnItem`s (each carrying a `Condition`: `GOOD`, `DAMAGED`, or
   `OPENED`).
7. Broadcast `RETURN_REQUESTED` over WebSocket and create notifications for
   `returns.manage` and `pos.view` permission holders.

### 3.2 Process (`POST /returns/:id/process`, `returns.manage`)

Accepts `{"action": "approve" | "reject"}`. Rejects if the return is not
`PENDING`. Runs in a transaction:

- **Reject**: set `Status=REJECTED`, record `ApprovedBy`/`ApprovedAt`, save.
  No stock or money movement.
- **Approve**: set `Status=APPROVED`, then:
  1. **Batch restoration** — for each `GOOD`-condition item, add the quantity
     back to the most recent batch for that product/location
     (`ORDER BY created_at DESC`), or create a new batch numbered
     `RET-<returnID>` if none exists. `DAMAGED`/`OPENED` items are **not**
     restocked.
  2. Write a `StockAdjustment(Type=STOCK_IN, ReasonCode=RETURN)` per restocked
     line.
  3. Increment `returned_qty` on each order item and set `is_returned = true`.
  4. Create a `Transaction(Status=REFUNDED)` for the refund amount.
  5. **Loyalty reversal** — deduct points proportional to the returned
     quantity: `pointsRedeemed × returnQty / itemQty` (integer division),
     summed across lines. Fails the transaction if the customer's balance is
     insufficient to reverse.
  6. Save the return.
- After commit: broadcast `RETURN_UPDATED`, notify `returns.manage` and
  `pos.view`, and (async) invalidate the `RETURNS_ANALYSIS`, `SHRINKAGE`, and
  `COGS_GMROI` report caches.

---

## 4. Bulk Import Lifecycle

Bulk product import is asynchronous and confirmation-gated, decoupled from the
HTTP request via RabbitMQ. Handlers in `internal/handlers/bulk.go`; processing
in `internal/consumers/bulk_consumer.go` and
`internal/services/bulk_import_service.go`.

### 4.1 State machine

```
        upload                 consumer validates CSV
QUEUED  ──────►  (processing)  ──────────────────────►  PENDING_CONFIRMATION
                                                              │
                                          confirm endpoint     │
                                          (publish confirm)    ▼
                                                          COMPLETED
```

| Status | Meaning |
| :--- | :--- |
| `QUEUED` | Job created on upload; file stored in MinIO (`bulk-imports` bucket); message published to the `bulk.import` routing key. |
| `PENDING_CONFIRMATION` | Consumer parsed and validated the CSV. The job result holds `totalRecords`, `validRecords`, `invalidRecords`, `errors`, and the list of new categories/sub-categories/suppliers/locations that would be created. **No products exist yet.** |
| `COMPLETED` | `POST /bulk/products/import/:jobId/confirm` was called and the confirm consumer created the products. |

### 4.2 Rules

- **Two-pass validation.** Pass 1 scans the file to collect referenced entity
  names and builds an in-memory cache (fetching only those rows from the DB).
  Pass 2 validates each row against the cache. A row needs all 11 columns; SKU,
  Name, and LocationName are required.
- **New entity discovery.** Categories, sub-categories, suppliers, and
  locations not found in the cache are flagged for creation and surfaced in the
  preview. A sub-category cannot be created without a parent category.
- **Confirmation is the only write path.** Since P3.T1, the validating consumer
  never auto-finalizes; products are created only by the confirm consumer.
- **Idempotent confirm (P3.T2).** Confirming a `COMPLETED` job is a no-op, and
  `finalizeImport` uses `ON CONFLICT DO NOTHING` so re-runs cannot trip the
  `products.sku` unique constraint.
- **Retries.** Jobs are created with `MaxRetries = 3`; the consumer tracks
  `RetryCount`, `LastError`, and `LastAttemptAt`.
- **Concurrency.** The bulk consumer runs `CONSUMER_CONCURRENCY` workers
  (default 1).

Export (`BULK_EXPORT`) follows the same QUEUED → processing → COMPLETED pattern
without a confirmation step.

---

## 5. Loyalty Tiers (settings-driven)

Loyalty is configured entirely through `system_settings`, so operators can tune
it without redeploying (P4.T18). Accounts are created lazily on a customer's
first checkout, starting at `Bronze` with 0 points.

### 5.1 Settings

Seeded in `migrations/0002_seed.up.sql`:

| Key | Default | Meaning |
| :--- | :--- | :--- |
| `loyalty_points_earning_rate` | `1` | Points earned per currency unit of net spend. |
| `loyalty_points_redemption_rate` | `0.01` | Currency value of one redeemed point. |
| `loyalty_tier_silver` | `500` | Points required for Silver. |
| `loyalty_tier_gold` | `2500` | Points required for Gold. |
| `loyalty_tier_platinum` | `10000` | Points required for Platinum. |

### 5.2 Earning & redemption

- **Redemption** (`handleLoyalty`): requires an attached customer. Discount =
  `pointsToRedeem × redemption_rate`, capped so it never exceeds the order
  total. Redeemed points are subtracted from the balance.
- **Earning**: `pointsEarned = int(netAmount × earning_rate)`, where
  `netAmount = postTaxTotal − loyaltyDiscount` (floored at 0). Points are added
  after redemption.
- **Tier recompute** (`updateLoyaltyTier`): after every balance change the tier
  is recalculated from the thresholds — Platinum ≥ gold ≥ silver, otherwise the
  account stays at its current tier (tiers are not downgraded below the
  computed band by this function).

### 5.3 Reversal on return

When a return is approved, redeemed points are reversed proportionally to the
returned quantity (see §3.2), not by re-applying the earning rate. This keeps
the reversal symmetric with the original redemption.

---

## 6. Related documents

- [architecture.md](./architecture.md) — layering, ADRs, deployment topology.
- [security.md](./security.md) — threat model and controls.
- [runbook.md](./runbook.md) — incident response and deploy/rollback.
- [access_control.md](./access_control.md) — permission-to-route mapping.
- [reports_user_guide.md](./reports_user_guide.md) — reporting reference.
