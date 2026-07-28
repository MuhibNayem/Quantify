<script lang="ts">
	import {
		Card,
		CardContent,
		CardDescription,
		CardHeader,
		CardTitle
	} from '$lib/components/ui/card';
	import { Badge } from '$lib/components/ui/badge';
	import { formatCurrency } from '$lib/utils';
	import { t } from '$lib/i18n';

	type Props = {
		products: any[];
		fetchStatus: string;
		fetchLimit: number;
		getProductPrice: (
			product: any
		) => { price: number; originalPrice: number; discountName?: string };
		getAvailableStock: (product: any) => number;
		onAddToCart: (product: any) => void;
		onFilterChange: () => void;
	};

	let {
		products,
		fetchStatus = $bindable(),
		fetchLimit = $bindable(),
		getProductPrice,
		getAvailableStock,
		onAddToCart,
		onFilterChange
	}: Props = $props();
</script>

<Card
	class="flex-1 overflow-hidden rounded-2xl border-0 bg-gradient-to-br from-indigo-50 to-sky-100 shadow-lg transition-all duration-300 hover:scale-[1.01] hover:shadow-xl"
>
	<CardHeader class="border-b border-white/60 bg-white/70 pb-3 backdrop-blur">
		<div class="flex flex-col gap-3">
			<div>
				<CardTitle class="text-sm text-slate-800">{$t('pos.products.title')}</CardTitle>
				<CardDescription class="text-[0.75rem] text-slate-500">
					{$t('pos.products.description')}
				</CardDescription>
			</div>
			<div class="flex flex-wrap items-center gap-2">
				<!-- Status Filter -->
				<select
					bind:value={fetchStatus}
					onchange={onFilterChange}
					class="h-9 flex-1 cursor-pointer rounded-lg border-slate-200 bg-white/80 pr-2 pl-2 text-xs font-medium text-slate-600 shadow-sm outline-none hover:bg-white focus:border-indigo-300 focus:ring-indigo-300 sm:flex-none"
					aria-label={$t('pos.products.filter_status.label')}
				>
					<option value="">{$t('pos.products.filter_status.all')}</option>
					<option value="IN_STOCK">{$t('pos.products.filter_status.in_stock')}</option>
					<option value="LOW_STOCK">{$t('pos.products.filter_status.low_stock')}</option>
					<option value="OUT_OF_STOCK">{$t('pos.products.filter_status.out_of_stock')}</option>
				</select>

				<div class="flex flex-1 items-center gap-2 sm:flex-none">
					<select
						bind:value={fetchLimit}
						onchange={onFilterChange}
						class="h-9 w-full cursor-pointer rounded-lg border-slate-200 bg-white/80 pr-2 pl-2 text-xs font-medium text-slate-600 shadow-sm outline-none hover:bg-white focus:border-indigo-300 focus:ring-indigo-300 sm:w-auto"
					>
						<option value={100}>100 items</option>
						<option value={200}>200 items</option>
						<option value={500}>500 items</option>
						<option value={1000}>1,000 items</option>
					</select>
					<div class="text-[0.7rem] whitespace-nowrap text-slate-500">
						<span class="font-medium text-slate-700">{products.length}</span> results
					</div>
				</div>
			</div>
		</div>
	</CardHeader>
	<CardContent class="max-h-[22rem] overflow-y-auto p-3 pt-3">
		{#if products.length === 0}
			<div class="flex h-full items-center justify-center text-[0.8rem] text-slate-500">
				{$t('pos.products.no_results')}
			</div>
		{:else}
			<div class="grid grid-cols-2 gap-3 xl:grid-cols-3">
				{#each products as product}
					{@const { price, originalPrice, discountName } = getProductPrice(product)}
					<button
						type="button"
						onclick={() => onAddToCart(product)}
						class="group flex flex-col justify-between rounded-xl border border-indigo-50 bg-white/80 p-3 text-left shadow-sm transition-all duration-200 hover:border-indigo-200 hover:bg-indigo-50/80 hover:shadow-md"
					>
						<div class="space-y-1.5">
							<div class="line-clamp-2 text-[0.8rem] font-medium text-slate-900">
								{product.Name}
							</div>
							<div class="mt-0.5 flex flex-col gap-1">
								<div class="flex flex-wrap items-center gap-2">
									{#if price < originalPrice}
										<span class="text-[0.65rem] text-slate-400 line-through">
											{formatCurrency(originalPrice)}
										</span>
										<span class="text-[0.8rem] font-bold text-rose-600">
											{formatCurrency(price)}
										</span>
									{:else}
										<span class="text-[0.8rem] font-semibold text-indigo-600">
											{formatCurrency(price)}
										</span>
									{/if}
								</div>

								{#if discountName}
									<Badge
										variant="outline"
										class="w-fit border-rose-200 bg-rose-50 px-1 py-0 text-[0.6rem] text-rose-600"
									>
										{discountName}
									</Badge>
								{/if}

								<div class="flex items-center gap-1 text-[0.7rem] text-slate-500">
									<span
										class="h-1.5 w-1.5 rounded-full {getAvailableStock(product) > 0
											? 'bg-emerald-500'
											: 'bg-rose-400'}"
									/>
									<span>
										{$t('pos.products.in_stock', { count: getAvailableStock(product) })}
									</span>
								</div>
							</div>
						</div>
						<div class="mt-3 flex items-center justify-between text-[0.7rem] text-slate-400">
							<span class="flex items-center gap-1 font-medium group-hover:text-indigo-600">
								<span
									class="rounded-full border border-indigo-100 bg-indigo-50/90 px-1.5 py-0.5 text-[0.65rem] text-indigo-600"
								>
									Tap
								</span>
								{$t('pos.products.tap_to_add')}
							</span>
							<span class="text-[0.65rem] tracking-wide text-slate-400 uppercase"
								>#{product.ID}</span
							>
						</div>
					</button>
				{/each}
			</div>
		{/if}
	</CardContent>
</Card>
