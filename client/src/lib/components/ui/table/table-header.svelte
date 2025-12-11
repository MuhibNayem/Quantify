<script lang="ts">
	import { cn, type WithElementRef } from '$lib/utils.js';
	import type { HTMLAttributes } from 'svelte/elements';

	let {
		ref = $bindable(null),
		class: className,
		children,
		...restProps
	}: WithElementRef<HTMLAttributes<HTMLTableSectionElement>> = $props();
</script>

<thead
	bind:this={ref}
	data-slot="table-header"
	class={cn(
		'bg-white/50 shadow-[0_4px_20px_-5px_rgba(0,0,0,0.05)] backdrop-blur-xl [&_tr]:border-b [&_tr]:border-white/30',
		className
	)}
	{...restProps}
>
	{@render children?.()}
</thead>
