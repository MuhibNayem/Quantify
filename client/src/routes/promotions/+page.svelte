<script lang="ts">
	import { onMount } from 'svelte';
	import { toast } from 'svelte-sonner';
	import { fly } from 'svelte/transition';
	import { Input } from '$lib/components/ui/input';
	import { Button } from '$lib/components/ui/button';
	import { Select } from '$lib/components/ui/select';
	import {
		Dialog,
		DialogContent,
		DialogDescription,
		DialogHeader,
		DialogTitle,
		DialogFooter
	} from '$lib/components/ui/dialog';
	import { Label } from '$lib/components/ui/label';
	import { Switch } from '$lib/components/ui/switch';
	import HelpTip from '$lib/components/ui/help-tip.svelte';
	import { formatDateTime, formatCurrency } from '$lib/utils';
	import { cn } from '$lib/utils';
	import { promotionsApi, productsApi, categoriesApi, subCategoriesApi } from '$lib/api/resources';
	import type { Promotion, Product, Category, SubCategory } from '$lib/types';
	import {
		Plus,
		Pencil,
		Trash2,
		Search,
		Tag,
		CalendarRange,
		Percent,
		DollarSign,
		AlertCircle,
		Layers,
		CheckCircle2
	} from 'lucide-svelte';
	import { superForm } from 'sveltekit-superforms';
	import { zod4 } from 'sveltekit-superforms/adapters';
	import { promotionSchema } from '$lib/schemas/promotion';

	const inputClass =
		'liquid-input h-11 w-full rounded-md bg-card px-3 text-sm text-foreground placeholder:text-muted-foreground focus:border-primary focus:ring-0';

	let promotions = $state<Promotion[]>([]);
	let loading = $state(false);
	let isDialogOpen = $state(false);
	let editingId = $state<number | null>(null);

	// Selection Data
	let products = $state<Product[]>([]);
	let categories = $state<Category[]>([]);
	let subCategories = $state<SubCategory[]>([]);
	let productSearch = $state('');

	const defaultPromotion = () => ({
		name: '',
		description: '',
		discountType: 'PERCENTAGE' as 'PERCENTAGE' | 'FIXED_AMOUNT',
		discountValue: 0,
		startDate: new Date().toISOString().slice(0, 16),
		endDate: new Date(Date.now() + 7 * 24 * 60 * 60 * 1000).toISOString().slice(0, 16),
		priority: 0,
		isActive: true,
		targetType: 'CATEGORY' as 'PRODUCT' | 'CATEGORY' | 'SUBCATEGORY',
		productId: null as number | null,
		categoryId: null as number | null,
		subCategoryId: null as number | null
	});

	const { form, errors, enhance, reset } = superForm(defaultPromotion(), {
		validators: zod4(promotionSchema),
		onSubmit: async ({ cancel }) => {
			cancel();
			const data = $form;
			try {
				const payload: any = {
					name: data.name,
					description: data.description,
					discountType: data.discountType,
					discountValue: Number(data.discountValue),
					startDate: new Date(data.startDate).toISOString(),
					endDate: new Date(data.endDate).toISOString(),
					priority: Number(data.priority),
					isActive: data.isActive
				};

				if (data.targetType === 'PRODUCT') {
					if (!data.productId) {
						toast.error('Please select a product');
						return;
					}
					payload.productId = Number(data.productId);
				} else if (data.targetType === 'SUBCATEGORY') {
					if (!data.subCategoryId) {
						toast.error('Please select a sub-category');
						return;
					}
					payload.subCategoryId = Number(data.subCategoryId);
					payload.categoryId = Number(data.categoryId);
				} else {
					if (!data.categoryId) {
						toast.error('Please select a category');
						return;
					}
					payload.categoryId = Number(data.categoryId);
				}

				if (editingId) {
					await promotionsApi.update(editingId, payload);
					toast.success('Promotion updated');
				} else {
					await promotionsApi.create(payload);
					toast.success('Promotion created');
				}
				isDialogOpen = false;
				loadPromotions();
			} catch (error: any) {
				const msg = error.response?.data?.error || 'Operation failed';
				toast.error(msg);
			}
		}
	});

	// Derived Options for Select Components
	const categoryOptions = $derived(categories.map((c) => ({ value: String(c.ID), label: c.Name })));
	const subCategoryOptions = $derived(
		subCategories.map((s) => ({ value: String(s.ID), label: s.Name }))
	);
	const discountTypeOptions = [
		{ value: 'PERCENTAGE', label: 'Percentage (%)' },
		{ value: 'FIXED_AMOUNT', label: 'Fixed Amount ($)' }
	];

	onMount(async () => {
		await loadPromotions();
		await loadCategories();
	});

	async function loadPromotions() {
		loading = true;
		try {
			promotions = await promotionsApi.list();
		} catch (error) {
			toast.error('Failed to load promotions');
		} finally {
			loading = false;
		}
	}

	async function loadCategories() {
		try {
			categories = await categoriesApi.list();
		} catch (e) {
			console.error(e);
		}
	}

	async function handleCategorySelect(catIdStr: string) {
		const catId = Number(catIdStr);
		$form.categoryId = catId;
		$form.subCategoryId = null; // Reset subcat

		if (catId) {
			try {
				subCategories = await subCategoriesApi.list(catId);
			} catch (e) {
				console.error(e);
			}
		} else {
			subCategories = [];
		}
	}

	async function searchProducts(term: string) {
		if (term.length < 2) return;
		try {
			const res = await productsApi.list(1, 20, term);
			products = res.products;
		} catch (e) {
			console.error(e);
		}
	}

	function resetForm() {
		editingId = null;
		reset({ data: defaultPromotion() });
		subCategories = [];
		products = [];
		productSearch = '';
	}

	function openCreateDialog() {
		resetForm();
		isDialogOpen = true;
	}

	function openEditDialog(promo: Promotion) {
		editingId = promo.ID;
		if (promo.CategoryID) {
			subCategoriesApi.list(promo.CategoryID).then((subs) => {
				subCategories = subs;
			});
		} else {
			subCategories = [];
		}

		reset({
			data: {
				name: promo.Name,
				description: promo.Description || '',
				discountType: promo.DiscountType,
				discountValue: promo.DiscountValue,
				startDate: promo.StartDate ? promo.StartDate.slice(0, 16) : '',
				endDate: promo.EndDate ? promo.EndDate.slice(0, 16) : '',
				priority: promo.Priority,
				isActive: promo.IsActive,
				targetType: promo.ProductID ? 'PRODUCT' : promo.SubCategoryID ? 'SUBCATEGORY' : 'CATEGORY',
				productId: promo.ProductID || null,
				categoryId: promo.CategoryID || null,
				subCategoryId: promo.SubCategoryID || null
			}
		});

		if (promo.Product) {
			products = [promo.Product];
			productSearch = promo.Product.Name;
		} else {
			products = [];
			productSearch = '';
		}

		isDialogOpen = true;
	}

	async function handleDelete(id: number) {
		if (!confirm('Are you sure you want to delete this promotion?')) return;
		try {
			await promotionsApi.delete(id);
			promotions = promotions.filter((p) => p.ID !== id);
			toast.success('Promotion deleted');
		} catch (error) {
			toast.error('Failed to delete promotion');
		}
	}

	function getTargetLabel(p: Promotion) {
		if (p.Product) return { type: 'Product', name: p.Product.Name };
		if (p.SubCategory) return { type: 'Sub-Category', name: p.SubCategory.Name };
		if (p.Category) return { type: 'Category', name: p.Category.Name };
		return { type: 'Global', name: 'All Items' };
	}

	function onCategoryChange(e: CustomEvent<string>) {
		handleCategorySelect(e.detail);
	}

	function onSubCategoryChange(e: CustomEvent<string>) {
		$form.subCategoryId = Number(e.detail);
	}

	function onDiscountTypeChange(e: CustomEvent<string>) {
		$form.discountType = e.detail as any;
	}
