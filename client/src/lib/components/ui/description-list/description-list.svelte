
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
	import { t } from '$lib/i18n';

	

	export let items: DescriptionItem[] = [];
	export let columns: 1 | 2 | 3 = 2;
	export let dense = false;
	export let emptyText = 'No data available';

	const columnClass = {
		1: 'sm:grid-cols-1',
		2: 'sm:grid-cols-2',
		3: 'sm:grid-cols-3'
	}[columns];

	const badgeStyles: Record<NonNullable<NonNullable<DescriptionItem['badge']>['variant']>, string> = {
		default: 'badge-neutral',
		success: 'badge-success',
		warning: 'badge-warning',
		danger: 'badge-danger',
		info: 'badge-info'
	};
</script>

{#if items.length === 0}
	<div class="text-center text-sm text-muted-foreground py-6">{emptyText}</div>
{:else}
	<dl class={cn('grid grid-cols-1 gap-4', columnClass)}>
		{#each items as item (item.label)}
			<div class="rounded-md border border-border bg-card p-4 shadow-sm">
				<dt class="overline flex items-center gap-2">
					{#if item.icon}
						<svelte:component this={item.icon} class="h-4 w-4 text-muted-foreground" />
					{/if}
					{$t(item.label)}
				</dt>
				<dd class={cn('text-foreground', dense ? 'mt-1 text-sm font-semibold' : 'mt-2 text-base font-semibold')}>
					<span>{item.value ?? '—'}</span>
					{#if item.badge}
						<span class={cn('ml-2', badgeStyles[item.badge.variant ?? 'default'])}>
							{item.badge.text}
						</span>
					{/if}
				</dd>
				{#if item.hint}
					<p class="mt-1 text-xs text-muted-foreground">{item.hint}</p>
				{/if}
			</div>
		{/each}
	</dl>
{/if}
