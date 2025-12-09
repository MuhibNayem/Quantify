<script lang="ts">
	import { onMount } from 'svelte';
	import { reportsApi } from '$lib/api/resources';
	import {
		Tabs,
		TabsContent,
		TabsList,
		TabsTrigger
	} from '$lib/components/ui/tabs';
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
	let dateRange = $state({
		start: new Date(new Date().setDate(new Date().getDate() - 30)).toISOString().slice(0, 10),
		end: new Date().toISOString().slice(0, 10)
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
			const salesValues = (heatmapData as any[]).map(h => h.TotalSales || 0);
			const maxSale = Math.max(...salesValues, 1);
			const logMax = Math.log(maxSale + 1);

			for (let d = 0; d < 7; d++) {
				for (let h = 0; h < 24; h++) {
					const found = (heatmapData as any[]).find(i => i.DayOfWeek === d && i.HourOfDay === h);
					const sales = found ? found.TotalSales : 0;
					
					// Logarithmic scale: log(value) / log(max)
					// This boosts visibility of smaller values relative to massive outliers
					const intensity = sales > 0 
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
			employeeSales = Array.isArray(employeeData) ? (employeeData as any[]).map(e => ({
				name: e.Username || e.EmployeeName || 'Unknown',
				sales: e.TotalSales || 0,
				count: e.TotalOrders || e.TransactionCount || 0
			})).sort((a, b) => b.sales - a.sales) : [];

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
			console.error("Critical Report Load Error:", error);
			toast.error('Failed to load reports');
		} finally {
			loading = false;
		}
	}

	onMount(() => loadReports(false));

	const days = ['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat'];
	const hours = [0,4,8,12,16,20, 23];
</script>

<div class="relative min-h-screen overflow-hidden bg-slate-50/50 p-6 lg:p-10 font-sans selection:bg-blue-100 selection:text-blue-900">
	<!-- Dynamic Background Mesh -->
	<div class="fixed inset-0 -z-10 bg-[radial-gradient(circle_at_50%_0%,_rgba(120,119,198,0.1),transparent_50%),radial-gradient(circle_at_0%_0%,_rgba(59,130,246,0.1),transparent_50%),radial-gradient(circle_at_100%_0%,_rgba(37,99,235,0.1),transparent_50%)]"></div>
	<div class="fixed inset-0 -z-10 bg-[url('/noise.png')] opacity-[0.015] mix-blend-overlay pointer-events-none"></div>

	<div class="mx-auto max-w-7xl space-y-8 pb-12">
		<!-- Header -->
		<div class="flex flex-col gap-6 md:flex-row md:items-center md:justify-between">
			<div class="space-y-1">
				<h1 class="text-4xl font-bold tracking-tight text-slate-900 drop-shadow-sm">
					Intelligence Suite
				</h1>
				<p class="text-slate-500 font-medium text-lg">Deep insights into operational efficiency and sales performance.</p>
			</div>
			
			<div class="flex items-center gap-3 bg-white/60 p-2 rounded-[1.5rem] border border-white/40 shadow-lg shadow-slate-200/20 backdrop-blur-2xl transition-all hover:bg-white/70 hover:shadow-xl hover:scale-[1.01]">
				<div class="flex items-center gap-3 px-4 py-2 bg-gradient-to-b from-white/80 to-white/40 rounded-[1.2rem] shadow-sm border border-white/60">
					<CalendarRange class="h-4 w-4 text-blue-600" />
					<input 
						type="date" 
						bind:value={dateRange.start}
						class="text-xs font-bold text-slate-700 bg-transparent border-none focus:ring-0 p-0 cursor-pointer font-mono tracking-wide"
					/>
					<span class="text-slate-300 font-light">|</span>
					<input 
						type="date" 
						bind:value={dateRange.end}
						class="text-xs font-bold text-slate-700 bg-transparent border-none focus:ring-0 p-0 cursor-pointer font-mono tracking-wide"
					/>
				</div>
				<button 
					onclick={() => loadReports(true)}
					class="p-2.5 rounded-[1.2rem] bg-gradient-to-br from-blue-600 to-blue-700 text-white hover:from-blue-500 hover:to-blue-600 transition-all shadow-lg shadow-blue-500/25 active:scale-95 group"
					title="Refresh Data"
				>
					<Sparkles class="h-4 w-4 group-active:skew-12 transition-transform" />
				</button>
			</div>
		</div>

		<Tabs value="sales" class="space-y-8">
			<TabsList class="w-full md:w-auto h-auto justify-start gap-2 rounded-[2rem] bg-white/30 p-2 backdrop-blur-2xl border border-white/40 shadow-xl shadow-slate-200/20">
				<TabsTrigger value="sales" class="rounded-[1.5rem] px-8 py-3 text-sm font-bold data-[state=active]:bg-white data-[state=active]:text-blue-600 data-[state=active]:shadow-lg data-[state=active]:shadow-blue-900/5 transition-all duration-300 hover:bg-white/40">
					Sales & Staff
				</TabsTrigger>
				<TabsTrigger value="inventory" class="rounded-[1.5rem] px-8 py-3 text-sm font-bold data-[state=active]:bg-white data-[state=active]:text-orange-600 data-[state=active]:shadow-lg data-[state=active]:shadow-orange-900/5 transition-all duration-300 hover:bg-white/40">
					Inventory Health
				</TabsTrigger>
				<TabsTrigger value="financial" class="rounded-[1.5rem] px-8 py-3 text-sm font-bold data-[state=active]:bg-white data-[state=active]:text-emerald-600 data-[state=active]:shadow-lg data-[state=active]:shadow-emerald-900/5 transition-all duration-300 hover:bg-white/40">
					Financials
				</TabsTrigger>
			</TabsList>

			<!-- SALES CONTENT -->
			<TabsContent value="sales" class="space-y-6 outline-none">
				<div class="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
					
					<!-- Heatmap Card -->
					<div class="col-span-2 lg:col-span-2 relative overflow-hidden rounded-[2.5rem] border border-white/60 bg-white/40 p-8 shadow-2xl shadow-blue-900/5 backdrop-blur-3xl transition-all hover:bg-white/50 duration-500 group">
						<div class="absolute inset-0 bg-gradient-to-b from-white/40 to-transparent pointer-events-none"></div>
						<div class="relative z-10">
							<div class="flex items-center justify-between mb-8">
								<div class="flex items-center gap-4">
									<div class="h-12 w-12 flex items-center justify-center bg-gradient-to-br from-blue-50 to-indigo-50/50 rounded-2xl text-blue-600 shadow-inner border border-white/60">
										<Clock class="h-6 w-6" />
									</div>
									<div>
										<h3 class="text-xl font-bold text-slate-800 tracking-tight">Peak Hours</h3>
										<p class="text-sm text-slate-500 font-medium">Sales intensity analysis</p>
									</div>
								</div>
							</div>

							{#if loading}
								<Skeleton class="h-[300px] w-full rounded-[2rem] bg-white/50" />
							{:else}
								<div class="relative">
									<!-- X-Axis Labels -->
									<div class="flex justify-between pl-10 mb-3">
										{#each hours as h}
											<div class="text-[10px] font-bold text-slate-400 w-full text-left border-l border-white/20 pl-2">{h}:00</div>
										{/each}
									</div>
									
									<div class="space-y-1.5">
										{#each days as day, i}
											<div class="flex items-center gap-3">
												<span class="text-[11px] font-bold text-slate-400 w-8 text-right tracking-wide">{day.toUpperCase()}</span>
												<div class="flex-1 grid grid-cols-24 gap-[2px] p-1 bg-white/20 rounded-xl border border-white/20">
													{#each heatmapGrid.filter(c => c.day === i) as cell}
														<div 
															class="h-8 rounded-[4px] transition-all duration-300 hover:scale-[1.3] hover:shadow-lg hover:z-20 relative group/cell cursor-help"
															style="background-color: rgba(37, 99, 235, {Math.max(cell.intensity, 0.05)})"
														>
															<!-- Tooltip -->
															<div class="absolute bottom-full left-1/2 -translate-x-1/2 mb-2 hidden group-hover/cell:block z-50 pointer-events-none">
																<div class="bg-slate-900/90 text-white text-[10px] font-bold py-1.5 px-3 rounded-xl whitespace-nowrap shadow-xl backdrop-blur-md">
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
					<div class="col-span-2 lg:col-span-1 relative overflow-hidden rounded-[2.5rem] border border-white/60 bg-white/40 p-8 shadow-2xl shadow-purple-900/5 backdrop-blur-3xl transition-all hover:bg-white/50 duration-500">
						<div class="flex items-center gap-4 mb-8">
							<div class="h-12 w-12 flex items-center justify-center bg-gradient-to-br from-purple-50 to-pink-50/50 rounded-2xl text-purple-600 shadow-inner border border-white/60">
								<Users class="h-6 w-6" />
							</div>
							<div>
								<h3 class="text-xl font-bold text-slate-800 tracking-tight">Top Staff</h3>
								<p class="text-sm text-slate-500 font-medium">Revenue generators</p>
							</div>
						</div>

						<div class="space-y-3 relative z-10 max-h-[350px] overflow-y-auto custom-scrollbar pr-2">
							{#each employeeSales.slice(0, 50) as emp, i}
								<div class="group flex items-center justify-between p-4 rounded-[1.5rem] bg-white/40 border border-white/40 hover:bg-white/80 hover:scale-[1.02] hover:shadow-lg transition-all duration-300 cursor-default">
									<div class="flex items-center gap-4">
										<div class="flex h-10 w-10 items-center justify-center rounded-full bg-gradient-to-br from-slate-100 to-slate-200 font-bold text-slate-600 shadow-inner border border-white">
											{emp.name.charAt(0)}
										</div>
										<div>
											<p class="text-sm font-bold text-slate-700 group-hover:text-blue-900 transition-colors">{emp.name}</p>
											<p class="text-[11px] font-medium text-slate-400 group-hover:text-slate-500">{emp.count} transactions</p>
										</div>
									</div>
									<span class="font-bold text-purple-600 text-sm bg-purple-50/50 px-3 py-1 rounded-full border border-purple-100/50">{formatCurrency(emp.sales)}</span>
								</div>
							{/each}
						</div>
					</div>

					<!-- Top Customers (New) -->
					<div class="col-span-2 lg:col-span-1 relative overflow-hidden rounded-[2.5rem] border border-white/60 bg-white/40 p-8 shadow-2xl shadow-sky-900/5 backdrop-blur-3xl transition-all hover:bg-white/50 duration-500">
						<div class="flex items-center gap-4 mb-8">
							<div class="h-12 w-12 flex items-center justify-center bg-gradient-to-br from-sky-50 to-cyan-50/50 rounded-2xl text-sky-600 shadow-inner border border-white/60">
								<Users class="h-6 w-6" />
							</div>
							<div>
								<h3 class="text-xl font-bold text-slate-800 tracking-tight">Top Customers</h3>
								<p class="text-sm text-slate-500 font-medium">Most valuable clients</p>
							</div>
						</div>

						<div class="space-y-3 relative z-10 max-h-[350px] overflow-y-auto custom-scrollbar pr-2">
							{#each customerInsights.slice(0, 50) as cust}
								<div class="group flex items-center justify-between p-4 rounded-[1.5rem] bg-white/40 border border-white/40 hover:bg-white/80 hover:scale-[1.02] hover:shadow-lg transition-all duration-300 cursor-default">
									<div class="flex items-center gap-4">
										<div class="flex h-10 w-10 items-center justify-center rounded-full bg-gradient-to-br from-sky-100 to-sky-200 font-bold text-sky-700 shadow-inner border border-white">
											{(cust.FullName || cust.Username || '?').charAt(0)}
										</div>
										<div>
											<p class="text-sm font-bold text-slate-700 group-hover:text-sky-900 transition-colors">{cust.FullName || cust.Username}</p>
											<div class="flex items-center gap-2">
												<p class="text-[11px] font-medium text-slate-400 group-hover:text-slate-500">{cust.OrderCount} orders</p>
												{#if cust.DaysSinceLastOrder > 90}
													<span class="text-[9px] font-bold text-white bg-rose-400 px-1.5 rounded-full">At Risk</span>
												{/if}
											</div>
										</div>
									</div>
									<span class="font-bold text-sky-600 text-sm bg-sky-50/50 px-3 py-1 rounded-full border border-sky-100/50">{formatCurrency(cust.TotalSpent)}</span>
								</div>
							{/each}
						</div>
					</div>

					<!-- Category & Insights Grid -->
					<div class="col-span-full grid md:grid-cols-2 gap-6">
						<!-- Category Breakdown - Fixed Visibility -->
						<div class="rounded-[2.5rem] border border-white/60 bg-white/40 p-8 backdrop-blur-3xl shadow-2xl shadow-indigo-900/5">
							<h3 class="font-bold text-slate-700 mb-6 flex items-center gap-3 text-lg">
								<PieChart class="h-5 w-5 text-indigo-500" /> Category Breakdown
							</h3>
							
							<div class="space-y-3 max-h-[350px] overflow-y-auto custom-scrollbar pr-2">
								{#each categoryPerformance as cat}
									<!-- Added specific bg-white/60 for visibility as requested -->
									<div class="p-4 rounded-[1.5rem] bg-white/60 backdrop-blur-md border border-white/50 hover:bg-white/90 transition-all hover:scale-[1.01] hover:shadow-md space-y-3 group">
										<div class="flex justify-between items-center">
											<span class="text-base font-bold text-slate-700 group-hover:text-indigo-900 transition-colors">{cat.CategoryName}</span>
											<span class="text-[11px] font-bold px-3 py-1 rounded-full bg-indigo-50 text-indigo-600 shadow-sm">
												{cat.ItemCount || cat.TotalUnits || 0} items
											</span>
										</div>
										<div class="grid grid-cols-2 gap-4">
											<div class="bg-white/50 rounded-2xl p-3 border border-white/40">
												<div class="text-[10px] text-slate-400 uppercase tracking-wider font-bold mb-1">Sales</div>
												<div class="text-sm font-bold text-slate-800">{formatCurrency(cat.TotalSales)}</div>
											</div>
											<div class="bg-white/50 rounded-2xl p-3 border border-white/40">
												<div class="text-[10px] text-slate-400 uppercase tracking-wider font-bold mb-1">Margin</div>
												<div class="flex flex-col items-start">
													<div class="text-sm font-bold text-emerald-600">{formatCurrency(cat.GrossMargin)}</div>
													<div class="text-[10px] font-bold text-emerald-600/70 bg-emerald-50 px-1.5 rounded-md mt-0.5">
														{formatPercent(cat.MarginPercent / 100)}
													</div>
												</div>
											</div>
										</div>
									</div>
								{/each}
							</div>
						</div>

						<div class="rounded-[2.5rem] border border-white/60 bg-white/40 p-8 backdrop-blur-3xl shadow-2xl shadow-pink-900/5">
							<h3 class="font-bold text-slate-700 mb-6 flex items-center gap-3 text-lg">
								<ShoppingCart class="h-5 w-5 text-pink-500" /> Frequency Analysis
							</h3>
							<div class="space-y-3 max-h-[350px] overflow-y-auto custom-scrollbar pr-2">
								{#each basketAnalysis.slice(0, 10) as rule}
									<div class="p-4 rounded-[1.5rem] bg-pink-50/40 backdrop-blur-sm border border-pink-100/50 hover:bg-pink-50/80 transition-all hover:shadow-md hover:scale-[1.01] group">
										<div class="flex items-center justify-center gap-3 text-xs text-slate-500 mb-4 bg-white/40 p-2 rounded-xl border border-white/40">
											<span class="font-bold text-slate-700 bg-white px-3 py-1 rounded-full shadow-sm border border-slate-100">{rule.ProductAName || 'Unknown'}</span>
											<span class="text-pink-400 font-bold text-lg">+</span>
											<span class="font-bold text-slate-700 bg-white px-3 py-1 rounded-full shadow-sm border border-slate-100">{rule.ProductBName || 'Unknown'}</span>
										</div>
										<div class="flex items-center justify-between px-1">
											<span class="text-[10px] font-mono font-medium text-pink-400/80">ID: {rule.ProductA}-{rule.ProductB}</span>
											<div class="flex items-center gap-1.5">
												<span class="h-1.5 w-1.5 rounded-full bg-pink-500 animate-pulse"></span>
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
					<div class="col-span-2 relative overflow-hidden rounded-[2.5rem] border border-white/60 bg-white/40 p-0 shadow-2xl shadow-orange-900/5 backdrop-blur-3xl">
						<div class="p-8 border-b border-white/30 bg-gradient-to-r from-orange-50/40 to-transparent">
							<div class="flex items-center gap-4">
								<div class="h-12 w-12 flex items-center justify-center bg-gradient-to-br from-orange-50 to-amber-50/50 rounded-2xl text-orange-600 shadow-inner border border-white/60">
									<Package class="h-6 w-6" />
								</div>
								<div>
									<h3 class="text-xl font-bold text-slate-800 tracking-tight">Stock Aging</h3>
									<p class="text-sm text-slate-500 font-medium">Slow moving inventory > 30 days</p>
								</div>
							</div>
						</div>
						
						<div class="max-h-[500px] overflow-y-auto custom-scrollbar">
							<table class="w-full text-left text-sm border-collapse">
								<thead class="bg-white/40 text-slate-500 font-bold sticky top-0 backdrop-blur-xl border-b border-white/30 z-10">
									<tr>
										<th class="p-5 pl-8">Product / SKU</th>
										<th class="p-5 text-right">Age (Days)</th>
										<th class="p-5 text-right">Qty</th>
										<th class="p-5 text-right pr-8">Value</th>
									</tr>
								</thead>
								<tbody class="divide-y divide-white/20">
									{#each stockAgingFlat as item}
										<tr class="hover:bg-orange-50/30 transition-colors group">
											<td class="p-5 pl-8">
												<div class="font-bold text-slate-700 group-hover:text-orange-900 transition-colors">{item.ProductName}</div>
												<div class="text-[11px] text-slate-400 font-mono bg-white/60 inline-block px-1.5 rounded border border-white/40 mt-1">{item.SKU}</div>
											</td>
											<td class="p-5 text-right">
												<span class="inline-flex items-center px-2.5 py-1 rounded-lg bg-orange-100/80 text-orange-800 text-xs font-bold border border-orange-200/50 shadow-sm">
													{item.AgeDays}d
												</span>
											</td>
											<td class="p-5 text-right font-medium text-slate-600">{item.Quantity}</td>
											<td class="p-5 text-right pr-8 font-bold text-slate-800">{formatCurrency(item.Value)}</td>
										</tr>
									{/each}
								</tbody>
							</table>
						</div>
					</div>

					<!-- Dead Stock -->
					<div class="rounded-[2.5rem] border border-white/60 bg-white/40 p-8 backdrop-blur-3xl shadow-2xl shadow-red-900/5">
						<h3 class="font-bold text-slate-700 mb-6 flex items-center gap-3 text-lg">
							<AlertTriangle class="h-5 w-5 text-red-500" /> Dead Stock (180+ Days)
						</h3>
						<div class="space-y-3">
							{#each deadStock.slice(0,5) as item}
								<div class="flex justify-between items-center p-4 rounded-[1.5rem] bg-red-50/40 border border-red-100/50 hover:bg-red-50/70 transition-all hover:scale-[1.01]">
									<div class="text-sm font-bold text-slate-700 truncate max-w-[200px]">{item.ProductName}</div>
									<div class="text-right">
										<div class="font-bold text-red-600 text-sm">{formatCurrency(item.Value)}</div>
										<div class="text-[10px] font-bold text-red-400 bg-red-100/50 px-2 py-0.5 rounded-full inline-block mt-1">{item.DaysSinceLastSale} days idle</div>
									</div>
								</div>
							{/each}
						</div>
					</div>

					<!-- Supplier Performance -->
					<div class="rounded-[2.5rem] border border-white/60 bg-white/40 p-8 backdrop-blur-3xl shadow-2xl shadow-emerald-900/5">
						<h3 class="font-bold text-slate-700 mb-6 flex items-center gap-3 text-lg">
							<DollarSign class="h-5 w-5 text-emerald-500" /> Supplier Reliability
						</h3>
						<div class="space-y-3">
							{#each supplierPerf as sup}
								<div class="flex justify-between items-center p-4 rounded-[1.5rem] bg-white/60 border border-white/40 hover:bg-white/80 transition-all">
									<span class="text-sm font-bold text-slate-700">{sup.supplierName}</span>
									<div class="flex items-center gap-4">
										<div class="text-right bg-slate-50/50 rounded-xl p-2 min-w-[60px]">
											<div class="text-[9px] text-slate-400 uppercase tracking-wider font-bold">Time</div>
											<div class="font-medium text-slate-800">{sup.averageLeadTimeDays}d</div>
										</div>
										<div class="text-right bg-emerald-50/50 rounded-xl p-2 min-w-[60px]">
											<div class="text-[9px] text-emerald-600/60 uppercase tracking-wider font-bold">Rate</div>
											<div class="font-bold text-emerald-600">{formatPercent(sup.onTimeDeliveryRate)}</div>
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
				<div class="grid grid-cols-2 lg:grid-cols-4 gap-4">
					{#if gmroiStats}
						<div class="rounded-[2rem] bg-white/40 backdrop-blur-xl p-6 border border-white/50 shadow-lg shadow-emerald-900/5 hover:scale-[1.02] transition-transform">
							<p class="text-[10px] font-bold text-emerald-600 uppercase tracking-widest mb-2 bg-emerald-50 inline-block px-2 py-1 rounded-lg">Revenue</p>
							<p class="text-2xl font-bold text-slate-800">{formatCurrency(gmroiStats.TotalRevenue)}</p>
						</div>
						<div class="rounded-[2rem] bg-white/40 backdrop-blur-xl p-6 border border-white/50 shadow-lg shadow-slate-900/5 hover:scale-[1.02] transition-transform">
							<p class="text-[10px] font-bold text-slate-500 uppercase tracking-widest mb-2 bg-slate-50 inline-block px-2 py-1 rounded-lg">COGS</p>
							<p class="text-2xl font-bold text-slate-800">{formatCurrency(gmroiStats.COGS)}</p>
						</div>
						<div class="rounded-[2rem] bg-white/40 backdrop-blur-xl p-6 border border-white/50 shadow-lg shadow-indigo-900/5 hover:scale-[1.02] transition-transform">
							<p class="text-[10px] font-bold text-indigo-600 uppercase tracking-widest mb-2 bg-indigo-50 inline-block px-2 py-1 rounded-lg">Margin</p>
							<p class="text-2xl font-bold text-slate-800">{formatCurrency(gmroiStats.GrossMargin)}</p>
						</div>
						<div class="rounded-[2rem] bg-gradient-to-br from-emerald-500 to-teal-600 p-6 shadow-xl shadow-emerald-500/30 text-white relative overflow-hidden group hover:scale-[1.02] transition-transform">
							<div class="relative z-10">
								<p class="text-[10px] font-bold text-emerald-100 uppercase tracking-widest mb-2 bg-white/20 inline-block px-2 py-1 rounded-lg backdrop-blur-md">GMROI Index</p>
								<p class="text-3xl font-bold">{gmroiStats.GMROI.toFixed(2)}x</p>
								<p class="text-[10px] text-emerald-100 mt-2 font-medium">Return on Inventory</p>
							</div>
							<TrendingUp class="absolute -right-4 -bottom-4 h-24 w-24 text-white/10 group-hover:scale-110 transition-transform duration-500" />
						</div>
					{/if}
				</div>

				<div class="grid gap-6 md:grid-cols-2">
					<!-- Void Audit -->
					<div class="rounded-[2.5rem] border border-white/60 bg-white/40 p-8 backdrop-blur-3xl shadow-2xl shadow-red-900/5">
						<h3 class="font-bold text-slate-700 mb-6 flex items-center gap-3 text-lg">
							<AlertTriangle class="h-5 w-5 text-red-500" /> Void Audit
						</h3>
						<div class="space-y-3">
							{#each voidAudit.slice(0, 5) as item}
								<div class="flex justify-between items-center p-4 rounded-[1.5rem] bg-red-50/40 border border-red-100/50 backdrop-blur-sm hover:bg-red-50/70 transition-all">
									<div class="flex items-center gap-4">
										<div class="h-10 w-10 rounded-2xl bg-red-100 flex items-center justify-center text-red-600 font-bold text-sm shadow-inner">!</div>
										<div>
											<div class="text-sm font-bold text-slate-700">{item.CashierName}</div>
											<div class="text-[10px] font-medium text-slate-500 max-w-[150px] truncate bg-white/50 px-2 py-0.5 rounded-md mt-1 border border-white/50">{item.Reason}</div>
										</div>
									</div>
									<div class="font-bold text-red-600 bg-white/50 px-3 py-1 rounded-xl border border-red-100/50">{formatCurrency(item.VoidedAmount)}</div>
								</div>
							{/each}
						</div>
					</div>

					<!-- Cash Recon -->
					<div class="rounded-[2.5rem] border border-white/60 bg-white/40 p-8 backdrop-blur-3xl shadow-2xl shadow-emerald-900/5">
						<h3 class="font-bold text-slate-700 mb-6 flex items-center gap-3 text-lg">
							<DollarSign class="h-5 w-5 text-emerald-500" /> Cash Reconciliation
						</h3>
						<div class="space-y-3">
							{#each cashReconciliation as item}
								<div class="flex justify-between items-center p-4 rounded-[1.5rem] bg-white/60 border border-white/50 hover:bg-white/80 transition-all">
									<span class="text-sm font-bold text-slate-700 pl-2">{item.CashierName}</span>
									<div class="text-right">
										<div class="text-[9px] text-slate-400 uppercase font-bold tracking-wide mb-0.5">Discrepancy</div>
										<div class="font-bold px-3 py-1 rounded-xl bg-white/50 border border-slate-100 {item.Discrepancy < 0 ? 'text-red-500' : 'text-emerald-600'}">
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
