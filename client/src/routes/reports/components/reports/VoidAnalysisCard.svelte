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

<div
	class="rounded-[2.5rem] border border-white/60 bg-white/40 p-8 shadow-2xl shadow-rose-900/5 backdrop-blur-3xl"
>
	<h3 class="mb-2 flex items-center gap-2 text-lg font-bold text-slate-700">
		<AlertTriangle class="h-5 w-5 text-rose-500" />
		{$t('reports.void_analysis.title')}
		<Tooltip.Root>
			<Tooltip.Trigger>
				<HelpCircle class="h-4 w-4 text-slate-400 transition-colors hover:text-rose-500" />
			</Tooltip.Trigger>
			<Tooltip.Content>
				<p class="max-w-xs">{$t('reports.void_analysis.description')}</p>
			</Tooltip.Content>
		</Tooltip.Root>
	</h3>
	<p class="mb-6 text-sm font-medium text-slate-500">
		{$t('reports.void_analysis.subtitle')}
	</p>

	<div class="space-y-4">
		{#each voidAudit as item}
			<div
				class="relative overflow-hidden rounded-[1.8rem] border border-rose-100 bg-white/60 p-5 transition-all hover:bg-white/80"
			>
				<div class="mb-3 flex items-start justify-between">
					<div>
						<div class="font-bold text-slate-800">{item.Reason}</div>
						<div class="text-xs font-medium text-slate-400">
							{new Date(item.Timestamp).toLocaleString()}
						</div>
					</div>
					<span
						class="rounded-full px-2.5 py-1 text-[10px] font-bold tracking-wider uppercase shadow-sm
					{item.RiskScore === 'High'
							? 'bg-rose-100 text-rose-700'
							: item.RiskScore === 'Medium'
								? 'bg-amber-100 text-amber-700'
								: 'bg-slate-100 text-slate-600'}"
					>
						{item.RiskScore === 'High'
							? $t('reports.void_analysis.risk.high')
							: item.RiskScore === 'Medium'
								? $t('reports.void_analysis.risk.medium')
								: $t('reports.void_analysis.risk.low')}
					</span>
				</div>
				<div class="grid grid-cols-2 gap-4">
					<div class="rounded-xl bg-slate-50/50 p-2.5">
						<div class="text-[9px] font-bold tracking-wider text-slate-400 uppercase">
							{$t('reports.void_analysis.voids')}
						</div>
						<div class="text-lg font-black text-slate-700">{item.VoidCount}</div>
					</div>
					<div class="rounded-xl bg-rose-50/30 p-2.5">
						<div class="text-[9px] font-bold tracking-wider text-rose-400/80 uppercase">
							{$t('reports.void_analysis.risk_score')}
						</div>
						<div class="text-lg font-black text-rose-600">
							{formatCurrency(item.TotalVoidValue)}
						</div>
					</div>
				</div>
			</div>
		{/each}
	</div>
</div>
