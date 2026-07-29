<script lang="ts">
	import { onMount } from 'svelte';
	import { Button } from '$lib/components/ui/button';
	import { toast } from 'svelte-sonner';
	import { dashboardApi } from '$lib/api/resources';
	import type { Alert, Product, ReorderSuggestion } from '$lib/types';
	import { t } from '$lib/i18n';
	import { auth } from '$lib/stores/auth';
	import { formatDate, timeAgo } from '$lib/utils';
	import { AlertTriangle, Boxes, Plus, RefreshCcw, ShoppingCart } from 'lucide-svelte';

	let loading = $state(true);
	const stats = $state({ products: 0, categories: 0, suppliers: 0, alerts: 0 });
	let recentProducts = $state<Product[]>([]);
	let recentAlerts = $state<Alert[]>([]);
	let suggestions = $state<ReorderSuggestion[]>([]);
	let trend = $state({ direction: 'neutral', percentage: 0 });

	let chartSeries = $state<number[]>([]);
	let chartMax = $derived(Math.max(...chartSeries, 10));

	const loadDashboard = async () => {
		loading = true;
		try {
			const data = await dashboardApi.getSummary();

			stats.products = data.stats.products;
			stats.categories = data.stats.categories;
			stats.suppliers = data.stats.suppliers;
			stats.alerts = data.stats.alerts;

			recentProducts = data.recentProducts;
			recentAlerts = data.recentAlerts;
			suggestions = data.suggestions;
			chartSeries = data.chartData || [];
			trend = data.trend || { direction: 'neutral', percentage: 0 };
		} catch (error: any) {
			toast.error($t('dashboard.toasts.load_fail'), {
				description: error?.response?.data?.error || $t('dashboard.toasts.error_desc')
			});
		} finally {
			loading = false;
		}
	};

	/* ── Count-up: animates 0 → target whenever target changes ── */
	const reduceMotion =
		typeof window !== 'undefined' &&
		window.matchMedia('(prefers-reduced-motion: reduce)').matches;

	function countUp(getTarget: () => number, duration = 900) {
		let value = $state(0);
		$effect(() => {
			const target = getTarget();
			if (reduceMotion) {
				value = target;
				return;
			}
			let raf = 0;
			const start = performance.now();
			const tick = (now: number) => {
				const p = Math.min((now - start) / duration, 1);
				const eased = 1 - Math.pow(1 - p, 3);
				value = target * eased;
				if (p < 1) raf = requestAnimationFrame(tick);
			};
			raf = requestAnimationFrame(tick);
			return () => cancelAnimationFrame(raf);
		});
		return () => Math.round(value);
	}

	const demandTotal = $derived(chartSeries.reduce((a, b) => a + b, 0));
	const reorderUnits = $derived(
		suggestions.reduce((a, s) => a + (s.SuggestedOrderQuantity || 0), 0)
	);

	const productsCount = countUp(() => stats.products);
	const demandCount = countUp(() => demandTotal);
	const alertsCount = countUp(() => stats.alerts);
	const reorderCount = countUp(() => suggestions.length);

	/* ── Sparkline bar heights (%) ── */
	function sparkFromSeries(series: number[]): number[] {
		const max = Math.max(...series, 1);
		return series.map((v) => Math.max(8, (v / max) * 100));
	}
	function sparkBars(seed: number, n = 10): number[] {
		let s = Math.abs(Math.round(seed)) + 11;
		const out: number[] = [];
		for (let i = 0; i < n; i++) {
			s = (s * 9301 + 49297) % 233280;
			out.push(22 + (s / 233280) * 78);
		}
		return out;
	}
	const demandSpark = $derived(
		chartSeries.length ? sparkFromSeries(chartSeries) : sparkBars(7)
	);
	const productsSpark = $derived(sparkBars(stats.products));
	const alertsSpark = $derived(sparkBars(stats.alerts + 3));
	const reorderSpark = $derived(sparkBars(suggestions.length + 5));

	function alertBadgeClass(item: Alert): string {
		const probe = `${item.Type} ${item.Status}`.toLowerCase();
		if (/(out|expired|critical|danger|high)/.test(probe)) return 'badge-danger';
		if (/(low|warn|expir)/.test(probe)) return 'badge-warning';
		return 'badge-neutral';
	}

	onMount(loadDashboard);
</script>

