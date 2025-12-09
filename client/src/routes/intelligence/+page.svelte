<script lang="ts">
	import { onMount } from 'svelte';
	import { toast } from 'svelte-sonner';
	import {
		Card,
		CardContent,
		CardDescription,
		CardHeader,
		CardTitle
	} from '$lib/components/ui/card';
	import { Input } from '$lib/components/ui/input';
	import { Button } from '$lib/components/ui/button';
	import {
		Table,
		TableBody,
		TableCell,
		TableHead,
		TableHeader,
		TableRow
	} from '$lib/components/ui/table';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import { replenishmentApi, reportsApi } from '$lib/api/resources';
	import type { ReorderSuggestion } from '$lib/types';
	import { BarChart3 } from 'lucide-svelte';
	import { Download, TrendingUp, RotateCcw, Percent, ShoppingCart, RefreshCw } from 'lucide-svelte';
	import { auth } from '$lib/stores/auth';
	import { goto } from '$app/navigation';
    import DemandForecastWidget from '$lib/components/dashboard/DemandForecastWidget.svelte';
    import ChurnRiskWidget from '$lib/components/dashboard/ChurnRiskWidget.svelte';

	$effect(() => {
		if (!auth.hasPermission('reports.view')) {
			toast.error('Access Denied', { description: 'You do not have permission to view reports.' });
			goto('/');
		}
	});


    
    // Define report types
    interface SalesReport {
        salesTrends: any[];
        totalSales: number;
        averageDailySales: number;
        period: string;
    }
    interface TurnoverReport {
        averageInventoryValue: number;
        inventoryTurnoverRate: number;
        period: string;
    }
    interface MarginReport {
        grossProfitMargin: number;
        grossProfit: number;
        totalRevenue: number;
        period: string;
    }

	let suggestions = $state<ReorderSuggestion[]>([]);
	let suggestionsLoading = $state(false);

	const reportRange = $state({
		startDate: new Date(Date.now() - 30 * 24 * 60 * 60 * 1000).toISOString().slice(0, 10),
		endDate: new Date().toISOString().slice(0, 10)
	});
	const reportKeys = ['sales', 'turnover', 'margin'] as const;
	type ReportKey = (typeof reportKeys)[number];
	
    const reportResults = $state<{
        sales: SalesReport | null;
        turnover: TurnoverReport | null;
        margin: MarginReport | null;
    }>({
		sales: null,
		turnover: null,
		margin: null
	});
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

	const generateSparkline = (data: any[] | undefined) => {
		if (!data || data.length === 0) return '0,40 100,40';

		const values = data.map((d) => d.TotalSales);
		const min = Math.min(...values);
		const max = Math.max(...values);
		const range = max - min || 1;

		// Map to 100x40 coordinate system
		// X: 0 to 100
		// Y: 40 (bottom) to 0 (top)
		const points = values.map((val, i) => {
			const x = (i / (values.length - 1)) * 100;
			const y = 40 - ((val - min) / range) * 35; // Leave 5px padding at top
			return `${x},${y}`;
		});

		return points.join(' ');
	};

	const loadSuggestions = async () => {
		suggestionsLoading = true;
		try {
			suggestions = await replenishmentApi.listSuggestions();
		} catch (error: any) {
			const errorMessage = error.response?.data?.error || 'Unable to load suggestions';
			toast.error('Failed to Load Suggestions', { description: errorMessage });
		} finally {
			suggestionsLoading = false;
		}
	};
	onMount(loadSuggestions);



	const createPO = async (suggestionId: number) => {
		try {
			const po = await replenishmentApi.createPOFromSuggestion(suggestionId);
			toast.success(`PO ${po.ID ?? 'created'}`);
			await loadSuggestions();
		} catch (error: any) {
			const errorMessage = error.response?.data?.error || 'Unable to create PO';
			toast.error('Failed to Create PO', { description: errorMessage });
		}
	};

	const runReport = async (type: ReportKey) => {
		reportsLoading = true;
		const payload = {
			startDate: new Date(reportRange.startDate).toISOString(),
			endDate: new Date(reportRange.endDate).toISOString(),
			groupBy: 'daily'
		};
		try {
			if (type === 'sales') {
				reportResults.sales = await reportsApi.salesTrends(payload) as SalesReport;
			} else if (type === 'turnover') {
				reportResults.turnover = await reportsApi.inventoryTurnover(payload) as TurnoverReport;
			} else {
				reportResults.margin = await reportsApi.profitMargin(payload) as MarginReport;
			}
			toast.success('Report ready');
		} catch (error: any) {
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
<section
	class="animate-gradientShift relative w-full overflow-hidden bg-gradient-to-r from-sky-50 via-blue-50 to-cyan-100 px-6 py-20 text-center"
>
	<div class="absolute inset-0 bg-white/40 backdrop-blur-sm"></div>

	<div
		class="parallax-hero relative z-10 mx-auto flex max-w-3xl transform flex-col items-center justify-center space-y-4 transition-transform duration-700 ease-out will-change-transform"
	>
		<div
			class="animate-pulseGlow rounded-2xl bg-gradient-to-br from-sky-400 to-blue-500 p-4 shadow-lg"
		>
			<BarChart3 class="h-8 w-8 text-white" />
		</div>
		<h1
			class="animate-fadeUp bg-gradient-to-r from-sky-600 via-blue-600 to-cyan-600 bg-clip-text text-4xl font-bold text-transparent sm:text-5xl"
		>
			Forecasting, Reorder Suggestions & Business Reports
		</h1>
		<p class="animate-fadeUp text-base text-slate-600 delay-200">
			Plan ahead, act on signals, and align analytics across one horizon.
		</p>
	</div>
</section>

<!-- MAIN -->
<section class="mx-auto max-w-7xl space-y-10 bg-white px-6 py-14">
	<!-- Forecast + Churn Risk -->
	<div class="grid gap-8 lg:grid-cols-2">
		<!-- Demand forecast -->
        <div class="h-full">
            <DemandForecastWidget />
        </div>
        <!-- Churn Risk -->
        <div class="h-full">
            <ChurnRiskWidget />
        </div>
	</div>

    <!-- Report range -->
    <div class="mx-auto max-w-4xl">
		<Card
			class="rounded-2xl border-0 bg-gradient-to-br from-cyan-50 to-teal-100 shadow-lg transition-all duration-300 hover:scale-[1.02] hover:shadow-xl"
		>
			<CardHeader
				class="rounded-t-2xl border-b border-white/60 bg-white/80 px-6 py-5 backdrop-blur"
			>
				<CardTitle class="text-slate-800">Report Range</CardTitle>
				<CardDescription class="text-slate-600"
					>Align analytics across shared horizon</CardDescription
				>
			</CardHeader>
			<CardContent class="space-y-4 p-6">
				<div class="grid grid-cols-1 gap-3 sm:grid-cols-2">
					<Input
						type="date"
						bind:value={reportRange.startDate}
						class="rounded-xl border-cyan-200 bg-white/90 focus:ring-2 focus:ring-cyan-400"
					/>
					<Input
						type="date"
						bind:value={reportRange.endDate}
						class="rounded-xl border-cyan-200 bg-white/90 focus:ring-2 focus:ring-cyan-400"
					/>
				</div>

				<!-- Responsive 3-button group -->
				<div class="flex flex-col gap-3 sm:flex-row">
					<Button
						class="flex-1 rounded-xl border border-cyan-200 bg-white/80 font-medium text-cyan-700 shadow-sm transition-all hover:scale-105 hover:bg-cyan-50"
						variant="secondary"
						onclick={() => runReport('sales')}
					>
						Sales Trends
					</Button>
					<Button
						class="flex-1 rounded-xl border border-cyan-200 bg-white/80 font-medium text-cyan-700 shadow-sm transition-all hover:scale-105 hover:bg-cyan-50"
						variant="secondary"
						onclick={() => runReport('turnover')}
					>
						Inventory Turnover
					</Button>
					<Button
						class="flex-1 rounded-xl border border-cyan-200 bg-white/80 font-medium text-cyan-700 shadow-sm transition-all hover:scale-105 hover:bg-cyan-50"
						variant="secondary"
						onclick={() => runReport('margin')}
					>
						Profit Margin
					</Button>
				</div>
			</CardContent>
		</Card>
    </div>

	<!-- Reorder suggestions -->
	<Card
		class="rounded-2xl border-0 bg-gradient-to-br from-amber-50 to-orange-100 shadow-lg transition-all duration-300 hover:shadow-xl"
	>
		<CardHeader class="rounded-t-2xl border-b border-white/60 bg-white/80 px-6 py-5 backdrop-blur">
			<div class="flex items-center justify-between">
				<div>
					<CardTitle class="flex items-center gap-2 text-slate-800">
						<ShoppingCart class="h-5 w-5 text-emerald-600" />
						Reorder Suggestions
					</CardTitle>
					<CardDescription class="text-slate-600">AI-recommended purchase orders</CardDescription>
				</div>
				<Button
					variant="outline"
					size="sm"
					class="gap-2 border-emerald-200 text-emerald-700 hover:bg-emerald-50"
					onclick={async () => {
						suggestionsLoading = true;
						try {
							await replenishmentApi.generateSuggestions();
							await loadSuggestions();
							toast.success('Suggestions refreshed');
						} catch (e) {
							toast.error('Failed to refresh suggestions');
						} finally {
							suggestionsLoading = false;
						}
					}}
					disabled={suggestionsLoading}
				>
					<RefreshCw class={suggestionsLoading ? 'h-4 w-4 animate-spin' : 'h-4 w-4'} />
					Refresh
				</Button>
			</div>
		</CardHeader>
		<CardContent class="p-0">
			<Table class="overflow-hidden rounded-2xl border border-amber-200/60">
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
								<TableCell colspan={5} class="p-3"
									><Skeleton class="h-6 w-full bg-white/70" /></TableCell
								>
							</TableRow>
						{/each}
					{:else if suggestions.length === 0}
						<TableRow>
							<TableCell colspan={5} class="py-6 text-center text-sm text-slate-500"
								>No pending suggestions</TableCell
							>
						</TableRow>
					{:else}
						{#each suggestions as suggestion}
							<TableRow class="transition-colors hover:bg-white/90">
								<TableCell
									>{suggestion.Product?.Name ?? `Product ${suggestion.ProductID}`}</TableCell
								>
								<TableCell>{suggestion.Supplier?.Name ?? String(suggestion.SupplierID)}</TableCell>
								<TableCell>{suggestion.SuggestedOrderQuantity}</TableCell>
								<TableCell>
									<span
										class="inline-flex items-center rounded-full bg-orange-200/60 px-2 py-0.5 text-xs capitalize text-orange-800"
									>
										{suggestion.Status}
									</span>
								</TableCell>
								<TableCell class="text-right">
									<!-- Responsive action: single button stays compact -->
									<Button
										size="sm"
										variant="ghost"
										class="rounded-md px-3 py-1.5 text-amber-700 hover:bg-amber-50"
										onclick={() => createPO(suggestion.ID)}
									>
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
		{#if auth.hasPermission('reports.sales')}
			<Card
				class="rounded-2xl border-0 bg-gradient-to-br from-sky-50 to-blue-100 shadow-lg transition-all duration-300 hover:scale-[1.01] hover:shadow-xl"
			>
				<CardHeader
					class="flex items-center justify-between rounded-t-2xl border-b border-white/60 bg-white/80 px-6 py-5 backdrop-blur"
				>
					<div>
						<CardTitle class="flex items-center gap-2 text-slate-800">
							<TrendingUp class="h-5 w-5 text-sky-600" />
							Sales Report
						</CardTitle>
						<CardDescription class="text-slate-600">Trend of total vs average sales</CardDescription
						>
					</div>
					<button
						class="rounded-full p-2 hover:bg-sky-100"
						onclick={() => exportJSON(reportResults.sales, 'sales-report')}
					>
						<Download class="h-4 w-4 text-sky-700" />
					</button>
				</CardHeader>
				<CardContent class="space-y-4 p-6">
					<!-- Simple Line Chart (SVG sparkline) -->
					<div class="relative h-24 w-full">
						<svg viewBox="0 0 100 40" preserveAspectRatio="none" class="absolute inset-0">
							<polyline
								points={generateSparkline(reportResults.sales?.salesTrends)}
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
		{/if}

		<!-- TURNOVER REPORT -->
		{#if auth.hasPermission('reports.inventory')}
			<Card
				class="rounded-2xl border-0 bg-gradient-to-br from-cyan-50 to-teal-100 shadow-lg transition-all duration-300 hover:scale-[1.01] hover:shadow-xl"
			>
				<CardHeader
					class="flex items-center justify-between rounded-t-2xl border-b border-white/60 bg-white/80 px-6 py-5 backdrop-blur"
				>
					<div>
						<CardTitle class="flex items-center gap-2 text-slate-800">
							<RotateCcw class="h-5 w-5 text-teal-600" />
							Turnover Report
						</CardTitle>
						<CardDescription class="text-slate-600">Inventory efficiency over time</CardDescription>
					</div>
					<button
						class="rounded-full p-2 hover:bg-teal-100"
						onclick={() => exportJSON(reportResults.turnover, 'turnover-report')}
					>
						<Download class="h-4 w-4 text-teal-700" />
					</button>
				</CardHeader>
				<CardContent class="space-y-4 p-6">
					<!-- Bar Chart -->
					<div class="flex h-24 w-full items-end gap-1">
						{#each [40, 25, 30, 15, 35, 22, 28] as height}
							<div
								class="flex-1 rounded-t-md bg-gradient-to-t from-teal-400 to-cyan-400 transition-all duration-500 ease-in-out hover:scale-105"
								style={`height: ${height}%;`}
							></div>
						{/each}
					</div>
					<div class="flex justify-between text-sm text-slate-600">
						<p>Avg Inventory Value</p>
						<p class="font-semibold text-teal-700">
							${fmt(reportResults.turnover?.averageInventoryValue)}
						</p>
					</div>
					<div class="flex justify-between text-sm text-slate-600">
						<p>Turnover Rate</p>
						<p class="font-semibold text-teal-700">
							{fmt(reportResults.turnover?.inventoryTurnoverRate)}
						</p>
					</div>
					<p class="text-xs text-slate-500">Period: {reportResults.turnover?.period ?? '—'}</p>
				</CardContent>
			</Card>
		{/if}

		<!-- MARGIN REPORT -->
		{#if auth.hasPermission('reports.financial')}
			<Card
				class="rounded-2xl border-0 bg-gradient-to-br from-pink-50 to-rose-100 shadow-lg transition-all duration-300 hover:scale-[1.01] hover:shadow-xl"
			>
				<CardHeader
					class="flex items-center justify-between rounded-t-2xl border-b border-white/60 bg-white/80 px-6 py-5 backdrop-blur"
				>
					<div>
						<CardTitle class="flex items-center gap-2 text-slate-800">
							<Percent class="h-5 w-5 text-rose-600" />
							Margin Report
						</CardTitle>
						<CardDescription class="text-slate-600">Profitability visualization</CardDescription>
					</div>
					<button
						class="rounded-full p-2 hover:bg-pink-100"
						onclick={() => exportJSON(reportResults.margin, 'margin-report')}
					>
						<Download class="h-4 w-4 text-rose-700" />
					</button>
				</CardHeader>
				<CardContent class="space-y-4 p-6">
					<!-- Semi-donut Chart -->
					<div class="relative flex h-24 items-center justify-center">
						<svg viewBox="0 0 36 18" class="h-12 w-24 rotate-180">
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
		{/if}
	</div>
</section>

<style lang="postcss">
	@keyframes gradientShift {
		0% {
			background-position: 0% 50%;
		}
		50% {
			background-position: 100% 50%;
		}
		100% {
			background-position: 0% 50%;
		}
	}
	.animate-gradientShift {
		background-size: 200% 200%;
		animation: gradientShift 20s ease infinite;
	}

	@keyframes pulseGlow {
		0%,
		100% {
			transform: scale(1);
			box-shadow: 0 0 15px rgba(56, 189, 248, 0.3);
		}
		50% {
			transform: scale(1.08);
			box-shadow: 0 0 25px rgba(56, 189, 248, 0.5);
		}
	}
	.animate-pulseGlow {
		animation: pulseGlow 8s ease-in-out infinite;
	}

	@keyframes fadeUp {
		from {
			opacity: 0;
			transform: translateY(20px);
		}
		to {
			opacity: 1;
			transform: translateY(0);
		}
	}
	.animate-fadeUp {
		animation: fadeUp 1.5s ease forwards;
	}

	* {
		transition-property:
			color, background-color, border-color, text-decoration-color, fill, stroke, opacity,
			box-shadow, transform, filter, backdrop-filter;
		transition-timing-function: cubic-bezier(0.4, 0, 0.2, 1);
		transition-duration: 300ms;
	}
</style>
