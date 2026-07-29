<script lang="ts">
	import { onMount } from 'svelte';
	import { productsApi, inventoryApi, barcodeApi } from '$lib/api/resources';
	import { toast } from 'svelte-sonner';
	import { Input } from '$lib/components/ui/input';
	import { Button } from '$lib/components/ui/button';
	import { Activity, ArrowRightLeft, ClipboardCheck, ScanLine } from 'lucide-svelte';
	import type { Batch, Product } from '$lib/types';
	import ProductSelector from '$lib/components/ui/product-selector.svelte';
	import { auth } from '$lib/stores/auth';
	import { goto } from '$app/navigation';
	import { t } from '$lib/i18n';

	const inputClass =
		'liquid-input h-10 w-full rounded-md bg-card bg-none px-3 text-sm font-normal text-foreground shadow-none ring-0 backdrop-blur-none backdrop-saturate-100 from-card to-card placeholder:text-muted-foreground focus:scale-100 focus:border-primary focus:ring-0 md:text-sm';
	const btnPrimary =
		'rounded-md bg-primary bg-none px-4 text-sm font-medium text-primary-foreground shadow-none hover:bg-primary/90 hover:bg-none';
	const btnSecondary =
		'rounded-md border border-border bg-card bg-none text-foreground shadow-none backdrop-blur-none hover:bg-muted hover:bg-none';

	$effect(() => {
		if (!auth.hasPermission('inventory.view')) {
			toast.error($t('operations.toasts.access_denied'), {
				description: $t('operations.toasts.access_denied_desc')
			});
			goto('/');
		}
	});

	const stockQuery = $state({ productId: '', locationId: '' });
	let stockLoading = $state(false);
	let stockSnapshot = $state<{ currentQuantity: number; batches: Batch[] } | null>(null);

	const adjustmentForm = $state({
		productId: '',
		type: 'STOCK_IN',
		quantity: '0',
		reasonCode: 'MANUAL',
		notes: ''
	});
	const transferForm = $state({
		productId: '',
		sourceLocationId: '',
		destLocationId: '',
		quantity: '0'
	});
	const barcodeLookup = $state({ value: '', result: null as Product | null });
	let barcodeImage = $state<string | null>(null);

	const loadStock = async () => {
		if (!stockQuery.productId) return toast.warning($t('operations.toasts.product_id_required'));
		stockLoading = true;
		try {
			const snapshot = await productsApi.stock(Number(stockQuery.productId), {
				locationId: stockQuery.locationId || undefined
			});
			stockSnapshot = snapshot;
			toast.success($t('operations.toasts.snapshot_updated'));
		} catch (error) {
			toast.error($t('operations.toasts.fetch_stock_fail'), {
				description: (error as any)?.response?.data?.error || 'Unable to fetch stock'
			});
		} finally {
			stockLoading = false;
		}
	};

	const submitAdjustment = async () => {
		if (!adjustmentForm.productId) return toast.warning($t('operations.toasts.select_product'));
		try {
			await productsApi.adjustStock(Number(adjustmentForm.productId), {
				type: adjustmentForm.type,
				quantity: Number(adjustmentForm.quantity),
				reasonCode: adjustmentForm.reasonCode,
				notes: adjustmentForm.notes
			});
			toast.success($t('operations.toasts.stock_adjusted'));
			loadStock();
		} catch (error) {
			toast.error($t('operations.toasts.adjust_fail'), {
				description: (error as any)?.response?.data?.error || 'Unable to apply adjustment'
			});
		}
	};

	const submitTransfer = async () => {
		try {
			await inventoryApi.transfer({
				productId: Number(transferForm.productId),
				sourceLocationId: Number(transferForm.sourceLocationId),
				destLocationId: Number(transferForm.destLocationId),
				quantity: Number(transferForm.quantity)
			});
			toast.success($t('operations.toasts.transfer_queued'));
		} catch (error) {
			toast.error($t('operations.toasts.transfer_fail'), {
				description: (error as any)?.response?.data?.error || 'Unable to create transfer'
			});
		}
	};

	const runBarcodeLookup = async () => {
		if (!barcodeLookup.value) return toast.warning($t('operations.toasts.barcode_required'));
		try {
			const product = await barcodeApi.lookup(barcodeLookup.value);
			barcodeLookup.result = product;
			toast.success($t('operations.toasts.sku_resolved'));
		} catch (error) {
			toast.error($t('operations.toasts.lookup_fail'), {
				description: (error as any)?.response?.data?.error || $t('operations.toasts.product_not_found')
			});
		}
	};

	const generateBarcode = async () => {
		if (!barcodeLookup.value) return toast.warning($t('operations.toasts.sku_or_id_required'));
		try {
			const blob = await barcodeApi.generate({ sku: barcodeLookup.value });
			barcodeImage = URL.createObjectURL(blob);
		} catch (error) {
			toast.error($t('operations.toasts.generate_fail'), {
				description: (error as any)?.response?.data?.error || 'Unable to generate barcode'
			});
		}
	};

	// --- Parallax Effect ---
	onMount(() => {
		const hero = document.querySelector('.parallax-hero') as HTMLElement | null;
		if (!hero) return;
		const handleScroll = () => {
			const scrollY = window.scrollY / 6;
			hero.style.transform = `translateY(${scrollY}px)`;
		};
		window.addEventListener('scroll', handleScroll);
		return () => window.removeEventListener('scroll', handleScroll);
	});
