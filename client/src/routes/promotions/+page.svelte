<script lang="ts">
	import { onMount } from 'svelte';
	import { toast } from 'svelte-sonner';
	import { fly } from 'svelte/transition';
	import { Input } from '$lib/components/ui/input';
	import { Button } from '$lib/components/ui/button';
	import { Badge } from '$lib/components/ui/badge';
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

<div class="relative min-h-screen overflow-hidden bg-slate-50/50 p-6 lg:p-10">
	<!-- Background -->
	<div
		class="absolute inset-0 -z-10 bg-[radial-gradient(ellipse_at_top,_var(--tw-gradient-stops))] from-purple-100/20 via-slate-50/20 to-white/20"
	></div>

	<div class="mx-auto max-w-7xl space-y-8">
		<!-- Header -->
		<div class="flex flex-col gap-6 md:flex-row md:items-center md:justify-between">
			<div class="space-y-1">
				<h1
					class="bg-gradient-to-r from-slate-900 via-purple-900 to-slate-900 bg-clip-text text-3xl font-bold tracking-tight text-transparent"
				>
					Promotions & Discounts
				</h1>
				<p class="text-slate-500">
					Manage dynamic pricing rules, sales events, and granular discounts.
				</p>
			</div>

			<Button
				onclick={openCreateDialog}
				class="rounded-xl bg-gradient-to-r from-purple-600 to-indigo-600 px-6 font-semibold text-white shadow-lg shadow-purple-500/25 transition-all hover:scale-105 hover:shadow-purple-500/40"
			>
				<Plus class="mr-2 h-4 w-4" /> New Promotion
			</Button>
		</div>

		{#if loading}
			<div class="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
				{#each Array(6) as _}
					<div class="h-48 animate-pulse rounded-2xl bg-white/50"></div>
				{/each}
			</div>
		{:else}
			<div
				class="grid gap-6 md:grid-cols-2 lg:grid-cols-3"
				in:fly={{ y: 20, duration: 400, delay: 0 }}
			>
				{#each promotions as promo (promo.ID)}
					{@const target = getTargetLabel(promo)}
					<div
						class="liquid-hoverable group relative flex flex-col overflow-hidden rounded-2xl bg-gradient-to-br from-white/50 via-white/30 to-white/10 transition-all duration-300 hover:-translate-y-1 hover:bg-white/60 hover:shadow-xl hover:shadow-purple-500/10"
					>
						<!-- Active Indicator -->
						<div class="absolute top-0 right-0 p-4">
							<div
								class={cn(
									'h-2.5 w-2.5 rounded-full shadow-sm',
									promo.IsActive ? 'animate-pulse bg-emerald-500' : 'bg-slate-300'
								)}
							></div>
						</div>

						<!-- Card Header -->
						<div class="border-b border-slate-100/50 p-5">
							<div class="flex items-start justify-between pr-6">
								<div>
									<h3 class="line-clamp-1 text-lg font-bold text-slate-800">{promo.Name}</h3>
									<p class="mt-1 line-clamp-2 min-h-[1.5em] text-xs font-medium text-slate-400">
										{promo.Description || 'No description provided'}
									</p>
								</div>
							</div>

							<div class="mt-4 flex flex-wrap gap-2">
								<Badge
									variant="secondary"
									class="border-indigo-100 bg-indigo-50 text-indigo-700 backdrop-blur-sm"
								>
									<Layers class="mr-1 h-3 w-3" />
									{target.type}
								</Badge>
								<Badge variant="outline" class="border-slate-200 text-slate-600 backdrop-blur-sm">
									{target.name}
								</Badge>
							</div>
						</div>

						<!-- Card Content -->
						<div class="flex-1 space-y-4 p-5">
							<div
								class="flex items-center justify-between rounded-xl border border-white/50 bg-white/40 p-3 backdrop-blur-sm"
							>
								<span class="text-sm font-medium text-slate-500">Discount</span>
								<span
									class={cn(
										'text-lg font-bold',
										promo.DiscountType === 'PERCENTAGE' ? 'text-purple-600' : 'text-emerald-600'
									)}
								>
									{#if promo.DiscountType === 'PERCENTAGE'}
										{promo.DiscountValue}% OFF
									{:else}
										-{formatCurrency(promo.DiscountValue)}
									{/if}
								</span>
							</div>

							<div class="space-y-2 text-sm">
								<div class="flex items-center justify-between">
									<span class="flex items-center text-slate-500">
										<CalendarRange class="mr-2 h-3.5 w-3.5" /> Period
									</span>
									<span class="text-xs font-medium text-slate-700">
										{formatDateTime(promo.StartDate).split(',')[0]} - {formatDateTime(
											promo.EndDate
										).split(',')[0]}
									</span>
								</div>
								<div class="flex items-center justify-between">
									<span class="flex items-center text-slate-500">
										<AlertCircle class="mr-2 h-3.5 w-3.5" /> Priority
									</span>
									<Badge variant="outline" class="h-5 bg-slate-50 px-2">{promo.Priority}</Badge>
								</div>
							</div>
						</div>

						<!-- Card Footer -->
						<div
							class="mt-auto flex items-center justify-between border-t border-slate-100/50 bg-slate-50/30 p-4 backdrop-blur-sm"
						>
							<span
								class={cn(
									'text-[10px] font-bold tracking-wider uppercase',
									promo.IsActive ? 'text-emerald-600' : 'text-slate-400'
								)}
							>
								{promo.IsActive ? 'Active Now' : 'Inactive'}
							</span>
							<div class="flex gap-1">
								<Button
									variant="ghost"
									size="icon"
									class="h-8 w-8 rounded-full text-slate-500 hover:bg-purple-50 hover:text-purple-600"
									onclick={() => openEditDialog(promo)}
								>
									<Pencil class="h-4 w-4" />
								</Button>
								<Button
									variant="ghost"
									size="icon"
									class="h-8 w-8 rounded-full text-slate-500 hover:bg-red-50 hover:text-red-600"
									onclick={() => handleDelete(promo.ID)}
								>
									<Trash2 class="h-4 w-4" />
								</Button>
							</div>
						</div>
					</div>
				{:else}
					<div class="col-span-full py-20 text-center">
						<div
							class="mx-auto flex h-20 w-20 items-center justify-center rounded-3xl bg-purple-50 text-purple-200"
						>
							<Tag class="h-10 w-10" />
						</div>
						<h3 class="mt-4 text-lg font-bold text-slate-700">No Promotions Found</h3>
						<p class="text-slate-500">Create your first promotion to boost sales.</p>
						<Button
							onclick={openCreateDialog}
							variant="outline"
							class="mt-4 border-purple-200 text-purple-700 hover:bg-purple-50"
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
			class="gap-0 overflow-visible rounded-[2rem] border-white/20 bg-white/60 p-0 shadow-[0_20px_50px_-12px_rgba(0,0,0,0.2)] backdrop-blur-3xl sm:max-w-[650px]"
		>
			<!-- Glassy Header -->
			<div
				class="rounded-t-[2rem] border-b border-white/30 bg-gradient-to-r from-indigo-50/40 to-purple-50/40 px-8 pt-8 pb-6"
			>
				<DialogHeader>
					<DialogTitle
						class="bg-gradient-to-r from-indigo-900 to-purple-900 bg-clip-text text-2xl font-bold text-transparent"
					>
						{editingId ? 'Edit Promotion' : 'Create Promotion'}</DialogTitle
					>
					<DialogDescription class="font-medium text-slate-500">
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
						<Label class="text-xs font-bold tracking-wider text-slate-500 uppercase">Name</Label>
						<Input
							bind:value={$form.name}
							placeholder="e.g. Summer Sale"
							class="h-11 rounded-xl border-white/40 bg-white/40 font-medium placeholder:text-slate-400 focus:border-indigo-500 focus:ring-indigo-500/20"
						/>
						{#if $errors.name}
							<p class="text-xs font-semibold text-rose-600">{$errors.name}</p>
						{/if}
					</div>
					<div class="space-y-2">
						<Label class="text-xs font-bold tracking-wider text-slate-500 uppercase">Priority</Label
						>
						<Input
							type="number"
							bind:value={$form.priority}
							min="0"
							class="h-11 rounded-xl border-white/40 bg-white/40 font-medium focus:border-indigo-500 focus:ring-indigo-500/20"
						/>
					</div>
				</div>

				<div class="space-y-2">
					<Label class="text-xs font-bold tracking-wider text-slate-500 uppercase"
						>Description</Label
					>
					<Input
						bind:value={$form.description}
						placeholder="Internal note..."
						class="h-11 rounded-xl border-white/40 bg-white/40 placeholder:text-slate-400 focus:border-indigo-500 focus:ring-indigo-500/20"
					/>
				</div>

				<div
					class="grid grid-cols-2 gap-6 rounded-2xl border border-white/30 bg-gradient-to-br from-white/40 to-indigo-50/10 p-5 shadow-sm"
				>
					<div class="space-y-2">
						<Label class="font-semibold text-indigo-900">Discount Type</Label>
						<Select
							value={$form.discountType}
							options={discountTypeOptions}
							on:change={onDiscountTypeChange}
							placeholder="Select Type"
							style="bg-white/60 border-white/40 h-11 rounded-xl"
						/>
					</div>
					<div class="space-y-2">
						<Label class="font-semibold text-indigo-900">Value</Label>
						<div class="relative">
							<Input
								type="number"
								bind:value={$form.discountValue}
								min="0"
								step="0.01"
								class="h-11 rounded-xl border-white/40 bg-white/80 pl-10 text-lg font-bold text-indigo-600 focus:border-indigo-500 focus:ring-indigo-500/20"
							/>
							<div
								class="pointer-events-none absolute top-1/2 left-3 -translate-y-1/2 text-indigo-400"
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
						<Label class="text-xs font-bold tracking-wider text-slate-500 uppercase"
							>Start Date</Label
						>
						<Input
							type="datetime-local"
							bind:value={$form.startDate}
							class="h-11 rounded-xl border-white/40 bg-white/40 text-sm font-medium text-slate-600"
						/>
						{#if $errors.startDate}
							<p class="text-xs font-semibold text-rose-600">{$errors.startDate}</p>
						{/if}
					</div>
					<div class="space-y-2">
						<Label class="text-xs font-bold tracking-wider text-slate-500 uppercase">End Date</Label
						>
						<Input
							type="datetime-local"
							bind:value={$form.endDate}
							class="h-11 rounded-xl border-white/40 bg-white/40 text-sm font-medium text-slate-600"
						/>
						{#if $errors.endDate}
							<p class="text-xs font-semibold text-rose-600">{$errors.endDate}</p>
						{/if}
					</div>
				</div>

				<div class="space-y-4 rounded-2xl border border-white/40 bg-white/20 p-6 shadow-sm">
					<div class="mb-2 flex items-center gap-2">
						<Layers class="h-5 w-5 text-indigo-600" />
						<Label class="text-base font-bold text-slate-800">Targeting Rules</Label>
					</div>

					<div class="flex rounded-xl bg-slate-100/50 p-1.5 backdrop-blur-md">
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
									class="flex items-center justify-center rounded-lg py-2.5 text-sm font-semibold text-slate-500 transition-all duration-300 peer-checked:bg-white peer-checked:text-indigo-600 peer-checked:shadow-sm"
								>
									{option.label}
								</div>
							</label>
						{/each}
					</div>

					<div class="min-h-[80px] space-y-4 pt-2">
						{#if $form.targetType === 'CATEGORY' || $form.targetType === 'SUBCATEGORY'}
							<div in:fly={{ y: 5, duration: 200 }} class="space-y-2">
								<Label class="text-xs font-semibold text-slate-500 uppercase">Select Category</Label
								>
								<Select
									value={$form.categoryId ? String($form.categoryId) : ''}
									options={categoryOptions}
									on:change={onCategoryChange}
									placeholder="Choose a category..."
									style="bg-white/80 h-11 rounded-xl border-white/40"
								/>
							</div>
						{/if}

						{#if $form.targetType === 'SUBCATEGORY'}
							<div in:fly={{ y: 5, duration: 200 }} class="space-y-2">
								<Label class="text-xs font-semibold text-slate-500 uppercase"
									>Select Sub-Category</Label
								>
								<Select
									value={$form.subCategoryId ? String($form.subCategoryId) : ''}
									options={subCategoryOptions}
									on:change={onSubCategoryChange}
									disabled={!$form.categoryId || subCategories.length === 0}
									placeholder="Choose a sub-category..."
									style="bg-white/80 h-11 rounded-xl border-white/40"
								/>
							</div>
						{/if}

						{#if $form.targetType === 'PRODUCT'}
							<div in:fly={{ y: 5, duration: 200 }} class="space-y-2">
								<Label class="text-xs font-semibold text-slate-500 uppercase">Find Product</Label>
								<div class="group relative">
									<Search
										class="absolute top-3.5 left-3 h-4 w-4 text-slate-400 transition-colors group-focus-within:text-indigo-500"
									/>
									<Input
										class="h-11 rounded-xl border-white/40 bg-white/60 pl-10 focus:border-indigo-500 focus:ring-indigo-500/20"
										placeholder="Search by name or SKU..."
										bind:value={productSearch}
										oninput={(e) => searchProducts(e.currentTarget.value)}
									/>
								</div>

								{#if products.length > 0}
									<div class="mt-2 px-1 text-xs font-medium text-slate-400">
										Found {products.length} matches
									</div>
									<div
										class="custom-scrollbar mt-2 max-h-[160px] overflow-y-auto rounded-xl border border-slate-100/50 bg-white/60 p-1 shadow-lg backdrop-blur-md"
									>
										{#each products as textProd}
											<button
												class="group flex w-full items-center justify-between rounded-lg px-3 py-2.5 text-left text-sm transition-all hover:bg-indigo-50/50 {$form.productId ===
												textProd.ID
													? 'bg-indigo-50 ring-1 ring-indigo-200'
													: ''}"
												onclick={() => ($form.productId = textProd.ID)}
											>
												<span
													class="font-medium text-slate-700 transition-colors group-hover:text-indigo-700"
													>{textProd.Name}</span
												>
												<Badge
													variant="secondary"
													class="bg-slate-100 font-mono text-[10px] tracking-wider text-slate-500 group-hover:bg-white"
													>{textProd.SKU}</Badge
												>
											</button>
										{/each}
									</div>
								{/if}
								{#if $form.productId}
									<div
										class="animate-in fade-in slide-in-from-top-1 mt-3 flex items-center justify-between rounded-xl border border-emerald-100 bg-emerald-50/50 p-3 text-emerald-700"
									>
										<div class="flex items-center gap-2 text-sm font-medium">
											<CheckCircle2 class="h-4 w-4 text-emerald-600" />
											Product Selected
										</div>
										<span
											class="rounded-md border border-emerald-200/50 bg-white/50 px-2 py-0.5 font-mono text-xs"
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
						'flex items-center space-x-4 rounded-2xl border p-4 shadow-sm transition-all duration-300',
						$form.isActive
							? 'border-emerald-200/60 bg-emerald-50/60 shadow-emerald-500/10'
							: 'border-slate-200/60 bg-slate-50/60'
					)}
				>
					<Switch bind:checked={$form.isActive} id="active-mode" />
					<div class="flex flex-col">
						<Label
							for="active-mode"
							class={cn(
								'cursor-pointer text-sm font-bold transition-colors',
								$form.isActive ? 'text-emerald-800' : 'text-slate-600'
							)}
						>
							{$form.isActive ? 'Promotion Active' : 'Promotion Inactive'}
						</Label>
						<span
							class={cn(
								'text-xs transition-colors',
								$form.isActive ? 'text-emerald-600/80' : 'text-slate-400'
							)}
						>
							{$form.isActive
								? 'Live immediately upon saving'
								: 'Saved as draft, not visible to customers'}
						</span>
					</div>
				</div>
			</form>

			<!-- Glassy Footer -->
			<div class="rounded-b-[2rem] border-t border-white/20 bg-slate-50/30 p-6 backdrop-blur-md">
				<DialogFooter class="gap-3 sm:gap-0">
					<Button
						variant="outline"
						onclick={() => (isDialogOpen = false)}
						class="rounded-xl border-white/40 bg-white/40 font-medium transition-all hover:bg-white/60 hover:text-slate-900"
						>Cancel</Button
					>
					<Button
						type="submit"
						form="promotion-form"
						class="rounded-xl bg-gradient-to-r from-indigo-600 to-purple-600 px-8 font-semibold text-white shadow-lg shadow-indigo-500/20 transition-all hover:scale-[1.02] hover:from-indigo-500 hover:to-purple-500"
					>
						{editingId ? 'Save Changes' : 'Create Promotion'}
					</Button>
				</DialogFooter>
			</div>
		</DialogContent>
	</Dialog>
</div>
