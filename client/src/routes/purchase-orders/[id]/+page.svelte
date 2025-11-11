<!-- client/src/routes/purchase-orders/[id]/+page.svelte -->
<script lang="ts">
	import { page } from '$app/stores';
	import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '$lib/components/ui/card';
	import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '$lib/components/ui/table';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import { ArrowLeft, ShoppingCart } from 'lucide-svelte';
	import type { PageData } from './$types';
	import { cn } from '$lib/utils';

	export let data: PageData;

	const { purchaseOrder } = data;
</script>

<div class="w-full max-w-6xl mx-auto py-8 px-6">
	<section class="space-y-8">
		<!-- HEADER -->
		<div class="flex items-center justify-between">
			<a href="/intelligence" class="flex items-center text-blue-600 hover:text-blue-800 transition-colors">
				<ArrowLeft class="h-5 w-5 mr-2" />
				Back to Intelligence
			</a>
		</div>

		{#if purchaseOrder}
			<div class="grid gap-8 lg:grid-cols-3">
				<!-- PO Details -->
				<div class="lg:col-span-1">
					<Card class="rounded-2xl shadow-lg hover:shadow-xl transition-all duration-300 hover:scale-[1.01] overflow-hidden border-0 bg-gradient-to-br from-blue-50 to-indigo-100">
						<CardHeader class="space-y-1 bg-white/70 backdrop-blur px-6 py-5 border-b border-white/60">
							<CardTitle class="text-slate-800 flex items-center">
								<ShoppingCart class="h-5 w-5 mr-2 text-blue-600" />
								Purchase Order Details
							</CardTitle>
							<CardDescription class="text-slate-600">ID: {purchaseOrder.ID}</CardDescription>
						</CardHeader>
						<CardContent class="p-6 space-y-4 text-sm">
							<div>
								<p class="font-medium text-slate-500">Supplier</p>
								<a href="/suppliers/{purchaseOrder.SupplierID}" class="text-slate-800 hover:underline">
									Supplier ID: {purchaseOrder.SupplierID}
								</a>
							</div>
							<div>
								<p class="font-medium text-slate-500">Status</p>
								<span
									class={cn(
										'rounded-full px-2.5 py-1 text-xs capitalize border shadow-sm',
										purchaseOrder.Status === 'DRAFT'
											? 'bg-gray-100 text-gray-700 border-gray-200'
											: purchaseOrder.Status === 'APPROVED'
											? 'bg-green-100 text-green-700 border-green-200'
											: purchaseOrder.Status === 'SENT'
											? 'bg-blue-100 text-blue-700 border-blue-200'
											: 'bg-red-100 text-red-700 border-red-200'
									)}
								>
									{purchaseOrder.Status}
								</span>
							</div>
							<div>
								<p class="font-medium text-slate-500">Order Date</p>
								<p class="text-slate-800">{new Date(purchaseOrder.OrderDate).toLocaleDateString()}</p>
							</div>
							<div>
								<p class="font-medium text-slate-500">Expected Delivery</p>
								<p class="text-slate-800">{purchaseOrder.ExpectedDeliveryDate ? new Date(purchaseOrder.ExpectedDeliveryDate).toLocaleDateString() : 'N/A'}</p>
							</div>
						</CardContent>
					</Card>
				</div>

				<!-- PO Items -->
				<div class="lg:col-span-2">
					<Card class="rounded-2xl shadow-lg hover:shadow-xl transition-all duration-300 hover:scale-[1.01] overflow-hidden border-0 bg-gradient-to-br from-gray-50 to-slate-100">
						<CardHeader class="space-y-1 bg-white/70 backdrop-blur px-6 py-5 border-b border-white/60">
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
											<TableCell colspan="4" class="text-center text-slate-500 py-4">
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
