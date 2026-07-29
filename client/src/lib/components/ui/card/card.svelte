<script lang="ts">
	import type { HTMLAttributes } from 'svelte/elements';
	import { cn, type WithElementRef } from '$lib/utils.js';

	let {
		ref = $bindable(null),
		class: className,
		children,
		...restProps
	}: WithElementRef<HTMLAttributes<HTMLDivElement>> = $props();
</script>

<div
	bind:this={ref}
	data-slot="card"
	class={cn(
		'text-card-foreground relative flex flex-col gap-6 overflow-visible rounded-lg border border-border bg-card py-6 shadow-sm',
		className
	)}
	{...restProps}
>
	{@render children?.()}
</div>
