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

<div class="panel flex-1 overflow-hidden">
	<div class="border-b border-border/80 bg-muted/40 px-4 py-3">
		<div class="flex flex-col gap-3">
			<div>
				<h3 class="section-title">{$t('pos.products.title')}</h3>
				<p class="mt-0.5 text-xs text-muted-foreground">
					{$t('pos.products.description')}
				</p>
			</div>
			<div class="flex flex-wrap items-center gap-2">
				<!-- Status Filter -->
				<select
					bind:value={fetchStatus}
					onchange={onFilterChange}
					class="h-9 flex-1 cursor-pointer px-2 text-xs font-medium text-foreground sm:flex-none"
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
						class="h-9 w-full cursor-pointer px-2 text-xs font-medium text-foreground sm:w-auto"
					>
						<option value={100}>100 items</option>
						<option value={200}>200 items</option>
						<option value={500}>500 items</option>
						<option value={1000}>1,000 items</option>
					</select>
					<div class="text-[0.7rem] whitespace-nowrap text-muted-foreground">
						<span class="num font-semibold text-foreground">{products.length}</span> results
					</div>
				</div>
			</div>
		</div>
	</div>
	<div class="custom-scrollbar max-h-[22rem] overflow-y-auto p-3">
		{#if products.length === 0}
			<div class="flex h-full min-h-32 items-center justify-center text-sm text-muted-foreground">
				{$t('pos.products.no_results')}
			</div>
		{:else}
			<div class="grid grid-cols-2 gap-3 xl:grid-cols-3">
				{#each products as product}
					{@const { price, originalPrice, discountName } = getProductPrice(product)}
					<button
						type="button"
						onclick={() => onAddToCart(product)}
						class="panel group flex flex-col justify-between p-3 text-left transition-all duration-200 hover:-translate-y-0.5 hover:border-primary/40 hover:shadow-md"
					>
						<div class="space-y-1.5">
							<div class="line-clamp-2 text-[0.8125rem] font-medium text-foreground">
								{product.Name}
							</div>
							<div class="mt-0.5 flex flex-col gap-1">
								<div class="flex flex-wrap items-center gap-2">
									{#if price < originalPrice}
										<span class="num-mono text-[0.65rem] text-muted-foreground line-through">
											{formatCurrency(originalPrice)}
										</span>
										<span class="num-mono text-sm font-semibold text-primary">
											{formatCurrency(price)}
										</span>
									{:else}
										<span class="num-mono text-sm font-semibold text-foreground">
											{formatCurrency(price)}
										</span>
									{/if}
								</div>

								{#if discountName}
									<span class="badge-info w-fit px-1.5 py-0 text-[0.6rem]">
										{discountName}
									</span>
								{/if}

								<div>
									<span
										class="{getAvailableStock(product) > 0
											? 'badge-success'
											: 'badge-danger'} px-1.5 py-0 text-[0.65rem]"
									>
										{$t('pos.products.in_stock', { count: getAvailableStock(product) })}
									</span>
								</div>
							</div>
						</div>
						<div
							class="mt-3 flex items-center justify-between gap-2 text-[0.7rem] text-muted-foreground"
						>
							<span class="flex items-center gap-1 font-medium group-hover:text-primary">
								<span
									class="rounded-md bg-muted px-1.5 py-0.5 text-[0.65rem] text-muted-foreground"
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
