<script lang="ts">
	import { toast } from 'svelte-sonner';
	import { TableCell } from '$lib/components/ui/table';
	import { Input } from '$lib/components/ui/input';
	import { Button } from '$lib/components/ui/button';
	import EntityTable from './components/EntityTable.svelte';
	import CategoryForm from './components/CategoryForm.svelte';
	import { categoriesApi } from '$lib/api/resources';
	import type { Category } from '$lib/types';
	import type { CatalogTab } from '$lib/catalog/detailSections';
	import { auth } from '$lib/stores/auth';
	import { t } from '$lib/i18n';

	type Props = {
		refreshToken?: number;
		onViewDetails: (resource: Record<string, any>, type: CatalogTab) => void;
	};

	let { refreshToken = 0, onViewDetails }: Props = $props();

	let loading = $state(false);
	let categories = $state<Category[]>([]);
	let editingCategory = $state<Category | null>(null);
	let categorySearchTerm = $state('');

	const categoryColumns = $derived([
		{ header: $t('catalog.categories.form.name'), accessorKey: 'Name' },
		...(auth.hasPermission('categories.write')
			? [{ header: $t('catalog.common.actions'), accessorKey: 'id', class: 'text-right' }]
			: [])
	]);

	const loadCategories = async () => {
		loading = true;
		try {
			const categoryList = await categoriesApi.list();
			categories = Array.isArray(categoryList) ? categoryList : [categoryList];
		} catch (error: any) {
			const errorMessage = error.response?.data?.error || 'Unable to load categories';
			toast.error('Failed to Load Catalog', { description: errorMessage });
		} finally {
			loading = false;
		}
	};

	const handleCategorySearch = async () => {
		if (!categorySearchTerm.trim()) return;
		loading = true;
		try {
			const category = await categoriesApi.getByName(categorySearchTerm.trim());
			if (category) {
				onViewDetails(category, 'categories');
			} else {
				toast.error('Category not found');
			}
		} catch (error: any) {
			const errorMessage = error.response?.data?.error || 'Search failed';
			toast.error('Search Failed', { description: errorMessage });
		} finally {
			loading = false;
		}
	};

	const clearCategorySearch = () => {
		categorySearchTerm = '';
		loadCategories();
	};

	const deleteCategory = async (category: Category) => {
		if (!window.confirm($t('catalog.toasts.confirm_delete', { name: category.Name }))) return;
		try {
			await categoriesApi.remove(category.ID);
			toast.success($t('catalog.toasts.category_removed'));
			await loadCategories();
		} catch (error: any) {
			const errorMessage = error.response?.data?.error || 'Unable to delete category';
			toast.error($t('catalog.toasts.category_remove_fail'), { description: errorMessage });
		}
	};

	$effect(() => {
		if (refreshToken >= 0) loadCategories();
	});
</script>

<div
	class="grid gap-8 {auth.hasPermission('categories.write')
		? 'lg:grid-cols-[2fr,1fr]'
		: 'grid-cols-1'}"
>
	<div class="flex flex-col gap-6">
		<div
			class="flex flex-col gap-4 rounded-2xl border border-emerald-100 bg-white/50 p-4 shadow-sm backdrop-blur sm:flex-row sm:items-center sm:justify-between"
		>
			<div class="space-y-1">
				<h2 class="text-lg font-semibold text-slate-800">{$t('catalog.categories.title')}</h2>
				<p class="text-sm text-slate-500">{$t('catalog.categories.subtitle')}</p>
			</div>
			<div class="flex flex-wrap items-center gap-2">
				<Input
					class="w-full rounded-xl border-emerald-200 bg-white/80 px-3 text-sm focus:ring-2 focus:ring-emerald-400 sm:w-48"
					placeholder={$t('catalog.categories.search_placeholder')}
					bind:value={categorySearchTerm}
				/>
				<Button
					class="rounded-xl bg-emerald-500 px-4 py-2 text-white hover:bg-emerald-600"
					onclick={handleCategorySearch}>{$t('catalog.common.search')}</Button
				>
				<Button
					variant="ghost"
					class="rounded-xl px-4 py-2 text-emerald-600 hover:bg-emerald-50"
					onclick={clearCategorySearch}>{$t('catalog.common.clear')}</Button
				>
			</div>
		</div>

		<EntityTable
			data={categories}
			columns={categoryColumns}
			{loading}
			onRowClick={(category) => onViewDetails(category, 'categories')}
		>
			{#snippet children(category)}
				<TableCell class="font-medium text-slate-800">{category.Name}</TableCell>
				{#if auth.hasPermission('categories.write')}
					<TableCell class="text-right">
						<div class="flex items-center justify-end gap-1">
							<Button
								size="sm"
								variant="ghost"
								class="h-8 text-emerald-600 hover:bg-emerald-50 hover:text-emerald-700"
								onclick={(event) => {
									event.stopPropagation();
									editingCategory = category;
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
									deleteCategory(category);
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

	{#if auth.hasPermission('categories.write')}
		<CategoryForm editing={editingCategory} onsaved={() => loadCategories()} />
	{/if}
</div>
