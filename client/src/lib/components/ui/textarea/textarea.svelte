<script lang="ts">
	import { cn, type WithElementRef, type WithoutChildren } from '$lib/utils.js';
	import type { HTMLTextareaAttributes } from 'svelte/elements';

	let {
		ref = $bindable(null),
		value = $bindable(),
		class: className,
		...restProps
	}: WithoutChildren<WithElementRef<HTMLTextareaAttributes>> = $props();
</script>

<textarea
	bind:this={ref}
	data-slot="textarea"
	class={cn(
		'glass-textarea flex field-sizing-content min-h-[120px] w-full rounded-lg px-4 py-3 text-base text-foreground transition-all outline-none placeholder:text-muted-foreground disabled:cursor-not-allowed disabled:opacity-50 md:text-sm',
		'aria-invalid:border-destructive aria-invalid:ring-destructive/20',
		className
	)}
	bind:value
	{...restProps}
></textarea>
