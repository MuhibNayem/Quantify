<script lang="ts">
	import * as Tooltip from '$lib/components/ui/tooltip';
	import { DollarSign, HelpCircle } from 'lucide-svelte';
	import type { TaxLiabilityReport } from '$lib/types';
	import { formatCurrency, formatPercent } from '$lib/utils';
	import { t } from '$lib/i18n';

	type Props = {
		taxLiability: TaxLiabilityReport | null;
	};

	let { taxLiability }: Props = $props();
</script>

<div class="liquid-panel rounded-lg p-6">
	<h3 class="font-display text-foreground mb-4 flex items-center gap-2 text-lg font-semibold">
		<DollarSign class="text-primary h-5 w-5" />
		{$t('reports.tax.title')}
		<Tooltip.Root>
			<Tooltip.Trigger>
				<HelpCircle class="text-muted-foreground hover:text-primary h-4 w-4 transition-colors" />
			</Tooltip.Trigger>
			<Tooltip.Content>
				<p class="max-w-xs">{$t('reports.tax.description')}</p>
			</Tooltip.Content>
		</Tooltip.Root>
	</h3>
	{#if taxLiability}
		<div class="space-y-4">
			<div class="bg-primary text-primary-foreground rounded-lg p-6 shadow-sm">
				<div class="text-primary-foreground/80 mb-1 text-sm font-semibold">
					{$t('reports.tax.collected')}
				</div>
				<div class="kpi-num text-3xl">{formatCurrency(taxLiability.TotalTax)}</div>
			</div>

			<div class="space-y-2">
				{#each taxLiability.Breakdown as rate}
					<div
						class="border-border bg-card flex items-center justify-between rounded-md border p-4"
					>
						<div>
							<div class="overline">
								{$t('reports.tax.rate')}
							</div>
							<div class="text-foreground font-semibold">
								{rate.RateName} ({formatPercent(rate.Rate / 100)})
							</div>
						</div>
						<div class="text-right">
							<div class="overline">
								{$t('reports.tax.taxable_sales')}
							</div>
							<div class="text-foreground font-semibold num-mono">
								{formatCurrency(rate.TaxableAmount)}
							</div>
						</div>
					</div>
				{/each}
			</div>
		</div>
	{/if}
</div>
