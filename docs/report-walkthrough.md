# Real-Time Reports Implementation Walkthrough

This document outlines the implementation of 12 new real-time reports and how to verify them.

## 1. Overview of Changes

We have implemented a comprehensive reporting suite with real-time capabilities.

### Key Components Modified:
- **[internal/repository/reporting_repository.go](file:///home/amnayem/Projects/Quantify/backend/internal/repository/reporting_repository.go)**: Added SQL logic for 12 new reports (Stock Aging, Dead Stock, Supplier Performance, Heatmap, etc.).
- **[internal/services/reporting_service.go](file:///home/amnayem/Projects/Quantify/backend/internal/services/reporting_service.go)**: Added wrapper methods and updated [NotifyReportUpdate](file:///home/amnayem/Projects/Quantify/backend/internal/services/reporting_service.go#45-90) to broadcast actual data payloads.
- **[internal/handlers/reports.go](file:///home/amnayem/Projects/Quantify/backend/internal/handlers/reports.go)**: Added API endpoints for each report.
- **[internal/handlers/sales.go](file:///home/amnayem/Projects/Quantify/backend/internal/handlers/sales.go)**: Updated [Checkout](file:///home/amnayem/Projects/Quantify/backend/internal/handlers/sales.go#28-338) to trigger real-time updates via [ReportingService](file:///home/amnayem/Projects/Quantify/backend/internal/services/reporting_service.go#23-32).
- **[internal/router/router.go](file:///home/amnayem/Projects/Quantify/backend/internal/router/router.go)**: Registered new routes and updated dependency injection.
- **[internal/domain/audit.go](file:///home/amnayem/Projects/Quantify/backend/internal/domain/audit.go) & [pos.go](file:///home/amnayem/Projects/Quantify/backend/internal/domain/pos.go)**: Added new models for Audit Logs and Cash Drawer sessions.

## 2. New Reports Available

| Report Name | Endpoint | Permission | Description |
|---|---|---|---|
| Stock Aging | `/reports/stock-aging` | `reports.inventory` | Groups inventory by age (0-30, 31-60, 90+ days). |
| Dead Stock | `/reports/dead-stock` | `reports.inventory` | Products with stock > 0 but no sales in X days. |
| Supplier Performance | `/reports/supplier-performance` | `reports.financial` | Lead time and fill rate metrics. |
| Hourly Heatmap | `/reports/heatmap` | `reports.sales` | Sales intensity by hour and day of week. |
| Sales by Employee | `/reports/employee-sales` | `reports.sales` | Sales totals per user. |
| Category Drill-Down | `/reports/category-drilldown` | `reports.sales` | Sales and margin by category. |
| COGS & GMROI | `/reports/gmroi` | `reports.financial` | Gross Margin Return on Investment. |
| Void/Discount Audit | `/reports/audit/voids` | `reports.financial` | Log of sensitive POS actions. |
| Tax Liability | `/reports/tax-liability` | `reports.financial` | Tax collected (estimated). |
| Cash Reconciliation | `/reports/cash-reconciliation` | `reports.financial` | System vs Actual cash counts. |

## 3. Verification Steps

### Automated Tests
Run the backend tests to ensure no regressions:
```bash
go test ./...
```

### Manual Verification (Real-Time)
1. **Connect a WebSocket Client**:
   - Connect to `ws://localhost:8080/ws?token=<YOUR_JWT>`
   - Ensure the user has `reports.view` permission.

2. **Trigger a Sale**:
   - Perform a `POST /api/v1/sales/checkout` with items.

3. **Observe Updates**:
   - The WebSocket client should receive messages with `type` corresponding to reports (e.g., `HOURLY_HEATMAP`, `SALES_BY_EMPLOYEE`).
   - The `payload` will contain the updated report data JSON.

### API Verification
You can manually query the new endpoints using `curl` or Postman:

**Example: Get Hourly Heatmap**
```bash
curl -X GET "http://localhost:8080/api/v1/reports/heatmap?startDate=2023-01-01T00:00:00Z&endDate=2023-12-31T23:59:59Z" \
  -H "Authorization: Bearer <TOKEN>"
```

**Example: Get Dead Stock**
```bash
curl -X GET "http://localhost:8080/api/v1/reports/dead-stock?days=60" \
  -H "Authorization: Bearer <TOKEN>"
```
