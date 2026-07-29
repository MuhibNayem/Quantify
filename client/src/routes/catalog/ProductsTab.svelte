<script lang="ts">
	import { toast } from 'svelte-sonner';
	import { TableCell } from '$lib/components/ui/table';
	import { Input } from '$lib/components/ui/input';
	import { Button } from '$lib/components/ui/button';
	import { Checkbox } from '$lib/components/ui/checkbox';
	import EntityTable from './components/EntityTable.svelte';
	import ProductForm from './components/ProductForm.svelte';
	import ConfirmDialog from '$lib/components/ui/confirm-dialog.svelte';
	import { productsApi, categoriesApi, suppliersApi, locationsApi } from '$lib/api/resources';
	import type { Category, Location, Product, Supplier } from '$lib/types';
	import type { CatalogTab } from '$lib/catalog/detailSections';
	import { statusBadge } from '$lib/catalog/detailSections';
	import { PlusCircle, Search, Package, Download, Pencil, Trash2 } from 'lucide-svelte';
	import { auth } from '$lib/stores/auth';
	import { t } from '$lib/i18n';
	import EmptyState from '$lib/components/ui/empty-state.svelte';
	import { saveView, loadView } from '$lib/saved-views';

	type Props = {
		refreshToken?: number;
		onViewDetails: (resource: Record<string, any>, type: CatalogTab) => void;
		onNavigate: (tab: CatalogTab) => void;
	};

	let { refreshToken = 0, onViewDetails, onNavigate }: Props = $props();

	let loading = $state(false);
	let search = $state(loadView('catalog:products', { term: '', key: 'name' }));

	let products = $state<Product[]>([]);
	let categories = $state<Category[]>([]);
	let suppliers = $state<Supplier[]>([]);
	let locations = $state<Location[]>([]);
	let editingProduct = $state<Product | null>(null);

	let selectedIds = $state<Set<number>>(new Set());
	let bulkDeleteOpen = $state(false);
	let bulkEditOpen = $state(false);
	let bulkEditCategory = $state('');
	let bulkEditStatus = $state('');
	let bulkDeleting = $state(false);
	let bulkDeleteProgress = $state(0);
	let deleteTarget = $state<Product | null>(null);
	let singleDeleteOpen = $state(false);

	const selectedCount = $derived(selectedIds.size);
	const allSelected = $derived(products.length > 0 && selectedIds.size === products.length);
	const someSelected = $derived(selectedIds.size > 0);

	function toggleSelect(id: number) {
		const next = new Set(selectedIds);
		if (next.has(id)) next.delete(id);
		else next.add(id);
		selectedIds = next;
	}

	function toggleSelectAll() {
		if (selectedIds.size === products.length) {
			selectedIds = new Set();
		} else {
			selectedIds = new Set(products.map((p) => p.ID));
		}
	}

	async function bulkDelete() {
		bulkDeleting = true;
		bulkDeleteProgress = 0;
		const ids = [...selectedIds];
		for (const id of ids) {
			try {
				await productsApi.remove(id);
			} catch {
				// continue
			}
			bulkDeleteProgress++;
		}
		bulkDeleting = false;
		bulkDeleteOpen = false;
		selectedIds = new Set();
		toast.success(`Deleted ${bulkDeleteProgress} product(s)`);
		await loadAll(pagination.currentPage);
	}

	function bulkExport() {
		const selected = products.filter((p) => selectedIds.has(p.ID));
		const headers = ['ID', 'SKU', 'Name', 'Status'];
		const rows = selected.map((p) => [p.ID, p.SKU, `"${p.Name}"`, p.Status]);
		const csv = [headers.join(','), ...rows.map((r) => r.join(','))].join('\n');
		const blob = new Blob([csv], { type: 'text/csv' });
		const url = URL.createObjectURL(blob);
		const a = document.createElement('a');
		a.href = url;
		a.download = 'products-export.csv';
		a.click();
		URL.revokeObjectURL(url);
		toast.success(`Exported ${selected.length} product(s)`);
	}

	async function bulkEditApply() {
		const ids = [...selectedIds];
		const payload: Record<string, unknown> = {};
		if (bulkEditCategory) payload.CategoryID = Number(bulkEditCategory);
		if (bulkEditStatus) payload.Status = bulkEditStatus;
		if (Object.keys(payload).length === 0) {
			bulkEditOpen = false;
			return;
		}
		for (const id of ids) {
			try {
				await productsApi.update(id, payload);
			} catch {
				// continue
			}
		}
		bulkEditOpen = false;
		bulkEditCategory = '';
		bulkEditStatus = '';
		selectedIds = new Set();
		toast.success(`Updated ${ids.length} product(s)`);
		await loadAll(pagination.currentPage);
	}

	$effect(() => {
		saveView('catalog:products', { term: search.term, key: search.key });
	});

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

	const deleteProduct = (product: Product) => {
		deleteTarget = product;
		singleDeleteOpen = true;
	};

	const confirmDeleteProduct = async () => {
		if (!deleteTarget) return;
		try {
			await productsApi.remove(deleteTarget.ID);
			toast.success($t('catalog.toasts.product_removed'));
			await loadAll(pagination.currentPage);
		} catch (error: any) {
			const errorMessage = error.response?.data?.error || 'Unable to delete product';
			toast.error($t('catalog.toasts.product_remove_fail'), { description: errorMessage });
		}
		deleteTarget = null;
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

		{#if selectedCount > 0}
			<div
				class="flex items-center gap-3 rounded-2xl border border-sky-200 bg-sky-50/80 px-4 py-3 shadow-sm backdrop-blur"
			>
				<span class="text-sm font-semibold text-sky-800">{selectedCount} selected</span>
				<div class="ml-auto flex items-center gap-2">
					{#if auth.hasPermission('products.write')}
						<Button size="sm" variant="outline" onclick={() => (bulkEditOpen = true)}>
							<Pencil class="mr-1.5 h-3.5 w-3.5" />
							Edit
						</Button>
					{/if}
					<Button size="sm" variant="outline" onclick={bulkExport}>
						<Download class="mr-1.5 h-3.5 w-3.5" />
						Export
					</Button>
					{#if auth.hasPermission('products.delete')}
						<Button size="sm" variant="destructive" onclick={() => (bulkDeleteOpen = true)}>
							<Trash2 class="mr-1.5 h-3.5 w-3.5" />
							Delete
						</Button>
					{/if}
					<Button size="sm" variant="ghost" onclick={() => (selectedIds = new Set())}>Clear</Button>
				</div>
			</div>
		{/if}

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
				selectable
				allSelected={allSelected}
				someSelected={someSelected}
				onSelectAll={toggleSelectAll}
			>
				{#snippet children(product)}
					<TableCell class="w-12 px-4" onclick={(e) => e.stopPropagation()}>
						<Checkbox
							checked={selectedIds.has(product.ID)}
							onCheckedChange={() => toggleSelect(product.ID)}
						/>
					</TableCell>
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

<ConfirmDialog
	bind:open={singleDeleteOpen}
	title="Delete product"
	description={deleteTarget
		? `This will permanently delete "${deleteTarget.Name}". This action cannot be undone.`
		: ''}
	confirmText={deleteTarget?.Name ?? ''}
	confirmLabel="Delete"
	onConfirm={confirmDeleteProduct}
/>

<ConfirmDialog
	bind:open={bulkDeleteOpen}
	title="Delete selected products"
	description={`This will permanently delete ${selectedCount} product(s). This action cannot be undone.`}
	confirmLabel={bulkDeleting ? `Deleting ${bulkDeleteProgress}/${selectedCount}...` : 'Delete all'}
	onConfirm={bulkDelete}
/>

{#if bulkEditOpen}
	<div class="fixed inset-0 z-50 flex items-center justify-center">
		<button
			class="absolute inset-0 bg-black/40 backdrop-blur-sm"
			onclick={() => (bulkEditOpen = false)}
			aria-label="Close"
		></button>
		<div
			class="relative z-10 w-full max-w-sm space-y-4 rounded-3xl border border-white/50 bg-white/90 p-6 shadow-2xl backdrop-blur-xl"
		>
			<h3 class="text-lg font-semibold text-slate-900">Bulk edit ({selectedCount} selected)</h3>
			<div class="space-y-3">
				<div class="space-y-1">
					<label class="text-sm font-medium text-slate-600" for="bulk-category">Category</label>
					<select
						id="bulk-category"
						bind:value={bulkEditCategory}
						class="w-full rounded-xl border border-slate-200 bg-white px-3 py-2 text-sm"
					>
						<option value="">No change</option>
						{#each categories as cat}
							<option value={cat.ID}>{cat.Name}</option>
						{/each}
					</select>
				</div>
				<div class="space-y-1">
					<label class="text-sm font-medium text-slate-600" for="bulk-status">Status</label>
					<select
						id="bulk-status"
						bind:value={bulkEditStatus}
						class="w-full rounded-xl border border-slate-200 bg-white px-3 py-2 text-sm"
					>
						<option value="">No change</option>
						<option value="active">Active</option>
						<option value="inactive">Inactive</option>
					</select>
				</div>
			</div>
			<div class="flex justify-end gap-2">
				<Button variant="outline" onclick={() => (bulkEditOpen = false)}>Cancel</Button>
				<Button onclick={bulkEditApply}>Apply</Button>
			</div>
		</div>
	</div>
{/if}
