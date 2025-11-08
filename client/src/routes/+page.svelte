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
				replenishmentApi.listSuggestions(),
			]);
			stats.products = productList.length;
			stats.categories = (Array.isArray(categoryList) ? categoryList : [categoryList]).length;
			stats.suppliers = (Array.isArray(supplierList) ? supplierList : [supplierList]).length;
			stats.alerts = alertList.length;
			recentProducts = productList.slice(0, 5);
			recentAlerts = alertList.slice(0, 5);
			suggestions = suggestionList.slice(0, 5);
		} catch (error: any) {
			const errorMessage = error.response?.data?.error || 'Unable to load dashboard data';
			toast.error('Failed to Load Dashboard', {
				description: errorMessage,
			});
		} finally {
			loading = false;
		}
	};

	onMount(loadDashboard);
</script>

<div class="w-full max-w-7xl mx-auto py-8 px-4 sm:px-6 lg:px-8 bg-white dark:bg-slate-900 shadow-xl rounded-2xl">
  <section class="space-y-8">
    <div class="flex flex-wrap items-center justify-between gap-3">
      <div>
        <p class="text-sm uppercase tracking-wide text-muted-foreground">Control tower</p>
        <h1 class="text-3xl font-semibold">Realtime inventory intelligence</h1>
      </div>
      <div class="flex flex-wrap gap-2">
        <Button variant="secondary" onclick={loadDashboard}>
          <RefreshCcw class="mr-2 h-4 w-4" /> Refresh
        </Button>
        <Button href="/catalog" class="bg-gradient-to-r from-blue-600 to-blue-700 hover:from-blue-700 hover:to-blue-800 text-white font-semibold shadow-lg hover:shadow-xl transition-all duration-200">Update catalog</Button>
      </div>
    </div>

    <div class="grid gap-4 lg:grid-cols-4">
      <Card class="shadow-lg rounded-xl">
        <CardHeader class="flex flex-row items-center justify-between space-y-0 pb-2">
          <CardTitle class="text-sm font-medium">Active SKUs</CardTitle>
          <Boxes class="h-4 w-4 text-muted-foreground" />
        </CardHeader>
        <CardContent>
          {#if loading}
            <Skeleton class="h-10 w-24" />
          {:else}
            <div class="text-3xl font-bold">{stats.products}</div>
            <p class="text-xs text-muted-foreground">{Math.round(stats.products * 1.12)} forecasted for Q4</p>
          {/if}
        </CardContent>
      </Card>

      <Card class="shadow-lg rounded-xl">
        <CardHeader class="flex flex-row items-center justify-between space-y-0 pb-2">
          <CardTitle class="text-sm font-medium">Strategic categories</CardTitle>
          <Activity class="h-4 w-4 text-muted-foreground" />
        </CardHeader>
        <CardContent>
          {#if loading}
            <Skeleton class="h-10 w-24" />
          {:else}
            <div class="text-3xl font-bold">{stats.categories}</div>
            <p class="text-xs text-muted-foreground">Converged across {stats.suppliers} suppliers</p>
          {/if}
        </CardContent>
      </Card>

      <Card class="shadow-lg rounded-xl">
        <CardHeader class="flex flex-row items-center justify-between space-y-0 pb-2">
          <CardTitle class="text-sm font-medium">Partners online</CardTitle>
          <Factory class="h-4 w-4 text-muted-foreground" />
        </CardHeader>
        <CardContent>
          {#if loading}
            <Skeleton class="h-10 w-24" />
          {:else}
            <div class="text-3xl font-bold">{stats.suppliers}</div>
            <p class="text-xs text-muted-foreground">Supplier SLAs synced</p>
          {/if}
        </CardContent>
      </Card>

      <Card class="bg-destructive/5 shadow-lg rounded-xl">
        <CardHeader class="flex flex-row items-center justify-between space-y-0 pb-2">
          <CardTitle class="text-sm font-medium">Active alerts</CardTitle>
          <AlertTriangle class="h-4 w-4 text-destructive" />
        </CardHeader>
        <CardContent>
          {#if loading}
            <Skeleton class="h-10 w-24" />
          {:else}
            <div class="text-3xl font-bold text-destructive">{stats.alerts}</div>
            <p class="text-xs text-muted-foreground">Auto-escalations enabled</p>
          {/if}
        </CardContent>
      </Card>
    </div>

    <div class="grid gap-6 lg:grid-cols-[2fr,1fr]">
      <Card class="shadow-lg rounded-xl">
        <CardHeader>
          <CardTitle>Demand pulse</CardTitle>
          <CardDescription>Seven-day weighted trend (mock data)</CardDescription>
        </CardHeader>
        <CardContent>
          <div class="flex flex-col gap-2">
            <div class="flex items-end gap-2">
              {#each chartSeries as value, index}
                <div
                  class="w-full rounded-t-md bg-primary/70"
                  style={`height: ${(value / chartMax) * 140}px`}
                >
                  <span class="sr-only">Day {index + 1}: {value}</span>
                </div>
              {/each}
            </div>
            <p class="text-sm text-muted-foreground">Signal derived from sales velocity, lead times, and safety stock buffers.</p>
          </div>
        </CardContent>
      </Card>
      <Card class="shadow-lg rounded-xl">
        <CardHeader>
          <CardTitle>Quick actions</CardTitle>
          <CardDescription>Keep the network in sync</CardDescription>
        </CardHeader>
        <CardContent class="space-y-2">
          <Button class="w-full justify-between bg-gradient-to-r from-blue-600 to-blue-700 hover:from-blue-700 hover:to-blue-800 text-white font-semibold shadow-lg hover:shadow-xl transition-all duration-200" href="/operations">
            Balance stock <TrendingUp class="h-4 w-4" />
          </Button>
          <Button class="w-full justify-between bg-gradient-to-r from-blue-600 to-blue-700 hover:from-blue-700 hover:to-blue-800 text-white font-semibold shadow-lg hover:shadow-xl transition-all duration-200" href="/intelligence">
            Trigger forecast <Activity class="h-4 w-4" />
          </Button>
          <Button class="w-full justify-between bg-gradient-to-r from-blue-600 to-blue-700 hover:from-blue-700 hover:to-blue-800 text-white font-semibold shadow-lg hover:shadow-xl transition-all duration-200" href="/bulk">
            Bulk export catalog <Boxes class="h-4 w-4" />
          </Button>
        </CardContent>
      </Card>
    </div>

    <div class="grid gap-6 lg:grid-cols-2">
        <Card class="shadow-lg rounded-xl">
          <CardHeader>
            <CardTitle>Fresh inventory</CardTitle>
            <CardDescription>Latest SKUs created or touched</CardDescription>
          </CardHeader>
          <CardContent>
            <Table>
              <TableHeader>
                <TableRow>
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
                    <TableCell colspan="3" class="text-center text-sm text-muted-foreground">No recent changes</TableCell>
                  </TableRow>
                {:else}
                  {#each recentProducts as product}
                    <TableRow>
                      <TableCell class="font-mono text-xs">{product.SKU}</TableCell>
                      <TableCell>{product.Name}</TableCell>
                      <TableCell>
                        <span class="rounded-full bg-muted px-2 py-0.5 text-xs capitalize">{product.Status ?? 'active'}</span>
                      </TableCell>
                    </TableRow>
                  {/each}
                {/if}
              </TableBody>
            </Table>
          </CardContent>
        </Card>

        <Card class="shadow-lg rounded-xl">
          <CardHeader>
            <CardTitle>Priority alerts</CardTitle>
            <CardDescription>Signals requiring intervention</CardDescription>
          </CardHeader>
          <CardContent>
            <Table>
              <TableHeader>
                <TableRow>
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
                    <TableCell colspan="3" class="text-center text-sm text-muted-foreground">All clear</TableCell>
                  </TableRow>
                {:else}
                  {#each recentAlerts as item}
                    <TableRow>
                      <TableCell class="text-xs font-semibold">{item.Type}</TableCell>
                      <TableCell>{item.Product?.SKU ?? item.ProductID}</TableCell>
                      <TableCell>
                        <span class="rounded-full bg-destructive/10 px-2 py-0.5 text-xs text-destructive">{item.Status}</span>
                      </TableCell>
                    </TableRow>
                  {/each}
                {/if}
              </TableBody>
            </Table>
          </CardContent>
        </Card>
      </div>


      <Card class="shadow-lg rounded-xl">
        <CardHeader>
          <CardTitle>Procurement intelligence</CardTitle>
          <CardDescription>Top reorder suggestions waiting for conversion</CardDescription>
        </CardHeader>
        <CardContent>
          <Table>
            <TableHeader>
              <TableRow>
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
                  <TableCell colspan="4" class="text-center text-sm text-muted-foreground">No pending suggestions</TableCell>
                </TableRow>
              {:else}
                {#each suggestions as suggestion}
                  <TableRow>
                    <TableCell>{suggestion?.Product?.Name ?? `Product ${suggestion?.ProductID ?? 'N/A'}`}</TableCell>
                    <TableCell>{suggestion?.SuggestedOrderQuantity ?? 'N/A'}</TableCell>
                    <TableCell>{suggestion?.Supplier?.Name ?? suggestion?.SupplierID ?? 'N/A'}</TableCell>
                    <TableCell>
                      <span class="rounded-full bg-primary/10 px-2 py-0.5 text-xs capitalize text-primary">{suggestion.Status}</span>
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