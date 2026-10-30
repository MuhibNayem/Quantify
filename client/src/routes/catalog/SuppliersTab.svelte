<script lang="ts">
	import { toast } from 'svelte-sonner';
	import { Input } from '$lib/components/ui/input';
	import { Button } from '$lib/components/ui/button';
	import EntityTable from './components/EntityTable.svelte';
	import SupplierForm from './components/SupplierForm.svelte';
	import { suppliersApi } from '$lib/api/resources';
	import type { Supplier } from '$lib/types';
	import type { CatalogTab } from '$lib/catalog/detailSections';
	import { auth } from '$lib/stores/auth';
	import { t } from '$lib/i18n';

	type Props = {
		refreshToken?: number;
		onViewDetails: (resource: Record<string, any>, type: CatalogTab) => void;
	};

	let { refreshToken = 0, onViewDetails }: Props = $props();

	let loading = $state(false);
	let suppliers = $state<Supplier[]>([]);
	let editingSupplier = $state<Supplier | null>(null);
	let supplierSearchTerm = $state('');

	const inputClass =
		'glass-input h-10 w-full px-3 text-sm text-foreground placeholder:text-muted-foreground';
	const btnPrimary = 'btn-soothe h-10 px-4 text-sm font-semibold';

	const supplierColumns = $derived([
		{ header: $t('catalog.suppliers.form.name'), accessorKey: 'Name' },
		{ header: $t('catalog.suppliers.columns.contact'), accessorKey: 'ContactPerson' },
		...(auth.hasPermission('suppliers.write')
			? [{ header: $t('catalog.common.actions'), accessorKey: 'id', class: 'text-right' }]
			: [])
	]);

	const loadSuppliers = async () => {
		loading = true;
		try {
			const supplierList = await suppliersApi.list();
			suppliers = Array.isArray(supplierList) ? supplierList : [supplierList];
		} catch (error: any) {
			const errorMessage = error.response?.data?.error || 'Unable to load suppliers';
			toast.error('Failed to Load Catalog', { description: errorMessage });
		} finally {
			loading = false;
		}
	};

	const handleSupplierSearch = async () => {
		if (!supplierSearchTerm.trim()) return;
		loading = true;
		try {
			const supplier = await suppliersApi.getByName(supplierSearchTerm.trim());
			if (supplier) {
				onViewDetails(supplier, 'suppliers');
			} else {
				toast.error('Supplier not found');
			}
		} catch (error: any) {
			const errorMessage = error.response?.data?.error || 'Search failed';
			toast.error('Search Failed', { description: errorMessage });
		} finally {
			loading = false;
		}
	};

	const clearSupplierSearch = () => {
		supplierSearchTerm = '';
		loadSuppliers();
	};

	const deleteSupplier = async (supplier: Supplier) => {
		if (!window.confirm($t('catalog.toasts.confirm_delete', { name: supplier.Name }))) return;
		try {
			await suppliersApi.remove(supplier.ID);
			toast.success($t('catalog.toasts.supplier_removed'));
			await loadSuppliers();
		} catch (error: any) {
			const errorMessage = error.response?.data?.error || 'Unable to delete supplier';
			toast.error($t('catalog.toasts.supplier_remove_fail'), { description: errorMessage });
		}
	};

	$effect(() => {
		if (refreshToken >= 0) loadSuppliers();
	});
</script>

<div
	class="grid gap-6 {auth.hasPermission('suppliers.write')
		? 'lg:grid-cols-[minmax(0,2fr)_minmax(0,1fr)]'
		: 'grid-cols-1'}"
>
	<div class="flex min-w-0 flex-col gap-4">
		<div class="panel p-4">
			<div class="flex flex-col gap-4 lg:flex-row lg:items-center lg:justify-between">
				<div class="min-w-0 space-y-1">
					<h2 class="section-title">{$t('catalog.suppliers.title')}</h2>
					<p class="text-sm text-muted-foreground">{$t('catalog.suppliers.subtitle')}</p>
				</div>
				<div class="flex flex-col gap-3 sm:flex-row sm:items-center">
					<Input
						class="{inputClass} sm:w-52"
						placeholder={$t('catalog.suppliers.search_placeholder')}
						bind:value={supplierSearchTerm}
					/>
					<div class="flex items-center gap-2">
						<Button class={btnPrimary} onclick={handleSupplierSearch}
							>{$t('catalog.common.search')}</Button
						>
						<Button
							variant="ghost"
							class="h-10 px-3 text-sm text-muted-foreground hover:bg-muted hover:text-foreground"
							onclick={clearSupplierSearch}>{$t('catalog.common.clear')}</Button
						>
					</div>
				</div>
			</div>
		</div>

		<EntityTable
			data={suppliers}
			columns={supplierColumns}
			{loading}
			onRowClick={(supplier) => onViewDetails(supplier, 'suppliers')}
		>
			{#snippet children(supplier)}
				<td class="font-medium text-foreground">{supplier.Name}</td>
				<td>
					<p class="text-sm font-medium text-foreground">{supplier.ContactPerson ?? '—'}</p>
					<p class="text-xs text-muted-foreground">{supplier.Email ?? supplier.Phone ?? ''}</p>
				</td>
				{#if auth.hasPermission('suppliers.write')}
					<td class="text-right">
						<div class="flex items-center justify-end gap-1">
							<Button
								size="sm"
								variant="ghost"
								class="h-8 px-2.5 text-muted-foreground hover:bg-muted hover:text-foreground"
								onclick={(event) => {
									event.stopPropagation();
									editingSupplier = supplier;
								}}
							>
								{$t('catalog.common.edit')}
							</Button>
							<Button
								size="sm"
								variant="ghost"
								class="h-8 px-2.5 text-danger hover:bg-danger/10 hover:text-danger"
								onclick={(event) => {
									event.stopPropagation();
									deleteSupplier(supplier);
								}}
							>
								{$t('catalog.common.delete')}
							</Button>
						</div>
					</td>
				{/if}
			{/snippet}
		</EntityTable>
	</div>

	{#if auth.hasPermission('suppliers.write')}
		<SupplierForm editing={editingSupplier} onsaved={() => loadSuppliers()} />
	{/if}
</div>
