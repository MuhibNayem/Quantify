<script lang="ts">
	import { onMount } from 'svelte';
	import { productsApi, inventoryApi, barcodeApi } from '$lib/api/resources';
	import { toast } from 'svelte-sonner';
	import { Input } from '$lib/components/ui/input';
	import { Button } from '$lib/components/ui/button';
	import { Activity, ArrowRightLeft, ClipboardCheck, ScanLine } from 'lucide-svelte';
	import type { Batch, Product, ProductUnit } from '$lib/types';
	import ProductSelector from '$lib/components/ui/product-selector.svelte';
	import { auth } from '$lib/stores/auth';
	import { goto } from '$app/navigation';
	import { t } from '$lib/i18n';

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
	const barcodeLookup = $state({
		value: '',
		result: null as Product | null,
		unit: null as ProductUnit | null
	});
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
			const resolved = await barcodeApi.resolve(barcodeLookup.value);
			barcodeLookup.result = resolved.product;
			barcodeLookup.unit = resolved.unit;
			toast.success($t('operations.toasts.sku_resolved'));
		} catch (error) {
			toast.error($t('operations.toasts.lookup_fail'), {
				description:
					(error as any)?.response?.data?.error || $t('operations.toasts.product_not_found')
			});
		}
	};

	const generateBarcode = async () => {
		if (!barcodeLookup.value) return toast.warning($t('operations.toasts.sku_required'));
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

<div class="mx-auto w-full max-w-7xl px-4 py-8 sm:px-6 lg:px-8 lg:py-10">
	<!-- Page header -->
	<header class="page-header" data-animate="fade-up">
		<div class="parallax-hero transition-transform duration-700 ease-out will-change-transform">
			<p class="flex items-center gap-2 overline">
				<span class="inline-flex rounded-md bg-primary/10 p-1.5 text-primary">
					<Activity class="h-3.5 w-3.5" />
				</span>
				{$t('operations.hero.subtitle')}
			</p>
			<h1 class="page-title mt-1">{$t('operations.hero.title')}</h1>
		</div>
	</header>

	<!-- MAIN CONTENT -->
	<div class="mt-6 grid gap-4 lg:grid-cols-2">
		<!-- Inventory Snapshot -->
		<section class="panel overflow-hidden">
			<div class="panel-header items-start">
				<div class="flex items-start gap-3">
					<span
						class="flex h-9 w-9 shrink-0 items-center justify-center rounded-md bg-accent text-accent-foreground ring-1 ring-accent-foreground/10"
					>
						<ClipboardCheck class="h-4 w-4" />
					</span>
					<div class="space-y-0.5">
						<h2 class="section-title">{$t('operations.snapshot.title')}</h2>
						<p class="text-sm text-muted-foreground">{$t('operations.snapshot.subtitle')}</p>
					</div>
				</div>
			</div>
			<div class="panel-body space-y-4">
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
						class="num-mono h-11 w-full rounded-md text-sm"
						onkeydown={(e) => e.key === 'Enter' && loadStock()}
					/>
				</div>
				<Button class="btn-soothe w-full" onclick={loadStock}>
					{$t('operations.snapshot.fetch_button')}
				</Button>
				{#if stockLoading}
					<div class="h-32 w-full animate-pulse rounded-lg bg-muted"></div>
				{:else if stockSnapshot}
					<div class="rounded-lg bg-muted/40 p-4 ring-1 ring-border/70">
						<p class="overline">{$t('operations.snapshot.current_qty')}</p>
						<p class="kpi-num mt-1 text-3xl text-primary">{stockSnapshot.currentQuantity}</p>
					</div>
					<div class="custom-scrollbar overflow-x-auto rounded-lg ring-1 ring-border">
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
										<td colspan={3} class="py-6 text-center text-sm text-muted-foreground"
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
		</section>

		<!-- Manual Adjustment -->
		{#if auth.hasPermission('products.write')}
			<section class="panel overflow-hidden">
				<div class="panel-header items-start">
					<div class="flex items-start gap-3">
						<span
							class="flex h-9 w-9 shrink-0 items-center justify-center rounded-md bg-accent text-accent-foreground ring-1 ring-accent-foreground/10"
						>
							<Activity class="h-4 w-4" />
						</span>
						<div class="space-y-0.5">
							<h2 class="section-title">{$t('operations.adjustment.title')}</h2>
							<p class="text-sm text-muted-foreground">{$t('operations.adjustment.subtitle')}</p>
						</div>
					</div>
				</div>
				<div class="panel-body space-y-3">
					<ProductSelector
						bind:value={adjustmentForm.productId}
						placeholder={$t('operations.adjustment.select_product')}
						className="w-full"
					/>
					<div class="grid grid-cols-2 gap-3">
						<select
							bind:value={adjustmentForm.type}
							aria-label={$t('operations.adjustment.title')}
							class="h-11 text-sm"
						>
							<option value="STOCK_IN">{$t('operations.adjustment.stock_in')}</option>
							<option value="STOCK_OUT">{$t('operations.adjustment.stock_out')}</option>
						</select>
						<Input
							type="number"
							placeholder={$t('operations.adjustment.quantity')}
							bind:value={adjustmentForm.quantity}
							class="num-mono h-11 w-full rounded-md text-sm"
						/>
					</div>
					<Input
						placeholder={$t('operations.adjustment.reason_code')}
						bind:value={adjustmentForm.reasonCode}
						class="h-11 w-full rounded-md text-sm"
					/>
					<Input
						placeholder={$t('operations.adjustment.notes')}
						bind:value={adjustmentForm.notes}
						class="h-11 w-full rounded-md text-sm"
					/>
					<Button class="btn-soothe w-full" onclick={submitAdjustment}
						>{$t('operations.adjustment.submit_button')}</Button
					>
				</div>
			</section>
		{/if}
	</div>

	<div class="mt-4 grid gap-4 lg:grid-cols-2">
		<!-- Stock Transfer -->
		{#if auth.hasPermission('products.write')}
			<section class="panel overflow-hidden">
				<div class="panel-header items-start">
					<div class="flex items-start gap-3">
						<span
							class="flex h-9 w-9 shrink-0 items-center justify-center rounded-md bg-accent text-accent-foreground ring-1 ring-accent-foreground/10"
						>
							<ArrowRightLeft class="h-4 w-4" />
						</span>
						<div class="space-y-0.5">
							<h2 class="section-title">{$t('operations.transfer.title')}</h2>
							<p class="text-sm text-muted-foreground">{$t('operations.transfer.subtitle')}</p>
						</div>
					</div>
				</div>
				<div class="panel-body space-y-3">
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
							class="num-mono h-11 w-full rounded-md text-sm"
						/>
						<Input
							type="number"
							placeholder={$t('operations.transfer.dest')}
							bind:value={transferForm.destLocationId}
							class="num-mono h-11 w-full rounded-md text-sm"
						/>
					</div>
					<Input
						type="number"
						placeholder={$t('operations.transfer.quantity')}
						bind:value={transferForm.quantity}
						class="num-mono h-11 w-full rounded-md text-sm"
					/>
					<Button class="btn-soothe w-full" onclick={submitTransfer}
						>{$t('operations.transfer.submit_button')}</Button
					>
				</div>
			</section>
		{/if}

		<!-- Barcode Intelligence -->
		<section class="panel overflow-hidden">
			<div class="panel-header items-start">
				<div class="flex items-start gap-3">
					<span
						class="flex h-9 w-9 shrink-0 items-center justify-center rounded-md bg-accent text-accent-foreground ring-1 ring-accent-foreground/10"
					>
						<ScanLine class="h-4 w-4" />
					</span>
					<div class="space-y-0.5">
						<h2 class="section-title">{$t('operations.barcode.title')}</h2>
						<p class="text-sm text-muted-foreground">{$t('operations.barcode.subtitle')}</p>
					</div>
				</div>
			</div>
			<div class="panel-body space-y-3">
				<Input
					placeholder={$t('operations.barcode.input_placeholder')}
					bind:value={barcodeLookup.value}
					class="num-mono h-11 w-full rounded-md text-sm"
					onkeydown={(e) => e.key === 'Enter' && runBarcodeLookup()}
				/>
				<div class="flex flex-col gap-2.5 sm:flex-row">
					<Button variant="secondary" class="flex-1" onclick={runBarcodeLookup}>
						{$t('operations.barcode.lookup_button')}
					</Button>

					<Button class="btn-soothe flex-1" onclick={generateBarcode}>
						{$t('operations.barcode.generate_button')}
					</Button>
				</div>

				{#if barcodeLookup.result}
					<div class="rounded-lg bg-muted/40 p-3 text-sm ring-1 ring-border/70">
						<p class="font-semibold text-foreground">{barcodeLookup.result.Name}</p>
						<p class="text-muted-foreground">
							SKU: <span class="num-mono">{barcodeLookup.result.SKU}</span>
						</p>
						{#if barcodeLookup.unit}
							<p class="mt-1">
								<span class="badge-info">
									Unit: {barcodeLookup.unit.Unit?.Name ??
										barcodeLookup.unit.Unit?.Symbol ??
										'selling unit'}
									· ×{barcodeLookup.unit.Factor}
								</span>
							</p>
						{/if}
					</div>
				{/if}
				{#if barcodeImage}
					<div class="rounded-lg border border-dashed border-border/70 bg-muted/20 p-3 text-center">
						<img src={barcodeImage} alt={$t('operations.barcode.preview_alt')} class="mx-auto" />
					</div>
				{/if}
			</div>
		</section>
	</div>
</div>
