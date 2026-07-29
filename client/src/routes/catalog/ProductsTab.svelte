<script lang="ts">
	import { toast } from 'svelte-sonner';
	import { TableCell } from '$lib/components/ui/table';
	import { Input } from '$lib/components/ui/input';
	import { Button } from '$lib/components/ui/button';
	import EntityTable from './components/EntityTable.svelte';
	import ProductForm from './components/ProductForm.svelte';
	import { productsApi, categoriesApi, suppliersApi, locationsApi } from '$lib/api/resources';
	import type { Category, Location, Product, Supplier } from '$lib/types';
	import type { CatalogTab } from '$lib/catalog/detailSections';
	import { statusBadge } from '$lib/catalog/detailSections';
	import { PlusCircle, Search, Package } from 'lucide-svelte';
	import { auth } from '$lib/stores/auth';
	import { t } from '$lib/i18n';
	import EmptyState from '$lib/components/ui/empty-state.svelte';

	type Props = {
		refreshToken?: number;
		onViewDetails: (resource: Record<string, any>, type: CatalogTab) => void;
		onNavigate: (tab: CatalogTab) => void;
	};

	let { refreshToken = 0, onViewDetails, onNavigate }: Props = $props();

	let loading = $state(false);
	let search = $state({ term: '', key: 'name' });

	let products = $state<Product[]>([]);
	let categories = $state<Category[]>([]);
	let suppliers = $state<Supplier[]>([]);
	let locations = $state<Location[]>([]);
	let editingProduct = $state<Product | null>(null);

	const pagination = $state({
		currentPage: 1,
		totalPages: 1,
		totalItems: 0,
		itemsPerPage: 10
	});

	const productColumns = $derived([
		{ header: $t('catalog.products.columns.sku'), accessorKey: 'SKU' },
		{ header: $t('catalog.products.columns.name'), accessorKey: 'Name' },
		{ header: $t('catalog.products.columns.status'), accessorKey: 'Status' },
		...(auth.hasPermission('products.write') || auth.hasPermission('products.delete')
			? [{ header: $t('catalog.common.actions'), accessorKey: 'id', class: 'text-right' }]
			: [])
	]);

	const loadAll = async (page = 1) => {
		loading = true;
		try {
			const categoryList = await categoriesApi.list();
			categories = Array.isArray(categoryList) ? categoryList : [categoryList];

			const supplierList = await suppliersApi.list();
			suppliers = Array.isArray(supplierList) ? supplierList : [supplierList];

			const locationList = await locationsApi.list();
			locations = Array.isArray(locationList) ? locationList : [locationList];

			await loadProductPerPage(page);
		} catch (error: any) {
			const errorMessage = error.response?.data?.error || 'Unable to load catalog';
			toast.error('Failed to Load Catalog', { description: errorMessage });
		} finally {
			loading = false;
		}
	};

	const loadProductPerPage = async (page = 1) => {
		loading = true;
		try {
			const productResponse = await productsApi.list(page, 10);
			products = productResponse.products || [];
			pagination.currentPage = productResponse.currentPage;
			pagination.totalPages = productResponse.totalPages;
			pagination.totalItems = productResponse.totalItems;
			pagination.itemsPerPage = productResponse.itemsPerPage;
		} catch (error: any) {
			const errorMessage = error.response?.data?.error || 'Unable to load catalog';
			toast.error('Failed to Load Catalog', { description: errorMessage });
		} finally {
			loading = false;
		}
	};

	const handlePageChange = (page: number) => {
		if (page !== pagination.currentPage) {
			pagination.currentPage = page;
			loadProductPerPage(page);
		}
	};

	const handleSearch = async () => {
		if (!search.term.trim()) return;
		loading = true;
		try {
			if (search.key === 'sku') {
				const product = await productsApi.getBySku(search.term.trim());
				if (product) {
					onViewDetails(product, 'products');
				} else {
					toast.error('Product not found');
				}
			} else {
				const productResponse = await productsApi.list(1, 100, search.term.trim());
				products = productResponse.products || [];
				pagination.currentPage = productResponse.currentPage;
				pagination.totalPages = productResponse.totalPages;
				pagination.totalItems = productResponse.totalItems;
				pagination.itemsPerPage = productResponse.itemsPerPage;
			}
		} catch (error: any) {
			const errorMessage = error.response?.data?.error || 'Search failed';
			toast.error('Search Failed', { description: errorMessage });
		} finally {
			loading = false;
		}
	};

	const editProduct = (product: Product) => {
		editingProduct = product;
	};

	const deleteProduct = async (product: Product) => {
		if (!window.confirm($t('catalog.toasts.confirm_delete', { name: product.Name }))) return;
		try {
			await productsApi.remove(product.ID);
			toast.success($t('catalog.toasts.product_removed'));
			await loadAll(pagination.currentPage);
		} catch (error: any) {
			const errorMessage = error.response?.data?.error || 'Unable to delete product';
			toast.error($t('catalog.toasts.product_remove_fail'), { description: errorMessage });
		}
	};

	$effect(() => {
		if (refreshToken >= 0) loadAll();
	});
