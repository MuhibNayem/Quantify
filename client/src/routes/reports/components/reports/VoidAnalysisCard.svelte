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

<div class="panel p-5">
	<div class="flex items-start gap-3">
		<span
			class="flex h-9 w-9 shrink-0 items-center justify-center rounded-[calc(var(--radius)_-_4px)] border border-border bg-danger/10 text-danger"
		>
			<AlertTriangle class="h-4 w-4" />
		</span>
		<div class="min-w-0">
			<h3 class="section-title flex items-center gap-1.5">
				{$t('reports.void_analysis.title')}
				<Tooltip.Root>
					<Tooltip.Trigger>
						<HelpCircle class="h-4 w-4 text-muted-foreground transition-colors hover:text-danger" />
					</Tooltip.Trigger>
					<Tooltip.Content>
						<p class="max-w-xs font-normal text-foreground">
							{$t('reports.void_analysis.description')}
						</p>
					</Tooltip.Content>
				</Tooltip.Root>
			</h3>
			<p class="mt-0.5 text-xs text-muted-foreground">
				{$t('reports.void_analysis.subtitle')}
			</p>
		</div>
	</div>

	<div class="mt-4 space-y-2">
		{#if voidAudit.length === 0}
			<p class="py-8 text-center text-xs text-muted-foreground">
				{$t('common.no_data_available')}
			</p>
		{/if}
		{#each voidAudit as item}
			<div class="rounded-[calc(var(--radius)_-_6px)] border border-border/70 bg-muted/30 p-3.5">
				<div class="flex items-start justify-between gap-3">
					<div class="min-w-0">
						<p class="truncate text-sm font-medium text-foreground">{item.Reason}</p>
						<p class="num-mono mt-0.5 text-[11px] text-muted-foreground">
							{new Date(item.Timestamp).toLocaleString()}
						</p>
					</div>
					<span
						class="{item.RiskScore === 'High'
							? 'badge-danger'
							: item.RiskScore === 'Medium'
								? 'badge-warning'
								: 'badge-neutral'} shrink-0 tracking-wide uppercase"
					>
						{item.RiskScore === 'High'
							? $t('reports.void_analysis.risk.high')
							: item.RiskScore === 'Medium'
								? $t('reports.void_analysis.risk.medium')
								: $t('reports.void_analysis.risk.low')}
					</span>
				</div>
				<div class="mt-3 grid grid-cols-2 gap-2">
					<div
						class="rounded-[calc(var(--radius)_-_8px)] border border-border/70 bg-card px-2.5 py-2"
					>
						<p class="overline">{$t('reports.void_analysis.voids')}</p>
						<p class="num-mono mt-0.5 text-sm font-semibold text-foreground">{item.VoidCount}</p>
					</div>
					<div
						class="rounded-[calc(var(--radius)_-_8px)] border border-danger/20 bg-danger/10 px-2.5 py-2"
					>
						<p class="text-danger overline">{$t('reports.void_analysis.risk_score')}</p>
						<p class="num-mono mt-0.5 text-sm font-semibold text-danger">
							{formatCurrency(item.TotalVoidValue)}
						</p>
					</div>
				</div>
			</div>
		{/each}
	</div>
</div>
