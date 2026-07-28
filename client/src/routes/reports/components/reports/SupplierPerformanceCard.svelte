<script lang="ts">
	import * as Tooltip from '$lib/components/ui/tooltip';
	import { DollarSign, HelpCircle } from 'lucide-svelte';
	import type { SupplierPerformance } from '$lib/types';
	import { formatPercent } from '$lib/utils';
	import { t } from '$lib/i18n';

	type Props = {
		supplierPerf: SupplierPerformance[];
	};

	let { supplierPerf }: Props = $props();
</script>

<div
	class="rounded-[2.5rem] border border-white/60 bg-white/40 p-8 shadow-2xl shadow-emerald-900/5 backdrop-blur-3xl"
>
	<h3 class="mb-6 flex items-center gap-2 text-lg font-bold text-slate-700">
		<DollarSign class="h-5 w-5 text-emerald-500" />
		{$t('reports.supplier.title')}
		<Tooltip.Root>
			<Tooltip.Trigger>
				<HelpCircle class="h-4 w-4 text-slate-400 transition-colors hover:text-emerald-500" />
			</Tooltip.Trigger>
			<Tooltip.Content>
				<p class="max-w-xs">{$t('reports.supplier.description')}</p>
			</Tooltip.Content>
		</Tooltip.Root>
	</h3>
	<div class="space-y-3">
		{#each supplierPerf as sup}
			<div
				class="flex items-center justify-between rounded-[1.5rem] border border-white/40 bg-white/60 p-4 transition-all hover:bg-white/80"
			>
				<span class="text-sm font-bold text-slate-700">{sup.supplierName}</span>
				<div class="flex items-center gap-4">
					<div class="min-w-[60px] rounded-xl bg-slate-50/50 p-2 text-right">
						<div class="text-[9px] font-bold tracking-wider text-slate-400 uppercase">
							{$t('reports.supplier.time')}
						</div>
						<div class="font-medium text-slate-800">
							{sup.averageLeadTimeDays}{$t('reports.supplier.days_suffix')}
						</div>
					</div>
					<div class="min-w-[60px] rounded-xl bg-emerald-50/50 p-2 text-right">
						<div class="text-[9px] font-bold tracking-wider text-emerald-600/60 uppercase">
							{$t('reports.supplier.rate')}
						</div>
						<div class="font-bold text-emerald-600">
							{formatPercent(sup.onTimeDeliveryRate)}
						</div>
					</div>
				</div>
			</div>
		{/each}
	</div>
</div>
