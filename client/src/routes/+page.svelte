<script lang="ts">
	import { onMount } from 'svelte';
	import { Button } from '$lib/components/ui/button';
	import {
		Card,
		CardContent,
		CardDescription,
		CardHeader,
		CardTitle
	} from '$lib/components/ui/card';
	import {
		Table,
		TableBody,
		TableCell,
		TableHead,
		TableHeader,
		TableRow
	} from '$lib/components/ui/table';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import { toast } from 'svelte-sonner';
	import { dashboardApi } from '$lib/api/resources';
	import type { Alert, Product, ReorderSuggestion } from '$lib/types';
	import { t } from '$lib/i18n';
	import { auth } from '$lib/stores/auth';
	import {
		Activity,
		AlertTriangle,
		Boxes,
		RefreshCcw,
		TrendingUp,
		Zap,
		ShoppingCart,
		Users,
		BarChart3
	} from 'lucide-svelte';

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

	let demandGridClass = $derived(
		auth.hasPermission('reports.sales') ? 'lg:grid-cols-3' : 'lg:grid-cols-1'
	);
	let inventoryGridClass = $derived(
		auth.hasPermission('products.read') && auth.hasPermission('alerts.view')
			? 'lg:grid-cols-2'
			: 'lg:grid-cols-1'
	);

	onMount(loadDashboard);
</script>

