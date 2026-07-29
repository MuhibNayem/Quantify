<script lang="ts">
	import { toast } from 'svelte-sonner';
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
	import EditableCell from '$lib/components/ui/editable-cell.svelte';
	import { saveView, loadView } from '$lib/saved-views';
	import { deleteWithUndo } from '$lib/undo';
	import { settings } from '$lib/stores/settings';

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

	const selectedCount = $derived(selectedIds.size);
	const allSelected = $derived(products.length > 0 && selectedIds.size === products.length);
	const someSelected = $derived(selectedIds.size > 0);

	const inputClass =
		'liquid-input h-10 w-full rounded-md bg-card bg-none px-3 text-sm font-normal text-foreground shadow-none ring-0 backdrop-blur-none backdrop-saturate-100 from-card to-card placeholder:text-muted-foreground focus:scale-100 focus:border-primary focus:ring-0 md:text-sm';
	const btnPrimary =
		'rounded-md bg-primary bg-none px-4 text-sm font-medium text-primary-foreground shadow-none hover:bg-primary/90 hover:bg-none';
	const btnSecondary =
		'rounded-md border border-border bg-card bg-none text-foreground shadow-none backdrop-blur-none hover:bg-muted hover:bg-none';

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
		{ header: $t('catalog.products.columns.selling_price'), accessorKey: 'SellingPrice', class: 'text-right' },
		{ header: $t('catalog.products.columns.purchase_price'), accessorKey: 'PurchasePrice', class: 'text-right' },
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

	const canEditPrices = $derived(auth.hasPermission('products.write'));

	function formatMoney(amount?: number): string {
		return new Intl.NumberFormat($settings.locale, {
			minimumFractionDigits: 2,
			maximumFractionDigits: 2
		}).format(amount ?? 0);
	}

	async function savePrice(
		product: Product,
		field: 'SellingPrice' | 'PurchasePrice',
		newValue: number
	) {
		const previous = product[field];
		product[field] = newValue;
		try {
			await productsApi.update(product.ID, { [field]: newValue });
			toast.success(`${field === 'SellingPrice' ? 'Selling' : 'Purchase'} price updated`, {
				description: `${product.Name} → ${$settings.currency_symbol}${formatMoney(newValue)}`
			});
		} catch (error: any) {
			product[field] = previous;
			const errorMessage = error.response?.data?.error || 'Unable to update price';
			toast.error('Price update failed', { description: errorMessage });
			throw error;
		}
	}

	const deleteProduct = (product: Product) => {
		const snapshot = products;
		products = products.filter((p) => p.ID !== product.ID);
		deleteWithUndo(
			product.Name,
			async () => {
				await productsApi.remove(product.ID);
				await loadAll(pagination.currentPage);
			},
			() => {
				products = snapshot;
			}
		);
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
			class="liquid-panel flex flex-col gap-4 rounded-lg p-4 sm:flex-row sm:items-center sm:justify-between"
		>
			<div class="space-y-1">
				<h2 class="font-display text-foreground text-lg font-semibold">
					{$t('catalog.products.title')}
				</h2>
				<p class="text-muted-foreground text-sm">{$t('catalog.products.subtitle')}</p>
			</div>
			<div class="flex flex-col gap-3 sm:flex-row">
				<div class="relative flex-1">
					<Search class="text-muted-foreground absolute top-1/2 left-3 h-4 w-4 -translate-y-1/2" />
					<Input
						placeholder={$t('catalog.products.search_placeholder')}
						bind:value={search.term}
						onkeydown={(e) => e.key === 'Enter' && handleSearch()}
						class="{inputClass} pl-9"
					/>
				</div>
				<div class="flex gap-2">
					<Button variant="secondary" onclick={handleSearch} class={btnSecondary}>
						{$t('catalog.common.search')}
					</Button>
					{#if auth.hasPermission('products.write')}
						<Button onclick={() => (editingProduct = null)} class={btnPrimary}>
							<PlusCircle class="mr-2 h-4 w-4" />
							{$t('catalog.products.add_button')}
						</Button>
					{/if}
				</div>
			</div>
		</div>

		{#if selectedCount > 0}
			<div
				class="flex items-center gap-3 rounded-lg border border-primary/25 bg-accent px-4 py-3 shadow-sm"
			>
				<span class="text-accent-foreground text-sm font-semibold">{selectedCount} selected</span>
				<div class="ml-auto flex items-center gap-2">
					{#if auth.hasPermission('products.write')}
						<Button size="sm" variant="outline" onclick={() => (bulkEditOpen = true)} class="{btnSecondary} h-9">
							<Pencil class="mr-1.5 h-3.5 w-3.5" />
							Edit
						</Button>
					{/if}
					<Button size="sm" variant="outline" onclick={bulkExport} class="{btnSecondary} h-9">
						<Download class="mr-1.5 h-3.5 w-3.5" />
						Export
					</Button>
					{#if auth.hasPermission('products.delete')}
						<Button
							size="sm"
							variant="destructive"
							onclick={() => (bulkDeleteOpen = true)}
							class="rounded-md"
						>
							<Trash2 class="mr-1.5 h-3.5 w-3.5" />
							Delete
						</Button>
					{/if}
					<Button
						size="sm"
						variant="ghost"
						onclick={() => (selectedIds = new Set())}
						class="text-muted-foreground hover:bg-muted hover:text-foreground"
						>Clear</Button
					>
				</div>
			</div>
		{/if}

		{#if !loading && products.length === 0}
			<div class="liquid-panel rounded-lg">
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
					{@const badge = statusBadge(product.Status)}
					<td class="w-12" onclick={(e) => e.stopPropagation()}>
						<Checkbox
							checked={selectedIds.has(product.ID)}
							onCheckedChange={() => toggleSelect(product.ID)}
						/>
					</td>
					<td class="num-mono text-muted-foreground text-xs">{product.SKU}</td>
					<td class="text-foreground font-medium">{product.Name}</td>
					<td class="col-num" onclick={(e) => e.stopPropagation()}>
						{#if canEditPrices}
							<EditableCell
								type="number"
								value={product.SellingPrice ?? 0}
								display={formatMoney(product.SellingPrice)}
								prefix={$settings.currency_symbol}
								onSave={(v) => savePrice(product, 'SellingPrice', Number(v))}
							/>
						{:else}
							<span class="text-foreground text-sm font-semibold">
								<span class="text-muted-foreground mr-0.5 text-xs font-medium">
									{$settings.currency_symbol}
								</span>
								{formatMoney(product.SellingPrice)}
							</span>
						{/if}
					</td>
					<td class="col-num" onclick={(e) => e.stopPropagation()}>
						{#if canEditPrices}
							<EditableCell
								type="number"
								value={product.PurchasePrice ?? 0}
								display={formatMoney(product.PurchasePrice)}
								prefix={$settings.currency_symbol}
								onSave={(v) => savePrice(product, 'PurchasePrice', Number(v))}
							/>
						{:else}
							<span class="text-foreground text-sm font-semibold">
								<span class="text-muted-foreground mr-0.5 text-xs font-medium">
									{$settings.currency_symbol}
								</span>
								{formatMoney(product.PurchasePrice)}
							</span>
						{/if}
					</td>
					<td>
						{#if badge}
							<span class="badge-{badge.variant} capitalize">
								{badge.text}
							</span>
						{:else}
							<span class="text-muted-foreground">—</span>
						{/if}
					</td>
					{#if auth.hasPermission('products.write') || auth.hasPermission('products.delete')}
						<td class="text-right">
							<div class="flex items-center justify-end gap-1">
								{#if auth.hasPermission('products.write')}
									<Button
										size="sm"
										variant="ghost"
										class="h-8 text-primary hover:bg-muted hover:text-foreground"
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
										class="h-8 text-danger hover:bg-danger/10 hover:text-danger"
										onclick={(event) => {
											event.stopPropagation();
											deleteProduct(product);
										}}
									>
										{$t('catalog.common.delete')}
									</Button>
								{/if}
							</div>
						</td>
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
	bind:open={bulkDeleteOpen}
	title="Delete selected products"
	description={`This will permanently delete ${selectedCount} product(s). This action cannot be undone.`}
	confirmLabel={bulkDeleting ? `Deleting ${bulkDeleteProgress}/${selectedCount}...` : 'Delete all'}
	onConfirm={bulkDelete}
/>

{#if bulkEditOpen}
	<div class="fixed inset-0 z-50 flex items-center justify-center">
		<button
			class="bg-black/40 absolute inset-0"
			onclick={() => (bulkEditOpen = false)}
			aria-label="Close"
		></button>
		<div
			class="bg-card border-border relative z-10 w-full max-w-sm space-y-4 rounded-lg border p-6 shadow-lg"
		>
			<h3 class="text-foreground text-lg font-semibold">Bulk edit ({selectedCount} selected)</h3>
			<div class="space-y-3">
				<div class="space-y-1">
					<label class="text-muted-foreground text-sm font-medium" for="bulk-category">Category</label>
					<div class="select-wrapper">
						<select id="bulk-category" bind:value={bulkEditCategory}>
							<option value="">No change</option>
							{#each categories as cat}
								<option value={cat.ID}>{cat.Name}</option>
							{/each}
						</select>
					</div>
				</div>
				<div class="space-y-1">
					<label class="text-muted-foreground text-sm font-medium" for="bulk-status">Status</label>
					<div class="select-wrapper">
						<select id="bulk-status" bind:value={bulkEditStatus}>
							<option value="">No change</option>
							<option value="active">Active</option>
							<option value="inactive">Inactive</option>
						</select>
					</div>
				</div>
			</div>
			<div class="flex justify-end gap-2">
				<Button
					variant="outline"
					onclick={() => (bulkEditOpen = false)}
					class="rounded-md border-border bg-card bg-none text-foreground shadow-none backdrop-blur-none hover:bg-muted"
					>Cancel</Button
				>
				<Button onclick={bulkEditApply} class={btnPrimary}>Apply</Button>
			</div>
		</div>
	</div>
{/if}
