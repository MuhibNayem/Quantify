<!-- client/src/routes/products/[id]/+page.svelte -->
<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { toast } from 'svelte-sonner';
	import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '$lib/components/ui/card';
	import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '$lib/components/ui/table';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import { Button } from '$lib/components/ui/button';
	import { ArrowLeft, Info, History } from 'lucide-svelte';
	import type { PageData } from './$types';

	export let data: PageData;

	const { product, stockHistory } = data;
</script>

<div class="w-full max-w-7xl mx-auto py-8 px-6">
	<section class="space-y-8">
		<!-- HEADER -->
		<div class="flex items-center justify-between">
			<a href="/catalog" class="flex items-center text-sky-600 hover:text-sky-800 transition-colors">
				<ArrowLeft class="h-5 w-5 mr-2" />
				Back to Catalog
			</a>
		</div>

		{#if product}
			<div class="grid gap-8 lg:grid-cols-3">
				<!-- Product Details -->
				<div class="lg:col-span-2">
					<Card class="rounded-2xl shadow-lg hover:shadow-xl transition-all duration-300 hover:scale-[1.01] overflow-hidden border-0 bg-gradient-to-br from-sky-50 to-blue-100">
						<CardHeader class="space-y-1 bg-white/70 backdrop-blur px-6 py-5 border-b border-white/60">
							<CardTitle class="text-slate-800 flex items-center">
								<Info class="h-5 w-5 mr-2 text-sky-600" />
								{product.Name}
							</CardTitle>
							<CardDescription class="text-slate-600">{product.SKU}</CardDescription>
						</CardHeader>
						<CardContent class="p-6 space-y-4">
							<div class="grid grid-cols-1 sm:grid-cols-2 gap-4 text-sm">
								<div>
									<p class="font-medium text-slate-500">Description</p>
									<p class="text-slate-800">{product.Description || 'N/A'}</p>
								</div>
								<div>
									<p class="font-medium text-slate-500">Brand</p>
									<p class="text-slate-800">{product.Brand || 'N/A'}</p>
								</div>
								<div>
									<p class="font-medium text-slate-500">Category</p>
									<p class="text-slate-800">{product.Category?.Name || 'N/A'}</p>
								</div>
								<div>
									<p class="font-medium text-slate-500">Sub-Category</p>
									<p class="text-slate-800">{product.SubCategory?.Name || 'N/A'}</p>
								</div>
								<div>
									<p class="font-medium text-slate-500">Supplier</p>
									<p class="text-slate-800">{product.Supplier?.Name || 'N/A'}</p>
								</div>
								<div>
									<p class="font-medium text-slate-500">Location</p>
									<p class="text-slate-800">{product.Location?.Name || 'N/A'}</p>
								</div>
								<div>
									<p class="font-medium text-slate-500">Purchase Price</p>
									<p class="text-slate-800">${product.PurchasePrice?.toFixed(2) || 'N/A'}</p>
								</div>
								<div>
									<p class="font-medium text-slate-500">Selling Price</p>
									<p class="text-slate-800">${product.SellingPrice?.toFixed(2) || 'N/A'}</p>
								</div>
								<div>
									<p class="font-medium text-slate-500">Barcode/UPC</p>
									<p class="text-slate-800 font-mono">{product.BarcodeUPC || 'N/A'}</p>
								</div>
								<div>
									<p class="font-medium text-slate-500">Status</p>
									<span class="rounded-full bg-sky-100 text-sky-700 px-2.5 py-0.5 text-xs capitalize border border-sky-200 shadow-sm">
										{product.Status || 'active'}
									</span>
								</div>
							</div>
						</CardContent>
					</Card>
				</div>

				<!-- Stock History -->
				<div class="lg:col-span-1">
					<Card class="rounded-2xl shadow-lg hover:shadow-xl transition-all duration-300 hover:scale-[1.01] overflow-hidden border-0 bg-gradient-to-br from-emerald-50 to-green-100">
						<CardHeader class="space-y-1 bg-white/70 backdrop-blur px-6 py-5 border-b border-white/60">
							<CardTitle class="text-slate-800 flex items-center">
								<History class="h-5 w-5 mr-2 text-emerald-600" />
								Stock History
							</CardTitle>
							<CardDescription class="text-slate-600">Recent stock adjustments</CardDescription>
						</CardHeader>
						<CardContent class="p-0">
							<Table>
								<TableHeader>
									<TableRow>
										<TableHead>Type</TableHead>
										<TableHead>Qty</TableHead>
										<TableHead>Date</TableHead>
									</TableRow>
								</TableHeader>
								<TableBody>
									{#if stockHistory && stockHistory.length > 0}
										{#each stockHistory as item}
											<TableRow>
												<TableCell>{item.Type}</TableCell>
												<TableCell class="{item.Type === 'STOCK_IN' ? 'text-green-600' : 'text-red-600'}">
													{item.Type === 'STOCK_IN' ? '+' : '-'}{item.Quantity}
												</TableCell>
												<TableCell>{new Date(item.AdjustedAt).toLocaleDateString()}</TableCell>
											</TableRow>
										{/each}
									{:else}
										<TableRow>
											<TableCell colspan="3" class="text-center text-slate-500 py-4">
												No stock history found.
											</TableCell>
										</TableRow>
									{/if}
								</TableBody>
							</Table>
						</CardContent>
					</Card>
				</div>
			</div>
		{:else}
			<!-- Loading State -->
			<div class="grid gap-8 lg:grid-cols-3">
				<div class="lg:col-span-2">
					<Skeleton class="h-96 w-full" />
				</div>
				<div class="lg:col-span-1">
					<Skeleton class="h-96 w-full" />
				</div>
			</div>
		{/if}
	</section>
</div>
