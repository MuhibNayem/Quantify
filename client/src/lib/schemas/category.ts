import { z } from 'zod/v4';

export const categorySchema = z.object({
	name: z.string().min(1, 'Name is required').max(255, 'Name is too long')
});

export const subCategorySchema = z.object({
	name: z.string().min(1, 'Name is required').max(255, 'Name is too long'),
	categoryId: z.string().min(1, 'Category is required')
});

export type CategoryFormData = z.infer<typeof categorySchema>;
export type SubCategoryFormData = z.infer<typeof subCategorySchema>;
