<script lang="ts">
	import { onMount } from 'svelte';
	import { returnsApi } from '$lib/api/returns';
	import { Button } from '$lib/components/ui/button';
	import { Badge } from '$lib/components/ui/badge';
	import { toast } from 'svelte-sonner';
	import { CheckCircle2, XCircle, Eye } from 'lucide-svelte';
	import * as Dialog from '$lib/components/ui/dialog';
	import { Textarea } from '$lib/components/ui/textarea';

	let returns: any[] = [];
	let selectedReturn: any = null;
	let isDetailsOpen = false;
	let processingId: number | null = null;
	let rejectReason = '';
	let isRejectModalOpen = false;

	onMount(() => {
		loadReturns();
	});

	async function loadReturns() {
		try {
			const data = await returnsApi.listReturns('PENDING');
			returns = data.returns || [];
		} catch (e) {
			console.error(e);
			toast.error('Failed to load returns');
		}
	}

	async function handleProcess(returnId: number, action: 'approve' | 'reject', notes?: string) {
		processingId = returnId;
		try {
			await returnsApi.processReturn(returnId, action, notes);
			toast.success(`Return ${action}ed successfully`);
			loadReturns();
			if (isRejectModalOpen) isRejectModalOpen = false;
			if (isDetailsOpen) isDetailsOpen = false;
		} catch (e) {
			console.error(e);
			toast.error(`Failed to ${action} return`);
		} finally {
			processingId = null;
		}
	}

	function openDetails(ret: any) {
		selectedReturn = ret;
		isDetailsOpen = true;
	}

	function openRejectModal(ret: any) {
		selectedReturn = ret;
		rejectReason = '';
		isRejectModalOpen = true;
	}
</script>

<div class="grid gap-4">
	{#each returns as ret}
		<div
			class="flex flex-col gap-4 rounded-3xl border border-white/60 bg-white/60 p-6 shadow-xl backdrop-blur-2xl transition-all hover:bg-white/80 md:flex-row md:items-center md:justify-between"
		>
			<div>
				<div class="flex items-center gap-3">
					<h3 class="text-lg font-bold text-slate-800">Return #{ret.ID}</h3>
					<Badge variant="secondary">{ret.Status}</Badge>
				</div>
				<p class="text-sm text-slate-500">
					Requested by User #{ret.UserID} for Order #{ret.OrderID}
				</p>
				<p class="text-xs text-slate-400">
					{new Date(ret.CreatedAt).toLocaleDateString()}
				</p>
			</div>

			<div class="flex items-center gap-2">
				<Button variant="outline" size="sm" onclick={() => openDetails(ret)}>
					<Eye size={16} class="mr-2" /> Details
				</Button>
				<Button
					variant="default"
					size="sm"
					class="bg-green-600 hover:bg-green-700 text-white"
					disabled={processingId === ret.ID}
					onclick={() => handleProcess(ret.ID, 'approve')}
				>
					<CheckCircle2 size={16} class="mr-2" /> Approve
				</Button>
				<Button
					variant="destructive"
					size="sm"
					disabled={processingId === ret.ID}
					onclick={() => openRejectModal(ret)}
				>
					<XCircle size={16} class="mr-2" /> Reject
				</Button>
			</div>
		</div>
	{:else}
		<div
			class="flex h-64 flex-col items-center justify-center rounded-3xl border border-dashed border-slate-300 bg-white/40 text-center"
		>
			<CheckCircle2 size={48} class="mb-4 text-slate-300" />
			<p class="text-lg font-medium text-slate-600">All caught up!</p>
			<p class="text-slate-400">No pending returns to process.</p>
		</div>
	{/each}
</div>

<!-- Details Modal -->
<Dialog.Root bind:open={isDetailsOpen}>
	<Dialog.Content class="sm:max-w-[600px]">
		<Dialog.Header>
			<Dialog.Title>Return Details #{selectedReturn?.ID}</Dialog.Title>
		</Dialog.Header>

		{#if selectedReturn}
			<div class="space-y-4">
				<div class="rounded-xl bg-slate-50 p-4">
					<p class="text-sm font-medium text-slate-500">Refund Amount</p>
					<p class="text-2xl font-bold text-slate-900">${selectedReturn.RefundAmount.toFixed(2)}</p>
				</div>

				<div class="space-y-2">
					<h4 class="font-medium text-slate-900">Items</h4>
					{#each selectedReturn.Items as item}
						<div class="flex items-start justify-between rounded-lg border border-slate-100 p-3">
							<div>
								<p class="font-medium text-slate-800">Item #{item.OrderItemID}</p>
								<p class="text-sm text-slate-500">Qty: {item.Quantity} • Reason: {item.Reason}</p>
							</div>
							<Badge variant="outline">{item.Condition}</Badge>
						</div>
					{/each}
				</div>
			</div>

			<Dialog.Footer>
				<Button
					variant="default"
					class="bg-green-600 hover:bg-green-700 text-white"
					onclick={() => handleProcess(selectedReturn.ID, 'approve')}
				>
					Approve Return
				</Button>
			</Dialog.Footer>
		{/if}
	</Dialog.Content>
</Dialog.Root>

<!-- Reject Modal -->
<Dialog.Root bind:open={isRejectModalOpen}>
	<Dialog.Content>
		<Dialog.Header>
			<Dialog.Title>Reject Return</Dialog.Title>
			<Dialog.Description>Please provide a reason for rejection.</Dialog.Description>
		</Dialog.Header>

		<Textarea bind:value={rejectReason} placeholder="Reason for rejection..." />

		<Dialog.Footer>
			<Button variant="outline" onclick={() => (isRejectModalOpen = false)}>Cancel</Button>
			<Button
				variant="destructive"
				onclick={() => handleProcess(selectedReturn.ID, 'reject', rejectReason)}
				disabled={!rejectReason}
			>
				Reject Return
			</Button>
		</Dialog.Footer>
	</Dialog.Content>
</Dialog.Root>
