<script lang="ts">
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import {
		Table,
		TableBody,
		TableCell,
		TableHead,
		TableHeader,
		TableRow
	} from '$lib/components/ui/table';
	import {
		Card,
		CardContent,
		CardDescription,
		CardHeader,
		CardTitle
	} from '$lib/components/ui/card';
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

<Card
	class="overflow-hidden rounded-2xl border-0 bg-white/90 shadow-lg backdrop-blur-xl transition-all duration-300 hover:scale-[1.01] hover:shadow-xl"
	data-animate="fade-up"
	style="animation-delay:200ms"
>
	<CardHeader
		class="flex flex-row items-center justify-between border-b border-slate-100/80 bg-slate-50/70 pb-3"
	>
		<div>
			<CardTitle class="text-sm text-slate-900">{$t('pos.cart.title')}</CardTitle>
			<CardDescription class="text-[0.75rem] text-slate-500">
				{cart.length === 0
					? $t('pos.cart.empty_desc')
					: $t('pos.cart.items_desc', { count: cart.length, s: cart.length > 1 ? 's' : '' })}
			</CardDescription>
		</div>

		{#if cart.length > 0}
			<Button
				variant="ghost"
				size="sm"
				class="rounded-xl px-3 py-1.5 text-[0.75rem] text-slate-500 hover:bg-rose-50 hover:text-rose-500"
				onclick={onClear}
			>
				{$t('pos.cart.clear_btn')}
			</Button>
		{/if}
	</CardHeader>

	<CardContent class="max-h-[20rem] overflow-auto pt-0">
		{#if cart.length === 0}
			<div class="py-6 text-center text-[0.8rem] text-slate-400">
				{$t('pos.cart.empty_state')}
			</div>
		{:else}
			<Table class="w-full min-w-[30rem]">
				<TableHeader>
					<TableRow class="border-slate-100 bg-slate-50/80">
						<TableHead />
						<TableHead class="w-1/3 min-w-[8rem] text-[0.7rem] text-slate-500"
							>{$t('pos.cart.table.product')}</TableHead
						>
						<TableHead class="text-[0.7rem] text-slate-500">{$t('pos.cart.table.price')}</TableHead>
						<TableHead class="min-w-[5rem] text-[0.7rem] text-slate-500"
							>{$t('pos.cart.table.qty')}</TableHead
						>
						<TableHead class="text-right text-[0.7rem] text-slate-500"
							>{$t('pos.cart.table.total')}</TableHead
						>
					</TableRow>
				</TableHeader>

				<TableBody>
					{#each cart as item}
						{@const { price, originalPrice, discountName } = getProductPrice(item)}
						<TableRow class="border-slate-100 hover:bg-slate-50/60">
							<TableCell class="text-right align-top">
								<Button
									variant="ghost"
									size="icon"
									class="h-7 w-7 rounded-full text-slate-400 hover:bg-rose-50 hover:text-rose-500"
									onclick={() => onRemove(item.id)}
								>
									<X class="h-3 w-3" />
								</Button>
							</TableCell>
							<TableCell class="min-w-0 align-top whitespace-normal">
								<div
									class="line-clamp-2 text-[0.8rem] font-medium break-words hyphens-auto text-slate-900"
									style="display: -webkit-box; -webkit-line-clamp: 8; -webkit-box-orient: vertical; overflow: hidden; word-break: break-word;"
								>
									{item.Name}
								</div>

								<div class="mt-0.5 text-[0.7rem] break-words text-slate-400">
									#{item.id}
									{#if discountName}
										<span class="ml-1 text-[0.65rem] font-medium text-rose-500"
											>({discountName})</span
										>
									{/if}
								</div>
							</TableCell>

							<TableCell
								class="min-w-[8rem] align-top text-[0.8rem] whitespace-nowrap text-slate-800"
							>
								{#if price < originalPrice}
									<div class="flex flex-col">
										<span class="text-[0.65rem] text-slate-400 line-through"
											>{formatCurrency(originalPrice)}</span
										>
										<span class="font-bold text-rose-600">{formatCurrency(price)}</span>
									</div>
								{:else}
									{formatCurrency(price)}
								{/if}
							</TableCell>

							<!-- Quantity input -->
							<TableCell class="min-w-[5rem] align-top">
								<Input
									type="number"
									class="h-8 w-full min-w-[5rem] rounded-lg border-slate-200 bg-slate-50/80 px-2 text-[0.8rem]"
									min="1"
									value={item.quantity}
									onchange={(e) => onUpdateQuantity(item.id, parseInt(e.currentTarget.value))}
								/>
							</TableCell>

							<TableCell
								class="min-w-[12rem] text-right align-top text-[0.8rem] font-semibold whitespace-nowrap text-slate-900"
							>
								{formatCurrency(price * item.quantity)}
							</TableCell>
						</TableRow>
					{/each}
				</TableBody>
			</Table>
		{/if}
	</CardContent>
</Card>
