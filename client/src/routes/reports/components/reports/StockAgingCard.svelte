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

<div class="liquid-panel col-span-2 overflow-hidden rounded-lg p-0">
	<div class="border-border border-b p-6">
		<div class="flex items-center gap-3">
			<div
				class="flex h-10 w-10 items-center justify-center rounded-md border border-border bg-warning/10 text-warning"
			>
				<Package class="h-5 w-5" />
			</div>
			<div>
				<h3 class="font-display text-foreground mb-0 flex items-center gap-2 text-lg font-semibold">
					{$t('reports.stock_aging.title')}
					<Tooltip.Root>
						<Tooltip.Trigger>
							<HelpCircle class="text-muted-foreground hover:text-warning h-4 w-4 transition-colors" />
						</Tooltip.Trigger>
						<Tooltip.Content>
							<p class="max-w-xs">{$t('reports.stock_aging.description')}</p>
						</Tooltip.Content>
					</Tooltip.Root>
				</h3>
				<p class="text-muted-foreground text-sm">
					{$t('reports.stock_aging.subtitle')}
				</p>
			</div>
		</div>
	</div>

	<div class="custom-scrollbar max-h-[500px] overflow-y-auto">
		<table class="ledger-table">
			<thead>
				<tr>
					<th>{$t('reports.stock_aging.headers.product')}</th>
					<th class="text-right">{$t('reports.stock_aging.headers.age')}</th>
					<th class="text-right">{$t('reports.stock_aging.headers.quantity')}</th>
					<th class="text-right">{$t('reports.stock_aging.headers.value')}</th>
				</tr>
			</thead>
			<tbody>
				{#each stockAgingFlat as item}
					<tr class="group">
						<td>
							<div class="text-foreground font-semibold">
								{item.ProductName}
							</div>
							<div
								class="border-border bg-muted text-muted-foreground mt-1 inline-block rounded border px-1.5 text-[11px] num-mono"
							>
								{item.SKU}
							</div>
						</td>
						<td class="text-right">
							<span class="badge-warning num-mono">
								{item.AgeDays}{$t('reports.stock_aging.days_suffix')}
							</span>
						</td>
						<td class="col-num">{item.Quantity}</td>
						<td class="col-num">{formatCurrency(item.Value)}</td>
					</tr>
				{/each}
			</tbody>
		</table>
	</div>
</div>
