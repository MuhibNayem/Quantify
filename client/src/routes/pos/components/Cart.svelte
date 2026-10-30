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

<div class="panel overflow-hidden" data-animate="fade-up" style="animation-delay:200ms">
	<div class="panel-header">
		<div>
			<h3 class="section-title">{$t('pos.cart.title')}</h3>
			<p class="mt-0.5 text-xs text-muted-foreground">
				{cart.length === 0
					? $t('pos.cart.empty_desc')
					: $t('pos.cart.items_desc', { count: cart.length, s: cart.length > 1 ? 's' : '' })}
			</p>
		</div>

		{#if cart.length > 0}
			<Button
				variant="ghost"
				size="sm"
				class="h-8 rounded-lg px-2.5 text-xs text-muted-foreground hover:bg-danger/10 hover:text-danger"
				onclick={onClear}
			>
				{$t('pos.cart.clear_btn')}
			</Button>
		{/if}
	</div>

	<div class="custom-scrollbar max-h-[20rem] overflow-auto">
		{#if cart.length === 0}
			<div class="px-4 py-10 text-center text-sm text-muted-foreground">
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
									class="inline-flex h-6 w-6 items-center justify-center rounded-md text-muted-foreground transition-colors hover:bg-danger/10 hover:text-danger"
									onclick={() => onRemove(item.id)}
									aria-label="Remove item"
								>
									<X class="h-3 w-3" />
								</button>
							</td>
							<td class="min-w-0 align-top whitespace-normal">
								<div
									class="line-clamp-2 text-[0.8125rem] font-medium break-words hyphens-auto text-foreground"
									style="display: -webkit-box; -webkit-line-clamp: 8; -webkit-box-orient: vertical; overflow: hidden; word-break: break-word;"
								>
									{item.Name}
								</div>

								<div class="mt-0.5 text-[0.7rem] break-words text-muted-foreground">
									<span class="num-mono">#{item.id}</span>
									{#if discountName}
										<span class="ml-1 text-[0.65rem] font-medium text-success"
											>({discountName})</span
										>
									{/if}
								</div>
							</td>

							<td
								class="num-mono min-w-[8rem] align-top text-[0.8125rem] whitespace-nowrap text-foreground"
							>
								{#if item.unitId}
									<span class="font-semibold">{formatCurrency(unitPrice)}</span>
									{#if item.unitSymbol}
										<span class="text-[0.65rem] text-muted-foreground">/ {item.unitSymbol}</span>
									{/if}
								{:else if price < originalPrice}
									<div class="flex flex-col">
										<span class="text-[0.65rem] text-muted-foreground line-through"
											>{formatCurrency(originalPrice)}</span
										>
										<span class="font-semibold text-primary">{formatCurrency(price)}</span>
									</div>
								{:else}
									{formatCurrency(price)}
								{/if}

								{#if item.units && item.units.length > 0}
									<select
										class="glass-input mt-1 h-7 w-full min-w-[7.5rem] cursor-pointer rounded-md px-1.5 text-[0.7rem]"
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
										class="inline-flex h-8 w-7 shrink-0 items-center justify-center rounded-md border border-border text-muted-foreground transition-colors hover:bg-muted hover:text-foreground"
										onclick={() => onUpdateQuantity(item.id, item.quantity - 1)}
										aria-label="Decrease quantity"
									>
										<Minus class="h-3 w-3" />
									</button>
									<input
										type="number"
										class="glass-input num-mono h-8 w-full min-w-[3.5rem] rounded-md px-1 text-center text-[0.8125rem]"
										min="0.01"
										step="0.01"
										value={item.quantity}
										onchange={(e) => onUpdateQuantity(item.id, parseFloat(e.currentTarget.value))}
									/>
									<button
										type="button"
										class="inline-flex h-8 w-7 shrink-0 items-center justify-center rounded-md border border-border text-muted-foreground transition-colors hover:bg-muted hover:text-foreground"
										onclick={() => onUpdateQuantity(item.id, item.quantity + 1)}
										aria-label="Increase quantity"
									>
										<Plus class="h-3 w-3" />
									</button>
								</div>
							</td>

							<td
								class="col-num min-w-[10rem] align-top font-semibold whitespace-nowrap text-foreground"
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
