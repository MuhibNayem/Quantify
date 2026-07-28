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

<div
	class="rounded-[2.5rem] border border-white/60 bg-white/40 p-8 shadow-2xl shadow-indigo-900/5 backdrop-blur-3xl"
>
	<h3 class="mb-2 flex items-center gap-2 text-lg font-bold text-slate-700">
		<PieChart class="h-5 w-5 text-indigo-500" />
		{$t('reports.category.title')}
		<Tooltip.Root>
			<Tooltip.Trigger>
				<HelpCircle class="h-4 w-4 text-slate-400 transition-colors hover:text-indigo-500" />
			</Tooltip.Trigger>
			<Tooltip.Content>
				<p class="max-w-xs">{$t('reports.category.description')}</p>
			</Tooltip.Content>
		</Tooltip.Root>
	</h3>

	<div class="custom-scrollbar max-h-[350px] space-y-2 overflow-y-auto p-1 pr-2">
		{#each categoryPerformance as cat}
			<div
				class="group space-y-3 rounded-[1.25rem] border border-white/50 bg-white/60 p-3 backdrop-blur-md transition-all hover:scale-[1.01] hover:bg-white/90 hover:shadow-md"
			>
				<div class="flex items-center justify-between">
					<span
						class="text-base font-bold text-slate-700 transition-colors group-hover:text-indigo-900"
						>{cat.CategoryName}</span
					>
					<span
						class="rounded-full bg-indigo-50 px-3 py-1 text-[11px] font-bold text-indigo-600 shadow-sm"
					>
						{cat.ItemCount || cat.TotalUnits || 0}
						{$t('reports.category.items')}
					</span>
				</div>
				<div class="grid grid-cols-2 gap-4">
					<div class="rounded-2xl border border-white/40 bg-white/50 p-3">
						<div class="mb-1 text-[10px] font-bold tracking-wider text-slate-400 uppercase">
							{$t('reports.category.sales')}
						</div>
						<div class="text-sm font-bold text-slate-800">
							{formatCurrency(cat.TotalSales)}
						</div>
					</div>
					<div class="rounded-2xl border border-white/40 bg-white/50 p-3">
						<div class="mb-1 text-[10px] font-bold tracking-wider text-slate-400 uppercase">
							{$t('reports.category.margin')}
						</div>
						<div class="flex flex-col items-start">
							<div class="text-sm font-bold text-emerald-600">
								{formatCurrency(cat.GrossMargin)}
							</div>
							<div
								class="mt-0.5 rounded-md bg-emerald-50 px-1.5 text-[10px] font-bold text-emerald-600/70"
							>
								{formatPercent(cat.MarginPercent / 100)}
							</div>
						</div>
					</div>
				</div>
			</div>
		{/each}
	</div>
</div>
