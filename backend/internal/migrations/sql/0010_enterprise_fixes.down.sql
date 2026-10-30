DROP INDEX IF EXISTS public.idx_order_items_location;
ALTER TABLE public.order_items DROP COLUMN IF EXISTS location_id;

DELETE FROM public.system_settings
WHERE key IN (
  'manual_discount_enabled',
  'manual_discount_max_percent',
  'manual_discount_max_amount',
  'manual_discount_require_reason'
);

DELETE FROM public.role_permissions rp
USING public.roles r, public.permissions p
WHERE rp.role_id = r.id
  AND rp.permission_id = p.id
  AND p.name = 'pos.discount'
  AND r.name IN ('Admin', 'Manager');

DELETE FROM public.permissions WHERE name = 'pos.discount';
