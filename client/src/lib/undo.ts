import { toast } from 'svelte-sonner';

/**
 * Soft-delete with undo. Shows a toast with an Undo action;
 * only calls the real delete after the window expires.
 */
export function deleteWithUndo(
	label: string,
	performDelete: () => Promise<void>,
	onUndo?: () => void,
	windowMs = 5000
) {
	let cancelled = false;
	let timer: ReturnType<typeof setTimeout>;

	toast(`Deleted ${label}`, {
		duration: windowMs,
		action: {
			label: 'Undo',
			onClick: () => {
				cancelled = true;
				clearTimeout(timer);
				onUndo?.();
				toast.success(`Restored ${label}`);
			}
		}
	});

	timer = setTimeout(async () => {
		if (cancelled) return;
		try {
			await performDelete();
		} catch {
			toast.error(`Failed to delete ${label}`);
		}
	}, windowMs);
}
