<script lang="ts">
	import { onMount } from 'svelte';
	import GlassCard from '$lib/components/ui/GlassCard.svelte';
	import GlassTabs from '$lib/components/ui/glass-tabs.svelte';
	import api from '$lib/api';
	import { formatCurrency, orderNumber, returnNumber, poNumber } from '$lib/utils';
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
		ArrowRight,
		ShoppingCart,
		Undo2
	} from 'lucide-svelte';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { cn } from '$lib/utils';
	import TimeAgo from '$lib/components/ui/time-ago.svelte';
	import EmptyState from '$lib/components/ui/empty-state.svelte';
	import ColumnPicker from '$lib/components/ui/column-picker.svelte';
	import { loadColumnPrefs, saveColumnPrefs, type ColumnDef } from '$lib/column-prefs';

	const SALES_COLUMNS: ColumnDef[] = [
		{ key: 'customer', label: 'Customer' },
		{ key: 'date', label: 'Date' },
		{ key: 'items', label: 'Items' },
		{ key: 'total', label: 'Total' },
		{ key: 'status', label: 'Status' },
		{ key: 'payment', label: 'Payment' }
	];

	let salesVisible = $state(loadColumnPrefs('orders-sales', SALES_COLUMNS));

	function toggleSalesColumn(key: string) {
		if (salesVisible.has(key)) {
			salesVisible.delete(key);
		} else {
			salesVisible.add(key);
		}
		salesVisible = new Set(salesVisible);
		saveColumnPrefs('orders-sales', salesVisible);
	}

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
		Customer?: {
			FirstName: string;
			LastName: string;
		};
		HasPendingReturn?: boolean;
		AdjustedTotal?: number;
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

	import { t } from '$lib/i18n';
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
			toast.error($t('orders.errors.access_denied'), {
				description: $t('orders.errors.access_denied_desc')
			});
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

			salesOrders = salesRes.data.orders || salesRes.data.data || salesRes.data.items || [];
			customerReturns = customerReturnsRes.data.returns || [];
			purchaseOrders =
				purchasesRes.data.purchaseOrders || purchasesRes.data.data || purchasesRes.data.items || [];
			purchaseReturns = vendorReturnsRes.data.returns || [];

			// Set initial tab based on permission preference
			if (!auth.hasPermission('pos.view') && auth.hasPermission('inventory.view')) {
				activeTab = 'purchases';
			}
		} catch (err) {
			console.error('Failed to load data', err);
			error = $t('orders.errors.load_fail');
		} finally {
			loading = false;
		}
	}

	onMount(() => {
		loadData();

		const handleReturnUpdated = (e: Event) => handleRealTimeUpdate(e as CustomEvent);
		window.addEventListener('return-updated', handleReturnUpdated);
		return () => window.removeEventListener('return-updated', handleReturnUpdated);
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
		salesOrders.filter((o) => {
			const q = searchQuery.toLowerCase();
			const customer = o.Customer
				? `${o.Customer.FirstName} ${o.Customer.LastName}`.toLowerCase()
				: '';
			return (
				orderNumber(o).toLowerCase().includes(q) ||
				customer.includes(q) ||
				o.Status.toLowerCase().includes(q)
			);
		})
	);

	const filteredCustomerReturns = $derived(
		customerReturns.filter((r) => {
			const q = searchQuery.toLowerCase();
			return returnNumber(r.ID).toLowerCase().includes(q) || r.Status.toLowerCase().includes(q);
		})
	);

	const filteredPurchases = $derived(
		purchaseOrders.filter((po) => {
			const q = searchQuery.toLowerCase();
			return (
				poNumber(po.ID).toLowerCase().includes(q) ||
				po.Supplier?.Name.toLowerCase().includes(q) ||
				po.Status.toLowerCase().includes(q)
			);
		})
	);

	const filteredReturns = $derived(
		purchaseReturns.filter((r) => {
			const q = searchQuery.toLowerCase();
			return (
				returnNumber(r.ID).toLowerCase().includes(q) ||
				r.Supplier?.Name.toLowerCase().includes(q) ||
				r.Status.toLowerCase().includes(q)
			);
		})
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
					{$t('orders.title')}
				</h1>
				<p class="text-slate-500">
					{$t('orders.subtitle')}
				</p>
			</div>

			<div class="flex flex-col gap-2">
				<!-- Main Tabs -->
				<GlassTabs
					bind:value={activeTab}
					tabs={[
						{ value: 'sales', label: $t('orders.tabs.sales') },
						{ value: 'purchases', label: $t('orders.tabs.purchases') }
					]}
					class="self-start md:self-auto"
				/>

				<!-- Sub Tabs -->
				<!-- Sub Tabs -->
				<div in:slide={{ axis: 'y', duration: 200 }} class="flex gap-2 self-start md:self-end">
					{#if activeTab === 'sales'}
						<GlassTabs
							bind:value={subTab}
							tabs={[
								{ value: 'orders', label: $t('orders.subtabs.sales_orders') },
								{ value: 'returns', label: $t('orders.subtabs.customer_returns') }
							]}
						/>
					{:else}
						<GlassTabs
							bind:value={subTab}
							tabs={[
								{ value: 'orders', label: $t('orders.subtabs.purchase_orders') },
								{ value: 'returns', label: $t('orders.subtabs.vendor_returns') }
							]}
						/>
					{/if}
				</div>
			</div>
		</div>

		<!-- Search & Actions -->
		<div
			class="liquid-panel flex items-center justify-between rounded-2xl bg-gradient-to-br from-white/40 via-white/20 to-white/5 p-4 backdrop-blur-sm"
		>
			<div class="relative max-w-md flex-1">
				<Search class="absolute top-1/2 left-3 h-4 w-4 -translate-y-1/2 text-slate-400" />
				<Input
					bind:value={searchQuery}
					placeholder={activeTab === 'sales'
						? $t('orders.search.sales')
						: subTab === 'orders'
							? $t('orders.search.pos')
							: $t('orders.search.returns')}
					class="border-slate-200 bg-white pl-9 shadow-none focus:border-blue-500 focus:ring-blue-500/20"
				/>
			</div>
			{#if activeTab === 'sales' && subTab === 'orders'}
				<ColumnPicker columns={SALES_COLUMNS} visible={salesVisible} onToggle={toggleSalesColumn} />
			{/if}
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
					>{$t('orders.actions.try_again')}</Button
				>
			</div>
		{:else}
			{#key activeTab + subTab}
				<div
					in:fly={{ y: 20, duration: 400, delay: 0 }}
					out:fade={{ duration: 150 }}
					class="grid gap-6 md:grid-cols-2 lg:grid-cols-3"
				>
					{#if activeTab === 'sales' && subTab === 'orders'}
						{#each filteredSales as order (order.ID)}
							<div
								class="liquid-hoverable group relative flex flex-col overflow-hidden rounded-2xl bg-gradient-to-br from-white/50 via-white/30 to-white/10 transition-all duration-200 hover:bg-white/60"
							>
								{#if order.HasPendingReturn}
									<div
										class="absolute top-0 right-0 rounded-bl-lg bg-amber-100 px-2 py-1 text-[10px] font-bold text-amber-600 shadow-sm"
									>
										{$t('orders.status.return_pending')}
									</div>
								{/if}
							<div class="flex items-start justify-between border-b border-slate-100 p-4">
								<div>
									<h3 class="font-mono text-sm font-bold text-slate-800">{orderNumber(order)}</h3>
									{#if salesVisible.has('date')}
										<p class="text-xs text-slate-500"><TimeAgo date={order.OrderDate} /></p>
									{/if}
								</div>
								{#if salesVisible.has('status')}
									<div
										class={cn(
											'flex items-center gap-1.5 rounded-full border px-2 py-0.5 text-[10px] font-bold tracking-wide uppercase',
											getStatusColor(order.Status)
										)}
									>
										<svelte:component this={getStatusIcon(order.Status)} size={12} />
										{order.Status}
									</div>
								{/if}
							</div>

							<div class="flex-1 space-y-3 p-4">
								{#if salesVisible.has('customer')}
									<div class="flex items-center justify-between text-sm">
										<span class="text-slate-500">{$t('orders.labels.customer')}</span>
										<span class="font-medium text-slate-700">
											{order.Customer
												? `${order.Customer.FirstName} ${order.Customer.LastName}`
												: $t('orders.labels.guest')}
										</span>
									</div>
								{/if}
								{#if salesVisible.has('payment')}
									<div class="flex items-center justify-between text-sm">
										<span class="text-slate-500">Payment</span>
										<span class="font-medium text-slate-700">{order.PaymentMethod || '—'}</span>
									</div>
								{/if}
								{#if salesVisible.has('items')}
									<div class="flex items-center justify-between text-sm">
										<span class="text-slate-500">{$t('orders.labels.items')}</span>
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
											<div class="text-[10px] text-slate-400 italic">
												+{order.OrderItems.length - 2}
												{$t('orders.labels.more')}
											</div>
										{/if}
									</div>
								{/if}
							</div>

							<div class="mt-auto flex items-center justify-between bg-slate-50/50 p-4">
								<div>
									{#if salesVisible.has('total')}
										{#if order.AdjustedTotal !== undefined && order.AdjustedTotal < order.TotalAmount}
											<p class="text-lg font-bold text-emerald-600">
												{formatCurrency(order.AdjustedTotal)}
											</p>
											<p class="text-xs text-slate-400 line-through">
												{formatCurrency(order.TotalAmount)}
											</p>
										{:else}
											<p class="text-lg font-bold text-slate-800">
												{formatCurrency(order.TotalAmount)}
											</p>
										{/if}
									{/if}
								</div>
								<Button
									href="/orders/{order.OrderNumber}"
									variant="ghost"
									size="sm"
									class="text-blue-600 hover:bg-blue-50 hover:text-blue-700"
								>
									{$t('orders.actions.view_details')}
									<ArrowRight class="ml-1 h-3 w-3" />
								</Button>
							</div>
							</div>
						{:else}
							<div class="col-span-full">
								<EmptyState
									icon={ShoppingCart}
									title="No orders yet"
									description="Sales you complete at the POS will show up here."
									actionLabel="Go to POS"
									onAction={() => goto('/pos')}
								/>
							</div>
						{/each}
					{:else if activeTab === 'sales' && subTab === 'returns'}
						<!-- Customer Returns List -->
						{#each filteredCustomerReturns as ret (ret.ID)}
							<div
								class="liquid-hoverable group relative flex flex-col overflow-hidden rounded-2xl bg-gradient-to-br from-white/50 via-white/30 to-white/10 transition-all duration-200 hover:bg-white/60"
							>
								<div class="flex items-start justify-between border-b border-slate-100 p-4">
									<div>
										<h3 class="font-mono text-sm font-bold text-slate-800">
											{returnNumber(ret.ID)}
										</h3>
										<p class="text-xs text-slate-500"><TimeAgo date={ret.CreatedAt} /></p>
									</div>
									<div
										class={cn(
											'flex items-center gap-1.5 rounded-full border px-2 py-0.5 text-[10px] font-bold tracking-wide uppercase',
											getStatusColor(ret.Status)
										)}
									>
										{ret.Status}
									</div>
								</div>

								<div class="flex-1 space-y-3 p-4">
									<div class="flex items-center justify-between text-sm">
										<span class="text-slate-500">{$t('orders.labels.order')}</span>
										<span class="font-mono font-medium text-slate-700">#{ret.OrderID}</span>
									</div>
									<div class="flex items-center justify-between text-sm">
										<span class="text-slate-500">{$t('orders.labels.refund')}</span>
										<span class="font-bold text-slate-800">{formatCurrency(ret.RefundAmount)}</span>
									</div>
									<div class="flex items-center justify-between text-sm">
										<span class="text-slate-500">{$t('orders.labels.reason')}</span>
										<span class="font-medium text-slate-700">{ret.Reason}</span>
									</div>
									<div class="space-y-1 rounded-lg bg-orange-50/50 p-2 text-xs text-slate-600">
										<p class="mb-1 font-semibold text-orange-800">
											{$t('orders.status.items_returned')}
										</p>
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
											<p class="text-slate-400 italic">{$t('orders.status.no_items')}</p>
										{/if}
									</div>
								</div>

								<div class="mt-auto flex items-center justify-between bg-slate-50/50 p-4">
									<Button
										href={`/admin/returns/${ret.ID}`}
										variant="ghost"
										size="sm"
										class="w-full text-orange-600 hover:bg-orange-50 hover:text-orange-700"
									>
										{$t('orders.actions.view_details')}
										<ArrowRight class="ml-1 h-3 w-3" />
									</Button>
								</div>
							</div>
						{:else}
							<div class="col-span-full">
								<EmptyState
									icon={Undo2}
									title="No customer returns"
									description="Return requests from customers will appear here."
								/>
							</div>
						{/each}
					{:else if subTab === 'orders'}
						{#each filteredPurchases as po (po.ID)}
							<div
								class="liquid-hoverable group relative flex flex-col overflow-hidden rounded-2xl bg-gradient-to-br from-white/50 via-white/30 to-white/10 transition-all duration-200 hover:bg-white/60"
							>
								<div class="flex items-start justify-between border-b border-slate-100 p-4">
									<div>
										<h3 class="font-mono text-sm font-bold text-slate-800">{poNumber(po.ID)}</h3>
										<p class="text-xs text-slate-500"><TimeAgo date={po.OrderDate} /></p>
									</div>
									<div
										class={cn(
											'flex items-center gap-1.5 rounded-full border px-2 py-0.5 text-[10px] font-bold tracking-wide uppercase',
											getStatusColor(po.Status)
										)}
									>
										<svelte:component this={getStatusIcon(po.Status)} size={12} />
										{po.Status}
									</div>
								</div>

								<div class="flex-1 space-y-3 p-4">
									<div class="flex items-center justify-between text-sm">
										<span class="text-slate-500">{$t('orders.labels.supplier')}</span>
										<span class="font-medium text-slate-700"
											>{po.Supplier?.Name || $t('orders.labels.unknown')}</span
										>
									</div>
									<div class="space-y-1 rounded-lg bg-slate-50 p-2 text-xs text-slate-600">
										{#each po.PurchaseOrderItems.slice(0, 2) as item}
											<div class="flex justify-between">
												<span class="truncate pr-2">{item.Product?.Name || 'Product'}</span>
												<span class="font-medium text-purple-600"
													>{$t('orders.status.qty')}: {item.OrderedQuantity}</span
												>
											</div>
										{/each}
										{#if po.PurchaseOrderItems.length > 2}
											<div class="text-[10px] text-slate-400 italic">
												+{po.PurchaseOrderItems.length - 2}
												{$t('orders.labels.more')}
											</div>
										{/if}
									</div>
								</div>

								<div class="mt-auto flex items-center justify-between bg-slate-50/50 p-4">
									<div>
										{#if po.PurchaseOrderItems.length > 0}
											<p class="text-xs font-semibold text-slate-500">
												{$t('orders.status.total_items')}: {po.PurchaseOrderItems.reduce(
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
										{$t('orders.actions.manage_po')}
										<ArrowRight class="ml-1 h-3 w-3" />
									</Button>
								</div>
							</div>
						{:else}
							<div class="col-span-full">
								<EmptyState
									icon={Factory}
									title="No purchase orders"
									description="Purchase orders you create for suppliers will show up here."
								/>
							</div>
						{/each}
					{:else}
						<!-- Vendor Returns List -->
						{#each filteredReturns as ret (ret.ID)}
							<div
								class="liquid-hoverable group relative flex flex-col overflow-hidden rounded-2xl bg-gradient-to-br from-white/50 via-white/30 to-white/10 transition-all duration-200 hover:bg-white/60"
							>
								<div class="flex items-start justify-between border-b border-slate-100 p-4">
									<div>
										<h3 class="font-mono text-sm font-bold text-slate-800">
											{returnNumber(ret.ID)}
										</h3>
										<p class="text-xs text-slate-500"><TimeAgo date={ret.ReturnedAt} /></p>
									</div>
									<div
										class={cn(
											'flex items-center gap-1.5 rounded-full border px-2 py-0.5 text-[10px] font-bold tracking-wide uppercase',
											getStatusColor(ret.Status)
										)}
									>
										{ret.Status}
									</div>
								</div>

								<div class="flex-1 space-y-3 p-4">
									<div class="flex items-center justify-between text-sm">
										<span class="text-slate-500">{$t('orders.labels.supplier')}</span>
										<span class="font-medium text-slate-700"
											>{ret.Supplier?.Name || $t('orders.labels.unknown')}</span
										>
									</div>
									<div class="flex items-center justify-between text-sm">
										<span class="text-slate-500">{$t('orders.labels.reason')}</span>
										<span class="font-medium text-slate-700">{ret.Reason}</span>
									</div>

									<div class="space-y-1 rounded-lg bg-slate-50 p-2 text-xs text-slate-600">
										{#each ret.PurchaseReturnItems.slice(0, 2) as item}
											<div class="flex justify-between">
												<span class="truncate pr-2">{item.Product?.Name || 'Product'}</span>
												<span class="font-medium text-red-600"
													>{$t('orders.status.qty')}: -{item.Quantity}</span
												>
											</div>
										{/each}
									</div>
								</div>

								<div class="mt-auto flex items-center justify-between bg-slate-50/50 p-4">
									<p class="text-lg font-bold text-slate-800">{formatCurrency(ret.RefundAmount)}</p>
									<span class="text-xs font-semibold text-slate-400"
										>{$t('orders.status.refund_value')}</span
									>
								</div>
							</div>
						{:else}
							<div class="col-span-full">
								<EmptyState
									icon={Undo2}
									title="No vendor returns"
									description="Returns you send back to suppliers will appear here."
								/>
							</div>
						{/each}
					{/if}
				</div>
			{/key}
		{/if}
	</div>
</div>
