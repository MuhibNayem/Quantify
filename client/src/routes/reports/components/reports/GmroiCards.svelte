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
		<div
			class="min-w-[200px] flex-1 rounded-[2rem] border border-white/50 bg-white/40 p-6 shadow-lg shadow-emerald-900/5 backdrop-blur-xl transition-transform hover:scale-[1.02]"
		>
			<div
				class="mb-2 flex items-center gap-2 text-xs font-bold tracking-wider text-slate-500 uppercase"
			>
				{$t('reports.financials.revenue')}
				<Tooltip.Root>
					<Tooltip.Trigger>
						<HelpCircle
							class="h-3.5 w-3.5 text-slate-400 transition-colors hover:text-emerald-500"
						/>
					</Tooltip.Trigger>
					<Tooltip.Content>
						<p class="max-w-xs font-normal text-slate-700">
							{$t('reports.financials.revenue_desc')}
						</p>
					</Tooltip.Content>
				</Tooltip.Root>
			</div>
			<div class="text-xl font-black text-slate-800 md:text-2xl">
				{formatCurrency(gmroiStats.totalRevenue)}
			</div>
		</div>
		<div
			class="min-w-[200px] flex-1 rounded-[2rem] border border-white/50 bg-white/40 p-6 shadow-lg shadow-emerald-900/5 backdrop-blur-xl transition-transform hover:scale-[1.02]"
		>
			<div
				class="mb-2 flex items-center gap-2 text-xs font-bold tracking-wider text-slate-500 uppercase"
			>
				{$t('reports.financials.cogs')}
				<Tooltip.Root>
					<Tooltip.Trigger>
						<HelpCircle
							class="h-3.5 w-3.5 text-slate-400 transition-colors hover:text-emerald-500"
						/>
					</Tooltip.Trigger>
					<Tooltip.Content>
						<p class="max-w-xs font-normal text-slate-700">
							{$t('reports.financials.cogs_desc')}
						</p>
					</Tooltip.Content>
				</Tooltip.Root>
			</div>
			<div class="text-xl font-black text-slate-600 md:text-2xl">
				{formatCurrency(gmroiStats.totalCost)}
			</div>
		</div>
		<div
			class="min-w-[200px] flex-1 rounded-[2rem] border border-white/50 bg-white/40 p-6 shadow-lg shadow-emerald-900/5 backdrop-blur-xl transition-transform hover:scale-[1.02]"
		>
			<div
				class="mb-2 flex items-center gap-2 text-xs font-bold tracking-wider text-slate-500 uppercase"
			>
				{$t('reports.financials.margin')}
				<Tooltip.Root>
					<Tooltip.Trigger>
						<HelpCircle
							class="h-3.5 w-3.5 text-slate-400 transition-colors hover:text-emerald-500"
						/>
					</Tooltip.Trigger>
					<Tooltip.Content>
						<p class="max-w-xs font-normal text-slate-700">
							{$t('reports.financials.margin_desc')}
						</p>
					</Tooltip.Content>
				</Tooltip.Root>
			</div>
			<div class="text-xl font-black text-emerald-600 md:text-2xl">
				{formatPercent(gmroiStats.marginPercent)}
			</div>
		</div>
		<div
			class="min-w-[200px] flex-1 rounded-[2rem] border border-white/50 bg-white/40 p-6 shadow-lg shadow-emerald-900/5 backdrop-blur-xl transition-transform hover:scale-[1.02]"
		>
			<div
				class="mb-2 flex items-center gap-2 text-xs font-bold tracking-wider text-slate-500 uppercase"
			>
				<span class="border-b border-dotted border-slate-400 decoration-slate-400"
					>{$t('reports.financials.gmroi')}</span
				>
				<Tooltip.Root>
					<Tooltip.Trigger>
						<HelpCircle
							class="h-3.5 w-3.5 text-slate-400 transition-colors hover:text-emerald-500"
						/>
					</Tooltip.Trigger>
					<Tooltip.Content>
						<p class="max-w-xs font-normal text-slate-700">
							{$t('reports.financials.gmroi_desc')}
						</p>
					</Tooltip.Content>
				</Tooltip.Root>
			</div>
			<div class="text-xl font-black text-purple-600 md:text-2xl">
				{gmroiStats.gmroiIndex.toFixed(2)}
			</div>
		</div>
	{:else}
		<div class="col-span-full py-8 text-center text-slate-500">
			{$t('common.no_data_available')}
		</div>
	{/if}
</div>
