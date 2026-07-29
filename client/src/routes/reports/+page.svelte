<script lang="ts">
	import { onMount } from 'svelte';
	import { reportsApi } from '$lib/api/resources';
	import GlassTabs from '$lib/components/ui/glass-tabs.svelte';
	import { toast } from 'svelte-sonner';
	import * as Tooltip from '$lib/components/ui/tooltip';
	import type {
		StockAgingItem,
		GMROIReport,
		DeadStockItem,
		SupplierPerformance,
		CategoryPerformance,
		VoidAuditLog,
		TaxLiabilityReport,
		TaxLiabilityItem,
		CashReconciliation,
		CustomerInsight,
		ShrinkageReport,
		ReturnsAnalysis,
		BasketAnalysisItem
	} from '$lib/types';
	import { t } from '$lib/i18n';
	import { auth } from '$lib/stores/auth';
	import { globalConfig } from '$lib/stores/settings';
	import { goto } from '$app/navigation';
	import DateRangePicker from './components/DateRangePicker.svelte';
	import SalesReports from './components/SalesReports.svelte';
	import InventoryReports from './components/InventoryReports.svelte';
	import FinancialReports from './components/FinancialReports.svelte';

	let loading = $state(true);
	let currentTab = $state('sales');

	// Data State
	let heatmapGrid: { day: number; hour: number; sales: number; intensity: number }[] = $state([]);
	let employeeSales: any[] = $state([]);
	let categoryPerformance: CategoryPerformance[] = $state([]);
	let customerInsights: CustomerInsight[] = $state([]);
	let basketAnalysis: BasketAnalysisItem[] = $state([]);
	let stockAgingFlat: any[] = $state([]);
	let deadStock: DeadStockItem[] = $state([]);
	let supplierPerf: SupplierPerformance[] = $state([]);
	let shrinkage: ShrinkageReport[] = $state([]);
	let returnsAnalysis: ReturnsAnalysis[] = $state([]);
	let gmroiStats: any = $state(null);
	let voidAudit: VoidAuditLog[] = $state([]);
	let taxLiability: TaxLiabilityReport | null = $state(null);
	let cashReconciliation: CashReconciliation[] = $state([]);

	// Date State
	function getZonedDate(offsetDays = 0) {
		const d = new Date();
		d.setDate(d.getDate() + offsetDays);
		try {
			return new Intl.DateTimeFormat('en-CA', {
				timeZone: globalConfig.timezone,
				year: 'numeric',
				month: '2-digit',
				day: '2-digit'
			}).format(d);
		} catch (e) {
			return d.toISOString().slice(0, 10);
		}
	}

	let dateRange = $state({
		start: getZonedDate(-30),
		end: getZonedDate(0)
	});

	$effect(() => {
		if (!auth.hasPermission('reports.view')) {
			goto('/');
		}
	});

	async function loadReports(forceRefresh = false) {
		if (!auth.hasPermission('reports.view')) return;
		loading = true;

		try {
			const dateParams = {
				startDate: new Date(dateRange.start).toISOString(),
				endDate: new Date(dateRange.end).toISOString(),
				refresh: forceRefresh
			};

			const results = await Promise.allSettled([
				reportsApi.hourlyHeatmap(dateParams),
				reportsApi.stockAging(),
				reportsApi.salesByEmployee(dateParams),
				reportsApi.gmroi(dateParams),
				reportsApi.deadStock(),
				reportsApi.supplierPerformance(dateParams),
				reportsApi.categoryDrilldown(dateParams),
				reportsApi.voidAudit(dateParams),
				reportsApi.taxLiability(dateParams),
				reportsApi.cashReconciliation(dateParams),
				reportsApi.customerInsights(dateParams),
				reportsApi.shrinkage(dateParams),
				reportsApi.returnsAnalysis(dateParams),
				reportsApi.basketAnalysis(dateParams)
			]);

			const getResult = (index: number, defaultVal: any) => {
				const res = results[index];
				if (res.status === 'fulfilled') return res.value;
				console.error(`Report ${index} failed:`, res.reason);
				return defaultVal;
			};

			const heatmapData = getResult(0, []);
			const agingData = getResult(1, {});
			const employeeData = getResult(2, []);
			const gmroiRes = getResult(3, null);
			const deadStockData = getResult(4, []);
			const supplierData = getResult(5, []);
			const categoryData = getResult(6, []);
			const voidData = getResult(7, []);
			const taxData = getResult(8, null);
			const cashData = getResult(9, []);
			const customerData = getResult(10, []);
			const shrinkageData = getResult(11, []);
			const returnsData = getResult(12, []);
			const basketData = getResult(13, []);

			// 1. Transform Heatmap (Sparse -> Full Grid)
			const grid = [];
			const salesValues = (heatmapData as any[]).map((h) => h.TotalSales || 0);
			const maxSale = Math.max(...salesValues, 1);
			const logMax = Math.log(maxSale + 1);

			for (let d = 0; d < 7; d++) {
				for (let h = 0; h < 24; h++) {
					const found = (heatmapData as any[]).find((i) => i.DayOfWeek === d && i.HourOfDay === h);
					const sales = found ? found.TotalSales : 0;

					const intensity = sales > 0 ? Math.max(Math.log(sales + 1) / logMax, 0.15) : 0;

					grid.push({
						day: d,
						hour: h,
						sales,
						intensity
					});
				}
			}
			heatmapGrid = grid;

			// 2. Transform Stock Aging (Object -> Flat Array)
			const agingArray: any[] = [];
			const agingObj = agingData as any;
			if (agingObj) {
				for (const [range, items] of Object.entries(agingObj)) {
					if (Array.isArray(items)) {
						items.forEach((item: any) => {
							agingArray.push({ ...item, Range: range });
						});
					}
				}
			}
			stockAgingFlat = agingArray.sort((a, b) => b.DaysInStock - a.DaysInStock);

			// 3. Transform Employee Data (Safe Map)
			employeeSales = Array.isArray(employeeData)
				? (employeeData as any[])
						.map((e) => ({
							name: e.Username || e.EmployeeName || 'Unknown',
							sales: e.TotalSales || 0,
							count: e.TotalOrders || e.TransactionCount || 0
						}))
						.sort((a, b) => b.sales - a.sales)
				: [];

			// 4. Transform GMROI (Single Object Response)
			const gmroiObj = gmroiRes as GMROIReport | null;
			if (gmroiObj) {
				gmroiStats = {
					totalRevenue: gmroiObj.TotalRevenue || 0,
					totalCost: gmroiObj.COGS || 0,
					marginPercent:
						gmroiObj.TotalRevenue > 0 ? gmroiObj.GrossMargin / gmroiObj.TotalRevenue : 0,
					gmroiIndex: gmroiObj.GMROI || 0
				};
			} else {
				gmroiStats = null;
			}

			// Others
			deadStock = Array.isArray(deadStockData) ? deadStockData : [];
			// @ts-ignore
			supplierPerf = Array.isArray(supplierData) ? supplierData : [supplierData].filter(Boolean);
			categoryPerformance = Array.isArray(categoryData) ? categoryData : [];
			voidAudit = Array.isArray(voidData) ? voidData : [];
			// Transform Tax Liability
			const taxList = Array.isArray(taxData) ? (taxData as TaxLiabilityItem[]) : [];
			if (taxList.length > 0) {
				const totalTax = taxList.reduce((acc, item) => acc + (item.TaxAmount || 0), 0);
				taxLiability = {
					TotalTax: totalTax,
					Breakdown: taxList.map((item) => ({
						RateName: item.TaxRate > 0 ? `Tax ${item.TaxRate}%` : 'No Tax',
						Rate: item.TaxRate,
						TaxableAmount: item.TaxableAmount,
						TaxAmount: item.TaxAmount
					}))
				};
			} else {
				taxLiability = null;
			}
			cashReconciliation = Array.isArray(cashData) ? cashData : [];
			customerInsights = Array.isArray(customerData) ? customerData : [];
			shrinkage = Array.isArray(shrinkageData) ? shrinkageData : [];
			returnsAnalysis = Array.isArray(returnsData) ? returnsData : [];
			basketAnalysis = Array.isArray(basketData)
				? basketData
				: basketData?.items || basketData?.data || [];
		} catch (error) {
			console.error('Critical Report Load Error:', error);
			toast.error('Failed to load reports');
		} finally {
			loading = false;
		}
	}

	onMount(() => loadReports(false));
