<script lang="ts">
	import { onMount } from 'svelte';
	import GlassCard from '$lib/components/ui/GlassCard.svelte';
	import api from '$lib/api';
	import { formatDate, formatCurrency } from '$lib/utils';
	import { fade, fly, slide } from 'svelte/transition';
	import {
		ShoppingBag,
		Package,
		Truck,
		AlertCircle,
		CheckCircle2,
		XCircle,
		Search,
		TrendingUp,
		Factory,
		ArrowRight
	} from 'lucide-svelte';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { cn } from '$lib/utils';

	// Types
	interface Product {
		ID: number;
		Name: string;
		SKU: string;
		SellingPrice: number;
		PurchasePrice: number;
	}

	interface OrderItem {
		ID: number;
		ProductID: number;
		Product: Product;
		Quantity: number;
		UnitPrice: number;
		TotalPrice: number;
	}

	// Sales Order (Invoice)
	interface SalesOrder {
		ID: number;
		OrderNumber: string;
		OrderDate: string;
		TotalAmount: number;
		Status: string;
		PaymentMethod: string;
		OrderItems: OrderItem[];
		User?: {
			FirstName: string;
			LastName: string;
		};
	}

	// Purchase Order (Restock)
	interface PurchaseOrderItem {
		ID: number;
		ProductID: number;
		Product: Product;
		OrderedQuantity: number;
		ReceivedQuantity: number;
		UnitPrice: number;
	}

	interface PurchaseOrder {
		ID: number;
		SupplierID: number;
		Supplier: {
			Name: string;
		};
		Status: string;
		OrderDate: string;
		ExpectedDeliveryDate?: string;
		PurchaseOrderItems: PurchaseOrderItem[];
	}

	// Vendor Return
	interface PurchaseReturn {
		ID: number;
		Supplier: { Name: string };
		Status: string;
		Reason: string;
		RefundAmount: number;
		ReturnedAt: string;
		PurchaseReturnItems: {
			Product: { Name: string };
			Quantity: number;
			Reason: string;
		}[];
	}

	// Customer Return
	interface CustomerReturn {
		ID: number;
		OrderID: number;
		UserID: number;
		Status: string;
		Reason: string;
		RefundAmount: number;
		CreatedAt: string;
		ReturnItems: {
			Product: { Name: string };
			ProductID: number;
			Quantity: number;
			Reason: string;
		}[];
	}

	// State
	let activeTab: 'sales' | 'purchases' = $state('sales');
	let subTab: 'orders' | 'returns' = $state('orders'); // Sub-tab for purchases
	let salesOrders = $state<SalesOrder[]>([]);
	let purchaseOrders = $state<PurchaseOrder[]>([]);
	let purchaseReturns = $state<PurchaseReturn[]>([]);
	let customerReturns = $state<CustomerReturn[]>([]);
	let loading = $state(true);
	let error = $state<string | null>(null);
	let searchQuery = $state('');

	// Create Return Modal State
	let showReturnModal = $state(false);
	// Simplified create return state (mock data for now or would fetch suppliers/products)
	// For MVP, we will just show the list of returns. Creating requires complex "Batch selection" which is hard to UI without a deeper flow.
	// But I will add the UI to LIST returns first.

	import { notifications } from '$lib/stores/notifications';
	import { auth } from '$lib/stores/auth';
	import { goto } from '$app/navigation';
	import { toast } from 'svelte-sonner';

	// ... imports ...

	// Types
	// ... (Types remain same) ...

	// State
	// ... (State remains same) ...

	// Data Loading
	async function loadData() {
		if (!auth.hasPermission('pos.view') && !auth.hasPermission('inventory.view')) {
			toast.error('Access Denied', { description: 'You do not have permission to view orders.' });
			goto('/');
			return;
		}

		loading = true;
		error = null;
		try {
			const promises = [];

			// Load Sales only if allowed
			if (auth.hasPermission('pos.view')) {
				promises.push(api.get('/sales/history').catch(() => ({ data: { orders: [] } }))); // 0: sales
				promises.push(api.get('/returns').catch(() => ({ data: { returns: [] } }))); // 1: customer returns
			} else {
				promises.push(Promise.resolve({ data: { orders: [] } }));
				promises.push(Promise.resolve({ data: { returns: [] } }));
			}

			// Load Purchases only if allowed
			if (auth.hasPermission('inventory.view')) {
				promises.push(
					api.get('/replenishment/purchase-orders').catch(() => ({ data: { purchaseOrders: [] } }))
				); // 2: POs
				promises.push(api.get('/replenishment/returns').catch(() => ({ data: { returns: [] } }))); // 3: vendor returns
			} else {
				promises.push(Promise.resolve({ data: { purchaseOrders: [] } }));
				promises.push(Promise.resolve({ data: { returns: [] } }));
			}

			const [salesRes, customerReturnsRes, purchasesRes, vendorReturnsRes] =
				await Promise.all(promises);

			salesOrders = salesRes.data.orders || [];
			customerReturns = customerReturnsRes.data.returns || [];
			purchaseOrders = purchasesRes.data.purchaseOrders || [];
			purchaseReturns = vendorReturnsRes.data.returns || [];

			// Set initial tab based on permission preference
			if (!auth.hasPermission('pos.view') && auth.hasPermission('inventory.view')) {
				activeTab = 'purchases';
			}
		} catch (err) {
			console.error('Failed to load data', err);
			error = 'Failed to load order history.';
		} finally {
			loading = false;
		}
	}

	onMount(() => {
		loadData();
	});

	// Hybrid Sync Logic

	// 1. Push: Handle Real-time Update (Full Payload)
	function handleRealTimeUpdate(event: CustomEvent) {
		const payload = event.detail;
		if (!payload) return;

		// Handle Vendor Return Updates
		if (payload.type === 'VENDOR_RETURN' && payload.data) {
			const updatedReturn = payload.data as PurchaseReturn;

			// Update local state immediately
			const index = purchaseReturns.findIndex((r) => r.ID === updatedReturn.ID);
			if (index !== -1) {
				// Update existing
				purchaseReturns[index] = { ...purchaseReturns[index], ...updatedReturn };
			} else {
				// Add new (prepend)
				purchaseReturns = [updatedReturn, ...purchaseReturns];
			}

			// Optional: Show toast
			// toast.success(`Return #${updatedReturn.ID} updated`);
		}

		// Handle Customer Return Updates
		if (payload.type === 'CUSTOMER_RETURN' && payload.data) {
			const updatedReturn = payload.data as CustomerReturn;
			const index = customerReturns.findIndex((r) => r.ID === updatedReturn.ID);
			if (index !== -1) {
				customerReturns[index] = { ...customerReturns[index], ...updatedReturn };
			} else {
				customerReturns = [updatedReturn, ...customerReturns];
			}
		}
	}

	// 2. Pull: Handle Reconnection (Sync)
	let wasConnected = false;
	$effect(() => {
		const isConnected = $notifications.connected;
		// If we were disconnected and now connected, sync data
		if (!wasConnected && isConnected) {
			console.log('Connection restored, syncing data...');
			loadData();
		}
		wasConnected = isConnected;
	});

	// Filtering
	const filteredSales = $derived(
		salesOrders.filter(
			(o) =>
				o.OrderNumber.toLowerCase().includes(searchQuery.toLowerCase()) ||
				o.Status.toLowerCase().includes(searchQuery.toLowerCase())
		)
	);

	const filteredCustomerReturns = $derived(
		customerReturns.filter(
			(r) =>
				(r.ID + '').includes(searchQuery) ||
				r.Status.toLowerCase().includes(searchQuery.toLowerCase())
		)
	);

	const filteredPurchases = $derived(
		purchaseOrders.filter(
			(po) =>
				po.Supplier?.Name.toLowerCase().includes(searchQuery.toLowerCase()) ||
				po.Status.toLowerCase().includes(searchQuery.toLowerCase())
		)
	);

	const filteredReturns = $derived(
		purchaseReturns.filter(
			(r) =>
				r.Supplier?.Name.toLowerCase().includes(searchQuery.toLowerCase()) ||
				r.Status.toLowerCase().includes(searchQuery.toLowerCase())
		)
	);

	// Helpers
	function getStatusColor(status: string) {
		const s = status?.toUpperCase();
		if (['COMPLETED', 'RECEIVED', 'APPROVED'].includes(s))
			return 'text-emerald-500 bg-emerald-50 border-emerald-200';
		if (['PENDING', 'DRAFT', 'SENT', 'PARTIALLY_RECEIVED'].includes(s))
			return 'text-amber-500 bg-amber-50 border-amber-200';
		if (['CANCELLED', 'RETURNED', 'REJECTED'].includes(s))
			return 'text-red-500 bg-red-50 border-red-200';
		return 'text-slate-500 bg-slate-50 border-slate-200';
	}

	function getStatusIcon(status: string) {
		const s = status?.toUpperCase();
		if (['COMPLETED', 'RECEIVED', 'APPROVED'].includes(s)) return CheckCircle2;
		if (['PENDING', 'DRAFT', 'SENT'].includes(s)) return Truck;
		if (['CANCELLED', 'RETURNED', 'REJECTED'].includes(s)) return XCircle;
		return Package;
	}
