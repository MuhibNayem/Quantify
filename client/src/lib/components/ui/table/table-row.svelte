<script lang="ts">
	import { cn, type WithElementRef } from '$lib/utils.js';
	import type { HTMLAttributes } from 'svelte/elements';

	let {
		ref = $bindable(null),
		class: className,
		children,
		...restProps
	}: WithElementRef<HTMLAttributes<HTMLTableRowElement>> = $props();
</script>

<tr
	bind:this={ref}
	data-slot="table-row"
	class={cn(
		'border-b border-gray-100/10 transition-all duration-300 hover:bg-gradient-to-r hover:from-blue-200/30 hover:to-indigo-200/20 hover:shadow-[inset_0_0_0_1px_rgba(255,255,255,0.4),0_2px_8px_rgba(59,130,246,0.1)] hover:backdrop-blur-md data-[state=selected]:bg-blue-50/50',
		className
	)}
	{...restProps}
>
	{@render children?.()}
</tr>
