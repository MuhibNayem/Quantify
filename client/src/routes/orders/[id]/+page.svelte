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

<div
	class="relative min-h-screen overflow-hidden p-8 lg:p-12"
>
	<div class="relative mx-auto max-w-5xl">
		{#if order}
			<div class="mb-4">
				<Breadcrumbs
					items={[{ label: 'Orders', href: '/orders' }, { label: orderNumber(order) }]}
				/>
			</div>
		{/if}
		<button
			onclick={() => goto('/orders')}
			class="text-muted-foreground hover:text-primary group mb-8 flex items-center gap-2 text-sm font-medium transition-colors"
		>
			<div
				class="border-border bg-card flex h-8 w-8 items-center justify-center rounded-md border shadow-sm transition-all group-hover:bg-accent"
			>
				←
			</div>
			Back to Orders
		</button>

		{#if loading}
			<div class="flex h-64 items-center justify-center">
				<div
					class="border-border border-t-primary h-10 w-10 animate-spin rounded-full border-4"
				></div>
			</div>
		{:else if error}
			<div class="liquid-panel border-danger/25 bg-danger/10 rounded-lg p-12 text-center">
				<div
					class="text-danger bg-danger/10 mx-auto mb-4 flex h-16 w-16 items-center justify-center rounded-lg text-2xl font-bold"
				>
					!
				</div>
				<h3 class="text-foreground text-lg font-semibold">Unable to load order</h3>
				<p class="text-danger mt-2">{error}</p>
			</div>
		{:else if order}
			<div in:fade={{ duration: 300 }}>
				<!-- Header Section -->
				<div class="mb-8 flex flex-col justify-between gap-6 md:flex-row md:items-center">
					<div>
						<div class="flex items-center gap-4">
							<h1 class="num-mono font-display text-foreground text-4xl font-bold tracking-tight">
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
						<div class="text-muted-foreground mt-2 flex items-center gap-4 text-sm">
							<span class="flex items-center gap-1.5">
								<Calendar class="h-3.5 w-3.5" /> <span class="num">{formatDate(order.OrderDate)}</span>
							</span>
							<span class="bg-border h-1 w-1 rounded-full"></span>
							<span class="flex items-center gap-1.5">
								<Package class="h-3.5 w-3.5" /> <span class="num">{order.OrderItems?.length || 0}</span> Items
							</span>
						</div>
					</div>

					<div class="flex flex-col items-end">
						<span class="overline">Total Amount</span>
						<div class="flex flex-col items-end">
							{#if order.AdjustedTotal !== undefined && order.AdjustedTotal < order.TotalAmount}
								<span class="kpi-num text-foreground text-3xl">
									{formatCurrency(order.AdjustedTotal)}
								</span>
								<span class="text-muted-foreground num-mono text-sm font-medium line-through">
									{formatCurrency(order.TotalAmount)}
								</span>
							{:else}
								<span class="kpi-num text-foreground text-3xl">
									{formatCurrency(order.TotalAmount)}
								</span>
							{/if}
						</div>
					</div>
				</div>

				<div class="grid gap-8 lg:grid-cols-3">
					<!-- Order Items -->
					<div class="space-y-6 lg:col-span-2">
						<div class="liquid-panel overflow-hidden rounded-lg p-0">
							<div class="border-border bg-muted/50 border-b px-6 py-4">
								<h3 class="font-display text-foreground font-semibold">Order Items</h3>
							</div>
							<div class="space-y-3 px-6 py-4">
								{#each order.OrderItems as item}
									<div
										class="liquid-hoverable group relative flex items-center justify-between rounded-md p-4"
									>
										<div class="relative z-10 flex items-center gap-4">
											<div
												class="bg-muted text-muted-foreground flex h-12 min-w-[3rem] items-center justify-center rounded-md px-3 font-medium"
											>
												<span class="num-mono">{item.Quantity}x</span>
												{#if item.ReturnedQty > 0}
													<span class="text-warning ml-1 text-xs">(-{item.ReturnedQty})</span>
												{/if}
											</div>
											<div>
												<p class="text-foreground font-semibold">
													{item.Product?.Name || `Product #${item.ProductID}`}
												</p>
												<p class="text-muted-foreground text-sm">
													{#if item.ReturnedQty > 0}
														<span class="text-warning font-medium"
															>Net Qty: <span class="num-mono">{item.Quantity - item.ReturnedQty}</span></span
														> •
													{/if}
													<span class="num-mono">{formatCurrency(item.UnitPrice)}</span> / unit
												</p>
											</div>
										</div>
										<div class="relative z-10 text-right">
											<div class="num-mono text-foreground font-bold">
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
					<div class="space-y-6">
						<!-- Actions Card -->
						<div class="liquid-panel space-y-6 rounded-lg p-6">
							<h3 class="border-border font-display text-foreground border-b pb-2 font-semibold">Actions</h3>

							{#if order.Status === 'COMPLETED'}
								{#if order.HasPendingReturn}
									<div class="rounded-md border border-warning/25 bg-warning/10 p-4">
										<p class="text-warning mb-1 text-xs font-semibold tracking-wide uppercase">
											Return Pending
										</p>
										<p class="text-warning/80 text-sm">A request is currently being processed.</p>
									</div>
									<button
										disabled
										class="bg-muted text-muted-foreground w-full cursor-not-allowed rounded-md py-3.5 font-medium"
									>
										Return Requested
									</button>
								{:else if isReturnEligible}
									<div class="rounded-md border border-info/25 bg-info/10 p-4">
										<p class="text-info mb-1 text-xs font-semibold tracking-wide uppercase">
											Return Window Open
										</p>
										<p class="text-info/80 text-sm">
											Until <span class="num">{returnDeadline ? formatDate(returnDeadline.toISOString()) : 'N/A'}</span>
										</p>
									</div>
									<button
										onclick={() => (showReturnModal = true)}
										class="bg-primary text-primary-foreground hover:bg-primary/90 w-full rounded-md py-3.5 font-medium shadow-sm transition-all active:scale-[0.99]"
									>
										Request Return
									</button>
								{:else}
									<div class="border-border bg-muted rounded-md border p-4 text-center">
										<p class="text-muted-foreground text-sm font-semibold">Return Window Closed</p>
										<p class="text-muted-foreground mt-1 text-xs">
											Deadline was <span class="num">{returnDeadline
												? formatDate(returnDeadline.toISOString())
												: 'N/A'}</span>
										</p>
									</div>
								{/if}
							{:else}
								<div class="border-border bg-muted/60 rounded-md border p-4 text-center">
									<p class="text-muted-foreground text-sm">Actions available when order is completed.</p>
								</div>
							{/if}
						</div>

						<!-- Order Info -->
						<div class="liquid-panel space-y-4 rounded-lg p-6">
							<h3 class="border-border font-display text-foreground border-b pb-2 font-semibold">
								Payment Info
							</h3>
							<div class="flex justify-between text-sm">
								<span class="text-muted-foreground">Method</span>
								<span
									class="border-border bg-muted text-foreground num-mono rounded-md border px-2 py-0.5 font-medium"
									>{order.PaymentMethod || 'N/A'}</span
								>
							</div>
						<div class="flex justify-between text-sm">
							<span class="text-muted-foreground">Status</span>
							<span class="text-foreground font-medium">{order.Status}</span>
						</div>
					</div>

					<!-- Activity History -->
					<div class="liquid-panel rounded-lg p-6">
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
			class="bg-foreground/50 fixed inset-0 z-50 flex items-center justify-center p-4"
			transition:fade={{ duration: 200 }}
		>
			<div class="w-full max-w-lg" transition:fly={{ y: 20, duration: 300 }}>
				<!-- Ledger Modal -->
				<div
					class="liquid-panel relative rounded-lg p-8 shadow-lg"
				>
					<button
						onclick={() => (showReturnModal = false)}
						class="text-muted-foreground hover:text-foreground hover:bg-muted absolute top-4 right-4 rounded-md p-1.5 transition-colors"
					>
						✕
					</button>

					<div class="mb-6">
						<h2 class="font-display text-foreground text-2xl font-bold tracking-tight">Request Return</h2>
						<p class="text-muted-foreground text-sm">Select items and quantities to return</p>
					</div>

					<div class="custom-scrollbar max-h-[60vh] space-y-3 overflow-y-auto pr-2">
						{#each order.OrderItems as item}
							{@const remainingQty = item.Quantity - item.ReturnedQty}
							{#if remainingQty > 0}
								<div
									class={`rounded-md border p-4 transition-all ${
										selectedItems[item.ID]
											? 'border-primary bg-accent/60 border-l-4'
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
													class="border-border peer h-5 w-5 cursor-pointer appearance-none rounded-md border transition-all checked:border-primary checked:bg-primary hover:border-primary/60"
												/>
												<div
													class="text-primary-foreground pointer-events-none absolute text-xs opacity-0 peer-checked:opacity-100"
												>
													✓
												</div>
											</div>
											<div>
												<p class="text-foreground font-medium">{item.Product?.Name}</p>
												<p class="text-muted-foreground text-xs">
													Max Return: <span class="num-mono">{remainingQty}</span>
												</p>
											</div>
										</div>

										{#if selectedItems[item.ID]}
											<div
												class="border-border bg-card flex items-center rounded-md border shadow-sm"
											>
												<button
													class="text-muted-foreground hover:text-primary hover:bg-muted h-7 w-7 rounded-l-md transition-colors"
													onclick={() =>
														updateQuantity(item.ID, selectedItems[item.ID] - 1, remainingQty)}
												>
													-
												</button>
												<span class="num-mono text-foreground w-8 text-center text-sm font-semibold">
													{selectedItems[item.ID]}
												</span>
												<button
													class="text-muted-foreground hover:text-primary hover:bg-muted h-7 w-7 rounded-r-md transition-colors"
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
							<label class="text-foreground mb-2 block text-sm font-medium">Reason for Return</label>
							<textarea
								bind:value={returnReason}
								class="liquid-textarea text-foreground placeholder:text-muted-foreground w-full px-4 py-3 text-sm"
								rows="3"
								placeholder="Please tell us why you are returning these items..."
							></textarea>
						</div>
					</div>

					{#if Object.keys(selectedItems).length > 0}
						<div class="bg-accent text-accent-foreground mt-4 flex items-center justify-between rounded-md p-4">
							<span class="text-sm font-medium">Estimated Refund</span>
							<span class="num-mono text-lg font-bold">{formatCurrency(estimatedRefund)}</span>
						</div>
						<p class="text-muted-foreground mt-2 text-right text-xs">
							Includes {$settings.tax_rate_percentage || 0}% tax adjustment
						</p>
					{/if}

					<div class="border-border mt-8 flex gap-3 border-t pt-6">
						<button
							onclick={() => (showReturnModal = false)}
							class="border-border bg-card text-muted-foreground hover:text-foreground hover:bg-muted flex-1 rounded-md border py-3 font-medium transition-colors"
						>
							Cancel
						</button>
						<button
							onclick={submitReturn}
							disabled={Object.keys(selectedItems).length === 0 || !returnReason}
							class="bg-primary text-primary-foreground hover:bg-primary/90 flex-1 rounded-md py-3 font-medium shadow-sm transition-all disabled:cursor-not-allowed disabled:opacity-50 disabled:shadow-none"
						>
							Submit Request
						</button>
					</div>
				</div>
			</div>
		</div>
	{/if}
</div>
