<!-- client/src/routes/products/[sku]/+page.svelte -->
<script lang="ts">
	import { page } from '$app/stores';
	import { toast } from 'svelte-sonner';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import { ArrowLeft, Info, History } from 'lucide-svelte';
	import Breadcrumbs from '$lib/components/ui/breadcrumbs.svelte';
	import ActivityHistory from '$lib/components/activity-history.svelte';
	import type { PageData } from './$types';
	import { auth } from '$lib/stores/auth';
	import { goto } from '$app/navigation';
	import { addRecent } from '$lib/recent';

	$effect(() => {
		if (!auth.hasPermission('products.read')) {
			toast.error('Access Denied', {
				description: 'You do not have permission to view this product.'
			});
			goto('/');
		}
	});

	let { data }: { data: PageData } = $props();

	let product = $derived(data.product);
	let stockHistory = $derived(data.stockHistory);

	$effect(() => {
		if (product?.ID) {
			addRecent({
				type: 'product',
				id: product.ID,
				label: product.Name,
				href: `/products/${encodeURIComponent(product.SKU || String(product.ID))}`
			});
		}
	});

	const statusBadge = (status?: string) => {
		if (!status) return undefined;
		const normalized = status.toLowerCase();
		if (normalized === 'active') return { text: status, variant: 'success' as const };
		if (normalized === 'inactive') return { text: status, variant: 'warning' as const };
		if (normalized === 'archived') return { text: status, variant: 'danger' as const };
		return { text: status, variant: 'info' as const };
	};
</script>

<div class="mx-auto w-full max-w-7xl px-4 py-8 sm:px-6 lg:px-8">
	{#if product}
		{@const badge = statusBadge(product.Status)}
		<div class="space-y-6">
			<Breadcrumbs
				items={[
					{ label: 'Catalog', href: '/catalog' },
					{ label: 'Products', href: '/catalog' },
					{ label: product.Name }
				]}
			/>

			<!-- HEADER -->
			<div class="page-header">
				<div class="min-w-0 space-y-2">
					<p class="overline">Catalog · Product</p>
					<h1 class="page-title truncate">{product.Name}</h1>
					<div class="flex flex-wrap items-center gap-3">
						<span class="num-mono text-sm text-muted-foreground">{product.SKU}</span>
						{#if badge}
							<span class="badge-{badge.variant} capitalize">{badge.text}</span>
						{/if}
					</div>
				</div>
				<div class="flex flex-wrap items-center gap-2">
					<a
						href="/catalog"
						class="btn-soft inline-flex h-10 items-center gap-2 px-4 text-sm font-medium"
					>
						<ArrowLeft class="h-4 w-4" />
						Back to Catalog
					</a>
				</div>
			</div>

			<div class="grid gap-6 lg:grid-cols-3">
				<!-- Product Details -->
				<div class="min-w-0 lg:col-span-2">
					<div class="panel overflow-hidden">
						<div class="panel-header">
							<div class="flex items-center gap-2">
								<Info class="h-4 w-4 text-primary" />
								<h2 class="section-title">Product Details</h2>
							</div>
						</div>
						<div class="panel-body">
							<dl class="grid grid-cols-1 gap-x-6 gap-y-5 text-sm sm:grid-cols-2">
								<div class="min-w-0">
									<dt class="overline">Description</dt>
									<dd class="mt-1 text-foreground">{product.Description || 'N/A'}</dd>
								</div>
								<div class="min-w-0">
									<dt class="overline">Brand</dt>
									<dd class="mt-1 text-foreground">{product.Brand || 'N/A'}</dd>
								</div>
								<div class="min-w-0">
									<dt class="overline">Category</dt>
									<dd class="mt-1 text-foreground">{product.Category?.Name || 'N/A'}</dd>
								</div>
								<div class="min-w-0">
									<dt class="overline">Sub-Category</dt>
									<dd class="mt-1 text-foreground">{product.SubCategory?.Name || 'N/A'}</dd>
								</div>
								<div class="min-w-0">
									<dt class="overline">Supplier</dt>
									<dd class="mt-1 text-foreground">{product.Supplier?.Name || 'N/A'}</dd>
								</div>
								<div class="min-w-0">
									<dt class="overline">Location</dt>
									<dd class="mt-1 text-foreground">{product.Location?.Name || 'N/A'}</dd>
								</div>
								<div class="min-w-0">
									<dt class="overline">Purchase Price</dt>
									<dd class="num-mono mt-1 text-foreground">
										${product.PurchasePrice?.toFixed(2) || 'N/A'}
									</dd>
								</div>
								<div class="min-w-0">
									<dt class="overline">Selling Price</dt>
									<dd class="num-mono mt-1 text-foreground">
										${product.SellingPrice?.toFixed(2) || 'N/A'}
									</dd>
								</div>
								<div class="min-w-0">
									<dt class="overline">Barcode/UPC</dt>
									<dd class="num-mono mt-1 text-foreground">{product.BarcodeUPC || 'N/A'}</dd>
								</div>
								<div class="min-w-0">
									<dt class="overline">Status</dt>
									<dd class="mt-1.5">
										{#if product.Status}
											{#if badge}
												<span class="badge-{badge.variant} capitalize">
													{badge.text}
												</span>
											{:else}
												<span class="text-muted-foreground">—</span>
											{/if}
										{:else}
											<span class="text-muted-foreground">—</span>
										{/if}
									</dd>
								</div>
							</dl>
						</div>
					</div>
				</div>

				<!-- Stock History -->
				<div class="min-w-0 space-y-6 lg:col-span-1">
					<div class="panel overflow-hidden">
						<div class="panel-header">
							<div class="flex items-center gap-2">
								<History class="h-4 w-4 text-primary" />
								<h2 class="section-title">Stock History</h2>
							</div>
							<span class="text-xs text-muted-foreground">Recent</span>
						</div>
						<div class="custom-scrollbar max-h-[26rem] overflow-y-auto">
							<table class="glass-table">
								<thead>
									<tr>
										<th>Type</th>
										<th class="text-right">Qty</th>
										<th>Date</th>
									</tr>
								</thead>
								<tbody>
									{#if (stockHistory || []).length === 0}
										<tr>
											<td colspan={3} class="py-10 text-center text-sm text-muted-foreground">
												No stock adjustments yet
											</td>
										</tr>
									{:else}
										{#each stockHistory || [] as item}
											<tr>
												<td class="font-medium text-foreground">{item.Type}</td>
												<td
													class="col-num font-semibold {item.Type === 'STOCK_IN'
														? 'text-success'
														: 'text-danger'}"
												>
													{item.Type === 'STOCK_IN' ? '+' : '-'}{item.Quantity}
												</td>
												<td class="text-xs text-muted-foreground">
													{new Date(item.AdjustedAt).toLocaleDateString()}
												</td>
											</tr>
										{/each}
									{/if}
								</tbody>
							</table>
						</div>
					</div>

					<div class="panel p-5">
						<ActivityHistory entity="Product" entityId={product.ID} />
					</div>
				</div>
			</div>
		</div>
	{:else}
		<!-- Loading State -->
		<div class="space-y-6">
			<Skeleton class="h-9 w-64 rounded-lg" />
			<div class="grid gap-6 lg:grid-cols-3">
				<div class="lg:col-span-2">
					<Skeleton class="h-96 w-full rounded-xl" />
				</div>
				<div class="lg:col-span-1">
					<Skeleton class="h-96 w-full rounded-xl" />
				</div>
			</div>
		</div>
	{/if}
</div>
