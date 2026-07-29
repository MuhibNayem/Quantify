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
		'liquid-textarea field-sizing-content flex min-h-[120px] w-full px-4 py-3 text-base text-foreground outline-none transition-all placeholder:text-muted-foreground disabled:cursor-not-allowed disabled:opacity-50 md:text-sm',
		'aria-invalid:ring-destructive/20 aria-invalid:border-destructive',
		className
	)}
	bind:value
	{...restProps}
></textarea>
