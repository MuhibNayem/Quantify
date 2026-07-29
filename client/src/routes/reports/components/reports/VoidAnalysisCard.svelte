<script lang="ts">
	import * as Tooltip from '$lib/components/ui/tooltip';
	import { AlertTriangle, HelpCircle } from 'lucide-svelte';
	import type { VoidAuditLog } from '$lib/types';
	import { formatCurrency } from '$lib/utils';
	import { t } from '$lib/i18n';

	type Props = {
		voidAudit: VoidAuditLog[];
	};

	let { voidAudit }: Props = $props();
</script>

<div class="liquid-panel rounded-lg p-6">
	<h3 class="font-display text-foreground mb-1 flex items-center gap-2 text-lg font-semibold">
		<AlertTriangle class="text-danger h-5 w-5" />
		{$t('reports.void_analysis.title')}
		<Tooltip.Root>
			<Tooltip.Trigger>
				<HelpCircle class="text-muted-foreground hover:text-danger h-4 w-4 transition-colors" />
			</Tooltip.Trigger>
			<Tooltip.Content>
				<p class="max-w-xs">{$t('reports.void_analysis.description')}</p>
			</Tooltip.Content>
		</Tooltip.Root>
	</h3>
	<p class="text-muted-foreground mb-4 text-sm">
		{$t('reports.void_analysis.subtitle')}
	</p>

	<div class="space-y-3">
		{#each voidAudit as item}
			<div
				class="border-border bg-card rounded-md border p-5"
			>
				<div class="mb-3 flex items-start justify-between">
					<div>
						<div class="text-foreground font-semibold">{item.Reason}</div>
						<div class="text-muted-foreground text-xs num-mono">
							{new Date(item.Timestamp).toLocaleString()}
						</div>
					</div>
					<span
						class="{item.RiskScore === 'High'
							? 'badge-danger'
							: item.RiskScore === 'Medium'
								? 'badge-warning'
								: 'badge-neutral'} uppercase tracking-wider"
					>
						{item.RiskScore === 'High'
							? $t('reports.void_analysis.risk.high')
							: item.RiskScore === 'Medium'
								? $t('reports.void_analysis.risk.medium')
								: $t('reports.void_analysis.risk.low')}
					</span>
				</div>
				<div class="grid grid-cols-2 gap-3">
					<div class="bg-muted/50 rounded-md p-2.5">
						<div class="overline">
							{$t('reports.void_analysis.voids')}
						</div>
						<div class="text-foreground text-lg font-semibold num-mono">{item.VoidCount}</div>
					</div>
					<div class="bg-danger/10 rounded-md p-2.5">
						<div class="text-danger overline">
							{$t('reports.void_analysis.risk_score')}
						</div>
						<div class="text-danger text-lg font-semibold num-mono">
							{formatCurrency(item.TotalVoidValue)}
						</div>
					</div>
				</div>
			</div>
		{/each}
	</div>
</div>
