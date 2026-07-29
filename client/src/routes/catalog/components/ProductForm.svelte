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

	const inputClass =
		'liquid-input h-10 w-full rounded-md bg-card bg-none px-3 text-sm font-normal text-foreground shadow-none ring-0 backdrop-blur-none backdrop-saturate-100 from-card to-card placeholder:text-muted-foreground focus:scale-100 focus:border-primary focus:ring-0 md:text-sm';
	const btnPrimary =
		'rounded-md bg-primary bg-none px-4 text-sm font-medium text-primary-foreground shadow-none hover:bg-primary/90 hover:bg-none';
	const btnSecondary =
		'rounded-md border border-border bg-card bg-none text-foreground shadow-none backdrop-blur-none hover:bg-muted hover:bg-none';

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

<div
	class="liquid-panel overflow-hidden rounded-lg"
	data-animate="fade-up"
	style="animation-delay:240ms"
>
	<div class="border-border space-y-1 border-b px-6 py-5">
		<h3 class="text-foreground font-display text-lg font-semibold"
			>{editing
				? $t('catalog.products.form.update_title')
				: $t('catalog.products.form.create_title')}</h3
		>
		<p class="text-muted-foreground text-sm">{$t('catalog.products.form.subtitle')}</p>
	</div>
	<div class="p-6">
		<form method="POST" use:enhance class="space-y-4">
			<div class="space-y-1">
				<Input
					class="{inputClass} num-mono"
					placeholder={$t('catalog.products.form.sku')}
					bind:value={$form.sku}
				/>
				{#if $errors.sku}
					<p class="text-danger text-xs font-medium">{$errors.sku}</p>
				{/if}
			</div>
			<div class="space-y-1">
				<Input
					class={inputClass}
					placeholder={$t('catalog.products.form.name')}
					bind:value={$form.name}
				/>
				{#if $errors.name}
					<p class="text-danger text-xs font-medium">{$errors.name}</p>
				{/if}
			</div>
			<Input
				class={inputClass}
				placeholder={$t('catalog.products.form.description')}
				bind:value={$form.description}
			/>
			<div class="space-y-1">
				<Input
					class="{inputClass} num-mono"
					placeholder={$t('catalog.products.form.barcode')}
					bind:value={$form.barCodeUPC}
				/>
				{#if $errors.barCodeUPC}
					<p class="text-danger text-xs font-medium">{$errors.barCodeUPC}</p>
				{/if}
			</div>

			<div class="select-wrapper">
				<select
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
					<select bind:value={$form.subCategoryId}>
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
					class="{btnSecondary} shrink-0"
					onclick={() => onNavigate?.('sub-categories')}
				>
					{$t('catalog.common.new')}</Button
				>
			</div>

			<div class="select-wrapper">
				<select bind:value={$form.supplierId}>
					<option value="">{$t('catalog.products.form.select_supplier')}</option>
					{#each suppliers as supplier (supplier.ID)}
						<option value={String(supplier.ID)}>{supplier.Name}</option>
					{/each}
				</select>
			</div>

			<div class="select-wrapper">
				<select bind:value={$form.locationId}>
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
						class="{inputClass} num-mono"
						placeholder={$t('catalog.products.form.purchase_price')}
						bind:value={$form.purchasePrice}
					/>
					{#if $errors.purchasePrice}
						<p class="text-danger text-xs font-medium">{$errors.purchasePrice}</p>
					{/if}
				</div>
				<div class="space-y-1">
					<Input
						type="number"
						min="0"
						step="0.01"
						class="{inputClass} num-mono"
						placeholder={$t('catalog.products.form.selling_price')}
						bind:value={$form.sellingPrice}
					/>
					{#if $errors.sellingPrice}
						<p class="text-danger text-xs font-medium">{$errors.sellingPrice}</p>
					{/if}
				</div>
			</div>

			<div class="select-wrapper">
				<select bind:value={$form.status}>
					<option value="Active">Active</option>
					<option value="Archived">Archived</option>
					<option value="Discontinued">Discontinued</option>
				</select>
			</div>
			<div class="flex flex-col gap-3 pt-1 pr-2 sm:flex-row">
				<Button type="submit" class="{btnPrimary} w-full py-2.5 sm:w-1/2">
					{editing ? $t('catalog.common.update') : $t('catalog.common.create')}
				</Button>
				<Button
					type="button"
					class="{btnSecondary} w-full py-2.5 sm:w-1/2"
					onclick={() => {
						reset();
						oncancel?.();
					}}
				>
					{$t('catalog.common.reset')}
				</Button>
			</div>
		</form>
	</div>
</div>
