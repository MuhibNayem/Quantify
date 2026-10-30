-- 0010_enterprise_fixes
-- Closes seed gaps that broke the manual-discount path and adds the
-- order-line location needed to restock returns at the correct site.

-- POS discount permission (present only in the retired in-code seeder).
INSERT INTO public.permissions (name, "group", description, created_at, updated_at)
VALUES ('pos.discount', 'POS', 'Apply manual discounts at checkout', NOW(), NOW())
ON CONFLICT DO NOTHING;

-- Admins and managers may apply manual discounts. Admin permissions were
-- granted by a one-time CROSS JOIN in 0002, so new permissions need an
-- explicit grant.
INSERT INTO public.role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM public.roles r
CROSS JOIN public.permissions p
WHERE r.name IN ('Admin', 'Manager')
  AND p.name = 'pos.discount'
ON CONFLICT DO NOTHING;

-- Manual-discount policy settings (previously only in the retired seeder).
INSERT INTO public.system_settings (key, value, "group", type, description, created_at, updated_at)
VALUES
  ('manual_discount_enabled', 'true', 'Policy', 'boolean', 'Allow employees to apply manual discounts at checkout', NOW(), NOW()),
  ('manual_discount_max_percent', '10', 'Policy', 'number', 'Maximum percentage discount an employee may apply', NOW(), NOW()),
  ('manual_discount_max_amount', '0', 'Policy', 'number', 'Maximum fixed-amount discount (0 = no cap)', NOW(), NOW()),
  ('manual_discount_require_reason', 'true', 'Policy', 'boolean', 'Require a reason note for manual discounts', NOW(), NOW())
ON CONFLICT DO NOTHING;

-- Remember which location each sale line deducted from so returns restock
-- the same site instead of the product's default location.
ALTER TABLE public.order_items
  ADD COLUMN IF NOT EXISTS location_id bigint REFERENCES public.locations(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_order_items_location ON public.order_items (location_id);
