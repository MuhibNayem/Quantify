<script lang="ts">
	import { onMount } from 'svelte';
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

	const inputClass =
		'liquid-input h-10 w-full rounded-md bg-card bg-none px-3 text-sm font-normal text-foreground shadow-none ring-0 backdrop-blur-none backdrop-saturate-100 from-card to-card placeholder:text-muted-foreground focus:scale-100 focus:border-primary focus:ring-0 md:text-sm';
	const tabsClass =
		'rounded-lg border-border bg-card shadow-sm backdrop-blur-none';

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
		if (['COMPLETED', 'RECEIVED', 'APPROVED'].includes(s)) return 'badge-success';
		if (['PENDING', 'DRAFT', 'SENT', 'PARTIALLY_RECEIVED'].includes(s)) return 'badge-warning';
		if (['CANCELLED', 'RETURNED', 'REJECTED'].includes(s)) return 'badge-danger';
		return 'badge-neutral';
	}

	function getStatusIcon(status: string) {
		const s = status?.toUpperCase();
		if (['COMPLETED', 'RECEIVED', 'APPROVED'].includes(s)) return CheckCircle2;
		if (['PENDING', 'DRAFT', 'SENT'].includes(s)) return Truck;
		if (['CANCELLED', 'RETURNED', 'REJECTED'].includes(s)) return XCircle;
		return Package;
	}
</script>

