import { z } from 'zod/v4';

export const promotionSchema = z
	.object({
		name: z.string().min(1, 'Name is required'),
		description: z.string().optional(),
		discountType: z.enum(['PERCENTAGE', 'FIXED_AMOUNT']),
		discountValue: z.coerce.number().min(0, 'Must be positive'),
		startDate: z.string().min(1, 'Start date is required'),
		endDate: z.string().min(1, 'End date is required'),
		priority: z.coerce.number().min(0, 'Must be positive').optional(),
		isActive: z.boolean(),
		targetType: z.enum(['PRODUCT', 'CATEGORY', 'SUBCATEGORY']),
		productId: z.number().nullable().optional(),
		categoryId: z.number().nullable().optional(),
		subCategoryId: z.number().nullable().optional()
	})
	.refine(
		(data) => !data.startDate || !data.endDate || new Date(data.endDate) > new Date(data.startDate),
		{
			message: 'End date must be after start date',
			path: ['endDate']
		}
	);

export type PromotionFormData = z.infer<typeof promotionSchema>;
