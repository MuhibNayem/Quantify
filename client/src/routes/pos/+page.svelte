<script lang="ts">
    import { t } from '$lib/i18n';
	import { onMount } from 'svelte';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import {
		Table,
		TableBody,
		TableCell,
		TableHead,
		TableHeader,
		TableRow
	} from '$lib/components/ui/table';
	import {
		Card,
		CardContent,
		CardDescription,
		CardFooter,
		CardHeader,
		CardTitle
	} from '$lib/components/ui/card';
	import {
		Search,
		UserPlus,
		X,
		Zap,
		CreditCard,
		Loader2,
		Users,
		Mail,
		Phone,
		Banknote,
		QrCode,
		Wallet,
		Check
	} from 'lucide-svelte';
	import api from '$lib/api';
	import { crmApi, promotionsApi } from '$lib/api/resources';
	import { toast } from 'svelte-sonner';
	import * as Dialog from '$lib/components/ui/dialog';
	import { Label } from '$lib/components/ui/label';
	import { auth } from '$lib/stores/auth';
	import { settings } from '$lib/stores/settings';
	import { goto } from '$app/navigation';
	import { formatCurrency } from '$lib/utils';
    import { Badge } from '$lib/components/ui/badge';

	$effect(() => {
		if (!auth.hasPermission('pos.view')) {
			toast.error($t('pos.toasts.access_denied'), { description: $t('pos.toasts.access_denied_desc') });
			goto('/');
		}
	});

	// Runes state
	let products = $state<any[]>([]);
    let promotions = $state<any[]>([]);
	let cart = $state<any[]>([]);
	let searchTerm = $state('');
    let fetchLimit = $state(100); // Default limit
    let fetchStatus = $state('IN_STOCK'); // Default: In Stock
	let customerSearchTerm = $state('');
	let selectedCustomer = $state<any | null>(null);
	let paymentMethod = $state<string | null>(null);
	let isProcessing = $state(false);

	// New Customer Modal State
	let isNewCustomerModalOpen = $state(false);
	let newCustomerForm = $state({
		name: '',
		email: '',
		phone: ''
	});

	let pointsToRedeem = $state(0);

    const fetchPromotions = async () => {
        try {
            promotions = await promotionsApi.list(true);
        } catch (error) {
            console.error('Error fetching promotions:', error);
        }
    };

	const fetchProducts = async (search = '') => {
		try {
			// Optimized: Fetch products with server-side search and limit
			const response = await api.get('/sales/products', {
				params: {
					q: search,
					limit: fetchLimit,
                    status: fetchStatus
				}
			});
			const productsData = response.data.products;

			// Map to expected structure
			products = productsData.map((p: any) => ({
				...p,
				stock: { currentQuantity: p.StockQuantity } // Adapter for existing template usage
			}));
		} catch (error) {
            console.error('Error fetching products:', error);
		}
	};

	onMount(() => {
        Promise.all([fetchProducts(), fetchPromotions()]);

		// Parallax hero motion
		const hero = document.querySelector('.parallax-hero') as HTMLElement | null;
		let ticking = false;

		const handleScroll = () => {
			if (!hero) return;
			if (!ticking) {
				window.requestAnimationFrame(() => {
					const scrollY = window.scrollY || 0;
					const translate = Math.min(scrollY * 0.25, 60);
					const blur = Math.min(scrollY * 0.02, 6);
					hero.style.transform = `translateY(${translate}px)`;
					hero.style.filter = `blur(${blur}px)`;
					ticking = false;
				});
				ticking = true;
			}
		};

		window.addEventListener('scroll', handleScroll, { passive: true });

		return () => {
			window.removeEventListener('scroll', handleScroll);
		};
	});

	const handleSearch = async () => {
		await fetchProducts(searchTerm.trim());
	};

    function getProductPrice(product: any): { price: number, originalPrice: number, discountName?: string } {
        const originalPrice = product.SellingPrice;
        let bestPrice = originalPrice;
        let bestPromo = null;
        let bestScore = -1;

        for (const p of promotions) {
            let score = -1;
            
            if (p.ProductID === product.ID) {
                score = 2;
            } else if (p.SubCategoryID && p.SubCategoryID === product.SubCategoryID) {
                score = 1;
            } else if (p.CategoryID && p.CategoryID === product.CategoryID) {
                score = 0;
            }

            if (score > -1) {
                if (score > bestScore) {
                    bestScore = score;
                    bestPromo = p;
                } else if (score === bestScore) {
                    if (bestPromo && p.Priority > bestPromo.Priority) {
                        bestPromo = p;
                    }
                }
            }
        }

        if (bestPromo) {
            if (bestPromo.DiscountType === 'PERCENTAGE') {
                bestPrice = originalPrice * (1 - bestPromo.DiscountValue / 100);
            } else {
                bestPrice = originalPrice - bestPromo.DiscountValue;
            }
            if (bestPrice < 0) bestPrice = 0;
            return { price: bestPrice, originalPrice, discountName: bestPromo.Name };
        }

        return { price: originalPrice, originalPrice };
    }

	const handleCustomerSearch = async () => {
		if (!customerSearchTerm) {
			selectedCustomer = null;
			return;
		}
		try {
			// Search using the list endpoint which supports name/email/phone/id via 'q'
			const response = await api.get('/crm/customers', {
				params: {
					q: customerSearchTerm,
					limit: 1
				}
			});

			if (response.data.users && response.data.users.length > 0) {
				selectedCustomer = response.data.users[0];
			} else {
				// If no results from search, try strict ID fetch just in case user typed exact numeric ID (though search usually covers this)
				try {
					const idResp = await api.get(`/crm/customers/${customerSearchTerm}`);
					selectedCustomer = idResp.data;
				} catch (e) {
					selectedCustomer = null;
					toast.error($t('pos.toasts.customer_not_found'));
				}
			}
		} catch (error) {
			console.error('Error fetching customer:', error);
			selectedCustomer = null;
			toast.error($t('pos.toasts.search_error'));
		}
	};

	const getAvailableStock = (product: any) => {
		const cartItem = cart.find((item) => item.id === product.ID);
		const cartQty = cartItem ? cartItem.quantity : 0;
		const currentStock = product.stock ? product.stock.currentQuantity : 0;
		return Math.max(0, currentStock - cartQty);
	};

	const addToCart = (product: any) => {
		const existingItem = cart.find((item) => item.id === product.ID);
        const currentQty = existingItem ? existingItem.quantity : 0;
        const stockLimit = product.stock ? product.stock.currentQuantity : 0;

        if (currentQty >= stockLimit) {
            toast.error($t('pos.toasts.out_of_stock'));
            return;
        }

		if (existingItem) {
			existingItem.quantity++;
			cart = [...cart];
		} else {
			cart = [
				...cart,
				{
					...product,
					id: product.ID,
					quantity: 1
				}
			];
		}
	};

	const removeFromCart = (productId: number) => {
		cart = cart.filter((item) => item.id !== productId);
	};

	const updateQuantity = (productId: number, quantity: number) => {
		if (!quantity || quantity < 1) quantity = 1;

		const item = cart.find((item) => item.id === productId);
		if (item) {
            const stockLimit = item.stock ? item.stock.currentQuantity : 0;
            
            if (quantity > stockLimit) {
                toast.error($t('pos.toasts.stock_limit_reached', { stock: stockLimit }));
                item.quantity = stockLimit;
            } else {
			    item.quantity = quantity;
            }
			cart = [...cart];
		}
	};

	const clearCart = () => {
		cart = [];
	};

	const subtotal = $derived(
		cart.reduce((acc, item) => {
            const { price } = getProductPrice(item);
            return acc + price * item.quantity;
        }, 0)
	);
	// Tax Calculation
	const taxRate = $derived(($settings.tax_rate_percentage || 0) / 100);
	const tax = $derived(subtotal * taxRate);
	const total = $derived(subtotal + tax);

	const potentialPoints = $derived(
		Math.floor(subtotal * ($settings.loyalty_points_earning_rate || 0))
	);

	const canCompleteOrder = $derived(cart.length > 0 && !!paymentMethod);

	const setPayment = (method: string) => {
		paymentMethod = method;
	};

	const completeOrder = async () => {
		if (!canCompleteOrder || isProcessing) return;
		isProcessing = true;

		const toastId = toast.loading($t('pos.toasts.processing'));

		try {
			const payload = {
				items: cart.map((item) => ({
					productId: item.id,
					quantity: item.quantity
				})),
				customerId: selectedCustomer ? selectedCustomer.ID : null,
				paymentMethod: paymentMethod,
				pointsToRedeem: Math.min(
					pointsToRedeem,
					selectedCustomer ? selectedCustomer.loyalty?.Points || 0 : 0
				)
			};

			await api.post('/sales/checkout', payload);

			toast.success($t('pos.toasts.order_success'), { id: toastId });
			if (selectedCustomer && potentialPoints > 0) {
				toast.info($t('pos.toasts.loyalty_earned', { points: potentialPoints }));
			}
			// Reset all state
			cart = [];
			paymentMethod = null;
			selectedCustomer = null;
			customerSearchTerm = '';
			pointsToRedeem = 0;
			
			await fetchProducts(); // Refresh stock
		} catch (error: any) {
			console.error('Error completing order:', error);
			const msg = error.response?.data?.error || 'Failed to complete order';
			toast.error(`${$t('pos.toasts.transaction_fail')}: ${msg}`, { id: toastId });
		} finally {
			isProcessing = false;
		}
	};

	const saveNewCustomer = async () => {
		const nameParts = newCustomerForm.name.trim().split(' ');
		const firstName = nameParts[0] || '';
		const lastName = nameParts.slice(1).join(' ') || '';

		if (!firstName) {
			toast.error($t('pos.toasts.name_required'));
			return;
		}

		try {
			const newUser = await crmApi.createCustomer({
				username: newCustomerForm.email || `user${Date.now()}`,
				email: newCustomerForm.email,
				firstName: firstName,
				lastName: lastName,
				phoneNumber: newCustomerForm.phone,
				role: 'Customer',
				password: 'TempPassword123!'
			} as any);

			selectedCustomer = newUser;
			isNewCustomerModalOpen = false;
			toast.success($t('pos.toasts.customer_created'));
			newCustomerForm = { name: '', email: '', phone: '' };
		} catch (error) {
			console.error('Error creating customer:', error);
			toast.error($t('pos.toasts.create_fail'));
		}
	};