</script>

<div
	class="grid gap-8 {auth.hasPermission('products.write')
		? 'lg:grid-cols-[2fr,1fr]'
		: 'grid-cols-1'}"
>
	<div class="flex flex-col gap-6">
		<div
			class="flex gap-4 rounded-2xl border border-sky-100 bg-white/50 p-4 shadow-sm backdrop-blur sm:flex-row sm:items-center sm:justify-between"
		>
			<div class="space-y-1">
				<h2 class="text-lg font-semibold text-slate-800">{$t('catalog.products.title')}</h2>
				<p class="text-sm text-slate-500">{$t('catalog.products.subtitle')}</p>
			</div>
			<div class="flex flex-col gap-3 sm:flex-row">
				<div class="relative flex-1">
					<Search class="absolute top-1/2 left-3 h-4 w-4 -translate-y-1/2 text-slate-400" />
					<Input
						placeholder={$t('catalog.products.search_placeholder')}
						bind:value={search.term}
						onkeydown={(e) => e.key === 'Enter' && handleSearch()}
						class="rounded-xl border-sky-200 bg-white/90 pl-9 focus:ring-2 focus:ring-sky-400"
					/>
				</div>
				<div class="flex gap-2">
					<Button
						variant="secondary"
						onclick={handleSearch}
						class="rounded-xl border border-sky-200 bg-white text-sky-700 hover:bg-sky-50"
					>
						{$t('catalog.common.search')}
					</Button>
					{#if auth.hasPermission('products.write')}
						<Button
							onclick={() => (editingProduct = null)}
							class="rounded-xl bg-gradient-to-r from-sky-500 to-indigo-600 text-white shadow-md transition-all hover:from-sky-600 hover:to-indigo-700 hover:shadow-lg"
						>
							<PlusCircle class="mr-2 h-4 w-4" />
							{$t('catalog.products.add_button')}
						</Button>
					{/if}
				</div>
			</div>
		</div>

		{#if !loading && products.length === 0}
			<div class="rounded-2xl border border-sky-100 bg-white/50 shadow-sm backdrop-blur">
				<EmptyState
					icon={Package}
					title="Add your first product"
					description="Products you add will appear here and be available at the POS."
					actionLabel={auth.hasPermission('products.write')
						? $t('catalog.products.add_button')
						: ''}
					onAction={() => (editingProduct = null)}
				/>
			</div>
		{:else}
			<EntityTable
				data={products}
				columns={productColumns}
				totalItems={pagination.totalItems}
				pageSize={pagination.itemsPerPage}
				currentPage={pagination.currentPage}
				onPageChange={handlePageChange}
				{loading}
				onRowClick={(product) => onViewDetails(product, 'products')}
			>
				{#snippet children(product)}
					<TableCell class="font-mono text-xs text-slate-600">{product.SKU}</TableCell>
					<TableCell class="font-medium text-slate-900">{product.Name}</TableCell>
					<TableCell>
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
					</TableCell>
					{#if auth.hasPermission('products.write') || auth.hasPermission('products.delete')}
						<TableCell class="text-right">
							<div class="flex items-center justify-end gap-1">
								{#if auth.hasPermission('products.write')}
									<Button
										size="sm"
										variant="ghost"
										class="h-8 text-sky-600 hover:bg-sky-50 hover:text-sky-700"
										onclick={(event) => {
											event.stopPropagation();
											editProduct(product);
										}}
									>
										{$t('catalog.common.edit')}
									</Button>
								{/if}
								{#if auth.hasPermission('products.delete')}
									<Button
										size="sm"
										variant="ghost"
										class="h-8 text-rose-600 hover:bg-rose-50 hover:text-rose-700"
										onclick={(event) => {
											event.stopPropagation();
											deleteProduct(product);
										}}
									>
										{$t('catalog.common.delete')}
									</Button>
								{/if}
							</div>
						</TableCell>
					{/if}
				{/snippet}
			</EntityTable>
		{/if}
	</div>

	{#if auth.hasPermission('products.write')}
		<ProductForm
			editing={editingProduct}
			{categories}
			{suppliers}
			{locations}
			{products}
			onsaved={() => loadAll(pagination.currentPage)}
			oncancel={() => (editingProduct = null)}
			{onNavigate}
		/>
	{/if}
</div>
