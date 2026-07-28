import { z } from 'zod/v4';

export const productSchema = z.object({
	sku: z.string().min(1, 'SKU is required'),
	name: z.string().min(1, 'Name is required').max(255, 'Name is too long'),
	description: z.string().optional(),
	categoryId: z.string().optional(),
	subCategoryId: z.string().optional(),
	supplierId: z.string().optional(),
	locationId: z.string().optional(),
	purchasePrice: z.coerce.number().min(0, 'Must be positive').optional(),
	sellingPrice: z.coerce.number().min(0, 'Must be positive').optional(),
	status: z.string().min(1, 'Status is required'),
	barCodeUPC: z.string().min(1, 'Barcode/UPC is required')
});

export type ProductFormData = z.infer<typeof productSchema>;
