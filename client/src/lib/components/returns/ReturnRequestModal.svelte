<script lang="ts">
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import { Textarea } from '$lib/components/ui/textarea';
	import * as Dialog from '$lib/components/ui/dialog';
	import { Checkbox } from '$lib/components/ui/checkbox';
	import { Select } from '$lib/components/ui/select';
	import { returnsApi } from '$lib/api/returns';
	import { toast } from 'svelte-sonner';
	import { createEventDispatcher } from 'svelte';

	export let open = false;
	export let order: any;

	const dispatch = createEventDispatcher();

	let selectedItems: Record<number, boolean> = {};
	let quantities: Record<number, number> = {};
	let reasons: Record<number, string> = {};
	let conditions: Record<number, string> = {};
	let isSubmitting = false;

	const reasonOptions = [
		{ value: 'DAMAGED', label: 'Damaged / Defective' },
		{ value: 'WRONG_ITEM', label: 'Wrong Item Sent' },
		{ value: 'DONT_WANT', label: 'Changed Mind' },
		{ value: 'SIZE_ISSUE', label: 'Size / Fit Issue' }
	];

	const conditionOptions = [
		{ value: 'UNOPENED', label: 'Unopened' },
		{ value: 'OPENED', label: 'Opened' },
		{ value: 'DAMAGED', label: 'Damaged' }
	];

	function toggleItem(itemId: number) {
		selectedItems[itemId] = !selectedItems[itemId];
		if (selectedItems[itemId]) {
			quantities[itemId] = 1; // Default to 1
			reasons[itemId] = 'DONT_WANT';
			conditions[itemId] = 'UNOPENED';
		}
	}

	async function submitReturn() {
		const itemsToReturn = Object.keys(selectedItems)
			.filter((id) => selectedItems[Number(id)])
			.map((id) => ({
				order_item_id: Number(id),
				quantity: quantities[Number(id)],
				reason: reasons[Number(id)],
				condition: conditions[Number(id)]
			}));

		if (itemsToReturn.length === 0) {
			toast.error('Please select at least one item to return');
			return;
		}

		isSubmitting = true;
		try {
			await returnsApi.requestReturn(order.ID, itemsToReturn);
			dispatch('submit');
			open = false;
		} catch (e) {
			console.error(e);
			toast.error('Failed to submit return request');
		} finally {
			isSubmitting = false;
		}
	}
</script>

<Dialog.Root bind:open>
	<Dialog.Content class="sm:max-w-[600px]">
		<Dialog.Header>
			<Dialog.Title>Request Return for Order #{order.OrderNumber}</Dialog.Title>
			<Dialog.Description>
				Select the items you wish to return and specify the reason.
			</Dialog.Description>
		</Dialog.Header>

		<div class="grid gap-4 py-4 max-h-[60vh] overflow-y-auto pr-2">
			{#each order.Items as item}
				<div
					class="flex flex-col gap-3 rounded-xl border border-slate-200 p-4 transition-colors hover:bg-slate-50"
				>
					<div class="flex items-start gap-3">
						<Checkbox
							checked={selectedItems[item.ID]}
							onCheckedChange={() => toggleItem(item.ID)}
						/>
						<div class="flex-1">
							<p class="font-medium text-slate-900">{item.Product?.Name || 'Unknown Product'}</p>
							<p class="text-sm text-slate-500">Qty: {item.Quantity} • ${item.UnitPrice}</p>
						</div>
					</div>

					{#if selectedItems[item.ID]}
						<div class="ml-8 grid gap-3 pl-2 border-l-2 border-slate-100">
							<div class="grid grid-cols-2 gap-3">
								<div class="space-y-1">
									<Label class="text-xs">Quantity</Label>
									<Input
										type="number"
										min="1"
										max={item.Quantity}
										bind:value={quantities[item.ID]}
										class="h-8"
									/>
								</div>
								<div class="space-y-1">
									<Label class="text-xs">Condition</Label>
									<Select
										options={conditionOptions}
										bind:value={conditions[item.ID]}
										style="h-8 text-sm"
									/>
								</div>
							</div>
							<div class="space-y-1">
								<Label class="text-xs">Reason</Label>
								<Select options={reasonOptions} bind:value={reasons[item.ID]} style="h-8 text-sm" />
							</div>
						</div>
					{/if}
				</div>
			{/each}
		</div>

		<Dialog.Footer>
			<Button variant="outline" onclick={() => (open = false)}>Cancel</Button>
			<Button onclick={submitReturn} disabled={isSubmitting}>
				{isSubmitting ? 'Submitting...' : 'Submit Request'}
			</Button>
		</Dialog.Footer>
	</Dialog.Content>
</Dialog.Root>
