<script lang="ts">
	import { productsApi, inventoryApi, barcodeApi } from '$lib/api/resources';
	import { toast } from 'svelte-sonner';
	import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '$lib/components/ui/card';
	import { Input } from '$lib/components/ui/input';
	import { Button } from '$lib/components/ui/button';
	import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '$lib/components/ui/table';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import type { Batch, Product } from '$lib/types';

	const stockQuery = $state({ productId: '', locationId: '' });
	let stockLoading = $state(false);
	let stockSnapshot = $state<{ currentQuantity: number; batches: Batch[] } | null>(null);

	const adjustmentForm = $state({ productId: '', type: 'STOCK_IN', quantity: '0', reasonCode: 'MANUAL', notes: '' });

	const transferForm = $state({ productId: '', sourceLocationId: '', destLocationId: '', quantity: '0' });

	const barcodeLookup = $state({ value: '', result: null as Product | null });
	let barcodeImage = $state<string | null>(null);

	const loadStock = async () => {
		if (!stockQuery.productId) {
			toast.warning('Enter a product ID first');
			return;
		}
		stockLoading = true;
		try {
			const snapshot = await productsApi.stock(Number(stockQuery.productId), {
				locationId: stockQuery.locationId || undefined,
			});
			stockSnapshot = snapshot;
			toast.success('Inventory snapshot updated');
		} catch (error) {
			const errorMessage = error.response?.data?.error || 'Unable to fetch stock';
			toast.error('Failed to Fetch Stock', {
				description: errorMessage,
			});
		} finally {
			stockLoading = false;
		}
	};

	const submitAdjustment = async () => {
		if (!adjustmentForm.productId) {
			toast.warning('Select a product');
			return;
		}
		try {
			await productsApi.adjustStock(Number(adjustmentForm.productId), {
				type: adjustmentForm.type,
				quantity: Number(adjustmentForm.quantity),
				reasonCode: adjustmentForm.reasonCode,
				notes: adjustmentForm.notes,
			});
			toast.success('Stock adjusted');
			loadStock();
		} catch (error) {
			const errorMessage = error.response?.data?.error || 'Unable to apply adjustment';
			toast.error('Failed to Apply Adjustment', {
				description: errorMessage,
			});
		}
	};

	const submitTransfer = async () => {
		try {
			await inventoryApi.transfer({
				productId: Number(transferForm.productId),
				sourceLocationId: Number(transferForm.sourceLocationId),
				destLocationId: Number(transferForm.destLocationId),
				quantity: Number(transferForm.quantity),
			});
			toast.success('Transfer queued');
		} catch (error) {
			const errorMessage = error.response?.data?.error || 'Unable to create transfer';
			toast.error('Failed to Create Transfer', {
				description: errorMessage,
			});
		}
	};

	const runBarcodeLookup = async () => {
		if (!barcodeLookup.value) {
			toast.warning('Provide a barcode value');
			return;
		}
		try {
			const product = await barcodeApi.lookup(barcodeLookup.value);
			barcodeLookup.result = product;
			toast.success('SKU resolved');
		} catch (error) {
			const errorMessage = error.response?.data?.error || 'Product not found';
			toast.error('Failed to Lookup Barcode', {
				description: errorMessage,
			});
		}
	};

	const generateBarcode = async () => {
		if (!barcodeLookup.value) {
			toast.warning('Provide SKU or product ID');
			return;
		}
		try {
			const blob = await barcodeApi.generate({ sku: barcodeLookup.value });
			barcodeImage = URL.createObjectURL(blob);
		} catch (error) {
			const errorMessage = error.response?.data?.error || 'Unable to generate barcode';
			toast.error('Failed to Generate Barcode', {
				description: errorMessage,
			});
		}
	};
</script>

