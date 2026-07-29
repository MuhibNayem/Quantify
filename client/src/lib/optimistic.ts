import { toast } from 'svelte-sonner';

export async function optimistic<T>(
	apply: () => void,
	rollback: () => void,
	serverCall: () => Promise<T>,
	labels = { success: 'Saved', error: 'Failed to save' }
): Promise<T | undefined> {
	apply();
	try {
		const result = await serverCall();
		toast.success(labels.success);
		return result;
	} catch (e) {
		rollback();
		toast.error(labels.error);
		throw e;
	}
}
