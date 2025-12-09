<script lang="ts">
	import { onMount } from 'svelte';
	import GlassCard from '$lib/components/ui/GlassCard.svelte';
	import { formatDate, formatCurrency } from '$lib/utils';
	import { fade } from 'svelte/transition';
	import { auth } from '$lib/stores/auth';
	import { goto } from '$app/navigation';
	import { toast } from 'svelte-sonner';

	$effect(() => {
		if (!auth.hasPermission('returns.manage')) {
			toast.error('Access Denied', {
				description: 'You do not have permission to manage returns.'
			});
			goto('/');
		}
	});

	let returns = $state<any[]>([]);
	let loading = $state(true);

	// Mock Data
	const mockReturns = [
		{
			id: 1,
			orderNumber: 'ORD-1733600000-1',
			user: 'john_doe',
			reason: 'Defective',
			amount: 100.0,
			status: 'PENDING',
			date: '2025-12-08T09:00:00Z',
			items: [{ name: 'Wireless Headphones', quantity: 1, condition: 'GOOD' }]
		},
		{
			id: 2,
			orderNumber: 'ORD-1733500000-1',
			user: 'jane_smith',
			reason: 'Changed Mind',
			amount: 45.5,
			status: 'APPROVED',
			date: '2025-12-07T14:30:00Z',
			items: [{ name: 'USB-C Cable', quantity: 1, condition: 'OPENED' }]
		}
	];

	onMount(async () => {
		// Simulate fetch
		setTimeout(() => {
			returns = mockReturns;
			loading = false;
		}, 500);
	});

	async function processReturn(id, action) {
		// Call API: await api.post(`/returns/${id}/process`, { action });
		console.log(`Processing return ${id}: ${action}`);

		// Optimistic update
		returns = returns.map((r) =>
			r.id === id ? { ...r, status: action === 'approve' ? 'APPROVED' : 'REJECTED' } : r
		);
	}
</script>

<div class="container mx-auto space-y-8 p-6">
	<div class="flex items-center justify-between">
		<h1
			class="bg-gradient-to-r from-purple-400 to-pink-600 bg-clip-text text-3xl font-bold text-transparent"
		>
			Return Requests
		</h1>
	</div>

	{#if loading}
		<div class="text-center text-gray-400">Loading requests...</div>
	{:else if returns.length === 0}
		<div class="text-center text-gray-400">No return requests found.</div>
	{:else}
		<div class="space-y-4">
			{#each returns as req (req.id)}
				<GlassCard class="transition-all hover:bg-white/15">
					<div class="flex flex-col justify-between gap-4 md:flex-row">
						<!-- Info -->
						<div class="flex-1 space-y-2">
							<div class="flex items-center gap-3">
								<span class="text-lg font-semibold text-white">#{req.id}</span>
								<span class="text-sm text-purple-400">{req.orderNumber}</span>
								<span
									class={req.status === 'PENDING'
										? 'rounded bg-yellow-400/10 px-2 py-0.5 text-xs text-yellow-400'
										: req.status === 'APPROVED'
											? 'rounded bg-green-400/10 px-2 py-0.5 text-xs text-green-400'
											: 'rounded bg-red-400/10 px-2 py-0.5 text-xs text-red-400'}
								>
									{req.status}
								</span>
							</div>

							<div class="text-sm text-gray-400">
								<span class="text-white">{req.user}</span> • {formatDate(req.date)}
							</div>

							<div class="rounded-lg border border-white/5 bg-black/20 p-3">
								<p class="mb-2 text-sm text-gray-300">
									<span class="text-gray-500">Reason:</span>
									{req.reason}
								</p>
								<div class="space-y-1">
									{#each req.items as item}
										<div class="flex justify-between text-xs text-gray-400">
											<span>{item.quantity}x {item.name}</span>
											<span class="text-gray-500">Condition: {item.condition}</span>
										</div>
									{/each}
								</div>
							</div>
						</div>

						<!-- Actions -->
						<div class="flex min-w-[150px] flex-col items-end justify-between gap-4">
							<span class="text-xl font-bold text-white">{formatCurrency(req.amount)}</span>

							{#if req.status === 'PENDING'}
								<div class="flex w-full gap-2">
									<button
										onclick={() => processReturn(req.id, 'reject')}
										class="flex-1 rounded border border-red-500/20 bg-red-500/10 px-3 py-2 text-sm font-medium text-red-400 transition-colors hover:bg-red-500/20"
									>
										Reject
									</button>
									<button
										onclick={() => processReturn(req.id, 'approve')}
										class="flex-1 rounded border border-green-500/20 bg-green-500/10 px-3 py-2 text-sm font-medium text-green-400 transition-colors hover:bg-green-500/20"
									>
										Approve
									</button>
								</div>
							{:else}
								<div class="text-sm italic text-gray-500">Processed</div>
							{/if}
						</div>
					</div>
				</GlassCard>
			{/each}
		</div>
	{/if}
</div>
