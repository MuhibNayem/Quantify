<!-- client/src/routes/purchase-orders/[id]/+page.svelte -->
<script lang="ts">
	import { page } from '$app/stores';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import { ArrowLeft, ShoppingCart } from 'lucide-svelte';
	import { cn, poNumber } from '$lib/utils';
	import Breadcrumbs from '$lib/components/ui/breadcrumbs.svelte';
	import ActivityHistory from '$lib/components/activity-history.svelte';

	import { replenishmentApi } from '$lib/api/resources';
	import { toast } from 'svelte-sonner';
	import { invalidateAll, goto } from '$app/navigation';
	import { Button } from '$lib/components/ui/button';
	import { auth } from '$lib/stores/auth';

	// Types
	import type { PageData } from './$types';

	let { data }: { data: PageData } = $props();

	let purchaseOrder = $derived(data.purchaseOrder);

	$effect(() => {
		if (!auth.hasPermission('replenishment.read')) {
			toast.error('Access Denied', {
				description: 'You do not have permission to view purchase orders.'
			});
			goto('/');
		}
	});

	let isCancelling = false;

	async function handleCancelPO() {
		if (
			!confirm('Are you sure you want to cancel this Purchase Order? This action cannot be undone.')
		)
			return;

		isCancelling = true;
		try {
			await replenishmentApi.cancelPO(purchaseOrder.ID);
			toast.success('Purchase Order cancelled successfully');
			await invalidateAll();
		} catch (e: any) {
			console.error(e);
			toast.error('Failed to cancel PO', { description: e?.response?.data?.error || e.message });
		} finally {
			isCancelling = false;
		}
	}

	function getStatusBadge(status: string) {
		switch (status) {
			case 'APPROVED':
			case 'SENT':
				return 'badge-info';
			case 'RECEIVED':
				return 'badge-success';
			case 'CANCELLED':
				return 'badge-danger';
			default:
				return 'badge-neutral';
		}
	}
</script>

<div class="mx-auto w-full max-w-6xl px-4 py-8 sm:px-6 lg:px-8 lg:py-10">
	<section class="space-y-4">
		{#if purchaseOrder}
			<Breadcrumbs
				items={[
					{ label: 'Purchase Orders', href: '/purchase-orders' },
					{ label: poNumber(purchaseOrder.ID) }
				]}
			/>
		{/if}

		<!-- HEADER -->
		<header class="page-header">
			<a
				href="/intelligence"
				class="flex items-center gap-2 text-sm font-medium text-muted-foreground transition-colors hover:text-foreground"
			>
				<span
					class="flex h-8 w-8 items-center justify-center rounded-md border border-border/70 bg-card shadow-xs"
				>
					<ArrowLeft class="h-4 w-4" />
				</span>
				Back to Intelligence
			</a>
			{#if purchaseOrder && !['CANCELLED', 'COMPLETED', 'RECEIVED'].includes(purchaseOrder.Status)}
				<Button variant="destructive" size="sm" onclick={handleCancelPO} disabled={isCancelling}>
					{isCancelling ? 'Cancelling...' : 'Cancel Order'}
				</Button>
			{/if}
		</header>

		{#if purchaseOrder}
			<div class="grid gap-4 lg:grid-cols-3">
				<!-- PO Details -->
				<div class="space-y-4 lg:col-span-1">
					<section class="panel overflow-hidden" data-animate="fade-up">
						<div class="panel-header items-start">
							<div class="flex items-start gap-3">
								<span
									class="flex h-9 w-9 shrink-0 items-center justify-center rounded-md bg-accent text-accent-foreground ring-1 ring-accent-foreground/10"
								>
									<ShoppingCart class="h-4 w-4" />
								</span>
								<div class="min-w-0 space-y-0.5">
									<h2 class="num-mono section-title truncate">{poNumber(purchaseOrder.ID)}</h2>
									<p class="text-sm text-muted-foreground">Details and Status</p>
								</div>
							</div>
						</div>
						<div class="panel-body space-y-3.5 text-sm">
							<div>
								<p class="mb-1 overline">Supplier</p>
								<a
									href="/suppliers/{purchaseOrder.SupplierID}"
									class="text-base font-semibold text-foreground transition-colors hover:text-primary hover:underline"
								>
									{`Supplier #${purchaseOrder.SupplierID}`}
								</a>
							</div>
							<div>
								<p class="mb-1 overline">Status</p>
								<span class={cn('uppercase', getStatusBadge(purchaseOrder.Status))}>
									{purchaseOrder.Status}
								</span>
							</div>
							<div>
								<p class="mb-1 overline">Order Date</p>
								<p class="num-mono text-sm text-foreground">
									{new Date(purchaseOrder.OrderDate).toLocaleDateString()}
								</p>
							</div>
							<div>
								<p class="mb-1 overline">Expected Delivery</p>
								<p class="num-mono text-sm text-foreground">
									{purchaseOrder.ExpectedDeliveryDate
										? new Date(purchaseOrder.ExpectedDeliveryDate).toLocaleDateString()
										: 'N/A'}
								</p>
							</div>
						</div>
					</section>

					<section
						class="panel overflow-hidden"
						data-animate="fade-up"
						style="animation-delay: 100ms;"
					>
						<div class="panel-header">
							<h2 class="section-title">History</h2>
						</div>
						<div class="panel-body pt-3">
							<ActivityHistory entity="PurchaseOrder" entityId={purchaseOrder.ID} />
						</div>
					</section>
				</div>

				<!-- PO Items -->
				<div class="lg:col-span-2">
					<section
						class="panel overflow-hidden"
						data-animate="fade-up"
						style="animation-delay: 60ms;"
					>
						<div class="panel-header">
							<h2 class="section-title">Items</h2>
						</div>
						<div class="overflow-x-auto">
							<table class="ledger-table">
								<thead>
									<tr>
										<th>Product</th>
										<th class="text-right">Ordered</th>
										<th class="text-right">Received</th>
										<th class="text-right">Unit Price</th>
									</tr>
								</thead>
								<tbody>
									{#if purchaseOrder.PurchaseOrderItems && purchaseOrder.PurchaseOrderItems.length > 0}
										{#each purchaseOrder.PurchaseOrderItems as item}
											<tr>
												<td>
													<a
														href="/products/{item.ProductID}"
														class="font-medium text-foreground transition-colors hover:text-primary hover:underline"
													>
														{item.Product?.Name || `Product ID: ${item.ProductID}`}
													</a>
												</td>
												<td class="col-num">{item.OrderedQuantity}</td>
												<td class="col-num">{item.ReceivedQuantity}</td>
												<td class="col-num">${item.UnitPrice.toFixed(2)}</td>
											</tr>
										{/each}
									{:else}
										<tr>
											<td colspan={4} class="py-8 text-center text-sm text-muted-foreground">
												No items in this purchase order.
											</td>
										</tr>
									{/if}
								</tbody>
							</table>
						</div>
					</section>
				</div>
			</div>
		{:else}
			<!-- Loading State -->
			<div class="grid gap-4 lg:grid-cols-3">
				<div class="lg:col-span-1">
					<Skeleton class="h-64 w-full rounded-lg bg-muted" />
				</div>
				<div class="lg:col-span-2">
					<Skeleton class="h-64 w-full rounded-lg bg-muted" />
				</div>
			</div>
		{/if}
	</section>
</div>
