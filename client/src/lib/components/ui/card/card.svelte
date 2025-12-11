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
		'text-card-foreground relative flex flex-col gap-6 overflow-visible rounded-2xl border border-white/50 bg-gradient-to-br from-white/80 to-white/40 py-6 shadow-[inset_1px_1px_2px_0_rgba(255,255,255,0.8),inset_-1px_-1px_2px_0_rgba(255,255,255,0.2),0_10px_30px_-5px_rgba(0,0,0,0.1)] backdrop-blur-3xl backdrop-saturate-[180%]',
		className
	)}
	{...restProps}
>
	{@render children?.()}
</div>