</script>

<!-- HEADER -->
<section class="w-full">
	<div class="mx-auto max-w-7xl px-6 pt-10 pb-6 sm:pt-14">
		<div
			class="parallax-hero space-y-2 transition-transform duration-700 ease-out will-change-transform"
			data-animate="fade-up"
		>
			<p class="overline flex items-center gap-2">
				<span class="bg-primary/10 text-primary inline-flex rounded-md p-1.5">
					<Activity class="h-4 w-4" />
				</span>
				{$t('operations.hero.subtitle')}
			</p>
			<h1 class="font-display text-foreground text-3xl font-bold tracking-tight sm:text-4xl">
				{$t('operations.hero.title')}
			</h1>
		</div>
	</div>
</section>

<!-- MAIN CONTENT -->
<section class="mx-auto max-w-7xl space-y-8 px-6 py-8">
	<div class="grid gap-6 lg:grid-cols-2">
		<!-- Inventory Snapshot -->
		<div class="liquid-panel rounded-lg">
			<div class="border-border space-y-1 border-b px-6 py-5">
				<h2 class="text-foreground flex items-center gap-2 font-display text-lg font-semibold">
					<ClipboardCheck class="text-primary h-5 w-5" />
					{$t('operations.snapshot.title')}
				</h2>
				<p class="text-muted-foreground text-sm">{$t('operations.snapshot.subtitle')}</p>
			</div>
			<div class="space-y-4 p-6">
				<div class="grid gap-3 sm:grid-cols-2">
					<ProductSelector
						bind:value={stockQuery.productId}
						placeholder={$t('operations.snapshot.search_placeholder')}
						className="w-full"
						onSelect={() => setTimeout(loadStock, 100)}
					/>
					<Input
						type="number"
						placeholder={$t('operations.snapshot.location_id')}
						bind:value={stockQuery.locationId}
						class="{inputClass} num-mono"
						onkeydown={(e) => e.key === 'Enter' && loadStock()}
					/>
				</div>
				<Button class="{btnPrimary} w-full" onclick={loadStock}>
					{$t('operations.snapshot.fetch_button')}
				</Button>
				{#if stockLoading}
					<div class="bg-muted h-32 w-full animate-pulse rounded-md"></div>
				{:else if stockSnapshot}
					<div class="border-border bg-muted/50 rounded-lg border p-4">
						<p class="overline">{$t('operations.snapshot.current_qty')}</p>
						<p class="kpi-num text-primary mt-1 text-3xl">{stockSnapshot.currentQuantity}</p>
					</div>
					<div class="liquid-panel custom-scrollbar mt-4 overflow-x-auto rounded-lg">
						<table class="ledger-table min-w-[24rem]">
							<thead>
								<tr>
									<th>{$t('operations.snapshot.table.batch')}</th>
									<th class="text-right">{$t('operations.snapshot.table.qty')}</th>
									<th>{$t('operations.snapshot.table.expiry')}</th>
								</tr>
							</thead>
							<tbody>
								{#if stockSnapshot.batches.length === 0}
									<tr>
										<td colspan={3} class="text-muted-foreground py-4 text-center text-sm"
											>{$t('operations.snapshot.table.empty')}</td
										>
									</tr>
								{:else}
									{#each stockSnapshot.batches as batch}
										<tr>
											<td class="num-mono text-foreground">{batch.BatchNumber}</td>
											<td class="col-num">{batch.Quantity}</td>
											<td class="text-muted-foreground">{batch.ExpiryDate ?? '—'}</td>
										</tr>
									{/each}
								{/if}
							</tbody>
						</table>
					</div>
				{/if}
			</div>
		</div>

		<!-- Manual Adjustment -->
		{#if auth.hasPermission('products.write')}
			<div class="liquid-panel rounded-lg">
				<div class="border-border space-y-1 border-b px-6 py-5">
					<h2 class="text-foreground flex items-center gap-2 font-display text-lg font-semibold">
						<Activity class="text-success h-5 w-5" />
						{$t('operations.adjustment.title')}
					</h2>
					<p class="text-muted-foreground text-sm">{$t('operations.adjustment.subtitle')}</p>
				</div>
				<div class="space-y-4 p-6">
					<ProductSelector
						bind:value={adjustmentForm.productId}
						placeholder={$t('operations.adjustment.select_product')}
						className="w-full"
					/>
					<div class="grid grid-cols-2 gap-3">
						<div class="select-wrapper">
							<select bind:value={adjustmentForm.type}>
								<option value="STOCK_IN">{$t('operations.adjustment.stock_in')}</option>
								<option value="STOCK_OUT">{$t('operations.adjustment.stock_out')}</option>
							</select>
						</div>
						<Input
							type="number"
							placeholder={$t('operations.adjustment.quantity')}
							bind:value={adjustmentForm.quantity}
							class="{inputClass} num-mono"
						/>
					</div>
					<Input
						placeholder={$t('operations.adjustment.reason_code')}
						bind:value={adjustmentForm.reasonCode}
						class={inputClass}
					/>
					<Input
						placeholder={$t('operations.adjustment.notes')}
						bind:value={adjustmentForm.notes}
						class={inputClass}
					/>
					<Button class="{btnPrimary} w-full" onclick={submitAdjustment}
						>{$t('operations.adjustment.submit_button')}</Button
					>
				</div>
			</div>
		{/if}
	</div>

	<div class="grid gap-6 lg:grid-cols-2">
		<!-- Stock Transfer -->
		{#if auth.hasPermission('products.write')}
			<div class="liquid-panel rounded-lg">
				<div class="border-border space-y-1 border-b px-6 py-5">
					<h2 class="text-foreground flex items-center gap-2 font-display text-lg font-semibold">
						<ArrowRightLeft class="text-info h-5 w-5" />
						{$t('operations.transfer.title')}
					</h2>
					<p class="text-muted-foreground text-sm">{$t('operations.transfer.subtitle')}</p>
				</div>
				<div class="space-y-4 p-6">
					<ProductSelector
						bind:value={transferForm.productId}
						placeholder={$t('operations.transfer.select_product')}
						className="w-full"
					/>
					<div class="grid grid-cols-2 gap-3">
						<Input
							type="number"
							placeholder={$t('operations.transfer.source')}
							bind:value={transferForm.sourceLocationId}
							class="{inputClass} num-mono"
						/>
						<Input
							type="number"
							placeholder={$t('operations.transfer.dest')}
							bind:value={transferForm.destLocationId}
							class="{inputClass} num-mono"
						/>
					</div>
					<Input
						type="number"
						placeholder={$t('operations.transfer.quantity')}
						bind:value={transferForm.quantity}
						class="{inputClass} num-mono"
					/>
					<Button class="{btnPrimary} w-full" onclick={submitTransfer}
						>{$t('operations.transfer.submit_button')}</Button
					>
				</div>
			</div>
		{/if}

		<!-- Barcode Intelligence -->
		<div class="liquid-panel rounded-lg">
			<div class="border-border space-y-1 border-b px-6 py-5">
				<h2 class="text-foreground flex items-center gap-2 font-display text-lg font-semibold">
					<ScanLine class="text-warning h-5 w-5" />
					{$t('operations.barcode.title')}
				</h2>
				<p class="text-muted-foreground text-sm">{$t('operations.barcode.subtitle')}</p>
			</div>
			<div class="space-y-4 p-6">
				<Input
					placeholder={$t('operations.barcode.input_placeholder')}
					bind:value={barcodeLookup.value}
					class="{inputClass} num-mono"
					onkeydown={(e) => e.key === 'Enter' && runBarcodeLookup()}
				/>
				<div class="flex flex-col gap-3 sm:flex-row">
					<Button class="{btnSecondary} flex-1" variant="secondary" onclick={runBarcodeLookup}>
						{$t('operations.barcode.lookup_button')}
					</Button>

					<Button class="{btnPrimary} flex-1" onclick={generateBarcode}>
						{$t('operations.barcode.generate_button')}
					</Button>
				</div>

				{#if barcodeLookup.result}
					<div class="border-border bg-muted/50 rounded-lg border p-3 text-sm">
						<p class="text-foreground font-semibold">{barcodeLookup.result.Name}</p>
						<p class="text-muted-foreground">
							SKU: <span class="num-mono">{barcodeLookup.result.SKU}</span>
						</p>
					</div>
				{/if}
				{#if barcodeImage}
					<div class="border-border bg-card rounded-lg border border-dashed p-3 text-center shadow-sm">
						<img src={barcodeImage} alt={$t('operations.barcode.preview_alt')} class="mx-auto" />
					</div>
				{/if}
			</div>
		</div>
	</div>
</section>