</script>

<svelte:window on:return-updated={(e) => handleRealTimeUpdate(e)} />

<div class="relative min-h-screen overflow-hidden bg-slate-50/50 p-6 lg:p-10">
	<!-- Background -->
	<div
		class="absolute inset-0 -z-10 bg-[radial-gradient(ellipse_at_top,_var(--tw-gradient-stops))] from-blue-100/20 via-slate-50/20 to-white/20"
	></div>

	<div class="mx-auto max-w-7xl space-y-8">
		<!-- Header -->
		<div class="flex flex-col gap-6 md:flex-row md:items-center md:justify-between">
			<div class="space-y-1">
				<h1
					class="bg-gradient-to-r from-slate-900 via-blue-900 to-slate-900 bg-clip-text text-3xl font-bold tracking-tight text-transparent"
				>
					Orders Management
				</h1>
				<p class="text-slate-500">
					Manage customer sales invoices and supplier purchase orders & returns.
				</p>
			</div>

			<div class="flex flex-col gap-2">
				<!-- Main Tabs -->
				<div
					class="flex items-center gap-3 self-start rounded-xl border border-white/60 bg-white/40 p-1.5 shadow-sm backdrop-blur-md md:self-auto"
				>
					<Button
						variant="ghost"
						class={cn(
							'gap-2 rounded-lg font-semibold transition-all',
							activeTab === 'sales'
								? 'bg-white text-blue-600 shadow-sm ring-1 ring-black/5'
								: 'text-slate-500 hover:bg-white/50 hover:text-slate-700'
						)}
						onclick={() => (activeTab = 'sales')}
					>
						<TrendingUp class="h-4 w-4" /> Sales (Invoices)
					</Button>
					<Button
						variant="ghost"
						class={cn(
							'gap-2 rounded-lg font-semibold transition-all',
							activeTab === 'purchases'
								? 'bg-white text-purple-600 shadow-sm ring-1 ring-black/5'
								: 'text-slate-500 hover:bg-white/50 hover:text-slate-700'
						)}
						onclick={() => (activeTab = 'purchases')}
					>
						<Factory class="h-4 w-4" /> Purchases & Returns
					</Button>
				</div>

				<!-- Sub Tabs -->
				<div in:slide={{ axis: 'y', duration: 200 }} class="flex gap-2 self-start md:self-end">
					{#if activeTab === 'sales'}
						<button
							class={cn(
								'rounded-full px-3 py-1 text-sm font-medium transition-colors',
								subTab === 'orders'
									? 'bg-blue-100 text-blue-700'
									: 'text-slate-500 hover:text-slate-700'
							)}
							onclick={() => (subTab = 'orders')}
						>
							Sales Orders
						</button>
						<button
							class={cn(
								'rounded-full px-3 py-1 text-sm font-medium transition-colors',
								subTab === 'returns'
									? 'bg-blue-100 text-blue-700'
									: 'text-slate-500 hover:text-slate-700'
							)}
							onclick={() => (subTab = 'returns')}
						>
							Customer Returns
						</button>
					{:else}
						<button
							class={cn(
								'rounded-full px-3 py-1 text-sm font-medium transition-colors',
								subTab === 'orders'
									? 'bg-purple-100 text-purple-700'
									: 'text-slate-500 hover:text-slate-700'
							)}
							onclick={() => (subTab = 'orders')}
						>
							Purchase Orders
						</button>
						<button
							class={cn(
								'rounded-full px-3 py-1 text-sm font-medium transition-colors',
								subTab === 'returns'
									? 'bg-purple-100 text-purple-700'
									: 'text-slate-500 hover:text-slate-700'
							)}
							onclick={() => (subTab = 'returns')}
						>
							Vendor Returns
						</button>
					{/if}
				</div>
			</div>
		</div>

		<!-- Search & Actions -->
		<div
			class="flex items-center justify-between rounded-2xl border border-white/60 bg-white/60 p-4 shadow-sm backdrop-blur-sm"
		>
			<div class="relative max-w-md flex-1">
				<Search class="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-slate-400" />
				<Input
					bind:value={searchQuery}
					placeholder={activeTab === 'sales'
						? 'Search sales...'
						: subTab === 'orders'
							? 'Search POs...'
							: 'Search returns...'}
					class="border-slate-200 bg-white pl-9 shadow-none focus:border-blue-500 focus:ring-blue-500/20"
				/>
			</div>
			<!-- Add create buttons later if needed -->
		</div>

		{#if loading}
			<div class="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
				{#each Array(6) as _}
					<div class="h-48 animate-pulse rounded-2xl bg-white/50"></div>
				{/each}
			</div>
		{:else if error}
			<div
				class="flex h-64 flex-col items-center justify-center rounded-2xl bg-red-50/50 text-red-500"
			>
				<AlertCircle class="mb-4 h-8 w-8" />
				<p class="font-medium">{error}</p>
				<Button variant="link" onclick={loadData} class="mt-2 text-red-600 underline"
					>Try Again</Button
				>
			</div>
		{:else}
			{#key activeTab + subTab}
				<div
					in:fly={{ y: 20, duration: 400, delay: 0 }}
					out:fade={{ duration: 150 }}
					class="grid gap-6 md:grid-cols-2 lg:grid-cols-3"
				>
					{#if activeTab === 'sales'}
						{#each filteredSales as order (order.ID)}
							<GlassCard
								class="group relative flex flex-col overflow-hidden transition-all hover:shadow-lg"
							>
								<div class="flex items-start justify-between border-b border-slate-100 p-4">
									<div>
										<h3 class="font-mono text-sm font-bold text-slate-800">{order.OrderNumber}</h3>
										<p class="text-xs text-slate-500">{formatDate(order.OrderDate)}</p>
									</div>
									<div
										class={cn(
											'flex items-center gap-1.5 rounded-full border px-2 py-0.5 text-[10px] font-bold uppercase tracking-wide',
											getStatusColor(order.Status)
										)}
									>
										<svelte:component this={getStatusIcon(order.Status)} size={12} />
										{order.Status}
									</div>
								</div>

								<div class="flex-1 space-y-3 p-4">
									<div class="flex items-center justify-between text-sm">
										<span class="text-slate-500">Customer</span>
										<span class="font-medium text-slate-700">
											{order.User
												? `${order.User.FirstName} ${order.User.LastName}`
												: 'Guest / System'}
										</span>
									</div>
									<div class="flex items-center justify-between text-sm">
										<span class="text-slate-500">Items</span>
										<span class="font-medium text-slate-700">{order.OrderItems.length}</span>
									</div>
									<div class="space-y-1 rounded-lg bg-slate-50 p-2 text-xs text-slate-600">
										{#each order.OrderItems.slice(0, 2) as item}
											<div class="flex justify-between">
												<span class="truncate pr-2">{item.Product?.Name || 'Product'}</span>
												<span>x{item.Quantity}</span>
											</div>
										{/each}
										{#if order.OrderItems.length > 2}
											<div class="text-[10px] italic text-slate-400">
												+{order.OrderItems.length - 2} more
											</div>
										{/if}
									</div>
								</div>

								<div class="mt-auto flex items-center justify-between bg-slate-50/50 p-4">
									<p class="text-lg font-bold text-slate-800">
										{formatCurrency(order.TotalAmount)}
									</p>
									<Button
										href="/orders/{order.OrderNumber}"
										variant="ghost"
										size="sm"
										class="text-blue-600 hover:bg-blue-50 hover:text-blue-700"
									>
										View Details <ArrowRight class="ml-1 h-3 w-3" />
									</Button>
								</div>
							</GlassCard>
						{:else}
							<div class="col-span-full py-12 text-center text-slate-500">
								No sales orders found.
							</div>
						{/each}
					{:else if activeTab === 'sales' && subTab === 'returns'}
						<!-- Customer Returns List -->
						{#each filteredCustomerReturns as ret (ret.ID)}
							<GlassCard
								class="group relative flex flex-col overflow-hidden border-orange-100/50 transition-all hover:border-orange-200 hover:shadow-lg"
							>
								<div class="flex items-start justify-between border-b border-slate-100 p-4">
									<div>
										<h3 class="text-sm font-bold text-slate-800">Return #{ret.ID}</h3>
										<p class="text-xs text-slate-500">{formatDate(ret.CreatedAt)}</p>
									</div>
									<div
										class={cn(
											'flex items-center gap-1.5 rounded-full border px-2 py-0.5 text-[10px] font-bold uppercase tracking-wide',
											getStatusColor(ret.Status)
										)}
									>
										{ret.Status}
									</div>
								</div>

								<div class="flex-1 space-y-3 p-4">
									<div class="flex items-center justify-between text-sm">
										<span class="text-slate-500">Refund</span>
										<span class="font-bold text-slate-800">{formatCurrency(ret.RefundAmount)}</span>
									</div>
									<div class="flex items-center justify-between text-sm">
										<span class="text-slate-500">Reason</span>
										<span class="font-medium text-slate-700">{ret.Reason}</span>
									</div>
									<div class="space-y-1 rounded-lg bg-orange-50/50 p-2 text-xs text-slate-600">
										<p class="mb-1 font-semibold text-orange-800">Items Returned</p>
										{#if ret.ReturnItems && ret.ReturnItems.length > 0}
											{#each ret.ReturnItems.slice(0, 2) as item}
												<div class="flex justify-between">
													<span class="truncate pr-2"
														>{item.Product?.Name || `Item #${item.ProductID}`}</span
													>
													<span class="font-medium text-orange-600">x{item.Quantity}</span>
												</div>
											{/each}
										{:else}
											<p class="italic text-slate-400">No items detail</p>
										{/if}
									</div>
								</div>

								<div class="mt-auto flex items-center justify-between bg-slate-50/50 p-4">
									<Button
										variant="ghost"
										size="sm"
										class="w-full text-orange-600 hover:bg-orange-50 hover:text-orange-700"
									>
										View Details <ArrowRight class="ml-1 h-3 w-3" />
									</Button>
								</div>
							</GlassCard>
						{:else}
							<div class="col-span-full py-12 text-center text-slate-500">
								No customer returns found.
							</div>
						{/each}
					{:else if subTab === 'orders'}
						{#each filteredPurchases as po (po.ID)}
							<GlassCard
								class="group relative flex flex-col overflow-hidden border-purple-100/50 transition-all hover:border-purple-200 hover:shadow-lg"
							>
								<div class="flex items-start justify-between border-b border-slate-100 p-4">
									<div>
										<h3 class="text-sm font-bold text-slate-800">PO #{po.ID}</h3>
										<p class="text-xs text-slate-500">{formatDate(po.OrderDate)}</p>
									</div>
									<div
										class={cn(
											'flex items-center gap-1.5 rounded-full border px-2 py-0.5 text-[10px] font-bold uppercase tracking-wide',
											getStatusColor(po.Status)
										)}
									>
										<svelte:component this={getStatusIcon(po.Status)} size={12} />
										{po.Status}
									</div>
								</div>

								<div class="flex-1 space-y-3 p-4">
									<div class="flex items-center justify-between text-sm">
										<span class="text-slate-500">Supplier</span>
										<span class="font-medium text-slate-700">{po.Supplier?.Name || 'Unknown'}</span>
									</div>
									<div class="space-y-1 rounded-lg bg-slate-50 p-2 text-xs text-slate-600">
										{#each po.PurchaseOrderItems.slice(0, 2) as item}
											<div class="flex justify-between">
												<span class="truncate pr-2">{item.Product?.Name || 'Product'}</span>
												<span class="font-medium text-purple-600">Qty: {item.OrderedQuantity}</span>
											</div>
										{/each}
										{#if po.PurchaseOrderItems.length > 2}
											<div class="text-[10px] italic text-slate-400">
												+{po.PurchaseOrderItems.length - 2} more
											</div>
										{/if}
									</div>
								</div>

								<div class="mt-auto flex items-center justify-between bg-slate-50/50 p-4">
									<div>
										{#if po.PurchaseOrderItems.length > 0}
											<p class="text-xs font-semibold text-slate-500">
												Total Items: {po.PurchaseOrderItems.reduce(
													(acc, item) => acc + item.OrderedQuantity,
													0
												)}
											</p>
										{/if}
									</div>
									<Button
										href="/purchase-orders/{po.ID}"
										variant="ghost"
										size="sm"
										class="text-purple-600 hover:bg-purple-50 hover:text-purple-700"
									>
										Manage PO <ArrowRight class="ml-1 h-3 w-3" />
									</Button>
								</div>
							</GlassCard>
						{:else}
							<div class="col-span-full py-12 text-center text-slate-500">
								No purchase orders found.
							</div>
						{/each}
					{:else}
						<!-- Vendor Returns List -->
						{#each filteredReturns as ret (ret.ID)}
							<GlassCard
								class="group relative flex flex-col overflow-hidden border-red-100/50 transition-all hover:border-red-200 hover:shadow-lg"
							>
								<div class="flex items-start justify-between border-b border-slate-100 p-4">
									<div>
										<h3 class="text-sm font-bold text-slate-800">Return #{ret.ID}</h3>
										<p class="text-xs text-slate-500">{formatDate(ret.ReturnedAt)}</p>
									</div>
									<div
										class={cn(
											'flex items-center gap-1.5 rounded-full border px-2 py-0.5 text-[10px] font-bold uppercase tracking-wide',
											getStatusColor(ret.Status)
										)}
									>
										{ret.Status}
									</div>
								</div>

								<div class="flex-1 space-y-3 p-4">
									<div class="flex items-center justify-between text-sm">
										<span class="text-slate-500">Supplier</span>
										<span class="font-medium text-slate-700">{ret.Supplier?.Name || 'Unknown'}</span
										>
									</div>
									<div class="flex items-center justify-between text-sm">
										<span class="text-slate-500">Reason</span>
										<span class="font-medium text-slate-700">{ret.Reason}</span>
									</div>

									<div class="space-y-1 rounded-lg bg-slate-50 p-2 text-xs text-slate-600">
										{#each ret.PurchaseReturnItems.slice(0, 2) as item}
											<div class="flex justify-between">
												<span class="truncate pr-2">{item.Product?.Name || 'Product'}</span>
												<span class="font-medium text-red-600">Qty: -{item.Quantity}</span>
											</div>
										{/each}
									</div>
								</div>

								<div class="mt-auto flex items-center justify-between bg-slate-50/50 p-4">
									<p class="text-lg font-bold text-slate-800">{formatCurrency(ret.RefundAmount)}</p>
									<span class="text-xs font-semibold text-slate-400">Refund Value</span>
								</div>
							</GlassCard>
						{:else}
							<div class="col-span-full py-12 text-center text-slate-500">
								No vendor returns found.
							</div>
						{/each}
					{/if}
				</div>
			{/key}
		{/if}
	</div>
</div>
