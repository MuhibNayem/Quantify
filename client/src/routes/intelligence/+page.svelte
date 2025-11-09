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
	import { BarChart3 } from 'lucide-svelte';

	const forecastForm = $state({ periodInDays: '30', productId: '', result: '' });
	let suggestions = $state<ReorderSuggestion[]>([]);
	let suggestionsLoading = $state(false);

	const reportRange = $state({
		startDate: new Date(Date.now() - 30 * 24 * 60 * 60 * 1000).toISOString().slice(0, 10),
		endDate: new Date().toISOString().slice(0, 10)
	});
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
			toast.error('Failed to Load Suggestions', { description: errorMessage });
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
			toast.error('Failed to Run Forecast', { description: errorMessage });
		}
	};

	const createPO = async (suggestionId: number) => {
		try {
			const po = await replenishmentApi.createPOFromSuggestion(suggestionId);
			toast.success(`PO ${po.ID ?? 'created'}`);
			await loadSuggestions();
		} catch (error) {
			const errorMessage = error.response?.data?.error || 'Unable to create PO';
			toast.error('Failed to Create PO', { description: errorMessage });
		}
	};

	const runReport = async (type: ReportKey) => {
		reportsLoading = true;
		const payload = {
			startDate: new Date(reportRange.startDate).toISOString(),
			endDate: new Date(reportRange.endDate).toISOString()
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
			toast.error('Failed to Run Report', { description: errorMessage });
		} finally {
			reportsLoading = false;
		}
	};

	// Hero parallax (same system as Operations page)
	onMount(() => {
		const hero = document.querySelector('.parallax-hero') as HTMLElement | null;
		if (!hero) return;
		const handleScroll = () => {
			const scrollY = window.scrollY / 6;
			hero.style.transform = `translateY(${scrollY}px)`;
		};
		window.addEventListener('scroll', handleScroll);
		return () => window.removeEventListener('scroll', handleScroll);
	});
</script>

<!-- HERO -->
<section class="relative w-full overflow-hidden bg-gradient-to-r from-sky-50 via-blue-50 to-cyan-100 animate-gradientShift py-20 px-6 text-center">
	<div class="absolute inset-0 bg-white/40 backdrop-blur-sm"></div>

	<div class="relative z-10 max-w-3xl mx-auto flex flex-col items-center justify-center space-y-4 transform transition-transform duration-700 ease-out will-change-transform parallax-hero">
		<div class="p-4 bg-gradient-to-br from-sky-400 to-blue-500 rounded-2xl shadow-lg animate-pulseGlow">
			<BarChart3 class="h-8 w-8 text-white" />
		</div>
		<h1 class="text-4xl sm:text-5xl font-bold bg-gradient-to-r from-sky-600 via-blue-600 to-cyan-600 bg-clip-text text-transparent animate-fadeUp">
			Forecasting, Reorder Suggestions & Business Reports
		</h1>
		<p class="text-slate-600 text-base animate-fadeUp delay-200">
			Plan ahead, act on signals, and align analytics across one horizon.
		</p>
	</div>
</section>

