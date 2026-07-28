-- Remove default alert subscriptions
DELETE FROM public.alert_role_subscriptions
WHERE alert_type IN ('LOW_STOCK', 'OUT_OF_STOCK', 'OVERSTOCK', 'EXPIRY_ALERT')
  AND role = 'Admin';

-- Remove system settings
DELETE FROM public.system_settings
WHERE key IN (
  'business_name',
  'currency_symbol',
  'currency_code',
  'timezone',
  'locale',
  'return_window_days',
  'sales_trends_cache_ttl',
  'loyalty_points_earning_rate',
  'loyalty_points_redemption_rate',
  'loyalty_tier_silver',
  'loyalty_tier_gold',
  'loyalty_tier_platinum',
  'tax_rate_percentage'
);

-- Remove role-permission grants
DELETE FROM public.role_permissions
WHERE role_id IN (
  SELECT id FROM public.roles
  WHERE name IN ('Admin', 'Manager', 'Staff', 'Customer')
);

-- Remove roles
DELETE FROM public.roles
WHERE name IN ('Admin', 'Manager', 'Staff', 'Customer');

-- Remove permissions
DELETE FROM public.permissions
WHERE name IN (
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
  'bulk.import', 'bulk.export',
  'settings.view', 'settings.manage',
  'users.view', 'users.manage',
  'roles.view', 'roles.manage',
  'alerts.view', 'alerts.manage',
  'dashboard.view',
  'inventory.view', 'inventory.read', 'inventory.write',
  'crm.view', 'crm.read', 'crm.write',
  'pos.view',
  'reports.view',
  'notifications.read', 'notifications.write'
);
