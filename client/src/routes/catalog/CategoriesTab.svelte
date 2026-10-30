<script lang="ts">
	import { toast } from 'svelte-sonner';
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

	const inputClass =
		'glass-input h-10 w-full px-3 text-sm text-foreground placeholder:text-muted-foreground';
	const btnPrimary = 'btn-soothe h-10 px-4 text-sm font-semibold';

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
	class="grid gap-6 {auth.hasPermission('categories.write')
		? 'lg:grid-cols-[minmax(0,2fr)_minmax(0,1fr)]'
		: 'grid-cols-1'}"
>
	<div class="flex min-w-0 flex-col gap-4">
		<div class="panel p-4">
			<div class="flex flex-col gap-4 lg:flex-row lg:items-center lg:justify-between">
				<div class="min-w-0 space-y-1">
					<h2 class="section-title">{$t('catalog.categories.title')}</h2>
					<p class="text-sm text-muted-foreground">{$t('catalog.categories.subtitle')}</p>
				</div>
				<div class="flex flex-col gap-3 sm:flex-row sm:items-center">
					<Input
						class="{inputClass} sm:w-52"
						placeholder={$t('catalog.categories.search_placeholder')}
						bind:value={categorySearchTerm}
					/>
					<div class="flex items-center gap-2">
						<Button class={btnPrimary} onclick={handleCategorySearch}
							>{$t('catalog.common.search')}</Button
						>
						<Button
							variant="ghost"
							class="h-10 px-3 text-sm text-muted-foreground hover:bg-muted hover:text-foreground"
							onclick={clearCategorySearch}>{$t('catalog.common.clear')}</Button
						>
					</div>
				</div>
			</div>
		</div>

		<EntityTable
			data={categories}
			columns={categoryColumns}
			{loading}
			onRowClick={(category) => onViewDetails(category, 'categories')}
		>
			{#snippet children(category)}
				<td class="font-medium text-foreground">{category.Name}</td>
				{#if auth.hasPermission('categories.write')}
					<td class="text-right">
						<div class="flex items-center justify-end gap-1">
							<Button
								size="sm"
								variant="ghost"
								class="h-8 px-2.5 text-muted-foreground hover:bg-muted hover:text-foreground"
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
								class="h-8 px-2.5 text-danger hover:bg-danger/10 hover:text-danger"
								onclick={(event) => {
									event.stopPropagation();
									deleteCategory(category);
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

	{#if auth.hasPermission('categories.write')}
		<CategoryForm editing={editingCategory} onsaved={() => loadCategories()} />
	{/if}
</div>
