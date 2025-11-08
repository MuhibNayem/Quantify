<script lang="ts">
	import { onMount } from 'svelte';
	import { toast } from 'svelte-sonner';
	import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '$lib/components/ui/card';
	import { Input } from '$lib/components/ui/input';
	import { Button } from '$lib/components/ui/button';
	import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '$lib/components/ui/table';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import { replenishmentApi, reportsApi } from '$lib/api/resources';
	import type { ReorderSuggestion } from '$lib/types';

	const forecastForm = $state({ periodInDays: '30', productId: '', result: '' });
	let suggestions = $state<ReorderSuggestion[]>([]);
	let suggestionsLoading = $state(false);

	const reportRange = $state({ startDate: new Date(Date.now() - 30 * 24 * 60 * 60 * 1000).toISOString().slice(0, 10), endDate: new Date().toISOString().slice(0, 10) });
	const reportKeys = ['sales', 'turnover', 'margin'] as const;
	type ReportKey = (typeof reportKeys)[number];
	const reportResults = $state<Record<ReportKey, unknown>>({ sales: null, turnover: null, margin: null });
	let reportsLoading = $state(false);

	const loadSuggestions = async () => {
		suggestionsLoading = true;
		try {
			suggestions = await replenishmentApi.listSuggestions();
		} catch (error) {
			const errorMessage = error.response?.data?.error || 'Unable to load suggestions';
			toast.error('Failed to Load Suggestions', {
				description: errorMessage,
			});
		} finally {
			suggestionsLoading = false;
		}
	};

	onMount(loadSuggestions);

	const triggerForecast = async () => {
		try {
			const payload: Record<string, unknown> = { periodInDays: Number(forecastForm.periodInDays) };
			if (forecastForm.productId) payload.productId = Number(forecastForm.productId);
			const response = await replenishmentApi.generateForecast(payload);
			forecastForm.result = response.message ?? 'Forecast generated';
			toast.success('Forecast completed');
			await loadSuggestions();
		} catch (error) {
			const errorMessage = error.response?.data?.error || 'Unable to run forecast';
			toast.error('Failed to Run Forecast', {
				description: errorMessage,
			});
		}
	};

	const createPO = async (suggestionId: number) => {
		try {
			const po = await replenishmentApi.createPOFromSuggestion(suggestionId);
			toast.success(`PO ${po.ID ?? 'created'}`);
			await loadSuggestions();
		} catch (error) {
			const errorMessage = error.response?.data?.error || 'Unable to create PO';
			toast.error('Failed to Create PO', {
				description: errorMessage,
			});
		}
	};

	const runReport = async (type: ReportKey) => {
		reportsLoading = true;
		const payload = {
			startDate: new Date(reportRange.startDate).toISOString(),
			endDate: new Date(reportRange.endDate).toISOString(),
		};
		try {
			if (type === 'sales') {
				reportResults.sales = await reportsApi.salesTrends(payload);
			} else if (type === 'turnover') {
				reportResults.turnover = await reportsApi.inventoryTurnover(payload);
			} else {
				reportResults.margin = await reportsApi.profitMargin(payload);
			}
			toast.success('Report ready');
		} catch (error) {
			const errorMessage = error.response?.data?.error || 'Unable to run report';
			toast.error('Failed to Run Report', {
				description: errorMessage,
			});
		} finally {
			reportsLoading = false;
		}
	};
</script>

