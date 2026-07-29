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

<div class="mx-auto w-full max-w-6xl px-6 py-8">
	<section class="space-y-8">
		{#if purchaseOrder}
			<Breadcrumbs
				items={[
					{ label: 'Purchase Orders', href: '/purchase-orders' },
					{ label: poNumber(purchaseOrder.ID) }
				]}
			/>
		{/if}
		<!-- HEADER -->
		<div class="flex items-center justify-between">
			<a
				href="/intelligence"
				class="text-primary hover:text-primary/80 flex items-center font-medium transition-colors"
			>
				<ArrowLeft class="mr-2 h-5 w-5" />
				Back to Intelligence
			</a>
			{#if purchaseOrder && !['CANCELLED', 'COMPLETED', 'RECEIVED'].includes(purchaseOrder.Status)}
				<Button
					class="bg-danger text-primary-foreground font-semibold shadow-md hover:bg-danger/90"
					size="sm"
					onclick={handleCancelPO}
					disabled={isCancelling}
				>
					{isCancelling ? 'Cancelling...' : 'Cancel Order'}
				</Button>
			{/if}
		</div>

		{#if purchaseOrder}
			<div class="grid gap-8 lg:grid-cols-3">
				<!-- PO Details -->
				<div class="space-y-8 lg:col-span-1">
					<div class="liquid-panel overflow-hidden rounded-lg" data-animate="fade-up">
						<div class="border-border bg-muted/40 space-y-1 border-b px-6 py-5">
							<h2 class="font-display text-foreground flex items-center text-lg font-bold">
								<ShoppingCart class="text-primary mr-2 h-5 w-5" />
								{poNumber(purchaseOrder.ID)}
							</h2>
							<p class="text-muted-foreground text-sm">Details and Status</p>
						</div>
						<div class="space-y-4 p-6 text-sm">
							<div>
								<p class="overline mb-1">Supplier</p>
								<a
									href="/suppliers/{purchaseOrder.SupplierID}"
									class="text-foreground hover:text-primary text-lg font-semibold transition-colors hover:underline"
								>
									{`Supplier #${purchaseOrder.SupplierID}`}
								</a>
							</div>
							<div>
								<p class="overline mb-1">Status</p>
								<span class={cn('uppercase', getStatusBadge(purchaseOrder.Status))}>
									{purchaseOrder.Status}
								</span>
							</div>
							<div>
								<p class="overline mb-1">Order Date</p>
								<p class="text-foreground num-mono">
									{new Date(purchaseOrder.OrderDate).toLocaleDateString()}
								</p>
							</div>
							<div>
								<p class="overline mb-1">Expected Delivery</p>
								<p class="text-foreground num-mono">
									{purchaseOrder.ExpectedDeliveryDate
										? new Date(purchaseOrder.ExpectedDeliveryDate).toLocaleDateString()
										: 'N/A'}
								</p>
							</div>
						</div>
					</div>

					<div class="liquid-panel rounded-lg p-4" data-animate="fade-up" style="animation-delay: 100ms;">
						<ActivityHistory entity="PurchaseOrder" entityId={purchaseOrder.ID} />
					</div>
				</div>

				<!-- PO Items -->
				<div class="lg:col-span-2">
					<div class="liquid-panel overflow-hidden rounded-lg" data-animate="fade-up" style="animation-delay: 60ms;">
						<div class="border-border bg-muted/40 border-b px-6 py-5">
							<h2 class="font-display text-foreground text-lg font-bold">Items</h2>
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
														class="text-foreground hover:text-primary transition-colors hover:underline"
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
											<td colspan={4} class="text-muted-foreground py-4 text-center">
												No items in this purchase order.
											</td>
										</tr>
									{/if}
								</tbody>
							</table>
						</div>
					</div>
				</div>
			</div>
		{:else}
			<!-- Loading State -->
			<div class="grid gap-8 lg:grid-cols-3">
				<div class="lg:col-span-1">
					<Skeleton class="bg-muted h-64 w-full rounded-lg" />
				</div>
				<div class="lg:col-span-2">
					<Skeleton class="bg-muted h-64 w-full rounded-lg" />
				</div>
			</div>
		{/if}
	</section>
</div>
