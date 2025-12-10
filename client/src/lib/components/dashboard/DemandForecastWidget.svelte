<script lang="ts">
	import { Button } from '$lib/components/ui/button';
	import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '$lib/components/ui/card';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import { replenishmentApi, productsApi } from '$lib/api/resources';
	import type { DemandForecast, Product } from '$lib/types';
	import { Loader2, TrendingUp, Search, AlertCircle, CheckCircle2 } from 'lucide-svelte';
	import { toast } from 'svelte-sonner';
    import { onMount } from 'svelte';

	let loading = false;
	let productId: number | null = null;
    let productSearch = '';
    let searchResults: Product[] = [];
    let selectedProduct: Product | null = null;
    let searching = false;

	let forecast: DemandForecast | null = null;
	let days = 30;

    async function searchProducts() {
        if (productSearch.length < 2) return;
        searching = true;
        try {
            const res = await productsApi.list(1, 5, productSearch);
            searchResults = res.products;
        } catch (e) {
            console.error(e);
        } finally {
            searching = false;
        }
    }

    function selectProduct(product: Product) {
        selectedProduct = product;
        productId = product.ID;
        searchResults = [];
        productSearch = '';
    }

	async function generateForecast() {
		if (!productId) {
			toast.error('Please select a product');
			return;
		}

		loading = true;
		forecast = null;
		try {
			const res = await replenishmentApi.generateForecast({
				productId: productId,
				periodInDays: days
			});
            // Backend returns { message, forecast }
			forecast = res.forecast;
			toast.success('Forecast generated successfully');
		} catch (error: any) {
			toast.error('Failed to generate forecast', {
				description: error?.response?.data?.error || 'An unexpected error occurred'
			});
		} finally {
			loading = false;
		}
	}
</script>

<Card class="h-full border-blue-100 bg-gradient-to-br from-white to-blue-50/50 shadow-sm">
	<CardHeader>
		<CardTitle class="flex items-center gap-2 text-slate-800">
			<div class="rounded-lg bg-blue-100 p-2 text-blue-600">
				<TrendingUp class="h-5 w-5" />
			</div>
			AI Demand Forecast
		</CardTitle>
		<CardDescription>Predict future demand for specific products</CardDescription>
	</CardHeader>
	<CardContent class="space-y-4">
        <!-- Product Search -->
		<div class="space-y-2">
			<Label>Select Product</Label>
            {#if selectedProduct}
                <div class="flex items-center justify-between rounded-md border border-blue-200 bg-blue-50 p-2">
                    <div class="flex flex-col">
                        <span class="font-medium text-blue-900">{selectedProduct.Name}</span>
                        <span class="text-xs text-blue-700">SKU: {selectedProduct.SKU}</span>
                    </div>
                    <Button variant="ghost" size="sm" class="h-8 w-8 p-0 text-blue-500 hover:text-blue-700" onclick={() => { selectedProduct = null; productId = null; forecast = null; }}>
                        X
                    </Button>
                </div>
            {:else}
                <div class="relative">
                    <Search class="absolute left-2.5 top-2.5 h-4 w-4 text-slate-500" />
                    <Input 
                        placeholder="Search by name or SKU..." 
                        class="pl-9" 
                        bind:value={productSearch}
                        oninput={searchProducts}
                    />
                    {#if searchResults.length > 0}
                        <div class="absolute z-10 mt-1 w-full rounded-md border border-slate-200 bg-white shadow-lg">
                            {#each searchResults as product}
                                <button 
                                    class="w-full px-4 py-2 text-left text-sm hover:bg-slate-50"
                                    onclick={() => selectProduct(product)}
                                >
                                    <div class="font-medium">{product.Name}</div>
                                    <div class="text-xs text-slate-500">{product.SKU}</div>
                                </button>
                            {/each}
                        </div>
                    {/if}
                </div>
            {/if}
		</div>

		<div class="flex items-end gap-4">
			<div class="flex-1 space-y-2">
				<Label>Forecast Period (Days)</Label>
				<Input type="number" bind:value={days} min="7" max="90" />
			</div>
			<Button onclick={generateForecast} disabled={loading || !productId} class="bg-blue-600 hover:bg-blue-700">
				{#if loading}
					<Loader2 class="mr-2 h-4 w-4 animate-spin" />
					Generating...
				{:else}
					Generate
				{/if}
			</Button>
		</div>

		{#if forecast}
			<div class="mt-4 rounded-lg border border-slate-200 bg-white p-4 shadow-sm animate-in fade-in slide-in-from-bottom-4">
				<div class="mb-4 flex items-center justify-between">
					<div>
						<div class="text-sm text-slate-500">Predicted Demand</div>
						<div class="text-2xl font-bold text-slate-900">{forecast.PredictedDemand} units</div>
					</div>
                    {#if forecast.ConfidenceScore}
                        <div class="text-right">
                            <div class="text-sm text-slate-500">Confidence</div>
                            <div class="flex items-center gap-1 font-medium text-emerald-600">
                                <CheckCircle2 class="h-4 w-4" />
                                {(forecast.ConfidenceScore * 100).toFixed(0)}%
                            </div>
                        </div>
                    {/if}
				</div>
                
                {#if forecast.Reasoning}
                    <div class="rounded-md bg-slate-50 p-3 text-sm text-slate-700">
                        <div class="mb-1 font-medium text-slate-900 flex items-center gap-2">
                            <AlertCircle class="h-3 w-3" /> AI Reasoning
                        </div>
                        {forecast.Reasoning}
                    </div>
                {/if}
                
				<div class="mt-2 text-xs text-slate-400">
					Generated at {new Date(forecast.GeneratedAt).toLocaleString()}
				</div>
			</div>
		{/if}
	</CardContent>
</Card>
