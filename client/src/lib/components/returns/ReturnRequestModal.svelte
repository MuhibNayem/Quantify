<script lang="ts">
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
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

	const dialogClass =
		'rounded-lg border-border bg-card bg-none shadow-lg backdrop-blur-none backdrop-saturate-100 from-card to-card';
	const inputClass =
		'liquid-input h-8 w-full rounded-md bg-card bg-none px-3 text-sm font-normal text-foreground shadow-none ring-0 backdrop-blur-none backdrop-saturate-100 from-card to-card placeholder:text-muted-foreground focus:scale-100 focus:border-primary focus:ring-0 md:text-sm';
	const btnPrimary =
		'rounded-md bg-primary bg-none px-4 text-sm font-medium text-primary-foreground shadow-none hover:bg-primary/90 hover:bg-none';

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
	<Dialog.Content class="{dialogClass} sm:max-w-[600px]">
		<Dialog.Header>
			<Dialog.Title>Request Return for Order <span class="num-mono">#{order.OrderNumber}</span></Dialog.Title>
			<Dialog.Description>
				Select the items you wish to return and specify the reason.
			</Dialog.Description>
		</Dialog.Header>

		<div class="custom-scrollbar grid max-h-[60vh] gap-4 overflow-y-auto py-4 pr-2">
			{#each order.Items as item}
				<div
					class="border-border hover:bg-muted/60 flex flex-col gap-3 rounded-md border p-4 transition-colors"
				>
					<div class="flex items-start gap-3">
						<Checkbox
							checked={selectedItems[item.ID]}
							onCheckedChange={() => toggleItem(item.ID)}
						/>
						<div class="flex-1">
							<p class="text-foreground font-medium">{item.Product?.Name || 'Unknown Product'}</p>
							<p class="text-muted-foreground text-sm">Qty: <span class="num-mono">{item.Quantity}</span> • <span class="num-mono">${item.UnitPrice}</span></p>
						</div>
					</div>

					{#if selectedItems[item.ID]}
						<div class="border-border ml-8 grid gap-3 border-l-2 pl-2">
							<div class="grid grid-cols-2 gap-3">
								<div class="space-y-1">
									<Label class="text-muted-foreground text-xs">Quantity</Label>
									<Input
										type="number"
										min="1"
										max={item.Quantity}
										bind:value={quantities[item.ID]}
										class="{inputClass} num-mono"
									/>
								</div>
								<div class="space-y-1">
									<Label class="text-muted-foreground text-xs">Condition</Label>
									<Select
										options={conditionOptions}
										bind:value={conditions[item.ID]}
										style="h-8 text-sm"
									/>
								</div>
							</div>
							<div class="space-y-1">
								<Label class="text-muted-foreground text-xs">Reason</Label>
								<Select options={reasonOptions} bind:value={reasons[item.ID]} style="h-8 text-sm" />
							</div>
						</div>
					{/if}
				</div>
			{/each}
		</div>

		<Dialog.Footer>
			<Button variant="outline" onclick={() => (open = false)}>Cancel</Button>
			<Button onclick={submitReturn} disabled={isSubmitting} class={btnPrimary}>
				{isSubmitting ? 'Submitting...' : 'Submit Request'}
			</Button>
		</Dialog.Footer>
	</Dialog.Content>
</Dialog.Root>
