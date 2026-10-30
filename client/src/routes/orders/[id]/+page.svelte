<script lang="ts">
	import { page } from '$app/stores';
	import { onMount } from 'svelte';
	import { formatCurrency, formatDate, orderNumber } from '$lib/utils';
	import Breadcrumbs from '$lib/components/ui/breadcrumbs.svelte';
	import { fade, fly } from 'svelte/transition';
	import api from '$lib/api';
	import { auth } from '$lib/stores/auth';
	import { goto } from '$app/navigation';
	import { toast } from 'svelte-sonner';
	import { settings } from '$lib/stores/settings';
	import { addRecent } from '$lib/recent';
	import ActivityHistory from '$lib/components/activity-history.svelte';
	import { Calendar, Package } from 'lucide-svelte';

	let orderId = $page.params.id; // actually OrderNumber or ID, usually number in URL, but backend uses string OrderNumber.
	// Assuming route is /orders/[id], if id is the db ID, we fetch by ID. If orderNumber, by Number.
	// The previous code used $page.params.id.

	let order = $state<any>(null);
	let loading = $state(true);
	let showReturnModal = $state(false);
	let returnReason = $state('');
	let selectedItems = $state<Record<number, number>>({}); // { orderItemId: quantity }
	let error = $state<string | null>(null);

	// Dynamic settings derived state
	const returnWindowDays = $derived($settings.return_window_days || 30);
	const returnDeadline = $derived(
		order ? new Date(new Date(order.OrderDate).getTime() + returnWindowDays * 86400000) : null
	);
	const isReturnEligible = $derived(returnDeadline ? new Date() <= returnDeadline : false);
	const estimatedRefund = $derived(
		Object.entries(selectedItems).reduce((total, [itemIdStr, qty]) => {
			const item = order?.OrderItems?.find((i: any) => i.ID === parseInt(itemIdStr));
			if (!item) return total;
			return total + item.UnitPrice * qty;
		}, 0) *
			(1 + ($settings.tax_rate_percentage || 0) / 100)
	);

	onMount(async () => {
		if (!$auth.isAuthenticated) {
			goto('/login');
			return;
		}
		await fetchOrder();
	});

	const fetchOrder = async () => {
		loading = true;
		try {
			// Try fetching by ID or OrderNumber depending on what [id] is.
			// Usually [id] in frontend routes maps to an ID.
			// Backend endpoint: /sales/orders/{orderNumber}
			// If [id] is numeric ID, we might need a different endpoint or use the order number if available.
			// Let's assume the user navigated from a list where they clicked an Order Number.
			const response = await api.get(`/sales/orders/${orderId}`);
			order = response.data.order;
			if (order) {
				addRecent({
					type: 'order',
					id: Number(order.ID ?? orderId),
					label: order.OrderNumber ? `Order ${order.OrderNumber}` : `Order ${orderId}`,
					href: `/orders/${orderId}`
				});
			}
		} catch (err: any) {
			console.error('Error fetching order:', err);
			error = 'Failed to load order details.';
			toast.error('Could not fetch order details');
		} finally {
			loading = false;
		}
	};

	const toggleItemSelection = (itemId: number, maxQty: number) => {
		if (selectedItems[itemId]) {
			const newSelected = { ...selectedItems };
			delete newSelected[itemId];
			selectedItems = newSelected;
		} else {
			selectedItems = {
				...selectedItems,
				[itemId]: 1 // Start with 1
			};
		}
	};

	const updateQuantity = (itemId: number, newQty: number, maxQty: number) => {
		if (newQty < 1) return;
		if (newQty > maxQty) return;
		selectedItems = {
			...selectedItems,
			[itemId]: newQty
		};
	};

	const submitReturn = async () => {
		if (!order) return;

		try {
			const itemsToReturn = Object.entries(selectedItems).map(([itemIdStr, qty]) => ({
				order_item_id: parseInt(itemIdStr),
				quantity: qty,
				reason: returnReason,
				condition: 'GOOD' // Default, maybe add UI selector later
			}));

			const payload = {
				order_number: order.OrderNumber,
				items: itemsToReturn
			};

			await api.post('/returns/request', payload);
			toast.success('Return request submitted successfully');
			showReturnModal = false;
			returnReason = '';
			selectedItems = {};
			// Refresh order
			await fetchOrder();
		} catch (err: any) {
			console.error('Error requesting return:', err);
			toast.error(err.response?.data?.error || 'Failed to submit return request');
		}
	};
