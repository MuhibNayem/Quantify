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

<div class="liquid-panel rounded-lg p-6">
	<h3 class="font-display text-foreground mb-4 flex items-center gap-2 text-lg font-semibold">
		<DollarSign class="text-primary h-5 w-5" />
		{$t('reports.cash_reconciliation.title')}
		<Tooltip.Root>
			<Tooltip.Trigger>
				<HelpCircle class="text-muted-foreground hover:text-primary h-4 w-4 transition-colors" />
			</Tooltip.Trigger>
			<Tooltip.Content>
				<p class="max-w-xs">{$t('reports.cash_reconciliation.description')}</p>
			</Tooltip.Content>
		</Tooltip.Root>
	</h3>
	<div class="space-y-2">
		{#each cashReconciliation as item}
			<div
				class="border-border bg-card hover:bg-muted flex items-center justify-between rounded-md border p-4 transition-colors"
			>
				<span class="text-foreground pl-2 text-sm font-semibold">{item.CashierName}</span>
				<div class="text-right">
					<div class="overline mb-0.5">
						{$t('reports.cash_reconciliation.discrepancy')}
					</div>
					<div
						class="border-border bg-muted/50 rounded-md border px-3 py-1 font-semibold num-mono {item.Discrepancy <
						0
							? 'text-danger'
							: 'text-success'}"
					>
						{formatCurrency(item.Discrepancy)}
					</div>
				</div>
			</div>
		{/each}
	</div>
</div>
