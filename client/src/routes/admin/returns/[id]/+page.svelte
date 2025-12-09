<script lang="ts">
	import { page } from '$app/stores';
	import { onMount } from 'svelte';
	import { returnsApi } from '$lib/api/returns';
	import { goto } from '$app/navigation';
	import { formatDate, formatCurrency } from '$lib/utils';
	import GlassCard from '$lib/components/ui/GlassCard.svelte';
	import { toast } from 'svelte-sonner';
	import { fade, fly } from 'svelte/transition';

	let returnId = $page.params.id;
	let returnRequest: any = null;
	let loading = true;
	let processing = false;

	onMount(async () => {
		try {
			const res = await returnsApi.getReturn(Number(returnId));
			returnRequest = res.return; // Backend returns { return: ... }
		} catch (error) {
			console.error('Failed to load return:', error);
			toast.error('Failed to load return details');
		} finally {
			loading = false;
		}
	});

	async function handleProcess(action: 'approve' | 'reject') {
		if (processing) return;
		processing = true;
		try {
			await returnsApi.processReturn(Number(returnId), action);
			toast.success(`Return request ${action}ed successfully`);
			// Reload
			const res = await returnsApi.getReturn(Number(returnId));
			returnRequest = res.return;
		} catch (error) {
			console.error(`Failed to ${action} return:`, error);
			toast.error(`Failed to ${action} return`);
		} finally {
			processing = false;
		}
	}
</script>

