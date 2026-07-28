import { z } from 'zod/v4';

export const customerSchema = z.object({
	name: z.string().min(1, 'Name is required'),
	email: z.string().email('Enter a valid email address'),
	phone: z.string().min(1, 'Phone number is required')
});

export type CustomerFormData = z.infer<typeof customerSchema>;
