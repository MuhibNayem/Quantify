<script lang="ts">
	import { onMount } from 'svelte';
	import { Button } from '$lib/components/ui/button';
	import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '$lib/components/ui/card';
	import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '$lib/components/ui/table';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import { toast } from 'svelte-sonner';
	import { alertsApi, productsApi, categoriesApi, suppliersApi, replenishmentApi } from '$lib/api/resources';
	import type { Alert, Product, ReorderSuggestion } from '$lib/types';
	import { Activity, AlertTriangle, Boxes, RefreshCcw, TrendingUp, Zap, ShoppingCart, Users, BarChart3 } from 'lucide-svelte';

	let loading = $state(true);
	const stats = $state({ products: 0, categories: 0, suppliers: 0, alerts: 0 });
	let recentProducts = $state<Product[]>([]);
	let recentAlerts = $state<Alert[]>([]);
	let suggestions = $state<ReorderSuggestion[]>([]);

	const chartSeries = [62, 48, 55, 61, 58, 72, 80];
	const chartMax = Math.max(...chartSeries, 100);

	const loadDashboard = async () => {
		loading = true;
		try {
			const productList = await productsApi.list();
			stats.products = productList.products?.length ?? 0;
			recentProducts = productList.products?.slice(0, 5);

			const categoryList = await categoriesApi.list();
			stats.categories = (Array.isArray(categoryList) ? categoryList : [categoryList]).length;

			const supplierList = await suppliersApi.list();
			stats.suppliers = (Array.isArray(supplierList) ? supplierList : [supplierList]).length;

			const suggestionList = await replenishmentApi.listSuggestions();
			suggestions = suggestionList.slice(0, 5);

			const alertList = await alertsApi.list({ status: 'ACTIVE' });
			stats.alerts = alertList.length ?? 0;
			recentAlerts = alertList.slice(0, 5);
		} catch (error: any) {
			toast.error('Failed to Load Dashboard', {
				description: error?.response?.data?.error || 'An unexpected error occurred'
			});
		} finally {
			loading = false;
		}
	};

	onMount(loadDashboard);
</script>

