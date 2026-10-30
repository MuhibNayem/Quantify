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

{#if gmroiStats}
	<div class="grid gap-4 sm:grid-cols-2 xl:grid-cols-4">
		<div class="panel p-5">
			<div class="flex items-start justify-between gap-2">
				<p class="overline">{$t('reports.financials.revenue')}</p>
				<Tooltip.Root>
					<Tooltip.Trigger>
						<HelpCircle
							class="h-3.5 w-3.5 text-muted-foreground transition-colors hover:text-primary"
						/>
					</Tooltip.Trigger>
					<Tooltip.Content>
						<p class="max-w-xs font-normal text-foreground">
							{$t('reports.financials.revenue_desc')}
						</p>
					</Tooltip.Content>
				</Tooltip.Root>
			</div>
			<p class="kpi-num mt-2 text-2xl text-foreground">
				{formatCurrency(gmroiStats.totalRevenue)}
			</p>
		</div>

		<div class="panel p-5">
			<div class="flex items-start justify-between gap-2">
				<p class="overline">{$t('reports.financials.cogs')}</p>
				<Tooltip.Root>
					<Tooltip.Trigger>
						<HelpCircle
							class="h-3.5 w-3.5 text-muted-foreground transition-colors hover:text-primary"
						/>
					</Tooltip.Trigger>
					<Tooltip.Content>
						<p class="max-w-xs font-normal text-foreground">
							{$t('reports.financials.cogs_desc')}
						</p>
					</Tooltip.Content>
				</Tooltip.Root>
			</div>
			<p class="kpi-num mt-2 text-2xl text-foreground">
				{formatCurrency(gmroiStats.totalCost)}
			</p>
		</div>

		<div class="panel p-5">
			<div class="flex items-start justify-between gap-2">
				<p class="overline">{$t('reports.financials.margin')}</p>
				<Tooltip.Root>
					<Tooltip.Trigger>
						<HelpCircle
							class="h-3.5 w-3.5 text-muted-foreground transition-colors hover:text-primary"
						/>
					</Tooltip.Trigger>
					<Tooltip.Content>
						<p class="max-w-xs font-normal text-foreground">
							{$t('reports.financials.margin_desc')}
						</p>
					</Tooltip.Content>
				</Tooltip.Root>
			</div>
			<p class="kpi-num mt-2 text-2xl text-success">
				{formatPercent(gmroiStats.marginPercent)}
			</p>
		</div>

		<div class="panel p-5">
			<div class="flex items-start justify-between gap-2">
				<p class="overline">
					<span class="border-b border-dotted border-muted-foreground/60 pb-0.5"
						>{$t('reports.financials.gmroi')}</span
					>
				</p>
				<Tooltip.Root>
					<Tooltip.Trigger>
						<HelpCircle
							class="h-3.5 w-3.5 text-muted-foreground transition-colors hover:text-primary"
						/>
					</Tooltip.Trigger>
					<Tooltip.Content>
						<p class="max-w-xs font-normal text-foreground">
							{$t('reports.financials.gmroi_desc')}
						</p>
					</Tooltip.Content>
				</Tooltip.Root>
			</div>
			<p class="kpi-num mt-2 text-2xl text-primary">
				{gmroiStats.gmroiIndex.toFixed(2)}
			</p>
		</div>
	</div>
{:else}
	<div class="panel p-8 text-center">
		<p class="text-sm text-muted-foreground">{$t('common.no_data_available')}</p>
	</div>
{/if}