<section class="space-y-8">
	<header>
		<p class="text-sm uppercase tracking-wide text-muted-foreground">Planning intelligence</p>
		<h1 class="text-3xl font-semibold">Forecasting, reorder suggestions & business reports</h1>
	</header>

	<div class="grid gap-6 lg:grid-cols-2">
		<Card>
			<CardHeader>
				<CardTitle>Demand forecast</CardTitle>
				<CardDescription>Trigger rolling forecasts for targeted SKUs</CardDescription>
			</CardHeader>
			<CardContent class="space-y-3">
				<Input type="number" min="7" placeholder="Horizon (days)" bind:value={forecastForm.periodInDays} />
				<Input type="number" min="1" placeholder="Product ID (optional)" bind:value={forecastForm.productId} />
				<Button class="w-full" onclick={triggerForecast}>Generate forecast</Button>
				{#if forecastForm.result}
					<p class="rounded-lg border border-border/60 bg-muted/30 p-3 text-sm">{forecastForm.result}</p>
				{/if}
			</CardContent>
		</Card>
		<Card>
			<CardHeader>
				<CardTitle>Report range</CardTitle>
				<CardDescription>Align analytics across shared horizon</CardDescription>
			</CardHeader>
			<CardContent class="space-y-3">
				<div class="grid grid-cols-2 gap-3">
					<Input type="date" bind:value={reportRange.startDate} />
					<Input type="date" bind:value={reportRange.endDate} />
				</div>
				<div class="flex flex-wrap gap-2">
					<Button class="flex-1" variant="secondary" onclick={() => runReport('sales')}>Sales trends</Button>
					<Button class="flex-1" variant="secondary" onclick={() => runReport('turnover')}>Inventory turnover</Button>
					<Button class="flex-1" variant="secondary" onclick={() => runReport('margin')}>Profit margin</Button>
				</div>
			</CardContent>
		</Card>
	</div>

	<Card>
		<CardHeader>
			<CardTitle>Reorder suggestions</CardTitle>
			<CardDescription>Convert high-signal suggestions into purchase orders</CardDescription>
		</CardHeader>
		<CardContent>
			<Table>
				<TableHeader>
					<TableRow>
						<TableHead>Product</TableHead>
						<TableHead>Supplier</TableHead>
						<TableHead>Suggested qty</TableHead>
						<TableHead>Status</TableHead>
						<TableHead class="text-right">Actions</TableHead>
					</TableRow>
				</TableHeader>
				<TableBody>
					{#if suggestionsLoading}
						{#each Array(4) as _, i}
							<TableRow>
								<TableCell colspan="5"><Skeleton class="h-6 w-full" /></TableCell>
							</TableRow>
						{/each}
					{:else if suggestions.length === 0}
						<TableRow>
							<TableCell colspan="5" class="text-center text-sm text-muted-foreground">No pending suggestions</TableCell>
						</TableRow>
					{:else}
						{#each suggestions as suggestion}
							<TableRow>
								<TableCell>{suggestion.Product?.Name ?? `Product ${suggestion.ProductID}`}</TableCell>
								<TableCell>{suggestion.Supplier?.Name ?? suggestion.SupplierID}</TableCell>
								<TableCell>{suggestion.SuggestedOrderQuantity}</TableCell>
								<TableCell>
									<span class="rounded-full bg-primary/10 px-2 py-0.5 text-xs capitalize text-primary">{suggestion.Status}</span>
								</TableCell>
								<TableCell class="text-right">
									<Button size="sm" variant="ghost" onclick={() => createPO(suggestion.ID)}>Create PO</Button>
								</TableCell>
							</TableRow>
						{/each}
					{/if}
				</TableBody>
			</Table>
		</CardContent>
	</Card>

	<div class="grid gap-6 lg:grid-cols-3">
	{#each reportKeys as key}
			<Card>
				<CardHeader>
					<CardTitle class="capitalize">{key} report</CardTitle>
					<CardDescription>Raw payload for BI handoff</CardDescription>
				</CardHeader>
				<CardContent>
					{#if reportsLoading && !reportResults[key]}
						<Skeleton class="h-36 w-full" />
					{:else if reportResults[key]}
						<pre class="max-h-56 overflow-auto rounded-lg bg-muted/50 p-3 text-xs">{JSON.stringify(reportResults[key], null, 2)}</pre>
					{:else}
						<p class="text-sm text-muted-foreground">Run the {key} report to populate this block.</p>
					{/if}
				</CardContent>
			</Card>
		{/each}
	</div>
</section>
