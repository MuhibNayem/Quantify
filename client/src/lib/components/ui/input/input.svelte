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
			'flex h-12 w-full min-w-0 rounded-md border border-input bg-card px-4 py-2 text-sm text-foreground shadow-xs outline-none transition-all placeholder:text-muted-foreground disabled:cursor-not-allowed disabled:opacity-50',
			'focus:border-primary focus:ring-[3px] focus:ring-ring/30',
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
			'flex h-12 w-full min-w-0 rounded-md border border-input bg-card px-4 py-2 text-sm text-foreground shadow-xs outline-none transition-all placeholder:text-muted-foreground disabled:cursor-not-allowed disabled:opacity-50',
			'focus:border-primary focus:ring-[3px] focus:ring-ring/30',
			'aria-invalid:ring-destructive/20 aria-invalid:border-destructive',
			className
		)}
		{type}
		bind:value
		{...restProps}
	/>
{/if}