</script>

<div class="bg-background relative min-h-screen overflow-hidden p-6 lg:p-10">
	<div class="mx-auto max-w-7xl space-y-8">
		<!-- Header -->
		<div
			class="flex flex-col gap-6 md:flex-row md:items-center md:justify-between"
			data-animate="fade-up"
		>
			<div class="space-y-1">
				<p class="overline">Pricing · Campaigns</p>
				<h1 class="font-display text-foreground text-3xl font-bold tracking-tight">
					Promotions & Discounts
				</h1>
				<p class="text-muted-foreground">
					Manage dynamic pricing rules, sales events, and granular discounts.
				</p>
			</div>

			<Button
				onclick={openCreateDialog}
				class="bg-primary text-primary-foreground rounded-md px-6 font-semibold shadow-md hover:bg-primary/90"
			>
				<Plus class="mr-2 h-4 w-4" /> New Promotion
			</Button>
		</div>

		{#if loading}
			<div class="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
				{#each Array(6) as _}
					<div class="bg-muted h-48 animate-pulse rounded-lg"></div>
				{/each}
			</div>
		{:else}
			<div
				class="grid gap-6 md:grid-cols-2 lg:grid-cols-3"
				in:fly={{ y: 20, duration: 400, delay: 0 }}
			>
				{#each promotions as promo (promo.ID)}
					{@const target = getTargetLabel(promo)}
					<div class="liquid-hoverable group relative flex flex-col overflow-hidden rounded-lg">
						<!-- Active Indicator -->
						<div class="absolute top-0 right-0 p-4">
							<div
								class={cn(
									'h-2.5 w-2.5 rounded-full',
									promo.IsActive ? 'bg-success animate-pulse' : 'bg-border'
								)}
							></div>
						</div>

						<!-- Card Header -->
						<div class="border-border border-b p-5">
							<div class="flex items-start justify-between pr-6">
								<div>
									<h3 class="text-foreground line-clamp-1 text-lg font-bold">{promo.Name}</h3>
									<p class="text-muted-foreground mt-1 line-clamp-2 min-h-[1.5em] text-xs font-medium">
										{promo.Description || 'No description provided'}
									</p>
								</div>
							</div>

							<div class="mt-4 flex flex-wrap gap-2">
								<span class="badge-info">
									<Layers class="h-3 w-3" />
									{target.type}
								</span>
								<span class="badge-neutral">{target.name}</span>
							</div>
						</div>

						<!-- Card Content -->
						<div class="flex-1 space-y-4 p-5">
							<div class="bg-muted flex items-center justify-between rounded-md p-3">
								<span class="text-muted-foreground text-sm font-medium">Discount</span>
								<span class="num-mono text-primary text-lg font-bold">
									{#if promo.DiscountType === 'PERCENTAGE'}
										{promo.DiscountValue}% OFF
									{:else}
										-{formatCurrency(promo.DiscountValue)}
									{/if}
								</span>
							</div>

							<div class="space-y-2 text-sm">
								<div class="flex items-center justify-between">
									<span class="text-muted-foreground flex items-center">
										<CalendarRange class="mr-2 h-3.5 w-3.5" /> Period
									</span>
									<span class="num-mono text-foreground text-xs font-medium">
										{formatDateTime(promo.StartDate).split(',')[0]} - {formatDateTime(
											promo.EndDate
										).split(',')[0]}
									</span>
								</div>
								<div class="flex items-center justify-between">
									<span class="text-muted-foreground flex items-center">
										<AlertCircle class="mr-2 h-3.5 w-3.5" /> Priority
									</span>
									<span class="badge-neutral num-mono">{promo.Priority}</span>
								</div>
							</div>
						</div>

						<!-- Card Footer -->
						<div
							class="border-border bg-muted/40 mt-auto flex items-center justify-between border-t p-4"
						>
							<span
								class={cn(
									'overline',
									promo.IsActive ? 'text-success' : 'text-muted-foreground'
								)}
							>
								{promo.IsActive ? 'Active Now' : 'Inactive'}
							</span>
							<div class="flex gap-1">
								<Button
									variant="ghost"
									size="icon"
									class="text-muted-foreground hover:bg-muted hover:text-primary h-8 w-8 rounded-md"
									onclick={() => openEditDialog(promo)}
								>
									<Pencil class="h-4 w-4" />
								</Button>
								<Button
									variant="ghost"
									size="icon"
									class="text-muted-foreground hover:bg-danger/10 hover:text-danger h-8 w-8 rounded-md"
									onclick={() => handleDelete(promo.ID)}
								>
									<Trash2 class="h-4 w-4" />
								</Button>
							</div>
						</div>
					</div>
				{:else}
					<div class="col-span-full py-20 text-center">
						<div class="bg-muted text-muted-foreground mx-auto flex h-20 w-20 items-center justify-center rounded-lg">
							<Tag class="h-10 w-10" />
						</div>
						<h3 class="text-foreground mt-4 text-lg font-bold">No Promotions Found</h3>
						<p class="text-muted-foreground">Create your first promotion to boost sales.</p>
						<Button
							onclick={openCreateDialog}
							variant="outline"
							class="glass-button text-foreground mt-4 rounded-md"
						>
							Create Promotion
						</Button>
					</div>
				{/each}
			</div>
		{/if}
	</div>

	<!-- Create/Edit Dialog -->
	<Dialog bind:open={isDialogOpen}>
		<DialogContent
			class="bg-card border-border gap-0 overflow-visible rounded-lg border p-0 shadow-lg sm:max-w-[650px]"
		>
			<!-- Header -->
			<div class="border-border bg-muted/50 rounded-t-lg border-b px-8 pt-8 pb-6">
				<DialogHeader>
					<DialogTitle class="font-display text-foreground text-2xl font-bold">
						{editingId ? 'Edit Promotion' : 'Create Promotion'}</DialogTitle
					>
					<DialogDescription class="text-muted-foreground font-medium">
						Configure discount rules and targeting. Higher priority rules override overlapping ones.
					</DialogDescription>
				</DialogHeader>
			</div>

			<form
				method="POST"
				use:enhance
				id="promotion-form"
				class="custom-scrollbar grid max-h-[70vh] gap-6 overflow-y-auto px-8 py-6"
			>
				<div class="grid grid-cols-2 gap-6">
					<div class="space-y-2">
						<Label class="overline">Name</Label>
						<Input bind:value={$form.name} placeholder="e.g. Summer Sale" class={inputClass} />
						{#if $errors.name}
							<p class="text-danger text-xs font-semibold">{$errors.name}</p>
						{/if}
					</div>
					<div class="space-y-2">
						<Label class="overline">Priority</Label>
						<Input
							type="number"
							bind:value={$form.priority}
							min="0"
							class="{inputClass} num-mono"
						/>
					</div>
				</div>

				<div class="space-y-2">
					<Label class="overline">Description</Label>
					<Input
						bind:value={$form.description}
						placeholder="Internal note..."
						class={inputClass}
					/>
				</div>

				<div class="border-border bg-muted/40 grid grid-cols-2 gap-6 rounded-md border p-5">
					<div class="space-y-2">
						<Label class="text-foreground flex items-center gap-1 font-semibold">
							Discount Type
							<HelpTip text="PERCENTAGE applies a % off the price; FIXED_AMOUNT subtracts a flat currency value." />
						</Label>
						<Select
							value={$form.discountType}
							options={discountTypeOptions}
							on:change={onDiscountTypeChange}
							placeholder="Select Type"
							style="liquid-input h-11 w-full rounded-md bg-card px-3 text-sm text-foreground"
						/>
					</div>
					<div class="space-y-2">
						<Label class="text-foreground font-semibold">Value</Label>
						<div class="relative">
							<Input
								type="number"
								bind:value={$form.discountValue}
								min="0"
								step="0.01"
								class="liquid-input num-mono text-primary h-11 w-full rounded-md bg-card pl-10 text-lg font-bold focus:border-primary focus:ring-0"
							/>
							<div
								class="text-muted-foreground pointer-events-none absolute top-1/2 left-3 -translate-y-1/2"
							>
								{#if $form.discountType === 'PERCENTAGE'}
									<Percent class="h-5 w-5" />
								{:else}
									<DollarSign class="h-5 w-5" />
								{/if}
							</div>
						</div>
					</div>
				</div>

				<div class="grid grid-cols-2 gap-6">
					<div class="space-y-2">
						<Label class="overline">Start Date</Label>
						<Input
							type="datetime-local"
							bind:value={$form.startDate}
							class="{inputClass} num-mono"
						/>
						{#if $errors.startDate}
							<p class="text-danger text-xs font-semibold">{$errors.startDate}</p>
						{/if}
					</div>
					<div class="space-y-2">
						<Label class="overline">End Date</Label>
						<Input
							type="datetime-local"
							bind:value={$form.endDate}
							class="{inputClass} num-mono"
						/>
						{#if $errors.endDate}
							<p class="text-danger text-xs font-semibold">{$errors.endDate}</p>
						{/if}
					</div>
				</div>

				<div class="border-border bg-card space-y-4 rounded-md border p-6 shadow-xs">
					<div class="mb-2 flex items-center gap-2">
						<Layers class="text-primary h-5 w-5" />
						<Label class="text-foreground text-base font-bold">Targeting Rules</Label>
					</div>

					<div class="bg-muted flex rounded-md p-1.5">
						{#each [{ val: 'CATEGORY', label: 'Category' }, { val: 'SUBCATEGORY', label: 'Sub-Category' }, { val: 'PRODUCT', label: 'Product' }] as option}
							<label class="relative flex-1 cursor-pointer">
								<input
									type="radio"
									class="peer sr-only"
									name="target"
									value={option.val}
									bind:group={$form.targetType}
								/>
								<div
									class="text-muted-foreground peer-checked:bg-card peer-checked:text-primary peer-checked:shadow-sm flex items-center justify-center rounded-md py-2.5 text-sm font-semibold transition-all duration-200"
								>
									{option.label}
								</div>
							</label>
						{/each}
					</div>

					<div class="min-h-[80px] space-y-4 pt-2">
						{#if $form.targetType === 'CATEGORY' || $form.targetType === 'SUBCATEGORY'}
							<div in:fly={{ y: 5, duration: 200 }} class="space-y-2">
								<Label class="overline">Select Category</Label>
								<Select
									value={$form.categoryId ? String($form.categoryId) : ''}
									options={categoryOptions}
									on:change={onCategoryChange}
									placeholder="Choose a category..."
									style="liquid-input h-11 w-full rounded-md bg-card px-3 text-sm text-foreground"
								/>
							</div>
						{/if}

						{#if $form.targetType === 'SUBCATEGORY'}
							<div in:fly={{ y: 5, duration: 200 }} class="space-y-2">
								<Label class="overline">Select Sub-Category</Label>
								<Select
									value={$form.subCategoryId ? String($form.subCategoryId) : ''}
									options={subCategoryOptions}
									on:change={onSubCategoryChange}
									disabled={!$form.categoryId || subCategories.length === 0}
									placeholder="Choose a sub-category..."
									style="liquid-input h-11 w-full rounded-md bg-card px-3 text-sm text-foreground"
								/>
							</div>
						{/if}

						{#if $form.targetType === 'PRODUCT'}
							<div in:fly={{ y: 5, duration: 200 }} class="space-y-2">
								<Label class="overline">Find Product</Label>
								<div class="group relative">
									<Search
										class="text-muted-foreground absolute top-3.5 left-3 h-4 w-4"
									/>
									<Input
										class="{inputClass} pl-10"
										placeholder="Search by name or SKU..."
										bind:value={productSearch}
										oninput={(e) => searchProducts(e.currentTarget.value)}
									/>
								</div>

								{#if products.length > 0}
									<div class="text-muted-foreground mt-2 px-1 text-xs font-medium">
										Found {products.length} matches
									</div>
									<div
										class="custom-scrollbar border-border bg-card mt-2 max-h-[160px] overflow-y-auto rounded-md border p-1 shadow-md"
									>
										{#each products as textProd}
											<button
												class="group flex w-full items-center justify-between rounded-md px-3 py-2.5 text-left text-sm transition-colors hover:bg-muted {$form.productId ===
												textProd.ID
													? 'bg-accent ring-primary/30 ring-1'
													: ''}"
												onclick={() => ($form.productId = textProd.ID)}
											>
												<span class="text-foreground font-medium">{textProd.Name}</span>
												<span class="badge-neutral num-mono">{textProd.SKU}</span>
											</button>
										{/each}
									</div>
								{/if}
								{#if $form.productId}
									<div
										class="text-success mt-3 flex items-center justify-between rounded-md border border-success/25 bg-success/10 p-3"
									>
										<div class="flex items-center gap-2 text-sm font-medium">
											<CheckCircle2 class="text-success h-4 w-4" />
											Product Selected
										</div>
										<span
											class="num-mono border-success/30 bg-card rounded-md border px-2 py-0.5 text-xs"
											>ID: {$form.productId}</span
										>
									</div>
								{/if}
							</div>
						{/if}
					</div>
				</div>

				<div
					class={cn(
						'flex items-center space-x-4 rounded-md border p-4 transition-colors duration-200',
						$form.isActive ? 'border-success/30 bg-success/10' : 'border-border bg-muted/50'
					)}
				>
					<Switch bind:checked={$form.isActive} id="active-mode" />
					<div class="flex flex-col">
						<Label
							for="active-mode"
							class={cn(
								'cursor-pointer text-sm font-bold transition-colors',
								$form.isActive ? 'text-success' : 'text-muted-foreground'
							)}
						>
							{$form.isActive ? 'Promotion Active' : 'Promotion Inactive'}
						</Label>
						<span
							class={cn(
								'text-xs transition-colors',
								$form.isActive ? 'text-success/80' : 'text-muted-foreground'
							)}
						>
							{$form.isActive
								? 'Live immediately upon saving'
								: 'Saved as draft, not visible to customers'}
						</span>
					</div>
				</div>
			</form>

			<!-- Footer -->
			<div class="border-border bg-muted/50 rounded-b-lg border-t p-6">
				<DialogFooter class="gap-3 sm:gap-0">
					<Button
						variant="outline"
						onclick={() => (isDialogOpen = false)}
						class="glass-button text-foreground rounded-md font-medium"
						>Cancel</Button
					>
					<Button
						type="submit"
						form="promotion-form"
						class="bg-primary text-primary-foreground rounded-md px-8 font-semibold shadow-md hover:bg-primary/90"
					>
						{editingId ? 'Save Changes' : 'Create Promotion'}
					</Button>
				</DialogFooter>
			</div>
		</DialogContent>
	</Dialog>
</div>
