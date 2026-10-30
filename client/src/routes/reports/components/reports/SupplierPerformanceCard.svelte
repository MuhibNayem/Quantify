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

<div class="panel p-5">
	<div class="flex items-start gap-3">
		<span
			class="flex h-9 w-9 shrink-0 items-center justify-center rounded-[calc(var(--radius)_-_4px)] border border-border bg-primary/10 text-primary"
		>
			<DollarSign class="h-4 w-4" />
		</span>
		<div class="min-w-0">
			<h3 class="section-title flex items-center gap-1.5">
				{$t('reports.supplier.title')}
				<Tooltip.Root>
					<Tooltip.Trigger>
						<HelpCircle
							class="h-4 w-4 text-muted-foreground transition-colors hover:text-primary"
						/>
					</Tooltip.Trigger>
					<Tooltip.Content>
						<p class="max-w-xs font-normal text-foreground">{$t('reports.supplier.description')}</p>
					</Tooltip.Content>
				</Tooltip.Root>
			</h3>
		</div>
	</div>

	<div class="mt-4 space-y-1">
		{#if supplierPerf.length === 0}
			<p class="py-8 text-center text-xs text-muted-foreground">
				{$t('common.no_data_available')}
			</p>
		{/if}
		{#each supplierPerf as sup}
			<div
				class="flex items-center justify-between gap-3 rounded-[calc(var(--radius)_-_6px)] px-2.5 py-2 transition-colors hover:bg-muted/50"
			>
				<span class="min-w-0 truncate text-sm font-medium text-foreground">{sup.supplierName}</span>
				<div class="flex shrink-0 items-center gap-2">
					<div
						class="min-w-[64px] rounded-[calc(var(--radius)_-_8px)] border border-border/70 bg-muted/50 px-2 py-1.5 text-right"
					>
						<p class="text-[10px] overline">
							{$t('reports.supplier.time')}
						</p>
						<p class="num-mono mt-0.5 text-xs font-semibold text-foreground">
							{sup.averageLeadTimeDays}{$t('reports.supplier.days_suffix')}
						</p>
					</div>
					<div
						class="min-w-[64px] rounded-[calc(var(--radius)_-_8px)] border border-success/20 bg-success/10 px-2 py-1.5 text-right"
					>
						<p class="text-[10px] text-success overline">
							{$t('reports.supplier.rate')}
						</p>
						<p class="num-mono mt-0.5 text-xs font-semibold text-success">
							{formatPercent(sup.onTimeDeliveryRate)}
						</p>
					</div>
				</div>
			</div>
		{/each}
	</div>
</div>
