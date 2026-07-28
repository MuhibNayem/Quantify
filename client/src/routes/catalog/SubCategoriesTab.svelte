<script lang="ts">
	import { toast } from 'svelte-sonner';
	import { superForm } from 'sveltekit-superforms';
	import { zod4 } from 'sveltekit-superforms/adapters';
	import { subCategorySchema } from '$lib/schemas/category';
	import { TableCell } from '$lib/components/ui/table';
	import { Input } from '$lib/components/ui/input';
	import { Button } from '$lib/components/ui/button';
	import { Card, CardContent, CardHeader, CardTitle } from '$lib/components/ui/card';
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
	class="grid gap-8 {auth.hasPermission('categories.write')
		? 'lg:grid-cols-[2fr,1fr]'
		: 'grid-cols-1'}"
>
	<div class="flex flex-col gap-6">
		<div
			class="flex flex-col gap-4 rounded-2xl border border-amber-100 bg-white/50 p-4 shadow-sm backdrop-blur sm:flex-row sm:items-center sm:justify-between"
		>
			<div class="space-y-1">
				<h2 class="text-lg font-semibold text-slate-800">{$t('catalog.sub_categories.title')}</h2>
				<p class="text-sm text-slate-500">{$t('catalog.sub_categories.subtitle')}</p>
			</div>
			<div class="select-wrapper w-full sm:w-1/2">
				<select
					class="w-full rounded-xl border border-amber-200 bg-white/80 px-3 py-2 text-sm outline-none focus:ring-2 focus:ring-amber-400"
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

		{#if !$form.categoryId}
			<div
				class="rounded-2xl border-2 border-dashed border-amber-200/50 bg-white/40 py-12 text-center text-slate-500"
			>
				<Layers class="mx-auto mb-3 h-12 w-12 text-amber-200" />
				<p class="text-base font-medium">{$t('catalog.sub_categories.empty_state')}</p>
				<p class="text-sm text-slate-400">{$t('catalog.sub_categories.empty_subtitle')}</p>
			</div>
		{:else}
			<EntityTable
				data={subCategories.filter((sc) => sc.CategoryID === Number($form.categoryId))}
				columns={subCategoryColumns}
				{loading}
				onRowClick={(subCategory) => onViewDetails(subCategory, 'sub-categories')}
			>
				{#snippet children(subCategory)}
					<TableCell class="font-medium text-slate-800">{subCategory.Name}</TableCell>
					{#if auth.hasPermission('categories.write')}
						<TableCell class="text-right">
							<div class="flex items-center justify-end gap-1">
								<Button
									size="sm"
									variant="ghost"
									class="h-8 text-amber-700 hover:bg-amber-100 hover:text-amber-800"
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
									class="h-8 text-rose-600 hover:bg-rose-50 hover:text-rose-700"
									onclick={(event) => {
										event.stopPropagation();
										deleteSubCategory(subCategory);
									}}
								>
									Delete
								</Button>
							</div>
						</TableCell>
					{/if}
				{/snippet}
			</EntityTable>
		{/if}
	</div>

	{#if auth.hasPermission('categories.write')}
		<Card
			class="overflow-hidden rounded-2xl border-0 bg-gradient-to-br from-amber-50 to-orange-100 shadow-lg transition-all duration-300 hover:scale-[1.01] hover:shadow-xl"
			data-animate="fade-up"
			style="animation-delay:180ms"
		>
			<CardHeader class="space-y-1 border-b border-white/60 bg-white/70 px-6 py-5 backdrop-blur">
				<CardTitle class="text-slate-800"
					>{editingSubCategory
						? $t('catalog.sub_categories.form.update_title')
						: $t('catalog.sub_categories.form.create_title')}</CardTitle
				>
			</CardHeader>
			<CardContent class="p-6">
				<form method="POST" use:enhance class="space-y-4">
					<div class="space-y-1">
						<Input
							class="w-full rounded-xl border border-amber-200 bg-white/90 px-3.5 py-2.5 text-sm focus:ring-2 focus:ring-amber-400"
							placeholder={$t('catalog.sub_categories.form.name')}
							bind:value={$form.name}
						/>
						{#if $errors.name}
							<p class="text-xs font-medium text-rose-600">{$errors.name}</p>
						{/if}
					</div>
					<div class="space-y-1">
						<select
							class="w-full rounded-xl border border-amber-200 bg-white/90 px-3.5 py-2.5 text-sm focus:ring-2 focus:ring-amber-400"
							bind:value={$form.categoryId}
						>
							<option value="">{$t('catalog.sub_categories.form.select_category')}</option>
							{#each categories as category (category.ID)}
								<option value={category.ID}>{category.Name}</option>
							{/each}
						</select>
						{#if $errors.categoryId}
							<p class="text-xs font-medium text-rose-600">{$errors.categoryId}</p>
						{/if}
					</div>
					<div class="flex flex-col gap-3 pt-1 sm:flex-row">
						<Button
							type="submit"
							class="w-full rounded-xl bg-gradient-to-r from-amber-500 to-orange-600 py-2.5 text-white shadow-lg transition-all duration-300 hover:scale-105 hover:from-amber-600 hover:to-orange-700 hover:shadow-xl sm:w-1/2"
						>
							{editingSubCategory ? $t('catalog.common.update') : $t('catalog.common.create')}
						</Button>
						<Button
							type="button"
							class="w-full rounded-xl border border-amber-200 py-2.5 text-amber-700 transition hover:bg-amber-50 sm:w-1/2"
							onclick={() => {
								$form.name = '';
								editingSubCategory = null;
							}}
						>
							{$t('catalog.common.reset')}
						</Button>
					</div>
				</form>
			</CardContent>
		</Card>
	{/if}
</div>