<!-- MAIN -->
<section class="max-w-7xl mx-auto py-14 px-6 bg-white space-y-10">
	<!-- Forecast + Range -->
	<div class="grid gap-8 lg:grid-cols-2">
		<!-- Demand forecast -->
		<Card class="rounded-2xl border-0 shadow-lg hover:shadow-xl transition-all duration-300 hover:scale-[1.02] bg-gradient-to-br from-sky-50 to-blue-100">
			<CardHeader class="bg-white/80 backdrop-blur rounded-t-2xl border-b border-white/60 px-6 py-5">
				<CardTitle class="text-slate-800">Demand Forecast</CardTitle>
				<CardDescription class="text-slate-600">Trigger rolling forecasts for targeted SKUs</CardDescription>
			</CardHeader>
			<CardContent class="space-y-4 p-6">
				<div class="grid gap-3 sm:grid-cols-2">
					<Input type="number" min="7" placeholder="Horizon (days)" bind:value={forecastForm.periodInDays} class="rounded-xl border-sky-200 bg-white/90 focus:ring-2 focus:ring-sky-400" />
					<Input type="number" min="1" placeholder="Product ID (optional)" bind:value={forecastForm.productId} class="rounded-xl border-sky-200 bg-white/90 focus:ring-2 focus:ring-sky-400" />
				</div>

				<!-- Responsive action row -->
				<div class="flex flex-col sm:flex-row gap-3">
					<Button class="flex-1 bg-gradient-to-r from-sky-500 to-blue-600 hover:from-sky-600 hover:to-blue-700 text-white font-semibold rounded-xl shadow-md hover:shadow-lg hover:scale-105 transition-all" onclick={triggerForecast}>
						Generate Forecast
					</Button>
				</div>

				{#if forecastForm.result}
					<p class="rounded-xl border border-sky-200 bg-white/70 backdrop-blur p-3 text-sm text-slate-700">
						{forecastForm.result}
					</p>
				{/if}
			</CardContent>
		</Card>

		<!-- Report range -->
		<Card class="rounded-2xl border-0 shadow-lg hover:shadow-xl transition-all duration-300 hover:scale-[1.02] bg-gradient-to-br from-cyan-50 to-teal-100">
			<CardHeader class="bg-white/80 backdrop-blur rounded-t-2xl border-b border-white/60 px-6 py-5">
				<CardTitle class="text-slate-800">Report Range</CardTitle>
				<CardDescription class="text-slate-600">Align analytics across shared horizon</CardDescription>
			</CardHeader>
			<CardContent class="space-y-4 p-6">
				<div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
					<Input type="date" bind:value={reportRange.startDate} class="rounded-xl border-cyan-200 bg-white/90 focus:ring-2 focus:ring-cyan-400" />
					<Input type="date" bind:value={reportRange.endDate} class="rounded-xl border-cyan-200 bg-white/90 focus:ring-2 focus:ring-cyan-400" />
				</div>

				<!-- Responsive 3-button group -->
				<div class="flex flex-col sm:flex-row gap-3">
					<Button class="flex-1 bg-white/80 border border-cyan-200 text-cyan-700 hover:bg-cyan-50 rounded-xl font-medium hover:scale-105 transition-all shadow-sm" variant="secondary" onclick={() => runReport('sales')}>
						Sales Trends
					</Button>
					<Button class="flex-1 bg-white/80 border border-cyan-200 text-cyan-700 hover:bg-cyan-50 rounded-xl font-medium hover:scale-105 transition-all shadow-sm" variant="secondary" onclick={() => runReport('turnover')}>
						Inventory Turnover
					</Button>
					<Button class="flex-1 bg-white/80 border border-cyan-200 text-cyan-700 hover:bg-cyan-50 rounded-xl font-medium hover:scale-105 transition-all shadow-sm" variant="secondary" onclick={() => runReport('margin')}>
						Profit Margin
					</Button>
				</div>
			</CardContent>
		</Card>
	</div>

	<!-- Reorder suggestions -->
	<Card class="rounded-2xl border-0 shadow-lg hover:shadow-xl transition-all duration-300 bg-gradient-to-br from-amber-50 to-orange-100">
		<CardHeader class="bg-white/80 backdrop-blur rounded-t-2xl border-b border-white/60 px-6 py-5">
			<CardTitle class="text-slate-800">Reorder Suggestions</CardTitle>
			<CardDescription class="text-slate-600">Convert high-signal suggestions into purchase orders</CardDescription>
		</CardHeader>
		<CardContent class="p-0">
			<Table class="border border-amber-200/60 rounded-2xl overflow-hidden">
				<TableHeader class="bg-gradient-to-r from-amber-100 to-orange-100">
					<TableRow>
						<TableHead>Product</TableHead>
						<TableHead>Supplier</TableHead>
						<TableHead>Suggested qty</TableHead>
						<TableHead>Status</TableHead>
						<TableHead class="text-right">Actions</TableHead>
					</TableRow>
				</TableHeader>
				<TableBody class="[&>tr:nth-child(even)]:bg-white/70 [&>tr:nth-child(odd)]:bg-white/60">
					{#if suggestionsLoading}
						{#each Array(4) as _, i}
							<TableRow>
								<TableCell colspan="5" class="p-3"><Skeleton class="h-6 w-full bg-white/70" /></TableCell>
							</TableRow>
						{/each}
					{:else if suggestions.length === 0}
						<TableRow>
							<TableCell colspan="5" class="text-center text-sm text-slate-500 py-6">No pending suggestions</TableCell>
						</TableRow>
					{:else}
						{#each suggestions as suggestion}
							<TableRow class="hover:bg-white/90 transition-colors">
								<TableCell>{suggestion.Product?.Name ?? `Product ${suggestion.ProductID}`}</TableCell>
								<TableCell>{suggestion.Supplier?.Name ?? suggestion.SupplierID}</TableCell>
								<TableCell>{suggestion.SuggestedOrderQuantity}</TableCell>
								<TableCell>
									<span class="inline-flex items-center rounded-full bg-orange-200/60 text-orange-800 px-2 py-0.5 text-xs capitalize">
										{suggestion.Status}
									</span>
								</TableCell>
								<TableCell class="text-right">
									<!-- Responsive action: single button stays compact -->
									<Button size="sm" variant="ghost" class="text-amber-700 hover:bg-amber-50 rounded-md px-3 py-1.5" onclick={() => createPO(suggestion.ID)}>
										Create PO
									</Button>
								</TableCell>
							</TableRow>
						{/each}
					{/if}
				</TableBody>
			</Table>
		</CardContent>
	</Card>

	<!-- Reports payloads -->
	<div class="grid gap-8 md:grid-cols-2 lg:grid-cols-3">
		{#each reportKeys as key}
			<Card class="rounded-2xl border-0 shadow-lg hover:shadow-xl transition-all duration-300 hover:scale-[1.01] bg-gradient-to-br from-slate-50 to-slate-100">
				<CardHeader class="bg-white/80 backdrop-blur rounded-t-2xl border-b border-white/60 px-6 py-5">
					<CardTitle class="capitalize text-slate-800">{key} report</CardTitle>
					<CardDescription class="text-slate-600">Raw payload for BI handoff</CardDescription>
				</CardHeader>
				<CardContent class="p-6">
					{#if reportsLoading && !reportResults[key]}
						<Skeleton class="h-36 w-full bg-white/70" />
					{:else if reportResults[key]}
						<pre class="max-h-56 overflow-auto rounded-xl border border-slate-200 bg-white/70 backdrop-blur p-3 text-xs text-slate-800">{JSON.stringify(reportResults[key], null, 2)}</pre>
					{:else}
						<p class="text-sm text-slate-600">Run the {key} report to populate this block.</p>
					{/if}
				</CardContent>
			</Card>
		{/each}
	</div>
</section>

<style lang="postcss">
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
		0%, 100% { transform: scale(1); box-shadow: 0 0 15px rgba(56, 189, 248, 0.3); }
		50% { transform: scale(1.08); box-shadow: 0 0 25px rgba(56, 189, 248, 0.5); }
	}
	.animate-pulseGlow { animation: pulseGlow 8s ease-in-out infinite; }

	@keyframes fadeUp {
		from { opacity: 0; transform: translateY(20px); }
		to { opacity: 1; transform: translateY(0); }
	}
	.animate-fadeUp { animation: fadeUp 1.5s ease forwards; }

	* {
		transition-property: color, background-color, border-color, text-decoration-color, fill, stroke, opacity, box-shadow, transform, filter, backdrop-filter;
		transition-timing-function: cubic-bezier(0.4, 0, 0.2, 1);
		transition-duration: 300ms;
	}
</style>
