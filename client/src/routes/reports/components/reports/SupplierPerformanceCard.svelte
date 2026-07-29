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

<div class="liquid-panel rounded-lg p-6">
	<h3 class="font-display text-foreground mb-4 flex items-center gap-2 text-lg font-semibold">
		<DollarSign class="text-primary h-5 w-5" />
		{$t('reports.supplier.title')}
		<Tooltip.Root>
			<Tooltip.Trigger>
				<HelpCircle class="text-muted-foreground hover:text-primary h-4 w-4 transition-colors" />
			</Tooltip.Trigger>
			<Tooltip.Content>
				<p class="max-w-xs">{$t('reports.supplier.description')}</p>
			</Tooltip.Content>
		</Tooltip.Root>
	</h3>
	<div class="space-y-2">
		{#each supplierPerf as sup}
			<div
				class="border-border bg-card hover:bg-muted flex items-center justify-between rounded-md border p-4 transition-colors"
			>
				<span class="text-foreground text-sm font-semibold">{sup.supplierName}</span>
				<div class="flex items-center gap-3">
					<div class="bg-muted/50 min-w-[64px] rounded-md p-2 text-right">
						<div class="overline">
							{$t('reports.supplier.time')}
						</div>
						<div class="text-foreground font-semibold num-mono">
							{sup.averageLeadTimeDays}{$t('reports.supplier.days_suffix')}
						</div>
					</div>
					<div class="bg-success/10 min-w-[64px] rounded-md p-2 text-right">
						<div class="text-success overline">
							{$t('reports.supplier.rate')}
						</div>
						<div class="text-success font-semibold num-mono">
							{formatPercent(sup.onTimeDeliveryRate)}
						</div>
					</div>
				</div>
			</div>
		{/each}
	</div>
</div>
