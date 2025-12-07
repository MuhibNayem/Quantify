<!-- client/src/routes/purchase-orders/[id]/+page.svelte -->
<script lang="ts">
	import { page } from '$app/stores';
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
	import { ArrowLeft, ShoppingCart } from 'lucide-svelte';
	import type { PageData } from './$types';
	import { cn } from '$lib/utils';

	import { replenishmentApi } from '$lib/api/resources';
	import { toast } from 'svelte-sonner';
	import { invalidateAll } from '$app/navigation';
	import { Button } from '$lib/components/ui/button';

	export let data: PageData;

	const { purchaseOrder } = data;

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
</script>

<div class="mx-auto w-full max-w-6xl px-6 py-8">
	<section class="space-y-8">
		<!-- HEADER -->
		<div class="flex items-center justify-between">
			<a
				href="/intelligence"
				class="flex items-center text-blue-600 transition-colors hover:text-blue-800"
			>
				<ArrowLeft class="mr-2 h-5 w-5" />
				Back to Intelligence
			</a>
			{#if purchaseOrder && !['CANCELLED', 'COMPLETED', 'RECEIVED'].includes(purchaseOrder.Status)}
				<Button
					class="bg-red-600 font-semibold text-white shadow-md transition-all hover:bg-red-700 hover:shadow-lg"
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
				<div class="lg:col-span-1">
					<Card
						class="overflow-hidden rounded-2xl border-0 bg-gradient-to-br from-blue-50 to-indigo-100 shadow-lg transition-all duration-300 hover:scale-[1.01] hover:shadow-xl"
					>
						<CardHeader
							class="space-y-1 border-b border-white/60 bg-white/70 px-6 py-5 backdrop-blur"
						>
							<CardTitle class="flex items-center text-slate-800">
								<ShoppingCart class="mr-2 h-5 w-5 text-blue-600" />
								Purchase Order #{purchaseOrder.ID}
							</CardTitle>
							<CardDescription class="text-slate-600">Details and Status</CardDescription>
						</CardHeader>
						<CardContent class="space-y-4 p-6 text-sm">
							<div>
								<p class="font-medium text-slate-500">Supplier</p>
								<a
									href="/suppliers/{purchaseOrder.SupplierID}"
									class="text-lg font-semibold text-slate-800 hover:underline"
								>
									{purchaseOrder.Supplier?.Name || `Supplier #${purchaseOrder.SupplierID}`}
								</a>
							</div>
							<div>
								<p class="font-medium text-slate-500">Status</p>
								<span
									class={cn(
										'rounded-full border px-2.5 py-1 text-xs capitalize shadow-sm',
										purchaseOrder.Status === 'DRAFT'
											? 'border-gray-200 bg-gray-100 text-gray-700'
											: purchaseOrder.Status === 'APPROVED'
												? 'border-green-200 bg-green-100 text-green-700'
												: purchaseOrder.Status === 'SENT'
													? 'border-blue-200 bg-blue-100 text-blue-700'
													: 'border-red-200 bg-red-100 text-red-700'
									)}
								>
									{purchaseOrder.Status}
								</span>
							</div>
							<div>
								<p class="font-medium text-slate-500">Order Date</p>
								<p class="text-slate-800">
									{new Date(purchaseOrder.OrderDate).toLocaleDateString()}
								</p>
							</div>
							<div>
								<p class="font-medium text-slate-500">Expected Delivery</p>
								<p class="text-slate-800">
									{purchaseOrder.ExpectedDeliveryDate
										? new Date(purchaseOrder.ExpectedDeliveryDate).toLocaleDateString()
										: 'N/A'}
								</p>
							</div>
						</CardContent>
					</Card>
				</div>

				<!-- PO Items -->
				<div class="lg:col-span-2">
					<Card
						class="overflow-hidden rounded-2xl border-0 bg-gradient-to-br from-gray-50 to-slate-100 shadow-lg transition-all duration-300 hover:scale-[1.01] hover:shadow-xl"
					>
						<CardHeader
							class="space-y-1 border-b border-white/60 bg-white/70 px-6 py-5 backdrop-blur"
						>
							<CardTitle class="text-slate-800">Items</CardTitle>
						</CardHeader>
						<CardContent class="p-0">
							<Table>
								<TableHeader>
									<TableRow>
										<TableHead>Product</TableHead>
										<TableHead>Ordered</TableHead>
										<TableHead>Received</TableHead>
										<TableHead>Unit Price</TableHead>
									</TableRow>
								</TableHeader>
								<TableBody>
									{#if purchaseOrder.PurchaseOrderItems && purchaseOrder.PurchaseOrderItems.length > 0}
										{#each purchaseOrder.PurchaseOrderItems as item}
											<TableRow>
												<TableCell>
													<a href="/products/{item.ProductID}" class="hover:underline">
														{item.Product?.Name || `Product ID: ${item.ProductID}`}
													</a>
												</TableCell>
												<TableCell>{item.OrderedQuantity}</TableCell>
												<TableCell>{item.ReceivedQuantity}</TableCell>
												<TableCell>${item.UnitPrice.toFixed(2)}</TableCell>
											</TableRow>
										{/each}
									{:else}
										<TableRow>
											<TableCell colspan="4" class="py-4 text-center text-slate-500">
												No items in this purchase order.
											</TableCell>
										</TableRow>
									{/if}
								</TableBody>
							</Table>
						</CardContent>
					</Card>
				</div>
			</div>
		{:else}
			<!-- Loading State -->
			<div class="grid gap-8 lg:grid-cols-3">
				<div class="lg:col-span-1">
					<Skeleton class="h-64 w-full" />
				</div>
				<div class="lg:col-span-2">
					<Skeleton class="h-64 w-full" />
				</div>
			</div>
		{/if}
	</section>
</div>
