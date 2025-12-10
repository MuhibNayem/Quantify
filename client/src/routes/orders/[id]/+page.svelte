<script lang="ts">
	import { page } from '$app/stores';
	import { onMount } from 'svelte';
	import GlassCard from '$lib/components/ui/GlassCard.svelte';
	import { formatCurrency, formatDate } from '$lib/utils';
	import { adaptiveText, liquidGlass, glassCard } from '$lib/styles/liquid-glass';
	import { fade, fly } from 'svelte/transition';
	import api from '$lib/api';
	import { auth } from '$lib/stores/auth';
	import { goto } from '$app/navigation';
	import { toast } from 'svelte-sonner';
	import { settings } from '$lib/stores/settings';

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
			return total + (item.UnitPrice * qty);
		}, 0) * (1 + ($settings.tax_rate_percentage || 0) / 100)
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
	class="relative min-h-screen overflow-hidden bg-slate-50/50 p-8 font-sans text-slate-800 lg:p-12"
>
	<!-- Background -->
	<div
		class="absolute inset-0 -z-10 bg-[radial-gradient(ellipse_at_top,_var(--tw-gradient-stops))] from-blue-100/20 via-slate-50/20 to-white/20"
	></div>
	<!-- Organic Mesh Gradient Background (Apple-style) -->
	<div class="pointer-events-none absolute inset-0 overflow-hidden opacity-60">
		<div
			class="absolute left-[10%] top-[5%] h-[600px] w-[600px] rounded-full bg-gradient-to-br from-blue-200 via-cyan-100 to-transparent blur-[120px]"
		></div>
		<div
			class="absolute right-[5%] top-[30%] h-[500px] w-[500px] rounded-full bg-gradient-to-tr from-purple-200 via-pink-100 to-transparent blur-[100px]"
		></div>
		<div
			class="absolute bottom-[10%] left-[30%] h-[400px] w-[400px] rounded-full bg-gradient-to-tl from-indigo-200 via-violet-100 to-transparent blur-[90px]"
		></div>
	</div>

	<div class="relative mx-auto max-w-5xl">
		<button
			onclick={() => goto('/orders')}
			class="group mb-8 flex items-center gap-2 text-sm font-medium text-slate-400 transition-colors hover:text-blue-500"
		>
			<div
				class="flex h-8 w-8 items-center justify-center rounded-full bg-white shadow-sm ring-1 ring-slate-100 transition-all group-hover:bg-blue-50 group-hover:text-blue-500 group-hover:ring-blue-100"
			>
				←
			</div>
			Back to Orders
		</button>

		{#if loading}
			<div class="flex h-64 items-center justify-center">
				<div
					class="h-10 w-10 animate-spin rounded-full border-4 border-slate-100 border-t-blue-400"
				></div>
			</div>
		{:else if error}
			<GlassCard class="border-red-100 bg-red-50/40 p-12 text-center backdrop-blur-xl">
				<div
					class="mx-auto mb-4 flex h-16 w-16 items-center justify-center rounded-full bg-red-50 text-red-400"
				>
					!
				</div>
				<h3 class="text-lg font-semibold text-red-800">Unable to load order</h3>
				<p class="mt-2 text-red-600/70">{error}</p>
			</GlassCard>
		{:else if order}
			<div in:fade={{ duration: 300 }}>
				<!-- Header Section -->
				<div class="mb-8 flex flex-col justify-between gap-6 md:flex-row md:items-center">
					<div>
						<div class="flex items-center gap-4">
							<h1 class="text-4xl font-bold tracking-tight text-slate-900">
								Order #{order.OrderNumber}
							</h1>
							<span
								class={`rounded-full px-3 py-1 text-xs font-bold uppercase tracking-wide shadow-sm ring-1 backdrop-blur-md ${
									order.Status === 'COMPLETED'
										? 'bg-emerald-50/50 text-emerald-600 ring-emerald-100'
										: order.Status === 'PENDING'
											? 'bg-amber-50/50 text-amber-600 ring-amber-100'
											: 'bg-slate-50 text-slate-500 ring-slate-100'
								}`}
							>
								{order.Status}
							</span>
						</div>
						<div class="mt-2 flex items-center gap-4 text-sm text-slate-400">
							<span class="flex items-center gap-1.5">
								📅 {formatDate(order.OrderDate)}
							</span>
							<span class="h-1 w-1 rounded-full bg-slate-200"></span>
							<span class="flex items-center gap-1.5">
								📦 {order.OrderItems?.length || 0} Items
							</span>
						</div>
					</div>

					<div class="flex flex-col items-end">
						<span class="text-sm font-medium text-slate-500">Total Amount</span>
						<div class="flex flex-col items-end">
							{#if order.AdjustedTotal !== undefined && order.AdjustedTotal < order.TotalAmount}
								<span class="text-3xl font-bold tracking-tight text-slate-900">
									{formatCurrency(order.AdjustedTotal)}
								</span>
								<span class="text-sm font-medium text-slate-400 line-through decoration-slate-400">
									{formatCurrency(order.TotalAmount)}
								</span>
							{:else}
								<span class="text-3xl font-bold tracking-tight text-slate-900">
									{formatCurrency(order.TotalAmount)}
								</span>
							{/if}
						</div>
					</div>
				</div>

				<div class="grid gap-8 lg:grid-cols-3">
					<!-- Order Items -->
					<div class="space-y-6 lg:col-span-2">
						<GlassCard
							class="liquid-panel overflow-hidden rounded-[28px] bg-gradient-to-br from-white/40 via-white/20 to-white/5 p-0 shadow-[0_35px_90px_-60px_rgba(59,130,246,0.55)] transition-all hover:scale-[1.005]"
						>
							<div class="border-b border-white/20 bg-white/5 px-6 py-4 backdrop-blur-md">
								<h3 class="font-semibold text-slate-800 drop-shadow-sm">Order Items</h3>
							</div>
							<div class="space-y-3 px-6 py-4">
								{#each order.OrderItems as item}
									<div
										class="liquid-hoverable group relative flex items-center justify-between rounded-xl bg-gradient-to-br from-white/50 via-white/30 to-white/10 p-4 transition-all duration-200 hover:bg-white/60"
									>
										<!-- Glass Highlight Top -->
										<div
											class="absolute inset-x-0 top-0 h-px bg-gradient-to-r from-transparent via-white/40 to-transparent opacity-50"
										></div>
										<!-- Glass Sheen -->
										<div
											class="absolute inset-0 bg-gradient-to-b from-white/10 to-transparent opacity-0 transition-opacity group-hover:opacity-100"
										></div>

										<div class="relative z-10 flex items-center gap-4">
											<div
												class="flex h-12 min-w-[3rem] px-3 items-center justify-center rounded-2xl bg-white font-medium text-slate-500 shadow-sm ring-1 ring-slate-100"
											>
												{item.Quantity}x
												{#if item.ReturnedQty > 0}
													<span class="text-xs text-amber-500 ml-1">(-{item.ReturnedQty})</span>
												{/if}
											</div>
											<div>
												<p class="font-semibold text-slate-900">
													{item.Product?.Name || `Product #${item.ProductID}`}
												</p>
												<p class="text-sm text-slate-400">
													{#if item.ReturnedQty > 0}
														<span class="text-amber-600 font-medium">Net Qty: {item.Quantity - item.ReturnedQty}</span> • 
													{/if}
													{formatCurrency(item.UnitPrice)} / unit
												</p>
											</div>
										</div>
										<div class="relative z-10 text-right">
											<div class="font-bold text-slate-900 drop-shadow-sm">
												{formatCurrency(item.TotalPrice)}
											</div>
											{#if item.ReturnedQty > 0}
												<div class="mt-1 flex justify-end">
													<span
														class="rounded-full bg-amber-50 px-2 py-0.5 text-[10px] font-bold uppercase tracking-wide text-amber-500 ring-1 ring-amber-100"
													>
														{item.ReturnedQty} Returned
													</span>
												</div>
											{/if}
										</div>
									</div>
								{/each}
							</div>
						</GlassCard>
					</div>

					<!-- Sidebar -->
					<div class="space-y-6">
						<!-- Actions Card -->
						<GlassCard
							class="liquid-panel space-y-6 rounded-[28px] bg-gradient-to-br from-white/40 via-white/20 to-white/5 p-6 shadow-[0_35px_90px_-60px_rgba(59,130,246,0.55)] hover:scale-[1.01]"
						>
							<h3 class="border-b border-white/20 pb-2 font-semibold text-slate-800">Actions</h3>

							{#if order.Status === 'COMPLETED'}
								{#if order.HasPendingReturn}
									<div class="rounded-xl bg-amber-50/30 p-4 ring-1 ring-amber-100/50">
										<p class="mb-1 text-xs font-semibold uppercase tracking-wide text-amber-500">
											Return Pending
										</p>
										<p class="text-sm text-amber-700/80">A request is currently being processed.</p>
									</div>
									<button
										disabled
										class="w-full cursor-not-allowed rounded-xl bg-slate-100 py-3.5 font-medium text-slate-400"
									>
										Return Requested
									</button>
								{:else if isReturnEligible}
									<div class="rounded-xl bg-blue-50/30 p-4 ring-1 ring-blue-100/50">
										<p class="mb-1 text-xs font-semibold uppercase tracking-wide text-blue-500">
											Return Window Open
										</p>
										<p class="text-sm text-blue-700/80">
											Until {returnDeadline ? formatDate(returnDeadline.toISOString()) : 'N/A'}
										</p>
									</div>
									<button
										onclick={() => (showReturnModal = true)}
										class="w-full transform rounded-xl bg-blue-500 py-3.5 font-medium text-white shadow-lg shadow-blue-200 transition-all hover:scale-[1.01] hover:bg-blue-600 active:scale-[0.99]"
									>
										Request Return
									</button>
								{:else}
									<div class="rounded-xl bg-slate-50 p-4 text-center ring-1 ring-slate-100">
										<p class="text-sm font-semibold text-slate-500">Return Window Closed</p>
										<p class="mt-1 text-xs text-slate-400">
											Deadline was {returnDeadline
												? formatDate(returnDeadline.toISOString())
												: 'N/A'}
										</p>
									</div>
								{/if}
							{:else}
								<div class="rounded-xl bg-white/40 p-4 text-center ring-1 ring-white/60">
									<p class="text-sm text-slate-500">Actions available when order is completed.</p>
								</div>
							{/if}
						</GlassCard>

						<!-- Order Info -->
						<GlassCard
							class="liquid-panel space-y-4 rounded-[28px] bg-gradient-to-br from-white/40 via-white/20 to-white/5 p-6 shadow-[0_35px_90px_-60px_rgba(148,163,184,0.5)] hover:scale-[1.01]"
						>
							<h3 class="border-b border-white/20 pb-2 font-semibold text-slate-800">
								Payment Info
							</h3>
							<div class="flex justify-between text-sm">
								<span class="text-slate-500">Method</span>
								<span
									class="rounded-md bg-white/40 px-2 py-0.5 font-medium text-slate-800 shadow-sm ring-1 ring-black/5"
									>{order.PaymentMethod || 'N/A'}</span
								>
							</div>
							<div class="flex justify-between text-sm">
								<span class="text-slate-500">Status</span>
								<span class="font-medium text-slate-800">{order.Status}</span>
							</div>
						</GlassCard>
					</div>
				</div>
			</div>
		{/if}
	</div>

	<!-- Return Modal -->
	{#if showReturnModal && order}
		<div
			class="fixed inset-0 z-50 flex items-center justify-center bg-slate-900/40 p-4 backdrop-blur-sm"
			transition:fade={{ duration: 200 }}
		>
			<div class="w-full max-w-lg" transition:fly={{ y: 20, duration: 300 }}>
				<!-- Liquid Glass Modal -->
				<div
					class="liquid-panel relative rounded-[32px] bg-gradient-to-br from-white/95 via-blue-50/90 to-indigo-50/80 p-8 shadow-[0_45px_120px_-60px_rgba(0,0,0,0.3)] backdrop-blur-3xl"
				>
					<button
						onclick={() => (showReturnModal = false)}
						class="absolute right-4 top-4 rounded-full bg-slate-50 p-1.5 text-slate-300 transition-colors hover:bg-slate-100 hover:text-slate-500"
					>
						✕
					</button>

					<div class="mb-6">
						<h2 class="text-2xl font-bold tracking-tight text-slate-900">Request Return</h2>
						<p class="text-sm text-slate-500">Select items and quantities to return</p>
					</div>

					<div class="custom-scrollbar max-h-[60vh] space-y-3 overflow-y-auto pr-2">
						{#each order.OrderItems as item}
							{@const remainingQty = item.Quantity - item.ReturnedQty}
							{#if remainingQty > 0}
								<div
									class={`liquid-hoverable rounded-2xl bg-gradient-to-br from-white/50 via-white/30 to-white/10 p-4 transition-all ${
										selectedItems[item.ID]
											? 'border-l-4 border-blue-400 bg-white/60'
											: 'hover:bg-white/60'
									}`}
								>
									<div class="flex items-center justify-between">
										<div class="flex items-center gap-3">
											<div class="relative flex h-5 w-5 items-center justify-center">
												<input
													type="checkbox"
													checked={!!selectedItems[item.ID]}
													onchange={() => toggleItemSelection(item.ID, remainingQty)}
													class="peer h-5 w-5 cursor-pointer appearance-none rounded border border-slate-200 transition-all checked:border-blue-500 checked:bg-blue-500 hover:border-blue-400"
												/>
												<div
													class="pointer-events-none absolute text-xs text-white opacity-0 peer-checked:opacity-100"
												>
													✓
												</div>
											</div>
											<div>
												<p class="font-medium text-slate-900">{item.Product?.Name}</p>
												<p class="text-xs text-slate-500">
													Max Return: {remainingQty}
												</p>
											</div>
										</div>

										{#if selectedItems[item.ID]}
											<div
												class="flex items-center rounded-lg bg-white shadow-sm ring-1 ring-slate-100"
											>
												<button
													class="h-7 w-7 rounded-l-lg text-slate-400 transition-colors hover:bg-slate-50 hover:text-blue-500"
													onclick={() =>
														updateQuantity(item.ID, selectedItems[item.ID] - 1, remainingQty)}
												>
													-
												</button>
												<span class="w-8 text-center text-sm font-semibold text-slate-800">
													{selectedItems[item.ID]}
												</span>
												<button
													class="h-7 w-7 rounded-r-lg text-slate-400 transition-colors hover:bg-slate-50 hover:text-blue-500"
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
							<label class="mb-2 block text-sm font-medium text-slate-700">Reason for Return</label>
							<textarea
								bind:value={returnReason}
								class="liquid-textarea w-full px-4 py-3 text-sm text-slate-800 placeholder:text-slate-400"
								rows="3"
								placeholder="Please tell us why you are returning these items..."
							></textarea>
						</div>
					</div>

					{#if Object.keys(selectedItems).length > 0}
						<div class="mt-4 flex items-center justify-between rounded-xl bg-blue-50 p-4">
							<span class="text-sm font-medium text-blue-700">Estimated Refund</span>
							<span class="text-lg font-bold text-blue-800">{formatCurrency(estimatedRefund)}</span>
						</div>
						<p class="mt-2 text-xs text-slate-400 text-right">
							Includes {($settings.tax_rate_percentage || 0)}% tax adjustment
						</p>
					{/if}

					<div class="mt-8 flex gap-3 border-t border-slate-50 pt-6">
						<button
							onclick={() => (showReturnModal = false)}
							class="flex-1 rounded-xl border border-slate-100 bg-white py-3 font-medium text-slate-500 transition-colors hover:bg-slate-50 hover:text-slate-700"
						>
							Cancel
						</button>
						<button
							onclick={submitReturn}
							disabled={Object.keys(selectedItems).length === 0 || !returnReason}
							class="flex-1 rounded-xl bg-blue-500 py-3 font-medium text-white shadow-lg shadow-blue-200 transition-all hover:bg-blue-600 disabled:cursor-not-allowed disabled:opacity-50 disabled:shadow-none"
						>
							Submit Request
						</button>
					</div>
				</div>
			</div>
		</div>
	{/if}
</div>
