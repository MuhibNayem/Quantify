<script lang="ts">
	import { onMount } from 'svelte';
	import { reportsApi } from '$lib/api/resources';
	import {
		Card,
		CardContent,
		CardDescription,
		CardHeader,
		CardTitle
	} from '$lib/components/ui/card';
	import { Tabs, TabsContent, TabsList, TabsTrigger } from '$lib/components/ui/tabs';
	import {
		Table,
		TableBody,
		TableCell,
		TableHead,
		TableHeader,
		TableRow
	} from '$lib/components/ui/table';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import { toast } from 'svelte-sonner';
	import {
		BarChart3,
		TrendingUp,
		DollarSign,
		Package,
		Users,
		AlertTriangle,
		Clock,
		PieChart,
		ShoppingCart,
		Undo2,
		FileWarning,
		Scale
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
	import { formatCurrency, formatPercent } from '$lib/utils';
	import { auth } from '$lib/stores/auth';
	import { goto } from '$app/navigation';

	let loading = true;
	// Sales & Staff
	let heatmap: HourlySalesHeatmap[] = [];
	let employeeSales: EmployeeSalesPerformance[] = [];
	let categoryPerformance: CategoryPerformance[] = [];
	let customerInsights: CustomerInsight[] = [];
	let basketAnalysis: BasketAnalysisRule[] = [];

	// Inventory Health
	let stockAging: StockAgingItem[] = [];
	let deadStock: DeadStockItem[] = [];
	let supplierPerf: SupplierPerformance[] = []; // Note: API returns array? Type says single obj but list usually returns array. Assuming array for report.
	let shrinkage: ShrinkageReport[] = [];
	let returnsAnalysis: ReturnsAnalysis[] = [];

	// Financials
	let gmroiData: GMROIReport[] = [];
	let voidAudit: VoidAuditLog[] = [];
	let taxLiability: TaxLiabilityReport | null = null;
	let cashReconciliation: CashReconciliation[] = [];

	$effect(() => {
		if (!auth.hasPermission('reports.view')) {
			toast.error('Access Denied', {
				description: 'You do not have permission to view reports.'
			});
			goto('/');
		}
	});

	const loadReports = async () => {
		if (!auth.hasPermission('reports.view')) return;
		
		loading = true;
		try {
			const [
				heatmapData,
				agingData,
				employeeData,
				gmroiRes,
				deadStockData,
				supplierData,
				categoryData,
				voidData,
				taxData,
				cashData,
				customerData,
				shrinkageData,
				returnsData,
				basketData
			] = await Promise.all([
				reportsApi.hourlyHeatmap(),
				reportsApi.stockAging(),
				reportsApi.salesByEmployee(),
				reportsApi.gmroi(),
				reportsApi.deadStock(),
				reportsApi.supplierPerformance(),
				reportsApi.categoryDrilldown(),
				reportsApi.voidAudit(),
				reportsApi.taxLiability(),
				reportsApi.cashReconciliation(),
				reportsApi.customerInsights(),
				reportsApi.shrinkage(),
				reportsApi.returnsAnalysis(),
				reportsApi.basketAnalysis()
			]);

			heatmap = heatmapData;
			stockAging = agingData;
			employeeSales = employeeData;
			gmroiData = gmroiRes;
			deadStock = deadStockData;
			// @ts-ignore - API might return array or single object, handling as array for list
			supplierPerf = Array.isArray(supplierData) ? supplierData : [supplierData]; 
			categoryPerformance = categoryData;
			voidAudit = voidData;
			taxLiability = taxData;
			cashReconciliation = cashData;
			customerInsights = customerData;
			shrinkage = shrinkageData;
			returnsAnalysis = returnsData;
			basketAnalysis = basketData;

		} catch (error) {
			console.error(error);
			toast.error('Failed to load reports');
		} finally {
			loading = false;
		}
	};

	onMount(loadReports);
</script>

<div class="container mx-auto py-8">
	<div class="mb-8 flex items-center justify-between">
		<div>
			<h1 class="text-3xl font-bold tracking-tight text-slate-900">Advanced Reporting Suite</h1>
			<p class="text-slate-500">Real-time insights into your business performance</p>
		</div>
	</div>

	<Tabs value="sales" class="space-y-6">
		<TabsList class="grid w-full grid-cols-3 lg:w-[400px]">
			<TabsTrigger value="sales">Sales & Staff</TabsTrigger>
			<TabsTrigger value="inventory">Inventory Health</TabsTrigger>
			<TabsTrigger value="financial">Financials</TabsTrigger>
		</TabsList>

		<!-- SALES TAB -->
		<TabsContent value="sales" class="space-y-6">
			<div class="grid gap-6 md:grid-cols-2">
				<!-- Hourly Heatmap -->
				<Card class="col-span-2 md:col-span-1">
					<CardHeader>
						<CardTitle class="flex items-center gap-2">
							<Clock class="h-5 w-5 text-blue-500" />
							Hourly Sales Heatmap
						</CardTitle>
						<CardDescription>Peak sales hours by day of week</CardDescription>
					</CardHeader>
					<CardContent>
						{#if loading}
							<Skeleton class="h-[300px] w-full" />
						{:else}
							<div class="grid grid-cols-7 gap-1 text-center text-xs">
								{#each ['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat'] as day}
									<div class="font-semibold text-slate-500">{day}</div>
								{/each}
								{#each heatmap as cell}
									<div
										class="aspect-square rounded-sm transition-all hover:scale-110"
										style="background-color: rgba(59, 130, 246, {Math.min(
											cell.TotalSales / 1000,
											1
										)});"
										title={`${cell.DayOfWeek} ${cell.HourOfDay}:00 - $${cell.TotalSales}`}
									></div>
								{/each}
							</div>
						{/if}
					</CardContent>
				</Card>

				<!-- Employee Performance -->
				<Card class="col-span-2 md:col-span-1">
					<CardHeader>
						<CardTitle class="flex items-center gap-2">
							<Users class="h-5 w-5 text-purple-500" />
							Staff Performance
						</CardTitle>
						<CardDescription>Top performing employees</CardDescription>
					</CardHeader>
					<CardContent>
						<Table>
							<TableHeader>
								<TableRow>
									<TableHead>Name</TableHead>
									<TableHead class="text-right">Sales</TableHead>
									<TableHead class="text-right">Txns</TableHead>
								</TableRow>
							</TableHeader>
							<TableBody>
								{#if loading}
									<TableRow><TableCell colspan="3"><Skeleton class="h-8 w-full" /></TableCell></TableRow>
								{:else}
									{#each employeeSales as emp}
										<TableRow>
											<TableCell class="font-medium">{emp.EmployeeName}</TableCell>
											<TableCell class="text-right">{formatCurrency(emp.TotalSales)}</TableCell>
											<TableCell class="text-right">{emp.TransactionCount}</TableCell>
										</TableRow>
									{/each}
								{/if}
							</TableBody>
						</Table>
					</CardContent>
				</Card>

				<!-- Category Performance -->
				<Card class="col-span-2">
					<CardHeader>
						<CardTitle class="flex items-center gap-2">
							<PieChart class="h-5 w-5 text-indigo-500" />
							Category Performance
						</CardTitle>
						<CardDescription>Sales breakdown by category</CardDescription>
					</CardHeader>
					<CardContent>
						<Table>
							<TableHeader>
								<TableRow>
									<TableHead>Category</TableHead>
									<TableHead class="text-right">Total Sales</TableHead>
									<TableHead class="text-right">Units Sold</TableHead>
									<TableHead class="text-right">Gross Margin</TableHead>
								</TableRow>
							</TableHeader>
							<TableBody>
								{#if loading}
									<TableRow><TableCell colspan="4"><Skeleton class="h-8 w-full" /></TableCell></TableRow>
								{:else}
									{#each categoryPerformance as cat}
										<TableRow>
											<TableCell class="font-medium">{cat.CategoryName}</TableCell>
											<TableCell class="text-right">{formatCurrency(cat.TotalSales)}</TableCell>
											<TableCell class="text-right">{cat.TotalUnits}</TableCell>
											<TableCell class="text-right">{formatCurrency(cat.GrossMargin)}</TableCell>
										</TableRow>
									{/each}
								{/if}
							</TableBody>
						</Table>
					</CardContent>
				</Card>

				<!-- Customer Insights -->
				<Card class="col-span-2 md:col-span-1">
					<CardHeader>
						<CardTitle class="flex items-center gap-2">
							<Users class="h-5 w-5 text-teal-500" />
							Customer Insights
						</CardTitle>
						<CardDescription>Top customers and segments</CardDescription>
					</CardHeader>
					<CardContent>
						<Table>
							<TableHeader>
								<TableRow>
									<TableHead>Customer</TableHead>
									<TableHead>Segment</TableHead>
									<TableHead class="text-right">Spend</TableHead>
								</TableRow>
							</TableHeader>
							<TableBody>
								{#if loading}
									<TableRow><TableCell colspan="3"><Skeleton class="h-8 w-full" /></TableCell></TableRow>
								{:else}
									{#each customerInsights as cust}
										<TableRow>
											<TableCell class="font-medium">{cust.CustomerName}</TableCell>
											<TableCell>
												<span class="inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium bg-blue-100 text-blue-800">
													{cust.Segment}
												</span>
											</TableCell>
											<TableCell class="text-right">{formatCurrency(cust.TotalSpend)}</TableCell>
										</TableRow>
									{/each}
								{/if}
							</TableBody>
						</Table>
					</CardContent>
				</Card>

				<!-- Basket Analysis -->
				<Card class="col-span-2 md:col-span-1">
					<CardHeader>
						<CardTitle class="flex items-center gap-2">
							<ShoppingCart class="h-5 w-5 text-pink-500" />
							Market Basket Analysis
						</CardTitle>
						<CardDescription>Commonly bought together items</CardDescription>
					</CardHeader>
					<CardContent>
						<Table>
							<TableHeader>
								<TableRow>
									<TableHead>If bought...</TableHead>
									<TableHead>Also buys...</TableHead>
									<TableHead class="text-right">Conf.</TableHead>
								</TableRow>
							</TableHeader>
							<TableBody>
								{#if loading}
									<TableRow><TableCell colspan="3"><Skeleton class="h-8 w-full" /></TableCell></TableRow>
								{:else}
									{#each basketAnalysis as rule}
										<TableRow>
											<TableCell class="text-xs">{rule.Antecedents.join(', ')}</TableCell>
											<TableCell class="text-xs">{rule.Consequents.join(', ')}</TableCell>
											<TableCell class="text-right text-xs">{formatPercent(rule.Confidence)}</TableCell>
										</TableRow>
									{/each}
								{/if}
							</TableBody>
						</Table>
					</CardContent>
				</Card>
			</div>
		</TabsContent>

		<!-- INVENTORY TAB -->
		<TabsContent value="inventory" class="space-y-6">
			<div class="grid gap-6 md:grid-cols-2">
				<!-- Stock Aging -->
				<Card class="col-span-2">
					<CardHeader>
						<CardTitle class="flex items-center gap-2">
							<Package class="h-5 w-5 text-orange-500" />
							Stock Aging Report
						</CardTitle>
						<CardDescription>Items in stock for > 90 days</CardDescription>
					</CardHeader>
					<CardContent>
						<Table>
							<TableHeader>
								<TableRow>
									<TableHead>Product</TableHead>
									<TableHead>SKU</TableHead>
									<TableHead class="text-right">Qty</TableHead>
									<TableHead class="text-right">Days</TableHead>
									<TableHead class="text-right">Value</TableHead>
								</TableRow>
							</TableHeader>
							<TableBody>
								{#if loading}
									<TableRow><TableCell colspan="5"><Skeleton class="h-8 w-full" /></TableCell></TableRow>
								{:else}
									{#each stockAging as item}
										<TableRow>
											<TableCell class="font-medium">{item.ProductName}</TableCell>
											<TableCell>{item.SKU}</TableCell>
											<TableCell class="text-right">{item.Quantity}</TableCell>
											<TableCell class="text-right text-orange-600">{item.DaysInStock}</TableCell>
											<TableCell class="text-right">{formatCurrency(item.Value)}</TableCell>
										</TableRow>
									{/each}
								{/if}
							</TableBody>
						</Table>
					</CardContent>
				</Card>

				<!-- Dead Stock -->
				<Card class="col-span-2 md:col-span-1">
					<CardHeader>
						<CardTitle class="flex items-center gap-2">
							<FileWarning class="h-5 w-5 text-red-500" />
							Dead Stock
						</CardTitle>
						<CardDescription>No sales in 180+ days</CardDescription>
					</CardHeader>
					<CardContent>
						<Table>
							<TableHeader>
								<TableRow>
									<TableHead>Product</TableHead>
									<TableHead class="text-right">Days Idle</TableHead>
									<TableHead class="text-right">Value</TableHead>
								</TableRow>
							</TableHeader>
							<TableBody>
								{#if loading}
									<TableRow><TableCell colspan="3"><Skeleton class="h-8 w-full" /></TableCell></TableRow>
								{:else}
									{#each deadStock as item}
										<TableRow>
											<TableCell class="font-medium text-xs">{item.ProductName}</TableCell>
											<TableCell class="text-right text-red-600">{item.DaysSinceLastSale}</TableCell>
											<TableCell class="text-right">{formatCurrency(item.Value)}</TableCell>
										</TableRow>
									{/each}
								{/if}
							</TableBody>
						</Table>
					</CardContent>
				</Card>

				<!-- Supplier Performance -->
				<Card class="col-span-2 md:col-span-1">
					<CardHeader>
						<CardTitle class="flex items-center gap-2">
							<Package class="h-5 w-5 text-blue-500" />
							Supplier Performance
						</CardTitle>
						<CardDescription>Delivery times and reliability</CardDescription>
					</CardHeader>
					<CardContent>
						<Table>
							<TableHeader>
								<TableRow>
									<TableHead>Supplier</TableHead>
									<TableHead class="text-right">Lead Time</TableHead>
									<TableHead class="text-right">On-Time %</TableHead>
								</TableRow>
							</TableHeader>
							<TableBody>
								{#if loading}
									<TableRow><TableCell colspan="3"><Skeleton class="h-8 w-full" /></TableCell></TableRow>
								{:else}
									{#each supplierPerf as sup}
										<TableRow>
											<TableCell class="font-medium">{sup.supplierName}</TableCell>
											<TableCell class="text-right">{sup.averageLeadTimeDays}d</TableCell>
											<TableCell class="text-right">{formatPercent(sup.onTimeDeliveryRate)}</TableCell>
										</TableRow>
									{/each}
								{/if}
							</TableBody>
						</Table>
					</CardContent>
				</Card>

				<!-- Shrinkage -->
				<Card class="col-span-2 md:col-span-1">
					<CardHeader>
						<CardTitle class="flex items-center gap-2">
							<AlertTriangle class="h-5 w-5 text-amber-500" />
							Shrinkage Report
						</CardTitle>
						<CardDescription>Lost inventory analysis</CardDescription>
					</CardHeader>
					<CardContent>
						<Table>
							<TableHeader>
								<TableRow>
									<TableHead>Product</TableHead>
									<TableHead class="text-right">Lost Qty</TableHead>
									<TableHead class="text-right">Value</TableHead>
								</TableRow>
							</TableHeader>
							<TableBody>
								{#if loading}
									<TableRow><TableCell colspan="3"><Skeleton class="h-8 w-full" /></TableCell></TableRow>
								{:else}
									{#each shrinkage as item}
										<TableRow>
											<TableCell class="font-medium text-xs">{item.ProductName}</TableCell>
											<TableCell class="text-right text-red-600">{item.LostQuantity}</TableCell>
											<TableCell class="text-right">{formatCurrency(item.LostValue)}</TableCell>
										</TableRow>
									{/each}
								{/if}
							</TableBody>
						</Table>
					</CardContent>
				</Card>

				<!-- Returns Analysis -->
				<Card class="col-span-2 md:col-span-1">
					<CardHeader>
						<CardTitle class="flex items-center gap-2">
							<Undo2 class="h-5 w-5 text-purple-500" />
							Returns Analysis
						</CardTitle>
						<CardDescription>Most returned products</CardDescription>
					</CardHeader>
					<CardContent>
						<Table>
							<TableHeader>
								<TableRow>
									<TableHead>Product</TableHead>
									<TableHead class="text-right">Count</TableHead>
									<TableHead class="text-right">Rate</TableHead>
								</TableRow>
							</TableHeader>
							<TableBody>
								{#if loading}
									<TableRow><TableCell colspan="3"><Skeleton class="h-8 w-full" /></TableCell></TableRow>
								{:else}
									{#each returnsAnalysis as item}
										<TableRow>
											<TableCell class="font-medium text-xs">{item.ProductName}</TableCell>
											<TableCell class="text-right">{item.ReturnCount}</TableCell>
											<TableCell class="text-right text-red-500">{formatPercent(item.ReturnRate)}</TableCell>
										</TableRow>
									{/each}
								{/if}
							</TableBody>
						</Table>
					</CardContent>
				</Card>
			</div>
		</TabsContent>

		<!-- FINANCIAL TAB -->
		<TabsContent value="financial" class="space-y-6">
			<div class="grid gap-6 md:grid-cols-2">
				<!-- GMROI -->
				<Card class="col-span-2">
					<CardHeader>
						<CardTitle class="flex items-center gap-2">
							<TrendingUp class="h-5 w-5 text-green-500" />
							GMROI Analysis
						</CardTitle>
						<CardDescription>Gross Margin Return on Investment</CardDescription>
					</CardHeader>
					<CardContent>
						<Table>
							<TableHeader>
								<TableRow>
									<TableHead>Product</TableHead>
									<TableHead class="text-right">Revenue</TableHead>
									<TableHead class="text-right">Margin</TableHead>
									<TableHead class="text-right">GMROI</TableHead>
								</TableRow>
							</TableHeader>
							<TableBody>
								{#if loading}
									<TableRow><TableCell colspan="4"><Skeleton class="h-8 w-full" /></TableCell></TableRow>
								{:else}
									{#each gmroiData as item}
										<TableRow>
											<TableCell class="font-medium">{item.ProductName}</TableCell>
											<TableCell class="text-right">{formatCurrency(item.Revenue)}</TableCell>
											<TableCell class="text-right">{formatCurrency(item.GrossMargin)}</TableCell>
											<TableCell class="text-right font-bold text-green-600"
												>{item.GMROI.toFixed(2)}x</TableCell
											>
										</TableRow>
									{/each}
								{/if}
							</TableBody>
						</Table>
					</CardContent>
				</Card>

				<!-- Void Audit -->
				<Card class="col-span-2 md:col-span-1">
					<CardHeader>
						<CardTitle class="flex items-center gap-2">
							<AlertTriangle class="h-5 w-5 text-red-500" />
							Void Audit Log
						</CardTitle>
						<CardDescription>Suspicious void transactions</CardDescription>
					</CardHeader>
					<CardContent>
						<Table>
							<TableHeader>
								<TableRow>
									<TableHead>Cashier</TableHead>
									<TableHead>Reason</TableHead>
									<TableHead class="text-right">Amount</TableHead>
								</TableRow>
							</TableHeader>
							<TableBody>
								{#if loading}
									<TableRow><TableCell colspan="3"><Skeleton class="h-8 w-full" /></TableCell></TableRow>
								{:else}
									{#each voidAudit as item}
										<TableRow>
											<TableCell class="font-medium">{item.CashierName}</TableCell>
											<TableCell class="text-xs text-slate-500">{item.Reason}</TableCell>
											<TableCell class="text-right text-red-600">{formatCurrency(item.VoidedAmount)}</TableCell>
										</TableRow>
									{/each}
								{/if}
							</TableBody>
						</Table>
					</CardContent>
				</Card>

				<!-- Cash Reconciliation -->
				<Card class="col-span-2 md:col-span-1">
					<CardHeader>
						<CardTitle class="flex items-center gap-2">
							<DollarSign class="h-5 w-5 text-emerald-500" />
							Cash Reconciliation
						</CardTitle>
						<CardDescription>Register discrepancies</CardDescription>
					</CardHeader>
					<CardContent>
						<Table>
							<TableHeader>
								<TableRow>
									<TableHead>Cashier</TableHead>
									<TableHead class="text-right">System</TableHead>
									<TableHead class="text-right">Diff</TableHead>
								</TableRow>
							</TableHeader>
							<TableBody>
								{#if loading}
									<TableRow><TableCell colspan="3"><Skeleton class="h-8 w-full" /></TableCell></TableRow>
								{:else}
									{#each cashReconciliation as item}
										<TableRow>
											<TableCell class="font-medium">{item.CashierName}</TableCell>
											<TableCell class="text-right">{formatCurrency(item.SystemCalculated)}</TableCell>
											<TableCell class="text-right font-bold {item.Discrepancy < 0 ? 'text-red-500' : 'text-green-500'}">
												{formatCurrency(item.Discrepancy)}
											</TableCell>
										</TableRow>
									{/each}
								{/if}
							</TableBody>
						</Table>
					</CardContent>
				</Card>

				<!-- Tax Liability -->
				<Card class="col-span-2">
					<CardHeader>
						<CardTitle class="flex items-center gap-2">
							<Scale class="h-5 w-5 text-slate-500" />
							Tax Liability Estimate
						</CardTitle>
						<CardDescription>Estimated tax collected for current period</CardDescription>
					</CardHeader>
					<CardContent>
						<div class="flex items-center justify-between p-4 bg-slate-50 rounded-lg">
							<div>
								<p class="text-sm font-medium text-slate-500">Taxable Sales</p>
								<p class="text-2xl font-bold text-slate-900">{formatCurrency(taxLiability?.TaxableSales || 0)}</p>
							</div>
							<div class="text-right">
								<p class="text-sm font-medium text-slate-500">Tax Collected</p>
								<p class="text-2xl font-bold text-blue-600">{formatCurrency(taxLiability?.TaxCollected || 0)}</p>
							</div>
						</div>
					</CardContent>
				</Card>
			</div>
		</TabsContent>
	</Tabs>
</div>