</script>

<Tooltip.Provider>
	<div
		class="relative min-h-screen overflow-hidden bg-slate-50/50 p-6 font-sans selection:bg-blue-100 selection:text-blue-900 lg:p-10"
	>
		<!-- Dynamic Background Mesh -->
		<div
			class="fixed inset-0 -z-10 bg-[radial-gradient(circle_at_50%_0%,_rgba(120,119,198,0.1),transparent_50%),radial-gradient(circle_at_0%_0%,_rgba(59,130,246,0.1),transparent_50%),radial-gradient(circle_at_100%_0%,_rgba(37,99,235,0.1),transparent_50%)]"
		></div>
		<div
			class="pointer-events-none fixed inset-0 -z-10 bg-[url('/noise.png')] opacity-[0.015] mix-blend-overlay"
		></div>

		<div class="mx-auto max-w-7xl space-y-8 pb-12">
			<!-- Header -->
			<div class="flex flex-col gap-6 md:flex-row md:items-center md:justify-between">
				<div class="space-y-1">
					<h1 class="text-4xl font-bold tracking-tight text-slate-900 drop-shadow-sm">
						{$t('reports.title')}
					</h1>
					<p class="text-lg font-medium text-slate-500">
						{$t('reports.subtitle')}
					</p>
				</div>

				<DateRangePicker
					bind:start={dateRange.start}
					bind:end={dateRange.end}
					onRefresh={() => loadReports(true)}
				/>
			</div>

			<div class="space-y-8">
				<GlassTabs
					bind:value={currentTab}
					tabs={[
						{ value: 'sales', label: $t('reports.tabs.sales') },
						{ value: 'inventory', label: $t('reports.tabs.inventory') },
						{ value: 'financial', label: $t('reports.tabs.financial') }
					]}
				/>

				{#if currentTab === 'sales'}
					<SalesReports
						{heatmapGrid}
						{loading}
						{employeeSales}
						{customerInsights}
						{categoryPerformance}
						{basketAnalysis}
					/>
				{:else if currentTab === 'inventory'}
					<InventoryReports {stockAgingFlat} {deadStock} {supplierPerf} />
				{:else if currentTab === 'financial'}
					<FinancialReports {gmroiStats} {voidAudit} {taxLiability} {cashReconciliation} />
				{/if}
			</div>
		</div>
	</div>
</Tooltip.Provider>
