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

<div
	class="rounded-[2.5rem] border border-white/60 bg-white/40 p-8 shadow-2xl shadow-blue-900/5 backdrop-blur-3xl"
>
	<h3 class="mb-6 flex items-center gap-2 text-lg font-bold text-slate-700">
		<DollarSign class="h-5 w-5 text-blue-500" />
		{$t('reports.tax.title')}
		<Tooltip.Root>
			<Tooltip.Trigger>
				<HelpCircle class="h-4 w-4 text-slate-400 transition-colors hover:text-blue-500" />
			</Tooltip.Trigger>
			<Tooltip.Content>
				<p class="max-w-xs">{$t('reports.tax.description')}</p>
			</Tooltip.Content>
		</Tooltip.Root>
	</h3>
	{#if taxLiability}
		<div class="space-y-4">
			<div
				class="rounded-[2rem] bg-gradient-to-br from-blue-600 to-indigo-700 p-6 text-white shadow-xl shadow-blue-500/20"
			>
				<div class="mb-1 text-sm font-bold text-blue-100">
					{$t('reports.tax.collected')}
				</div>
				<div class="text-3xl font-black">{formatCurrency(taxLiability.TotalTax)}</div>
			</div>

			<div class="space-y-3">
				{#each taxLiability.Breakdown as rate}
					<div
						class="flex items-center justify-between rounded-[1.5rem] border border-white/50 bg-white/60 p-4"
					>
						<div>
							<div class="text-xs font-bold tracking-wider text-slate-400 uppercase">
								{$t('reports.tax.rate')}
							</div>
							<div class="font-bold text-slate-700">
								{rate.RateName} ({formatPercent(rate.Rate / 100)})
							</div>
						</div>
						<div class="text-right">
							<div class="text-xs font-bold tracking-wider text-slate-400 uppercase">
								{$t('reports.tax.taxable_sales')}
							</div>
							<div class="font-bold text-slate-700">
								{formatCurrency(rate.TaxableAmount)}
							</div>
						</div>
					</div>
				{/each}
			</div>
		</div>
	{/if}
</div>