<div class="relative min-h-screen overflow-hidden p-6 lg:p-10">
	<div class="mx-auto max-w-7xl space-y-8">
		<!-- Header -->
		<div class="flex flex-col gap-6 md:flex-row md:items-center md:justify-between">
			<div class="space-y-1" data-animate="fade-up">
				<p class="overline">{$t('orders.tabs.sales')} · {$t('orders.tabs.purchases')}</p>
				<h1
					class="font-display text-foreground text-3xl font-bold tracking-tight"
				>
					{$t('orders.title')}
				</h1>
				<p class="text-muted-foreground">
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
					class="{tabsClass} self-start md:self-auto"
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
							class={tabsClass}
						/>
					{:else}
						<GlassTabs
							bind:value={subTab}
							tabs={[
								{ value: 'orders', label: $t('orders.subtabs.purchase_orders') },
								{ value: 'returns', label: $t('orders.subtabs.vendor_returns') }
							]}
							class={tabsClass}
						/>
					{/if}
				</div>
			</div>
		</div>

		<!-- Search & Actions -->
		<div
			class="liquid-panel flex items-center justify-between rounded-lg p-4"
		>
			<div class="relative max-w-md flex-1">
				<Search class="text-muted-foreground absolute top-1/2 left-3 h-4 w-4 -translate-y-1/2" />
				<Input
					bind:value={searchQuery}
					placeholder={activeTab === 'sales'
						? $t('orders.search.sales')
						: subTab === 'orders'
							? $t('orders.search.pos')
							: $t('orders.search.returns')}
					class="{inputClass} pl-9"
				/>
			</div>
			{#if activeTab === 'sales' && subTab === 'orders'}
				<ColumnPicker columns={SALES_COLUMNS} visible={salesVisible} onToggle={toggleSalesColumn} />
			{/if}
		</div>

		{#if loading}
			<div class="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
				{#each Array(6) as _}
					<div class="bg-muted h-48 animate-pulse rounded-lg"></div>
				{/each}
			</div>
		{:else if error}
			<div
				class="text-danger flex h-64 flex-col items-center justify-center rounded-lg border border-danger/25 bg-danger/10"
			>
				<AlertCircle class="mb-4 h-8 w-8" />
				<p class="font-medium">{error}</p>
				<Button variant="link" onclick={loadData} class="text-danger mt-2 underline"
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
								class="liquid-hoverable group relative flex flex-col overflow-hidden rounded-lg"
							>
								{#if order.HasPendingReturn}
									<div
										class="bg-warning/15 text-warning absolute top-0 right-0 rounded-bl-md px-2 py-1 text-[10px] font-bold shadow-sm"
									>
										{$t('orders.status.return_pending')}
									</div>
								{/if}
							<div class="border-border flex items-start justify-between border-b p-4">
								<div>
									<h3 class="num-mono text-foreground text-sm font-bold">{orderNumber(order)}</h3>
									{#if salesVisible.has('date')}
										<p class="text-muted-foreground text-xs"><TimeAgo date={order.OrderDate} /></p>
									{/if}
								</div>
								{#if salesVisible.has('status')}
									<div class={cn('uppercase', getStatusColor(order.Status))}>
										<svelte:component this={getStatusIcon(order.Status)} size={12} />
										{order.Status}
									</div>
								{/if}
							</div>

							<div class="flex-1 space-y-3 p-4">
								{#if salesVisible.has('customer')}
									<div class="flex items-center justify-between text-sm">
										<span class="text-muted-foreground">{$t('orders.labels.customer')}</span>
										<span class="text-foreground font-medium">
											{order.Customer
												? `${order.Customer.FirstName} ${order.Customer.LastName}`
												: $t('orders.labels.guest')}
										</span>
									</div>
								{/if}
								{#if salesVisible.has('payment')}
									<div class="flex items-center justify-between text-sm">
										<span class="text-muted-foreground">Payment</span>
										<span class="text-foreground font-medium">{order.PaymentMethod || '—'}</span>
									</div>
								{/if}
								{#if salesVisible.has('items')}
									<div class="flex items-center justify-between text-sm">
										<span class="text-muted-foreground">{$t('orders.labels.items')}</span>
										<span class="num text-foreground font-medium">{order.OrderItems.length}</span>
									</div>
									<div class="bg-muted text-muted-foreground space-y-1 rounded-md p-2 text-xs">
										{#each order.OrderItems.slice(0, 2) as item}
											<div class="flex justify-between">
												<span class="truncate pr-2">{item.Product?.Name || 'Product'}</span>
												<span class="num-mono">x{item.Quantity}</span>
											</div>
										{/each}
										{#if order.OrderItems.length > 2}
											<div class="text-muted-foreground text-[10px] italic">
												+{order.OrderItems.length - 2}
												{$t('orders.labels.more')}
											</div>
										{/if}
									</div>
								{/if}
							</div>

							<div class="bg-muted/50 border-border mt-auto flex items-center justify-between border-t p-4">
								<div>
									{#if salesVisible.has('total')}
										{#if order.AdjustedTotal !== undefined && order.AdjustedTotal < order.TotalAmount}
											<p class="num-mono text-success text-lg font-bold">
												{formatCurrency(order.AdjustedTotal)}
											</p>
											<p class="text-muted-foreground text-xs line-through">
												{formatCurrency(order.TotalAmount)}
											</p>
										{:else}
											<p class="num-mono text-foreground text-lg font-bold">
												{formatCurrency(order.TotalAmount)}
											</p>
										{/if}
									{/if}
								</div>
								<Button
									href="/orders/{order.OrderNumber}"
									variant="ghost"
									size="sm"
									class="text-primary hover:bg-accent hover:text-accent-foreground"
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
								class="liquid-hoverable group relative flex flex-col overflow-hidden rounded-lg"
							>
								<div class="border-border flex items-start justify-between border-b p-4">
									<div>
										<h3 class="num-mono text-foreground text-sm font-bold">
											{returnNumber(ret.ID)}
										</h3>
										<p class="text-muted-foreground text-xs"><TimeAgo date={ret.CreatedAt} /></p>
									</div>
									<div class={cn('uppercase', getStatusColor(ret.Status))}>
										{ret.Status}
									</div>
								</div>

								<div class="flex-1 space-y-3 p-4">
									<div class="flex items-center justify-between text-sm">
										<span class="text-muted-foreground">{$t('orders.labels.order')}</span>
										<span class="num-mono text-foreground font-medium">#{ret.OrderID}</span>
									</div>
									<div class="flex items-center justify-between text-sm">
										<span class="text-muted-foreground">{$t('orders.labels.refund')}</span>
										<span class="num-mono text-foreground font-bold">{formatCurrency(ret.RefundAmount)}</span>
									</div>
									<div class="flex items-center justify-between text-sm">
										<span class="text-muted-foreground">{$t('orders.labels.reason')}</span>
										<span class="text-foreground font-medium">{ret.Reason}</span>
									</div>
									<div class="space-y-1 rounded-md border border-warning/25 bg-warning/10 p-2 text-xs">
										<p class="text-warning mb-1 font-semibold">
											{$t('orders.status.items_returned')}
										</p>
										{#if ret.ReturnItems && ret.ReturnItems.length > 0}
											{#each ret.ReturnItems.slice(0, 2) as item}
												<div class="text-muted-foreground flex justify-between">
													<span class="truncate pr-2"
														>{item.Product?.Name || `Item #${item.ProductID}`}</span
													>
													<span class="num-mono text-warning font-medium">x{item.Quantity}</span>
												</div>
											{/each}
										{:else}
											<p class="text-muted-foreground italic">{$t('orders.status.no_items')}</p>
										{/if}
									</div>
								</div>

								<div class="bg-muted/50 border-border mt-auto flex items-center justify-between border-t p-4">
									<Button
										href={`/admin/returns/${ret.ID}`}
										variant="ghost"
										size="sm"
										class="text-warning hover:bg-warning/10 hover:text-warning w-full"
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
								class="liquid-hoverable group relative flex flex-col overflow-hidden rounded-lg"
							>
								<div class="border-border flex items-start justify-between border-b p-4">
									<div>
										<h3 class="num-mono text-foreground text-sm font-bold">{poNumber(po.ID)}</h3>
										<p class="text-muted-foreground text-xs"><TimeAgo date={po.OrderDate} /></p>
									</div>
									<div class={cn('uppercase', getStatusColor(po.Status))}>
										<svelte:component this={getStatusIcon(po.Status)} size={12} />
										{po.Status}
									</div>
								</div>

								<div class="flex-1 space-y-3 p-4">
									<div class="flex items-center justify-between text-sm">
										<span class="text-muted-foreground">{$t('orders.labels.supplier')}</span>
										<span class="text-foreground font-medium"
											>{po.Supplier?.Name || $t('orders.labels.unknown')}</span
										>
									</div>
									<div class="bg-muted space-y-1 rounded-md p-2 text-xs">
										{#each po.PurchaseOrderItems.slice(0, 2) as item}
											<div class="flex justify-between">
												<span class="text-muted-foreground truncate pr-2">{item.Product?.Name || 'Product'}</span>
												<span class="num-mono text-primary font-medium"
													>{$t('orders.status.qty')}: {item.OrderedQuantity}</span
												>
											</div>
										{/each}
										{#if po.PurchaseOrderItems.length > 2}
											<div class="text-muted-foreground text-[10px] italic">
												+{po.PurchaseOrderItems.length - 2}
												{$t('orders.labels.more')}
											</div>
										{/if}
									</div>
								</div>

								<div class="bg-muted/50 border-border mt-auto flex items-center justify-between border-t p-4">
									<div>
										{#if po.PurchaseOrderItems.length > 0}
											<p class="text-muted-foreground text-xs font-semibold">
												{$t('orders.status.total_items')}: <span class="num-mono">{po.PurchaseOrderItems.reduce(
													(acc, item) => acc + item.OrderedQuantity,
													0
												)}</span>
											</p>
										{/if}
									</div>
									<Button
										href="/purchase-orders/{po.ID}"
										variant="ghost"
										size="sm"
										class="text-primary hover:bg-accent hover:text-accent-foreground"
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
								class="liquid-hoverable group relative flex flex-col overflow-hidden rounded-lg"
							>
								<div class="border-border flex items-start justify-between border-b p-4">
									<div>
										<h3 class="num-mono text-foreground text-sm font-bold">
											{returnNumber(ret.ID)}
										</h3>
										<p class="text-muted-foreground text-xs"><TimeAgo date={ret.ReturnedAt} /></p>
									</div>
									<div class={cn('uppercase', getStatusColor(ret.Status))}>
										{ret.Status}
									</div>
								</div>

								<div class="flex-1 space-y-3 p-4">
									<div class="flex items-center justify-between text-sm">
										<span class="text-muted-foreground">{$t('orders.labels.supplier')}</span>
										<span class="text-foreground font-medium"
											>{ret.Supplier?.Name || $t('orders.labels.unknown')}</span
										>
									</div>
									<div class="flex items-center justify-between text-sm">
										<span class="text-muted-foreground">{$t('orders.labels.reason')}</span>
										<span class="text-foreground font-medium">{ret.Reason}</span>
									</div>

									<div class="bg-muted space-y-1 rounded-md p-2 text-xs">
										{#each ret.PurchaseReturnItems.slice(0, 2) as item}
											<div class="flex justify-between">
												<span class="text-muted-foreground truncate pr-2">{item.Product?.Name || 'Product'}</span>
												<span class="num-mono text-danger font-medium"
													>{$t('orders.status.qty')}: -{item.Quantity}</span
												>
											</div>
										{/each}
									</div>
								</div>

								<div class="bg-muted/50 border-border mt-auto flex items-center justify-between border-t p-4">
									<p class="num-mono text-foreground text-lg font-bold">{formatCurrency(ret.RefundAmount)}</p>
									<span class="text-muted-foreground text-xs font-semibold"
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