<div class="mx-auto w-full max-w-7xl px-6 py-8">
	<section class="relative z-10 space-y-8">
		<!-- HEADER (matches Operations animations) -->
		<div
			class="animate-fadeUp relative overflow-hidden rounded-3xl border border-white/50 shadow-lg"
		>
			<div
				class="animate-gradientShift absolute inset-0 bg-gradient-to-r from-sky-50 via-blue-50 to-cyan-100"
			></div>
			<div class="relative rounded-3xl bg-white/70 p-8 backdrop-blur-lg">
				<div class="flex flex-wrap items-center justify-between gap-6">
					<div class="space-y-3">
						<div class="flex items-center gap-3">
							<div
								class="animate-pulseGlow rounded-2xl bg-gradient-to-br from-sky-400 to-blue-500 p-2 shadow-md"
							>
								<Zap class="h-6 w-6 text-white" />
							</div>
							<p class="text-sm font-semibold uppercase tracking-wider text-sky-600">
								Inventory Control Tower
							</p>
						</div>
						<h1
							class="bg-gradient-to-r from-sky-600 via-blue-600 to-cyan-600 bg-clip-text text-4xl font-bold text-transparent"
						>
							{$t('dashboard.title')}
						</h1>
						<p class="max-w-2xl text-slate-600">
							{$t('dashboard.subtitle')}
						</p>
					</div>
					<div class="flex flex-wrap gap-3">
						<Button
							variant="secondary"
							class="group rounded-xl border border-slate-200 bg-white/80 px-6 py-3 font-medium text-slate-700 transition-all duration-300 hover:scale-105 hover:border-sky-200 hover:bg-sky-50/70 hover:text-sky-700 hover:shadow-lg"
							onclick={loadDashboard}
						>
							<RefreshCcw
								class="mr-2 h-4 w-4 transition-transform duration-500 group-hover:rotate-180"
							/>
							{$t('dashboard.refresh')}
						</Button>
						<Button
							href="/catalog"
							class="rounded-xl bg-gradient-to-r from-sky-500 to-blue-500 px-6 py-3 font-semibold text-white shadow-lg transition-all duration-300 hover:scale-105 hover:from-sky-600 hover:to-blue-600 hover:shadow-xl"
						>
							<Boxes class="mr-2 h-4 w-4" />
							{$t('dashboard.update_catalog')}
						</Button>
					</div>
				</div>
			</div>
		</div>

		<!-- STATS GRID (staggered like Operations) -->
		<div class="grid gap-6 md:grid-cols-2 lg:grid-cols-4">
			<!-- Products -->
			<Card
				class="animate-fadeUp group relative cursor-pointer overflow-hidden border-0 bg-gradient-to-br from-blue-50 to-cyan-100 shadow-xl transition-all duration-500 hover:scale-105 hover:shadow-2xl"
				style="animation-delay: .05s;"
			>
				<div
					class="absolute right-0 top-0 h-20 w-20 -translate-y-10 translate-x-10 rounded-full bg-blue-200/30 transition-transform duration-500 group-hover:scale-150"
				></div>
				<CardHeader class="relative z-10 flex flex-row items-center justify-between pb-3">
					<CardTitle class="text-sm font-semibold text-blue-800/80">{$t('dashboard.stats.active_products')}</CardTitle>
					<div class="rounded-xl bg-white/50 p-2 shadow-sm">
						<ShoppingCart class="h-5 w-5 text-blue-600" />
					</div>
				</CardHeader>
				<CardContent class="relative z-10">
					{#if loading}
						<Skeleton class="h-8 w-20 bg-white/50" />
					{:else}
						<div class="mb-1 text-3xl font-bold text-blue-900">{stats.products}</div>
						<p class="text-xs font-medium text-blue-700/70">
							{$t('dashboard.stats.forecast_hint').replace('{value}', Math.round(stats.products * 1.12).toString())}
						</p>
					{/if}
				</CardContent>
			</Card>

			<!-- Categories -->
			<Card
				class="animate-fadeUp group relative cursor-pointer overflow-hidden border-0 bg-gradient-to-br from-green-50 to-emerald-100 shadow-xl transition-all duration-500 hover:scale-105 hover:shadow-2xl"
				style="animation-delay: .15s;"
			>
				<div
					class="absolute right-0 top-0 h-20 w-20 -translate-y-10 translate-x-10 rounded-full bg-green-200/30 transition-transform duration-500 group-hover:scale-150"
				></div>
				<CardHeader class="relative z-10 flex flex-row items-center justify-between pb-3">
					<CardTitle class="text-sm font-semibold text-green-800/80">{$t('dashboard.stats.categories')}</CardTitle>
					<div class="rounded-xl bg-white/50 p-2 shadow-sm">
						<BarChart3 class="h-5 w-5 text-green-600" />
					</div>
				</CardHeader>
				<CardContent class="relative z-10">
					{#if loading}
						<Skeleton class="h-8 w-20 bg-white/50" />
					{:else}
						<div class="mb-1 text-3xl font-bold text-green-900">{stats.categories}</div>
						<p class="text-xs font-medium text-green-700/70">
							{$t('dashboard.stats.supplier_hint').replace('{count}', stats.suppliers.toString())}
						</p>
					{/if}
				</CardContent>
			</Card>

			<!-- Suppliers -->
			<Card
				class="animate-fadeUp group relative cursor-pointer overflow-hidden border-0 bg-gradient-to-br from-purple-50 to-violet-100 shadow-xl transition-all duration-500 hover:scale-105 hover:shadow-2xl"
				style="animation-delay: .25s;"
			>
				<div
					class="absolute right-0 top-0 h-20 w-20 -translate-y-10 translate-x-10 rounded-full bg-purple-200/30 transition-transform duration-500 group-hover:scale-150"
				></div>
				<CardHeader class="relative z-10 flex flex-row items-center justify-between pb-3">
					<CardTitle class="text-sm font-semibold text-purple-800/80">{$t('dashboard.stats.suppliers')}</CardTitle>
					<div class="rounded-xl bg-white/50 p-2 shadow-sm">
						<Users class="h-5 w-5 text-purple-600" />
					</div>
				</CardHeader>
				<CardContent class="relative z-10">
					{#if loading}
						<Skeleton class="h-8 w-20 bg-white/50" />
					{:else}
						<div class="mb-1 text-3xl font-bold text-purple-900">{stats.suppliers}</div>
						<p class="text-xs font-medium text-purple-700/70">{$t('dashboard.stats.sla_hint')}</p>
					{/if}
				</CardContent>
			</Card>

			<!-- Alerts -->
			<Card
				class="animate-fadeUp group relative cursor-pointer overflow-hidden border-0 bg-gradient-to-br from-orange-50 to-red-100 shadow-xl transition-all duration-500 hover:scale-105 hover:shadow-2xl"
				style="animation-delay: .35s;"
			>
				<div
					class="absolute right-0 top-0 h-20 w-20 -translate-y-10 translate-x-10 rounded-full bg-red-200/30 transition-transform duration-500 group-hover:scale-150"
				></div>
				<CardHeader class="relative z-10 flex flex-row items-center justify-between pb-3">
					<CardTitle class="text-sm font-semibold text-red-800/80">{$t('dashboard.stats.active_alerts')}</CardTitle>
					<div class="rounded-xl bg-white/50 p-2 shadow-sm">
						<AlertTriangle class="h-5 w-5 text-red-600" />
					</div>
				</CardHeader>
				<CardContent class="relative z-10">
					{#if loading}
						<Skeleton class="h-8 w-20 bg-white/50" />
					{:else}
						<div class="mb-1 text-3xl font-bold text-red-900">{stats.alerts}</div>
						<p class="text-xs font-medium text-red-700/70">{$t('dashboard.stats.escalation_hint')}</p>
					{/if}
				</CardContent>
			</Card>
		</div>

		<!-- DEMAND PULSE + QUICK ACTIONS -->
		<div class="mt-8 grid gap-6 {demandGridClass}">
			<!-- Demand Pulse -->
			{#if auth.hasPermission('reports.sales')}
				<Card
					class="animate-fadeUp group rounded-2xl border border-blue-100 bg-gradient-to-br from-sky-50 to-blue-50 shadow-lg transition-all duration-500 hover:scale-[1.02] hover:shadow-xl lg:col-span-2"
				>
					<CardHeader class="pb-4">
						<CardTitle class="flex items-center gap-2 text-slate-800">
							<div class="rounded-xl bg-gradient-to-r from-sky-400 to-blue-400 p-2 shadow-sm">
								<Activity class="h-5 w-5 text-white" />
							</div>
							{$t('dashboard.demand.title')}
						</CardTitle>
						<CardDescription class="text-slate-600"
							>{$t('dashboard.demand.subtitle')}</CardDescription
						>
					</CardHeader>
					<CardContent>
						<div class="space-y-4">
							<div class="flex h-32 items-end gap-3">
								{#each chartSeries as value, index}
									<div class="group flex flex-1 flex-col items-center">
										<div
											class="animate-fadeUp relative w-full overflow-hidden rounded-t-2xl bg-gradient-to-t from-sky-400 to-blue-500 shadow-md transition-all duration-700 ease-out hover:scale-110 group-hover:brightness-110"
											style={`height: ${(value / chartMax) * 100}px; min-height: 8px; animation-delay: ${index * 0.08 + 0.1}s;`}
										>
											<div
												class="absolute inset-0 bg-gradient-to-b from-white/30 to-transparent"
											></div>
										</div>

										<span class="mt-2 text-xs font-medium text-slate-600">{$t('dashboard.demand.day_label').replace('{day}', (index + 1).toString())}</span>
									</div>
								{/each}
							</div>
							<div class="flex items-center justify-between border-t border-blue-100 pt-4">
								<p class="text-sm text-slate-600">{$t('dashboard.demand.chart_hint')}</p>
								<div class="flex gap-2">
									{#if trend.direction === 'up'}
										<span
											class="rounded-full border border-green-200 bg-green-50 px-3 py-1 text-xs font-medium text-green-700"
											>{$t('dashboard.demand.trend_positive')}</span
										>
										<span
											class="rounded-full border border-blue-200 bg-blue-50 px-3 py-1 text-xs font-medium text-blue-700"
											>{$t('dashboard.demand.growth').replace('{value}', trend.percentage.toFixed(1))}</span
										>
									{:else if trend.direction === 'down'}
										<span
											class="rounded-full border border-red-200 bg-red-50 px-3 py-1 text-xs font-medium text-red-700"
											>{$t('dashboard.demand.trend_negative')}</span
										>
										<span
											class="rounded-full border border-orange-200 bg-orange-50 px-3 py-1 text-xs font-medium text-orange-700"
											>{$t('dashboard.demand.decline').replace('{value}', Math.abs(trend.percentage).toFixed(1))}</span
										>
									{:else}
										<span
											class="rounded-full border border-slate-200 bg-slate-50 px-3 py-1 text-xs font-medium text-slate-700"
											>{$t('dashboard.demand.trend_stable')}</span
										>
										<span
											class="rounded-full border border-slate-200 bg-slate-50 px-3 py-1 text-xs font-medium text-slate-700"
											>{$t('dashboard.demand.no_change')}</span
										>
									{/if}
								</div>
							</div>
						</div>
					</CardContent>
				</Card>
			{/if}

			<!-- Quick Actions -->
			<Card
				class="animate-fadeUp group rounded-2xl border border-purple-100 bg-gradient-to-br from-violet-50 to-purple-50 shadow-lg transition-all duration-500 hover:scale-[1.02] hover:shadow-xl"
				style="animation-delay:.1s;"
			>
				<CardHeader>
					<CardTitle class="flex items-center gap-2 text-slate-800">
						<div class="rounded-xl bg-gradient-to-r from-violet-400 to-purple-400 p-2 shadow-sm">
							<Zap class="h-5 w-5 text-white" />
						</div>
						{$t('dashboard.quick_actions.title')}
					</CardTitle>
					<CardDescription class="text-slate-600">{$t('dashboard.quick_actions.subtitle')}</CardDescription>
				</CardHeader>
				<CardContent class="space-y-3">
					<Button
						class="group h-14 w-full justify-start rounded-xl bg-gradient-to-r from-emerald-400 to-green-500 font-semibold text-white shadow-md transition-all hover:scale-105 hover:from-emerald-500 hover:to-green-600 hover:shadow-lg"
						href="/operations"
					>
						<TrendingUp class="mr-3 h-5 w-5 transition-transform group-hover:scale-110" />
						<div class="text-left">
							<div class="font-semibold">{$t('dashboard.quick_actions.balance_stock')}</div>
							<div class="text-xs opacity-90">{$t('dashboard.quick_actions.balance_desc')}</div>
						</div>
					</Button>

					<Button
						class="group h-14 w-full justify-start rounded-xl bg-gradient-to-r from-sky-400 to-blue-500 font-semibold text-white shadow-md transition-all hover:scale-105 hover:from-sky-500 hover:to-blue-600 hover:shadow-lg"
						href="/intelligence"
					>
						<Activity class="mr-3 h-5 w-5 transition-transform group-hover:scale-110" />
						<div class="text-left">
							<div class="font-semibold">{$t('dashboard.quick_actions.run_forecast')}</div>
							<div class="text-xs opacity-90">{$t('dashboard.quick_actions.forecast_desc')}</div>
						</div>
					</Button>

					<Button
						class="group h-14 w-full justify-start rounded-xl bg-gradient-to-r from-violet-400 to-purple-500 font-semibold text-white shadow-md transition-all hover:scale-105 hover:from-violet-500 hover:to-purple-600 hover:shadow-lg"
						href="/bulk"
					>
						<Boxes class="mr-3 h-5 w-5 transition-transform group-hover:scale-110" />
						<div class="text-left">
							<div class="font-semibold">{$t('dashboard.quick_actions.export_catalog')}</div>
							<div class="text-xs opacity-90">{$t('dashboard.quick_actions.export_desc')}</div>
						</div>
					</Button>
				</CardContent>
			</Card>
		</div>

		<!-- FRESH INVENTORY + ALERTS -->
		<div class="grid gap-6 {inventoryGridClass}">
			<!-- Fresh Inventory -->
			{#if auth.hasPermission('products.read')}
				<Card
					class="animate-fadeUp group rounded-2xl border border-green-100 bg-gradient-to-br from-emerald-50 to-green-50 shadow-lg transition-all hover:scale-[1.01] hover:shadow-xl"
				>
					<CardHeader class="pb-4">
						<CardTitle class="flex items-center gap-2 text-slate-800">
							<div class="rounded-xl bg-gradient-to-r from-emerald-400 to-green-400 p-2 shadow-sm">
								<Boxes class="h-5 w-5 text-white" />
							</div>
							{$t('dashboard.fresh_inventory.title')}
						</CardTitle>
						<CardDescription class="text-slate-600">{$t('dashboard.fresh_inventory.subtitle')}</CardDescription>
					</CardHeader>
					<CardContent>
						<div class="scrollbar-gradient max-h-80 overflow-y-auto pr-2">
							<Table>
								<TableHeader class="sticky top-0 rounded-lg bg-white/90 backdrop-blur-sm">
									<TableRow class="border-b border-green-100">
										<TableHead class="py-3 font-semibold text-slate-700">{$t('dashboard.fresh_inventory.sku')}</TableHead>
										<TableHead class="py-3 font-semibold text-slate-700">{$t('dashboard.fresh_inventory.product_name')}</TableHead>
										<TableHead class="py-3 text-right font-semibold text-slate-700"
											>{$t('dashboard.fresh_inventory.status')}</TableHead
										>
									</TableRow>
								</TableHeader>
								<TableBody>
									{#if loading}
										{#each Array(5) as _, i}
											<TableRow
												class="animate-fadeUp transition-colors hover:bg-white/70"
												style={`animation-delay:${i * 0.08 + 0.1}s`}
											>
												<TableCell colspan="3" class="py-3"
													><Skeleton class="h-6 w-full bg-white/70" /></TableCell
												>
											</TableRow>
										{/each}
									{:else if recentProducts.length === 0}
										<TableRow>
											<TableCell colspan="3" class="py-8 text-center text-slate-500">
												<div class="space-y-2">
													<Boxes class="mx-auto h-8 w-8 text-slate-400" />
													<div>{$t('dashboard.fresh_inventory.no_data')}</div>
												</div>
											</TableCell>
										</TableRow>
									{:else}
										{#each recentProducts as product, i}
											<TableRow
												class="animate-fadeUp group transition-all duration-300 hover:bg-white/70"
												style={`animation-delay:${i * 0.08 + 0.1}s`}
											>
												<TableCell class="py-3 font-mono text-sm font-medium text-blue-600"
													>{product.SKU}</TableCell
												>
												<TableCell class="py-3 text-slate-700 group-hover:text-slate-900"
													>{product.Name}</TableCell
												>
												<TableCell class="py-3 text-right">
													<span
														class="inline-flex items-center rounded-full border border-emerald-200 bg-emerald-100 px-3 py-1.5 text-xs font-medium text-emerald-700 shadow-sm"
													>
														● {product.Status ?? 'active'}
													</span>
												</TableCell>
											</TableRow>
										{/each}
									{/if}
								</TableBody>
							</Table>
						</div>
					</CardContent>
				</Card>
			{/if}

			<!-- Priority Alerts -->
			{#if auth.hasPermission('alerts.view')}
				<Card
					class="animate-fadeUp group rounded-2xl border border-amber-100 bg-gradient-to-br from-orange-50 to-amber-50 shadow-lg transition-all hover:scale-[1.01] hover:shadow-xl"
					style="animation-delay:.1s;"
				>
					<CardHeader class="pb-4">
						<CardTitle class="flex items-center gap-2 text-slate-800">
							<div class="rounded-xl bg-gradient-to-r from-orange-400 to-amber-400 p-2 shadow-sm">
								<AlertTriangle class="h-5 w-5 text-white" />
							</div>
							{$t('dashboard.priority_alerts.title')}
						</CardTitle>
						<CardDescription class="text-slate-600">{$t('dashboard.priority_alerts.subtitle')}</CardDescription>
					</CardHeader>
					<CardContent>
						<div class="scrollbar-gradient max-h-80 overflow-y-auto pr-2">
							<Table>
								<TableHeader class="sticky top-0 rounded-lg bg-white/90 backdrop-blur-sm">
									<TableRow class="border-b border-amber-100">
										<TableHead class="py-3 font-semibold text-slate-700">{$t('dashboard.priority_alerts.type')}</TableHead>
										<TableHead class="py-3 font-semibold text-slate-700">{$t('dashboard.priority_alerts.product')}</TableHead>
										<TableHead class="py-3 text-right font-semibold text-slate-700"
											>{$t('dashboard.priority_alerts.status')}</TableHead
										>
									</TableRow>
								</TableHeader>
								<TableBody>
									{#if loading}
										{#each Array(5) as _, i}
											<TableRow
												class="animate-fadeUp transition-colors hover:bg-white/70"
												style={`animation-delay:${i * 0.08 + 0.1}s`}
											>
												<TableCell colspan="3" class="py-3"
													><Skeleton class="h-6 w-full bg-white/70" /></TableCell
												>
											</TableRow>
										{/each}
									{:else if recentAlerts.length === 0}
										<TableRow>
											<TableCell colspan="3" class="py-8 text-center text-slate-500">
												<div class="space-y-2">
													<AlertTriangle class="mx-auto h-8 w-8 text-emerald-400" />
													<div>{$t('dashboard.priority_alerts.no_data')}</div>
												</div>
											</TableCell>
										</TableRow>
									{:else}
										{#each recentAlerts as item, i}
											<TableRow
												class="animate-fadeUp group transition-all duration-300 hover:bg-white/70"
												style={`animation-delay:${i * 0.08 + 0.1}s`}
											>
												<TableCell class="py-3 font-medium text-slate-700">{item.Type}</TableCell>
												<TableCell class="py-3 text-slate-600"
													>{item.Product?.SKU ?? item.ProductID}</TableCell
												>
												<TableCell class="py-3 text-right">
													<span
														class="inline-flex items-center rounded-full border border-amber-200 bg-amber-100 px-3 py-1.5 text-xs font-medium text-amber-700 shadow-sm"
													>
														🚨 {item.Status}
													</span>
												</TableCell>
											</TableRow>
										{/each}
									{/if}
								</TableBody>
							</Table>
						</div>
					</CardContent>
				</Card>
			{/if}
		</div>

		<!-- PROCUREMENT INTELLIGENCE -->
		{#if auth.hasPermission('replenishment.read')}
			<Card
				class="animate-fadeUp group rounded-2xl border border-teal-100 bg-gradient-to-br from-cyan-50 to-teal-50 shadow-lg transition-all hover:scale-[1.01] hover:shadow-xl"
			>
				<CardHeader class="pb-4">
					<CardTitle class="flex items-center gap-2 text-slate-800">
						<div class="rounded-xl bg-gradient-to-r from-cyan-400 to-teal-400 p-2 shadow-sm">
							<ShoppingCart class="h-5 w-5 text-white" />
						</div>
						{$t('dashboard.procurement.title')}
					</CardTitle>
					<CardDescription class="text-slate-600">{$t('dashboard.procurement.subtitle')}</CardDescription>
				</CardHeader>
				<CardContent>
					<div class="scrollbar-gradient max-h-80 overflow-y-auto pr-2">
						<Table>
							<TableHeader class="sticky top-0 rounded-lg bg-white/90 backdrop-blur-sm">
								<TableRow class="border-b border-teal-100">
									<TableHead class="py-3 font-semibold text-slate-700">{$t('dashboard.procurement.product')}</TableHead>
									<TableHead class="py-3 text-right font-semibold text-slate-700"
										>{$t('dashboard.procurement.suggested_qty')}</TableHead
									>
									<TableHead class="py-3 font-semibold text-slate-700">{$t('dashboard.procurement.supplier')}</TableHead>
									<TableHead class="py-3 text-right font-semibold text-slate-700">{$t('dashboard.procurement.status')}</TableHead>
								</TableRow>
							</TableHeader>
							<TableBody>
								{#if loading}
									{#each Array(5) as _, i}
										<TableRow
											class="animate-fadeUp transition-colors hover:bg-white/70"
											style={`animation-delay:${i * 0.08 + 0.1}s`}
										>
											<TableCell colspan="4" class="py-3"
												><Skeleton class="h-6 w-full bg-white/70" /></TableCell
											>
										</TableRow>
									{/each}
								{:else if suggestions.length === 0}
									<TableRow>
										<TableCell colspan="4" class="py-8 text-center text-slate-500">
											<div class="space-y-2">
												<ShoppingCart class="mx-auto h-8 w-8 text-slate-400" />
												<div>{$t('dashboard.procurement.no_data')}</div>
											</div>
										</TableCell>
									</TableRow>
								{:else}
									{#each suggestions as suggestion, i}
										<TableRow
											class="animate-fadeUp group transition-all duration-300 hover:bg-white/70"
											style={`animation-delay:${i * 0.08 + 0.1}s`}
										>
											<TableCell class="py-3 font-medium text-slate-700">
												{suggestion?.Product?.Name ??
													suggestion?.product?.name ??
													`Product ${suggestion?.ProductID ?? suggestion?.productId ?? 'N/A'}`}
											</TableCell>
											<TableCell class="py-3 text-right font-semibold text-blue-600">
												{suggestion?.SuggestedOrderQuantity ??
													suggestion?.suggestedOrderQuantity ??
													suggestion?.quantity ??
													'N/A'}
											</TableCell>
											<TableCell class="py-3 text-slate-600">
												{suggestion?.Supplier?.Name ??
													suggestion?.supplier?.name ??
													suggestion?.SupplierID ??
													suggestion?.supplierId ??
													'N/A'}
											</TableCell>
											<TableCell class="py-3 text-right">
												<span
													class="inline-flex items-center rounded-full border border-cyan-200 bg-cyan-100 px-3 py-1.5 text-xs font-medium text-cyan-700 shadow-sm"
												>
													💡 {suggestion?.Status ?? suggestion?.status ?? $t('dashboard.procurement.ready_to_order')}
												</span>
											</TableCell>
										</TableRow>
									{/each}
								{/if}
							</TableBody>
						</Table>
					</div>
				</CardContent>
			</Card>
		{/if}
	</section>
</div>

<style lang="postcss">
	/* ===== Reused animation system from Operations ===== */
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
			transform: translateY(18px);
		}
		to {
			opacity: 1;
			transform: translateY(0);
		}
	}
	.animate-fadeUp {
		animation: fadeUp 1.2s ease forwards;
	}

	/* Scrollbar soft styling */
	.scrollbar-gradient {
		scrollbar-width: thin;
		scrollbar-color: rgba(139, 92, 246, 0.2) transparent;
	}
	.scrollbar-gradient::-webkit-scrollbar {
		width: 6px;
	}
	.scrollbar-gradient::-webkit-scrollbar-track {
		background: transparent;
		border-radius: 10px;
	}
	.scrollbar-gradient::-webkit-scrollbar-thumb {
		background: rgba(139, 92, 246, 0.2);
		border-radius: 10px;
	}
	.scrollbar-gradient::-webkit-scrollbar-thumb:hover {
		background: rgba(139, 92, 246, 0.3);
	}

	/* Smooth hover transitions */
	* {
		transition-property:
			color, background-color, border-color, text-decoration-color, fill, stroke, opacity,
			box-shadow, transform, filter, backdrop-filter;
		transition-timing-function: cubic-bezier(0.4, 0, 0.2, 1);
		transition-duration: 300ms;
	}

	@media (prefers-reduced-motion: reduce) {
		.animate-gradientShift,
		.animate-pulseGlow,
		.animate-fadeUp {
			animation: none !important;
		}
	}
</style>