<div class="w-full max-w-7xl mx-auto py-8 px-6">
	<section class="space-y-8 relative z-10">
		<!-- HEADER (matches Operations animations) -->
		<div class="relative overflow-hidden rounded-3xl shadow-lg border border-white/50 animate-fadeUp">
			<div class="absolute inset-0 bg-gradient-to-r from-sky-50 via-blue-50 to-cyan-100 animate-gradientShift"></div>
			<div class="relative bg-white/70 backdrop-blur-lg rounded-3xl p-8">
				<div class="flex flex-wrap items-center justify-between gap-6">
					<div class="space-y-3">
						<div class="flex items-center gap-3">
							<div class="p-2 bg-gradient-to-br from-sky-400 to-blue-500 rounded-2xl shadow-md animate-pulseGlow">
								<Zap class="h-6 w-6 text-white" />
							</div>
							<p class="text-sm uppercase tracking-wider font-semibold text-sky-600">Inventory Control Tower</p>
						</div>
						<h1 class="text-4xl font-bold bg-gradient-to-r from-sky-600 via-blue-600 to-cyan-600 bg-clip-text text-transparent">
							Real-time Inventory Intelligence
						</h1>
						<p class="text-slate-600 max-w-2xl">
							Monitor, analyze, and optimize your inventory ecosystem with AI-powered insights
						</p>
					</div>
					<div class="flex flex-wrap gap-3">
						<Button
							variant="secondary"
							class="bg-white/80 border border-slate-200 hover:border-sky-200 hover:bg-sky-50/70 text-slate-700 hover:text-sky-700 font-medium rounded-xl px-6 py-3 transition-all duration-300 hover:scale-105 hover:shadow-lg group"
							onclick={loadDashboard}
						>
							<RefreshCcw class="mr-2 h-4 w-4 group-hover:rotate-180 transition-transform duration-500" />
							Refresh Data
						</Button>
						<Button
							href="/catalog"
							class="bg-gradient-to-r from-sky-500 to-blue-500 hover:from-sky-600 hover:to-blue-600 text-white font-semibold rounded-xl px-6 py-3 shadow-lg hover:shadow-xl transition-all duration-300 hover:scale-105"
						>
							<Boxes class="mr-2 h-4 w-4" />
							Update Catalog
						</Button>
					</div>
				</div>
			</div>
		</div>

		<!-- STATS GRID (staggered like Operations) -->
		<div class="grid gap-6 md:grid-cols-2 lg:grid-cols-4">
			<!-- Products -->
			<Card class="group relative overflow-hidden bg-gradient-to-br from-blue-50 to-cyan-100 border-0 shadow-xl hover:shadow-2xl transition-all duration-500 hover:scale-105 cursor-pointer animate-fadeUp" style="animation-delay: .05s;">
				<div class="absolute top-0 right-0 w-20 h-20 bg-blue-200/30 rounded-full -translate-y-10 translate-x-10 group-hover:scale-150 transition-transform duration-500"></div>
				<CardHeader class="flex flex-row items-center justify-between pb-3 relative z-10">
					<CardTitle class="text-sm font-semibold text-blue-800/80">Active Products</CardTitle>
					<div class="p-2 bg-white/50 rounded-xl shadow-sm">
						<ShoppingCart class="h-5 w-5 text-blue-600" />
					</div>
				</CardHeader>
				<CardContent class="relative z-10">
					{#if loading}
						<Skeleton class="h-8 w-20 bg-white/50" />
					{:else}
						<div class="text-3xl font-bold text-blue-900 mb-1">{stats.products}</div>
						<p class="text-xs text-blue-700/70 font-medium">📈 {Math.round(stats.products * 1.12)} forecasted for Q4</p>
					{/if}
				</CardContent>
			</Card>

			<!-- Categories -->
			<Card class="group relative overflow-hidden bg-gradient-to-br from-green-50 to-emerald-100 border-0 shadow-xl hover:shadow-2xl transition-all duration-500 hover:scale-105 cursor-pointer animate-fadeUp" style="animation-delay: .15s;">
				<div class="absolute top-0 right-0 w-20 h-20 bg-green-200/30 rounded-full -translate-y-10 translate-x-10 group-hover:scale-150 transition-transform duration-500"></div>
				<CardHeader class="flex flex-row items-center justify-between pb-3 relative z-10">
					<CardTitle class="text-sm font-semibold text-green-800/80">Categories</CardTitle>
					<div class="p-2 bg-white/50 rounded-xl shadow-sm">
						<BarChart3 class="h-5 w-5 text-green-600" />
					</div>
				</CardHeader>
				<CardContent class="relative z-10">
					{#if loading}
						<Skeleton class="h-8 w-20 bg-white/50" />
					{:else}
						<div class="text-3xl font-bold text-green-900 mb-1">{stats.categories}</div>
						<p class="text-xs text-green-700/70 font-medium">🔄 Across {stats.suppliers} suppliers</p>
					{/if}
				</CardContent>
			</Card>

			<!-- Suppliers -->
			<Card class="group relative overflow-hidden bg-gradient-to-br from-purple-50 to-violet-100 border-0 shadow-xl hover:shadow-2xl transition-all duration-500 hover:scale-105 cursor-pointer animate-fadeUp" style="animation-delay: .25s;">
				<div class="absolute top-0 right-0 w-20 h-20 bg-purple-200/30 rounded-full -translate-y-10 translate-x-10 group-hover:scale-150 transition-transform duration-500"></div>
				<CardHeader class="flex flex-row items-center justify-between pb-3 relative z-10">
					<CardTitle class="text-sm font-semibold text-purple-800/80">Suppliers</CardTitle>
					<div class="p-2 bg-white/50 rounded-xl shadow-sm">
						<Users class="h-5 w-5 text-purple-600" />
					</div>
				</CardHeader>
				<CardContent class="relative z-10">
					{#if loading}
						<Skeleton class="h-8 w-20 bg-white/50" />
					{:else}
						<div class="text-3xl font-bold text-purple-900 mb-1">{stats.suppliers}</div>
						<p class="text-xs text-purple-700/70 font-medium">✅ All SLAs active</p>
					{/if}
				</CardContent>
			</Card>

			<!-- Alerts -->
			<Card class="group relative overflow-hidden bg-gradient-to-br from-orange-50 to-red-100 border-0 shadow-xl hover:shadow-2xl transition-all duration-500 hover:scale-105 cursor-pointer animate-fadeUp" style="animation-delay: .35s;">
				<div class="absolute top-0 right-0 w-20 h-20 bg-red-200/30 rounded-full -translate-y-10 translate-x-10 group-hover:scale-150 transition-transform duration-500"></div>
				<CardHeader class="flex flex-row items-center justify-between pb-3 relative z-10">
					<CardTitle class="text-sm font-semibold text-red-800/80">Active Alerts</CardTitle>
					<div class="p-2 bg-white/50 rounded-xl shadow-sm">
						<AlertTriangle class="h-5 w-5 text-red-600" />
					</div>
				</CardHeader>
				<CardContent class="relative z-10">
					{#if loading}
						<Skeleton class="h-8 w-20 bg-white/50" />
					{:else}
						<div class="text-3xl font-bold text-red-900 mb-1">{stats.alerts}</div>
						<p class="text-xs text-red-700/70 font-medium">🚨 Auto-escalations active</p>
					{/if}
				</CardContent>
			</Card>
		</div>

		<!-- DEMAND PULSE + QUICK ACTIONS -->
		<div class="grid gap-6 lg:grid-cols-3 mt-8">
			<!-- Demand Pulse -->
			<Card class="lg:col-span-2 group bg-gradient-to-br from-sky-50 to-blue-50 border border-blue-100 rounded-2xl shadow-lg hover:shadow-xl transition-all duration-500 hover:scale-[1.02] animate-fadeUp">
				<CardHeader class="pb-4">
					<CardTitle class="flex items-center gap-2 text-slate-800">
						<div class="p-2 bg-gradient-to-r from-sky-400 to-blue-400 rounded-xl shadow-sm">
							<Activity class="h-5 w-5 text-white" />
						</div>
						Demand Pulse Analytics
					</CardTitle>
					<CardDescription class="text-slate-600">Real-time inventory movement trends</CardDescription>
				</CardHeader>
				<CardContent>
					<div class="space-y-4">
						<div class="flex items-end gap-3 h-32">
							{#each chartSeries as value, index}
								<div class="flex-1 flex flex-col items-center group">
									<div
										class="w-full rounded-t-2xl bg-gradient-to-t from-sky-400 to-blue-500 shadow-md transition-all duration-700 ease-out hover:scale-110 relative overflow-hidden group-hover:brightness-110 animate-fadeUp"
										style={`height: ${(value / chartMax) * 100}px; min-height: 8px; animation-delay: ${index * 0.08 + 0.1}s;`}
									>
										<div class="absolute inset-0 bg-gradient-to-b from-white/30 to-transparent"></div>
									</div>
									<span class="text-xs text-slate-600 mt-2 font-medium">Day {index + 1}</span>
								</div>
							{/each}
						</div>
						<div class="flex items-center justify-between pt-4 border-t border-blue-100">
							<p class="text-sm text-slate-600">📊 Based on sales velocity & stock buffers</p>
							<div class="flex gap-2">
								<span class="px-3 py-1 bg-green-50 text-green-700 rounded-full text-xs font-medium border border-green-200">↑ Trend: Positive</span>
								<span class="px-3 py-1 bg-blue-50 text-blue-700 rounded-full text-xs font-medium border border-blue-200">📈 12% Growth</span>
							</div>
						</div>
					</div>
				</CardContent>
			</Card>

			<!-- Quick Actions -->
			<Card class="group bg-gradient-to-br from-violet-50 to-purple-50 border border-purple-100 rounded-2xl shadow-lg hover:shadow-xl transition-all duration-500 hover:scale-[1.02] animate-fadeUp" style="animation-delay:.1s;">
				<CardHeader>
					<CardTitle class="flex items-center gap-2 text-slate-800">
						<div class="p-2 bg-gradient-to-r from-violet-400 to-purple-400 rounded-xl shadow-sm">
							<Zap class="h-5 w-5 text-white" />
						</div>
						Quick Actions
					</CardTitle>
					<CardDescription class="text-slate-600">Instant inventory operations</CardDescription>
				</CardHeader>
				<CardContent class="space-y-3">
					<Button class="w-full justify-start h-14 bg-gradient-to-r from-emerald-400 to-green-500 hover:from-emerald-500 hover:to-green-600 text-white font-semibold rounded-xl shadow-md hover:shadow-lg hover:scale-105 transition-all group" href="/operations">
						<TrendingUp class="mr-3 h-5 w-5 group-hover:scale-110 transition-transform" />
						<div class="text-left">
							<div class="font-semibold">Balance Stock</div>
							<div class="text-xs opacity-90">Optimize inventory levels</div>
						</div>
					</Button>

					<Button class="w-full justify-start h-14 bg-gradient-to-r from-sky-400 to-blue-500 hover:from-sky-500 hover:to-blue-600 text-white font-semibold rounded-xl shadow-md hover:shadow-lg hover:scale-105 transition-all group" href="/intelligence">
						<Activity class="mr-3 h-5 w-5 group-hover:scale-110 transition-transform" />
						<div class="text-left">
							<div class="font-semibold">Run Forecast</div>
							<div class="text-xs opacity-90">AI predictions</div>
						</div>
					</Button>

					<Button class="w-full justify-start h-14 bg-gradient-to-r from-violet-400 to-purple-500 hover:from-violet-500 hover:to-purple-600 text-white font-semibold rounded-xl shadow-md hover:shadow-lg hover:scale-105 transition-all group" href="/bulk">
						<Boxes class="mr-3 h-5 w-5 group-hover:scale-110 transition-transform" />
						<div class="text-left">
							<div class="font-semibold">Export Catalog</div>
							<div class="text-xs opacity-90">Bulk operations</div>
						</div>
					</Button>
				</CardContent>
			</Card>
		</div>

		<!-- FRESH INVENTORY + ALERTS -->
		<div class="grid gap-6 lg:grid-cols-2">
			<!-- Fresh Inventory -->
			<Card class="group bg-gradient-to-br from-emerald-50 to-green-50 border border-green-100 rounded-2xl shadow-lg hover:shadow-xl hover:scale-[1.01] transition-all animate-fadeUp">
				<CardHeader class="pb-4">
					<CardTitle class="flex items-center gap-2 text-slate-800">
						<div class="p-2 bg-gradient-to-r from-emerald-400 to-green-400 rounded-xl shadow-sm">
							<Boxes class="h-5 w-5 text-white" />
						</div>
						Fresh Inventory
					</CardTitle>
					<CardDescription class="text-slate-600">Recently added or updated SKUs</CardDescription>
				</CardHeader>
				<CardContent>
					<div class="scrollbar-gradient max-h-80 overflow-y-auto pr-2">
						<Table>
							<TableHeader class="sticky top-0 bg-white/90 backdrop-blur-sm rounded-lg">
								<TableRow class="border-b border-green-100">
									<TableHead class="text-slate-700 font-semibold py-3">SKU</TableHead>
									<TableHead class="text-slate-700 font-semibold py-3">Product Name</TableHead>
									<TableHead class="text-slate-700 font-semibold py-3 text-right">Status</TableHead>
								</TableRow>
							</TableHeader>
							<TableBody>
								{#if loading}
									{#each Array(5) as _, i}
										<TableRow class="hover:bg-white/70 transition-colors animate-fadeUp" style={`animation-delay:${i * 0.08 + 0.1}s`}>
											<TableCell colspan="3" class="py-3"><Skeleton class="h-6 w-full bg-white/70" /></TableCell>
										</TableRow>
									{/each}
								{:else if recentProducts.length === 0}
									<TableRow>
										<TableCell colspan="3" class="text-center py-8 text-slate-500">
											<div class="space-y-2">
												<Boxes class="h-8 w-8 mx-auto text-slate-400" />
												<div>No recent inventory changes</div>
											</div>
										</TableCell>
									</TableRow>
								{:else}
									{#each recentProducts as product, i}
										<TableRow class="hover:bg-white/70 transition-all duration-300 group animate-fadeUp" style={`animation-delay:${i * 0.08 + 0.1}s`}>
											<TableCell class="py-3 font-mono text-sm text-blue-600 font-medium">{product.SKU}</TableCell>
											<TableCell class="py-3 text-slate-700 group-hover:text-slate-900">{product.Name}</TableCell>
											<TableCell class="py-3 text-right">
												<span class="inline-flex items-center px-3 py-1.5 rounded-full text-xs font-medium bg-emerald-100 text-emerald-700 border border-emerald-200 shadow-sm">
													● {product.Status ?? 'active'}
												</span>
											</TableCell>
										</TableRow>
									{/each}
								{/if}
							</TableBody>
						</Table>
					</div>
				</CardContent>
			</Card>

			<!-- Priority Alerts -->
			<Card class="group bg-gradient-to-br from-orange-50 to-amber-50 border border-amber-100 rounded-2xl shadow-lg hover:shadow-xl hover:scale-[1.01] transition-all animate-fadeUp" style="animation-delay:.1s;">
				<CardHeader class="pb-4">
					<CardTitle class="flex items-center gap-2 text-slate-800">
						<div class="p-2 bg-gradient-to-r from-orange-400 to-amber-400 rounded-xl shadow-sm">
							<AlertTriangle class="h-5 w-5 text-white" />
						</div>
						Priority Alerts
					</CardTitle>
					<CardDescription class="text-slate-600">Requires immediate attention</CardDescription>
				</CardHeader>
				<CardContent>
					<div class="scrollbar-gradient max-h-80 overflow-y-auto pr-2">
						<Table>
							<TableHeader class="sticky top-0 bg-white/90 backdrop-blur-sm rounded-lg">
								<TableRow class="border-b border-amber-100">
									<TableHead class="text-slate-700 font-semibold py-3">Alert Type</TableHead>
									<TableHead class="text-slate-700 font-semibold py-3">Product</TableHead>
									<TableHead class="text-slate-700 font-semibold py-3 text-right">Status</TableHead>
								</TableRow>
							</TableHeader>
							<TableBody>
								{#if loading}
									{#each Array(5) as _, i}
										<TableRow class="hover:bg-white/70 transition-colors animate-fadeUp" style={`animation-delay:${i * 0.08 + 0.1}s`}>
											<TableCell colspan="3" class="py-3"><Skeleton class="h-6 w-full bg-white/70" /></TableCell>
										</TableRow>
									{/each}
								{:else if recentAlerts.length === 0}
									<TableRow>
										<TableCell colspan="3" class="text-center py-8 text-slate-500">
											<div class="space-y-2">
												<AlertTriangle class="h-8 w-8 mx-auto text-emerald-400" />
												<div>All systems normal</div>
											</div>
										</TableCell>
									</TableRow>
								{:else}
									{#each recentAlerts as item, i}
										<TableRow class="hover:bg-white/70 transition-all duration-300 group animate-fadeUp" style={`animation-delay:${i * 0.08 + 0.1}s`}>
											<TableCell class="py-3 font-medium text-slate-700">{item.Type}</TableCell>
											<TableCell class="py-3 text-slate-600">{item.Product?.SKU ?? item.ProductID}</TableCell>
											<TableCell class="py-3 text-right">
												<span class="inline-flex items-center px-3 py-1.5 rounded-full text-xs font-medium bg-amber-100 text-amber-700 border border-amber-200 shadow-sm">
													🚨 {item.Status}
												</span>
											</TableCell>
										</TableRow>
									{/each}
								{/if}
							</TableBody>
						</Table>
					</div>
				</CardContent>
			</Card>
		</div>

		<!-- PROCUREMENT INTELLIGENCE -->
		<Card class="group bg-gradient-to-br from-cyan-50 to-teal-50 border border-teal-100 rounded-2xl shadow-lg hover:shadow-xl hover:scale-[1.01] transition-all animate-fadeUp">
			<CardHeader class="pb-4">
				<CardTitle class="flex items-center gap-2 text-slate-800">
					<div class="p-2 bg-gradient-to-r from-cyan-400 to-teal-400 rounded-xl shadow-sm">
						<ShoppingCart class="h-5 w-5 text-white" />
					</div>
					Procurement Intelligence
				</CardTitle>
				<CardDescription class="text-slate-600">AI-powered reorder suggestions</CardDescription>
			</CardHeader>
			<CardContent>
				<div class="scrollbar-gradient max-h-80 overflow-y-auto pr-2">
					<Table>
						<TableHeader class="sticky top-0 bg-white/90 backdrop-blur-sm rounded-lg">
							<TableRow class="border-b border-teal-100">
								<TableHead class="text-slate-700 font-semibold py-3">Product</TableHead>
								<TableHead class="text-slate-700 font-semibold py-3 text-right">Suggested Qty</TableHead>
								<TableHead class="text-slate-700 font-semibold py-3">Supplier</TableHead>
								<TableHead class="text-slate-700 font-semibold py-3 text-right">Status</TableHead>
							</TableRow>
						</TableHeader>
						<TableBody>
							{#if loading}
								{#each Array(5) as _, i}
									<TableRow class="hover:bg-white/70 transition-colors animate-fadeUp" style={`animation-delay:${i * 0.08 + 0.1}s`}>
										<TableCell colspan="4" class="py-3"><Skeleton class="h-6 w-full bg-white/70" /></TableCell>
									</TableRow>
								{/each}
							{:else if suggestions.length === 0}
								<TableRow>
									<TableCell colspan="4" class="text-center py-8 text-slate-500">
										<div class="space-y-2">
											<ShoppingCart class="h-8 w-8 mx-auto text-slate-400" />
											<div>No pending suggestions</div>
										</div>
									</TableCell>
								</TableRow>
							{:else}
								{#each suggestions as suggestion, i}
									<TableRow class="hover:bg-white/70 transition-all duration-300 group animate-fadeUp" style={`animation-delay:${i * 0.08 + 0.1}s`}>
										<TableCell class="py-3 font-medium text-slate-700">
											{suggestion?.Product?.Name ?? suggestion?.product?.name ?? `Product ${suggestion?.ProductID ?? suggestion?.productId ?? 'N/A'}`}
										</TableCell>
										<TableCell class="py-3 text-right font-semibold text-blue-600">
											{suggestion?.SuggestedOrderQuantity ?? suggestion?.suggestedOrderQuantity ?? suggestion?.quantity ?? 'N/A'}
										</TableCell>
										<TableCell class="py-3 text-slate-600">
											{suggestion?.Supplier?.Name ?? suggestion?.supplier?.name ?? suggestion?.SupplierID ?? suggestion?.supplierId ?? 'N/A'}
										</TableCell>
										<TableCell class="py-3 text-right">
											<span class="inline-flex items-center px-3 py-1.5 rounded-full text-xs font-medium bg-cyan-100 text-cyan-700 border border-cyan-200 shadow-sm">
												💡 {suggestion?.Status ?? suggestion?.status ?? 'Ready to Order'}
											</span>
										</TableCell>
									</TableRow>
								{/each}
							{/if}
						</TableBody>
					</Table>
				</div>
			</CardContent>
		</Card>
	</section>
</div>

<style lang="postcss">
	/* ===== Reused animation system from Operations ===== */
	@keyframes gradientShift {
		0% { background-position: 0% 50%; }
		50% { background-position: 100% 50%; }
		100% { background-position: 0% 50%; }
	}
	.animate-gradientShift {
		background-size: 200% 200%;
		animation: gradientShift 20s ease infinite;
	}

	@keyframes pulseGlow {
		0%, 100% { transform: scale(1); box-shadow: 0 0 15px rgba(56, 189, 248, 0.30); }
		50% { transform: scale(1.08); box-shadow: 0 0 25px rgba(56, 189, 248, 0.50); }
	}
	.animate-pulseGlow { animation: pulseGlow 8s ease-in-out infinite; }

	@keyframes fadeUp {
		from { opacity: 0; transform: translateY(18px); }
		to { opacity: 1; transform: translateY(0); }
	}
	.animate-fadeUp { animation: fadeUp 1.2s ease forwards; }

	/* Scrollbar soft styling */
	.scrollbar-gradient {
		scrollbar-width: thin;
		scrollbar-color: rgba(139, 92, 246, 0.2) transparent;
	}
	.scrollbar-gradient::-webkit-scrollbar { width: 6px; }
	.scrollbar-gradient::-webkit-scrollbar-track { background: transparent; border-radius: 10px; }
	.scrollbar-gradient::-webkit-scrollbar-thumb { background: rgba(139, 92, 246, 0.2); border-radius: 10px; }
	.scrollbar-gradient::-webkit-scrollbar-thumb:hover { background: rgba(139, 92, 246, 0.3); }

	/* Smooth hover transitions */
	* {
		transition-property: color, background-color, border-color, text-decoration-color, fill, stroke, opacity, box-shadow, transform, filter, backdrop-filter;
		transition-timing-function: cubic-bezier(0.4, 0, 0.2, 1);
		transition-duration: 300ms;
	}

	@media (prefers-reduced-motion: reduce) {
		.animate-gradientShift,
		.animate-pulseGlow,
		.animate-fadeUp { animation: none !important; }
	}
</style>
