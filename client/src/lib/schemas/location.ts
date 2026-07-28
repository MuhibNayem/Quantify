import { z } from 'zod/v4';

export const locationSchema = z.object({
	name: z.string().min(1, 'Name is required').max(255, 'Name is too long'),
	address: z.string().optional()
});

export type LocationFormData = z.infer<typeof locationSchema>;
