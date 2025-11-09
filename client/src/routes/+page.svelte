<script lang="ts">
	import { onMount } from 'svelte';
	import { Button } from '$lib/components/ui/button';
	import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '$lib/components/ui/card';
	import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '$lib/components/ui/table';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import { toast } from 'svelte-sonner';
	import { alertsApi, productsApi, categoriesApi, suppliersApi, replenishmentApi } from '$lib/api/resources';
	import type { Alert, Product, ReorderSuggestion } from '$lib/types';
	import { Activity, AlertTriangle, Boxes, Factory, RefreshCcw, TrendingUp } from 'lucide-svelte';

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
		try {
			const productList = await productsApi.list();
			stats.products = productList.products?.length ?? 0;
			recentProducts = productList.products?.slice(0, 5);
		} catch (error: any) {
			stats.products = 0;
			recentProducts = [];
			toast.error('Failed to Load Products', {
				description: error?.response?.data?.error || 'Unable to load product list'
			});
		}

		try {
			const categoryList = await categoriesApi.list();
			stats.categories = (Array.isArray(categoryList) ? categoryList : [categoryList]).length;
		} catch (error: any) {
			stats.categories = 0;
			toast.error('Failed to Load Categories', {
				description: error?.response?.data?.error || 'Unable to load categories'
			});
		}

		try {
			const supplierList = await suppliersApi.list();
			stats.suppliers = (Array.isArray(supplierList) ? supplierList : [supplierList]).length;
		} catch (error: any) {
			stats.suppliers = 0;
			toast.error('Failed to Load Suppliers', {
				description: error?.response?.data?.error || 'Unable to load suppliers'
			});
		}

		try {
			const alertList = await alertsApi.list({ status: 'ACTIVE' });
			stats.alerts = alertList.length ?? 0;
			recentAlerts = alertList.slice(0, 5);
		} catch (error: any) {
			stats.alerts = 0;
			recentAlerts = [];
			toast.error('Failed to Load Alerts', {
				description: error?.response?.data?.error || 'Unable to load alert data'
			});
		}

		// 🟨 REPLENISHMENT SUGGESTIONS
		try {
			const suggestionList = await replenishmentApi.listSuggestions();
			suggestions = suggestionList.slice(0, 5);
		} catch (error: any) {
			suggestions = [];
			toast.error('Failed to Load Replenishment Suggestions', {
				description: error?.response?.data?.error || 'Unable to load replenishment data'
			});
		}
	} catch (error: any) {
		// 🚨 fallback if something unexpected breaks outside the inner blocks
		toast.error('Failed to Load Dashboard', {
			description: error?.response?.data?.error || 'An unexpected error occurred while loading dashboard data'
		});
	} finally {
		loading = false;
	}
};


	onMount(loadDashboard);
</script>

