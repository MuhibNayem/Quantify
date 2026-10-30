<script lang="ts">
	import { t } from '$lib/i18n';
	import { onMount } from 'svelte';
	import { toast } from 'svelte-sonner';
	import { Input } from '$lib/components/ui/input';
	import { Button } from '$lib/components/ui/button';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import HelpTip from '$lib/components/ui/help-tip.svelte';
	import { replenishmentApi, reportsApi } from '$lib/api/resources';
	import type { ReorderSuggestion } from '$lib/types';
	import { BarChart3 } from 'lucide-svelte';
	import {
		Download,
		TrendingUp,
		RotateCcw,
		Percent,
		ShoppingCart,
		RefreshCw,
		FileSpreadsheet,
		FileText,
		Loader2,
		Package,
		ClipboardList,
		AlertTriangle,
		ArrowRight
	} from 'lucide-svelte';
	import { auth } from '$lib/stores/auth';
	import { goto } from '$app/navigation';
	import DemandForecastWidget from '$lib/components/dashboard/DemandForecastWidget.svelte';
	import ChurnRiskWidget from '$lib/components/dashboard/ChurnRiskWidget.svelte';

	const inputClass =
		'glass-input h-10 w-full px-3 text-sm font-normal text-foreground placeholder:text-muted-foreground';
	const btnPrimary = 'btn-soothe h-10 px-4 text-sm font-semibold';
	const btnSecondary = 'shadow-none';

	$effect(() => {
		if (!auth.hasPermission('reports.view')) {
			toast.error($t('intelligence.toasts.access_denied'), {
				description: $t('intelligence.toasts.access_denied_desc')
			});
			goto('/');
		}
	});

	let replenishmentDashboard = $state<any>(null);
	let replenishmentLoading = $state(false);

	const loadReplenishmentDashboard = async () => {
		replenishmentLoading = true;
		try {
			replenishmentDashboard = await replenishmentApi.getDashboard();
		} catch (e) {
			console.error('Failed to load replenishment dashboard', e);
		} finally {
			replenishmentLoading = false;
		}
	};

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

	let exportFormat = $state<'csv' | 'pdf' | 'excel'>('csv');
	let exportJobId = $state<string | null>(null);
	let exportStatus = $state<'idle' | 'pending' | 'processing' | 'complete' | 'failed'>('idle');
	let exportPollTimer: ReturnType<typeof setInterval>;

	const startExport = async () => {
		exportStatus = 'pending';
		try {
			const job = await reportsApi.exportReport({ format: exportFormat });
			exportJobId = job.jobId || job.id || job.ID;
			exportStatus = 'processing';
			toast.success('Export started', {
				description: `Generating ${exportFormat.toUpperCase()} report...`
			});
			pollExportStatus();
		} catch (error: any) {
			exportStatus = 'failed';
			toast.error('Export failed', {
				description: error.response?.data?.error || 'Could not start export job'
			});
		}
	};

	const pollExportStatus = () => {
		clearInterval(exportPollTimer);
		exportPollTimer = setInterval(async () => {
			if (!exportJobId) return;
			try {
				const status = await reportsApi.getJobStatus(exportJobId);
				const state = (status.status || status.state || '').toLowerCase();
				if (state === 'complete' || state === 'completed' || state === 'done') {
					clearInterval(exportPollTimer);
					exportStatus = 'complete';
					toast.success('Export ready', { description: 'Your report is ready to download.' });
					const url = reportsApi.getDownloadUrl(exportJobId);
					window.open(url, '_blank');
				} else if (state === 'failed' || state === 'error') {
					clearInterval(exportPollTimer);
					exportStatus = 'failed';
					toast.error('Export failed', { description: status.error || 'The export job failed.' });
				}
			} catch {
				clearInterval(exportPollTimer);
				exportStatus = 'failed';
				toast.error('Export failed', { description: 'Could not check export status.' });
			}
		}, 2000);
	};

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
			const errorMessage =
				error.response?.data?.error || $t('intelligence.toasts.load_suggestions_fail');
			toast.error($t('intelligence.toasts.load_suggestions_fail'), { description: errorMessage });
		} finally {
			suggestionsLoading = false;
		}
	};
	onMount(() => {
		loadSuggestions();
		loadReplenishmentDashboard();
	});

	const createPO = async (suggestionId: number) => {
		try {
			const po = await replenishmentApi.createPOFromSuggestion(suggestionId);
			toast.success($t('intelligence.toasts.po_created', { id: po.ID ?? 'created' }));
			await loadSuggestions();
		} catch (error: any) {
			const errorMessage = error.response?.data?.error || 'Unable to create PO';
			toast.error($t('intelligence.toasts.po_create_fail'), { description: errorMessage });
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
				reportResults.sales = (await reportsApi.salesTrends(payload)) as SalesReport;
			} else if (type === 'turnover') {
				reportResults.turnover = (await reportsApi.inventoryTurnover(payload)) as TurnoverReport;
			} else {
				reportResults.margin = (await reportsApi.profitMargin(payload)) as MarginReport;
			}
			toast.success($t('intelligence.toasts.report_ready'));
		} catch (error: any) {
			const errorMessage = error.response?.data?.error || 'Unable to run report';
			toast.error($t('intelligence.toasts.report_fail'), { description: errorMessage });
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
<section class="w-full">
	<div class="mx-auto max-w-7xl px-6 pt-8 pb-4 sm:pt-10">
		<header
			class="parallax-hero page-header transition-transform duration-700 ease-out will-change-transform"
			data-animate="fade-up"
		>
			<div class="min-w-0">
				<p class="flex items-center gap-2 overline">
					<span class="inline-flex rounded-md bg-primary/10 p-1 text-primary">
						<BarChart3 class="h-3.5 w-3.5" />
					</span>
					{$t('intelligence.hero.subtitle')}
				</p>
				<h1 class="page-title mt-2">{$t('intelligence.hero.title')}</h1>
			</div>
		</header>
	</div>
</section>

<!-- MAIN -->
<section class="mx-auto max-w-7xl space-y-6 px-6 py-8">
	<!-- Replenishment Dashboard -->
	<div class="panel">
		<div class="panel-header">
			<div class="min-w-0">
				<h2 class="section-title flex items-center gap-2">
					<Package class="h-4 w-4 text-primary" />
					Replenishment Overview
				</h2>
				<p class="mt-0.5 text-xs text-muted-foreground">
					Quick stats for your replenishment pipeline
				</p>
			</div>
		</div>
		<div class="panel-body space-y-4">
			{#if replenishmentLoading}
				<div class="grid gap-4 sm:grid-cols-3">
					{#each Array(3) as _}
						<Skeleton class="h-24 w-full rounded-md bg-muted" />
					{/each}
				</div>
			{:else if replenishmentDashboard}
				<div class="grid gap-4 sm:grid-cols-3">
					<div class="surface-1 p-5 text-center">
						<ClipboardList class="mx-auto mb-2 h-5 w-5 text-primary" />
						<p class="kpi-num text-3xl text-foreground">
							{replenishmentDashboard.pendingSuggestions ??
								replenishmentDashboard.PendingSuggestions ??
								0}
						</p>
						<p class="mt-1 text-xs text-muted-foreground">Pending Suggestions</p>
					</div>
					<div class="surface-1 p-5 text-center">
						<ShoppingCart class="mx-auto mb-2 h-5 w-5 text-info" />
						<p class="kpi-num text-3xl text-foreground">
							{replenishmentDashboard.openPOs ?? replenishmentDashboard.OpenPOs ?? 0}
						</p>
						<p class="mt-1 text-xs text-muted-foreground">Open Purchase Orders</p>
					</div>
					<div class="surface-1 p-5 text-center">
						<AlertTriangle class="mx-auto mb-2 h-5 w-5 text-warning" />
						<p class="kpi-num text-3xl text-foreground">
							{replenishmentDashboard.lowStockItems ?? replenishmentDashboard.LowStockItems ?? 0}
						</p>
						<p class="mt-1 text-xs text-muted-foreground">Low Stock Items</p>
					</div>
				</div>
				<div class="flex flex-wrap gap-2">
					<a href="/intelligence" class="no-underline">
						<Button variant="outline" class="{btnSecondary} h-9 gap-1.5 px-3 text-xs font-semibold">
							View Suggestions <ArrowRight class="h-3.5 w-3.5" />
						</Button>
					</a>
					<a href="/purchase-orders" class="no-underline">
						<Button variant="outline" class="{btnSecondary} h-9 gap-1.5 px-3 text-xs font-semibold">
							View Purchase Orders <ArrowRight class="h-3.5 w-3.5" />
						</Button>
					</a>
				</div>
			{:else}
				<p class="py-4 text-center text-sm text-muted-foreground">
					Unable to load replenishment data.
				</p>
			{/if}
		</div>
	</div>

	<!-- Forecast + Churn Risk -->
	<div class="grid gap-4 lg:grid-cols-2">
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
	<div class="panel">
		<div class="panel-header">
			<div class="min-w-0">
				<h2 class="section-title">{$t('intelligence.report_range.title')}</h2>
				<p class="mt-0.5 text-xs text-muted-foreground">
					{$t('intelligence.report_range.subtitle')}
				</p>
			</div>
		</div>
		<div class="panel-body space-y-4">
			<div class="grid grid-cols-1 gap-3 sm:grid-cols-2">
				<Input type="date" bind:value={reportRange.startDate} class="{inputClass} num-mono" />
				<Input type="date" bind:value={reportRange.endDate} class="{inputClass} num-mono" />
			</div>

			<!-- Responsive 3-button group -->
			<div class="flex flex-col gap-3 sm:flex-row">
				<Button
					class="{btnSecondary} h-10 flex-1 text-sm"
					variant="outline"
					onclick={() => runReport('sales')}
				>
					{$t('intelligence.report_range.sales_trends')}
				</Button>
				<Button
					class="{btnSecondary} h-10 flex-1 text-sm"
					variant="outline"
					onclick={() => runReport('turnover')}
				>
					{$t('intelligence.report_range.turnover')}
				</Button>
				<Button
					class="{btnSecondary} h-10 flex-1 text-sm"
					variant="outline"
					onclick={() => runReport('margin')}
				>
					{$t('intelligence.report_range.margin')}
				</Button>
			</div>

			<div class="flex flex-wrap items-center gap-3 border-t border-border pt-4">
				<div
					class="inline-flex items-center gap-1 rounded-[calc(var(--radius)_-_4px)] border border-border bg-muted/60 p-1"
				>
					<button
						class="inline-flex items-center gap-1.5 rounded-md px-3 py-1.5 text-xs font-medium transition-colors {exportFormat ===
						'csv'
							? 'bg-card text-foreground shadow-xs'
							: 'text-muted-foreground hover:text-foreground'}"
						onclick={() => (exportFormat = 'csv')}
					>
						<FileText class="h-3.5 w-3.5" /> CSV
					</button>
					<button
						class="inline-flex items-center gap-1.5 rounded-md px-3 py-1.5 text-xs font-medium transition-colors {exportFormat ===
						'pdf'
							? 'bg-card text-foreground shadow-xs'
							: 'text-muted-foreground hover:text-foreground'}"
						onclick={() => (exportFormat = 'pdf')}
					>
						<FileText class="h-3.5 w-3.5" /> PDF
					</button>
					<button
						class="inline-flex items-center gap-1.5 rounded-md px-3 py-1.5 text-xs font-medium transition-colors {exportFormat ===
						'excel'
							? 'bg-card text-foreground shadow-xs'
							: 'text-muted-foreground hover:text-foreground'}"
						onclick={() => (exportFormat = 'excel')}
					>
						<FileSpreadsheet class="h-3.5 w-3.5" /> Excel
					</button>
				</div>
				<Button
					class="{btnPrimary} gap-2"
					onclick={startExport}
					disabled={exportStatus === 'pending' || exportStatus === 'processing'}
				>
					{#if exportStatus === 'pending' || exportStatus === 'processing'}
						<Loader2 class="h-4 w-4 animate-spin" />
						Exporting...
					{:else}
						<Download class="h-4 w-4" />
						Export
					{/if}
				</Button>
				{#if exportStatus === 'complete'}
					<span class="badge-success">Ready</span>
				{:else if exportStatus === 'failed'}
					<span class="badge-danger">Failed</span>
				{/if}
			</div>
		</div>
	</div>

	<!-- Reorder suggestions -->
	<div class="panel overflow-hidden">
		<div
			class="panel-header flex-col items-stretch gap-3 sm:flex-row sm:items-center sm:justify-between"
		>
			<div class="min-w-0">
				<h2 class="section-title flex items-center gap-2">
					<ShoppingCart class="h-4 w-4 text-primary" />
					{$t('intelligence.reorder_suggestions.title')}
				</h2>
				<p class="mt-0.5 text-xs text-muted-foreground">
					{$t('intelligence.reorder_suggestions.subtitle')}
				</p>
			</div>
			<Button
				variant="outline"
				class="{btnSecondary} h-9 gap-2 px-3 text-xs font-semibold"
				onclick={async () => {
					suggestionsLoading = true;
					try {
						await replenishmentApi.generateSuggestions();
						await loadSuggestions();
						toast.success($t('intelligence.toasts.suggestions_refreshed'));
					} catch (e) {
						toast.error($t('intelligence.toasts.refresh_fail'));
					} finally {
						suggestionsLoading = false;
					}
				}}
				disabled={suggestionsLoading}
			>
				<RefreshCw class={suggestionsLoading ? 'h-4 w-4 animate-spin' : 'h-4 w-4'} />
				{$t('intelligence.reorder_suggestions.refresh_btn')}
			</Button>
		</div>
		<div class="overflow-x-auto">
			<table class="ledger-table">
				<thead>
					<tr>
						<th>{$t('intelligence.reorder_suggestions.table.product')}</th>
						<th>{$t('intelligence.reorder_suggestions.table.supplier')}</th>
						<th class="text-right">
							<span class="inline-flex items-center gap-1">
								{$t('intelligence.reorder_suggestions.table.suggested_qty')}
								<HelpTip
									text="Recommended reorder quantity based on recent sales velocity and lead time."
								/>
							</span>
						</th>
						<th>{$t('intelligence.reorder_suggestions.table.status')}</th>
						<th class="text-right">{$t('intelligence.reorder_suggestions.table.actions')}</th>
					</tr>
				</thead>
				<tbody>
					{#if suggestionsLoading}
						{#each Array(4) as _, i}
							<tr>
								<td colspan={5} class="p-3"><Skeleton class="h-6 w-full bg-muted" /></td>
							</tr>
						{/each}
					{:else if suggestions.length === 0}
						<tr>
							<td colspan={5} class="py-6 text-center text-sm text-muted-foreground"
								>{$t('intelligence.reorder_suggestions.table.empty')}</td
							>
						</tr>
					{:else}
						{#each suggestions as suggestion}
							<tr>
								<td class="text-sm font-medium text-foreground"
									>{suggestion.Product?.Name ?? `Product ${suggestion.ProductID}`}</td
								>
								<td class="text-sm text-muted-foreground"
									>{suggestion.Supplier?.Name ?? String(suggestion.SupplierID)}</td
								>
								<td class="col-num">{suggestion.SuggestedOrderQuantity}</td>
								<td>
									<span class="badge-warning capitalize">
										{suggestion.Status}
									</span>
								</td>
								<td class="text-right">
									<!-- Responsive action: single button stays compact -->
									<Button
										variant="outline"
										class="{btnSecondary} h-8 gap-1.5 px-2.5 text-xs font-semibold"
										onclick={() => createPO(suggestion.ID)}
									>
										{$t('intelligence.reorder_suggestions.table.create_po')}
									</Button>
								</td>
							</tr>
						{/each}
					{/if}
				</tbody>
			</table>
		</div>
	</div>

	<!-- Reports payloads -->

	<div class="grid gap-4 md:grid-cols-2 lg:grid-cols-3" data-animate="fade-up">
		<!-- SALES REPORT -->
		{#if auth.hasPermission('reports.sales')}
			<div class="panel">
				<div class="panel-header">
					<div class="min-w-0">
						<h2 class="section-title flex items-center gap-2">
							<TrendingUp class="h-4 w-4 text-primary" />
							{$t('intelligence.reports.sales.title')}
						</h2>
						<p class="mt-0.5 text-xs text-muted-foreground">
							{$t('intelligence.reports.sales.subtitle')}
						</p>
					</div>
					<button
						class="-mr-1 inline-flex size-8 shrink-0 items-center justify-center rounded-md text-muted-foreground transition-colors hover:bg-muted hover:text-foreground"
						onclick={() => exportJSON(reportResults.sales, 'sales-report')}
						aria-label="Export sales report"
					>
						<Download class="h-4 w-4" />
					</button>
				</div>
				<div class="panel-body space-y-3">
					<!-- Simple Line Chart (SVG sparkline) -->
					<div class="relative h-24 w-full">
						<svg
							viewBox="0 0 100 40"
							preserveAspectRatio="none"
							class="absolute inset-0 h-full w-full"
							style="color: hsl(var(--primary))"
						>
							<polyline
								points={generateSparkline(reportResults.sales?.salesTrends)}
								fill="none"
								stroke="currentColor"
								stroke-width="2.5"
								stroke-linecap="round"
							/>
						</svg>
					</div>
					<div class="flex justify-between text-sm text-muted-foreground">
						<p>{$t('intelligence.reports.sales.total_sales')}</p>
						<p class="num-mono font-semibold text-foreground">
							{fmt(reportResults.sales?.totalSales)}
						</p>
					</div>
					<div class="flex justify-between text-sm text-muted-foreground">
						<p>{$t('intelligence.reports.sales.avg_daily_sales')}</p>
						<p class="num-mono font-semibold text-foreground">
							{fmt(reportResults.sales?.averageDailySales)}
						</p>
					</div>
					<p class="text-xs text-muted-foreground">
						{$t('intelligence.reports.period', { period: reportResults.sales?.period ?? '—' })}
					</p>
				</div>
			</div>
		{/if}

		<!-- TURNOVER REPORT -->
		{#if auth.hasPermission('reports.inventory')}
			<div class="panel">
				<div class="panel-header">
					<div class="min-w-0">
						<h2 class="section-title flex items-center gap-2">
							<RotateCcw class="h-4 w-4 text-primary" />
							{$t('intelligence.reports.turnover.title')}
						</h2>
						<p class="mt-0.5 text-xs text-muted-foreground">
							{$t('intelligence.reports.turnover.subtitle')}
						</p>
					</div>
					<button
						class="-mr-1 inline-flex size-8 shrink-0 items-center justify-center rounded-md text-muted-foreground transition-colors hover:bg-muted hover:text-foreground"
						onclick={() => exportJSON(reportResults.turnover, 'turnover-report')}
						aria-label="Export turnover report"
					>
						<Download class="h-4 w-4" />
					</button>
				</div>
				<div class="panel-body space-y-3">
					<!-- Bar Chart -->
					<div class="flex h-24 w-full items-end gap-1">
						{#each [40, 25, 30, 15, 35, 22, 28] as height, i}
							<div
								class="flex-1 rounded-t-[3px] transition-all duration-500 ease-in-out hover:scale-105"
								style={`height: ${height}%; background-color: hsl(var(--primary) / ${i === 6 ? 1 : 0.3});`}
							></div>
						{/each}
					</div>
					<div class="flex justify-between text-sm text-muted-foreground">
						<p>{$t('intelligence.reports.turnover.avg_inventory_value')}</p>
						<p class="num-mono font-semibold text-foreground">
							${fmt(reportResults.turnover?.averageInventoryValue)}
						</p>
					</div>
					<div class="flex justify-between text-sm text-muted-foreground">
						<p>{$t('intelligence.reports.turnover.turnover_rate')}</p>
						<p class="num-mono font-semibold text-foreground">
							{fmt(reportResults.turnover?.inventoryTurnoverRate)}
						</p>
					</div>
					<p class="text-xs text-muted-foreground">
						{$t('intelligence.reports.period', { period: reportResults.turnover?.period ?? '—' })}
					</p>
				</div>
			</div>
		{/if}

		<!-- MARGIN REPORT -->
		{#if auth.hasPermission('reports.financial')}
			<div class="panel">
				<div class="panel-header">
					<div class="min-w-0">
						<h2 class="section-title flex items-center gap-2">
							<Percent class="h-4 w-4 text-primary" />
							{$t('intelligence.reports.margin.title')}
						</h2>
						<p class="mt-0.5 text-xs text-muted-foreground">
							{$t('intelligence.reports.margin.subtitle')}
						</p>
					</div>
					<button
						class="-mr-1 inline-flex size-8 shrink-0 items-center justify-center rounded-md text-muted-foreground transition-colors hover:bg-muted hover:text-foreground"
						onclick={() => exportJSON(reportResults.margin, 'margin-report')}
						aria-label="Export margin report"
					>
						<Download class="h-4 w-4" />
					</button>
				</div>
				<div class="panel-body space-y-3">
					<!-- Semi-donut Chart -->
					<div class="relative flex h-24 items-center justify-center">
						<svg viewBox="0 0 36 18" class="h-12 w-24 rotate-180">
							<path
								d="M2 18 A16 16 0 0 1 34 18"
								fill="none"
								stroke="hsl(var(--muted))"
								stroke-width="4"
								stroke-linecap="round"
							/>
							<path
								d="M2 18 A16 16 0 0 1 34 18"
								fill="none"
								stroke="hsl(var(--primary))"
								stroke-width="4"
								stroke-linecap="round"
								stroke-dasharray="{(reportResults.margin?.grossProfitMargin ?? 0) * 0.5},100"
							/>
						</svg>
						<span class="num-mono absolute text-lg font-semibold text-primary">
							{fmt(reportResults.margin?.grossProfitMargin)}%
						</span>
					</div>
					<div class="flex justify-between text-sm text-muted-foreground">
						<p>{$t('intelligence.reports.margin.gross_profit')}</p>
						<p class="num-mono font-semibold text-foreground">
							${fmt(reportResults.margin?.grossProfit)}
						</p>
					</div>
					<div class="flex justify-between text-sm text-muted-foreground">
						<p>{$t('intelligence.reports.margin.total_revenue')}</p>
						<p class="num-mono font-semibold text-foreground">
							${fmt(reportResults.margin?.totalRevenue)}
						</p>
					</div>
					<p class="text-xs text-muted-foreground">
						{$t('intelligence.reports.period', { period: reportResults.margin?.period ?? '—' })}
					</p>
				</div>
			</div>
		{/if}
	</div>
</section>
