<script lang="ts">
	import { Button } from '$lib/components/ui/button';
	import { X } from 'lucide-svelte';
	import { formatCurrency } from '$lib/utils';
	import { t } from '$lib/i18n';

	type Props = {
		cart: any[];
		getProductPrice: (
			product: any
		) => { price: number; originalPrice: number; discountName?: string };
		onRemove: (productId: number) => void;
		onUpdateQuantity: (productId: number, quantity: number) => void;
		onClear: () => void;
	};

	let { cart, getProductPrice, onRemove, onUpdateQuantity, onClear }: Props = $props();
</script>

<div
	class="liquid-panel overflow-hidden rounded-lg"
	data-animate="fade-up"
	style="animation-delay:200ms"
>
	<div class="border-border bg-muted/50 flex flex-row items-center justify-between border-b px-4 py-3">
		<div>
			<h3 class="font-display text-foreground text-sm font-semibold">{$t('pos.cart.title')}</h3>
			<p class="text-muted-foreground text-[0.75rem]">
				{cart.length === 0
					? $t('pos.cart.empty_desc')
					: $t('pos.cart.items_desc', { count: cart.length, s: cart.length > 1 ? 's' : '' })}
			</p>
		</div>

		{#if cart.length > 0}
			<Button
				variant="ghost"
				size="sm"
				class="text-muted-foreground hover:text-danger hover:bg-danger/10 rounded-md px-3 py-1.5 text-[0.75rem]"
				onclick={onClear}
			>
				{$t('pos.cart.clear_btn')}
			</Button>
		{/if}
	</div>

	<div class="custom-scrollbar max-h-[20rem] overflow-auto">
		{#if cart.length === 0}
			<div class="text-muted-foreground py-8 text-center text-[0.8rem]">
				{$t('pos.cart.empty_state')}
			</div>
		{:else}
			<table class="ledger-table min-w-[30rem]">
				<thead>
					<tr>
						<th class="w-8"></th>
						<th class="w-1/3 min-w-[8rem]">{$t('pos.cart.table.product')}</th>
						<th>{$t('pos.cart.table.price')}</th>
						<th class="min-w-[5rem]">{$t('pos.cart.table.qty')}</th>
						<th class="text-right">{$t('pos.cart.table.total')}</th>
					</tr>
				</thead>

				<tbody>
					{#each cart as item}
						{@const { price, originalPrice, discountName } = getProductPrice(item)}
						<tr>
							<td class="text-right align-top">
								<button
									type="button"
									class="text-muted-foreground hover:text-danger hover:bg-danger/10 inline-flex h-6 w-6 items-center justify-center rounded-md transition-colors"
									onclick={() => onRemove(item.id)}
									aria-label="Remove item"
								>
									<X class="h-3 w-3" />
								</button>
							</td>
							<td class="min-w-0 align-top whitespace-normal">
								<div
									class="text-foreground line-clamp-2 text-[0.8rem] font-medium break-words hyphens-auto"
									style="display: -webkit-box; -webkit-line-clamp: 8; -webkit-box-orient: vertical; overflow: hidden; word-break: break-word;"
								>
									{item.Name}
								</div>

								<div class="text-muted-foreground mt-0.5 text-[0.7rem] break-words">
									<span class="num-mono">#{item.id}</span>
									{#if discountName}
										<span class="text-danger ml-1 text-[0.65rem] font-medium">({discountName})</span>
									{/if}
								</div>
							</td>

							<td class="num-mono text-foreground min-w-[8rem] align-top text-[0.8rem] whitespace-nowrap">
								{#if price < originalPrice}
									<div class="flex flex-col">
										<span class="text-muted-foreground text-[0.65rem] line-through"
											>{formatCurrency(originalPrice)}</span
										>
										<span class="text-danger font-bold">{formatCurrency(price)}</span>
									</div>
								{:else}
									{formatCurrency(price)}
								{/if}
							</td>

							<!-- Quantity input -->
							<td class="min-w-[5rem] align-top">
								<input
									type="number"
									class="liquid-input h-8 w-full min-w-[5rem] rounded-md px-2 text-center text-[0.8rem]"
									min="1"
									value={item.quantity}
									onchange={(e) => onUpdateQuantity(item.id, parseInt(e.currentTarget.value))}
								/>
							</td>

							<td
								class="num-mono text-foreground min-w-[12rem] text-right align-top text-[0.8rem] font-semibold whitespace-nowrap"
							>
								{formatCurrency(price * item.quantity)}
							</td>
						</tr>
					{/each}
				</tbody>
			</table>
		{/if}
	</div>
</div>
