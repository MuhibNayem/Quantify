# Role-Based Access Control (RBAC) & Navigation

This document defines the permission requirements for each sidebar item and application route. These permissions control visibility in the UI and access enforcement in the backend.

## 1. Sidebar Navigation

| Section | Item | Route | Required Permission | Description |
| :--- | :--- | :--- | :--- | :--- |
| **Core** | Overview | `/` | `(Public/Auth)` | Basic dashboard access. Visible to all authenticated users. |
| **Workspace** | Catalog | `/catalog` | `products.read` | View product lists, categories, and suppliers. |
| | Operations | `/operations` | `inventory.view` | Access stock movements, barcode lookup, and logistics. |
| | Time Tracking | `/time-tracking` | `(Public/Auth)` | Employee clock-in/out. Access depends on employment status. |
| | Intelligence | `/intelligence` | `reports.view` | View basic business reports and forecasts. |
| | POS | `/pos` | `pos.access` | Access the Point of Sale terminal interface. |
| **Business** | CRM | `/crm` | `customers.read` | View customer lists and loyalty programs. |
| **Control** | Alerts | `/alerts` | `alerts.view` | View system alerts and notifications. |
| | Bulk Ops | `/bulk` | `bulk.import` OR `bulk.export` | Access bulk data import/export tools. |
| | User Access | `/users` | `users.view` | View system users list. |
| | Settings | `/settings` | `settings.view` | View system configuration and role management. |

---

## 2. Granular Route Permissions

Detailed breakdown of specific sub-routes and actions.

### /catalog
- **View List**: `products.read`
- **Create Product**: `products.write`
- **Edit Product**: `products.write`
- **Delete Product**: `products.delete`

### /operations
- **View Stock**: `inventory.view`
- **Adjust Stock**: `inventory.manage`
- **Stock Transfer**: `inventory.manage`

### /intelligence
- **Sales Reports**: `reports.sales`
- **Inventory Reports**: `reports.inventory`
- **Financial Reports**: `reports.financial`

### /crm
- **View Customers**: `customers.read`
- **Edit Customer**: `customers.write`
- **Manage Loyalty**: `loyalty.write`

### /users
- **View Users**: `users.view`
- **Create/Edit User**: `users.manage`
- **Approve User**: `users.manage` (Admin only usually)

### /settings
- **View General**: `settings.view`
- **Edit General**: `settings.manage`
- **Role Manager**: `roles.view`, `roles.manage`

---

## 3. Recommended Role Configuration

Default permission sets for common roles.

### Admin
- **Access**: Full System Access
- **Permissions**: `*` (All permissions)

### Manager
- **Access**: Store Operations, Staff, Reports
- **Permissions**:
    - `products.*`, `categories.*`, `suppliers.*`
    - `inventory.*`, `alerts.*`
    - `customers.*`, `loyalty.*`
    - `pos.access`
    - `reports.*`
    - `users.view`, `users.manage`
    - `bulk.*`

### Staff / Cashier
- **Access**: POS, Customer Lookup, basic stock view
- **Permissions**:
    - `pos.access`
    - `products.read`
    - `customers.read`
    - `inventory.view` (optional)
    - `time_tracking.self` (implied)

### Stock Clerk
- **Access**: Inventory Operations
- **Permissions**:
    - `products.read`
    - `inventory.view`, `inventory.manage`
    - `barcode.read`
    - `alerts.view`
