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
			'group-[.toaster]:bg-white/90 group-[.toaster]:backdrop-blur-2xl group-[.toaster]:border-white/60 group-[.toaster]:text-slate-900 group-[.toaster]:rounded-2xl group-[.toaster]:shadow-[0_20px_40px_-10px_rgba(0,0,0,0.1)]',
		descriptionClass: 'group-[.toast]:text-slate-500',
		actionButtonClass: 'group-[.toast]:bg-slate-900 group-[.toast]:text-white',
		cancelButtonClass: 'group-[.toast]:bg-slate-100 group-[.toast]:text-slate-500',
		...toastOptions
	}}
	{...rest}
/>
