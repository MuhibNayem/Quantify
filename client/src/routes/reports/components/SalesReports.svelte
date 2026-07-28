<script lang="ts">
	import { fade } from 'svelte/transition';
	import type { CategoryPerformance, CustomerInsight, BasketAnalysisItem } from '$lib/types';
	import HeatmapCard from './reports/HeatmapCard.svelte';
	import StaffPerformanceCard from './reports/StaffPerformanceCard.svelte';
	import CustomerInsightsCard from './reports/CustomerInsightsCard.svelte';
	import CategoryBreakdownCard from './reports/CategoryBreakdownCard.svelte';
	import BasketAnalysisCard from './reports/BasketAnalysisCard.svelte';

	type Props = {
		heatmapGrid: { day: number; hour: number; sales: number; intensity: number }[];
		loading: boolean;
		employeeSales: { name: string; sales: number; count: number }[];
		customerInsights: CustomerInsight[];
		categoryPerformance: CategoryPerformance[];
		basketAnalysis: BasketAnalysisItem[];
	};

	let {
		heatmapGrid,
		loading,
		employeeSales,
		customerInsights,
		categoryPerformance,
		basketAnalysis
	}: Props = $props();
</script>

<div
	in:fade={{ duration: 300, delay: 150 }}
	out:fade={{ duration: 150 }}
	class="grid gap-6 md:grid-cols-2 lg:grid-cols-3"
>
	<HeatmapCard grid={heatmapGrid} {loading} />

	<StaffPerformanceCard {employeeSales} />

	<CustomerInsightsCard {customerInsights} />

	<!-- Category & Insights Grid -->
	<div class="col-span-full grid gap-6 md:grid-cols-2">
		<CategoryBreakdownCard {categoryPerformance} />
		<BasketAnalysisCard {basketAnalysis} />
	</div>
</div>
