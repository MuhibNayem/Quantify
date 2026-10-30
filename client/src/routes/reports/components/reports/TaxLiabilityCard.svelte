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

<div class="panel p-5">
	<div class="flex items-start gap-3">
		<span
			class="flex h-9 w-9 shrink-0 items-center justify-center rounded-[calc(var(--radius)_-_4px)] border border-border bg-primary/10 text-primary"
		>
			<DollarSign class="h-4 w-4" />
		</span>
		<div class="min-w-0">
			<h3 class="section-title flex items-center gap-1.5">
				{$t('reports.tax.title')}
				<Tooltip.Root>
					<Tooltip.Trigger>
						<HelpCircle
							class="h-4 w-4 text-muted-foreground transition-colors hover:text-primary"
						/>
					</Tooltip.Trigger>
					<Tooltip.Content>
						<p class="max-w-xs font-normal text-foreground">{$t('reports.tax.description')}</p>
					</Tooltip.Content>
				</Tooltip.Root>
			</h3>
		</div>
	</div>

	{#if taxLiability}
		<div class="mt-4 space-y-3">
			<div class="rounded-[calc(var(--radius)_-_2px)] border border-primary/20 bg-primary/5 p-4">
				<p class="text-primary overline">{$t('reports.tax.collected')}</p>
				<p class="kpi-num mt-1.5 text-3xl text-foreground">
					{formatCurrency(taxLiability.TotalTax)}
				</p>
			</div>

			<div class="space-y-1">
				{#each taxLiability.Breakdown as rate}
					<div
						class="flex items-center justify-between gap-3 rounded-[calc(var(--radius)_-_6px)] px-2.5 py-2 transition-colors hover:bg-muted/50"
					>
						<div class="min-w-0">
							<p class="overline">{$t('reports.tax.rate')}</p>
							<p class="truncate text-sm font-medium text-foreground">
								{rate.RateName} ({formatPercent(rate.Rate / 100)})
							</p>
						</div>
						<div class="shrink-0 text-right">
							<p class="overline">{$t('reports.tax.taxable_sales')}</p>
							<p class="num-mono text-sm font-semibold text-foreground">
								{formatCurrency(rate.TaxableAmount)}
							</p>
						</div>
					</div>
				{/each}
			</div>
		</div>
	{:else}
		<p class="mt-4 py-8 text-center text-xs text-muted-foreground">
			{$t('common.no_data_available')}
		</p>
	{/if}
</div>
