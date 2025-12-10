
<script context="module" lang="ts">
	export type DescriptionItem = {
			label: string;
			value?: string | number | null;
			hint?: string;
			icon?: ComponentType | null;
			badge?: {
				text: string;
				variant?: 'default' | 'success' | 'warning' | 'danger' | 'info';
			};
	};
</script>

<script lang="ts">
	import type { ComponentType } from 'svelte';
	import { cn } from '$lib/utils';

	

	export let items: DescriptionItem[] = [];
	export let columns: 1 | 2 | 3 = 2;
	export let dense = false;
	export let emptyText = 'No data available';

	const columnClass = {
		1: 'sm:grid-cols-1',
		2: 'sm:grid-cols-2',
		3: 'sm:grid-cols-3'
	}[columns];

	const badgeStyles: Record<NonNullable<DescriptionItem['badge']>['variant'], string> = {
		default: 'bg-slate-100 text-slate-700 border-slate-200',
		success: 'bg-emerald-100 text-emerald-700 border-emerald-200',
		warning: 'bg-amber-100 text-amber-700 border-amber-200',
		danger: 'bg-rose-100 text-rose-700 border-rose-200',
		info: 'bg-sky-100 text-sky-700 border-sky-200'
	};
</script>

{#if items.length === 0}
	<div class="text-center text-sm text-slate-500 py-6">{emptyText}</div>
{:else}
	<dl class={cn('grid grid-cols-1 gap-4', columnClass)}>
		{#each items as item (item.label)}
			<div class="rounded-2xl border border-slate-100 bg-white/90 p-4 shadow-sm">
				<dt class="text-xs font-semibold uppercase tracking-wide text-slate-500 flex items-center gap-2">
					{#if item.icon}
						<svelte:component this={item.icon} class="h-4 w-4 text-slate-400" />
					{/if}
					{item.label}
				</dt>
				<dd class={cn('text-slate-900', dense ? 'mt-1 text-sm font-semibold' : 'mt-2 text-base font-semibold')}>
					<span>{item.value ?? '—'}</span>
					{#if item.badge}
						<span
							class={cn(
								'ml-2 inline-flex items-center rounded-full border px-2 py-0.5 text-xs font-medium',
								badgeStyles[item.badge.variant ?? 'default']
							)}
						>
							{item.badge.text}
						</span>
					{/if}
				</dd>
				{#if item.hint}
					<p class="mt-1 text-xs text-slate-500">{item.hint}</p>
				{/if}
			</div>
		{/each}
	</dl>
{/if}
