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
		'liquid-input h-10 w-full rounded-md bg-card bg-none px-3 text-sm font-normal text-foreground shadow-none ring-0 backdrop-blur-none backdrop-saturate-100 from-card to-card placeholder:text-muted-foreground focus:scale-100 focus:border-primary focus:ring-0 md:text-sm';
	const btnPrimary =
		'rounded-md bg-primary bg-none px-4 text-sm font-medium text-primary-foreground shadow-none hover:bg-primary/90 hover:bg-none';
	const btnSecondary =
		'rounded-md border border-border bg-card bg-none text-foreground shadow-none backdrop-blur-none hover:bg-muted hover:bg-none';

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
			class="liquid-panel flex flex-col gap-4 rounded-lg p-4 sm:flex-row sm:items-center sm:justify-between"
		>
			<div class="space-y-1">
				<h2 class="font-display text-foreground text-lg font-semibold">
					{$t('catalog.sub_categories.title')}
				</h2>
				<p class="text-muted-foreground text-sm">{$t('catalog.sub_categories.subtitle')}</p>
			</div>
			<div class="select-wrapper w-full sm:w-1/2">
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

		{#if !$form.categoryId}
			<div
				class="border-border bg-card text-muted-foreground rounded-lg border-2 border-dashed py-12 text-center shadow-sm"
			>
				<Layers class="text-muted-foreground/40 mx-auto mb-3 h-12 w-12" />
				<p class="text-foreground text-base font-medium">{$t('catalog.sub_categories.empty_state')}</p>
				<p class="text-muted-foreground text-sm">{$t('catalog.sub_categories.empty_subtitle')}</p>
			</div>
		{:else}
			<EntityTable
				data={subCategories.filter((sc) => sc.CategoryID === Number($form.categoryId))}
				columns={subCategoryColumns}
				{loading}
				onRowClick={(subCategory) => onViewDetails(subCategory, 'sub-categories')}
			>
				{#snippet children(subCategory)}
					<td class="text-foreground font-medium">{subCategory.Name}</td>
					{#if auth.hasPermission('categories.write')}
						<td class="text-right">
							<div class="flex items-center justify-end gap-1">
								<Button
									size="sm"
									variant="ghost"
									class="h-8 text-primary hover:bg-muted hover:text-foreground"
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
									class="h-8 text-danger hover:bg-danger/10 hover:text-danger"
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
		<div
			class="liquid-panel overflow-hidden rounded-lg"
			data-animate="fade-up"
			style="animation-delay:180ms"
		>
			<div class="border-border space-y-1 border-b px-6 py-5">
				<h3 class="text-foreground font-display text-lg font-semibold"
					>{editingSubCategory
						? $t('catalog.sub_categories.form.update_title')
						: $t('catalog.sub_categories.form.create_title')}</h3
				>
			</div>
			<div class="p-6">
				<form method="POST" use:enhance class="space-y-4">
					<div class="space-y-1">
						<Input
							class={inputClass}
							placeholder={$t('catalog.sub_categories.form.name')}
							bind:value={$form.name}
						/>
						{#if $errors.name}
							<p class="text-danger text-xs font-medium">{$errors.name}</p>
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
							<p class="text-danger text-xs font-medium">{$errors.categoryId}</p>
						{/if}
					</div>
					<div class="flex flex-col gap-3 pt-1 sm:flex-row">
						<Button type="submit" class="{btnPrimary} w-full py-2.5 sm:w-1/2">
							{editingSubCategory ? $t('catalog.common.update') : $t('catalog.common.create')}
						</Button>
						<Button
							type="button"
							class="{btnSecondary} w-full py-2.5 sm:w-1/2"
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
