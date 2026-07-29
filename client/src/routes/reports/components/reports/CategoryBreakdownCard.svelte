<script lang="ts">
	import * as Tooltip from '$lib/components/ui/tooltip';
	import { PieChart, HelpCircle } from 'lucide-svelte';
	import type { CategoryPerformance } from '$lib/types';
	import { formatCurrency, formatPercent } from '$lib/utils';
	import { t } from '$lib/i18n';

	type Props = {
		categoryPerformance: CategoryPerformance[];
	};

	let { categoryPerformance }: Props = $props();
</script>

<div class="liquid-panel rounded-lg p-6">
	<h3 class="font-display text-foreground mb-4 flex items-center gap-2 text-lg font-semibold">
		<PieChart class="text-primary h-5 w-5" />
		{$t('reports.category.title')}
		<Tooltip.Root>
			<Tooltip.Trigger>
				<HelpCircle class="text-muted-foreground hover:text-primary h-4 w-4 transition-colors" />
			</Tooltip.Trigger>
			<Tooltip.Content>
				<p class="max-w-xs">{$t('reports.category.description')}</p>
			</Tooltip.Content>
		</Tooltip.Root>
	</h3>

	<div class="custom-scrollbar max-h-[350px] space-y-2 overflow-y-auto p-1 pr-2">
		{#each categoryPerformance as cat}
			<div
				class="group border-border bg-card hover:bg-muted space-y-3 rounded-md border p-3 transition-colors"
			>
				<div class="flex items-center justify-between">
					<span class="text-foreground text-base font-semibold">{cat.CategoryName}</span>
					<span class="bg-muted text-muted-foreground rounded-md px-2.5 py-1 text-[11px] font-semibold num-mono">
						{cat.ItemCount || cat.TotalUnits || 0}
						{$t('reports.category.items')}
					</span>
				</div>
				<div class="grid grid-cols-2 gap-3">
					<div class="border-border bg-muted/40 rounded-md border p-3">
						<div class="overline mb-1">
							{$t('reports.category.sales')}
						</div>
						<div class="text-foreground text-sm font-semibold num-mono">
							{formatCurrency(cat.TotalSales)}
						</div>
					</div>
					<div class="border-border bg-muted/40 rounded-md border p-3">
						<div class="overline mb-1">
							{$t('reports.category.margin')}
						</div>
						<div class="flex flex-col items-start">
							<div class="text-success text-sm font-semibold num-mono">
								{formatCurrency(cat.GrossMargin)}
							</div>
							<div class="text-success mt-0.5 text-[10px] font-semibold num-mono">
								{formatPercent(cat.MarginPercent / 100)}
							</div>
						</div>
					</div>
				</div>
			</div>
		{/each}
	</div>
</div>
