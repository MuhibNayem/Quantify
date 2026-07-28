<script lang="ts">
	import { fade } from 'svelte/transition';
	import type { VoidAuditLog, TaxLiabilityReport, CashReconciliation } from '$lib/types';
	import GmroiCards from './reports/GmroiCards.svelte';
	import VoidAnalysisCard from './reports/VoidAnalysisCard.svelte';
	import TaxLiabilityCard from './reports/TaxLiabilityCard.svelte';
	import CashReconciliationCard from './reports/CashReconciliationCard.svelte';

	type Props = {
		gmroiStats: {
			totalRevenue: number;
			totalCost: number;
			marginPercent: number;
			gmroiIndex: number;
		} | null;
		voidAudit: VoidAuditLog[];
		taxLiability: TaxLiabilityReport | null;
		cashReconciliation: CashReconciliation[];
	};

	let { gmroiStats, voidAudit, taxLiability, cashReconciliation }: Props = $props();
</script>

<div in:fade={{ duration: 300, delay: 150 }} out:fade={{ duration: 150 }} class="space-y-6">
	<GmroiCards {gmroiStats} />

	<div class="grid gap-6 md:grid-cols-2">
		<VoidAnalysisCard {voidAudit} />

		<TaxLiabilityCard {taxLiability} />

		<CashReconciliationCard {cashReconciliation} />
	</div>
</div>
