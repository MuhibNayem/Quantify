<script lang="ts">
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

<div class="liquid-panel flex-1 overflow-hidden rounded-lg">
	<div class="border-border bg-muted/50 border-b p-4 pb-3">
		<div class="flex flex-col gap-3">
			<div>
				<h3 class="font-display text-foreground text-sm font-semibold">
					{$t('pos.products.title')}
				</h3>
				<p class="text-muted-foreground text-[0.75rem]">
					{$t('pos.products.description')}
				</p>
			</div>
			<div class="flex flex-wrap items-center gap-2">
				<!-- Status Filter -->
				<select
					bind:value={fetchStatus}
					onchange={onFilterChange}
					class="text-muted-foreground h-9 flex-1 cursor-pointer px-2 text-xs font-medium sm:flex-none"
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
						class="text-muted-foreground h-9 w-full cursor-pointer px-2 text-xs font-medium sm:w-auto"
					>
						<option value={100}>100 items</option>
						<option value={200}>200 items</option>
						<option value={500}>500 items</option>
						<option value={1000}>1,000 items</option>
					</select>
					<div class="text-muted-foreground text-[0.7rem] whitespace-nowrap">
						<span class="num text-foreground font-medium">{products.length}</span> results
					</div>
				</div>
			</div>
		</div>
	</div>
	<div class="custom-scrollbar max-h-[22rem] overflow-y-auto p-3 pt-3">
		{#if products.length === 0}
			<div class="text-muted-foreground flex h-full items-center justify-center text-[0.8rem]">
				{$t('pos.products.no_results')}
			</div>
		{:else}
			<div class="grid grid-cols-2 gap-3 xl:grid-cols-3">
				{#each products as product}
					{@const { price, originalPrice, discountName } = getProductPrice(product)}
					<button
						type="button"
						onclick={() => onAddToCart(product)}
						class="liquid-panel group flex flex-col justify-between p-3 text-left transition-all duration-200 hover:-translate-y-0.5 hover:border-primary/40 hover:shadow-md"
					>
						<div class="space-y-1.5">
							<div class="text-foreground line-clamp-2 text-[0.8rem] font-medium">
								{product.Name}
							</div>
							<div class="mt-0.5 flex flex-col gap-1">
								<div class="flex flex-wrap items-center gap-2">
									{#if price < originalPrice}
										<span class="num-mono text-muted-foreground text-[0.65rem] line-through">
											{formatCurrency(originalPrice)}
										</span>
										<span class="num-mono text-danger text-[0.85rem] font-bold">
											{formatCurrency(price)}
										</span>
									{:else}
										<span class="num-mono text-primary text-[0.85rem] font-semibold">
											{formatCurrency(price)}
										</span>
									{/if}
								</div>

								{#if discountName}
									<span class="badge-danger w-fit px-1 py-0 text-[0.6rem]">
										{discountName}
									</span>
								{/if}

								<div>
									<span
										class="{getAvailableStock(product) > 0 ? 'badge-success' : 'badge-danger'} px-1.5 py-0 text-[0.65rem]"
									>
										{$t('pos.products.in_stock', { count: getAvailableStock(product) })}
									</span>
								</div>
							</div>
						</div>
						<div class="text-muted-foreground mt-3 flex items-center justify-between text-[0.7rem]">
							<span class="group-hover:text-primary flex items-center gap-1 font-medium">
								<span
									class="border-border bg-muted text-muted-foreground rounded-md border px-1.5 py-0.5 text-[0.65rem]"
								>
									Tap
								</span>
								{$t('pos.products.tap_to_add')}
							</span>
							<span class="num-mono text-[0.65rem] tracking-wide uppercase">#{product.ID}</span>
						</div>
					</button>
				{/each}
			</div>
		{/if}
	</div>
</div>