<section class="space-y-8">
	<header>
		<p class="text-sm uppercase tracking-wide text-muted-foreground">Network operations</p>
		<h1 class="text-3xl font-semibold">Stock adjustments, transfers & barcode intelligence</h1>
	</header>

	<div class="grid gap-6 lg:grid-cols-2">
		<Card>
			<CardHeader>
				<CardTitle>Inventory snapshot</CardTitle>
				<CardDescription>Read a product's current balance per location</CardDescription>
			</CardHeader>
			<CardContent class="space-y-4">
				<div class="grid gap-3 sm:grid-cols-2">
					<Input type="number" min="1" placeholder="Product ID" bind:value={stockQuery.productId} />
					<Input type="number" min="1" placeholder="Location ID (optional)" bind:value={stockQuery.locationId} />
				</div>
				<Button class="w-full" onclick={loadStock}>Fetch stock levels</Button>
				{#if stockLoading}
					<Skeleton class="h-32 w-full" />
				{:else if stockSnapshot}
					<div class="rounded-2xl border border-border/70 p-4">
						<p class="text-sm text-muted-foreground">Current quantity</p>
						<p class="text-3xl font-semibold">{stockSnapshot.currentQuantity}</p>
					</div>
					<Table class="mt-4">
						<TableHeader>
							<TableRow>
								<TableHead>Batch</TableHead>
								<TableHead>Qty</TableHead>
								<TableHead>Expiry</TableHead>
							</TableRow>
						</TableHeader>
						<TableBody>
							{#if stockSnapshot.batches.length === 0}
								<TableRow>
									<TableCell colspan="3" class="text-center text-sm text-muted-foreground">No batch detail available</TableCell>
								</TableRow>
							{:else}
								{#each stockSnapshot.batches as batch}
									<TableRow>
										<TableCell>{batch.BatchNumber}</TableCell>
										<TableCell>{batch.Quantity}</TableCell>
										<TableCell>{batch.ExpiryDate ?? '—'}</TableCell>
									</TableRow>
								{/each}
							{/if}
						</TableBody>
					</Table>
				{/if}
			</CardContent>
		</Card>
		<Card>
			<CardHeader>
				<CardTitle>Manual adjustment</CardTitle>
				<CardDescription>Adhoc cycle counts, write-offs, or inbound receipts</CardDescription>
			</CardHeader>
			<CardContent class="space-y-3">
				<Input type="number" min="1" placeholder="Product ID" bind:value={adjustmentForm.productId} />
				<select class="w-full rounded-md border border-border bg-background px-3 py-2 text-sm" bind:value={adjustmentForm.type}>
					<option value="STOCK_IN">Stock in</option>
					<option value="STOCK_OUT">Stock out</option>
				</select>
				<Input type="number" min="1" placeholder="Quantity" bind:value={adjustmentForm.quantity} />
				<Input placeholder="Reason code" bind:value={adjustmentForm.reasonCode} />
				<Input placeholder="Notes" bind:value={adjustmentForm.notes} />
				<Button class="w-full" onclick={submitAdjustment}>Apply adjustment</Button>
			</CardContent>
		</Card>
	</div>

	<div class="grid gap-6 lg:grid-cols-2">
		<Card>
			<CardHeader>
				<CardTitle>Stock transfer</CardTitle>
				<CardDescription>Move inventory across locations with audit</CardDescription>
			</CardHeader>
			<CardContent class="space-y-3">
				<Input type="number" min="1" placeholder="Product ID" bind:value={transferForm.productId} />
				<div class="grid grid-cols-2 gap-3">
					<Input type="number" min="1" placeholder="Source location" bind:value={transferForm.sourceLocationId} />
					<Input type="number" min="1" placeholder="Destination location" bind:value={transferForm.destLocationId} />
				</div>
				<Input type="number" min="1" placeholder="Quantity" bind:value={transferForm.quantity} />
				<Button class="w-full" onclick={submitTransfer}>Create transfer</Button>
			</CardContent>
		</Card>
		<Card>
			<CardHeader>
				<CardTitle>Barcode intelligence</CardTitle>
				<CardDescription>Resolve SKUs and render printable codes</CardDescription>
			</CardHeader>
			<CardContent class="space-y-3">
				<Input placeholder="Scan or type barcode / SKU" bind:value={barcodeLookup.value} />
				<div class="flex gap-2">
					<Button class="w-full" variant="secondary" onclick={runBarcodeLookup}>Lookup product</Button>
					<Button class="w-full" onclick={generateBarcode}>Generate image</Button>
				</div>
				{#if barcodeLookup.result}
					<div class="rounded-xl border border-border/70 p-3 text-sm">
						<p class="font-semibold">{barcodeLookup.result.Name}</p>
						<p class="text-muted-foreground">SKU: {barcodeLookup.result.SKU}</p>
					</div>
				{/if}
				{#if barcodeImage}
					<div class="rounded-xl border border-dashed border-border/70 p-3 text-center">
						<img src={barcodeImage} alt="Barcode preview" class="mx-auto" />
					</div>
				{/if}
			</CardContent>
		</Card>
	</div>
</section>
