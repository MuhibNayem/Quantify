<script lang="ts">
	import { t } from '$lib/i18n';
	import { onMount } from 'svelte';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import {
		Card,
		CardContent,
		CardDescription,
		CardHeader,
		CardTitle
	} from '$lib/components/ui/card';
	import { Search, Zap, CreditCard, ScanBarcode } from 'lucide-svelte';
	import api from '$lib/api';
	import { crmApi, promotionsApi, paymentApi, productsApi } from '$lib/api/resources';
	import { toast } from 'svelte-sonner';
	import { auth } from '$lib/stores/auth';
	import { settings } from '$lib/stores/settings';
	import { goto } from '$app/navigation';
	import ProductGrid from './components/ProductGrid.svelte';
	import Cart from './components/Cart.svelte';
	import CustomerCard from './components/CustomerCard.svelte';
	import PaymentMethods from './components/PaymentMethods.svelte';
	import CheckoutSummary from './components/CheckoutSummary.svelte';

	$effect(() => {
		if (!auth.hasPermission('pos.view')) {
			toast.error($t('pos.toasts.access_denied'), {
				description: $t('pos.toasts.access_denied_desc')
			});
			goto('/');
		}
	});

	let products = $state<any[]>([]);
	let promotions = $state<any[]>([]);
	let cart = $state<any[]>([]);
	let searchTerm = $state('');
	let fetchLimit = $state(100);
	let fetchStatus = $state('IN_STOCK');
	let customerSearchTerm = $state('');
	let selectedCustomer = $state<any | null>(null);
	let paymentMethod = $state<string | null>(null);
	let isProcessing = $state(false);
	let pointsToRedeem = $state(0);
	let barcodeInput = $state('');
	let barcodeRef = $state<HTMLInputElement | null>(null);

	async function findProductByCode(code: string) {
		const product = products.find((p) => p.BarcodeUPC === code || p.SKU === code);
		if (product) return product;
		try {
			return await productsApi.getByBarcode(code);
		} catch {
			return null;
		}
	}

	async function handleBarcodeKeydown(e: KeyboardEvent) {
		if (e.key !== 'Enter') return;
		e.preventDefault();
		const code = barcodeInput.trim();
		if (!code) return;
		try {
			const product = await findProductByCode(code);
			if (product) {
				const mapped = product.stock
					? product
					: { ...product, stock: { currentQuantity: product.StockQuantity ?? 0 } };
				addToCart(mapped);
				toast.success($t('pos.toasts.barcode_added', { name: product.Name }));
			} else {
				toast.error($t('pos.toasts.barcode_not_found', { code }));
			}
		} finally {
			barcodeInput = '';
			barcodeRef?.focus();
		}
	}

	const fetchPromotions = async () => {
		try {
			promotions = await promotionsApi.list(true);
		} catch (error) {
			console.error('Error fetching promotions:', error);
		}
	};

	const fetchProducts = async (search = '') => {
		try {
			const response = await api.get('/sales/products', {
				params: {
					q: search,
					limit: fetchLimit,
					status: fetchStatus
				}
			});
			const productsData = response.data.products;

			products = productsData.map((p: any) => ({
				...p,
				stock: { currentQuantity: p.StockQuantity }
			}));
		} catch (error) {
			console.error('Error fetching products:', error);
		}
	};

	onMount(() => {
		Promise.all([fetchProducts(), fetchPromotions()]);
		barcodeRef?.focus();

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

	function getProductPrice(product: any): {
		price: number;
		originalPrice: number;
		discountName?: string;
	} {
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
			const response = await api.get('/crm/customers', {
				params: {
					q: customerSearchTerm,
					limit: 1
				}
			});

			if (response.data.users && response.data.users.length > 0) {
				selectedCustomer = response.data.users[0];
			} else {
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
			if (paymentMethod === 'CARD' || paymentMethod === 'BKASH') {
				const paymentResp = await paymentApi.create({
					payment_method: paymentMethod === 'CARD' ? 'card' : 'bkash',
					amount: total,
					currency: 'BDT',
					description: `POS sale - ${cart.length} item(s)`,
					customer_id: selectedCustomer ? selectedCustomer.ID : undefined
				});

				if (paymentResp.payment_url) {
					toast.success('Redirecting to payment gateway...', { id: toastId });
					window.location.href = paymentResp.payment_url;
					return;
				}
			}

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
			cart = [];
			paymentMethod = null;
			selectedCustomer = null;
			customerSearchTerm = '';
			pointsToRedeem = 0;

			await fetchProducts();
		} catch (error: any) {
			console.error('Error completing order:', error);
			const msg = error.response?.data?.error || 'Failed to complete order';
			toast.error(`${$t('pos.toasts.transaction_fail')}: ${msg}`, { id: toastId });
		} finally {
			isProcessing = false;
		}
	};

	const saveNewCustomer = async (form: {
		name: string;
		email: string;
		phone: string;
	}): Promise<boolean> => {
		const nameParts = form.name.trim().split(' ');
		const firstName = nameParts[0] || '';
		const lastName = nameParts.slice(1).join(' ') || '';

		if (!firstName) {
			toast.error($t('pos.toasts.name_required'));
			return false;
		}

		try {
			const newUser = await crmApi.createCustomer({
				username: form.email || `user${Date.now()}`,
				email: form.email,
				firstName: firstName,
				lastName: lastName,
				phoneNumber: form.phone,
				role: 'Customer',
				password: 'TempPassword123!'
			} as any);

			selectedCustomer = newUser;
			toast.success($t('pos.toasts.customer_created'));
			return true;
		} catch (error) {
			console.error('Error creating customer:', error);
			toast.error($t('pos.toasts.create_fail'));
			return false;
		}
	};
</script>

<!-- Page background gradient + hero -->
<section
	class="relative isolate min-h-screen w-full overflow-hidden bg-gradient-to-br from-indigo-50 via-sky-50 to-slate-50"
>
	<div
		class="animate-gradientShift absolute inset-0 -z-20 bg-gradient-to-r from-indigo-100 via-sky-100 to-blue-100 bg-[length:220%_220%] opacity-70"
	/>

	<div
		class="animate-pulseGlow pointer-events-none absolute -top-32 -left-24 h-80 w-80 rounded-full bg-indigo-200/40 blur-3xl"
	/>
	<div
		class="animate-pulseGlow pointer-events-none absolute -right-24 -bottom-36 h-72 w-72 rounded-full bg-sky-200/40 blur-3xl delay-700"
	/>

	<div
		class="parallax-hero relative mx-auto max-w-7xl px-4 pt-14 pb-6 text-center sm:px-6 sm:pt-20 sm:pb-10 sm:text-left lg:px-8"
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
					class="text-[0.65rem] font-semibold tracking-[0.22em] text-indigo-700 uppercase sm:text-xs"
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

	<!-- Main content -->
	<div class="flex h-full min-h-0 flex-col px-4 py-4">
		<div
			class="mb-4 flex items-center gap-3 rounded-2xl border border-indigo-200/60 bg-white/80 px-4 py-3 shadow-md backdrop-blur transition-all duration-300 focus-within:border-indigo-400 focus-within:shadow-lg focus-within:ring-2 focus-within:ring-indigo-200"
		>
			<ScanBarcode class="h-5 w-5 shrink-0 text-indigo-500" />
			<input
				bind:this={barcodeRef}
				bind:value={barcodeInput}
				onkeydown={handleBarcodeKeydown}
				placeholder={$t('pos.barcode.placeholder')}
				aria-label={$t('pos.barcode.aria_label')}
				class="flex-1 bg-transparent text-sm text-slate-800 outline-none placeholder:text-slate-400"
			/>
			<span
				class="hidden items-center gap-1.5 rounded-full bg-emerald-50 px-2.5 py-1 text-[0.65rem] font-medium text-emerald-600 sm:inline-flex"
			>
				<span class="h-1.5 w-1.5 animate-pulse rounded-full bg-emerald-500"></span>
				{$t('pos.barcode.scanner_ready')}
			</span>
		</div>

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
			<CardContent class="pt-0 pb-3">
				<div class="flex flex-col items-stretch gap-3 sm:flex-row">
					<div class="relative flex-1">
						<Search class="absolute top-1/2 left-2.5 h-4 w-4 -translate-y-1/2 text-slate-400" />
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
				<ProductGrid
					{products}
					bind:fetchStatus
					bind:fetchLimit
					{getProductPrice}
					{getAvailableStock}
					onAddToCart={addToCart}
					onFilterChange={() => fetchProducts(searchTerm.trim())}
				/>

				<Cart
					{cart}
					{getProductPrice}
					onRemove={removeFromCart}
					onUpdateQuantity={updateQuantity}
					onClear={clearCart}
				/>
			</div>

			<!-- RIGHT STACK: Customer + Payment + Summary -->
			<div class="space-y-5">
				<CustomerCard
					{selectedCustomer}
					bind:customerSearchTerm
					onSearchCustomer={handleCustomerSearch}
					onCreateCustomer={saveNewCustomer}
				/>

				<PaymentMethods
					{paymentMethod}
					{selectedCustomer}
					bind:pointsToRedeem
					onSelectPayment={setPayment}
				/>

				<CheckoutSummary
					{subtotal}
					{tax}
					{total}
					{taxRate}
					cartLength={cart.length}
					{paymentMethod}
					{selectedCustomer}
					{potentialPoints}
					{canCompleteOrder}
					{isProcessing}
					onCompleteOrder={completeOrder}
				/>
			</div>
		</div>
	</div>
</section>

<style lang="postcss">
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
