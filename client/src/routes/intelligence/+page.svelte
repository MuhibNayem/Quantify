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
	import { Download, TrendingUp, RotateCcw, Percent } from 'lucide-svelte';

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

	const exportJSON = (data: unknown, filename: string) => {
		const blob = new Blob([JSON.stringify(data, null, 2)], { type: 'application/json' });
		const url = URL.createObjectURL(blob);
		const a = document.createElement('a');
		a.href = url;
		a.download = `${filename}.json`;
		a.click();
		URL.revokeObjectURL(url);
	};

	// Helper for safe number formatting
	const fmt = (v: number | undefined) => (v ?? 0).toLocaleString();

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

		
		<div class="grid gap-8 md:grid-cols-2 lg:grid-cols-3" data-animate="fade-up">
	<!-- SALES REPORT -->
	<Card class="rounded-2xl border-0 shadow-lg hover:shadow-xl transition-all duration-300 hover:scale-[1.01] bg-gradient-to-br from-sky-50 to-blue-100">
		<CardHeader class="bg-white/80 backdrop-blur rounded-t-2xl border-b border-white/60 flex items-center justify-between px-6 py-5">
			<div>
				<CardTitle class="flex items-center gap-2 text-slate-800">
					<TrendingUp class="h-5 w-5 text-sky-600" />
					Sales Report
				</CardTitle>
				<CardDescription class="text-slate-600">Trend of total vs average sales</CardDescription>
			</div>
			<button class="rounded-full hover:bg-sky-100 p-2" onclick={() => exportJSON(reportResults.sales, 'sales-report')}>
				<Download class="h-4 w-4 text-sky-700" />
			</button>
		</CardHeader>
		<CardContent class="p-6 space-y-4">
			<!-- Simple Line Chart (SVG sparkline) -->
			<div class="h-24 w-full relative">
				<svg viewBox="0 0 100 40" preserveAspectRatio="none" class="absolute inset-0">
					<polyline
						points="0,35 10,28 20,25 30,18 40,20 50,14 60,12 70,18 80,22 90,15 100,10"
						fill="none"
						stroke="url(#salesGrad)"
						stroke-width="2.5"
						stroke-linecap="round"
					/>
					<defs>
						<linearGradient id="salesGrad" x1="0%" y1="0%" x2="100%" y2="0%">
							<stop offset="0%" stop-color="#38bdf8" />
							<stop offset="100%" stop-color="#2563eb" />
						</linearGradient>
					</defs>
				</svg>
			</div>
			<div class="flex justify-between text-sm text-slate-600">
				<p>Total Sales</p>
				<p class="font-semibold text-sky-700">{fmt(reportResults.sales?.totalSales)}</p>
			</div>
			<div class="flex justify-between text-sm text-slate-600">
				<p>Avg Daily Sales</p>
				<p class="font-semibold text-blue-700">{fmt(reportResults.sales?.averageDailySales)}</p>
			</div>
			<p class="text-xs text-slate-500">Period: {reportResults.sales?.period ?? '—'}</p>
		</CardContent>
	</Card>

	<!-- TURNOVER REPORT -->
	<Card class="rounded-2xl border-0 shadow-lg hover:shadow-xl transition-all duration-300 hover:scale-[1.01] bg-gradient-to-br from-cyan-50 to-teal-100">
		<CardHeader class="bg-white/80 backdrop-blur rounded-t-2xl border-b border-white/60 flex items-center justify-between px-6 py-5">
			<div>
				<CardTitle class="flex items-center gap-2 text-slate-800">
					<RotateCcw class="h-5 w-5 text-teal-600" />
					Turnover Report
				</CardTitle>
				<CardDescription class="text-slate-600">Inventory efficiency over time</CardDescription>
			</div>
			<button class="rounded-full hover:bg-teal-100 p-2" onclick={() => exportJSON(reportResults.turnover, 'turnover-report')}>
				<Download class="h-4 w-4 text-teal-700" />
			</button>
		</CardHeader>
		<CardContent class="p-6 space-y-4">
			<!-- Bar Chart -->
			<div class="h-24 w-full flex items-end gap-1">
				{#each [40, 25, 30, 15, 35, 22, 28] as height}
					<div
						class="flex-1 bg-gradient-to-t from-teal-400 to-cyan-400 rounded-t-md transition-all duration-500 ease-in-out hover:scale-105"
						style={`height: ${height}%;`}
					></div>
				{/each}
			</div>
			<div class="flex justify-between text-sm text-slate-600">
				<p>Avg Inventory Value</p>
				<p class="font-semibold text-teal-700">${fmt(reportResults.turnover?.averageInventoryValue)}</p>
			</div>
			<div class="flex justify-between text-sm text-slate-600">
				<p>Turnover Rate</p>
				<p class="font-semibold text-teal-700">{fmt(reportResults.turnover?.inventoryTurnoverRate)}</p>
			</div>
			<p class="text-xs text-slate-500">Period: {reportResults.turnover?.period ?? '—'}</p>
		</CardContent>
	</Card>

	<!-- MARGIN REPORT -->
	<Card class="rounded-2xl border-0 shadow-lg hover:shadow-xl transition-all duration-300 hover:scale-[1.01] bg-gradient-to-br from-pink-50 to-rose-100">
		<CardHeader class="bg-white/80 backdrop-blur rounded-t-2xl border-b border-white/60 flex items-center justify-between px-6 py-5">
			<div>
				<CardTitle class="flex items-center gap-2 text-slate-800">
					<Percent class="h-5 w-5 text-rose-600" />
					Margin Report
				</CardTitle>
				<CardDescription class="text-slate-600">Profitability visualization</CardDescription>
			</div>
			<button class="rounded-full hover:bg-pink-100 p-2" onclick={() => exportJSON(reportResults.margin, 'margin-report')}>
				<Download class="h-4 w-4 text-rose-700" />
			</button>
		</CardHeader>
		<CardContent class="p-6 space-y-4">
			<!-- Semi-donut Chart -->
			<div class="relative flex items-center justify-center h-24">
				<svg viewBox="0 0 36 18" class="w-24 h-12 rotate-180">
					<path
						d="M2 18 A16 16 0 0 1 34 18"
						fill="none"
						stroke="url(#marginGrad)"
						stroke-width="4"
						stroke-linecap="round"
						stroke-dasharray="{(reportResults.margin?.grossProfitMargin ?? 0) * 0.5},100"
					/>
					<defs>
						<linearGradient id="marginGrad" x1="0%" y1="0%" x2="100%" y2="0%">
							<stop offset="0%" stop-color="#fb7185" />
							<stop offset="100%" stop-color="#e11d48" />
						</linearGradient>
					</defs>
				</svg>
				<span class="absolute text-lg font-semibold text-rose-700">
					{fmt(reportResults.margin?.grossProfitMargin)}%
				</span>
			</div>
			<div class="flex justify-between text-sm text-slate-600">
				<p>Gross Profit</p>
				<p class="font-semibold text-rose-700">${fmt(reportResults.margin?.grossProfit)}</p>
			</div>
			<div class="flex justify-between text-sm text-slate-600">
				<p>Total Revenue</p>
				<p class="font-semibold text-rose-700">${fmt(reportResults.margin?.totalRevenue)}</p>
			</div>
			<p class="text-xs text-slate-500">Period: {reportResults.margin?.period ?? '—'}</p>
		</CardContent>
	</Card>
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
