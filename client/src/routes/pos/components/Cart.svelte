<script lang="ts">
	import { Button } from '$lib/components/ui/button';
	import { X, Minus, Plus } from 'lucide-svelte';
	import { formatCurrency } from '$lib/utils';
	import { t } from '$lib/i18n';

	type Props = {
		cart: any[];
		getProductPrice: (
			product: any
		) => { price: number; originalPrice: number; discountName?: string };
		getLinePrice: (item: any) => number;
		onRemove: (productId: number) => void;
		onUpdateQuantity: (productId: number, quantity: number) => void;
		onUpdateUnit: (productId: number, unitId: number | undefined) => void;
		onClear: () => void;
	};

	let {
		cart,
		getProductPrice,
		getLinePrice,
		onRemove,
		onUpdateQuantity,
		onUpdateUnit,
		onClear
	}: Props = $props();
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
					{@const unitPrice = getLinePrice(item)}
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
							{#if item.unitId}
								<span class="font-bold">{formatCurrency(unitPrice)}</span>
								{#if item.unitSymbol}
									<span class="text-muted-foreground text-[0.65rem]">/ {item.unitSymbol}</span>
								{/if}
							{:else if price < originalPrice}
								<div class="flex flex-col">
									<span class="text-muted-foreground text-[0.65rem] line-through"
										>{formatCurrency(originalPrice)}</span
									>
									<span class="text-danger font-bold">{formatCurrency(price)}</span>
								</div>
							{:else}
								{formatCurrency(price)}
							{/if}

							{#if item.units && item.units.length > 0}
								<select
									class="liquid-input mt-1 h-7 w-full min-w-[7.5rem] cursor-pointer rounded-md px-1.5 text-[0.7rem]"
									value={item.unitId ?? ''}
									onchange={(e) => {
										const v = e.currentTarget.value;
										onUpdateUnit(item.id, v ? Number(v) : undefined);
									}}
								>
									<option value="">{$t('pos.cart.base_unit')}</option>
									{#each item.units as u}
										<option value={u.UnitID}>
											{u.Unit?.Name ?? u.Unit?.Symbol ?? u.UnitID}
											{#if u.Price > 0}· {formatCurrency(u.Price)}{/if}
										</option>
									{/each}
								</select>
							{/if}
						</td>

						<!-- Quantity input -->
						<td class="min-w-[6.5rem] align-top">
							<div class="flex items-center gap-1">
								<button
									type="button"
									class="text-muted-foreground hover:text-foreground hover:bg-muted inline-flex h-8 w-7 shrink-0 items-center justify-center rounded-md border border-border transition-colors"
									onclick={() => onUpdateQuantity(item.id, item.quantity - 1)}
									aria-label="Decrease quantity"
								>
									<Minus class="h-3 w-3" />
								</button>
								<input
									type="number"
									class="liquid-input num-mono h-8 w-full min-w-[3.5rem] rounded-md px-1 text-center text-[0.8rem]"
									min="0.01"
									step="0.01"
									value={item.quantity}
									onchange={(e) => onUpdateQuantity(item.id, parseFloat(e.currentTarget.value))}
								/>
								<button
									type="button"
									class="text-muted-foreground hover:text-foreground hover:bg-muted inline-flex h-8 w-7 shrink-0 items-center justify-center rounded-md border border-border transition-colors"
									onclick={() => onUpdateQuantity(item.id, item.quantity + 1)}
									aria-label="Increase quantity"
								>
									<Plus class="h-3 w-3" />
								</button>
							</div>
						</td>

						<td
							class="num-mono text-foreground min-w-[12rem] text-right align-top text-[0.8rem] font-semibold whitespace-nowrap"
						>
							{formatCurrency(unitPrice * item.quantity)}
						</td>
					</tr>
				{/each}
				</tbody>
			</table>
		{/if}
	</div>
</div>
