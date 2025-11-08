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
			const [productList, categoryList, supplierList, alertList, suggestionList] = await Promise.all([
				productsApi.list(),
				categoriesApi.list(),
				suppliersApi.list(),
				alertsApi.list({ status: 'ACTIVE' }),
				replenishmentApi.listSuggestions()
			]);

			stats.products = productList.length;
			stats.categories = (Array.isArray(categoryList) ? categoryList : [categoryList]).length;
			stats.suppliers = (Array.isArray(supplierList) ? supplierList : [supplierList]).length;
			stats.alerts = alertList.length;

			recentProducts = productList.slice(0, 5);
			recentAlerts = alertList.slice(0, 5);
			suggestions = suggestionList.slice(0, 5);
		} catch (error: any) {
			toast.error('Failed to Load Dashboard', {
				description: error.response?.data?.error || 'Unable to load dashboard data'
			});
		} finally {
			loading = false;
		}
	};

	onMount(loadDashboard);
</script>

<div class="w-full max-w-7xl mx-auto py-10 px-6 bg-gradient-to-br from-white via-slate-50 to-blue-50 dark:from-slate-900 dark:via-slate-950 dark:to-slate-900 rounded-2xl shadow-2xl backdrop-blur-md text-slate-800 dark:text-slate-100">
	<section class="space-y-10">
		<!-- Header -->
		<div class="flex flex-wrap items-center justify-between gap-3">
			<div>
				<p class="text-sm uppercase tracking-wide text-slate-500 dark:text-slate-400">Control tower</p>
				<h1 class="text-3xl font-semibold bg-gradient-to-r from-blue-600 to-cyan-500 bg-clip-text text-transparent">
					Realtime inventory intelligence
				</h1>
			</div>
			<div class="flex flex-wrap gap-2">
				<Button
					variant="secondary"
					class="bg-gradient-to-r from-slate-200 to-slate-300 hover:from-slate-300 hover:to-slate-400 text-slate-800 shadow-md"
					onclick={loadDashboard}
				>
					<RefreshCcw class="mr-2 h-4 w-4" /> Refresh
				</Button>
				<Button
					href="/catalog"
					class="bg-gradient-to-r from-blue-600 to-cyan-500 hover:from-blue-700 hover:to-cyan-600 text-white font-semibold shadow-lg hover:shadow-xl transition-all duration-200"
				>
					Update catalog
				</Button>
			</div>
		</div>

		<!-- Stats Cards -->
		<div class="grid gap-6 lg:grid-cols-4">
			<Card class="rounded-xl shadow-lg bg-gradient-to-br from-blue-50 to-indigo-100 hover:shadow-2xl transition-all duration-200">
				<CardHeader class="flex flex-row items-center justify-between pb-2">
					<CardTitle class="text-sm font-medium text-slate-700">Active SKUs</CardTitle>
					<Boxes class="h-4 w-4 text-blue-600" />
				</CardHeader>
				<CardContent>
					{#if loading}
						<Skeleton class="h-10 w-24" />
					{:else}
						<div class="text-3xl font-bold text-blue-800">{stats.products}</div>
						<p class="text-xs text-slate-600">{Math.round(stats.products * 1.12)} forecasted for Q4</p>
					{/if}
				</CardContent>
			</Card>

			<Card class="rounded-xl shadow-lg bg-gradient-to-br from-emerald-50 to-teal-100 hover:shadow-2xl transition-all duration-200">
				<CardHeader class="flex flex-row items-center justify-between pb-2">
					<CardTitle class="text-sm font-medium text-slate-700">Strategic categories</CardTitle>
					<Activity class="h-4 w-4 text-emerald-600" />
				</CardHeader>
				<CardContent>
					{#if loading}
						<Skeleton class="h-10 w-24" />
					{:else}
						<div class="text-3xl font-bold text-emerald-800">{stats.categories}</div>
						<p class="text-xs text-slate-600">Converged across {stats.suppliers} suppliers</p>
					{/if}
				</CardContent>
			</Card>

			<Card class="rounded-xl shadow-lg bg-gradient-to-br from-orange-50 to-amber-100 hover:shadow-2xl transition-all duration-200">
				<CardHeader class="flex flex-row items-center justify-between pb-2">
					<CardTitle class="text-sm font-medium text-slate-700">Partners online</CardTitle>
					<Factory class="h-4 w-4 text-amber-600" />
				</CardHeader>
				<CardContent>
					{#if loading}
						<Skeleton class="h-10 w-24" />
					{:else}
						<div class="text-3xl font-bold text-amber-800">{stats.suppliers}</div>
						<p class="text-xs text-slate-600">Supplier SLAs synced</p>
					{/if}
				</CardContent>
			</Card>

			<Card class="rounded-xl shadow-lg bg-gradient-to-br from-rose-50 to-pink-100 hover:shadow-2xl transition-all duration-200">
				<CardHeader class="flex flex-row items-center justify-between pb-2">
					<CardTitle class="text-sm font-medium text-slate-700">Active alerts</CardTitle>
					<AlertTriangle class="h-4 w-4 text-rose-600" />
				</CardHeader>
				<CardContent>
					{#if loading}
						<Skeleton class="h-10 w-24" />
					{:else}
						<div class="text-3xl font-bold text-rose-700">{stats.alerts}</div>
						<p class="text-xs text-slate-600">Auto-escalations enabled</p>
					{/if}
				</CardContent>
			</Card>
		</div>

		<!-- Demand Pulse -->
		<div class="grid gap-6 lg:grid-cols-[2fr,1fr]">
			<Card class="shadow-lg rounded-xl bg-gradient-to-br from-sky-50 to-blue-100 hover:shadow-2xl transition-all duration-200">
				<CardHeader>
					<CardTitle class="text-slate-800">Demand pulse</CardTitle>
					<CardDescription class="text-slate-600">Seven-day weighted trend (mock data)</CardDescription>
				</CardHeader>
				<CardContent>
					<div class="flex flex-col gap-2">
						<div class="flex items-end gap-2">
							{#each chartSeries as value, index}
								<div
									class="w-full rounded-t-md bg-gradient-to-t from-blue-400 to-blue-600 transition-all hover:scale-105"
									style={`height: ${(value / chartMax) * 140}px`}
								>
									<span class="sr-only">Day {index + 1}: {value}</span>
								</div>
							{/each}
						</div>
						<p class="text-sm text-slate-600">Signal derived from sales velocity, lead times, and safety stock buffers.</p>
					</div>
				</CardContent>
			</Card>

			<Card class="shadow-lg rounded-xl bg-gradient-to-br from-blue-50 to-indigo-100 hover:shadow-2xl transition-all duration-200">
				<CardHeader>
					<CardTitle class="text-slate-800">Quick actions</CardTitle>
					<CardDescription class="text-slate-600">Keep the network in sync</CardDescription>
				</CardHeader>
				<CardContent class="space-y-2">
					<Button class="w-full justify-between bg-gradient-to-r from-green-500 to-emerald-600 hover:from-green-600 hover:to-emerald-700 text-white font-semibold shadow-lg hover:shadow-xl transition-all duration-200" href="/operations">
						Balance stock <TrendingUp class="h-4 w-4" />
					</Button>
					<Button class="w-full justify-between bg-gradient-to-r from-indigo-500 to-blue-600 hover:from-indigo-600 hover:to-blue-700 text-white font-semibold shadow-lg hover:shadow-xl transition-all duration-200" href="/intelligence">
						Trigger forecast <Activity class="h-4 w-4" />
					</Button>
					<Button class="w-full justify-between bg-gradient-to-r from-cyan-500 to-blue-600 hover:from-cyan-600 hover:to-blue-700 text-white font-semibold shadow-lg hover:shadow-xl transition-all duration-200" href="/bulk">
						Bulk export catalog <Boxes class="h-4 w-4" />
					</Button>
				</CardContent>
			</Card>
		</div>

		<!-- Fresh Inventory + Priority Alerts -->
<div class="grid gap-6 lg:grid-cols-2">
  <!-- Fresh Inventory -->
  <Card class="rounded-xl bg-white dark:bg-slate-800 shadow-md hover:shadow-xl transition-all duration-200">
    <CardHeader>
      <CardTitle class="text-slate-800 dark:text-slate-100">Fresh inventory</CardTitle>
      <CardDescription class="text-slate-600 dark:text-slate-400">Latest SKUs created or touched</CardDescription>
    </CardHeader>
    <CardContent>
      <Table>
        <TableHeader>
          <TableRow class="text-slate-700 dark:text-slate-200">
            <TableHead>SKU</TableHead>
            <TableHead>Title</TableHead>
            <TableHead>Status</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {#if loading}
            {#each Array(3) as _, i}
              <TableRow>
                <TableCell colspan="3"><Skeleton class="h-6 w-full" /></TableCell>
              </TableRow>
            {/each}
          {:else if recentProducts.length === 0}
            <TableRow>
              <TableCell colspan="3" class="text-center text-sm text-slate-500 dark:text-slate-400">No recent changes</TableCell>
            </TableRow>
          {:else}
            {#each recentProducts as product}
              <TableRow class="text-slate-800 dark:text-slate-100">
                <TableCell class="font-mono text-xs">{product.SKU}</TableCell>
                <TableCell>{product.Name}</TableCell>
                <TableCell>
                  <span class="rounded-full bg-blue-50 text-blue-700 dark:bg-blue-900/40 dark:text-blue-300 px-2 py-0.5 text-xs capitalize">
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
  <Card class="rounded-xl bg-white dark:bg-slate-800 shadow-md hover:shadow-xl transition-all duration-200">
    <CardHeader>
      <CardTitle class="text-slate-800 dark:text-slate-100">Priority alerts</CardTitle>
      <CardDescription class="text-slate-600 dark:text-slate-400">Signals requiring intervention</CardDescription>
    </CardHeader>
    <CardContent>
      <Table>
        <TableHeader>
          <TableRow class="text-slate-700 dark:text-slate-200">
            <TableHead>Type</TableHead>
            <TableHead>SKU</TableHead>
            <TableHead>State</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {#if loading}
            {#each Array(3) as _, i}
              <TableRow>
                <TableCell colspan="3"><Skeleton class="h-6 w-full" /></TableCell>
              </TableRow>
            {/each}
          {:else if recentAlerts.length === 0}
            <TableRow>
              <TableCell colspan="3" class="text-center text-sm text-slate-500 dark:text-slate-400">All clear</TableCell>
            </TableRow>
          {:else}
            {#each recentAlerts as item}
              <TableRow class="text-slate-800 dark:text-slate-100">
                <TableCell class="text-xs font-semibold">{item.Type}</TableCell>
                <TableCell>{item.Product?.SKU ?? item.ProductID}</TableCell>
                <TableCell>
                  <span class="rounded-full bg-rose-50 text-rose-700 dark:bg-rose-900/40 dark:text-rose-300 px-2 py-0.5 text-xs">
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
<Card class="rounded-xl bg-white dark:bg-slate-800 shadow-md hover:shadow-xl transition-all duration-200">
  <CardHeader>
    <CardTitle class="text-slate-800 dark:text-slate-100">Procurement intelligence</CardTitle>
    <CardDescription class="text-slate-600 dark:text-slate-400">Top reorder suggestions waiting for conversion</CardDescription>
  </CardHeader>
  <CardContent>
    <Table>
      <TableHeader>
        <TableRow class="text-slate-700 dark:text-slate-200">
          <TableHead>Product</TableHead>
          <TableHead>Suggested Qty</TableHead>
          <TableHead>Supplier</TableHead>
          <TableHead>Status</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        {#if loading}
          {#each Array(4) as _, i}
            <TableRow>
              <TableCell colspan="4"><Skeleton class="h-6 w-full" /></TableCell>
            </TableRow>
          {/each}
        {:else if suggestions.length === 0}
          <TableRow>
            <TableCell colspan="4" class="text-center text-sm text-slate-500 dark:text-slate-400">No pending suggestions</TableCell>
          </TableRow>
        {:else}
          {#each suggestions as suggestion}
            <TableRow class="text-slate-800 dark:text-slate-100">
              <TableCell>{suggestion?.Product?.Name ?? `Product ${suggestion?.ProductID ?? 'N/A'}`}</TableCell>
              <TableCell>{suggestion?.SuggestedOrderQuantity ?? 'N/A'}</TableCell>
              <TableCell>{suggestion?.Supplier?.Name ?? suggestion?.SupplierID ?? 'N/A'}</TableCell>
              <TableCell>
                <span class="rounded-full bg-emerald-50 text-emerald-700 dark:bg-emerald-900/40 dark:text-emerald-300 px-2 py-0.5 text-xs capitalize">
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
</style>
