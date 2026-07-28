<script lang="ts">
	import { toast } from 'svelte-sonner';
	import { TableCell } from '$lib/components/ui/table';
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
	class="grid gap-8 {auth.hasPermission('suppliers.write')
		? 'lg:grid-cols-[2fr,1fr]'
		: 'grid-cols-1'}"
>
	<div class="flex flex-col gap-6">
		<div
			class="flex flex-col gap-4 rounded-2xl border border-violet-100 bg-white/50 p-4 shadow-sm backdrop-blur sm:flex-row sm:items-center sm:justify-between"
		>
			<div class="space-y-1">
				<h2 class="text-lg font-semibold text-slate-800">{$t('catalog.suppliers.title')}</h2>
				<p class="text-sm text-slate-500">{$t('catalog.suppliers.subtitle')}</p>
			</div>
			<div class="flex flex-wrap items-center gap-2">
				<Input
					class="w-full rounded-xl border-violet-200 bg-white/80 px-3 text-sm focus:ring-2 focus:ring-violet-400 sm:w-48"
					placeholder={$t('catalog.suppliers.search_placeholder')}
					bind:value={supplierSearchTerm}
				/>
				<Button
					class="rounded-xl bg-violet-500 px-4 py-2 text-white hover:bg-violet-600"
					onclick={handleSupplierSearch}>{$t('catalog.common.search')}</Button
				>
				<Button
					variant="ghost"
					class="rounded-xl px-4 py-2 text-violet-600 hover:bg-violet-50"
					onclick={clearSupplierSearch}>{$t('catalog.common.clear')}</Button
				>
			</div>
		</div>

		<EntityTable
			data={suppliers}
			columns={supplierColumns}
			{loading}
			onRowClick={(supplier) => onViewDetails(supplier, 'suppliers')}
		>
			{#snippet children(supplier)}
				<TableCell class="font-medium text-slate-800">{supplier.Name}</TableCell>
				<TableCell>
					<p class="text-sm font-medium text-slate-700">{supplier.ContactPerson ?? '—'}</p>
					<p class="text-xs text-slate-500">{supplier.Email ?? supplier.Phone ?? ''}</p>
				</TableCell>
				{#if auth.hasPermission('suppliers.write')}
					<TableCell class="text-right">
						<div class="flex items-center justify-end gap-1">
							<Button
								size="sm"
								variant="ghost"
								class="h-8 text-violet-600 hover:bg-violet-50 hover:text-violet-700"
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
								class="h-8 text-rose-600 hover:bg-rose-50 hover:text-rose-700"
								onclick={(event) => {
									event.stopPropagation();
									deleteSupplier(supplier);
								}}
							>
								{$t('catalog.common.delete')}
							</Button>
						</div>
					</TableCell>
				{/if}
			{/snippet}
		</EntityTable>
	</div>

	{#if auth.hasPermission('suppliers.write')}
		<SupplierForm editing={editingSupplier} onsaved={() => loadSuppliers()} />
	{/if}
</div>
