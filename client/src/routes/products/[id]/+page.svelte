<!-- client/src/routes/products/[id]/+page.svelte -->
<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { toast } from 'svelte-sonner';
	import {
		Card,
		CardContent,
		CardDescription,
		CardHeader,
		CardTitle
	} from '$lib/components/ui/card';
	import {
		Table,
		TableBody,
		TableCell,
		TableHead,
		TableHeader,
		TableRow
	} from '$lib/components/ui/table';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import { Button } from '$lib/components/ui/button';
	import { ArrowLeft, Info, History } from 'lucide-svelte';
	import DataTable from '$lib/components/ui/data-table/DataTable.svelte';
	import type { PageData } from './$types';

	export let data: PageData;

	$: ({ product, stockHistory } = data);

	const statusBadge = (status?: string) => {
		if (!status) return undefined;
		const normalized = status.toLowerCase();
		if (normalized === 'active') return { text: status, variant: 'success' as const };
		if (normalized === 'inactive') return { text: status, variant: 'warning' as const };
		if (normalized === 'archived') return { text: status, variant: 'danger' as const };
		return { text: status, variant: 'info' as const };
	};
</script>

<div class="mx-auto w-full max-w-7xl px-6 py-8">
	<section class="space-y-8">
		<!-- HEADER -->
		<div class="flex items-center justify-between">
			<a
				href="/catalog"
				class="flex items-center text-sky-600 transition-colors hover:text-sky-800"
			>
				<ArrowLeft class="mr-2 h-5 w-5" />
				Back to Catalog
			</a>
		</div>

		{#if product}
			<div class="grid gap-8 lg:grid-cols-3">
				<!-- Product Details -->
				<div class="lg:col-span-2">
					<Card
						class="overflow-hidden rounded-2xl border-0 bg-gradient-to-br from-sky-50 to-blue-100 shadow-lg transition-all duration-300 hover:scale-[1.01] hover:shadow-xl"
					>
						<CardHeader
							class="space-y-1 border-b border-white/60 bg-white/70 px-6 py-5 backdrop-blur"
						>
							<CardTitle class="flex items-center text-slate-800">
								<Info class="mr-2 h-5 w-5 text-sky-600" />
								{product.Name}
							</CardTitle>
							<CardDescription class="text-slate-600">{product.SKU}</CardDescription>
						</CardHeader>
						<CardContent class="space-y-4 p-6">
							<div class="grid grid-cols-1 gap-4 text-sm sm:grid-cols-2">
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
									<p class="font-mono text-slate-800">{product.BarcodeUPC || 'N/A'}</p>
								</div>
								<div>
									<p class="font-medium text-slate-500">Status</p>
									{#if product.Status}
										{@const badge = statusBadge(product.Status)}
										{#if badge}
											<span
												class="inline-flex items-center rounded-full border px-2.5 py-0.5 text-xs font-medium capitalize shadow-sm
								{badge.variant === 'success' ? 'border-emerald-200 bg-emerald-50 text-emerald-700' : ''}
								{badge.variant === 'warning' ? 'border-amber-200 bg-amber-50 text-amber-700' : ''}
								{badge.variant === 'danger' ? 'border-rose-200 bg-rose-50 text-rose-700' : ''}
								{badge.variant === 'info' ? 'border-sky-200 bg-sky-50 text-sky-700' : ''}"
											>
												{badge.text}
											</span>
										{:else}
											<span class="text-slate-400">—</span>
										{/if}
									{:else}
										<span class="text-slate-400">—</span>
									{/if}
								</div>
							</div>
						</CardContent>
					</Card>
				</div>

				<!-- Stock History -->
				<div class="lg:col-span-1">
					<div class="flex flex-col gap-4">
						<div
							class="flex flex-col gap-2 rounded-2xl border border-emerald-100 bg-white/50 p-4 shadow-sm backdrop-blur"
						>
							<h3 class="flex items-center text-lg font-semibold text-slate-800">
								<History class="mr-2 h-5 w-5 text-emerald-600" />
								Stock History
							</h3>
							<p class="text-sm text-slate-500">Recent stock adjustments</p>
						</div>

						<DataTable
							data={stockHistory || []}
							columns={[
								{ header: 'Type', accessorKey: 'Type' },
								{ header: 'Qty', accessorKey: 'Quantity' },
								{ header: 'Date', accessorKey: 'AdjustedAt' }
							]}
						>
							{#snippet children(item)}
								<TableCell class="font-medium text-slate-700">{item.Type}</TableCell>
								<TableCell
									class="font-semibold {item.Type === 'STOCK_IN'
										? 'text-emerald-600'
										: 'text-rose-600'}"
								>
									{item.Type === 'STOCK_IN' ? '+' : '-'}{item.Quantity}
								</TableCell>
								<TableCell class="text-slate-500">
									{new Date(item.AdjustedAt).toLocaleDateString()}
								</TableCell>
							{/snippet}
						</DataTable>
					</div>
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