</script>

<!-- Page background gradient + hero -->
<section
	class="relative isolate min-h-screen w-full overflow-hidden bg-gradient-to-br from-indigo-50 via-sky-50 to-slate-50"
>
	<!-- Animated gradient wash -->
	<div
		class="animate-gradientShift absolute inset-0 -z-20 bg-gradient-to-r from-indigo-100 via-sky-100 to-blue-100 bg-[length:220%_220%] opacity-70"
	/>

	<!-- Floating glow blobs -->
	<div
		class="animate-pulseGlow pointer-events-none absolute -left-24 -top-32 h-80 w-80 rounded-full bg-indigo-200/40 blur-3xl"
	/>
	<div
		class="animate-pulseGlow pointer-events-none absolute -bottom-36 -right-24 h-72 w-72 rounded-full bg-sky-200/40 blur-3xl delay-700"
	/>

	<!-- Hero -->
	<div
		class="parallax-hero relative mx-auto max-w-7xl px-4 pb-6 pt-14 text-center sm:px-6 sm:pb-10 sm:pt-20 sm:text-left lg:px-8"
	>
		<div
			class="animate-cardFloat mb-3 inline-flex items-center justify-center gap-3 sm:justify-start"
		>
			<span
				class="inline-flex rounded-2xl bg-gradient-to-br from-indigo-500 to-sky-500 p-2.5 shadow-md"
			>
				<CreditCard class="h-5 w-5 text-white" />
			</span>
			<div class="flex flex-col items-start">
				<p
					class="text-[0.65rem] font-semibold uppercase tracking-[0.22em] text-indigo-700 sm:text-xs"
				>
					{$t('pos.hero.label')}
				</p>
				<p class="text-[0.65rem] text-slate-500 sm:text-[0.7rem]">
					{$t('pos.hero.sub_label')}
				</p>
			</div>
		</div>

		<h1
			class="mb-3 bg-gradient-to-r from-slate-900 via-indigo-700 to-sky-700 bg-clip-text text-3xl font-bold text-transparent sm:text-4xl lg:text-5xl"
		>
			{$t('pos.hero.title')}
		</h1>
		<p class="mx-auto max-w-2xl text-sm text-slate-600 sm:mx-0 sm:text-base">
			{$t('pos.hero.subtitle')}
		</p>

		<div class="mt-6 flex flex-col justify-center gap-3 sm:flex-row sm:justify-start">
			<Button
				variant="secondary"
				onclick={clearCart}
				class="w-full rounded-xl bg-gradient-to-r from-indigo-500 to-sky-500 px-5 py-2.5 font-medium text-white shadow-lg transition-all duration-300 hover:scale-105 hover:from-indigo-600 hover:to-sky-600 hover:shadow-xl focus:ring-2 focus:ring-indigo-300 sm:w-auto"
			>
				<Zap class="mr-2 h-4 w-4" />
				{$t('pos.hero.new_sale_btn')}
			</Button>
			<Button
				variant="outline"
				onclick={() => fetchProducts(searchTerm.trim())}
				class="w-full rounded-xl border border-indigo-100 bg-white/80 px-5 py-2.5 font-medium text-indigo-700 shadow-md transition-all duration-300 hover:scale-105 hover:bg-indigo-50 hover:shadow-lg focus:ring-2 focus:ring-indigo-200 sm:w-auto"
			>
				<Search class="mr-2 h-4 w-4" />
				{$t('pos.hero.refresh_catalog_btn')}
			</Button>
		</div>
	</div>

	<!-- POS header bar -->

	<!-- Main content -->
	<div class="flex h-full min-h-0 flex-col px-4 py-4">
		<Card
			class="overflow-hidden rounded-2xl border-0 bg-white/80 shadow-lg backdrop-blur transition-all duration-300 hover:scale-[1.01] hover:shadow-xl"
		>
			<CardHeader class="flex flex-row items-center justify-between gap-3 pb-3">
				<div>
					<CardTitle
						class="flex items-center gap-2 text-lg font-semibold tracking-tight text-slate-900"
					>
						<span
							class="inline-flex h-8 w-8 items-center justify-center rounded-xl bg-indigo-100 text-indigo-600"
						>
							<CreditCard class="h-4 w-4" />
						</span>
						{$t('pos.header.title')}
					</CardTitle>
					<CardDescription class="text-[0.75rem] text-slate-500">
						{$t('pos.header.description')}
					</CardDescription>
				</div>
				<div class="hidden items-center gap-2 text-[0.7rem] text-slate-500 sm:flex">
					<span class="rounded-full border border-slate-200 bg-slate-50 px-2 py-0.5"
						>{$t('pos.header.super_shop_mode')}</span
					>
				</div>
			</CardHeader>
			<CardContent class="pb-3 pt-0">
				<div class="flex flex-col items-stretch gap-3 sm:flex-row">
					<div class="relative flex-1">
						<Search class="absolute left-2.5 top-1/2 h-4 w-4 -translate-y-1/2 text-slate-400" />
						<Input
							bind:value={searchTerm}
							placeholder={$t('pos.header.search_placeholder')}
							class="rounded-xl border-slate-200 bg-slate-50/80 pl-8 text-sm focus-visible:ring-indigo-300"
							onkeydown={(e) => e.key === 'Enter' && handleSearch()}
						/>
					</div>

					<Button
						class="rounded-xl bg-gradient-to-r from-indigo-500 to-sky-500 text-white shadow-md hover:from-indigo-600 hover:to-sky-600"
						onclick={handleSearch}
					>
						<Search class="mr-2 h-4 w-4" />
						{$t('pos.header.search_btn')}
					</Button>
				</div>
			</CardContent>
		</Card>
		<div class="mx-auto grid w-full max-w-7xl grid-cols-1 gap-4 px-4 py-4 lg:grid-cols-2">
			<!-- LEFT STACK: Products + Cart -->
			<div class="space-y-5" data-animate="fade-up" style="animation-delay:120ms">
				<!-- Product grid -->
				<Card
					class="flex-1 overflow-hidden rounded-2xl border-0 bg-gradient-to-br from-indigo-50 to-sky-100 shadow-lg transition-all duration-300 hover:scale-[1.01] hover:shadow-xl"
				>
					<CardHeader class="border-b border-white/60 bg-white/70 pb-3 backdrop-blur">
						<div class="flex flex-col gap-3">
							<div>
								<CardTitle class="text-sm text-slate-800">{$t('pos.products.title')}</CardTitle>
								<CardDescription class="text-[0.75rem] text-slate-500">
									{$t('pos.products.description')}
								</CardDescription>
							</div>
                            <div class="flex flex-wrap items-center gap-2">
                                <!-- Status Filter -->
                                <select
                                    bind:value={fetchStatus}
                                    onchange={() => fetchProducts(searchTerm.trim())}
                                    class="h-9 rounded-lg border-slate-200 bg-white/80 text-xs font-medium text-slate-600 shadow-sm focus:border-indigo-300 focus:ring-indigo-300 outline-none pl-2 pr-2 cursor-pointer hover:bg-white flex-1 sm:flex-none"
                                    aria-label={$t('pos.products.filter_status.label')}
                                >
                                    <option value="">{$t('pos.products.filter_status.all')}</option>
                                    <option value="IN_STOCK">{$t('pos.products.filter_status.in_stock')}</option>
                                    <option value="LOW_STOCK">{$t('pos.products.filter_status.low_stock')}</option>
                                    <option value="OUT_OF_STOCK">{$t('pos.products.filter_status.out_of_stock')}</option>
                                </select>

                                <div class="flex items-center gap-2 flex-1 sm:flex-none">
                                    <select
                                        bind:value={fetchLimit}
                                        onchange={() => fetchProducts(searchTerm.trim())}
                                        class="h-9 w-full sm:w-auto rounded-lg border-slate-200 bg-white/80 text-xs font-medium text-slate-600 shadow-sm focus:border-indigo-300 focus:ring-indigo-300 outline-none pl-2 pr-2 cursor-pointer hover:bg-white"
                                    >
                                        <option value={100}>100 items</option>
                                        <option value={200}>200 items</option>
                                        <option value={500}>500 items</option>
                                        <option value={1000}>1,000 items</option>
                                    </select>
                                    <div class="text-[0.7rem] text-slate-500 whitespace-nowrap">
                                        <span class="font-medium text-slate-700">{products.length}</span> results
                                    </div>
                                </div>
                            </div>
						</div>
					</CardHeader>
					<CardContent class="max-h-[22rem] overflow-y-auto p-3 pt-3">
						{#if products.length === 0}
							<div class="flex h-full items-center justify-center text-[0.8rem] text-slate-500">
								{$t('pos.products.no_results')}
							</div>
						{:else}
							<div class="grid grid-cols-2 gap-3 xl:grid-cols-3">
								{#each products as product}
                                    {@const { price, originalPrice, discountName } = getProductPrice(product)}
									<button
										type="button"
										onclick={() => addToCart(product)}
										class="group flex flex-col justify-between rounded-xl border border-indigo-50 bg-white/80 p-3 text-left shadow-sm transition-all duration-200 hover:border-indigo-200 hover:bg-indigo-50/80 hover:shadow-md"
									>
										<div class="space-y-1.5">
											<div class="line-clamp-2 text-[0.8rem] font-medium text-slate-900">
												{product.Name}
											</div>
											<div class="mt-0.5 flex flex-col gap-1">
                                                <div class="flex flex-wrap items-center gap-2">
                                                    {#if price < originalPrice}
                                                        <span class="text-[0.65rem] text-slate-400 line-through">
                                                            {formatCurrency(originalPrice)}
                                                        </span>
                                                        <span class="text-[0.8rem] font-bold text-rose-600">
                                                            {formatCurrency(price)}
                                                        </span>
                                                    {:else}
                                                        <span class="text-[0.8rem] font-semibold text-indigo-600">
                                                            {formatCurrency(price)}
                                                        </span>
                                                    {/if}
                                                </div>
                                                
                                                {#if discountName}
                                                     <Badge variant="outline" class="w-fit text-[0.6rem] px-1 py-0 border-rose-200 bg-rose-50 text-rose-600">
                                                        {discountName}
                                                     </Badge>
                                                {/if}

												<div class="flex items-center gap-1 text-[0.7rem] text-slate-500">
													<span
														class="h-1.5 w-1.5 rounded-full {getAvailableStock(product) > 0
															? 'bg-emerald-500'
															: 'bg-rose-400'}"
													/>
													<span>
														{$t('pos.products.in_stock', { count: getAvailableStock(product) })}
													</span>
												</div>
											</div>
										</div>
										<div
											class="mt-3 flex items-center justify-between text-[0.7rem] text-slate-400"
										>
											<span class="flex items-center gap-1 font-medium group-hover:text-indigo-600">
												<span
													class="rounded-full border border-indigo-100 bg-indigo-50/90 px-1.5 py-0.5 text-[0.65rem] text-indigo-600"
												>
													Tap
												</span>
												{$t('pos.products.tap_to_add')}
											</span>
											<span class="text-[0.65rem] uppercase tracking-wide text-slate-400"
												>#{product.ID}</span
											>
										</div>
									</button>
								{/each}
							</div>
						{/if}
					</CardContent>
				</Card>

				<!-- Cart -->
				<Card
					class="overflow-hidden rounded-2xl border-0 bg-white/90 shadow-lg backdrop-blur-xl transition-all duration-300 hover:scale-[1.01] hover:shadow-xl"
					data-animate="fade-up"
					style="animation-delay:200ms"
				>
					<CardHeader
						class="flex flex-row items-center justify-between border-b border-slate-100/80 bg-slate-50/70 pb-3"
					>
						<div>
							<CardTitle class="text-sm text-slate-900">{$t('pos.cart.title')}</CardTitle>
							<CardDescription class="text-[0.75rem] text-slate-500">
								{cart.length === 0
									? $t('pos.cart.empty_desc')
									: $t('pos.cart.items_desc', { count: cart.length, s: cart.length > 1 ? 's' : '' })}
							</CardDescription>
						</div>

						{#if cart.length > 0}
							<Button
								variant="ghost"
								size="sm"
								class="rounded-xl px-3 py-1.5 text-[0.75rem] text-slate-500 hover:bg-rose-50 hover:text-rose-500"
								onclick={clearCart}
							>
								{$t('pos.cart.clear_btn')}
							</Button>
						{/if}
					</CardHeader>

					<CardContent class="max-h-[20rem] overflow-auto pt-0">
						{#if cart.length === 0}
							<div class="py-6 text-center text-[0.8rem] text-slate-400">
								{$t('pos.cart.empty_state')}
							</div>
						{:else}
							<!-- ⭐ FIX: horizontal scroll enabled, no fixed width constraint -->
							<Table class="w-full min-w-[30rem]">
								<TableHeader>
									<TableRow class="border-slate-100 bg-slate-50/80">
										<TableHead />
										<TableHead class="w-1/3 min-w-[8rem] text-[0.7rem] text-slate-500">{$t('pos.cart.table.product')}</TableHead>
										<TableHead class="text-[0.7rem] text-slate-500">{$t('pos.cart.table.price')}</TableHead>
										<TableHead class="text-[0.7rem] min-w-[5rem] text-slate-500">{$t('pos.cart.table.qty')}</TableHead>
										<TableHead class="text-right text-[0.7rem] text-slate-500">{$t('pos.cart.table.total')}</TableHead>
									</TableRow>
								</TableHeader>

								<TableBody>
									{#each cart as item}
                                        {@const { price, originalPrice, discountName } = getProductPrice(item)}
										<TableRow class="border-slate-100 hover:bg-slate-50/60">
											<TableCell class="text-right align-top">
												<Button
													variant="ghost"
													size="icon"
													class="h-7 w-7 rounded-full text-slate-400 hover:bg-rose-50 hover:text-rose-500"
													onclick={() => removeFromCart(item.id)}
												>
													<X class="h-3 w-3" />
												</Button>
											</TableCell>
											<TableCell class="min-w-0 whitespace-normal align-top">
												<div
													class="
			line-clamp-2 hyphens-auto break-words
			text-[0.8rem] font-medium
			text-slate-900
		"
													style="
			display: -webkit-box;
			-webkit-line-clamp: 8;
			-webkit-box-orient: vertical;
			overflow: hidden;
			word-break: break-word;
		"
												>
													{item.Name}
												</div>

												<div class="mt-0.5 break-words text-[0.7rem] text-slate-400">
													#{item.id}
                                                    {#if discountName}
                                                        <span class="ml-1 text-rose-500 font-medium text-[0.65rem]">({discountName})</span>
                                                    {/if}
												</div>
											</TableCell>

											<TableCell class="min-w-[8rem] align-top text-[0.8rem] text-slate-800 whitespace-nowrap">
                                                {#if price < originalPrice}
                                                    <div class="flex flex-col">
    												    <span class="line-through text-slate-400 text-[0.65rem]">{formatCurrency(originalPrice)}</span>
                                                        <span class="font-bold text-rose-600">{formatCurrency(price)}</span>
                                                    </div>
                                                {:else}
												    {formatCurrency(price)}
                                                {/if}
											</TableCell>

											<!-- Quantity input -->
											<TableCell class="min-w-[5rem] align-top">
												<Input
													type="number"
													class="min-w-[5rem] h-8 w-full rounded-lg border-slate-200 bg-slate-50/80 px-2 text-[0.8rem]"
													min="1"
													value={item.quantity}
													onchange={(e) => updateQuantity(item.id, parseInt(e.currentTarget.value))}
												/>
											</TableCell>

											<TableCell
												class="min-w-[12rem] text-right align-top text-[0.8rem] font-semibold text-slate-900 whitespace-nowrap"
											>
												{formatCurrency(price * item.quantity)}
											</TableCell>

											
										</TableRow>
									{/each}
								</TableBody>
							</Table>
						{/if}
					</CardContent>
				</Card>
			</div>

			<!-- RIGHT STACK: Customer + Payment + Summary -->
			<div class="space-y-5">
				<!-- Customer card -->
				<Card
					class="overflow-hidden rounded-2xl border-0 bg-white/90 shadow-lg backdrop-blur-xl transition-all duration-300 hover:scale-[1.01] hover:shadow-xl"
					data-animate="fade-up"
					style="animation-delay:160ms"
				>
					<CardHeader class="border-b border-slate-100/80 bg-white/80 pb-3">
						<CardTitle class="text-sm text-slate-900">{$t('pos.customer.title')}</CardTitle>
						<CardDescription class="text-[0.75rem] text-slate-500">
							{$t('pos.customer.description')}
						</CardDescription>
					</CardHeader>
					<CardContent class="space-y-3 pt-3">
						<div class="flex items-center gap-2">
							<div class="relative flex-1">
								<Search class="absolute left-2.5 top-1/2 h-4 w-4 -translate-y-1/2 text-slate-400" />
								<Input
									bind:value={customerSearchTerm}
									placeholder={$t('pos.customer.search_placeholder')}
									class="rounded-xl border-slate-200 bg-slate-50/80 pl-8 text-sm focus-visible:ring-emerald-300"
									onkeydown={(e) => e.key === 'Enter' && handleCustomerSearch()}
								/>
							</div>
							<Button
								variant="outline"
								size="icon"
								class="rounded-xl border-slate-200 bg-white/80 hover:bg-slate-50"
								onclick={handleCustomerSearch}
							>
								<Search class="h-4 w-4" />
							</Button>
							<Button
								variant="secondary"
								class="flex items-center gap-1 rounded-xl border border-indigo-100 bg-indigo-50 px-3 py-2 text-[0.75rem] text-indigo-700 hover:bg-indigo-100"
								onclick={() => (isNewCustomerModalOpen = true)}
							>
								<UserPlus class="h-4 w-4" />
								<span class="font-medium">{$t('pos.customer.new_btn')}</span>
							</Button>
						</div>

						{#if selectedCustomer}
							<div
								class="flex flex-col gap-0.5 rounded-xl border border-emerald-100 bg-emerald-50/70 px-3 py-2.5 text-[0.75rem] text-slate-700"
							>
								<div class="flex items-center justify-between gap-2 font-medium text-slate-900">
									<span>
										{selectedCustomer.FirstName}
										{selectedCustomer.LastName}
									</span>
									<span
										class="rounded-full border border-emerald-100 bg-white/80 px-1.5 py-0.5 text-[0.65rem] text-emerald-600"
									>
										{selectedCustomer.Username}
									</span>
								</div>
								{#if selectedCustomer.Email}
									<div>{selectedCustomer.Email}</div>
								{/if}
								{#if selectedCustomer.PhoneNumber}
									<div>{selectedCustomer.PhoneNumber}</div>
								{/if}
								{#if selectedCustomer.loyalty}
									<div class="mt-1 flex items-center gap-2 border-t border-emerald-100/50 pt-1">
										<span class="font-semibold text-emerald-700">
											{$t('pos.customer.loyalty_pts', { points: selectedCustomer.loyalty.Points })}
										</span>
										<span
											class="rounded-sm bg-emerald-100 px-1 text-[0.65rem] uppercase text-emerald-800"
										>
											{$t('pos.customer.tier', { tier: selectedCustomer.loyalty.Tier })}
										</span>
									</div>
								{/if}
							</div>
						{:else}
							<div class="text-[0.75rem] text-slate-400">
								{$t('pos.customer.no_selected')}
							</div>
						{/if}
					</CardContent>
				</Card>

				<!-- Payment methods -->
				<Card
					class="overflow-hidden rounded-2xl border-0 bg-white/90 shadow-lg backdrop-blur-xl transition-all duration-300 hover:scale-[1.01] hover:shadow-xl"
					data-animate="fade-up"
					style="animation-delay:200ms"
				>
					<CardHeader class="border-b border-slate-100/80 bg-white/80 pb-3">
						<CardTitle class="text-sm text-slate-900">{$t('pos.payment.title')}</CardTitle>
						<CardDescription class="text-[0.75rem] text-slate-500">
							{$t('pos.payment.description')}
						</CardDescription>
					</CardHeader>
					<CardContent class="pt-3">
						<div class="grid grid-cols-2 gap-3 sm:grid-cols-4">
							{#each [{ id: 'CASH', label: $t('pos.payment.methods.cash'), icon: Banknote, color: 'text-emerald-500', sub: $t('pos.payment.sub.physical') }, { id: 'CARD', label: $t('pos.payment.methods.card'), icon: CreditCard, color: 'text-violet-500', sub: $t('pos.payment.sub.terminal') }, { id: 'BKASH', label: $t('pos.payment.methods.bkash'), icon: QrCode, color: 'text-pink-500', sub: $t('pos.payment.sub.mobile') }, { id: 'OTHER', label: $t('pos.payment.methods.other'), icon: Wallet, color: 'text-orange-500', sub: $t('pos.payment.sub.check_due') }] as method}
								<button
									type="button"
									onclick={() => setPayment(method.id)}
									class={`group relative flex flex-col items-center justify-center gap-2 rounded-2xl border p-3 transition-all duration-200 hover:shadow-md ${
										paymentMethod === method.id
											? 'scale-[1.02] border-transparent bg-gradient-to-br from-indigo-600 to-violet-600 shadow-lg shadow-indigo-500/25 ring-2 ring-indigo-500/10'
											: 'border-slate-100 bg-white/50 text-slate-600 hover:border-indigo-100 hover:bg-white'
									}`}
								>
									{#if paymentMethod === method.id}
										<div class="absolute right-2 top-2 rounded-full bg-white/20 p-0.5">
											<Check class="h-3 w-3 text-white" />
										</div>
									{/if}

									<div
										class={`rounded-xl p-2.5 transition-colors ${
											paymentMethod === method.id
												? 'bg-white/10 text-white'
												: 'bg-slate-50 ' + method.color + ' group-hover:bg-indigo-50/50'
										}`}
									>
										<svelte:component this={method.icon} class="h-5 w-5" />
									</div>

									<div class="text-center">
										<span
											class={`block text-xs font-semibold ${
												paymentMethod === method.id ? 'text-white' : 'text-slate-700'
											}`}
										>
											{method.label}
										</span>
										<span
											class={`block text-[0.65rem] ${
												paymentMethod === method.id ? 'text-indigo-100' : 'text-slate-400'
											}`}
										>
											{method.sub}
										</span>
									</div>
								</button>
							{/each}
						</div>

						<!-- Loyalty Redemption -->
						{#if selectedCustomer && selectedCustomer.loyalty && selectedCustomer.loyalty.Points > 0}
							<div class="mt-4 border-t border-slate-100/80 pt-4">
								<div class="flex items-center justify-between">
									<Label class="text-[0.7rem] font-medium text-slate-500"
										>{$t('pos.loyalty.redeem_label')}</Label
									>
									<span class="text-[0.7rem] font-medium text-emerald-600">
										{$t('pos.loyalty.available', {
											points: selectedCustomer.loyalty.Points,
											value: formatCurrency(selectedCustomer.loyalty.Points * ($settings.loyalty_points_redemption_rate || 0.01))
										})}
									</span>
								</div>
								<div class="mt-2 flex items-center gap-3">
									<Input
										type="number"
										min="0"
										max={selectedCustomer.loyalty.Points}
										bind:value={pointsToRedeem}
										class="h-9 w-24 rounded-lg border-slate-200 bg-slate-50 text-right text-sm"
									/>
									<span class="text-xs text-slate-400">{$t('pos.loyalty.points')}</span>

									{#if pointsToRedeem > 0}
										<span class="ml-auto text-sm font-semibold text-emerald-500">
											- {formatCurrency(
												pointsToRedeem * ($settings.loyalty_points_redemption_rate || 0.01)
											)}
										</span>
									{/if}
								</div>
								{#if pointsToRedeem > selectedCustomer.loyalty.Points}
									<p class="mt-1 text-[0.65rem] text-rose-500">{$t('pos.loyalty.error_exceed')}</p>
								{/if}
							</div>
						{/if}
					</CardContent>
				</Card>

				<!-- Summary & Complete -->
				<Card
					class="relative overflow-hidden rounded-2xl border-0 bg-slate-900 text-slate-50 shadow-xl"
					data-animate="fade-up"
					style="animation-delay:240ms"
				>
					<div
						class="absolute inset-0 bg-[radial-gradient(circle_at_top,_rgba(129,140,248,0.35),_transparent_55%),_radial-gradient(circle_at_bottom,_rgba(56,189,248,0.25),_transparent_55%)] opacity-80"
					/>
					<div class="relative">
						<CardHeader class="pb-3">
							<CardTitle class="text-[0.9rem] text-slate-50/90">{$t('pos.summary.title')}</CardTitle>
							<CardDescription class="text-[0.7rem] text-slate-300">
								{$t('pos.summary.description')}
							</CardDescription>
						</CardHeader>
						<CardContent class="space-y-2 pt-0 text-[0.8rem]">
							<div class="flex justify-between text-slate-200">
								<span>{$t('pos.summary.subtotal')}</span>
								<span>{formatCurrency(subtotal)}</span>
							</div>
							<div class="flex justify-between text-[0.75rem] text-slate-300/90">
								<span>{$t('pos.summary.tax', { rate: Number((taxRate * 100).toFixed(2)) })}</span>
								<span>{formatCurrency(tax)}</span>
							</div>
							<div class="flex items-center justify-between border-t border-slate-700/70 pt-2">
								<span class="text-[0.85rem] font-medium text-slate-100">{$t('pos.summary.total')}</span>
								<span class="text-lg font-semibold tracking-tight text-slate-50">
									{formatCurrency(total)}
								</span>
							</div>

							<div class="flex items-center justify-between pt-1 text-[0.7rem] text-slate-300/90">
								<div>
									<span class="font-medium">{$t('pos.summary.payment')}</span>
									<span class="ml-1">
										{paymentMethod ? paymentMethod : $t('pos.summary.not_selected')}
									</span>
								</div>
								<div>
									<span class="font-medium">{$t('pos.summary.items')}</span>
									<span class="ml-1">{cart.length}</span>
								</div>
							</div>

							{#if selectedCustomer && potentialPoints > 0}
								<div
									class="flex items-center justify-between border-t border-slate-700/50 pt-2 text-[0.7rem] text-emerald-400"
								>
									<div class="flex items-center gap-1">
										<Zap class="h-3 w-3" />
										<span class="font-medium">{$t('pos.summary.loyalty_earnings')}</span>
									</div>
									<span class="font-bold">+{potentialPoints} pt</span>
								</div>
							{/if}
						</CardContent>
						<CardFooter class="pt-3">
							<Button
								size="lg"
								class="h-11 w-full rounded-xl bg-gradient-to-r from-emerald-400 via-sky-400 to-indigo-500 text-sm font-semibold tracking-wide shadow-lg shadow-indigo-500/40 hover:from-emerald-400 hover:via-sky-400 hover:to-indigo-500 disabled:cursor-not-allowed disabled:opacity-60"
								onclick={completeOrder}
								disabled={!canCompleteOrder || isProcessing}
							>
								{#if isProcessing}
									<Loader2 class="mr-2 h-4 w-4 animate-spin" />
									{$t('pos.summary.processing_btn')}
								{:else if !cart.length}
									{$t('pos.summary.add_items_hint')}
								{:else if !paymentMethod}
									{$t('pos.summary.select_payment_hint')}
								{:else}
									{$t('pos.summary.complete_btn')}
								{/if}
							</Button>
						</CardFooter>
					</div>
				</Card>
			</div>
		</div>
	</div>

	<Dialog.Root bind:open={isNewCustomerModalOpen}>
		<Dialog.Content
			class="overflow-hidden rounded-3xl border-0 bg-white/95 p-0 shadow-2xl backdrop-blur-xl transition-all duration-300 sm:max-w-[425px]"
		>
			<!-- Header with gradient and pattern -->
			<div
				class="relative overflow-hidden bg-gradient-to-br from-violet-600 to-indigo-600 px-6 py-6 sm:px-10 sm:py-10"
			>
				<!-- Background decorative elements -->
				<div class="absolute -right-10 -top-10 h-40 w-40 rounded-full bg-white/10 blur-3xl"></div>
				<div
					class="absolute bottom-0 right-0 h-32 w-32 translate-x-12 translate-y-12 rounded-full bg-indigo-500/30 blur-2xl"
				></div>

				<div class="relative z-10">
					<Dialog.Title class="text-2xl font-bold tracking-tight text-white">
						{$t('pos.new_customer_modal.title')}
					</Dialog.Title>
					<Dialog.Description class="mt-1.5 text-violet-100">
						{$t('pos.new_customer_modal.description')}
					</Dialog.Description>
				</div>
			</div>

			<!-- Body -->
			<div class="space-y-6 px-6 py-8 sm:px-10">
				{#key isNewCustomerModalOpen}
					<!-- Name Field -->
					<div class="group space-y-2">
						<label
							for="name"
							class="ml-1 text-xs font-semibold uppercase tracking-wider text-slate-500 transition-colors group-focus-within:text-violet-600"
						>
							{$t('pos.new_customer_modal.name_label')}
						</label>
						<div class="relative">
							<div class="pointer-events-none absolute inset-y-0 left-0 flex items-center pl-3">
								<Users class="h-4 w-4 text-slate-400 group-focus-within:text-violet-500" />
							</div>
							<Input
								id="name"
								bind:value={newCustomerForm.name}
								placeholder={$t('pos.new_customer_modal.name_placeholder')}
								class="rounded-xl border-slate-200 bg-slate-50 pl-10 transition-all duration-300 focus:border-violet-500 focus:bg-white focus:ring-4 focus:ring-violet-500/10 group-hover:border-violet-300"
							/>
						</div>
					</div>

					<!-- Email Field -->
					<div class="group space-y-2">
						<label
							for="email"
							class="ml-1 text-xs font-semibold uppercase tracking-wider text-slate-500 transition-colors group-focus-within:text-violet-600"
						>
							{$t('pos.new_customer_modal.email_label')}
						</label>
						<div class="relative">
							<div class="pointer-events-none absolute inset-y-0 left-0 flex items-center pl-3">
								<Mail class="h-4 w-4 text-slate-400 group-focus-within:text-violet-500" />
							</div>
							<Input
								id="email"
								type="email"
								bind:value={newCustomerForm.email}
								placeholder={$t('pos.new_customer_modal.email_placeholder')}
								class="rounded-xl border-slate-200 bg-slate-50 pl-10 transition-all duration-300 focus:border-violet-500 focus:bg-white focus:ring-4 focus:ring-violet-500/10 group-hover:border-violet-300"
							/>
						</div>
					</div>

					<!-- Phone Field -->
					<div class="group space-y-2">
						<label
							for="phone"
							class="ml-1 text-xs font-semibold uppercase tracking-wider text-slate-500 transition-colors group-focus-within:text-violet-600"
						>
							{$t('pos.new_customer_modal.phone_label')}
						</label>
						<div class="relative">
							<div class="pointer-events-none absolute inset-y-0 left-0 flex items-center pl-3">
								<Phone class="h-4 w-4 text-slate-400 group-focus-within:text-violet-500" />
							</div>
							<Input
								id="phone"
								bind:value={newCustomerForm.phone}
								placeholder={$t('pos.new_customer_modal.phone_placeholder')}
								class="rounded-xl border-slate-200 bg-slate-50 pl-10 transition-all duration-300 focus:border-violet-500 focus:bg-white focus:ring-4 focus:ring-violet-500/10 group-hover:border-violet-300"
							/>
						</div>
					</div>
				{/key}
			</div>

			<!-- Footer -->
			<div
				class="flex items-center justify-end gap-3 bg-slate-50/50 px-6 py-4 backdrop-blur sm:px-10"
			>
				<Button
					variant="ghost"
					onclick={() => (isNewCustomerModalOpen = false)}
					class="rounded-xl text-slate-600 hover:bg-slate-100 hover:text-slate-900"
				>
					{$t('pos.new_customer_modal.cancel_btn')}
				</Button>
				<Button
					onclick={saveNewCustomer}
					class="rounded-xl bg-gradient-to-r from-violet-600 to-indigo-600 px-8 font-semibold text-white shadow-lg shadow-violet-500/25 transition-all duration-300 hover:scale-[1.02] hover:shadow-xl hover:shadow-violet-500/35 active:scale-[0.98]"
				>
					{$t('pos.new_customer_modal.create_btn')}
				</Button>
			</div>
		</Dialog.Content>
	</Dialog.Root>
</section>

<style lang="postcss">
	/* Smooth transitions globally */
	* {
		transition-property:
			color, background-color, border-color, text-decoration-color, fill, stroke, opacity,
			box-shadow, transform, filter, backdrop-filter;
		transition-timing-function: cubic-bezier(0.4, 0, 0.2, 1);
		transition-duration: 300ms;
	}

	.parallax-hero {
		transform: translateY(0);
		will-change: transform, filter;
		transition:
			transform 0.1s ease-out,
			filter 0.2s ease-out;
	}

	/* Hero gradient animation */
	@keyframes gradientShift {
		0% {
			background-position: 0% 50%;
		}
		50% {
			background-position: 100% 50%;
		}
		100% {
			background-position: 0% 50%;
		}
	}
	.animate-gradientShift {
		background-size: 200% 200%;
		animation: gradientShift 18s ease-in-out infinite;
	}

	/* Soft glowing blobs */
	@keyframes pulseGlow {
		0%,
		100% {
			transform: scale(1);
			opacity: 0.45;
		}
		50% {
			transform: scale(1.08);
			opacity: 0.7;
		}
	}
	.animate-pulseGlow {
		animation: pulseGlow 12s ease-in-out infinite;
	}

	/* Card float micro-motion */
	@keyframes cardFloat {
		0%,
		100% {
			transform: translateY(0);
		}
		50% {
			transform: translateY(-4px);
		}
	}
	.animate-cardFloat {
		animation: cardFloat 4s ease-in-out infinite;
	}

	/* Fade-up reveal */
	@keyframes fadeUp {
		from {
			opacity: 0;
			transform: translateY(12px);
		}
		to {
			opacity: 1;
			transform: translateY(0);
		}
	}
	[data-animate='fade-up'] {
		opacity: 0;
		transform: translateY(12px);
		animation: fadeUp 0.6s ease-out forwards;
	}

	/* Pastel scrollbar */
	::-webkit-scrollbar {
		width: 8px;
		height: 8px;
	}
	::-webkit-scrollbar-track {
		background: transparent;
	}
	::-webkit-scrollbar-thumb {
		background: rgba(79, 70, 229, 0.25);
		border-radius: 9999px;
	}
	::-webkit-scrollbar-thumb:hover {
		background: rgba(79, 70, 229, 0.35);
	}

	@media (max-width: 640px) {
		.parallax-hero {
			padding-top: 4.5rem;
			padding-bottom: 2.5rem;
		}
	}
</style>
