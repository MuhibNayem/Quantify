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
		'field-sizing-content flex min-h-[120px] w-full rounded-2xl border border-white/60 bg-gradient-to-br from-white/90 to-white/60 px-4 py-3 text-base font-bold text-slate-900 shadow-[inset_0_2px_4px_rgba(255,255,255,0.9),0_4px_10px_rgba(0,0,0,0.05)] outline-none ring-1 ring-white/40 backdrop-blur-xl backdrop-saturate-[180%] transition-all placeholder:text-slate-500/80 disabled:cursor-not-allowed disabled:opacity-50 md:text-sm',
		'focus:scale-[1.005] focus:border-blue-500/30 focus:ring-4 focus:ring-blue-500/10',
		'aria-invalid:ring-destructive/20 aria-invalid:border-destructive',
		className
	)}
	bind:value
	{...restProps}
></textarea>
