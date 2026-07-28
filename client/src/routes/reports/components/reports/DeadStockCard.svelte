<script lang="ts">
	import * as Tooltip from '$lib/components/ui/tooltip';
	import { AlertTriangle, HelpCircle } from 'lucide-svelte';
	import type { DeadStockItem } from '$lib/types';
	import { formatCurrency } from '$lib/utils';
	import { t } from '$lib/i18n';

	type Props = {
		deadStock: DeadStockItem[];
	};

	let { deadStock }: Props = $props();
</script>

<div
	class="rounded-[2.5rem] border border-white/60 bg-white/40 p-8 shadow-2xl shadow-red-900/5 backdrop-blur-3xl"
>
	<h3 class="mb-4 flex items-center gap-2 text-lg font-bold text-slate-700">
		<AlertTriangle class="h-5 w-5 text-red-500" />
		{$t('reports.dead_stock.title')}
		<Tooltip.Root>
			<Tooltip.Trigger>
				<HelpCircle class="h-4 w-4 text-slate-400 transition-colors hover:text-red-500" />
			</Tooltip.Trigger>
			<Tooltip.Content>
				<p class="max-w-xs">{$t('reports.dead_stock.description')}</p>
			</Tooltip.Content>
		</Tooltip.Root>
	</h3>
	<div class="space-y-3">
		{#each deadStock.slice(0, 5) as item}
			<div
				class="flex items-center justify-between rounded-[1.5rem] border border-red-100/50 bg-red-50/40 p-4 transition-all hover:scale-[1.01] hover:bg-red-50/70"
			>
				<div class="max-w-[200px] truncate text-sm font-bold text-slate-700">
					{item.ProductName}
				</div>
				<div class="text-right">
					<div class="text-sm font-bold text-red-600">{formatCurrency(item.Value)}</div>
					<div
						class="mt-1 inline-block rounded-full bg-red-100/50 px-2 py-0.5 text-[10px] font-bold text-red-400"
					>
						{item.DaysSinceLastSale}
						{$t('reports.dead_stock.days_idle')}
					</div>
				</div>
			</div>
		{/each}
	</div>
</div>
