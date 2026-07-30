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
	import { Area, AreaChart, LinearGradient, PieChart, Tooltip } from 'layerchart';
	import { curveMonotoneX } from 'd3-shape';

	let loading = $state(true);
	const stats = $state({ products: 0, categories: 0, suppliers: 0, alerts: 0 });
	let recentProducts = $state<Product[]>([]);
	let recentAlerts = $state<Alert[]>([]);
	let suggestions = $state<ReorderSuggestion[]>([]);
	let trend = $state({ direction: 'neutral', percentage: 0 });

	let chartSeries = $state<number[]>([]);
	let chartRevealed = $state(false);
	let mixContext = $state<any>(undefined);

	const loadDashboard = async () => {
		loading = true;
		chartRevealed = false;
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

	/* ── Demand Pulse chart data ── */
	const chartPeak = $derived(Math.max(...chartSeries, 1));
	const peakDay = $derived(chartSeries.indexOf(chartPeak) + 1);
	const chartAvg = $derived(chartSeries.length ? demandTotal / chartSeries.length : 0);

	/* Draw-in: mount flat at the baseline, then tween up to the real values */
	$effect(() => {
		if (!loading && chartSeries.length > 0 && !chartRevealed) {
			let raf2 = 0;
			const raf1 = requestAnimationFrame(() => {
				raf2 = requestAnimationFrame(() => {
					chartRevealed = true;
				});
			});
			return () => {
				cancelAnimationFrame(raf1);
				cancelAnimationFrame(raf2);
			};
		}
	});

	const demandData = $derived.by(() => {
		const revealed = chartRevealed || reduceMotion;
		return chartSeries.map((value, i) => ({ day: i + 1, value: revealed ? value : 0 }));
	});

	const dayLabel = (day: number) =>
		$t('dashboard.demand.day_label').replace('{day}', String(day));

	/* ── Alert mix (donut) ── */
	type Severity = 'critical' | 'warning' | 'info';

	const SEVERITY_COLORS: Record<Severity, string> = {
		critical: 'hsl(var(--danger))',
		warning: 'hsl(var(--warning))',
		info: 'hsl(var(--primary))'
	};

	function alertSeverity(item: Alert): Severity {
		const probe = `${item.Type} ${item.Status}`.toLowerCase();
		if (/(out|expired|critical|danger|high)/.test(probe)) return 'critical';
		if (/(low|warn|expir)/.test(probe)) return 'warning';
		return 'info';
	}

	const alertMix = $derived.by((): { key: Severity; label: string; value: number }[] => {
		const counts: Record<Severity, number> = { critical: 0, warning: 0, info: 0 };
		for (const item of recentAlerts) counts[alertSeverity(item)]++;
		return [
			{ key: 'critical', label: 'Critical', value: counts.critical },
			{ key: 'warning', label: 'Warning', value: counts.warning },
			{ key: 'info', label: 'Info', value: counts.info }
		];
	});
	const mixTotal = $derived(alertMix.reduce((a, s) => a + s.value, 0));

	function highlightSeverity(key: Severity | null) {
		if (mixContext) mixContext.series.highlightKey = key;
	}

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
	<div class="liquid-hoverable group p-4 lg:p-5" data-animate="fade-up">
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

{#snippet demandMarks()}
	<LinearGradient
		id="demand-pulse-gradient"
		vertical
		stops={['hsl(var(--primary) / 0.3)', 'hsl(var(--primary) / 0.02)']}
	/>
	<Area
		seriesKey="demand"
		curve={curveMonotoneX}
		fill="url(#demand-pulse-gradient)"
		fillOpacity={1}
		motion={reduceMotion ? { type: 'none' } : { type: 'tween', duration: 900 }}
		line={{ curve: curveMonotoneX, strokeWidth: 2 }}
	/>
{/snippet}

{#snippet demandTooltip({ context }: { context: any })}
	<Tooltip.Root
		{context}
		x="data"
		y="pointer"
		anchor="bottom-left"
		xOffset={14}
		yOffset={14}
		variant="none"
		class="ledger-tip"
	>
		{#snippet children({ data })}
			<p class="ledger-tip-label">{dayLabel(Number(context.x(data)))}</p>
			<p class="ledger-tip-value num-mono">
				{Math.round(Number(context.y(data))).toLocaleString()}
				<span>units</span>
			</p>
		{/snippet}
	</Tooltip.Root>
{/snippet}

{#snippet mixTooltip({ context }: { context: any })}
	<Tooltip.Root {context} anchor="bottom-left" xOffset={12} yOffset={10} variant="none" class="ledger-tip">
		{#snippet children({ data })}
			<p class="ledger-tip-label">
				<span
					class="mr-1.5 inline-block h-2 w-2 rounded-full"
					style="background:{SEVERITY_COLORS[data.key as Severity]}"
				></span>
				{data.label}
			</p>
			<p class="ledger-tip-value num-mono">
				{data.value}
				<span>{mixTotal > 0 ? Math.round((data.value / mixTotal) * 100) : 0}%</span>
			</p>
		{/snippet}
	</Tooltip.Root>
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

	<!-- ── DEMAND PULSE HERO ── -->
	{#if auth.hasPermission('reports.sales')}
		<section class="liquid-panel mt-4" data-animate="fade-up" aria-label="Demand pulse">
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

			{#if loading}
				<div class="px-5 py-6">
					<div class="flex flex-col gap-6 lg:flex-row lg:items-stretch">
						<div class="shrink-0 space-y-3 lg:w-56">
							<div class="bg-muted animate-pulse rounded-md h-3 w-24">&nbsp;</div>
							<div class="bg-muted animate-pulse rounded-md h-10 w-36">&nbsp;</div>
							<div class="bg-muted animate-pulse rounded-md h-3 w-28">&nbsp;</div>
						</div>
						<div class="flex h-60 flex-1 items-end gap-2 sm:h-64 lg:h-72">
							{#each [42, 68, 55, 80, 38, 72, 60, 88, 46, 76, 52, 84, 64, 70] as h}
								<div class="bg-muted animate-pulse rounded-md flex-1" style="height:{h}%">&nbsp;</div>
							{/each}
						</div>
					</div>
				</div>
			{:else if chartSeries.length === 0}
				<div class="text-muted-foreground flex h-60 flex-col items-center justify-center gap-2 text-sm sm:h-64">
					<ShoppingCart class="h-6 w-6 opacity-50" />
					{$t('dashboard.procurement.no_data')}
				</div>
			{:else}
				<div class="flex flex-col gap-6 px-5 py-6 lg:flex-row lg:items-stretch">
					<div class="flex shrink-0 flex-row items-end justify-between gap-6 lg:w-56 lg:flex-col lg:items-start lg:justify-center">
						<div>
							<p class="overline">Total demand</p>
							<div class="kpi-num mt-1.5 text-4xl lg:text-[2.75rem]">
								{demandCount().toLocaleString()}
							</div>
							<p class="text-muted-foreground mt-1.5 text-xs">
								units · last <span class="num font-semibold">{chartSeries.length}</span> days
							</p>
						</div>
						<div class="w-full space-y-1.5 border-t border-[hsl(var(--border))] pt-2.5 lg:mt-2">
							<div class="flex items-baseline justify-between gap-4">
								<span class="text-muted-foreground text-xs">Peak day</span>
								<span class="num-mono text-xs font-semibold">
									{dayLabel(peakDay)} · {chartPeak.toLocaleString()}
								</span>
							</div>
							<div class="flex items-baseline justify-between gap-4">
								<span class="text-muted-foreground text-xs">Daily average</span>
								<span class="num-mono text-xs font-semibold">
									{Math.round(chartAvg).toLocaleString()}
								</span>
							</div>
						</div>
					</div>

					<div class="min-w-0 flex-1">
						<AreaChart
							data={demandData}
							x="day"
							y="value"
							series={[
								{ key: 'demand', label: 'Demand', value: 'value', color: 'hsl(var(--primary))' }
							]}
							xDomain={[1, Math.max(chartSeries.length, 2)]}
							yDomain={[0, chartPeak * 1.12]}
							padding={{ top: 16, right: 8, bottom: 4, left: 8 }}
							axis="x"
							grid={{ y: { stroke: 'hsl(var(--border))', dashArray: '2 4' } }}
							highlight={{
								lines: { stroke: 'hsl(var(--primary) / 0.4)', dashArray: '3 3' },
								points: { r: 4, fill: 'hsl(var(--primary))', stroke: 'hsl(var(--card))', strokeWidth: 2 }
							}}
							tooltip={demandTooltip}
							marks={demandMarks}
							props={{
								xAxis: {
									tickMarks: false,
									tickSpacing: 72,
									format: 'integer',
									tickLabelProps: { fill: 'hsl(var(--muted-foreground))' }
								}
							}}
							class="h-60 w-full sm:h-64 lg:h-72"
						/>
						<div class="text-muted-foreground mt-2 flex justify-between text-[0.6875rem]">
							<span>{dayLabel(1)}</span>
							<span class="hidden sm:inline">{$t('dashboard.demand.chart_hint')}</span>
							<span>{dayLabel(chartSeries.length)}</span>
						</div>
					</div>
				</div>
			{/if}
		</section>
	{/if}

	<!-- ── ATTENTION ROW: alerts + mix + fresh inventory ── -->
	<section class="mt-4 grid grid-cols-1 gap-4 xl:grid-cols-3">
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

			<div class="liquid-panel flex flex-col" data-animate="fade-up">
				<div class="border-b border-[hsl(var(--border))] px-5 py-4">
					<p class="overline">By severity</p>
					<h2 class="font-display mt-0.5 text-base font-semibold">Alert mix</h2>
				</div>
				<div class="flex flex-1 flex-col justify-center px-5 py-5">
					{#if loading}
						<div class="bg-muted mx-auto aspect-square w-full max-w-40 animate-pulse rounded-full">&nbsp;</div>
						<div class="mx-auto mt-4 w-full max-w-44 space-y-2">
							{#each Array(3) as _}
								<div class="bg-muted animate-pulse rounded-md h-4">&nbsp;</div>
							{/each}
						</div>
					{:else if recentAlerts.length === 0}
						<div class="text-muted-foreground flex min-h-40 flex-col items-center justify-center gap-2 text-sm">
							<AlertTriangle class="text-success h-6 w-6 opacity-60" />
							{$t('dashboard.priority_alerts.no_data')}
						</div>
					{:else}
						<div class="relative mx-auto aspect-square w-full max-w-40">
							<PieChart
								bind:context={mixContext}
								data={alertMix}
								value="value"
								label="label"
								cDomain={['critical', 'warning', 'info']}
								cRange={[
									SEVERITY_COLORS.critical,
									SEVERITY_COLORS.warning,
									SEVERITY_COLORS.info
								]}
								innerRadius={0.72}
								padAngle={0.03}
								cornerRadius={2}
								padding={4}
								tooltip={mixTooltip}
								class="absolute inset-0"
							/>
							<div class="pointer-events-none absolute inset-0 flex flex-col items-center justify-center">
								<span class="kpi-num text-2xl">{mixTotal}</span>
								<span class="overline">alerts</span>
							</div>
						</div>
						<ul class="mx-auto mt-4 w-full max-w-44 space-y-1">
							{#each alertMix as s (s.key)}
								<li>
									<button
										type="button"
										class="hover:bg-muted/60 flex w-full items-center gap-2 rounded-md px-2 py-1 transition-colors"
										onpointerenter={() => highlightSeverity(s.key)}
										onpointerleave={() => highlightSeverity(null)}
										onfocus={() => highlightSeverity(s.key)}
										onblur={() => highlightSeverity(null)}
									>
										<span
											class="h-2 w-2 shrink-0 rounded-full"
											style="background:{SEVERITY_COLORS[s.key]}"
											aria-hidden="true"
										></span>
										<span class="text-xs">{s.label}</span>
										<span class="num-mono ml-auto text-xs font-semibold">{s.value}</span>
									</button>
								</li>
							{/each}
						</ul>
					{/if}
				</div>
			</div>
		{/if}

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
	</section>

	<!-- ── PROCUREMENT ── -->
	{#if auth.hasPermission('replenishment.read')}
		<section class="liquid-panel mt-4 overflow-hidden" data-animate="fade-up">
			<div class="border-b border-[hsl(var(--border))] px-5 py-4">
				<p class="overline">{$t('dashboard.procurement.subtitle')}</p>
				<h2 class="font-display mt-0.5 text-base font-semibold">
					{$t('dashboard.procurement.title')}
				</h2>
			</div>
			<div class="custom-scrollbar max-h-96 overflow-y-auto">
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
		</section>
	{/if}
</div>

<style>
	:global(.ledger-tip) {
		background: hsl(var(--card));
		border: 1px solid hsl(var(--border));
		border-radius: var(--radius);
		box-shadow: var(--shadow-md);
		padding: 0.5rem 0.625rem;
		min-width: 6.5rem;
	}
	:global(.ledger-tip .ledger-tip-label) {
		display: flex;
		align-items: center;
		font-size: 0.6875rem;
		font-weight: 600;
		letter-spacing: 0.08em;
		text-transform: uppercase;
		color: hsl(var(--muted-foreground));
	}
	:global(.ledger-tip .ledger-tip-value) {
		margin-top: 0.125rem;
		font-size: 0.9375rem;
		font-weight: 600;
		color: hsl(var(--foreground));
	}
	:global(.ledger-tip .ledger-tip-value span) {
		margin-left: 0.25rem;
		font-size: 0.6875rem;
		font-weight: 500;
		color: hsl(var(--muted-foreground));
	}
</style>
