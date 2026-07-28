<script lang="ts">
	import * as Tooltip from '$lib/components/ui/tooltip';
	import { DollarSign, HelpCircle } from 'lucide-svelte';
	import type { CashReconciliation } from '$lib/types';
	import { formatCurrency } from '$lib/utils';
	import { t } from '$lib/i18n';

	type Props = {
		cashReconciliation: CashReconciliation[];
	};

	let { cashReconciliation }: Props = $props();
</script>

<div
	class="rounded-[2.5rem] border border-white/60 bg-white/40 p-8 shadow-2xl shadow-emerald-900/5 backdrop-blur-3xl"
>
	<h3 class="mb-6 flex items-center gap-3 text-lg font-bold text-slate-700">
		<DollarSign class="h-5 w-5 text-emerald-500" />
		{$t('reports.cash_reconciliation.title')}
		<Tooltip.Root>
			<Tooltip.Trigger>
				<HelpCircle class="h-4 w-4 text-slate-400 transition-colors hover:text-emerald-500" />
			</Tooltip.Trigger>
			<Tooltip.Content>
				<p class="max-w-xs">{$t('reports.cash_reconciliation.description')}</p>
			</Tooltip.Content>
		</Tooltip.Root>
	</h3>
	<div class="space-y-3">
		{#each cashReconciliation as item}
			<div
				class="flex items-center justify-between rounded-[1.5rem] border border-white/50 bg-white/60 p-4 transition-all hover:bg-white/80"
			>
				<span class="pl-2 text-sm font-bold text-slate-700">{item.CashierName}</span>
				<div class="text-right">
					<div class="mb-0.5 text-[9px] font-bold tracking-wide text-slate-400 uppercase">
						{$t('reports.cash_reconciliation.discrepancy')}
					</div>
					<div
						class="rounded-xl border border-slate-100 bg-white/50 px-3 py-1 font-bold {item.Discrepancy <
						0
							? 'text-red-500'
							: 'text-emerald-600'}"
					>
						{formatCurrency(item.Discrepancy)}
					</div>
				</div>
			</div>
		{/each}
	</div>
</div>
