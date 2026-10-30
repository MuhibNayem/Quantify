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

<div class="panel col-span-2 overflow-hidden">
	<div class="panel-header">
		<div class="flex items-start gap-3">
			<span
				class="flex h-9 w-9 shrink-0 items-center justify-center rounded-[calc(var(--radius)_-_4px)] border border-border bg-warning/10 text-warning"
			>
				<Package class="h-4 w-4" />
			</span>
			<div class="min-w-0">
				<h3 class="section-title flex items-center gap-1.5">
					{$t('reports.stock_aging.title')}
					<Tooltip.Root>
						<Tooltip.Trigger>
							<HelpCircle
								class="h-4 w-4 text-muted-foreground transition-colors hover:text-warning"
							/>
						</Tooltip.Trigger>
						<Tooltip.Content>
							<p class="max-w-xs font-normal text-foreground">
								{$t('reports.stock_aging.description')}
							</p>
						</Tooltip.Content>
					</Tooltip.Root>
				</h3>
				<p class="mt-0.5 text-xs text-muted-foreground">
					{$t('reports.stock_aging.subtitle')}
				</p>
			</div>
		</div>
		<span class="badge-neutral num-mono">{stockAgingFlat.length}</span>
	</div>

	<div class="custom-scrollbar max-h-[480px] overflow-y-auto">
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
					<tr>
						<td>
							<div class="text-sm font-medium text-foreground">
								{item.ProductName}
							</div>
							<div class="badge-neutral num-mono mt-1">
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
