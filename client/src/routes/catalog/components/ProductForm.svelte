<script lang="ts">
	import { untrack } from 'svelte';
	import { superForm } from 'sveltekit-superforms';
	import { zod4 } from 'sveltekit-superforms/adapters';
	import { productSchema } from '$lib/schemas/product';
	import { productsApi, subCategoriesApi } from '$lib/api/resources';
	import type { Category, Location, Product, SubCategory, Supplier } from '$lib/types';
	import type { CatalogTab } from '$lib/catalog/detailSections';
	import { toast } from 'svelte-sonner';
	import { t } from '$lib/i18n';
	import {
		Card,
		CardContent,
		CardDescription,
		CardHeader,
		CardTitle
	} from '$lib/components/ui/card';
	import { Input } from '$lib/components/ui/input';
	import { Button } from '$lib/components/ui/button';

	type Props = {
		editing: Product | null;
		categories: Category[];
		suppliers: Supplier[];
		locations: Location[];
		products: Product[];
		onsaved: () => void;
		oncancel?: () => void;
		onNavigate?: (tab: CatalogTab) => void;
	};

	let {
		editing = null,
		categories = [],
		suppliers = [],
		locations = [],
		products = [],
		onsaved,
		oncancel,
		onNavigate
	}: Props = $props();

	let subCategories = $state<SubCategory[]>([]);

	const loadSubCategories = async (categoryId: number) => {
		if (!categoryId) {
			subCategories = [];
			return;
		}
		try {
			const list = await subCategoriesApi.list(categoryId);
			subCategories = Array.isArray(list) ? list : [list];
		} catch (error) {
			const errorMessage = (error as any)?.response?.data?.error || 'Unable to load sub-categories';
			toast.error('Failed to Load Sub-Categories', { description: errorMessage });
		}
	};

	const { form, errors, enhance, reset } = superForm(
		{
			sku: '',
			name: '',
			description: '',
			categoryId: '',
			subCategoryId: '',
			supplierId: '',
			locationId: '',
			purchasePrice: undefined as number | undefined,
			sellingPrice: undefined as number | undefined,
			status: 'Active',
			barCodeUPC: ''
		},
		{
			validators: zod4(productSchema),
			onSubmit: async ({ cancel }) => {
				cancel();
				const data = $form;

				const duplicate = products.find(
					(p) =>
						p.BarcodeUPC?.toLowerCase() === data.barCodeUPC.trim().toLowerCase() &&
						p.ID !== editing?.ID
				);
				if (duplicate) {
					toast.error('Duplicate Barcode/UPC Detected', {
						description: `The BarcodeUPC "${data.barCodeUPC}" is already used by product "${duplicate.Name}".`
					});
					return;
				}

				const payload = {
					sku: data.sku,
					name: data.name,
					description: data.description,
					categoryId: data.categoryId ? Number(data.categoryId) : undefined,
					subCategoryId: data.subCategoryId ? Number(data.subCategoryId) : undefined,
					supplierId: data.supplierId ? Number(data.supplierId) : undefined,
					locationId: data.locationId ? Number(data.locationId) : undefined,
					purchasePrice: data.purchasePrice || undefined,
					sellingPrice: data.sellingPrice || undefined,
					status: data.status,
					BarcodeUPC: data.barCodeUPC.trim()
				};

				try {
					if (editing) {
						await productsApi.update(editing.ID, payload);
					} else {
						await productsApi.create(payload);
					}
					toast.success($t('catalog.toasts.product_saved'));
					reset();
					subCategories = [];
					onsaved();
				} catch (error: any) {
					const backendMessage =
						error?.response?.data?.error ||
						error?.response?.data?.message ||
						'Unable to save product';
					toast.error($t('catalog.toasts.product_save_fail'), { description: backendMessage });
				}
			}
		}
	);

	$effect(() => {
		const current = editing;
		untrack(() => {
			if (current) {
				$form.sku = current.SKU;
				$form.name = current.Name;
				$form.description = current.Description ?? '';
				$form.categoryId = current.CategoryID ? String(current.CategoryID) : '';
				$form.subCategoryId = current.SubCategoryID ? String(current.SubCategoryID) : '';
				$form.supplierId = current.SupplierID ? String(current.SupplierID) : '';
				$form.locationId = current.LocationID ? String(current.LocationID) : '';
				$form.purchasePrice = current.PurchasePrice;
				$form.sellingPrice = current.SellingPrice;
				$form.status = current.Status ?? 'Active';
				$form.barCodeUPC = current.BarcodeUPC ?? '';
				if (current.CategoryID) loadSubCategories(current.CategoryID);
			} else {
				reset();
				subCategories = [];
			}
		});
	});
