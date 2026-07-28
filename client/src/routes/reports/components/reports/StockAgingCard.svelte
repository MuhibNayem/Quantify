<script lang="ts">
	import * as Tooltip from '$lib/components/ui/tooltip';
	import { Package, HelpCircle } from 'lucide-svelte';
	import { formatCurrency } from '$lib/utils';
	import { t } from '$lib/i18n';

	type Props = {
		stockAgingFlat: any[];
	};

	let { stockAgingFlat }: Props = $props();
</script>

<div
	class="relative col-span-2 overflow-hidden rounded-[2.5rem] border border-white/60 bg-white/40 p-0 shadow-2xl shadow-orange-900/5 backdrop-blur-3xl"
>
	<div class="border-b border-white/30 bg-gradient-to-r from-orange-50/40 to-transparent p-8">
		<div class="flex items-center gap-4">
			<div
				class="flex h-12 w-12 items-center justify-center rounded-2xl border border-white/60 bg-gradient-to-br from-orange-50 to-amber-50/50 text-orange-600 shadow-inner"
			>
				<Package class="h-6 w-6" />
			</div>
			<div>
				<h3 class="mb-0 flex items-center gap-2 text-xl font-bold tracking-tight text-slate-800">
					{$t('reports.stock_aging.title')}
					<Tooltip.Root>
						<Tooltip.Trigger>
							<HelpCircle class="h-4 w-4 text-slate-400 transition-colors hover:text-orange-500" />
						</Tooltip.Trigger>
						<Tooltip.Content>
							<p class="max-w-xs">{$t('reports.stock_aging.description')}</p>
						</Tooltip.Content>
					</Tooltip.Root>
				</h3>
				<p class="text-sm font-medium text-slate-500">
					{$t('reports.stock_aging.subtitle')}
				</p>
			</div>
		</div>
	</div>

	<div class="custom-scrollbar max-h-[500px] overflow-y-auto">
		<table class="w-full border-collapse text-left text-sm">
			<thead
				class="sticky top-0 z-10 border-b border-white/30 bg-white/40 font-bold text-slate-500 backdrop-blur-xl"
			>
				<tr>
					<th class="p-5 pl-8">{$t('reports.stock_aging.headers.product')}</th>
					<th class="p-5 text-right">{$t('reports.stock_aging.headers.age')}</th>
					<th class="p-5 text-right">{$t('reports.stock_aging.headers.quantity')}</th>
					<th class="p-5 pr-8 text-right">{$t('reports.stock_aging.headers.value')}</th>
				</tr>
			</thead>
			<tbody class="divide-y divide-white/20">
				{#each stockAgingFlat as item}
					<tr class="group transition-colors hover:bg-orange-50/30">
						<td class="p-5 pl-8">
							<div class="font-bold text-slate-700 transition-colors group-hover:text-orange-900">
								{item.ProductName}
							</div>
							<div
								class="mt-1 inline-block rounded border border-white/40 bg-white/60 px-1.5 font-mono text-[11px] text-slate-400"
							>
								{item.SKU}
							</div>
						</td>
						<td class="p-5 text-right">
							<span
								class="inline-flex items-center rounded-lg border border-orange-200/50 bg-orange-100/80 px-2.5 py-1 text-xs font-bold text-orange-800 shadow-sm"
							>
								{item.AgeDays}{$t('reports.stock_aging.days_suffix')}
							</span>
						</td>
						<td class="p-5 text-right font-medium text-slate-600">{item.Quantity}</td>
						<td class="p-5 pr-8 text-right font-bold text-slate-800"
							>{formatCurrency(item.Value)}</td
						>
					</tr>
				{/each}
			</tbody>
		</table>
	</div>
</div>
