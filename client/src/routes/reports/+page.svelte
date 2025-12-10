<script lang="ts">
	import { onMount } from 'svelte';
	import { reportsApi } from '$lib/api/resources';
	import { Tabs, TabsContent, TabsList, TabsTrigger } from '$lib/components/ui/tabs';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import { toast } from 'svelte-sonner';
	import {
		Clock,
		Users,
		PieChart,
		TrendingUp,
		AlertTriangle,
		Package,
		ShoppingCart,
		DollarSign,
		Sparkles,
		ArrowUpRight,
		ArrowDownRight,
		CalendarRange
	} from 'lucide-svelte';
	import type {
		HourlySalesHeatmap,
		StockAgingItem,
		EmployeeSalesPerformance,
		GMROIReport,
		DeadStockItem,
		SupplierPerformance,
		CategoryPerformance,
		VoidAuditLog,
		TaxLiabilityReport,
		CashReconciliation,
		CustomerInsight,
		ShrinkageReport,
		ReturnsAnalysis,
		BasketAnalysisRule
	} from '$lib/types';
	import { formatCurrency, formatPercent, cn } from '$lib/utils';
	import { auth } from '$lib/stores/auth';
	import { globalConfig } from '$lib/stores/settings';
	import { goto } from '$app/navigation';
	import { fly, fade } from 'svelte/transition';

	let loading = $state(true);

	// Data State
	let heatmapGrid: { day: number; hour: number; sales: number; intensity: number }[] = $state([]);
	let employeeSales: any[] = $state([]); // Transformed
	let categoryPerformance: CategoryPerformance[] = $state([]);
	let customerInsights: CustomerInsight[] = $state([]);
	let basketAnalysis: BasketAnalysisRule[] = $state([]);
	let stockAgingFlat: any[] = $state([]); // Transformed
	let deadStock: DeadStockItem[] = $state([]);
	let supplierPerf: SupplierPerformance[] = $state([]);
	let shrinkage: ShrinkageReport[] = $state([]);
	let returnsAnalysis: ReturnsAnalysis[] = $state([]);
	let gmroiStats: any = $state(null); // Transformed to object
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

			// Helper to extract data or return default
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

					// Logarithmic scale: log(value) / log(max)
					// This boosts visibility of smaller values relative to massive outliers
					const intensity =
						sales > 0
							? Math.max(Math.log(sales + 1) / logMax, 0.15) // Min 15% opacity for visibility
							: 0;

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
			const agingArray = [];
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

			// 4. Transform GMROI (Single Object)
			gmroiStats = gmroiRes;

			// Others
			deadStock = Array.isArray(deadStockData) ? deadStockData : [];
			// @ts-ignore
			supplierPerf = Array.isArray(supplierData) ? supplierData : [supplierData].filter(Boolean);
			categoryPerformance = Array.isArray(categoryData) ? categoryData : [];
			voidAudit = Array.isArray(voidData) ? voidData : [];
			taxLiability = taxData;
			cashReconciliation = Array.isArray(cashData) ? cashData : [];
			customerInsights = Array.isArray(customerData) ? customerData : [];
			shrinkage = Array.isArray(shrinkageData) ? shrinkageData : [];
			returnsAnalysis = Array.isArray(returnsData) ? returnsData : [];
			basketAnalysis = Array.isArray(basketData) ? basketData : [];
		} catch (error) {
			console.error('Critical Report Load Error:', error);
			toast.error('Failed to load reports');
		} finally {
			loading = false;
		}
	}

	onMount(() => loadReports(false));

	const days = ['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat'];
	const hours = [0, 4, 8, 12, 16, 20, 23];