</script>

<Card
	class="overflow-hidden rounded-2xl border-0 bg-gradient-to-br from-sky-50 to-blue-100 shadow-lg transition-all duration-300 hover:scale-[1.01] hover:shadow-xl"
	data-animate="fade-up"
	style="animation-delay:240ms"
>
	<CardHeader class="space-y-1 border-b border-white/60 bg-white/70 px-6 py-5 backdrop-blur">
		<CardTitle class="text-slate-800"
			>{editing
				? $t('catalog.products.form.update_title')
				: $t('catalog.products.form.create_title')}</CardTitle
		>
		<CardDescription class="text-slate-600">{$t('catalog.products.form.subtitle')}</CardDescription>
	</CardHeader>
	<CardContent class="p-6">
		<form method="POST" use:enhance class="space-y-4">
			<div class="space-y-1">
				<Input
					class="w-full rounded-xl border border-sky-200 bg-white/90 px-3.5 py-2.5 text-sm focus:ring-2 focus:ring-sky-400"
					placeholder={$t('catalog.products.form.sku')}
					bind:value={$form.sku}
				/>
				{#if $errors.sku}
					<p class="text-xs font-medium text-rose-600">{$errors.sku}</p>
				{/if}
			</div>
			<div class="space-y-1">
				<Input
					class="w-full rounded-xl border border-sky-200 bg-white/90 px-3.5 py-2.5 text-sm focus:ring-2 focus:ring-sky-400"
					placeholder={$t('catalog.products.form.name')}
					bind:value={$form.name}
				/>
				{#if $errors.name}
					<p class="text-xs font-medium text-rose-600">{$errors.name}</p>
				{/if}
			</div>
			<Input
				class="w-full rounded-xl border border-sky-200 bg-white/90 px-3.5 py-2.5 text-sm focus:ring-2 focus:ring-sky-400"
				placeholder={$t('catalog.products.form.description')}
				bind:value={$form.description}
			/>
			<div class="space-y-1">
				<Input
					class="w-full rounded-xl border border-sky-200 bg-white/90 px-3.5 py-2.5 text-sm focus:ring-2 focus:ring-sky-400"
					placeholder={$t('catalog.products.form.barcode')}
					bind:value={$form.barCodeUPC}
				/>
				{#if $errors.barCodeUPC}
					<p class="text-xs font-medium text-rose-600">{$errors.barCodeUPC}</p>
				{/if}
			</div>

			<div class="select-wrapper">
				<select
					class="w-full rounded-xl border border-sky-200 bg-white/90 px-3.5 py-2.5 text-sm focus:ring-2 focus:ring-sky-400"
					bind:value={$form.categoryId}
					onchange={(e) => {
						$form.subCategoryId = '';
						loadSubCategories(Number(e.currentTarget.value));
					}}
				>
					<option value="">{$t('catalog.products.form.select_category')}</option>
					{#each categories as category (category.ID)}
						<option value={String(category.ID)}>{category.Name}</option>
					{/each}
				</select>
			</div>

			<div class="flex items-center gap-3">
				<div class="select-wrapper">
					<select
						class="w-full rounded-xl border border-sky-200 bg-white/90 px-3.5 py-2.5 text-sm focus:ring-2 focus:ring-sky-400"
						bind:value={$form.subCategoryId}
					>
						<option value="">{$t('catalog.products.form.select_sub_category')}</option>
						{#each subCategories.filter((sc) => sc.CategoryID === Number($form.categoryId)) as subCategory (subCategory.ID)}
							<option value={String(subCategory.ID)}>{subCategory.Name}</option>
						{/each}
					</select>
				</div>

				<Button
					type="button"
					size="sm"
					variant="outline"
					class="rounded-xl border border-sky-200 px-3 py-2 text-sky-700 hover:bg-sky-50"
					onclick={() => onNavigate?.('sub-categories')}
				>
					{$t('catalog.common.new')}</Button
				>
			</div>

			<div class="select-wrapper">
				<select
					class="w-full rounded-xl border border-sky-200 bg-white/90 px-3.5 py-2.5 text-sm focus:ring-2 focus:ring-sky-400"
					bind:value={$form.supplierId}
				>
					<option value="">{$t('catalog.products.form.select_supplier')}</option>
					{#each suppliers as supplier (supplier.ID)}
						<option value={String(supplier.ID)}>{supplier.Name}</option>
					{/each}
				</select>
			</div>

			<div class="select-wrapper">
				<select
					class="w-full rounded-xl border border-sky-200 bg-white/90 px-3.5 py-2.5 text-sm focus:ring-2 focus:ring-sky-400"
					bind:value={$form.locationId}
				>
					<option value="">{$t('catalog.products.form.default_location')}</option>
					{#each locations as location (location.ID)}
						<option value={String(location.ID)}>{location.Name}</option>
					{/each}
				</select>
			</div>

			<div class="grid grid-cols-1 gap-3 sm:grid-cols-2">
				<div class="space-y-1">
					<Input
						type="number"
						min="0"
						step="0.01"
						class="w-full rounded-xl border border-sky-200 bg-white/90 px-3.5 py-2.5 text-sm focus:ring-2 focus:ring-sky-400"
						placeholder={$t('catalog.products.form.purchase_price')}
						bind:value={$form.purchasePrice}
					/>
					{#if $errors.purchasePrice}
						<p class="text-xs font-medium text-rose-600">{$errors.purchasePrice}</p>
					{/if}
				</div>
				<div class="space-y-1">
					<Input
						type="number"
						min="0"
						step="0.01"
						class="w-full rounded-xl border border-sky-200 bg-white/90 px-3.5 py-2.5 text-sm focus:ring-2 focus:ring-sky-400"
						placeholder={$t('catalog.products.form.selling_price')}
						bind:value={$form.sellingPrice}
					/>
					{#if $errors.sellingPrice}
						<p class="text-xs font-medium text-rose-600">{$errors.sellingPrice}</p>
					{/if}
				</div>
			</div>

			<div class="select-wrapper">
				<select
					class="w-full rounded-xl border border-sky-200 bg-white/90 px-3.5 py-2.5 text-sm focus:ring-2 focus:ring-sky-400"
					bind:value={$form.status}
				>
					<option value="Active">Active</option>
					<option value="Archived">Archived</option>
					<option value="Discontinued">Discontinued</option>
				</select>
			</div>
			<div class="flex flex-col gap-3 pt-1 pr-2 sm:flex-row">
				<Button
					type="submit"
					class="w-full rounded-xl bg-gradient-to-r from-sky-500 to-blue-600 py-2.5 text-white shadow-lg transition-all duration-300 hover:scale-105 hover:from-sky-600 hover:to-blue-700 hover:shadow-xl sm:w-1/2"
				>
					{editing ? $t('catalog.common.update') : $t('catalog.common.create')}
				</Button>
				<Button
					type="button"
					class="w-full rounded-xl border border-sky-200 py-2.5 text-sky-700 transition hover:bg-sky-50 sm:w-1/2"
					onclick={() => {
						reset();
						oncancel?.();
					}}
				>
					{$t('catalog.common.reset')}
				</Button>
			</div>
		</form>
	</CardContent>
</Card>