<div class="w-full max-w-7xl mx-auto py-8 px-4">
	<section class="space-y-8">
		<!-- Header -->
		<div class="flex flex-wrap items-center justify-between gap-4">
			<div>
				<p class="text-sm uppercase tracking-wide text-slate-500 mb-2">Control tower</p>
				<h1 class="text-3xl font-semibold text-slate-800">
					Realtime inventory intelligence
				</h1>
			</div>
			<div class="flex flex-wrap gap-3">
				<Button
					variant="secondary"
					class="bg-white border border-slate-200 text-slate-700 hover:bg-slate-50 hover:border-slate-300 transition-all duration-200"
					onclick={loadDashboard}
				>
					<RefreshCcw class="mr-2 h-4 w-4" /> Refresh
				</Button>
				<Button
					href="/catalog"
					class="bg-blue-500 hover:bg-blue-600 text-white font-medium transition-all duration-200"
				>
					Update catalog
				</Button>
			</div>
		</div>

		<!-- Stats Cards -->
		<div class="grid gap-5 lg:grid-cols-4">
			<Card class="bg-white border border-slate-100 shadow-sm hover:shadow-md transition-all duration-200">
				<CardHeader class="flex flex-row items-center justify-between pb-2">
					<CardTitle class="text-sm font-medium text-slate-600">Active SKUs</CardTitle>
					<div class="p-2 bg-blue-50 rounded-lg">
						<Boxes class="h-4 w-4 text-blue-500" />
					</div>
				</CardHeader>
				<CardContent>
					{#if loading}
						<Skeleton class="h-8 w-20 bg-slate-100" />
					{:else}
						<div class="text-2xl font-bold text-slate-800 mb-1">{stats.products}</div>
						<p class="text-xs text-slate-500">{Math.round(stats.products * 1.12)} forecasted for Q4</p>
					{/if}
				</CardContent>
			</Card>

			<Card class="bg-white border border-slate-100 shadow-sm hover:shadow-md transition-all duration-200">
				<CardHeader class="flex flex-row items-center justify-between pb-2">
					<CardTitle class="text-sm font-medium text-slate-600">Strategic categories</CardTitle>
					<div class="p-2 bg-emerald-50 rounded-lg">
						<Activity class="h-4 w-4 text-emerald-500" />
					</div>
				</CardHeader>
				<CardContent>
					{#if loading}
						<Skeleton class="h-8 w-20 bg-slate-100" />
					{:else}
						<div class="text-2xl font-bold text-slate-800 mb-1">{stats.categories}</div>
						<p class="text-xs text-slate-500">Converged across {stats.suppliers} suppliers</p>
					{/if}
				</CardContent>
			</Card>

			<Card class="bg-white border border-slate-100 shadow-sm hover:shadow-md transition-all duration-200">
				<CardHeader class="flex flex-row items-center justify-between pb-2">
					<CardTitle class="text-sm font-medium text-slate-600">Partners online</CardTitle>
					<div class="p-2 bg-amber-50 rounded-lg">
						<Factory class="h-4 w-4 text-amber-500" />
					</div>
				</CardHeader>
				<CardContent>
					{#if loading}
						<Skeleton class="h-8 w-20 bg-slate-100" />
					{:else}
						<div class="text-2xl font-bold text-slate-800 mb-1">{stats.suppliers}</div>
						<p class="text-xs text-slate-500">Supplier SLAs synced</p>
					{/if}
				</CardContent>
			</Card>

			<Card class="bg-white border border-slate-100 shadow-sm hover:shadow-md transition-all duration-200">
				<CardHeader class="flex flex-row items-center justify-between pb-2">
					<CardTitle class="text-sm font-medium text-slate-600">Active alerts</CardTitle>
					<div class="p-2 bg-rose-50 rounded-lg">
						<AlertTriangle class="h-4 w-4 text-rose-500" />
					</div>
				</CardHeader>
				<CardContent>
					{#if loading}
						<Skeleton class="h-8 w-20 bg-slate-100" />
					{:else}
						<div class="text-2xl font-bold text-slate-800 mb-1">{stats.alerts}</div>
						<p class="text-xs text-slate-500">Auto-escalations enabled</p>
					{/if}
				</CardContent>
			</Card>
		</div>

		<!-- Demand Pulse -->
		<div class="grid gap-6 lg:grid-cols-[2fr,1fr]">
			<Card class="bg-white border border-slate-100 shadow-sm hover:shadow-md transition-all duration-200">
				<CardHeader>
					<CardTitle class="text-slate-800">Demand pulse</CardTitle>
					<CardDescription class="text-slate-500">Seven-day weighted trend (mock data)</CardDescription>
				</CardHeader>
				<CardContent>
					<div class="flex flex-col gap-3">
						<div class="flex items-end gap-1.5 h-24">
							{#each chartSeries as value, index}
								<div
									class="flex-1 bg-gradient-to-t from-blue-400 to-blue-300 rounded-t-md transition-all duration-300 hover:opacity-80"
									style={`height: ${(value / chartMax) * 80}%`}
								>
									<span class="sr-only">Day {index + 1}: {value}</span>
								</div>
							{/each}
						</div>
						<p class="text-sm text-slate-500">Signal derived from sales velocity, lead times, and safety stock buffers.</p>
					</div>
				</CardContent>
			</Card>

			<Card class="bg-white border border-slate-100 shadow-sm hover:shadow-md transition-all duration-200">
				<CardHeader>
					<CardTitle class="text-slate-800">Quick actions</CardTitle>
					<CardDescription class="text-slate-500">Keep the network in sync</CardDescription>
				</CardHeader>
				<CardContent class="space-y-3">
					<Button class="w-full justify-between bg-emerald-500 hover:bg-emerald-600 text-white transition-all duration-200" href="/operations">
						Balance stock <TrendingUp class="h-4 w-4" />
					</Button>
					<Button class="w-full justify-between bg-blue-500 hover:bg-blue-600 text-white transition-all duration-200" href="/intelligence">
						Trigger forecast <Activity class="h-4 w-4" />
					</Button>
					<Button class="w-full justify-between bg-indigo-500 hover:bg-indigo-600 text-white transition-all duration-200" href="/bulk">
						Bulk export catalog <Boxes class="h-4 w-4" />
					</Button>
				</CardContent>
			</Card>
		</div>

		<!-- Fresh Inventory + Priority Alerts -->
		<div class="grid gap-6 lg:grid-cols-2">
			<!-- Fresh Inventory -->
			<Card class="bg-white border border-slate-100 shadow-sm hover:shadow-md transition-all duration-200">
				<CardHeader>
					<CardTitle class="text-slate-800">Fresh inventory</CardTitle>
					<CardDescription class="text-slate-500">Latest SKUs created or touched</CardDescription>
				</CardHeader>
				<CardContent>
					<Table>
						<TableHeader>
							<TableRow class="border-b border-slate-100">
								<TableHead class="text-slate-600 font-medium">SKU</TableHead>
								<TableHead class="text-slate-600 font-medium">Title</TableHead>
								<TableHead class="text-slate-600 font-medium">Status</TableHead>
							</TableRow>
						</TableHeader>
						<TableBody>
							{#if loading}
								{#each Array(3) as _, i}
									<TableRow>
										<TableCell colspan="3"><Skeleton class="h-6 w-full bg-slate-100" /></TableCell>
									</TableRow>
								{/each}
							{:else if recentProducts.length === 0}
								<TableRow>
									<TableCell colspan="3" class="text-center text-sm text-slate-500 py-4">No recent changes</TableCell>
								</TableRow>
							{:else}
								{#each recentProducts as product}
									<TableRow class="hover:bg-slate-50 transition-colors">
										<TableCell class="font-mono text-xs text-slate-700">{product.SKU}</TableCell>
										<TableCell class="text-slate-700">{product.Name}</TableCell>
										<TableCell>
											<span class="rounded-full bg-blue-50 text-blue-600 px-2 py-1 text-xs">
												{product.Status ?? 'active'}
											</span>
										</TableCell>
									</TableRow>
								{/each}
							{/if}
						</TableBody>
					</Table>
				</CardContent>
			</Card>

			<!-- Priority Alerts -->
			<Card class="bg-white border border-slate-100 shadow-sm hover:shadow-md transition-all duration-200">
				<CardHeader>
					<CardTitle class="text-slate-800">Priority alerts</CardTitle>
					<CardDescription class="text-slate-500">Signals requiring intervention</CardDescription>
				</CardHeader>
				<CardContent>
					<Table>
						<TableHeader>
							<TableRow class="border-b border-slate-100">
								<TableHead class="text-slate-600 font-medium">Type</TableHead>
								<TableHead class="text-slate-600 font-medium">SKU</TableHead>
								<TableHead class="text-slate-600 font-medium">State</TableHead>
							</TableRow>
						</TableHeader>
						<TableBody>
							{#if loading}
								{#each Array(3) as _, i}
									<TableRow>
										<TableCell colspan="3"><Skeleton class="h-6 w-full bg-slate-100" /></TableCell>
									</TableRow>
								{/each}
							{:else if recentAlerts.length === 0}
								<TableRow>
									<TableCell colspan="3" class="text-center text-sm text-slate-500 py-4">All clear</TableCell>
								</TableRow>
							{:else}
								{#each recentAlerts as item}
									<TableRow class="hover:bg-slate-50 transition-colors">
										<TableCell class="text-xs font-medium text-slate-700">{item.Type}</TableCell>
										<TableCell class="text-slate-700">{item.Product?.SKU ?? item.ProductID}</TableCell>
										<TableCell>
											<span class="rounded-full bg-rose-50 text-rose-600 px-2 py-1 text-xs">
												{item.Status}
											</span>
										</TableCell>
									</TableRow>
								{/each}
							{/if}
						</TableBody>
					</Table>
				</CardContent>
			</Card>
		</div>

		<!-- Procurement Intelligence -->
		<Card class="bg-white border border-slate-100 shadow-sm hover:shadow-md transition-all duration-200">
			<CardHeader>
				<CardTitle class="text-slate-800">Procurement intelligence</CardTitle>
				<CardDescription class="text-slate-500">Top reorder suggestions waiting for conversion</CardDescription>
			</CardHeader>
			<CardContent>
				<Table>
					<TableHeader>
						<TableRow class="border-b border-slate-100">
							<TableHead class="text-slate-600 font-medium">Product</TableHead>
							<TableHead class="text-slate-600 font-medium">Suggested Qty</TableHead>
							<TableHead class="text-slate-600 font-medium">Supplier</TableHead>
							<TableHead class="text-slate-600 font-medium">Status</TableHead>
						</TableRow>
					</TableHeader>
					<TableBody>
						{#if loading}
							{#each Array(4) as _, i}
								<TableRow>
									<TableCell colspan="4"><Skeleton class="h-6 w-full bg-slate-100" /></TableCell>
								</TableRow>
							{/each}
						{:else if suggestions.length === 0}
							<TableRow>
								<TableCell colspan="4" class="text-center text-sm text-slate-500 py-4">No pending suggestions</TableCell>
							</TableRow>
						{:else}
							{#each suggestions as suggestion}
								<TableRow class="hover:bg-slate-50 transition-colors">
									<TableCell class="text-slate-700">{suggestion?.Product?.Name ?? `Product ${suggestion?.ProductID ?? 'N/A'}`}</TableCell>
									<TableCell class="text-slate-700">{suggestion?.SuggestedOrderQuantity ?? 'N/A'}</TableCell>
									<TableCell class="text-slate-700">{suggestion?.Supplier?.Name ?? suggestion?.SupplierID ?? 'N/A'}</TableCell>
									<TableCell>
										<span class="rounded-full bg-emerald-50 text-emerald-600 px-2 py-1 text-xs capitalize">
											{suggestion.Status}
										</span>
									</TableCell>
								</TableRow>
							{/each}
						{/if}
					</TableBody>
				</Table>
			</CardContent>
		</Card>
	</section>
</div>

<style>
	/* Smooth transitions for all elements */
	* {
		transition-property: color, background-color, border-color, opacity, box-shadow, transform;
		transition-timing-function: cubic-bezier(0.4, 0, 0.2, 1);
		transition-duration: 200ms;
	}
</style>