{#snippet kpiCard(label: string, value: string, bars: number[], deltaClass: string, deltaText: string, subText: string)}
	<div class="liquid-panel group p-4 lg:p-5" data-animate="fade-up">
		<p class="overline">{label}</p>
		<div class="mt-3 flex items-end justify-between gap-3">
			<div class="min-w-0">
				<div class="kpi-num text-[1.75rem] lg:text-[2rem]">{value}</div>
				<p class="mt-1.5 truncate text-xs">
					<span class="font-semibold {deltaClass}">{deltaText}</span>
					<span class="text-muted-foreground"> {subText}</span>
				</p>
			</div>
			<div class="spark w-20 shrink-0" aria-hidden="true">
				{#each bars as h, i}
					<span
						style="height:{h}%; transition-delay:{i * 20}ms"
						class="transition-all duration-500 group-hover:bg-primary/60"
					></span>
				{/each}
			</div>
		</div>
	</div>
{/snippet}

{#snippet tableSkeleton(cols: number)}
	{#each Array(5) as _}
		<tr>
			<td colspan={cols} class="py-3">
				<div class="bg-muted animate-pulse rounded-md px-3 py-2.5">&nbsp;</div>
			</td>
		</tr>
	{/each}
{/snippet}

{#snippet emptyRow(cols: number, message: string)}
	<tr>
		<td colspan={cols} class="text-muted-foreground py-10 text-center text-sm">
			{message}
		</td>
	</tr>
{/snippet}

<div class="mx-auto w-full max-w-7xl px-6 py-8">
	<!-- ── HEADER ── -->
	<header class="flex flex-wrap items-end justify-between gap-4" data-animate="fade-up">
		<div>
			<h1 class="font-display text-2xl font-bold tracking-tight">Overview</h1>
			<p class="text-muted-foreground mt-1 flex items-center gap-2 text-sm">
				<span class="num">{formatDate(new Date())}</span>
				<span class="bg-border inline-block h-3 w-px" aria-hidden="true"></span>
				<span class="live-dot" aria-hidden="true"></span>
				<span class="text-success text-xs font-semibold tracking-wide uppercase">Live</span>
			</p>
		</div>
		<div class="flex items-center gap-2">
			<Button
				variant="secondary"
				size="sm"
				onclick={loadDashboard}
				class="glass-button rounded-md text-muted-foreground hover:text-foreground"
				title={$t('dashboard.refresh')}
			>
				<RefreshCcw class="h-4 w-4" />
				{$t('dashboard.refresh')}
			</Button>
			<Button
				href="/pos"
				size="sm"
				class="bg-primary text-primary-foreground hover:bg-primary/90 rounded-md bg-none shadow-xs"
			>
				<Plus class="h-4 w-4" />
				New Sale
			</Button>
		</div>
	</header>

	<!-- ── KPI STRIP ── -->
	<section class="mt-6 grid grid-cols-2 gap-3 lg:grid-cols-4 lg:gap-4" aria-label="Key metrics">
		{#if loading}
			{#each Array(4) as _}
				<div class="liquid-panel space-y-3 p-4 lg:p-5">
					<div class="bg-muted animate-pulse rounded-md h-3 w-24">&nbsp;</div>
					<div class="bg-muted animate-pulse rounded-md h-8 w-20">&nbsp;</div>
					<div class="bg-muted animate-pulse rounded-md h-3 w-28">&nbsp;</div>
				</div>
			{/each}
		{:else}
			{@render kpiCard(
				$t('dashboard.stats.active_products'),
				productsCount().toLocaleString(),
				productsSpark,
				'text-muted-foreground',
				`${stats.categories}`,
				'categories'
			)}
			{@render kpiCard(
				'Demand pulse',
				demandCount().toLocaleString(),
				demandSpark,
				trend.direction === 'up'
					? 'text-success'
					: trend.direction === 'down'
						? 'text-danger'
						: 'text-muted-foreground',
				trend.direction === 'up'
					? `▲ ${trend.percentage.toFixed(1)}%`
					: trend.direction === 'down'
						? `▼ ${Math.abs(trend.percentage).toFixed(1)}%`
						: '→ 0.0%',
				'vs prior period'
			)}
			{@render kpiCard(
				$t('dashboard.stats.active_alerts'),
				alertsCount().toLocaleString(),
				alertsSpark,
				stats.alerts > 0 ? 'text-danger' : 'text-success',
				stats.alerts > 0 ? 'needs review' : 'all clear',
				stats.alerts > 0 ? 'in queue' : 'no action'
			)}
			{@render kpiCard(
				'Reorder queue',
				reorderCount().toLocaleString(),
				reorderSpark,
				'text-primary',
				reorderUnits > 0 ? `${reorderUnits.toLocaleString()} units` : '—',
				'suggested'
			)}
		{/if}
	</section>

	<!-- ── MAIN GRID: demand chart + attention ── -->
	<section class="mt-4 grid grid-cols-1 gap-4 xl:grid-cols-3">
		{#if auth.hasPermission('reports.sales')}
			<div class="liquid-panel xl:col-span-2" data-animate="fade-up">
				<div class="flex flex-wrap items-center justify-between gap-3 border-b border-[hsl(var(--border))] px-5 py-4">
					<div>
						<p class="overline">{$t('dashboard.demand.subtitle')}</p>
						<h2 class="font-display mt-0.5 text-base font-semibold">
							{$t('dashboard.demand.title')}
						</h2>
					</div>
					{#if !loading}
						{#if trend.direction === 'up'}
							<span class="badge-success num">▲ {trend.percentage.toFixed(1)}%</span>
						{:else if trend.direction === 'down'}
							<span class="badge-danger num">▼ {Math.abs(trend.percentage).toFixed(1)}%</span>
						{:else}
							<span class="badge-neutral">→ {$t('dashboard.demand.no_change')}</span>
						{/if}
					{/if}
				</div>
				<div class="px-5 py-5">
					{#if loading}
						<div class="flex h-44 items-end gap-2">
							{#each [42, 68, 55, 80, 38, 72, 60, 88, 46, 76, 52, 84, 64, 70] as h}
								<div class="bg-muted animate-pulse rounded-md flex-1" style="height:{h}%">&nbsp;</div>
							{/each}
						</div>
					{:else if chartSeries.length === 0}
						<div class="text-muted-foreground flex h-44 flex-col items-center justify-center gap-2 text-sm">
							<ShoppingCart class="h-6 w-6 opacity-50" />
							{$t('dashboard.procurement.no_data')}
						</div>
					{:else}
						<div class="border-border flex h-44 items-end gap-1.5 border-b pb-px">
							{#each chartSeries as value, i}
								<div class="group relative flex h-full flex-1 items-end">
									<div
										class="bg-primary/60 hover:bg-primary w-full rounded-t-sm transition-all duration-500 ease-out group-hover:-translate-y-0.5"
										style="height:{Math.max(4, (value / chartMax) * 100)}%"
										title="{$t('dashboard.demand.day_label').replace('{day}', String(i + 1))}: {value}"
										role="img"
										aria-label="{$t('dashboard.demand.day_label').replace('{day}', String(i + 1))}: {value}"
									></div>
								</div>
							{/each}
						</div>
						<div class="text-muted-foreground mt-2 flex justify-between text-[0.6875rem]">
							<span>{$t('dashboard.demand.day_label').replace('{day}', '1')}</span>
							<span class="hidden sm:inline">{$t('dashboard.demand.chart_hint')}</span>
							<span>{$t('dashboard.demand.day_label').replace('{day}', String(chartSeries.length))}</span>
						</div>
					{/if}
				</div>
			</div>
		{/if}

		{#if auth.hasPermission('alerts.view')}
			<div class="liquid-panel flex flex-col" data-animate="fade-up">
				<div class="flex items-center justify-between gap-3 border-b border-[hsl(var(--border))] px-5 py-4">
					<div>
						<p class="overline">{$t('dashboard.priority_alerts.subtitle')}</p>
						<h2 class="font-display mt-0.5 text-base font-semibold">Needs attention</h2>
					</div>
					{#if !loading && stats.alerts > 0}
						<span class="badge-danger num">{stats.alerts}</span>
					{/if}
				</div>
				<div class="custom-scrollbar max-h-80 flex-1 overflow-y-auto">
					{#if loading}
						<div class="space-y-2 p-4">
							{#each Array(5) as _}
								<div class="bg-muted animate-pulse rounded-md h-12">&nbsp;</div>
							{/each}
						</div>
					{:else if recentAlerts.length === 0}
						<div class="text-muted-foreground flex h-full min-h-40 flex-col items-center justify-center gap-2 p-6 text-sm">
							<AlertTriangle class="text-success h-6 w-6 opacity-60" />
							{$t('dashboard.priority_alerts.no_data')}
						</div>
					{:else}
						<ul class="divide-y divide-[hsl(var(--border)/0.7)]">
							{#each recentAlerts as item}
								<li class="hover:bg-muted/40 flex items-center justify-between gap-3 px-5 py-3 transition-colors">
									<div class="min-w-0">
										<p class="truncate text-sm font-medium">
											{item.Product?.Name ?? item.Product?.SKU ?? `#${item.ProductID}`}
										</p>
										<p class="text-muted-foreground num-mono truncate text-xs">
											{item.Product?.SKU ?? item.ProductID} · {timeAgo(item.TriggeredAt)}
										</p>
									</div>
									<span class={alertBadgeClass(item)}>{item.Type}</span>
								</li>
							{/each}
						</ul>
					{/if}
				</div>
				<a
					href="/alerts"
					class="text-primary hover:bg-accent border-t border-[hsl(var(--border))] px-5 py-2.5 text-center text-xs font-semibold transition-colors"
				>
					View all alerts →
				</a>
			</div>
		{/if}
	</section>

	<!-- ── SECONDARY GRID: fresh inventory + procurement ── -->
	<section class="mt-4 grid grid-cols-1 gap-4 xl:grid-cols-2">
		{#if auth.hasPermission('products.read')}
			<div class="liquid-panel overflow-hidden" data-animate="fade-up">
				<div class="border-b border-[hsl(var(--border))] px-5 py-4">
					<p class="overline">{$t('dashboard.fresh_inventory.subtitle')}</p>
					<h2 class="font-display mt-0.5 text-base font-semibold">
						{$t('dashboard.fresh_inventory.title')}
					</h2>
				</div>
				<div class="custom-scrollbar max-h-80 overflow-y-auto">
					<table class="ledger-table">
						<thead>
							<tr>
								<th>{$t('dashboard.fresh_inventory.sku')}</th>
								<th>{$t('dashboard.fresh_inventory.product_name')}</th>
								<th class="text-right">{$t('dashboard.fresh_inventory.status')}</th>
							</tr>
						</thead>
						<tbody>
							{#if loading}
								{@render tableSkeleton(3)}
							{:else if recentProducts.length === 0}
								{@render emptyRow(3, $t('dashboard.fresh_inventory.no_data'))}
							{:else}
								{#each recentProducts as product}
									<tr>
										<td class="num-mono text-primary text-[0.8125rem] font-medium">{product.SKU}</td>
										<td class="max-w-48 truncate">{product.Name}</td>
										<td class="text-right">
											<span class={(product.Status ?? 'active').toLowerCase() === 'active' ? 'badge-success' : 'badge-neutral'}>
												{product.Status ?? 'active'}
											</span>
										</td>
									</tr>
								{/each}
							{/if}
						</tbody>
					</table>
				</div>
			</div>
		{/if}

		{#if auth.hasPermission('replenishment.read')}
			<div class="liquid-panel overflow-hidden" data-animate="fade-up">
				<div class="border-b border-[hsl(var(--border))] px-5 py-4">
					<p class="overline">{$t('dashboard.procurement.subtitle')}</p>
					<h2 class="font-display mt-0.5 text-base font-semibold">
						{$t('dashboard.procurement.title')}
					</h2>
				</div>
				<div class="custom-scrollbar max-h-80 overflow-y-auto">
					<table class="ledger-table">
						<thead>
							<tr>
								<th>{$t('dashboard.procurement.product')}</th>
								<th class="text-right">On hand</th>
								<th class="text-right">{$t('dashboard.procurement.suggested_qty')}</th>
								<th>{$t('dashboard.procurement.supplier')}</th>
								<th class="text-right">{$t('dashboard.procurement.status')}</th>
							</tr>
						</thead>
						<tbody>
							{#if loading}
								{@render tableSkeleton(5)}
							{:else if suggestions.length === 0}
								{@render emptyRow(5, $t('dashboard.procurement.no_data'))}
							{:else}
								{#each suggestions as s}
									<tr>
										<td class="max-w-44 truncate font-medium">
											{s.Product?.Name ?? `Product ${s.ProductID}`}
										</td>
										<td class="col-num">{s.CurrentStock ?? '—'}</td>
										<td class="col-num text-primary font-semibold">{s.SuggestedOrderQuantity ?? '—'}</td>
										<td class="text-muted-foreground max-w-32 truncate">{s.Supplier?.Name ?? s.SupplierID}</td>
										<td class="text-right">
											<span class="badge-info">{s.Status ?? $t('dashboard.procurement.ready_to_order')}</span>
										</td>
									</tr>
								{/each}
							{/if}
						</tbody>
					</table>
				</div>
			</div>
		{/if}
	</section>
</div>
