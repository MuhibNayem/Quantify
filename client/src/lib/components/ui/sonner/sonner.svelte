<!-- client/src/lib/components/ui/sonner/sonner.svelte -->
<script lang="ts">
	import { Toaster as SvelteToaster } from 'svelte-sonner';
	import { cn } from '$lib/utils';

	type Placement =
		| 'top-left'
		| 'top-right'
		| 'bottom-left'
		| 'bottom-right'
		| 'top-center'
		| 'bottom-center';

	let {
		richColors = true,
		closeButton = false,
		position = 'bottom-right',
		theme = 'system',
		class: className = '',
		toastOptions,
		...rest
	} = $props<{
		richColors?: boolean;
		closeButton?: boolean;
		position?: Placement;
		theme?: 'light' | 'dark' | 'system';
		class?: string;
		toastOptions?: any;
	}>();
</script>

<SvelteToaster
	{richColors}
	{closeButton}
	{position}
	{theme}
	class={cn('toaster group', className)}
	toastOptions={{
		class:
			'group-[.toaster]:bg-card group-[.toaster]:border-border group-[.toaster]:text-foreground group-[.toaster]:rounded-lg group-[.toaster]:shadow-lg',
		descriptionClass: 'group-[.toast]:text-muted-foreground',
		actionButtonClass: 'group-[.toast]:bg-primary group-[.toast]:text-primary-foreground',
		cancelButtonClass: 'group-[.toast]:bg-muted group-[.toast]:text-foreground',
		...toastOptions
	}}
	{...rest}
/>
