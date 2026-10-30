-- Seed permissions
INSERT INTO public.permissions (name, "group", description, created_at, updated_at)
VALUES
  ('products.read', 'Product Management', 'View products', NOW(), NOW()),
  ('products.write', 'Product Management', 'Create/Edit products', NOW(), NOW()),
  ('products.delete', 'Product Management', 'Delete products', NOW(), NOW()),
  ('categories.read', 'Product Management', 'View categories', NOW(), NOW()),
  ('categories.write', 'Product Management', 'Manage categories', NOW(), NOW()),
  ('locations.read', 'Inventory', 'View locations', NOW(), NOW()),
  ('locations.write', 'Inventory', 'Manage locations', NOW(), NOW()),
  ('suppliers.read', 'Inventory', 'View suppliers', NOW(), NOW()),
  ('suppliers.write', 'Inventory', 'Manage suppliers', NOW(), NOW()),
  ('barcode.read', 'Inventory', 'Lookup barcodes', NOW(), NOW()),
  ('replenishment.read', 'Inventory', 'View forecasts/suggestions', NOW(), NOW()),
  ('replenishment.write', 'Inventory', 'Generate forecasts and manage POs', NOW(), NOW()),
  ('customers.read', 'CRM', 'View customers', NOW(), NOW()),
  ('customers.write', 'CRM', 'Manage customers', NOW(), NOW()),
  ('loyalty.read', 'CRM', 'View loyalty info', NOW(), NOW()),
  ('loyalty.write', 'CRM', 'Manage loyalty points', NOW(), NOW()),
  ('pos.access', 'POS', 'Access Point of Sale terminal', NOW(), NOW()),
  ('orders.read', 'Orders', 'View order history', NOW(), NOW()),
  ('orders.manage', 'Orders', 'Manage orders', NOW(), NOW()),
  ('returns.request', 'POS', 'Request a return', NOW(), NOW()),
  ('returns.manage', 'POS', 'Approve/Reject returns', NOW(), NOW()),
  ('reports.sales', 'Reports', 'View sales reports', NOW(), NOW()),
  ('reports.inventory', 'Reports', 'View inventory reports', NOW(), NOW()),
  ('reports.financial', 'Reports', 'View financial reports', NOW(), NOW()),
  ('bulk.import', 'System', 'Import data in bulk', NOW(), NOW()),
  ('bulk.export', 'System', 'Export data in bulk', NOW(), NOW()),
  ('settings.view', 'Settings', 'View system settings', NOW(), NOW()),
  ('settings.manage', 'Settings', 'Edit system settings', NOW(), NOW()),
  ('users.view', 'Access Control', 'View users', NOW(), NOW()),
  ('users.manage', 'Access Control', 'Manage users', NOW(), NOW()),
  ('roles.view', 'Access Control', 'View roles', NOW(), NOW()),
  ('roles.manage', 'Access Control', 'Manage roles', NOW(), NOW()),
  ('alerts.view', 'Inventory', 'View alerts', NOW(), NOW()),
  ('alerts.manage', 'Inventory', 'Resolve alerts', NOW(), NOW()),
  ('dashboard.view', 'System', 'View dashboard overview', NOW(), NOW()),
  ('inventory.view', 'Inventory', 'View inventory module', NOW(), NOW()),
  ('inventory.read', 'Inventory', 'Read inventory data', NOW(), NOW()),
  ('inventory.write', 'Inventory', 'Modify inventory data', NOW(), NOW()),
  ('crm.view', 'CRM', 'View CRM module', NOW(), NOW()),
  ('crm.read', 'CRM', 'Read CRM data', NOW(), NOW()),
  ('crm.write', 'CRM', 'Modify CRM data', NOW(), NOW()),
  ('pos.view', 'POS', 'View POS module', NOW(), NOW()),
  ('reports.view', 'Reports', 'View reports', NOW(), NOW()),
  ('notifications.read', 'System', 'View consolidated notifications', NOW(), NOW()),
  ('notifications.write', 'System', 'Mark notifications as read/unread', NOW(), NOW())
ON CONFLICT DO NOTHING;

