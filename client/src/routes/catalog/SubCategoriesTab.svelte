<script lang="ts">
	import { toast } from 'svelte-sonner';
	import { superForm } from 'sveltekit-superforms';
	import { zod4 } from 'sveltekit-superforms/adapters';
	import { subCategorySchema } from '$lib/schemas/category';
	import { Input } from '$lib/components/ui/input';
	import { Button } from '$lib/components/ui/button';
	import EntityTable from './components/EntityTable.svelte';
	import { categoriesApi, subCategoriesApi } from '$lib/api/resources';
	import type { Category, SubCategory } from '$lib/types';
	import type { CatalogTab } from '$lib/catalog/detailSections';
	import { Layers } from 'lucide-svelte';
	import { auth } from '$lib/stores/auth';
	import { t } from '$lib/i18n';

	type Props = {
		refreshToken?: number;
		onViewDetails: (resource: Record<string, any>, type: CatalogTab) => void;
	};

	let { refreshToken = 0, onViewDetails }: Props = $props();

	let loading = $state(false);
	let categories = $state<Category[]>([]);
	let subCategories = $state<SubCategory[]>([]);
	let editingSubCategory = $state<SubCategory | null>(null);

	const inputClass =
		'glass-input h-10 w-full px-3 text-sm text-foreground placeholder:text-muted-foreground';
	const btnPrimary = 'btn-soothe h-10 px-4 text-sm font-semibold';
	const btnSecondary = 'btn-soft h-10 px-4 text-sm font-medium';

	const { form, errors, enhance } = superForm(
		{ name: '', categoryId: '' },
		{
			validators: zod4(subCategorySchema),
			onSubmit: async ({ cancel }) => {
				cancel();
				const data = $form;
				try {
					const payload = { name: data.name };
					if (editingSubCategory) {
						await subCategoriesApi.update(editingSubCategory.ID, payload);
					} else {
						await subCategoriesApi.create(payload, Number(data.categoryId));
					}
					toast.success($t('catalog.toasts.sub_category_saved'));
					$form.name = '';
					editingSubCategory = null;
					await loadSubCategories(Number(data.categoryId));
				} catch (error) {
					const errorMessage =
						(error as any)?.response?.data?.error || 'Unable to save sub-category';
					toast.error($t('catalog.toasts.sub_category_save_fail'), { description: errorMessage });
				}
			}
		}
	);

	const subCategoryColumns = $derived([
		{ header: $t('catalog.sub_categories.form.name'), accessorKey: 'Name' },
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

	const loadSubCategories = async (categoryId: number) => {
		if (!categoryId) {
			subCategories = [];
			return;
		}
		loading = true;
		try {
			const list = await subCategoriesApi.list(categoryId);
			subCategories = Array.isArray(list) ? list : [list];
		} catch (error) {
			const errorMessage = (error as any)?.response?.data?.error || 'Unable to load sub-categories';
			toast.error('Failed to Load Sub-Categories', { description: errorMessage });
		} finally {
			loading = false;
		}
	};

	const deleteSubCategory = async (subCategory: SubCategory) => {
		if (!window.confirm($t('catalog.toasts.confirm_delete', { name: subCategory.Name }))) return;
		try {
			await subCategoriesApi.remove(subCategory.ID);
			toast.success($t('catalog.toasts.sub_category_removed'));
			await loadSubCategories(Number($form.categoryId));
		} catch (error: any) {
			const errorMessage = error.response?.data?.error || 'Unable to delete sub-category';
			toast.error($t('catalog.toasts.sub_category_remove_fail'), { description: errorMessage });
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
			<div class="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
				<div class="min-w-0 space-y-1">
					<h2 class="section-title">{$t('catalog.sub_categories.title')}</h2>
					<p class="text-sm text-muted-foreground">{$t('catalog.sub_categories.subtitle')}</p>
				</div>
				<div class="select-wrapper w-full sm:w-64">
					<select
						bind:value={$form.categoryId}
						onchange={(e) => loadSubCategories(Number(e.currentTarget.value))}
					>
						<option value="">{$t('catalog.sub_categories.select_category')}</option>
						{#if categories.length > 0}
							{#each categories as category (category.ID)}
								<option value={category.ID}>{category.Name}</option>
							{/each}
						{:else}
							<option disabled>{$t('catalog.sub_categories.loading_categories')}</option>
						{/if}
					</select>
				</div>
			</div>
		</div>

		{#if !$form.categoryId}
			<div class="panel flex flex-col items-center justify-center gap-2 px-6 py-14 text-center">
				<div
					class="mb-1 flex h-14 w-14 items-center justify-center rounded-2xl bg-muted text-muted-foreground"
				>
					<Layers class="h-7 w-7" strokeWidth={1.5} />
				</div>
				<p class="text-base font-medium text-foreground">
					{$t('catalog.sub_categories.empty_state')}
				</p>
				<p class="max-w-sm text-sm text-muted-foreground">
					{$t('catalog.sub_categories.empty_subtitle')}
				</p>
			</div>
		{:else}
			<EntityTable
				data={subCategories.filter((sc) => sc.CategoryID === Number($form.categoryId))}
				columns={subCategoryColumns}
				{loading}
				onRowClick={(subCategory) => onViewDetails(subCategory, 'sub-categories')}
			>
				{#snippet children(subCategory)}
					<td class="font-medium text-foreground">{subCategory.Name}</td>
					{#if auth.hasPermission('categories.write')}
						<td class="text-right">
							<div class="flex items-center justify-end gap-1">
								<Button
									size="sm"
									variant="ghost"
									class="h-8 px-2.5 text-muted-foreground hover:bg-muted hover:text-foreground"
									onclick={(event) => {
										event.stopPropagation();
										editingSubCategory = subCategory;
										$form.name = subCategory.Name;
										$form.categoryId = String(subCategory.CategoryID);
									}}
								>
									Edit
								</Button>
								<Button
									size="sm"
									variant="ghost"
									class="h-8 px-2.5 text-danger hover:bg-danger/10 hover:text-danger"
									onclick={(event) => {
										event.stopPropagation();
										deleteSubCategory(subCategory);
									}}
								>
									Delete
								</Button>
							</div>
						</td>
					{/if}
				{/snippet}
			</EntityTable>
		{/if}
	</div>

	{#if auth.hasPermission('categories.write')}
		<div class="panel overflow-hidden" data-animate="fade-up" style="animation-delay:180ms">
			<div class="panel-header flex-col items-start gap-1">
				<h3 class="section-title">
					{editingSubCategory
						? $t('catalog.sub_categories.form.update_title')
						: $t('catalog.sub_categories.form.create_title')}
				</h3>
			</div>
			<div class="panel-body">
				<form method="POST" use:enhance class="space-y-4">
					<div class="space-y-1">
						<Input
							class={inputClass}
							placeholder={$t('catalog.sub_categories.form.name')}
							bind:value={$form.name}
						/>
						{#if $errors.name}
							<p class="text-xs font-medium text-danger">{$errors.name}</p>
						{/if}
					</div>
					<div class="space-y-1">
						<div class="select-wrapper">
							<select bind:value={$form.categoryId}>
								<option value="">{$t('catalog.sub_categories.form.select_category')}</option>
								{#each categories as category (category.ID)}
									<option value={category.ID}>{category.Name}</option>
								{/each}
							</select>
						</div>
						{#if $errors.categoryId}
							<p class="text-xs font-medium text-danger">{$errors.categoryId}</p>
						{/if}
					</div>
					<div class="flex flex-col gap-3 pt-1 sm:flex-row">
						<Button type="submit" class="{btnPrimary} w-full sm:flex-1">
							{editingSubCategory ? $t('catalog.common.update') : $t('catalog.common.create')}
						</Button>
						<Button
							type="button"
							variant="secondary"
							class="{btnSecondary} w-full sm:flex-1"
							onclick={() => {
								$form.name = '';
								editingSubCategory = null;
							}}
						>
							{$t('catalog.common.reset')}
						</Button>
					</div>
				</form>
			</div>
		</div>
	{/if}
</div>
