<script lang="ts">
	import { untrack } from 'svelte';
	import { superForm } from 'sveltekit-superforms';
	import { zod4 } from 'sveltekit-superforms/adapters';
	import { productSchema } from '$lib/schemas/product';
	import { productsApi, subCategoriesApi, unitsApi } from '$lib/api/resources';
	import type {
		Category,
		Location,
		Product,
		ProductUnit,
		SubCategory,
		Supplier,
		Unit
	} from '$lib/types';
	import type { CatalogTab } from '$lib/catalog/detailSections';
	import { toast } from 'svelte-sonner';
	import { t } from '$lib/i18n';
	import { Input } from '$lib/components/ui/input';
	import { Button } from '$lib/components/ui/button';
	import { Switch } from '$lib/components/ui/switch';

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

	type ProductWithUnits = Product & {
		BaseUnitID?: number;
		AllowFractional?: boolean;
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

	let units = $state<Unit[]>([]);
	let productUnits = $state<ProductUnit[]>([]);
	let baseUnitId = $state<string>('');
	let allowFractional = $state(false);

	let newUnitId = $state('');
	let newFactor = $state<number | undefined>(undefined);
	let newPrice = $state<number | undefined>(undefined);
	let newPurchasePrice = $state<number | undefined>(undefined);
	let newBarcode = $state('');
	let newIsActive = $state(true);
	let savingUnit = $state(false);

	const inputClass =
		'liquid-input h-10 w-full rounded-md bg-card bg-none px-3 text-sm font-normal text-foreground shadow-none ring-0 backdrop-blur-none backdrop-saturate-100 from-card to-card placeholder:text-muted-foreground focus:scale-100 focus:border-primary focus:ring-0 md:text-sm';
	const btnPrimary =
		'rounded-md bg-primary bg-none px-4 text-sm font-medium text-primary-foreground shadow-none hover:bg-primary/90 hover:bg-none';
	const btnSecondary =
		'rounded-md border border-border bg-card bg-none text-foreground shadow-none backdrop-blur-none hover:bg-muted hover:bg-none';

	const baseUnit = $derived(units.find((u) => String(u.ID) === baseUnitId));

	const dimensionGroups = $derived.by(() => {
		const map = new Map<number, { name: string; units: Unit[] }>();
		for (const unit of units) {
			const name = unit.Dimension?.Name ?? 'Other';
			const group = map.get(unit.DimensionID);
			if (group) {
				group.units.push(unit);
			} else {
				map.set(unit.DimensionID, { name, units: [unit] });
			}
		}
		return [...map.values()];
	});

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

	const loadUnits = async () => {
		try {
			const list = await unitsApi.list();
			units = Array.isArray(list) ? list : [];
		} catch (error) {
			const errorMessage = (error as any)?.response?.data?.error || 'Unable to load units';
			toast.error('Failed to Load Units', { description: errorMessage });
		}
	};

	const loadProductUnits = async (productId: number) => {
		try {
			const list = await unitsApi.productUnits(productId);
			productUnits = Array.isArray(list) ? list : [];
		} catch {
			productUnits = [];
		}
	};

	const resetNewUnit = () => {
		newUnitId = '';
		newFactor = undefined;
		newPrice = undefined;
		newPurchasePrice = undefined;
		newBarcode = '';
		newIsActive = true;
	};

	const onNewUnitSelect = (unitIdStr: string) => {
		newUnitId = unitIdStr;
		const selected = units.find((u) => String(u.ID) === unitIdStr);
		if (selected && baseUnit && selected.DimensionID === baseUnit.DimensionID && baseUnit.Factor) {
			newFactor = Math.round((selected.Factor / baseUnit.Factor) * 10000) / 10000;
		}
	};

	const saveSellingUnit = async () => {
		if (!editing || !newUnitId) return;
		savingUnit = true;
		try {
			await unitsApi.saveProductUnit(editing.ID, {
				unitId: Number(newUnitId),
				factor: newFactor ?? 0,
				price: newPrice ?? 0,
				purchasePrice: newPurchasePrice ?? 0,
				barcode: newBarcode.trim(),
				isActive: newIsActive
			});
			toast.success('Selling unit saved');
			resetNewUnit();
			await loadProductUnits(editing.ID);
		} catch (error: any) {
			const backendMessage =
				error?.response?.data?.error || error?.response?.data?.message || 'Unable to save selling unit';
			toast.error('Failed to Save Selling Unit', { description: backendMessage });
		} finally {
			savingUnit = false;
		}
	};

	const deleteSellingUnit = async (unitId: number) => {
		if (!editing) return;
		try {
			await unitsApi.deleteProductUnit(editing.ID, unitId);
			toast.success('Selling unit removed');
			await loadProductUnits(editing.ID);
		} catch (error: any) {
			const backendMessage =
				error?.response?.data?.error ||
				error?.response?.data?.message ||
				'Unable to remove selling unit';
			toast.error('Failed to Remove Selling Unit', { description: backendMessage });
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
					BarcodeUPC: data.barCodeUPC.trim(),
					baseUnitId: baseUnitId ? Number(baseUnitId) : undefined,
					allowFractional
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
		loadUnits();
	});

	$effect(() => {
		const list = units;
		const current = editing;
		untrack(() => {
			if (list.length && !current && !baseUnitId) {
				const fallback =
					list.find((u) => u.Name.toLowerCase() === 'each') ?? list.find((u) => u.IsBase);
				if (fallback) baseUnitId = String(fallback.ID);
			}
		});
	});

	$effect(() => {
		const current = editing;
		untrack(() => {
			if (current) {
				const withUnits = current as ProductWithUnits;
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
				baseUnitId = withUnits.BaseUnitID ? String(withUnits.BaseUnitID) : '';
				allowFractional = Boolean(withUnits.AllowFractional);
				resetNewUnit();
				if (current.CategoryID) loadSubCategories(current.CategoryID);
				loadProductUnits(current.ID);
			} else {
				reset();
				subCategories = [];
				productUnits = [];
				allowFractional = false;
				resetNewUnit();
			}
		});
	});
</script>

<div class="flex flex-col gap-6">
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

				<div class="border-border bg-muted/30 space-y-4 rounded-md border p-4">
					<p class="overline">Units of Measure</p>
					<div class="space-y-1">
						<label class="text-muted-foreground text-xs font-medium" for="base-unit"
							>Base Unit</label
						>
						<div class="select-wrapper">
							<select id="base-unit" bind:value={baseUnitId}>
								<option value="">Select base unit</option>
								{#each dimensionGroups as group (group.name)}
									<optgroup label={group.name}>
										{#each group.units as unit (unit.ID)}
											<option value={String(unit.ID)}>{unit.Name} ({unit.Symbol})</option>
										{/each}
									</optgroup>
								{/each}
							</select>
						</div>
						<p class="text-muted-foreground text-xs">The unit stock is tracked in.</p>
					</div>
					<div class="flex items-center justify-between gap-3">
						<div class="space-y-0.5">
							<label
								class="text-foreground cursor-pointer text-sm font-medium"
								for="allow-fractional"
							>
								Allow fractional quantities
							</label>
							<p class="text-muted-foreground text-xs">For measured goods like wire or fabric.</p>
						</div>
						<Switch id="allow-fractional" bind:checked={allowFractional} />
					</div>
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

	{#if editing}
		<div
			class="liquid-panel overflow-hidden rounded-lg"
			data-animate="fade-up"
			style="animation-delay:300ms"
		>
			<div class="border-border space-y-1 border-b px-6 py-5">
				<h3 class="text-foreground font-display text-lg font-semibold">Selling Units</h3>
				<p class="text-muted-foreground text-sm">
					Units this product can be sold in. 1 selling unit = factor ×
					{baseUnit?.Name?.toLowerCase() ?? 'base unit'}.
				</p>
			</div>
			<div class="space-y-4 p-6">
				<div class="border-border bg-muted/40 rounded-md border px-3 py-2">
					<p class="text-muted-foreground text-xs">
						Example: stocked in {baseUnit?.Name?.toLowerCase() ?? 'feet'}, a coil with factor 100
						means 1 coil = 100 {baseUnit?.Symbol ?? 'ft'}; set its price accordingly (e.g. 45).
					</p>
				</div>

				<div class="custom-scrollbar border-border overflow-x-auto rounded-md border">
					<table class="ledger-table min-w-[40rem]">
						<thead>
							<tr>
								<th>Unit</th>
								<th>Factor</th>
								<th>Price</th>
								<th>Purchase</th>
								<th>Barcode</th>
								<th>Active</th>
								<th class="w-16"></th>
							</tr>
						</thead>
						<tbody>
							{#if productUnits.length === 0}
								<tr>
									<td
										colspan="7"
										class="text-muted-foreground py-8 text-center text-sm"
									>
										No selling units yet.
									</td>
								</tr>
							{:else}
								{#each productUnits as pu (pu.UnitID)}
									<tr>
										<td>
											<span class="text-foreground text-sm font-medium"
												>{pu.Unit?.Name ?? '—'}</span
											>
											<span class="text-muted-foreground text-xs">
												({pu.Unit?.Symbol ?? ''})</span
											>
										</td>
										<td class="num-mono">{pu.Factor}</td>
										<td class="num-mono">{pu.Price}</td>
										<td class="num-mono">{pu.PurchasePrice}</td>
										<td class="num-mono">{pu.Barcode || '—'}</td>
										<td>
											<span class="badge-{pu.IsActive ? 'success' : 'neutral'}">
												{pu.IsActive ? 'Active' : 'Inactive'}
											</span>
										</td>
										<td>
											<Button
												type="button"
												size="sm"
												variant="ghost"
												class="text-danger hover:text-danger h-8 px-2"
												onclick={() => deleteSellingUnit(pu.UnitID)}
											>
												Delete
											</Button>
										</td>
									</tr>
								{/each}
							{/if}
						</tbody>
					</table>
				</div>

				<div class="border-border bg-muted/30 space-y-3 rounded-md border p-4">
					<p class="overline">Add selling unit</p>
					<div class="select-wrapper">
						<select
							value={newUnitId}
							onchange={(e) => onNewUnitSelect(e.currentTarget.value)}
						>
							<option value="">Select unit</option>
							{#each dimensionGroups as group (group.name)}
								<optgroup label={group.name}>
									{#each group.units as unit (unit.ID)}
										<option value={String(unit.ID)}>{unit.Name} ({unit.Symbol})</option>
									{/each}
								</optgroup>
							{/each}
						</select>
					</div>
					<div class="grid grid-cols-1 gap-3 sm:grid-cols-2">
						<Input
							type="number"
							min="0"
							step="any"
							class="{inputClass} num-mono"
							placeholder="Factor (1 unit = X base)"
							bind:value={newFactor}
						/>
						<Input
							type="number"
							min="0"
							step="0.01"
							class="{inputClass} num-mono"
							placeholder="Price"
							bind:value={newPrice}
						/>
						<Input
							type="number"
							min="0"
							step="0.01"
							class="{inputClass} num-mono"
							placeholder="Purchase price"
							bind:value={newPurchasePrice}
						/>
						<Input
							class="{inputClass} num-mono"
							placeholder="Barcode"
							bind:value={newBarcode}
						/>
					</div>
					<div class="flex items-center justify-between gap-3">
						<div class="flex items-center gap-2">
							<Switch id="new-unit-active" bind:checked={newIsActive} />
							<label class="text-foreground cursor-pointer text-sm" for="new-unit-active"
								>Active</label
							>
						</div>
						<Button
							type="button"
							class={btnPrimary}
							disabled={savingUnit || !newUnitId}
							onclick={saveSellingUnit}
						>
							{savingUnit ? 'Saving...' : 'Add selling unit'}
						</Button>
					</div>
				</div>
			</div>
		</div>
	{/if}
</div>
