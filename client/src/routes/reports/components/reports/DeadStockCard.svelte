<script lang="ts">
	import * as Tooltip from '$lib/components/ui/tooltip';
	import { AlertTriangle, HelpCircle } from 'lucide-svelte';
	import type { DeadStockItem } from '$lib/types';
	import { formatCurrency } from '$lib/utils';
	import { t } from '$lib/i18n';

	type Props = {
		deadStock: DeadStockItem[];
	};

	let { deadStock }: Props = $props();
</script>

<div class="panel p-5">
	<div class="flex items-start gap-3">
		<span
			class="flex h-9 w-9 shrink-0 items-center justify-center rounded-[calc(var(--radius)_-_4px)] border border-border bg-danger/10 text-danger"
		>
			<AlertTriangle class="h-4 w-4" />
		</span>
		<div class="min-w-0">
			<h3 class="section-title flex items-center gap-1.5">
				{$t('reports.dead_stock.title')}
				<Tooltip.Root>
					<Tooltip.Trigger>
						<HelpCircle class="h-4 w-4 text-muted-foreground transition-colors hover:text-danger" />
					</Tooltip.Trigger>
					<Tooltip.Content>
						<p class="max-w-xs font-normal text-foreground">
							{$t('reports.dead_stock.description')}
						</p>
					</Tooltip.Content>
				</Tooltip.Root>
			</h3>
		</div>
	</div>

	<div class="mt-4 space-y-1">
		{#if deadStock.length === 0}
			<p class="py-8 text-center text-xs text-muted-foreground">
				{$t('common.no_data_available')}
			</p>
		{/if}
		{#each deadStock.slice(0, 5) as item}
			<div
				class="flex items-center justify-between gap-3 rounded-[calc(var(--radius)_-_6px)] px-2.5 py-2.5 transition-colors hover:bg-muted/50"
			>
				<p class="min-w-0 truncate text-sm font-medium text-foreground">
					{item.ProductName}
				</p>
				<div class="shrink-0 text-right">
					<p class="num-mono text-sm font-semibold text-danger">{formatCurrency(item.Value)}</p>
					<span class="badge-danger mt-1 text-[10px]">
						{item.DaysSinceLastSale}
						{$t('reports.dead_stock.days_idle')}
					</span>
				</div>
			</div>
		{/each}
	</div>
</div>
