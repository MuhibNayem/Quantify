<script lang="ts">
	import type { HTMLInputAttributes, HTMLInputTypeAttribute } from 'svelte/elements';
	import { cn, type WithElementRef } from '$lib/utils.js';

	type InputType = Exclude<HTMLInputTypeAttribute, 'file'>;

	type Props = WithElementRef<
		Omit<HTMLInputAttributes, 'type'> &
			({ type: 'file'; files?: FileList } | { type?: InputType; files?: undefined })
	>;

	let {
		ref = $bindable(null),
		value = $bindable(),
		type,
		files = $bindable(),
		class: className,
		...restProps
	}: Props = $props();
</script>

{#if type === 'file'}
	<input
		bind:this={ref}
		data-slot="input"
		class={cn(
			'flex h-12 w-full min-w-0 rounded-2xl border border-white/60 bg-gradient-to-br from-white/90 to-white/60 px-4 py-2 text-base font-bold text-slate-900 shadow-[inset_0_2px_4px_rgba(255,255,255,0.9),0_4px_10px_rgba(0,0,0,0.05)] outline-none ring-1 ring-white/40 backdrop-blur-xl backdrop-saturate-[180%] transition-all placeholder:text-slate-500/80 disabled:cursor-not-allowed disabled:opacity-50 md:text-sm',
			'focus:scale-[1.01] focus:border-blue-500/30 focus:ring-4 focus:ring-blue-500/10',
			'aria-invalid:ring-destructive/20 aria-invalid:border-destructive',
			className
		)}
		type="file"
		bind:files
		bind:value
		{...restProps}
	/>
{:else}
	<input
		bind:this={ref}
		data-slot="input"
		class={cn(
			'flex h-12 w-full min-w-0 rounded-2xl border border-white/60 bg-gradient-to-br from-white/90 to-white/60 px-4 py-2 text-base font-bold text-slate-900 shadow-[inset_0_2px_4px_rgba(255,255,255,0.9),0_4px_10px_rgba(0,0,0,0.05)] outline-none ring-1 ring-white/40 backdrop-blur-xl backdrop-saturate-[180%] transition-all placeholder:text-slate-500/80 disabled:cursor-not-allowed disabled:opacity-50 md:text-base',
			'focus:scale-[1.01] focus:border-blue-500/30 focus:ring-4 focus:ring-blue-500/10',
			'aria-invalid:ring-destructive/20 aria-invalid:border-destructive',
			className
		)}
		{type}
		bind:value
		{...restProps}
	/>
{/if}
