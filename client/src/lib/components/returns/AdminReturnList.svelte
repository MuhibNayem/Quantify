<script lang="ts">
	import { onMount } from 'svelte';
	import { returnsApi } from '$lib/api/returns';
	import { Button } from '$lib/components/ui/button';
	import { toast } from 'svelte-sonner';
	import { CheckCircle2, XCircle, Eye } from 'lucide-svelte';
	import * as Dialog from '$lib/components/ui/dialog';
	import { Textarea } from '$lib/components/ui/textarea';
	import { returnNumber } from '$lib/utils';
	import TimeAgo from '$lib/components/ui/time-ago.svelte';
	import EmptyState from '$lib/components/ui/empty-state.svelte';

	let returns: any[] = [];
	let selectedReturn: any = null;
	let isDetailsOpen = false;
	let processingId: number | null = null;
	let rejectReason = '';
	let isRejectModalOpen = false;

	const textareaClass = 'rounded-md text-sm';

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

<div class="grid gap-3">
	{#each returns as ret}
		<div
			class="panel flex flex-col gap-4 p-5 transition-colors duration-200 hover:border-primary/30 md:flex-row md:items-center md:justify-between"
		>
			<div class="min-w-0 space-y-1">
				<div class="flex flex-wrap items-center gap-3">
					<h3 class="num-mono text-base font-bold text-foreground">{returnNumber(ret.ID)}</h3>
					<span class="badge-warning">{ret.Status}</span>
				</div>
				<p class="text-sm text-muted-foreground">
					Requested by User <span class="num-mono">#{ret.UserID}</span> for Order
					<span class="num-mono">#{ret.OrderID}</span>
				</p>
				<p class="text-xs text-muted-foreground">
					<TimeAgo date={ret.CreatedAt} />
				</p>
			</div>

			<div class="flex flex-wrap items-center gap-2">
				<Button variant="outline" size="sm" onclick={() => openDetails(ret)}>
					<Eye size={16} class="mr-2" /> Details
				</Button>
				<Button
					variant="default"
					size="sm"
					class="btn-soothe"
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
		<div class="rounded-lg border border-dashed border-border/70 bg-muted/30">
			<EmptyState
				icon={CheckCircle2}
				title="All caught up!"
				description="No pending returns to process."
			/>
		</div>
	{/each}
</div>

<!-- Details Modal -->
<Dialog.Root bind:open={isDetailsOpen}>
	<Dialog.Content class="sm:max-w-[600px]">
		<Dialog.Header>
			<Dialog.Title
				>Return Details {selectedReturn ? returnNumber(selectedReturn.ID) : ''}</Dialog.Title
			>
		</Dialog.Header>

		{#if selectedReturn}
			<div class="space-y-4">
				<div class="rounded-lg bg-muted/50 p-4">
					<p class="overline">Refund Amount</p>
					<p class="num-mono mt-1 text-2xl font-bold text-foreground">
						${selectedReturn.RefundAmount.toFixed(2)}
					</p>
				</div>

				<div class="space-y-2">
					<h4 class="section-title">Items</h4>
					{#each selectedReturn.Items as item}
						<div
							class="flex items-start justify-between gap-3 rounded-md border border-border/70 p-3"
						>
							<div class="min-w-0">
								<p class="font-medium text-foreground">
									Item <span class="num-mono">#{item.OrderItemID}</span>
								</p>
								<p class="text-sm text-muted-foreground">
									Qty: <span class="num-mono">{item.Quantity}</span> • Reason: {item.Reason}
								</p>
							</div>
							<span class="badge-neutral">{item.Condition}</span>
						</div>
					{/each}
				</div>
			</div>

			<Dialog.Footer>
				<Button
					variant="default"
					class="btn-soothe"
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

		<Textarea
			bind:value={rejectReason}
			placeholder="Reason for rejection..."
			class={textareaClass}
		/>

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