</script>

<div class="mx-auto w-full max-w-5xl px-4 py-8 sm:px-6 lg:px-8 lg:py-10">
	<div class="relative">
		{#if order}
			<div class="mb-4">
				<Breadcrumbs
					items={[{ label: 'Orders', href: '/orders' }, { label: orderNumber(order) }]}
				/>
			</div>
		{/if}
		<button
			onclick={() => goto('/orders')}
			class="group mb-8 flex items-center gap-2 text-sm font-medium text-muted-foreground transition-colors hover:text-foreground"
		>
			<span
				class="flex h-8 w-8 items-center justify-center rounded-lg border border-border bg-card transition-colors group-hover:bg-muted"
			>
				←
			</span>
			Back to Orders
		</button>

		{#if loading}
			<div class="flex h-64 items-center justify-center">
				<div
					class="h-10 w-10 animate-spin rounded-full border-4 border-muted border-t-primary"
				></div>
			</div>
		{:else if error}
			<div class="panel flex flex-col items-center gap-3 p-12 text-center">
				<div
					class="flex h-14 w-14 items-center justify-center rounded-xl bg-danger/10 text-2xl font-semibold text-danger"
				>
					!
				</div>
				<h3 class="text-lg font-semibold text-foreground">Unable to load order</h3>
				<p class="text-muted-foreground">{error}</p>
			</div>
		{:else if order}
			<div in:fade={{ duration: 300 }}>
				<!-- Header Section -->
				<header class="page-header mb-8">
					<div class="min-w-0">
						<div class="flex flex-wrap items-center gap-3">
							<h1 class="num-mono page-title">
								{orderNumber(order)}
							</h1>
							<span
								class={order.Status === 'COMPLETED'
									? 'badge-success'
									: order.Status === 'PENDING'
										? 'badge-warning'
										: 'badge-neutral'}
							>
								{order.Status}
							</span>
						</div>
						<div class="mt-2 flex flex-wrap items-center gap-3 text-sm text-muted-foreground">
							<span class="flex items-center gap-1.5">
								<Calendar class="h-3.5 w-3.5" />
								<span class="num">{formatDate(order.OrderDate)}</span>
							</span>
							<span class="h-1 w-1 rounded-full bg-border"></span>
							<span class="flex items-center gap-1.5">
								<Package class="h-3.5 w-3.5" />
								<span class="num">{order.OrderItems?.length || 0}</span> Items
							</span>
						</div>
					</div>

					<div class="flex flex-col items-end">
						<span class="overline">Total Amount</span>
						<div class="flex flex-col items-end">
							{#if order.AdjustedTotal !== undefined && order.AdjustedTotal < order.TotalAmount}
								<span class="kpi-num text-3xl text-foreground">
									{formatCurrency(order.AdjustedTotal)}
								</span>
								<span class="num-mono text-sm font-medium text-muted-foreground line-through">
									{formatCurrency(order.TotalAmount)}
								</span>
							{:else}
								<span class="kpi-num text-3xl text-foreground">
									{formatCurrency(order.TotalAmount)}
								</span>
							{/if}
						</div>
					</div>
				</header>

				<div class="grid gap-6 lg:grid-cols-3">
					<!-- Order Items -->
					<div class="lg:col-span-2">
						<div class="panel overflow-hidden">
							<div class="panel-header">
								<h3 class="section-title">Order Items</h3>
							</div>
							<div class="space-y-2 p-4">
								{#each order.OrderItems as item}
									<div
										class="group flex items-center justify-between gap-4 rounded-lg border border-border/70 p-4 transition-colors hover:bg-muted/40"
									>
										<div class="flex min-w-0 items-center gap-3">
											<div
												class="flex h-11 min-w-[3rem] items-center justify-center rounded-lg bg-muted px-3 text-sm font-medium text-muted-foreground"
											>
												<span class="num-mono">{item.Quantity}x</span>
												{#if item.ReturnedQty > 0}
													<span class="ml-1 text-xs text-warning">(-{item.ReturnedQty})</span>
												{/if}
											</div>
											<div class="min-w-0">
												<p class="truncate font-medium text-foreground">
													{item.Product?.Name || `Product #${item.ProductID}`}
												</p>
												<p class="text-sm text-muted-foreground">
													{#if item.ReturnedQty > 0}
														<span class="font-medium text-warning"
															>Net Qty: <span class="num-mono"
																>{item.Quantity - item.ReturnedQty}</span
															></span
														> •
													{/if}
													<span class="num-mono">{formatCurrency(item.UnitPrice)}</span> / unit
												</p>
											</div>
										</div>
										<div class="text-right">
											<div class="num-mono font-semibold text-foreground">
												{formatCurrency(item.TotalPrice)}
											</div>
											{#if item.ReturnedQty > 0}
												<div class="mt-1 flex justify-end">
													<span class="badge-warning px-1.5 py-0 text-[10px] uppercase">
														{item.ReturnedQty} Returned
													</span>
												</div>
											{/if}
										</div>
									</div>
								{/each}
							</div>
						</div>
					</div>

					<!-- Sidebar -->
					<div class="space-y-4">
						<!-- Actions Card -->
						<div class="panel space-y-4 p-5">
							<h3 class="section-title border-b border-border/80 pb-2.5">Actions</h3>

							{#if order.Status === 'COMPLETED'}
								{#if order.HasPendingReturn}
									<div class="rounded-lg border border-warning/25 bg-warning/10 p-4">
										<p class="mb-1 text-xs font-semibold tracking-wide text-warning uppercase">
											Return Pending
										</p>
										<p class="text-sm text-muted-foreground">
											A request is currently being processed.
										</p>
									</div>
									<button
										disabled
										class="w-full cursor-not-allowed rounded-lg bg-muted py-3 font-medium text-muted-foreground"
									>
										Return Requested
									</button>
								{:else if isReturnEligible}
									<div class="rounded-lg border border-info/25 bg-info/10 p-4">
										<p class="mb-1 text-xs font-semibold tracking-wide text-info uppercase">
											Return Window Open
										</p>
										<p class="text-sm text-muted-foreground">
											Until <span class="num"
												>{returnDeadline ? formatDate(returnDeadline.toISOString()) : 'N/A'}</span
											>
										</p>
									</div>
									<button
										onclick={() => (showReturnModal = true)}
										class="btn-soothe w-full py-3 font-semibold"
									>
										Request Return
									</button>
								{:else}
									<div class="rounded-lg border border-border bg-muted/60 p-4 text-center">
										<p class="text-sm font-semibold text-foreground">Return Window Closed</p>
										<p class="mt-1 text-xs text-muted-foreground">
											Deadline was <span class="num"
												>{returnDeadline ? formatDate(returnDeadline.toISOString()) : 'N/A'}</span
											>
										</p>
									</div>
								{/if}
							{:else}
								<div class="rounded-lg border border-border bg-muted/60 p-4 text-center">
									<p class="text-sm text-muted-foreground">
										Actions available when order is completed.
									</p>
								</div>
							{/if}
						</div>

						<!-- Order Info -->
						<div class="panel p-5">
							<h3 class="section-title border-b border-border/80 pb-2.5">Payment Info</h3>
							<div class="mt-3 space-y-2 text-sm">
								<div class="flex items-center justify-between gap-3">
									<span class="text-muted-foreground">Method</span>
									<span
										class="num-mono rounded-md bg-muted px-2 py-0.5 text-xs font-medium text-foreground"
										>{order.PaymentMethod || 'N/A'}</span
									>
								</div>
								<div class="flex items-center justify-between gap-3">
									<span class="text-muted-foreground">Status</span>
									<span class="font-medium text-foreground">{order.Status}</span>
								</div>
							</div>
						</div>

						<!-- Activity History -->
						<div class="panel p-5">
							<ActivityHistory entity="Order" entityId={order.ID ?? orderId} />
						</div>
					</div>
				</div>
			</div>
		{/if}
	</div>

	<!-- Return Modal -->
	{#if showReturnModal && order}
		<div
			class="fixed inset-0 z-50 flex items-center justify-center bg-background/70 p-4 backdrop-blur-sm"
			transition:fade={{ duration: 200 }}
		>
			<div class="w-full max-w-lg" transition:fly={{ y: 20, duration: 300 }}>
				<!-- Ledger Modal -->
				<div class="surface-3 relative rounded-xl p-6 shadow-lg sm:p-8">
					<button
						onclick={() => (showReturnModal = false)}
						class="absolute top-4 right-4 rounded-lg p-1.5 text-muted-foreground transition-colors hover:bg-muted hover:text-foreground"
					>
						✕
					</button>

					<div class="mb-6">
						<h2 class="font-display text-xl font-semibold tracking-tight text-foreground">
							Request Return
						</h2>
						<p class="mt-1 text-sm text-muted-foreground">Select items and quantities to return</p>
					</div>

					<div class="custom-scrollbar max-h-[60vh] space-y-3 overflow-y-auto pr-2">
						{#each order.OrderItems as item}
							{@const remainingQty = item.Quantity - item.ReturnedQty}
							{#if remainingQty > 0}
								<div
									class={`rounded-lg border p-4 transition-all ${
										selectedItems[item.ID]
											? 'border-primary/45 bg-accent'
											: 'border-border hover:bg-muted/60'
									}`}
								>
									<div class="flex items-center justify-between">
										<div class="flex items-center gap-3">
											<div class="relative flex h-5 w-5 items-center justify-center">
												<input
													type="checkbox"
													checked={!!selectedItems[item.ID]}
													onchange={() => toggleItemSelection(item.ID, remainingQty)}
													class="peer h-5 w-5 cursor-pointer appearance-none rounded-md border border-border transition-all checked:border-primary checked:bg-primary hover:border-primary/60"
												/>
												<div
													class="pointer-events-none absolute text-xs text-primary-foreground opacity-0 peer-checked:opacity-100"
												>
													✓
												</div>
											</div>
											<div>
												<p class="font-medium text-foreground">{item.Product?.Name}</p>
												<p class="text-xs text-muted-foreground">
													Max Return: <span class="num-mono">{remainingQty}</span>
												</p>
											</div>
										</div>

										{#if selectedItems[item.ID]}
											<div
												class="flex items-center rounded-lg border border-border bg-card shadow-xs"
											>
												<button
													class="h-7 w-7 rounded-l-lg text-muted-foreground transition-colors hover:bg-muted hover:text-foreground"
													onclick={() =>
														updateQuantity(item.ID, selectedItems[item.ID] - 1, remainingQty)}
												>
													-
												</button>
												<span
													class="num-mono w-8 text-center text-sm font-semibold text-foreground"
												>
													{selectedItems[item.ID]}
												</span>
												<button
													class="h-7 w-7 rounded-r-lg text-muted-foreground transition-colors hover:bg-muted hover:text-foreground"
													onclick={() =>
														updateQuantity(item.ID, selectedItems[item.ID] + 1, remainingQty)}
												>
													+
												</button>
											</div>
										{/if}
									</div>
								</div>
							{/if}
						{/each}

						<div class="pt-4">
							<label class="mb-2 block text-sm font-medium text-foreground">Reason for Return</label
							>
							<textarea
								bind:value={returnReason}
								class="glass-textarea w-full px-3 py-2.5 text-sm text-foreground placeholder:text-muted-foreground"
								rows="3"
								placeholder="Please tell us why you are returning these items..."
							></textarea>
						</div>
					</div>

					{#if Object.keys(selectedItems).length > 0}
						<div
							class="mt-4 flex items-center justify-between rounded-lg bg-accent p-4 text-accent-foreground"
						>
							<span class="text-sm font-medium">Estimated Refund</span>
							<span class="kpi-num text-lg">{formatCurrency(estimatedRefund)}</span>
						</div>
						<p class="mt-2 text-right text-xs text-muted-foreground">
							Includes {$settings.tax_rate_percentage || 0}% tax adjustment
						</p>
					{/if}

					<div class="mt-6 flex gap-3 border-t border-border/80 pt-5">
						<button
							onclick={() => (showReturnModal = false)}
							class="btn-soft flex-1 py-3 font-medium"
						>
							Cancel
						</button>
						<button
							onclick={submitReturn}
							disabled={Object.keys(selectedItems).length === 0 || !returnReason}
							class="btn-soothe flex-1 py-3 font-semibold disabled:cursor-not-allowed disabled:opacity-50"
						>
							Submit Request
						</button>
					</div>
				</div>
			</div>
		</div>
	{/if}
</div>