-- Seed roles
INSERT INTO public.roles (name, description, is_system, created_at, updated_at)
VALUES
  ('Admin', 'Full system access', true, NOW(), NOW()),
  ('Manager', 'Store management access', true, NOW(), NOW()),
  ('Staff', 'Standard employee access', true, NOW(), NOW()),
  ('Customer', 'Customer portal access', true, NOW(), NOW())
ON CONFLICT DO NOTHING;

-- Grant all permissions to Admin
INSERT INTO public.role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM public.roles r
CROSS JOIN public.permissions p
WHERE r.name = 'Admin'
ON CONFLICT DO NOTHING;

-- Grant permissions to Manager
INSERT INTO public.role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM public.roles r
CROSS JOIN public.permissions p
WHERE r.name = 'Manager'
  AND p.name IN (
    'products.read', 'products.write', 'products.delete',
    'categories.read', 'categories.write',
    'locations.read', 'locations.write',
    'suppliers.read', 'suppliers.write',
    'barcode.read',
    'replenishment.read', 'replenishment.write',
    'customers.read', 'customers.write',
    'loyalty.read', 'loyalty.write',
    'pos.access',
    'orders.read', 'orders.manage',
    'returns.request', 'returns.manage',
    'reports.sales', 'reports.inventory', 'reports.financial',
    'alerts.view', 'alerts.manage',
    'bulk.import', 'bulk.export',
    'users.view', 'users.manage',
    'settings.view',
    'dashboard.view',
    'inventory.view', 'crm.view', 'pos.view',
    'notifications.read', 'notifications.write'
  )
ON CONFLICT DO NOTHING;

-- Grant permissions to Staff
INSERT INTO public.role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM public.roles r
CROSS JOIN public.permissions p
WHERE r.name = 'Staff'
  AND p.name IN (
    'products.read',
    'categories.read',
    'locations.read',
    'suppliers.read',
    'barcode.read',
    'customers.read',
    'pos.access',
    'orders.read',
    'returns.request',
    'alerts.view',
    'dashboard.view',
    'inventory.view', 'crm.view', 'pos.view',
    'notifications.read', 'notifications.write'
  )
ON CONFLICT DO NOTHING;

-- Seed system settings
INSERT INTO public.system_settings (key, value, "group", type, description, created_at, updated_at)
VALUES
  ('business_name', 'Quantify Business', 'General', 'string', NULL, NOW(), NOW()),
  ('currency_symbol', '$', 'General', 'string', NULL, NOW(), NOW()),
  ('currency_code', 'USD', 'General', 'string', 'ISO 4217 Currency Code', NOW(), NOW()),
  ('timezone', 'UTC', 'General', 'string', NULL, NOW(), NOW()),
  ('locale', 'en-US', 'General', 'string', 'System Locale (e.g., en-US, bn-BD)', NOW(), NOW()),
  ('return_window_days', '30', 'Policy', 'number', 'Number of days allowing returns after purchase', NOW(), NOW()),
  ('sales_trends_cache_ttl', '5m', 'System', 'string', 'Cache TTL for sales trends report', NOW(), NOW()),
  ('loyalty_points_earning_rate', '1', 'Loyalty', 'number', 'Points earned per currency unit spent', NOW(), NOW()),
  ('loyalty_points_redemption_rate', '0.01', 'Loyalty', 'number', 'Currency value of 1 point', NOW(), NOW()),
  ('loyalty_tier_silver', '500', 'Loyalty', 'number', 'Points required for Silver tier', NOW(), NOW()),
  ('loyalty_tier_gold', '2500', 'Loyalty', 'number', 'Points required for Gold tier', NOW(), NOW()),
  ('loyalty_tier_platinum', '10000', 'Loyalty', 'number', 'Points required for Platinum tier', NOW(), NOW()),
  ('tax_rate_percentage', '0', 'Financial', 'number', 'Default tax rate percentage', NOW(), NOW())
ON CONFLICT DO NOTHING;

-- Seed default alert subscriptions
INSERT INTO public.alert_role_subscriptions (alert_type, role, created_at, updated_at)
VALUES
  ('LOW_STOCK', 'Admin', NOW(), NOW()),
  ('OUT_OF_STOCK', 'Admin', NOW(), NOW()),
  ('OVERSTOCK', 'Admin', NOW(), NOW()),
  ('EXPIRY_ALERT', 'Admin', NOW(), NOW())
ON CONFLICT DO NOTHING;
