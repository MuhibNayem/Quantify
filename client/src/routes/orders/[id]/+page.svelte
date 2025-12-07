<script lang="ts">
	import { page } from '$app/stores';
	import { onMount } from 'svelte';
	import GlassCard from '$lib/components/ui/GlassCard.svelte';
	import { formatDate, formatCurrency } from '$lib/utils';
	import { fade, fly } from 'svelte/transition';
	import api from '$lib/api';
	import { auth } from '$lib/stores/auth';
	import { goto } from '$app/navigation';
	import { toast } from 'svelte-sonner';

	let orderId = $page.params.id; // actually OrderNumber
	let order = $state(null);
	let loading = $state(true);
	let showReturnModal = $state(false);
	let returnReason = $state('');
	let selectedItems = $state({}); // { orderItemId: quantity }
	let error = $state(null);

	onMount(async () => {
		if (!$auth.permissions.includes('pos.view')) {
			toast.error('Permission denied', { description: 'You cannot view order details.' });
			goto('/orders');
			return;
		}

		try {
			const res = await api.get(`/sales/orders/${orderId}`);
			order = res.data.order;
		} catch (err) {
			console.error('Failed to fetch order', err);
			// Fallback or error state
			error = 'Failed to load order.';
		} finally {
			loading = false;
		}
	});

	function toggleItemSelection(orderItemId, maxQty) {
		if (selectedItems[orderItemId]) {
			const newItems = { ...selectedItems };
			delete newItems[orderItemId];
			selectedItems = newItems;
		} else {
			selectedItems = { ...selectedItems, [orderItemId]: 1 };
		}
	}

	function updateQuantity(orderItemId, qty, max) {
		if (qty > 0 && qty <= max) {
			selectedItems = { ...selectedItems, [orderItemId]: qty };
		}
	}

	async function submitReturn() {
		// Construct payload strictly matching ReturnRequest in backend
		/* 
       Items: []struct {
		OrderItemID uint   `json:"order_item_id"`
		Quantity    int    `json:"quantity"`
        ...
       }
    */
		const itemsToReturn = Object.entries(selectedItems).map(([oid, qty]) => ({
			order_item_id: parseInt(oid),
			quantity: qty,
			condition: 'GOOD', // Default
			reason: returnReason
		}));

		const payload = {
			order_number: order.OrderNumber,
			items: itemsToReturn
		};

		console.log('Submitting Return:', payload);

		try {
			await api.post('/returns/request', payload);
			toast.success('Return Request Submitted', {
				description: 'The return has been created successfully.'
			});
			showReturnModal = false;
			// Refresh?
			setTimeout(() => window.location.reload(), 1500);
		} catch (err: any) {
			toast.error('Return Failed', { description: err.response?.data?.error || 'Unknown error' });
		}
	}
</script>