<div class="container mx-auto max-w-4xl space-y-8 p-6">
	<!-- Header -->
	<div class="flex items-center justify-between">
		<button
			onclick={() => goto('/orders')}
			class="flex items-center gap-2 text-sm text-slate-400 transition-colors hover:text-white"
		>
			← Back to Orders
		</button>
		<div class="flex items-center gap-2">
			{#if returnRequest}
				<span
					class={returnRequest.Status === 'PENDING'
						? 'rounded-full bg-yellow-500/20 px-3 py-1 text-xs font-semibold text-yellow-400'
						: returnRequest.Status === 'APPROVED'
							? 'rounded-full bg-green-500/20 px-3 py-1 text-xs font-semibold text-green-400'
							: 'rounded-full bg-red-500/20 px-3 py-1 text-xs font-semibold text-red-400'}
				>
					{returnRequest.Status}
				</span>
			{/if}
		</div>
	</div>

	{#if loading}
		<div class="flex h-64 items-center justify-center">
			<div
				class="h-8 w-8 animate-spin rounded-full border-2 border-slate-600 border-t-blue-500"
			></div>
		</div>
	{:else if returnRequest}
		<div in:fade={{ duration: 300 }} class="space-y-6">
			<!-- Top Section: Overview -->
			<div class="grid gap-6 md:grid-cols-2">
				<GlassCard title="Request Details">
					<div class="space-y-4">
						<div class="flex justify-between border-b border-white/5 pb-2">
							<span class="text-slate-400">Return ID</span>
							<span class="font-medium text-white">#{returnRequest.ID}</span>
						</div>
						<div class="flex justify-between border-b border-white/5 pb-2">
							<span class="text-slate-400">Date Requested</span>
							<span class="text-white">{formatDate(returnRequest.CreatedAt)}</span>
						</div>
						<div class="flex justify-between border-b border-white/5 pb-2">
							<span class="text-slate-400">Refund Amount</span>
							<span class="font-bold text-emerald-400"
								>{formatCurrency(returnRequest.RefundAmount)}</span
							>
						</div>
						<div class="pt-2">
							<span class="mb-1 block text-xs text-slate-500">Reason for Return</span>
							<div class="rounded-lg bg-white/5 p-3 text-sm text-slate-300">
								"{returnRequest.Reason}"
							</div>
						</div>
					</div>
				</GlassCard>

				<GlassCard title="Order & Customer">
					<div class="space-y-4">
						<div class="flex justify-between border-b border-white/5 pb-2">
							<span class="text-slate-400">Order Number</span>
							<span class="font-medium text-blue-400">{returnRequest.Order?.OrderNumber}</span>
						</div>
						<div class="flex justify-between border-b border-white/5 pb-2">
							<span class="text-slate-400">Customer</span>
							<span class="text-white">
								{returnRequest.User?.FirstName}
								{returnRequest.User?.LastName}
							</span>
						</div>
						<div class="flex justify-between border-b border-white/5 pb-2">
							<span class="text-slate-400">Contact</span>
							<span class="text-white">{returnRequest.User?.Email}</span>
						</div>
					</div>
				</GlassCard>
			</div>

			<!-- Items List -->
			<GlassCard title="Items to Return">
				<div class="space-y-4">
					<div
						class="hidden grid-cols-12 gap-4 border-b border-white/10 pb-2 text-xs font-semibold uppercase tracking-wider text-slate-500 md:grid"
					>
						<div class="col-span-6">Product</div>
						<div class="col-span-2 text-center">Condition</div>
						<div class="col-span-2 text-center">Quantity</div>
						<div class="col-span-2 text-right">Refund Value</div>
					</div>

					{#each returnRequest.ReturnItems as item}
						<div
							class="grid grid-cols-1 items-center gap-4 border-b border-white/5 py-4 last:border-0 md:grid-cols-12"
						>
							<div class="col-span-6 flex items-center gap-4">
								<div class="h-12 w-12 flex-shrink-0 overflow-hidden rounded-lg bg-slate-800">
									{#if item.Product?.ImageURLs}
										<img
											src={item.Product.ImageURLs.split(',')[0]}
											alt={item.Product.Name}
											class="h-full w-full object-cover"
										/>
									{:else}
										<div
											class="flex h-full w-full items-center justify-center text-xs text-slate-600"
										>
											No Img
										</div>
									{/if}
								</div>
								<div>
									<div class="font-medium text-white">{item.Product?.Name}</div>
									<div class="text-xs text-slate-400">SKU: {item.Product?.SKU}</div>
									<div class="mt-1 text-xs italic text-slate-500">"{item.Reason}"</div>
								</div>
							</div>
							<div class="col-span-2 flex justify-between md:justify-center">
								<span class="text-xs text-slate-500 md:hidden">Condition:</span>
								<span
									class="rounded-full border border-slate-600/50 bg-slate-700/50 px-2 py-0.5 text-xs text-slate-300"
									>{item.Condition}</span
								>
							</div>
							<div class="col-span-2 flex justify-between text-slate-300 md:justify-center">
								<span class="text-xs text-slate-500 md:hidden">Qty:</span>
								x{item.Quantity}
							</div>
							<div class="col-span-2 flex justify-between font-medium text-white md:justify-end">
								<span class="text-xs text-slate-500 md:hidden">Value:</span>
								{formatCurrency(item.Product?.SellingPrice * item.Quantity)}
							</div>
						</div>
					{/each}
				</div>
			</GlassCard>

			<!-- Actions -->
			{#if returnRequest.Status === 'PENDING'}
				<div
					class="sticky bottom-6 z-20 flex justify-end gap-3 rounded-2xl border border-white/10 bg-slate-900/80 p-4 shadow-2xl backdrop-blur-xl"
				>
					<button
						disabled={processing}
						onclick={() => handleProcess('reject')}
						class="rounded-xl border border-red-500/30 bg-red-500/10 px-6 py-2.5 font-medium text-red-400 transition-all hover:bg-red-500/20 active:scale-95 disabled:opacity-50"
					>
						Reject Request
					</button>
					<button
						disabled={processing}
						onclick={() => handleProcess('approve')}
						class="rounded-xl border border-emerald-500/30 bg-emerald-500/20 px-8 py-2.5 font-medium text-emerald-400 transition-all hover:bg-emerald-500/30 hover:shadow-[0_0_20px_rgba(16,185,129,0.3)] active:scale-95 disabled:opacity-50"
					>
						{processing ? 'Processing...' : 'Approve & Refund'}
					</button>
				</div>
			{/if}
		</div>
	{:else}
		<div class="rounded-xl border border-dashed border-slate-700 p-12 text-center text-slate-500">
			Return request not found.
		</div>
	{/if}
</div>
