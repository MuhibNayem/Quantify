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
    import { t } from '$lib/i18n';

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
			toast.error($t('intelligence.toasts.select_product'));
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
			toast.success($t('intelligence.toasts.forecast_success'));
		} catch (error: any) {
			toast.error($t('intelligence.toasts.forecast_fail'), {
				description: error?.response?.data?.error || 'An unexpected error occurred'
			});
		} finally {
			loading = false;
		}
	}
</script>

<Card class="h-full border-border bg-card shadow-sm">
	<CardHeader>
		<CardTitle class="flex items-center gap-2 text-foreground">
			<div class="rounded-md bg-muted p-2 text-primary">
				<TrendingUp class="h-5 w-5" />
			</div>
			{$t('intelligence.demand_forecast.title')}
		</CardTitle>
		<CardDescription>{$t('intelligence.demand_forecast.subtitle')}</CardDescription>
	</CardHeader>
	<CardContent class="space-y-4">
        <!-- Product Search -->
		<div class="space-y-2">
			<Label>{$t('intelligence.demand_forecast.select_product')}</Label>
            {#if selectedProduct}
                <div class="flex items-center justify-between rounded-md border border-border bg-muted p-2">
                    <div class="flex flex-col">
                        <span class="font-medium text-foreground">{selectedProduct.Name}</span>
                        <span class="text-xs text-muted-foreground">SKU: {selectedProduct.SKU}</span>
                    </div>
                    <Button variant="ghost" size="sm" class="h-8 w-8 p-0 text-muted-foreground hover:text-foreground" onclick={() => { selectedProduct = null; productId = null; forecast = null; }}>
                        X
                    </Button>
                </div>
            {:else}
                <div class="relative">
                    <Search class="absolute left-2.5 top-2.5 h-4 w-4 text-muted-foreground" />
                    <Input
                        placeholder={$t('intelligence.demand_forecast.placeholder')}
                        class="pl-9"
                        bind:value={productSearch}
                        oninput={searchProducts}
                    />
                    {#if searchResults.length > 0}
                        <div class="absolute z-10 mt-1 w-full rounded-md border border-border bg-card shadow-lg">
                            {#each searchResults as product}
                                <button
                                    class="w-full px-4 py-2 text-left text-sm transition-colors hover:bg-muted"
                                    onclick={() => selectProduct(product)}
                                >
                                    <div class="font-medium text-foreground">{product.Name}</div>
                                    <div class="text-xs text-muted-foreground">{product.SKU}</div>
                                </button>
                            {/each}
                        </div>
                    {/if}
                </div>
            {/if}
		</div>

		<div class="flex items-end gap-4">
			<div class="flex-1 space-y-2">
				<Label>{$t('intelligence.demand_forecast.period_label')}</Label>
				<Input type="number" bind:value={days} min="7" max="90" />
			</div>
			<Button onclick={generateForecast} disabled={loading || !productId}>
				{#if loading}
					<Loader2 class="mr-2 h-4 w-4 animate-spin" />
					{$t('intelligence.demand_forecast.generating_btn')}
				{:else}
					{$t('intelligence.demand_forecast.generate_btn')}
				{/if}
			</Button>
		</div>

		{#if forecast}
			<div class="mt-4 rounded-lg border border-border bg-card p-4 shadow-sm animate-in fade-in slide-in-from-bottom-4">
				<div class="mb-4 flex items-center justify-between">
					<div>
						<div class="overline">{$t('intelligence.demand_forecast.predicted_demand')}</div>
						<div class="kpi-num text-2xl text-foreground">{forecast.PredictedDemand} units</div>
					</div>
                    {#if forecast.ConfidenceScore}
                        <div class="text-right">
                            <div class="overline">{$t('intelligence.demand_forecast.confidence')}</div>
                            <div class="num flex items-center justify-end gap-1 font-medium text-success">
                                <CheckCircle2 class="h-4 w-4" />
                                {(forecast.ConfidenceScore * 100).toFixed(0)}%
                            </div>
                        </div>
                    {/if}
				</div>

                {#if forecast.Reasoning}
                    <div class="rounded-md bg-muted p-3 text-sm text-foreground">
                        <div class="mb-1 font-medium flex items-center gap-2">
                            <AlertCircle class="h-3 w-3 text-info" /> {$t('intelligence.demand_forecast.reasoning')}
                        </div>
                        {forecast.Reasoning}
                    </div>
                {/if}

				<div class="mt-2 text-xs text-muted-foreground">
					{$t('intelligence.demand_forecast.generated_at')} {new Date(forecast.GeneratedAt).toLocaleString()}
				</div>
			</div>
		{/if}
	</CardContent>
</Card>