<div class="container mx-auto space-y-8 p-6">
	<div class="flex items-center justify-between">
		<h1
			class="bg-gradient-to-r from-purple-400 to-pink-600 bg-clip-text text-3xl font-bold text-transparent"
		>
			Order Details
		</h1>
		<a href="/orders" class="text-gray-400 transition-colors hover:text-white">← Back to Orders</a>
	</div>

	{#if loading}
		<div class="text-center text-gray-400">Loading details...</div>
	{:else if error}
		<div class="text-center text-red-400">{error}</div>
	{:else if order}
		<div class="grid gap-6 lg:grid-cols-3">
			<!-- Order Info -->
			<div class="space-y-6 lg:col-span-2">
				<GlassCard>
					<div class="mb-6 flex items-start justify-between">
						<div>
							<h2 class="text-xl font-semibold text-white">Order #{order.OrderNumber}</h2>
							<p class="text-sm text-gray-400">{formatDate(order.OrderDate)}</p>
						</div>
						<span class="rounded-full bg-green-400/10 px-3 py-1 text-sm text-green-400">
							{order.Status}
						</span>
					</div>

					<div class="space-y-4">
						{#each order.OrderItems as item}
							<div
								class="flex items-center justify-between rounded-lg bg-white/5 p-4 transition-colors hover:bg-white/10"
							>
								<div class="flex items-center gap-4">
									<!-- Placeholder image or item image -->
									<div
										class="flex h-12 w-12 items-center justify-center rounded bg-slate-700 text-xs text-slate-400"
									>
										IMG
									</div>
									<div>
										<h3 class="font-medium text-white">
											{item.Product?.Name || 'Unknown Product'}
										</h3>
										<p class="text-sm text-gray-400">
											Qty: {item.Quantity} x {formatCurrency(item.UnitPrice)}
										</p>
										{#if item.IsReturned}
											<span class="text-xs text-orange-400">Returned: {item.ReturnedQty}</span>
										{/if}
									</div>
								</div>
								<span class="font-semibold text-white">{formatCurrency(item.TotalPrice)}</span>
							</div>
						{/each}
					</div>
				</GlassCard>
			</div>

			<!-- Summary & Actions -->
			<div class="space-y-6">
				<GlassCard>
					<h3 class="mb-4 text-lg font-semibold text-white">Summary</h3>
					<div class="space-y-2 text-sm">
						<div class="flex justify-between text-gray-400">
							<span>Subtotal</span>
							<span>{formatCurrency(order.TotalAmount)}</span>
							<!-- Note: TotalAmount is usually total. Adding shipping logic if needed -->
						</div>
						<div
							class="my-2 flex justify-between border-t border-white/10 pt-2 text-lg font-bold text-white"
						>
							<span>Total</span>
							<span>{formatCurrency(order.TotalAmount)}</span>
						</div>
					</div>

					{#if order.Status === 'COMPLETED'}
						<button
							onclick={() => (showReturnModal = true)}
							class="mt-6 w-full transform rounded-lg bg-gradient-to-r from-purple-600 to-pink-600 py-3 font-semibold text-white shadow-lg transition-all hover:scale-[1.02] hover:from-purple-500 hover:to-pink-500"
						>
							Request Return
						</button>
					{/if}
				</GlassCard>
			</div>
		</div>
	{/if}

	<!-- Return Modal -->
	{#if showReturnModal && order}
		<div
			class="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm"
			transition:fade
		>
			<div class="w-full max-w-lg" transition:fly={{ y: 20 }}>
				<GlassCard class="relative">
					<button
						onclick={() => (showReturnModal = false)}
						class="absolute right-4 top-4 text-gray-400 hover:text-white"
					>
						✕
					</button>

					<h2 class="mb-6 text-2xl font-bold text-white">Request Return</h2>

					<div class="max-h-[60vh] space-y-4 overflow-y-auto pr-2">
						<p class="text-sm text-gray-300">Select items to return:</p>
						{#each order.OrderItems as item}
							{@const remainingQty = item.Quantity - item.ReturnedQty}
							{#if remainingQty > 0}
								<div
									class="flex items-center justify-between rounded-lg border border-white/10 bg-black/20 p-3"
								>
									<div class="flex items-center gap-3">
										<input
											type="checkbox"
											checked={!!selectedItems[item.ID]}
											onchange={() => toggleItemSelection(item.ID, remainingQty)}
											class="rounded border-gray-600 bg-gray-700 text-purple-600 focus:ring-purple-500"
										/>
										<div>
											<p class="text-sm font-medium text-white">{item.Product?.Name}</p>
											<p class="text-xs text-gray-500">
												{formatCurrency(item.UnitPrice)} (Max: {remainingQty})
											</p>
										</div>
									</div>

									{#if selectedItems[item.ID]}
										<div class="flex items-center gap-2">
											<button
												class="flex h-6 w-6 items-center justify-center rounded bg-white/10 text-white hover:bg-white/20"
												onclick={() =>
													updateQuantity(item.ID, selectedItems[item.ID] - 1, remainingQty)}
												>-</button
											>
											<span class="w-4 text-center text-sm text-white"
												>{selectedItems[item.ID]}</span
											>
											<button
												class="flex h-6 w-6 items-center justify-center rounded bg-white/10 text-white hover:bg-white/20"
												onclick={() =>
													updateQuantity(item.ID, selectedItems[item.ID] + 1, remainingQty)}
												>+</button
											>
										</div>
									{/if}
								</div>
							{/if}
						{/each}

						<div class="pt-4">
							<label class="mb-2 block text-sm font-medium text-gray-300">Reason for Return</label>
							<textarea
								bind:value={returnReason}
								class="w-full rounded-lg border border-white/10 bg-black/20 p-3 text-white outline-none focus:ring-2 focus:ring-purple-500"
								rows="3"
								placeholder="Please describe why you are returning these items..."
							></textarea>
						</div>
					</div>

					<div class="mt-8 flex gap-3">
						<button
							onclick={() => (showReturnModal = false)}
							class="flex-1 rounded-lg border border-white/10 py-2 text-gray-300 transition-colors hover:bg-white/5"
						>
							Cancel
						</button>
						<button
							onclick={submitReturn}
							disabled={Object.keys(selectedItems).length === 0 || !returnReason}
							class="flex-1 rounded-lg bg-purple-600 py-2 font-medium text-white transition-colors hover:bg-purple-500 disabled:cursor-not-allowed disabled:opacity-50"
						>
							Submit Request
						</button>
					</div>
				</GlassCard>
			</div>
		</div>
	{/if}
</div>
