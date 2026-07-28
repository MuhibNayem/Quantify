import { z } from 'zod/v4';

const optionalEmail = z
	.string()
	.refine((value) => value === '' || /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(value), {
		message: 'Enter a valid email address'
	})
	.optional();

export const supplierSchema = z.object({
	name: z.string().min(1, 'Name is required').max(255, 'Name is too long'),
	contactPerson: z.string().optional(),
	email: optionalEmail,
	phone: z.string().optional(),
	address: z.string().optional()
});

export type SupplierFormData = z.infer<typeof supplierSchema>;
