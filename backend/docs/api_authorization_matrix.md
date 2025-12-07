# API Authorization Matrix

This table summarizes the required roles for each group of `/api/v1` endpoints. Roles are hierarchical (`Admin` ≥ `Manager` ≥ `Staff`). Any role listed can access the endpoint; unlisted roles are denied.

| Endpoint Group | Description | Required Role(s) |
| --- | --- | --- |
| `POST /api/v1/products`<br>`PUT /api/v1/products/:id`<br>`DELETE /api/v1/products/:id`<br>`POST /api/v1/products/:id/stock/batches` | Create/update/delete products and batches | Admin, Manager |
| `POST /api/v1/products/:id/stock/adjustments`<br>`POST /api/v1/inventory/transfers` | Operational stock adjustments & transfers | Admin, Manager, Staff |
| `GET /api/v1/products/*` (read-only) | Product catalog queries | Admin, Manager, Staff |
| `POST/PUT/DELETE /api/v1/categories/*`<br>`/sub-categories/*` | Category maintenance | Admin, Manager |
| `POST/PUT/DELETE /api/v1/suppliers/*` | Supplier maintenance | Admin, Manager |
| `POST/PUT/DELETE /api/v1/locations/*` | Location maintenance | Admin, Manager |
| `POST /api/v1/replenishment/*` (forecasting, PO lifecycle) | Demand & purchase order workflows | Admin, Manager |
| `POST/GET /api/v1/reports/*`<br>`POST /api/v1/jobs/:id/cancel` | Reporting APIs and job control | Admin, Manager |
| `/api/v1/alerts/*` | Alert listing, resolution, notification settings | Admin, Manager |
| `/api/v1/bulk/*` | Bulk import/export flows | Admin, Manager |
| `/api/v1/crm/*` | CRM & loyalty management | Admin, Manager |
| `/api/v1/time-tracking/*` | Time clock endpoints | Admin, Manager, Staff |
| `GET /api/v1/users`<br>`PUT/DELETE /api/v1/users/:id`<br>`PUT /api/v1/users/:id/approve` | User administration | Admin only |
| `GET /api/v1/users/:id` | Fetch a user; Admin can access any user, others can only access themselves | Admin (any user) or resource owner |
| `POST /api/v1/users/refresh-token`<br>`POST /api/v1/users/logout` | Token refresh & logout | Authenticated user |

Public (unauthenticated) endpoints remain:

- `/health`, `/metrics`, `/ws`, `/swagger/*`
- `/webhooks`, `/payment/*`
- `/api/v1/users/register`, `/api/v1/users/login`

Refer to `internal/router/router.go` for the definitive middleware assignments. Any new endpoint should extend this matrix when merged.