</script>

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
					Intelligence Suite
				</h1>
				<p class="text-lg font-medium text-slate-500">
					Deep insights into operational efficiency and sales performance.
				</p>
			</div>

			<div
				class="flex items-center gap-3 rounded-[1.5rem] border border-white/40 bg-white/60 p-2 shadow-lg shadow-slate-200/20 backdrop-blur-2xl transition-all hover:scale-[1.01] hover:bg-white/70 hover:shadow-xl"
			>
				<div
					class="flex items-center gap-3 rounded-[1.2rem] border border-white/60 bg-gradient-to-b from-white/80 to-white/40 px-4 py-2 shadow-sm"
				>
					<CalendarRange class="h-4 w-4 text-blue-600" />
					<input
						type="date"
						bind:value={dateRange.start}
						class="cursor-pointer border-none bg-transparent p-0 font-mono text-xs font-bold tracking-wide text-slate-700 focus:ring-0"
					/>
					<span class="font-light text-slate-300">|</span>
					<input
						type="date"
						bind:value={dateRange.end}
						class="cursor-pointer border-none bg-transparent p-0 font-mono text-xs font-bold tracking-wide text-slate-700 focus:ring-0"
					/>
				</div>
				<button
					onclick={() => loadReports(true)}
					class="group rounded-[1.2rem] bg-gradient-to-br from-blue-600 to-blue-700 p-2.5 text-white shadow-lg shadow-blue-500/25 transition-all hover:from-blue-500 hover:to-blue-600 active:scale-95"
					title="Refresh Data"
				>
					<Sparkles class="group-active:skew-12 h-4 w-4 transition-transform" />
				</button>
			</div>
		</div>

		<Tabs value="sales" class="space-y-8">
			<TabsList
				class="h-auto w-full justify-start gap-2 rounded-[2rem] border border-white/40 bg-white/30 p-2 shadow-xl shadow-slate-200/20 backdrop-blur-2xl md:w-auto"
			>
				<TabsTrigger
					value="sales"
					class="rounded-[1.5rem] px-8 py-3 text-sm font-bold transition-all duration-300 hover:bg-white/40 data-[state=active]:bg-white data-[state=active]:text-blue-600 data-[state=active]:shadow-lg data-[state=active]:shadow-blue-900/5"
				>
					Sales & Staff
				</TabsTrigger>
				<TabsTrigger
					value="inventory"
					class="rounded-[1.5rem] px-8 py-3 text-sm font-bold transition-all duration-300 hover:bg-white/40 data-[state=active]:bg-white data-[state=active]:text-orange-600 data-[state=active]:shadow-lg data-[state=active]:shadow-orange-900/5"
				>
					Inventory Health
				</TabsTrigger>
				<TabsTrigger
					value="financial"
					class="rounded-[1.5rem] px-8 py-3 text-sm font-bold transition-all duration-300 hover:bg-white/40 data-[state=active]:bg-white data-[state=active]:text-emerald-600 data-[state=active]:shadow-lg data-[state=active]:shadow-emerald-900/5"
				>
					Financials
				</TabsTrigger>
			</TabsList>

			<!-- SALES CONTENT -->
			<TabsContent value="sales" class="space-y-6 outline-none">
				<div class="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
					<!-- Heatmap Card -->
					<div
						class="group relative col-span-2 overflow-hidden rounded-[2.5rem] border border-white/60 bg-white/40 p-8 shadow-2xl shadow-blue-900/5 backdrop-blur-3xl transition-all duration-500 hover:bg-white/50 lg:col-span-2"
					>
						<div
							class="pointer-events-none absolute inset-0 bg-gradient-to-b from-white/40 to-transparent"
						></div>
						<div class="relative z-10">
							<div class="mb-8 flex items-center justify-between">
								<div class="flex items-center gap-4">
									<div
										class="flex h-12 w-12 items-center justify-center rounded-2xl border border-white/60 bg-gradient-to-br from-blue-50 to-indigo-50/50 text-blue-600 shadow-inner"
									>
										<Clock class="h-6 w-6" />
									</div>
									<div>
										<h3 class="text-xl font-bold tracking-tight text-slate-800">Peak Hours</h3>
										<p class="text-sm font-medium text-slate-500">Sales intensity analysis</p>
									</div>
								</div>
							</div>

							{#if loading}
								<Skeleton class="h-[300px] w-full rounded-[2rem] bg-white/50" />
							{:else}
								<div class="relative">
									<!-- X-Axis Labels -->
									<div class="mb-3 flex justify-between pl-10">
										{#each hours as h}
											<div
												class="w-full border-l border-white/20 pl-2 text-left text-[10px] font-bold text-slate-400"
											>
												{h}:00
											</div>
										{/each}
									</div>

									<div class="space-y-1.5">
										{#each days as day, i}
											<div class="flex items-center gap-3">
												<span
													class="w-8 text-right text-[11px] font-bold tracking-wide text-slate-400"
													>{day.toUpperCase()}</span
												>
												<div
													class="grid-cols-24 grid flex-1 gap-[2px] rounded-xl border border-white/20 bg-white/20 p-1"
												>
													{#each heatmapGrid.filter((c) => c.day === i) as cell}
														<div
															class="group/cell relative h-8 cursor-help rounded-[4px] transition-all duration-300 hover:z-20 hover:scale-[1.3] hover:shadow-lg"
															style="background-color: rgba(37, 99, 235, {Math.max(
																cell.intensity,
																0.05
															)})"
														>
															<!-- Tooltip -->
															<div
																class="pointer-events-none absolute bottom-full left-1/2 z-50 mb-2 hidden -translate-x-1/2 group-hover/cell:block"
															>
																<div
																	class="whitespace-nowrap rounded-xl bg-slate-900/90 px-3 py-1.5 text-[10px] font-bold text-white shadow-xl backdrop-blur-md"
																>
																	{formatCurrency(cell.sales)}
																</div>
															</div>
														</div>
													{/each}
												</div>
											</div>
										{/each}
									</div>
								</div>
							{/if}
						</div>
					</div>

					<!-- Staff Performance -->
					<div
						class="relative col-span-2 overflow-hidden rounded-[2.5rem] border border-white/60 bg-white/40 p-8 shadow-2xl shadow-purple-900/5 backdrop-blur-3xl transition-all duration-500 hover:bg-white/50 lg:col-span-1"
					>
						<div class="mb-8 flex items-center gap-4">
							<div
								class="flex h-12 w-12 items-center justify-center rounded-2xl border border-white/60 bg-gradient-to-br from-purple-50 to-pink-50/50 text-purple-600 shadow-inner"
							>
								<Users class="h-6 w-6" />
							</div>
							<div>
								<h3 class="text-xl font-bold tracking-tight text-slate-800">Top Staff</h3>
								<p class="text-sm font-medium text-slate-500">Revenue generators</p>
							</div>
						</div>

						<div
							class="custom-scrollbar relative z-10 max-h-[350px] space-y-3 overflow-y-auto pr-2"
						>
							{#each employeeSales.slice(0, 50) as emp, i}
								<div
									class="group flex cursor-default items-center justify-between rounded-[1.5rem] border border-white/40 bg-white/40 p-4 transition-all duration-300 hover:scale-[1.02] hover:bg-white/80 hover:shadow-lg"
								>
									<div class="flex items-center gap-4">
										<div
											class="flex h-10 w-10 items-center justify-center rounded-full border border-white bg-gradient-to-br from-slate-100 to-slate-200 font-bold text-slate-600 shadow-inner"
										>
											{emp.name.charAt(0)}
										</div>
										<div>
											<p
												class="text-sm font-bold text-slate-700 transition-colors group-hover:text-blue-900"
											>
												{emp.name}
											</p>
											<p class="text-[11px] font-medium text-slate-400 group-hover:text-slate-500">
												{emp.count} transactions
											</p>
										</div>
									</div>
									<span
										class="rounded-full border border-purple-100/50 bg-purple-50/50 px-3 py-1 text-sm font-bold text-purple-600"
										>{formatCurrency(emp.sales)}</span
									>
								</div>
							{/each}
						</div>
					</div>

					<!-- Top Customers (New) -->
					<div
						class="relative col-span-2 overflow-hidden rounded-[2.5rem] border border-white/60 bg-white/40 p-8 shadow-2xl shadow-sky-900/5 backdrop-blur-3xl transition-all duration-500 hover:bg-white/50 lg:col-span-1"
					>
						<div class="mb-8 flex items-center gap-4">
							<div
								class="flex h-12 w-12 items-center justify-center rounded-2xl border border-white/60 bg-gradient-to-br from-sky-50 to-cyan-50/50 text-sky-600 shadow-inner"
							>
								<Users class="h-6 w-6" />
							</div>
							<div>
								<h3 class="text-xl font-bold tracking-tight text-slate-800">Top Customers</h3>
								<p class="text-sm font-medium text-slate-500">Most valuable clients</p>
							</div>
						</div>

						<div
							class="custom-scrollbar relative z-10 max-h-[350px] space-y-3 overflow-y-auto pr-2"
						>
							{#each customerInsights.slice(0, 50) as cust}
								<div
									class="group flex cursor-default items-center justify-between rounded-[1.5rem] border border-white/40 bg-white/40 p-4 transition-all duration-300 hover:scale-[1.02] hover:bg-white/80 hover:shadow-lg"
								>
									<div class="flex items-center gap-4">
										<div
											class="flex h-10 w-10 items-center justify-center rounded-full border border-white bg-gradient-to-br from-sky-100 to-sky-200 font-bold text-sky-700 shadow-inner"
										>
											{(cust.FullName || cust.Username || '?').charAt(0)}
										</div>
										<div>
											<p
												class="text-sm font-bold text-slate-700 transition-colors group-hover:text-sky-900"
											>
												{cust.FullName || cust.Username}
											</p>
											<div class="flex items-center gap-2">
												<p
													class="text-[11px] font-medium text-slate-400 group-hover:text-slate-500"
												>
													{cust.OrderCount} orders
												</p>
												{#if cust.DaysSinceLastOrder > 90}
													<span
														class="rounded-full bg-rose-400 px-1.5 text-[9px] font-bold text-white"
														>At Risk</span
													>
												{/if}
											</div>
										</div>
									</div>
									<span
										class="rounded-full border border-sky-100/50 bg-sky-50/50 px-3 py-1 text-sm font-bold text-sky-600"
										>{formatCurrency(cust.TotalSpent)}</span
									>
								</div>
							{/each}
						</div>
					</div>

					<!-- Category & Insights Grid -->
					<div class="col-span-full grid gap-6 md:grid-cols-2">
						<!-- Category Breakdown - Fixed Visibility -->
						<div
							class="rounded-[2.5rem] border border-white/60 bg-white/40 p-8 shadow-2xl shadow-indigo-900/5 backdrop-blur-3xl"
						>
							<h3 class="mb-6 flex items-center gap-3 text-lg font-bold text-slate-700">
								<PieChart class="h-5 w-5 text-indigo-500" /> Category Breakdown
							</h3>

							<div class="custom-scrollbar max-h-[350px] space-y-3 overflow-y-auto pr-2">
								{#each categoryPerformance as cat}
									<!-- Added specific bg-white/60 for visibility as requested -->
									<div
										class="group space-y-3 rounded-[1.5rem] border border-white/50 bg-white/60 p-4 backdrop-blur-md transition-all hover:scale-[1.01] hover:bg-white/90 hover:shadow-md"
									>
										<div class="flex items-center justify-between">
											<span
												class="text-base font-bold text-slate-700 transition-colors group-hover:text-indigo-900"
												>{cat.CategoryName}</span
											>
											<span
												class="rounded-full bg-indigo-50 px-3 py-1 text-[11px] font-bold text-indigo-600 shadow-sm"
											>
												{cat.ItemCount || cat.TotalUnits || 0} items
											</span>
										</div>
										<div class="grid grid-cols-2 gap-4">
											<div class="rounded-2xl border border-white/40 bg-white/50 p-3">
												<div
													class="mb-1 text-[10px] font-bold uppercase tracking-wider text-slate-400"
												>
													Sales
												</div>
												<div class="text-sm font-bold text-slate-800">
													{formatCurrency(cat.TotalSales)}
												</div>
											</div>
											<div class="rounded-2xl border border-white/40 bg-white/50 p-3">
												<div
													class="mb-1 text-[10px] font-bold uppercase tracking-wider text-slate-400"
												>
													Margin
												</div>
												<div class="flex flex-col items-start">
													<div class="text-sm font-bold text-emerald-600">
														{formatCurrency(cat.GrossMargin)}
													</div>
													<div
														class="mt-0.5 rounded-md bg-emerald-50 px-1.5 text-[10px] font-bold text-emerald-600/70"
													>
														{formatPercent(cat.MarginPercent / 100)}
													</div>
												</div>
											</div>
										</div>
									</div>
								{/each}
							</div>
						</div>

						<div
							class="rounded-[2.5rem] border border-white/60 bg-white/40 p-8 shadow-2xl shadow-pink-900/5 backdrop-blur-3xl"
						>
							<h3 class="mb-6 flex items-center gap-3 text-lg font-bold text-slate-700">
								<ShoppingCart class="h-5 w-5 text-pink-500" /> Frequency Analysis
							</h3>
							<div class="custom-scrollbar max-h-[350px] space-y-3 overflow-y-auto pr-2">
								{#each basketAnalysis.slice(0, 10) as rule}
									<div
										class="group rounded-[1.5rem] border border-pink-100/50 bg-pink-50/40 p-4 backdrop-blur-sm transition-all hover:scale-[1.01] hover:bg-pink-50/80 hover:shadow-md"
									>
										<div
											class="mb-4 flex items-center justify-center gap-3 rounded-xl border border-white/40 bg-white/40 p-2 text-xs text-slate-500"
										>
											<span
												class="rounded-full border border-slate-100 bg-white px-3 py-1 font-bold text-slate-700 shadow-sm"
												>{rule.ProductAName || 'Unknown'}</span
											>
											<span class="text-lg font-bold text-pink-400">+</span>
											<span
												class="rounded-full border border-slate-100 bg-white px-3 py-1 font-bold text-slate-700 shadow-sm"
												>{rule.ProductBName || 'Unknown'}</span
											>
										</div>
										<div class="flex items-center justify-between px-1">
											<span class="font-mono text-[10px] font-medium text-pink-400/80"
												>ID: {rule.ProductA}-{rule.ProductB}</span
											>
											<div class="flex items-center gap-1.5">
												<span class="h-1.5 w-1.5 animate-pulse rounded-full bg-pink-500"></span>
												<span class="text-xs font-bold text-pink-700">
													{rule.Frequency} Orders
												</span>
											</div>
										</div>
									</div>
								{/each}
							</div>
						</div>
					</div>
				</div>
			</TabsContent>

			<!-- INVENTORY CONTENT -->
			<TabsContent value="inventory" class="space-y-6 outline-none">
				<div class="grid gap-6 md:grid-cols-2">
					<!-- Stock Aging Table -->
					<div
						class="relative col-span-2 overflow-hidden rounded-[2.5rem] border border-white/60 bg-white/40 p-0 shadow-2xl shadow-orange-900/5 backdrop-blur-3xl"
					>
						<div
							class="border-b border-white/30 bg-gradient-to-r from-orange-50/40 to-transparent p-8"
						>
							<div class="flex items-center gap-4">
								<div
									class="flex h-12 w-12 items-center justify-center rounded-2xl border border-white/60 bg-gradient-to-br from-orange-50 to-amber-50/50 text-orange-600 shadow-inner"
								>
									<Package class="h-6 w-6" />
								</div>
								<div>
									<h3 class="text-xl font-bold tracking-tight text-slate-800">Stock Aging</h3>
									<p class="text-sm font-medium text-slate-500">Slow moving inventory > 30 days</p>
								</div>
							</div>
						</div>

						<div class="custom-scrollbar max-h-[500px] overflow-y-auto">
							<table class="w-full border-collapse text-left text-sm">
								<thead
									class="sticky top-0 z-10 border-b border-white/30 bg-white/40 font-bold text-slate-500 backdrop-blur-xl"
								>
									<tr>
										<th class="p-5 pl-8">Product / SKU</th>
										<th class="p-5 text-right">Age (Days)</th>
										<th class="p-5 text-right">Qty</th>
										<th class="p-5 pr-8 text-right">Value</th>
									</tr>
								</thead>
								<tbody class="divide-y divide-white/20">
									{#each stockAgingFlat as item}
										<tr class="group transition-colors hover:bg-orange-50/30">
											<td class="p-5 pl-8">
												<div
													class="font-bold text-slate-700 transition-colors group-hover:text-orange-900"
												>
													{item.ProductName}
												</div>
												<div
													class="mt-1 inline-block rounded border border-white/40 bg-white/60 px-1.5 font-mono text-[11px] text-slate-400"
												>
													{item.SKU}
												</div>
											</td>
											<td class="p-5 text-right">
												<span
													class="inline-flex items-center rounded-lg border border-orange-200/50 bg-orange-100/80 px-2.5 py-1 text-xs font-bold text-orange-800 shadow-sm"
												>
													{item.AgeDays}d
												</span>
											</td>
											<td class="p-5 text-right font-medium text-slate-600">{item.Quantity}</td>
											<td class="p-5 pr-8 text-right font-bold text-slate-800"
												>{formatCurrency(item.Value)}</td
											>
										</tr>
									{/each}
								</tbody>
							</table>
						</div>
					</div>

					<!-- Dead Stock -->
					<div
						class="rounded-[2.5rem] border border-white/60 bg-white/40 p-8 shadow-2xl shadow-red-900/5 backdrop-blur-3xl"
					>
						<h3 class="mb-6 flex items-center gap-3 text-lg font-bold text-slate-700">
							<AlertTriangle class="h-5 w-5 text-red-500" /> Dead Stock (180+ Days)
						</h3>
						<div class="space-y-3">
							{#each deadStock.slice(0, 5) as item}
								<div
									class="flex items-center justify-between rounded-[1.5rem] border border-red-100/50 bg-red-50/40 p-4 transition-all hover:scale-[1.01] hover:bg-red-50/70"
								>
									<div class="max-w-[200px] truncate text-sm font-bold text-slate-700">
										{item.ProductName}
									</div>
									<div class="text-right">
										<div class="text-sm font-bold text-red-600">{formatCurrency(item.Value)}</div>
										<div
											class="mt-1 inline-block rounded-full bg-red-100/50 px-2 py-0.5 text-[10px] font-bold text-red-400"
										>
											{item.DaysSinceLastSale} days idle
										</div>
									</div>
								</div>
							{/each}
						</div>
					</div>

					<!-- Supplier Performance -->
					<div
						class="rounded-[2.5rem] border border-white/60 bg-white/40 p-8 shadow-2xl shadow-emerald-900/5 backdrop-blur-3xl"
					>
						<h3 class="mb-6 flex items-center gap-3 text-lg font-bold text-slate-700">
							<DollarSign class="h-5 w-5 text-emerald-500" /> Supplier Reliability
						</h3>
						<div class="space-y-3">
							{#each supplierPerf as sup}
								<div
									class="flex items-center justify-between rounded-[1.5rem] border border-white/40 bg-white/60 p-4 transition-all hover:bg-white/80"
								>
									<span class="text-sm font-bold text-slate-700">{sup.supplierName}</span>
									<div class="flex items-center gap-4">
										<div class="min-w-[60px] rounded-xl bg-slate-50/50 p-2 text-right">
											<div class="text-[9px] font-bold uppercase tracking-wider text-slate-400">
												Time
											</div>
											<div class="font-medium text-slate-800">{sup.averageLeadTimeDays}d</div>
										</div>
										<div class="min-w-[60px] rounded-xl bg-emerald-50/50 p-2 text-right">
											<div
												class="text-[9px] font-bold uppercase tracking-wider text-emerald-600/60"
											>
												Rate
											</div>
											<div class="font-bold text-emerald-600">
												{formatPercent(sup.onTimeDeliveryRate)}
											</div>
										</div>
									</div>
								</div>
							{/each}
						</div>
					</div>
				</div>
			</TabsContent>

			<!-- FINANCIAL CONTENT -->
			<TabsContent value="financial" class="space-y-6 outline-none">
				<!-- GMROI Cards -->
				<div class="grid grid-cols-2 gap-4 lg:grid-cols-4">
					{#if gmroiStats}
						<div
							class="rounded-[2rem] border border-white/50 bg-white/40 p-6 shadow-lg shadow-emerald-900/5 backdrop-blur-xl transition-transform hover:scale-[1.02]"
						>
							<p
								class="mb-2 inline-block rounded-lg bg-emerald-50 px-2 py-1 text-[10px] font-bold uppercase tracking-widest text-emerald-600"
							>
								Revenue
							</p>
							<p class="text-2xl font-bold text-slate-800">
								{formatCurrency(gmroiStats.TotalRevenue)}
							</p>
						</div>
						<div
							class="rounded-[2rem] border border-white/50 bg-white/40 p-6 shadow-lg shadow-slate-900/5 backdrop-blur-xl transition-transform hover:scale-[1.02]"
						>
							<p
								class="mb-2 inline-block rounded-lg bg-slate-50 px-2 py-1 text-[10px] font-bold uppercase tracking-widest text-slate-500"
							>
								COGS
							</p>
							<p class="text-2xl font-bold text-slate-800">{formatCurrency(gmroiStats.COGS)}</p>
						</div>
						<div
							class="rounded-[2rem] border border-white/50 bg-white/40 p-6 shadow-lg shadow-indigo-900/5 backdrop-blur-xl transition-transform hover:scale-[1.02]"
						>
							<p
								class="mb-2 inline-block rounded-lg bg-indigo-50 px-2 py-1 text-[10px] font-bold uppercase tracking-widest text-indigo-600"
							>
								Margin
							</p>
							<p class="text-2xl font-bold text-slate-800">
								{formatCurrency(gmroiStats.GrossMargin)}
							</p>
						</div>
						<div
							class="group relative overflow-hidden rounded-[2rem] bg-gradient-to-br from-emerald-500 to-teal-600 p-6 text-white shadow-xl shadow-emerald-500/30 transition-transform hover:scale-[1.02]"
						>
							<div class="relative z-10">
								<p
									class="mb-2 inline-block rounded-lg bg-white/20 px-2 py-1 text-[10px] font-bold uppercase tracking-widest text-emerald-100 backdrop-blur-md"
								>
									GMROI Index
								</p>
								<p class="text-3xl font-bold">{gmroiStats.GMROI.toFixed(2)}x</p>
								<p class="mt-2 text-[10px] font-medium text-emerald-100">Return on Inventory</p>
							</div>
							<TrendingUp
								class="absolute -bottom-4 -right-4 h-24 w-24 text-white/10 transition-transform duration-500 group-hover:scale-110"
							/>
						</div>
					{/if}
				</div>

				<div class="grid gap-6 md:grid-cols-2">
					<!-- Void Audit -->
					<div
						class="rounded-[2.5rem] border border-white/60 bg-white/40 p-8 shadow-2xl shadow-red-900/5 backdrop-blur-3xl"
					>
						<h3 class="mb-6 flex items-center gap-3 text-lg font-bold text-slate-700">
							<AlertTriangle class="h-5 w-5 text-red-500" /> Void Audit
						</h3>
						<div class="space-y-3">
							{#each voidAudit.slice(0, 5) as item}
								<div
									class="flex items-center justify-between rounded-[1.5rem] border border-red-100/50 bg-red-50/40 p-4 backdrop-blur-sm transition-all hover:bg-red-50/70"
								>
									<div class="flex items-center gap-4">
										<div
											class="flex h-10 w-10 items-center justify-center rounded-2xl bg-red-100 text-sm font-bold text-red-600 shadow-inner"
										>
											!
										</div>
										<div>
											<div class="text-sm font-bold text-slate-700">{item.CashierName}</div>
											<div
												class="mt-1 max-w-[150px] truncate rounded-md border border-white/50 bg-white/50 px-2 py-0.5 text-[10px] font-medium text-slate-500"
											>
												{item.Reason}
											</div>
										</div>
									</div>
									<div
										class="rounded-xl border border-red-100/50 bg-white/50 px-3 py-1 font-bold text-red-600"
									>
										{formatCurrency(item.VoidedAmount)}
									</div>
								</div>
							{/each}
						</div>
					</div>

					<!-- Cash Recon -->
					<div
						class="rounded-[2.5rem] border border-white/60 bg-white/40 p-8 shadow-2xl shadow-emerald-900/5 backdrop-blur-3xl"
					>
						<h3 class="mb-6 flex items-center gap-3 text-lg font-bold text-slate-700">
							<DollarSign class="h-5 w-5 text-emerald-500" /> Cash Reconciliation
						</h3>
						<div class="space-y-3">
							{#each cashReconciliation as item}
								<div
									class="flex items-center justify-between rounded-[1.5rem] border border-white/50 bg-white/60 p-4 transition-all hover:bg-white/80"
								>
									<span class="pl-2 text-sm font-bold text-slate-700">{item.CashierName}</span>
									<div class="text-right">
										<div class="mb-0.5 text-[9px] font-bold uppercase tracking-wide text-slate-400">
											Discrepancy
										</div>
										<div
											class="rounded-xl border border-slate-100 bg-white/50 px-3 py-1 font-bold {item.Discrepancy <
											0
												? 'text-red-500'
												: 'text-emerald-600'}"
										>
											{formatCurrency(item.Discrepancy)}
										</div>
									</div>
								</div>
							{/each}
						</div>
					</div>
				</div>
			</TabsContent>
		</Tabs>
	</div>
</div>

<style>
	.grid-cols-24 {
		grid-template-columns: repeat(24, minmax(0, 1fr));
	}
	.custom-scrollbar::-webkit-scrollbar {
		width: 4px;
	}
	.custom-scrollbar::-webkit-scrollbar-track {
		background: rgba(255, 255, 255, 0.1);
	}
	.custom-scrollbar::-webkit-scrollbar-thumb {
		background: rgba(203, 213, 225, 0.4);
		border-radius: 10px;
	}
	.custom-scrollbar::-webkit-scrollbar-thumb:hover {
		background: rgba(148, 163, 184, 0.6);
	}
</style>
