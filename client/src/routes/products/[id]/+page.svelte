<!-- client/src/routes/products/[id]/+page.svelte -->
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
				href: `/products/${product.ID}`
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

<div class="mx-auto w-full max-w-7xl px-6 py-8">
	<section class="space-y-8">
		{#if product}
			<Breadcrumbs
				items={[
					{ label: 'Catalog', href: '/catalog' },
					{ label: 'Products', href: '/catalog' },
					{ label: product.Name }
				]}
			/>
		{/if}
		<!-- HEADER -->
		<div class="flex items-center justify-between">
			<a
				href="/catalog"
				class="text-primary hover:text-foreground flex items-center text-sm font-medium transition-colors"
			>
				<ArrowLeft class="mr-2 h-4 w-4" />
				Back to Catalog
			</a>
		</div>

		{#if product}
			<div class="grid gap-8 lg:grid-cols-3">
				<!-- Product Details -->
				<div class="lg:col-span-2">
					<div class="liquid-panel overflow-hidden rounded-lg">
						<div class="border-border space-y-1 border-b px-6 py-5">
							<h1 class="text-foreground flex items-center font-display text-xl font-semibold">
								<Info class="text-primary mr-2 h-5 w-5" />
								{product.Name}
							</h1>
							<p class="num-mono text-muted-foreground text-sm">{product.SKU}</p>
						</div>
						<div class="space-y-4 p-6">
							<div class="grid grid-cols-1 gap-x-6 gap-y-4 text-sm sm:grid-cols-2">
								<div>
									<p class="overline">Description</p>
									<p class="text-foreground mt-0.5">{product.Description || 'N/A'}</p>
								</div>
								<div>
									<p class="overline">Brand</p>
									<p class="text-foreground mt-0.5">{product.Brand || 'N/A'}</p>
								</div>
								<div>
									<p class="overline">Category</p>
									<p class="text-foreground mt-0.5">{product.Category?.Name || 'N/A'}</p>
								</div>
								<div>
									<p class="overline">Sub-Category</p>
									<p class="text-foreground mt-0.5">{product.SubCategory?.Name || 'N/A'}</p>
								</div>
								<div>
									<p class="overline">Supplier</p>
									<p class="text-foreground mt-0.5">{product.Supplier?.Name || 'N/A'}</p>
								</div>
								<div>
									<p class="overline">Location</p>
									<p class="text-foreground mt-0.5">{product.Location?.Name || 'N/A'}</p>
								</div>
								<div>
									<p class="overline">Purchase Price</p>
									<p class="num-mono text-foreground mt-0.5">
										${product.PurchasePrice?.toFixed(2) || 'N/A'}
									</p>
								</div>
								<div>
									<p class="overline">Selling Price</p>
									<p class="num-mono text-foreground mt-0.5">${product.SellingPrice?.toFixed(2) || 'N/A'}</p>
								</div>
								<div>
									<p class="overline">Barcode/UPC</p>
									<p class="num-mono text-foreground mt-0.5">{product.BarcodeUPC || 'N/A'}</p>
								</div>
								<div>
									<p class="overline">Status</p>
									<div class="mt-0.5">
										{#if product.Status}
											{@const badge = statusBadge(product.Status)}
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
									</div>
								</div>
							</div>
						</div>
					</div>
				</div>

				<!-- Stock History -->
				<div class="lg:col-span-1">
					<div class="flex flex-col gap-4">
						<div class="liquid-panel flex flex-col gap-1 rounded-lg p-4">
							<h3 class="text-foreground flex items-center font-display text-lg font-semibold">
								<History class="text-primary mr-2 h-5 w-5" />
								Stock History
							</h3>
							<p class="text-muted-foreground text-sm">Recent stock adjustments</p>
						</div>

						<div class="liquid-panel overflow-hidden rounded-lg">
							<table class="ledger-table">
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
											<td colspan={3} class="text-muted-foreground py-8 text-center text-sm">
												No stock adjustments yet
											</td>
										</tr>
									{:else}
										{#each stockHistory || [] as item}
											<tr>
												<td class="text-foreground font-medium">{item.Type}</td>
												<td
													class="col-num font-semibold {item.Type === 'STOCK_IN'
														? 'text-success'
														: 'text-danger'}"
												>
													{item.Type === 'STOCK_IN' ? '+' : '-'}{item.Quantity}
												</td>
												<td class="text-muted-foreground">
													{new Date(item.AdjustedAt).toLocaleDateString()}
												</td>
											</tr>
										{/each}
									{/if}
								</tbody>
							</table>
						</div>

						<div class="liquid-panel rounded-lg p-4 transition-shadow duration-300 hover:shadow-md">
							<ActivityHistory entity="Product" entityId={product.ID} />
						</div>
					</div>
				</div>
			</div>
		{:else}
			<!-- Loading State -->
			<div class="grid gap-8 lg:grid-cols-3">
				<div class="lg:col-span-2">
					<Skeleton class="h-96 w-full" />
				</div>
				<div class="lg:col-span-1">
					<Skeleton class="h-96 w-full" />
				</div>
			</div>
		{/if}
	</section>
</div>
