<script lang="ts">
	import { fade } from 'svelte/transition';
	import type { DeadStockItem, SupplierPerformance } from '$lib/types';
	import StockAgingCard from './reports/StockAgingCard.svelte';
	import DeadStockCard from './reports/DeadStockCard.svelte';
	import SupplierPerformanceCard from './reports/SupplierPerformanceCard.svelte';

	type Props = {
		stockAgingFlat: any[];
		deadStock: DeadStockItem[];
		supplierPerf: SupplierPerformance[];
	};

	let { stockAgingFlat, deadStock, supplierPerf }: Props = $props();
</script>

<div
	in:fade={{ duration: 300, delay: 150 }}
	out:fade={{ duration: 150 }}
	class="grid gap-6 md:grid-cols-2"
>
	<StockAgingCard {stockAgingFlat} />

	<DeadStockCard {deadStock} />

	<SupplierPerformanceCard {supplierPerf} />
</div>
