<script lang="ts">
	import { onMount } from 'svelte';
	import { toast } from 'svelte-sonner';
	import { fade, fly } from 'svelte/transition';
	import {
		Card,
		CardContent
	} from '$lib/components/ui/card';
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
	import {
		promotionsApi,
		productsApi,
		categoriesApi,
		subCategoriesApi
	} from '$lib/api/resources';
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

	let promotions = $state<Promotion[]>([]);
	let loading = $state(false);
	let isDialogOpen = $state(false);
	let editingId = $state<number | null>(null);

	// Form State
	let formData = $state({
		name: '',
		description: '',
		discountType: 'PERCENTAGE',
		discountValue: 0,
		startDate: '',
		endDate: '',
		priority: 0,
		isActive: true,
		targetType: 'CATEGORY' as 'PRODUCT' | 'CATEGORY' | 'SUBCATEGORY',
		productId: null as number | null,
		categoryId: null as number | null,
		subCategoryId: null as number | null
	});

	// Selection Data
	let products = $state<Product[]>([]);
	let categories = $state<Category[]>([]);
	let subCategories = $state<SubCategory[]>([]);
	let productSearch = $state('');

	// Derived Options for Select Components
	const categoryOptions = $derived(categories.map(c => ({ value: String(c.ID), label: c.Name })));
	const subCategoryOptions = $derived(subCategories.map(s => ({ value: String(s.ID), label: s.Name })));
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
		formData.categoryId = catId;
		formData.subCategoryId = null; // Reset subcat
		
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
		formData = {
			name: '',
			description: '',
			discountType: 'PERCENTAGE',
			discountValue: 0,
			startDate: new Date().toISOString().slice(0, 16),
			endDate: new Date(Date.now() + 7 * 24 * 60 * 60 * 1000).toISOString().slice(0, 16),
			priority: 0,
			isActive: true,
			targetType: 'CATEGORY',
			productId: null,
			categoryId: null,
			subCategoryId: null
		};
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
             subCategoriesApi.list(promo.CategoryID).then(subs => {
                 subCategories = subs;
             });
		} else {
			subCategories = [];
		}
		
		formData = {
			name: promo.Name,
			description: promo.Description || '',
			discountType: promo.DiscountType,
			discountValue: promo.DiscountValue,
			startDate: promo.StartDate ? promo.StartDate.slice(0, 16) : '',
			endDate: promo.EndDate ? promo.EndDate.slice(0, 16) : '',
			priority: promo.Priority,
			isActive: promo.IsActive,
			targetType: promo.ProductID
				? 'PRODUCT'
				: promo.SubCategoryID
					? 'SUBCATEGORY'
					: 'CATEGORY',
			productId: promo.ProductID || null,
			categoryId: promo.CategoryID || null,
			subCategoryId: promo.SubCategoryID || null
		};

		if (promo.Product) {
			products = [promo.Product];
			productSearch = promo.Product.Name;
		} else {
			products = [];
			productSearch = '';
		}

		isDialogOpen = true;
	}

	async function handleSubmit() {
		try {
			const payload: any = {
				name: formData.name,
				description: formData.description,
				discountType: formData.discountType,
				discountValue: Number(formData.discountValue),
				startDate: new Date(formData.startDate).toISOString(),
				endDate: new Date(formData.endDate).toISOString(),
				priority: Number(formData.priority),
				isActive: formData.isActive
			};

			if (formData.targetType === 'PRODUCT') {
				if (!formData.productId) {
					toast.error('Please select a product');
					return;
				}
				payload.productId = Number(formData.productId);
			} else if (formData.targetType === 'SUBCATEGORY') {
				if (!formData.subCategoryId) {
					toast.error('Please select a sub-category');
					return;
				}
				payload.subCategoryId = Number(formData.subCategoryId);
				payload.categoryId = Number(formData.categoryId);
			} else {
				if (!formData.categoryId) {
					toast.error('Please select a category');
					return;
				}
				payload.categoryId = Number(formData.categoryId);
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
		formData.subCategoryId = Number(e.detail);
	}
	
	function onDiscountTypeChange(e: CustomEvent<string>) {
		formData.discountType = e.detail as any;
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
						class="liquid-hoverable group relative flex flex-col overflow-hidden rounded-2xl bg-gradient-to-br from-white/50 via-white/30 to-white/10 transition-all duration-300 hover:bg-white/60 hover:-translate-y-1 hover:shadow-xl hover:shadow-purple-500/10"
					>
						<!-- Active Indicator -->
						<div class="absolute right-0 top-0 p-4">
							<div class={cn(
								"h-2.5 w-2.5 rounded-full shadow-sm",
								promo.IsActive ? "bg-emerald-500 animate-pulse" : "bg-slate-300"
							)}></div>
						</div>

						<!-- Card Header -->
						<div class="border-b border-slate-100/50 p-5">
							<div class="flex items-start justify-between pr-6">
								<div>
									<h3 class="font-bold text-slate-800 line-clamp-1 text-lg">{promo.Name}</h3>
									<p class="mt-1 text-xs font-medium text-slate-400 line-clamp-2 min-h-[1.5em]">
										{promo.Description || 'No description provided'}
									</p>
								</div>
							</div>
							
							<div class="mt-4 flex flex-wrap gap-2">
								<Badge variant="secondary" class="bg-indigo-50 text-indigo-700 border-indigo-100 backdrop-blur-sm">
									<Layers class="mr-1 h-3 w-3" /> {target.type}
								</Badge>
								<Badge variant="outline" class="border-slate-200 text-slate-600 backdrop-blur-sm">
									{target.name}
								</Badge>
							</div>
						</div>

						<!-- Card Content -->
						<div class="flex-1 p-5 space-y-4">
							<div class="flex items-center justify-between rounded-xl bg-white/40 p-3 backdrop-blur-sm border border-white/50">
								<span class="text-sm font-medium text-slate-500">Discount</span>
								<span class={cn(
									"text-lg font-bold",
									promo.DiscountType === 'PERCENTAGE' ? "text-purple-600" : "text-emerald-600"
								)}>
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
									<span class="font-medium text-slate-700 text-xs">
										{formatDateTime(promo.StartDate).split(',')[0]} - {formatDateTime(promo.EndDate).split(',')[0]}
									</span>
								</div>
								<div class="flex items-center justify-between">
									<span class="flex items-center text-slate-500">
										<AlertCircle class="mr-2 h-3.5 w-3.5" /> Priority
									</span>
									<Badge variant="outline" class="h-5 px-2 bg-slate-50">{promo.Priority}</Badge>
								</div>
							</div>
						</div>

						<!-- Card Footer -->
						<div class="mt-auto flex items-center justify-between bg-slate-50/30 p-4 backdrop-blur-sm border-t border-slate-100/50">
							<span class={cn(
								"text-[10px] font-bold uppercase tracking-wider",
								promo.IsActive ? "text-emerald-600" : "text-slate-400"
							)}>
								{promo.IsActive ? 'Active Now' : 'Inactive'}
							</span>
							<div class="flex gap-1">
								<Button
									variant="ghost"
									size="icon"
									class="h-8 w-8 text-slate-500 hover:text-purple-600 hover:bg-purple-50 rounded-full"
									onclick={() => openEditDialog(promo)}
								>
									<Pencil class="h-4 w-4" />
								</Button>
								<Button
									variant="ghost"
									size="icon"
									class="h-8 w-8 text-slate-500 hover:text-red-600 hover:bg-red-50 rounded-full"
									onclick={() => handleDelete(promo.ID)}
								>
									<Trash2 class="h-4 w-4" />
								</Button>
							</div>
						</div>
					</div>
				{:else}
					<div class="col-span-full py-20 text-center">
						<div class="mx-auto flex h-20 w-20 items-center justify-center rounded-3xl bg-purple-50 text-purple-200">
							<Tag class="h-10 w-10" />
						</div>
						<h3 class="mt-4 text-lg font-bold text-slate-700">No Promotions Found</h3>
						<p class="text-slate-500">Create your first promotion to boost sales.</p>
						<Button onclick={openCreateDialog} variant="outline" class="mt-4 border-purple-200 text-purple-700 hover:bg-purple-50">
							Create Promotion
						</Button>
					</div>
				{/each}
			</div>
		{/if}
	</div>

	<!-- Create/Edit Dialog -->
	<Dialog bind:open={isDialogOpen}>
		<DialogContent class="sm:max-w-[650px] gap-0 border-white/20 bg-white/60 backdrop-blur-3xl shadow-[0_20px_50px_-12px_rgba(0,0,0,0.2)] p-0 rounded-[2rem] overflow-visible">
			
            <!-- Glassy Header -->
            <div class="px-8 pt-8 pb-6 border-b border-white/30 bg-gradient-to-r from-indigo-50/40 to-purple-50/40 rounded-t-[2rem]">
                <DialogHeader>
                    <DialogTitle class="text-2xl font-bold bg-gradient-to-r from-indigo-900 to-purple-900 bg-clip-text text-transparent">
                        {editingId ? 'Edit Promotion' : 'Create Promotion'}</DialogTitle>
                    <DialogDescription class="text-slate-500 font-medium">
                        Configure discount rules and targeting. Higher priority rules override overlapping ones.
                    </DialogDescription>
                </DialogHeader>
            </div>

			<div class="grid gap-6 px-8 py-6 overflow-y-auto max-h-[70vh] custom-scrollbar">
				<div class="grid grid-cols-2 gap-6">
					<div class="space-y-2">
						<Label class="text-xs font-bold uppercase tracking-wider text-slate-500">Name</Label>
						<Input 
                            bind:value={formData.name} 
                            placeholder="e.g. Summer Sale" 
                            class="h-11 rounded-xl border-white/40 bg-white/40 focus:border-indigo-500 focus:ring-indigo-500/20 font-medium placeholder:text-slate-400" 
                        />
					</div>
					<div class="space-y-2">
						<Label class="text-xs font-bold uppercase tracking-wider text-slate-500">Priority</Label>
						<Input 
                            type="number" 
                            bind:value={formData.priority} 
                            min="0" 
                            class="h-11 rounded-xl border-white/40 bg-white/40 focus:border-indigo-500 focus:ring-indigo-500/20 font-medium" 
                        />
					</div>
				</div>

				<div class="space-y-2">
					<Label class="text-xs font-bold uppercase tracking-wider text-slate-500">Description</Label>
					<Input 
                        bind:value={formData.description} 
                        placeholder="Internal note..." 
                        class="h-11 rounded-xl border-white/40 bg-white/40 focus:border-indigo-500 focus:ring-indigo-500/20 placeholder:text-slate-400" 
                    />
				</div>

				<div class="grid grid-cols-2 gap-6 p-5 rounded-2xl bg-gradient-to-br from-white/40 to-indigo-50/10 border border-white/30 shadow-sm">
					<div class="space-y-2">
						<Label class="text-indigo-900 font-semibold">Discount Type</Label>
						<Select 
							value={formData.discountType} 
							options={discountTypeOptions} 
							on:change={onDiscountTypeChange} 
							placeholder="Select Type"
                            style="bg-white/60 border-white/40 h-11 rounded-xl"
						/>
					</div>
					<div class="space-y-2">
						<Label class="text-indigo-900 font-semibold">Value</Label>
						<div class="relative">
							<Input 
                                type="number" 
                                bind:value={formData.discountValue} 
                                min="0" 
                                step="0.01" 
                                class="pl-10 h-11 rounded-xl border-white/40 bg-white/80 font-bold text-lg text-indigo-600 focus:border-indigo-500 focus:ring-indigo-500/20" 
                            />
							<div class="absolute left-3 top-1/2 -translate-y-1/2 text-indigo-400 pointer-events-none">
								{#if formData.discountType === 'PERCENTAGE'}
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
						<Label class="text-xs font-bold uppercase tracking-wider text-slate-500">Start Date</Label>
						<Input 
                            type="datetime-local" 
                            bind:value={formData.startDate} 
                            class="h-11 rounded-xl border-white/40 bg-white/40 text-sm font-medium text-slate-600" 
                        />
					</div>
					<div class="space-y-2">
						<Label class="text-xs font-bold uppercase tracking-wider text-slate-500">End Date</Label>
						<Input 
                            type="datetime-local" 
                            bind:value={formData.endDate} 
                            class="h-11 rounded-xl border-white/40 bg-white/40 text-sm font-medium text-slate-600" 
                        />
					</div>
				</div>

				<div class="space-y-4 rounded-2xl border border-white/40 bg-white/20 p-6 shadow-sm">
					<div class="flex items-center gap-2 mb-2">
                        <Layers class="h-5 w-5 text-indigo-600" />
                        <Label class="text-base font-bold text-slate-800">Targeting Rules</Label>
                    </div>
					
					<div class="flex p-1.5 rounded-xl bg-slate-100/50 backdrop-blur-md">
						{#each [
                            { val: 'CATEGORY', label: 'Category' },
                            { val: 'SUBCATEGORY', label: 'Sub-Category' },
                            { val: 'PRODUCT', label: 'Product' }
                        ] as option}
                            <label class="flex-1 cursor-pointer relative">
                                <input 
                                    type="radio" 
                                    class="peer sr-only" 
                                    name="target" 
                                    value={option.val} 
                                    bind:group={formData.targetType} 
                                />
                                <div class="flex items-center justify-center py-2.5 text-sm font-semibold rounded-lg text-slate-500 transition-all duration-300 peer-checked:bg-white peer-checked:text-indigo-600 peer-checked:shadow-sm">
                                    {option.label}
                                </div>
                            </label>
                        {/each}
					</div>

					<div class="min-h-[80px] space-y-4 pt-2">
                        {#if formData.targetType === 'CATEGORY' || formData.targetType === 'SUBCATEGORY'}
                            <div in:fly={{ y: 5, duration: 200 }} class="space-y-2">
                                <Label class="text-xs font-semibold text-slate-500 uppercase">Select Category</Label>
                                <Select 
                                    value={formData.categoryId ? String(formData.categoryId) : ''} 
                                    options={categoryOptions}
                                    on:change={onCategoryChange}
                                    placeholder="Choose a category..."
                                    style="bg-white/80 h-11 rounded-xl border-white/40"
                                />
                            </div>
                        {/if}

                        {#if formData.targetType === 'SUBCATEGORY'}
                            <div in:fly={{ y: 5, duration: 200 }} class="space-y-2">
                                <Label class="text-xs font-semibold text-slate-500 uppercase">Select Sub-Category</Label>
                                <Select 
                                    value={formData.subCategoryId ? String(formData.subCategoryId) : ''}
                                    options={subCategoryOptions}
                                    on:change={onSubCategoryChange}
                                    disabled={!formData.categoryId || subCategories.length === 0}
                                    placeholder="Choose a sub-category..."
                                    style="bg-white/80 h-11 rounded-xl border-white/40"
                                />
                            </div>
                        {/if}

                        {#if formData.targetType === 'PRODUCT'}
                            <div in:fly={{ y: 5, duration: 200 }} class="space-y-2">
                                <Label class="text-xs font-semibold text-slate-500 uppercase">Find Product</Label>
                                <div class="relative group">
                                    <Search class="absolute left-3 top-3.5 h-4 w-4 text-slate-400 group-focus-within:text-indigo-500 transition-colors" />
                                    <Input 
                                        class="pl-10 h-11 rounded-xl border-white/40 bg-white/60 focus:border-indigo-500 focus:ring-indigo-500/20" 
                                        placeholder="Search by name or SKU..." 
                                        bind:value={productSearch}
                                        oninput={(e) => searchProducts(e.currentTarget.value)}
                                    />
                                </div>
                                
                                {#if products.length > 0}
                                    <div class="mt-2 text-xs text-slate-400 font-medium px-1">Found {products.length} matches</div>
                                    <div class="mt-2 border border-slate-100/50 rounded-xl max-h-[160px] overflow-y-auto bg-white/60 backdrop-blur-md shadow-lg p-1 custom-scrollbar">
                                        {#each products as textProd}
                                            <button 
                                                class="w-full text-left px-3 py-2.5 hover:bg-indigo-50/50 rounded-lg text-sm flex justify-between items-center transition-all group {formData.productId === textProd.ID ? 'bg-indigo-50 ring-1 ring-indigo-200' : ''}"
                                                onclick={() => formData.productId = textProd.ID}
                                            >
                                                <span class="font-medium text-slate-700 group-hover:text-indigo-700 transition-colors">{textProd.Name}</span>
                                                <Badge variant="secondary" class="bg-slate-100 text-slate-500 text-[10px] tracking-wider font-mono group-hover:bg-white">{textProd.SKU}</Badge>
                                            </button>
                                        {/each}
                                    </div>
                                {/if}
                                {#if formData.productId}
                                    <div class="mt-3 flex items-center justify-between p-3 rounded-xl bg-emerald-50/50 border border-emerald-100 text-emerald-700 animate-in fade-in slide-in-from-top-1">
                                        <div class="flex items-center gap-2 font-medium text-sm">
                                            <CheckCircle2 class="h-4 w-4 text-emerald-600" />
                                            Product Selected
                                        </div>
                                        <span class="text-xs font-mono bg-white/50 px-2 py-0.5 rounded-md border border-emerald-200/50">ID: {formData.productId}</span>
                                    </div>
                                {/if}
                            </div>
                        {/if}
                    </div>
				</div>

				<div class={cn(
                    "flex items-center space-x-4 p-4 rounded-2xl border transition-all duration-300 shadow-sm",
                    formData.isActive 
                        ? "bg-emerald-50/60 border-emerald-200/60 shadow-emerald-500/10" 
                        : "bg-slate-50/60 border-slate-200/60"
                )}>
					<Switch bind:checked={formData.isActive} id="active-mode" />
					<div class="flex flex-col">
						<Label for="active-mode" class={cn(
                            "cursor-pointer text-sm font-bold transition-colors",
                            formData.isActive ? "text-emerald-800" : "text-slate-600"
                        )}>
                            {formData.isActive ? 'Promotion Active' : 'Promotion Inactive'}
                        </Label>
						<span class={cn(
                            "text-xs transition-colors",
                            formData.isActive ? "text-emerald-600/80" : "text-slate-400"
                        )}>
                            {formData.isActive ? 'Live immediately upon saving' : 'Saved as draft, not visible to customers'}
                        </span>
					</div>
				</div>
			</div>

            <!-- Glassy Footer -->
			<div class="p-6 bg-slate-50/30 border-t border-white/20 backdrop-blur-md rounded-b-[2rem]">
                <DialogFooter class="gap-3 sm:gap-0">
                    <Button variant="outline" onclick={() => isDialogOpen = false} class="rounded-xl border-white/40 bg-white/40 hover:bg-white/60 hover:text-slate-900 font-medium transition-all">Cancel</Button>
                    <Button onclick={handleSubmit} class="rounded-xl bg-gradient-to-r from-indigo-600 to-purple-600 hover:from-indigo-500 hover:to-purple-500 text-white shadow-lg shadow-indigo-500/20 font-semibold px-8 transition-all hover:scale-[1.02]">
                        {editingId ? 'Save Changes' : 'Create Promotion'}
                    </Button>
                </DialogFooter>
            </div>
		</DialogContent>
	</Dialog>
</div>

