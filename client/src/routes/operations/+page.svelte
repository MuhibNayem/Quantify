<script lang="ts">
	import { onMount } from 'svelte';
	import { productsApi, inventoryApi, barcodeApi } from '$lib/api/resources';
	import { toast } from 'svelte-sonner';
	import {
		Card,
		CardContent,
		CardDescription,
		CardHeader,
		CardTitle
	} from '$lib/components/ui/card';
	import { Input } from '$lib/components/ui/input';
	import { Button } from '$lib/components/ui/button';
	import {
		Table,
		TableBody,
		TableCell,
		TableHead,
		TableHeader,
		TableRow
	} from '$lib/components/ui/table';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import { Activity, ArrowRightLeft, ClipboardCheck, ScanLine } from 'lucide-svelte';
	import type { Batch, Product } from '$lib/types';
	import ProductSelector from '$lib/components/ui/product-selector.svelte';
	import { auth } from '$lib/stores/auth';
	import { goto } from '$app/navigation';

	$effect(() => {
		if (!auth.hasPermission('inventory.view')) {
			toast.error('Access Denied', {
				description: 'You do not have permission to access operations.'
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
		if (!stockQuery.productId) return toast.warning('Enter a product ID first');
		stockLoading = true;
		try {
			const snapshot = await productsApi.stock(Number(stockQuery.productId), {
				locationId: stockQuery.locationId || undefined
			});
			stockSnapshot = snapshot;
			toast.success('Inventory snapshot updated');
		} catch (error) {
			toast.error('Failed to Fetch Stock', {
				description: error.response?.data?.error || 'Unable to fetch stock'
			});
		} finally {
			stockLoading = false;
		}
	};

	const submitAdjustment = async () => {
		if (!adjustmentForm.productId) return toast.warning('Select a product');
		try {
			await productsApi.adjustStock(Number(adjustmentForm.productId), {
				type: adjustmentForm.type,
				quantity: Number(adjustmentForm.quantity),
				reasonCode: adjustmentForm.reasonCode,
				notes: adjustmentForm.notes
			});
			toast.success('Stock adjusted');
			loadStock();
		} catch (error) {
			toast.error('Failed to Apply Adjustment', {
				description: error.response?.data?.error || 'Unable to apply adjustment'
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
			toast.success('Transfer queued');
		} catch (error) {
			toast.error('Failed to Create Transfer', {
				description: error.response?.data?.error || 'Unable to create transfer'
			});
		}
	};

	const runBarcodeLookup = async () => {
		if (!barcodeLookup.value) return toast.warning('Provide a barcode value');
		try {
			const product = await barcodeApi.lookup(barcodeLookup.value);
			barcodeLookup.result = product;
			toast.success('SKU resolved');
		} catch (error) {
			toast.error('Failed to Lookup Barcode', {
				description: error.response?.data?.error || 'Product not found'
			});
		}
	};

	const generateBarcode = async () => {
		if (!barcodeLookup.value) return toast.warning('Provide SKU or product ID');
		try {
			const blob = await barcodeApi.generate({ sku: barcodeLookup.value });
			barcodeImage = URL.createObjectURL(blob);
		} catch (error) {
			toast.error('Failed to Generate Barcode', {
				description: error.response?.data?.error || 'Unable to generate barcode'
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

<!-- HERO SECTION -->
<section
	class="animate-gradientShift relative w-full overflow-hidden bg-gradient-to-r from-sky-50 via-blue-50 to-cyan-100 px-6 py-20 text-center"
>
	<div class="absolute inset-0 bg-white/40 backdrop-blur-sm"></div>

	<div
		class="parallax-hero relative z-10 mx-auto flex max-w-3xl transform flex-col items-center justify-center space-y-4 transition-transform duration-700 ease-out will-change-transform"
	>
		<div
			class="animate-pulseGlow rounded-2xl bg-gradient-to-br from-sky-400 to-blue-500 p-4 shadow-lg"
		>
			<Activity class="h-8 w-8 text-white" />
		</div>
		<h1
			class="animate-fadeUp bg-gradient-to-r from-sky-600 via-blue-600 to-cyan-600 bg-clip-text text-4xl font-bold text-transparent sm:text-5xl"
		>
			Stock Adjustments, Transfers & Barcode Intelligence
		</h1>
		<p class="animate-fadeUp text-base text-slate-600 delay-200">
			Unified real-time control for stock, movement & labeling.
		</p>
	</div>
</section>

<!-- MAIN CONTENT -->
<section class="mx-auto max-w-7xl space-y-10 bg-white px-6 py-14">
	<div class="grid gap-8 lg:grid-cols-2">
		<!-- Inventory Snapshot -->
		<Card
			class="rounded-2xl border-0 bg-gradient-to-br from-sky-50 to-blue-100 shadow-lg transition-all duration-300 hover:scale-[1.02] hover:shadow-xl"
		>
			<CardHeader
				class="rounded-t-2xl border-b border-white/60 bg-white/80 px-6 py-5 backdrop-blur"
			>
				<CardTitle class="flex items-center gap-2 text-slate-800">
					<ClipboardCheck class="h-5 w-5 text-sky-600" />
					Inventory Snapshot
				</CardTitle>
				<CardDescription class="text-slate-600"
					>View product balance and batch details</CardDescription
				>
			</CardHeader>
			<CardContent class="space-y-4 p-6">
				<div class="grid gap-3 sm:grid-cols-2">
					<ProductSelector
						bind:value={stockQuery.productId}
						placeholder="Search product..."
						className="w-full"
						onSelect={() => setTimeout(loadStock, 100)}
					/>
					<Input
						type="number"
						placeholder="Location ID (optional)"
						bind:value={stockQuery.locationId}
						class="rounded-xl border-sky-200 bg-white/90 focus:ring-2 focus:ring-sky-400"
						onkeydown={(e) => e.key === 'Enter' && loadStock()}
					/>
				</div>
				<Button
					class="w-full rounded-xl bg-gradient-to-r from-sky-500 to-blue-600 font-semibold text-white shadow-md transition-all hover:scale-[1.02] hover:from-sky-600 hover:to-blue-700 hover:shadow-lg"
					onclick={loadStock}
				>
					Fetch stock levels
				</Button>
				{#if stockLoading}
					<Skeleton class="h-32 w-full bg-white/70" />
				{:else if stockSnapshot}
					<div class="rounded-2xl border border-sky-200 bg-white/80 p-4 shadow-sm backdrop-blur">
						<p class="text-sm text-slate-500">Current quantity</p>
						<p class="text-3xl font-semibold text-sky-700">{stockSnapshot.currentQuantity}</p>
					</div>
					<Table class="mt-4 overflow-hidden rounded-xl border border-sky-100">
						<TableHeader class="bg-gradient-to-r from-sky-100 to-blue-100">
							<TableRow>
								<TableHead>Batch</TableHead>
								<TableHead>Qty</TableHead>
								<TableHead>Expiry</TableHead>
							</TableRow>
						</TableHeader>
						<TableBody class="[&>tr:nth-child(even)]:bg-white/70 [&>tr:nth-child(odd)]:bg-white/50">
							{#if stockSnapshot.batches.length === 0}
								<TableRow>
									<TableCell colspan="3" class="py-4 text-center text-sm text-slate-500"
										>No batch detail available</TableCell
									>
								</TableRow>
							{:else}
								{#each stockSnapshot.batches as batch}
									<TableRow class="transition-colors hover:bg-white/90">
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

		<!-- Manual Adjustment -->
		{#if auth.hasPermission('products.write')}
			<Card
				class="rounded-2xl border-0 bg-gradient-to-br from-emerald-50 to-green-100 shadow-lg transition-all duration-300 hover:scale-[1.02] hover:shadow-xl"
			>
				<CardHeader
					class="rounded-t-2xl border-b border-white/60 bg-white/80 px-6 py-5 backdrop-blur"
				>
					<CardTitle class="flex items-center gap-2 text-slate-800">
						<Activity class="h-5 w-5 text-emerald-600" />
						Manual Adjustment
					</CardTitle>
					<CardDescription class="text-slate-600"
						>Perform adhoc cycle counts or receipts</CardDescription
					>
				</CardHeader>
				<CardContent class="space-y-4 p-6">
					<ProductSelector
						bind:value={adjustmentForm.productId}
						placeholder="Select product to adjust..."
						className="w-full border-emerald-200"
					/>
					<div class="grid grid-cols-2 gap-3">
						<select
							class="w-full rounded-xl border border-emerald-200 bg-white/90 px-3 py-2.5 text-sm focus:ring-2 focus:ring-emerald-400"
							bind:value={adjustmentForm.type}
						>
							<option value="STOCK_IN">Stock In (+)</option>
							<option value="STOCK_OUT">Stock Out (-)</option>
						</select>
						<Input
							type="number"
							placeholder="Quantity"
							bind:value={adjustmentForm.quantity}
							class="rounded-xl border-emerald-200 bg-white/90 focus:ring-2 focus:ring-emerald-400"
						/>
					</div>
					<Input
						placeholder="Reason code"
						bind:value={adjustmentForm.reasonCode}
						class="rounded-xl border-emerald-200 bg-white/90 focus:ring-2 focus:ring-emerald-400"
					/>
					<Input
						placeholder="Notes"
						bind:value={adjustmentForm.notes}
						class="rounded-xl border-emerald-200 bg-white/90 focus:ring-2 focus:ring-emerald-400"
					/>
					<Button
						class="w-full rounded-xl bg-gradient-to-r from-emerald-500 to-green-600 font-semibold text-white shadow-md transition-all hover:scale-105 hover:from-emerald-600 hover:to-green-700 hover:shadow-lg"
						onclick={submitAdjustment}>Apply adjustment</Button
					>
				</CardContent>
			</Card>
		{/if}
	</div>

	<div class="grid gap-8 lg:grid-cols-2">
		<!-- Stock Transfer -->
		{#if auth.hasPermission('products.write')}
			<Card
				class="rounded-2xl border-0 bg-gradient-to-br from-violet-50 to-purple-100 shadow-lg transition-all duration-300 hover:scale-[1.02] hover:shadow-xl"
			>
				<CardHeader
					class="rounded-t-2xl border-b border-white/60 bg-white/80 px-6 py-5 backdrop-blur"
				>
					<CardTitle class="flex items-center gap-2 text-slate-800">
						<ArrowRightLeft class="h-5 w-5 text-violet-600" />
						Stock Transfer
					</CardTitle>
					<CardDescription class="text-slate-600">Move inventory across locations</CardDescription>
				</CardHeader>
				<CardContent class="space-y-4 p-6">
					<ProductSelector
						bind:value={transferForm.productId}
						placeholder="Select product to transfer..."
						className="w-full border-violet-200"
					/>
					<div class="grid grid-cols-2 gap-3">
						<Input
							type="number"
							placeholder="Source location"
							bind:value={transferForm.sourceLocationId}
							class="rounded-xl border-violet-200 bg-white/90 focus:ring-2 focus:ring-violet-400"
						/>
						<Input
							type="number"
							placeholder="Destination location"
							bind:value={transferForm.destLocationId}
							class="rounded-xl border-violet-200 bg-white/90 focus:ring-2 focus:ring-violet-400"
						/>
					</div>
					<Input
						type="number"
						placeholder="Quantity"
						bind:value={transferForm.quantity}
						class="rounded-xl border-violet-200 bg-white/90 focus:ring-2 focus:ring-violet-400"
					/>
					<Button
						class="w-full rounded-xl bg-gradient-to-r from-violet-500 to-purple-600 font-semibold text-white shadow-md transition-all hover:scale-105 hover:from-violet-600 hover:to-purple-700 hover:shadow-lg"
						onclick={submitTransfer}>Create transfer</Button
					>
				</CardContent>
			</Card>
		{/if}

		<!-- Barcode Intelligence -->
		<Card
			class="rounded-2xl border-0 bg-gradient-to-br from-amber-50 to-orange-100 shadow-lg transition-all duration-300 hover:scale-[1.02] hover:shadow-xl"
		>
			<CardHeader
				class="rounded-t-2xl border-b border-white/60 bg-white/80 px-6 py-5 backdrop-blur"
			>
				<CardTitle class="flex items-center gap-2 text-slate-800">
					<ScanLine class="h-5 w-5 text-amber-600" />
					Barcode Intelligence
				</CardTitle>
				<CardDescription class="text-slate-600"
					>Lookup and generate barcodes for SKUs</CardDescription
				>
			</CardHeader>
			<CardContent class="space-y-4 p-6">
				<Input
					placeholder="Scan or type barcode / SKU"
					bind:value={barcodeLookup.value}
					class="rounded-xl border-amber-200 bg-white/90 focus:ring-2 focus:ring-amber-400"
					onkeydown={(e) => e.key === 'Enter' && runBarcodeLookup()}
				/>
				<div class="flex flex-col gap-3 sm:flex-row">
					<Button
						class="flex-1 rounded-xl border border-amber-200 bg-white/80 font-medium text-amber-700 shadow-sm transition-all hover:scale-105 hover:bg-amber-50"
						variant="secondary"
						onclick={runBarcodeLookup}
					>
						Lookup Product
					</Button>

					<Button
						class="flex-1 rounded-xl bg-gradient-to-r from-amber-500 to-orange-600 font-semibold text-white shadow-md transition-all hover:scale-105 hover:from-amber-600 hover:to-orange-700 hover:shadow-lg"
						onclick={generateBarcode}
					>
						Generate Image
					</Button>
				</div>

				{#if barcodeLookup.result}
					<div
						class="rounded-xl border border-amber-200 bg-white/70 p-3 text-sm shadow-sm backdrop-blur"
					>
						<p class="font-semibold text-slate-700">{barcodeLookup.result.Name}</p>
						<p class="text-slate-500">SKU: {barcodeLookup.result.SKU}</p>
					</div>
				{/if}
				{#if barcodeImage}
					<div
						class="rounded-xl border border-dashed border-amber-200 bg-white/70 p-3 text-center shadow-sm backdrop-blur"
					>
						<img src={barcodeImage} alt="Barcode preview" class="mx-auto" />
					</div>
				{/if}
			</CardContent>
		</Card>
	</div>
</section>

<style lang="postcss">
	@keyframes gradientShift {
		0% {
			background-position: 0% 50%;
		}
		50% {
			background-position: 100% 50%;
		}
		100% {
			background-position: 0% 50%;
		}
	}
	.animate-gradientShift {
		background-size: 200% 200%;
		animation: gradientShift 20s ease infinite;
	}

	@keyframes pulseGlow {
		0%,
		100% {
			transform: scale(1);
			box-shadow: 0 0 15px rgba(56, 189, 248, 0.3);
		}
		50% {
			transform: scale(1.08);
			box-shadow: 0 0 25px rgba(56, 189, 248, 0.5);
		}
	}
	.animate-pulseGlow {
		animation: pulseGlow 8s ease-in-out infinite;
	}

	@keyframes fadeUp {
		from {
			opacity: 0;
			transform: translateY(20px);
		}
		to {
			opacity: 1;
			transform: translateY(0);
		}
	}
	.animate-fadeUp {
		animation: fadeUp 1.5s ease forwards;
	}

	* {
		transition-property:
			color, background-color, border-color, text-decoration-color, fill, stroke, opacity,
			box-shadow, transform, filter, backdrop-filter;
		transition-timing-function: cubic-bezier(0.4, 0, 0.2, 1);
		transition-duration: 300ms;
	}
</style>
