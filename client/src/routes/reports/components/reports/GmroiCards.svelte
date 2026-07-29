<script lang="ts">
	import * as Tooltip from '$lib/components/ui/tooltip';
	import { HelpCircle } from 'lucide-svelte';
	import { formatCurrency, formatPercent } from '$lib/utils';
	import { t } from '$lib/i18n';

	type Props = {
		gmroiStats: {
			totalRevenue: number;
			totalCost: number;
			marginPercent: number;
			gmroiIndex: number;
		} | null;
	};

	let { gmroiStats }: Props = $props();
</script>

<div class="flex flex-wrap gap-4">
	{#if gmroiStats}
		<div class="liquid-panel min-w-[200px] flex-1 rounded-lg p-6">
			<div class="overline mb-2 flex items-center gap-2">
				{$t('reports.financials.revenue')}
				<Tooltip.Root>
					<Tooltip.Trigger>
						<HelpCircle
							class="text-muted-foreground hover:text-primary h-3.5 w-3.5 transition-colors"
						/>
					</Tooltip.Trigger>
					<Tooltip.Content>
						<p class="text-foreground max-w-xs font-normal">
							{$t('reports.financials.revenue_desc')}
						</p>
					</Tooltip.Content>
				</Tooltip.Root>
			</div>
			<div class="kpi-num text-foreground text-xl md:text-2xl">
				{formatCurrency(gmroiStats.totalRevenue)}
			</div>
		</div>
		<div class="liquid-panel min-w-[200px] flex-1 rounded-lg p-6">
			<div class="overline mb-2 flex items-center gap-2">
				{$t('reports.financials.cogs')}
				<Tooltip.Root>
					<Tooltip.Trigger>
						<HelpCircle
							class="text-muted-foreground hover:text-primary h-3.5 w-3.5 transition-colors"
						/>
					</Tooltip.Trigger>
					<Tooltip.Content>
						<p class="text-foreground max-w-xs font-normal">
							{$t('reports.financials.cogs_desc')}
						</p>
					</Tooltip.Content>
				</Tooltip.Root>
			</div>
			<div class="kpi-num text-foreground text-xl md:text-2xl">
				{formatCurrency(gmroiStats.totalCost)}
			</div>
		</div>
		<div class="liquid-panel min-w-[200px] flex-1 rounded-lg p-6">
			<div class="overline mb-2 flex items-center gap-2">
				{$t('reports.financials.margin')}
				<Tooltip.Root>
					<Tooltip.Trigger>
						<HelpCircle
							class="text-muted-foreground hover:text-primary h-3.5 w-3.5 transition-colors"
						/>
					</Tooltip.Trigger>
					<Tooltip.Content>
						<p class="text-foreground max-w-xs font-normal">
							{$t('reports.financials.margin_desc')}
						</p>
					</Tooltip.Content>
				</Tooltip.Root>
			</div>
			<div class="kpi-num text-success text-xl md:text-2xl">
				{formatPercent(gmroiStats.marginPercent)}
			</div>
		</div>
		<div class="liquid-panel min-w-[200px] flex-1 rounded-lg p-6">
			<div class="overline mb-2 flex items-center gap-2">
				<span class="border-muted-foreground border-b border-dotted"
					>{$t('reports.financials.gmroi')}</span
				>
				<Tooltip.Root>
					<Tooltip.Trigger>
						<HelpCircle
							class="text-muted-foreground hover:text-primary h-3.5 w-3.5 transition-colors"
						/>
					</Tooltip.Trigger>
					<Tooltip.Content>
						<p class="text-foreground max-w-xs font-normal">
							{$t('reports.financials.gmroi_desc')}
						</p>
					</Tooltip.Content>
				</Tooltip.Root>
			</div>
			<div class="kpi-num text-primary text-xl md:text-2xl">
				{gmroiStats.gmroiIndex.toFixed(2)}
			</div>
		</div>
	{:else}
		<div class="text-muted-foreground col-span-full py-8 text-center">
			{$t('common.no_data_available')}
		</div>
	{/if}
</div>
