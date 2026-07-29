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

<div class="liquid-panel rounded-lg p-6">
	<h3 class="font-display text-foreground mb-4 flex items-center gap-2 text-lg font-semibold">
		<AlertTriangle class="text-danger h-5 w-5" />
		{$t('reports.dead_stock.title')}
		<Tooltip.Root>
			<Tooltip.Trigger>
				<HelpCircle class="text-muted-foreground hover:text-danger h-4 w-4 transition-colors" />
			</Tooltip.Trigger>
			<Tooltip.Content>
				<p class="max-w-xs">{$t('reports.dead_stock.description')}</p>
			</Tooltip.Content>
		</Tooltip.Root>
	</h3>
	<div class="space-y-2">
		{#each deadStock.slice(0, 5) as item}
			<div
				class="border-border bg-card hover:bg-muted flex items-center justify-between rounded-md border p-4 transition-colors"
			>
				<div class="text-foreground max-w-[200px] truncate text-sm font-semibold">
					{item.ProductName}
				</div>
				<div class="text-right">
					<div class="text-danger text-sm font-semibold num-mono">{formatCurrency(item.Value)}</div>
					<div class="badge-danger mt-1 text-[10px]">
						{item.DaysSinceLastSale}
						{$t('reports.dead_stock.days_idle')}
					</div>
				</div>
			</div>
		{/each}
	</div>
</div>
