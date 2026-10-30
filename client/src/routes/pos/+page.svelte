<script lang="ts">
	import { t } from '$lib/i18n';
	import { onMount } from 'svelte';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import Kbd from '$lib/components/ui/kbd.svelte';
	import { Search, Zap, CreditCard, ScanBarcode } from 'lucide-svelte';
	import api from '$lib/api';
	import {
		barcodeApi,
		crmApi,
		promotionsApi,
		paymentApi,
		productsApi,
		unitsApi
	} from '$lib/api/resources';
	import type { ProductUnit } from '$lib/types';
	import { toast } from 'svelte-sonner';
	import { formatCurrency } from '$lib/utils';
	import { auth } from '$lib/stores/auth';
	import { settings } from '$lib/stores/settings';
	import { goto } from '$app/navigation';
	import ProductGrid from './components/ProductGrid.svelte';
	import Cart from './components/Cart.svelte';
	import CustomerCard from './components/CustomerCard.svelte';
	import PaymentMethods from './components/PaymentMethods.svelte';
	import DiscountControl from './components/DiscountControl.svelte';
	import CheckoutSummary from './components/CheckoutSummary.svelte';
	import CashDrawerBar from './components/CashDrawerBar.svelte';

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
	let dueDate = $state('');
	let barcodeInput = $state('');
	let barcodeRef = $state<HTMLInputElement | null>(null);
	let discountType = $state('percent');
	let discountValue = $state(0);
	let discountReason = $state('');

	const unitCache: Record<number, ProductUnit[]> = {};

	const inputClass = 'glass-input h-10 w-full rounded-lg px-3 text-sm font-normal';
	const btnPrimary = 'btn-soothe h-10 px-5 text-sm font-semibold';

	async function findProductByCode(code: string) {
		const product = products.find((p) => p.BarcodeUPC === code || p.SKU === code);
		if (product) return { product, unit: null };
		try {
			// Resolves product barcodes, SKUs, and per-selling-unit barcodes.
			return await barcodeApi.resolve(code);
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
			const resolved = await findProductByCode(code);
			if (resolved) {
				const product = resolved.product;
				const mapped = product.stock
					? product
					: { ...product, stock: { currentQuantity: product.StockQuantity ?? 0 } };
				await addToCart(mapped, resolved.unit);
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
					// Natural-key lookup only: email, phone, or username.
					selectedCustomer = await crmApi.lookupCustomer(customerSearchTerm);
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

	const applyScannedUnit = (item: any, unit: ProductUnit) => {
		const matched =
			item.units?.find(
				(u: ProductUnit) => u.UnitID === unit.UnitID || (u.Barcode && u.Barcode === unit.Barcode)
			) ?? unit;
		item.unitId = matched.UnitID;
		item.unitSymbol = matched.Unit?.Symbol;
		item.unitPrice =
			matched.Price > 0 ? matched.Price : getProductPrice(item).price * (matched.Factor || 1);
	};

	const addToCart = async (product: any, scannedUnit: ProductUnit | null = null) => {
		const existingItem = cart.find((item) => item.id === product.ID);
		const currentQty = existingItem ? existingItem.quantity : 0;
		const stockLimit = product.stock ? product.stock.currentQuantity : 0;

		if (currentQty >= stockLimit) {
			toast.error($t('pos.toasts.out_of_stock'));
			return;
		}

		if (existingItem) {
			existingItem.quantity++;
			if (scannedUnit) applyScannedUnit(existingItem, scannedUnit);
			cart = [...cart];
		} else {
			let units: ProductUnit[] = [];
			try {
				if (unitCache[product.ID]) {
					units = unitCache[product.ID];
				} else {
					units = (await unitsApi.productUnits(product.ID)).filter((u) => u.IsActive !== false);
					unitCache[product.ID] = units;
				}
			} catch {
				units = [];
			}
			const item: any = {
				...product,
				id: product.ID,
				quantity: 1,
				units,
				unitId: undefined,
				unitSymbol: undefined,
				unitPrice: undefined
			};
			if (scannedUnit) applyScannedUnit(item, scannedUnit);
			cart = [...cart, item];
		}
	};

	const removeFromCart = (productId: number) => {
		cart = cart.filter((item) => item.id !== productId);
	};

	const updateQuantity = (productId: number, quantity: number) => {
		if (!(quantity > 0)) quantity = 0.01;

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

	const updateUnit = (productId: number, unitId: number | undefined) => {
		const item = cart.find((item) => item.id === productId);
		if (!item) return;

		item.unitId = unitId;
		if (unitId && item.units) {
			const pu = item.units.find((u: ProductUnit) => u.UnitID === unitId);
			const basePrice = getProductPrice(item).price;
			item.unitSymbol = pu?.Unit?.Symbol;
			item.unitPrice = pu ? (pu.Price > 0 ? pu.Price : basePrice * pu.Factor) : basePrice;
		} else {
			item.unitSymbol = undefined;
			item.unitPrice = undefined;
		}
		cart = [...cart];
	};

	const clearCart = () => {
		cart = [];
		discountType = 'percent';
		discountValue = 0;
		discountReason = '';
	};

	const adjustLastItemQty = (delta: number) => {
		const lastItem = cart[cart.length - 1];
		if (!lastItem) return;
		updateQuantity(lastItem.id, lastItem.quantity + delta);
	};

	function focusBarcodeInput() {
		barcodeRef?.focus();
		barcodeRef?.select();
	}

	function handlePosKeydown(e: KeyboardEvent) {
		const el = e.target as HTMLElement;
		const isTyping = el && ['INPUT', 'TEXTAREA', 'SELECT'].includes(el.tagName);
		switch (e.key) {
			case 'F2':
				e.preventDefault();
				focusBarcodeInput();
				break;
			case 'F4':
				e.preventDefault();
				if (!isTyping) setPayment('CASH');
				break;
			case 'F5':
				e.preventDefault();
				if (!isTyping) setPayment('CARD');
				break;
			case 'F6':
				e.preventDefault();
				if (!isTyping) setPayment('BKASH');
				break;
			case 'F8':
				e.preventDefault();
				if (!isTyping && cart.length) completeOrder();
				break;
			case 'F9':
				e.preventDefault();
				if (!isTyping) clearCart();
				break;
			case '+':
			case '=':
				if (!isTyping) adjustLastItemQty(1);
				break;
			case '-':
				if (!isTyping) adjustLastItemQty(-1);
				break;
		}
	}

	const getLinePrice = (item: any): number =>
		item.unitId && item.unitPrice != null ? item.unitPrice : getProductPrice(item).price;

	const subtotal = $derived(
		cart.reduce((acc, item) => acc + getLinePrice(item) * item.quantity, 0)
	);

	const discountAvailable = $derived(
		$settings.manual_discount_enabled === 'true' && auth.hasPermission('pos.discount')
	);
	const discountMaxPercent = $derived(parseFloat($settings.manual_discount_max_percent) || 10);
	const discountMaxAmount = $derived(parseFloat($settings.manual_discount_max_amount) || 0);
	const discountRequireReason = $derived($settings.manual_discount_require_reason === 'true');

	const manualDiscountAmount = $derived.by(() => {
		if (!discountAvailable || !(discountValue > 0)) return 0;
		if (discountType === 'percent') return (subtotal * discountValue) / 100;
		if (discountType === 'fixed') return Math.min(discountValue, subtotal);
		return 0;
	});

	const discountValueInvalid = $derived(
		discountAvailable &&
			discountValue > 0 &&
			((discountType === 'percent' && discountValue > discountMaxPercent) ||
				(discountType === 'fixed' && discountMaxAmount > 0 && discountValue > discountMaxAmount))
	);
	const discountReasonMissing = $derived(
		discountAvailable && discountValue > 0 && discountRequireReason && !discountReason.trim()
	);
	const discountInvalid = $derived(discountValueInvalid || discountReasonMissing);

	const taxRate = $derived(($settings.tax_rate_percentage || 0) / 100);
	const taxableSubtotal = $derived(subtotal - manualDiscountAmount);
	const tax = $derived(taxableSubtotal * taxRate);
	const total = $derived(taxableSubtotal + tax);

	const potentialPoints = $derived(
		Math.floor(subtotal * ($settings.loyalty_points_earning_rate || 0))
	);

	const canCompleteOrder = $derived(
		cart.length > 0 &&
			!!paymentMethod &&
			(paymentMethod !== 'CREDIT' || !!selectedCustomer) &&
			!discountInvalid
	);

	const setPayment = (method: string) => {
		paymentMethod = method;
	};

	const completeOrder = async () => {
		if (!canCompleteOrder || isProcessing) return;
		if (paymentMethod === 'CREDIT' && !selectedCustomer) {
			toast.error($t('pos.toasts.customer_required_credit'));
			return;
		}
		if (discountInvalid) {
			toast.error(
				discountValueInvalid
					? discountType === 'percent'
						? $t('pos.discount.error_exceed_percent', { max: discountMaxPercent })
						: $t('pos.discount.error_exceed_amount', { max: formatCurrency(discountMaxAmount) })
					: $t('pos.discount.error_reason_required')
			);
			return;
		}
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

			const payload: Record<string, unknown> = {
				items: cart.map((item) => ({
					productId: item.id,
					quantity: item.quantity,
					...(item.unitId ? { unitId: item.unitId } : {})
				})),
				customerId: selectedCustomer ? selectedCustomer.ID : null,
				paymentMethod: paymentMethod ? paymentMethod.toLowerCase() : paymentMethod,
				pointsToRedeem: Math.min(
					pointsToRedeem,
					selectedCustomer ? selectedCustomer.loyalty?.Points || 0 : 0
				),
				manualDiscountType: manualDiscountAmount > 0 ? discountType : '',
				manualDiscountValue: manualDiscountAmount > 0 ? discountValue : 0,
				manualDiscountReason: manualDiscountAmount > 0 ? discountReason.trim() : ''
			};

			if (paymentMethod === 'CREDIT' && dueDate) {
				payload.dueDate = new Date(dueDate).toISOString();
			}

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
			dueDate = '';
			discountType = 'percent';
			discountValue = 0;
			discountReason = '';

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

<svelte:window onkeydown={handlePosKeydown} />

<div class="mx-auto w-full max-w-7xl px-4 py-8 sm:px-6 lg:px-8">
	<!-- ── Terminal header ── -->
	<header class="page-header" data-animate="fade-up">
		<div class="flex items-center gap-3">
			<span
				class="inline-flex h-11 w-11 shrink-0 items-center justify-center rounded-xl bg-primary text-primary-foreground"
			>
				<CreditCard class="h-5 w-5" />
			</span>
			<div class="min-w-0">
				<p class="overline">{$t('pos.hero.label')} · {$t('pos.hero.sub_label')}</p>
				<h1 class="page-title mt-0.5">{$t('pos.hero.title')}</h1>
				<p class="page-sub">{$t('pos.hero.subtitle')}</p>
			</div>
		</div>

		<div class="flex flex-wrap items-center gap-2">
			<Button
				variant="outline"
				onclick={() => fetchProducts(searchTerm.trim())}
				class="h-10 px-4 text-sm font-medium"
			>
				<Search class="mr-2 h-4 w-4" />
				{$t('pos.hero.refresh_catalog_btn')}
			</Button>
			<Button onclick={clearCart} class={btnPrimary}>
				<Zap class="mr-2 h-4 w-4" />
				{$t('pos.hero.new_sale_btn')}
				<Kbd
					class="ml-2 border-primary-foreground/30 bg-primary-foreground/15 text-primary-foreground"
					>F9</Kbd
				>
			</Button>
		</div>
	</header>

	<div class="mt-4" data-animate="fade-up" style="animation-delay:40ms">
		<CashDrawerBar />
	</div>

	<!-- ── Barcode capture bar ── -->
	<div
		class="panel mt-5 flex items-center gap-3 px-4 py-3 focus-within:border-primary/40"
		data-animate="fade-up"
		style="animation-delay:60ms"
	>
		<ScanBarcode class="h-5 w-5 shrink-0 text-primary" />
		<input
			bind:this={barcodeRef}
			bind:value={barcodeInput}
			onkeydown={handleBarcodeKeydown}
			placeholder={$t('pos.barcode.placeholder')}
			aria-label={$t('pos.barcode.aria_label')}
			class="flex-1 bg-transparent text-sm text-foreground outline-none placeholder:text-muted-foreground"
		/>
		<Kbd class="hidden border-border bg-muted text-muted-foreground sm:inline-block">F2</Kbd>
		<span class="badge-success hidden sm:inline-flex">
			<span class="live-dot" aria-hidden="true"></span>
			{$t('pos.barcode.scanner_ready')}
		</span>
	</div>

	<!-- ── Catalog search ── -->
	<div class="panel mt-4 p-4" data-animate="fade-up" style="animation-delay:100ms">
		<div class="flex flex-row items-center justify-between gap-3">
			<div>
				<h2 class="section-title">{$t('pos.header.title')}</h2>
				<p class="mt-0.5 text-xs text-muted-foreground">{$t('pos.header.description')}</p>
			</div>
			<div class="hidden items-center gap-2 text-[0.7rem] text-muted-foreground sm:flex">
				<span class="badge-neutral">{$t('pos.header.super_shop_mode')}</span>
			</div>
		</div>

		<div class="mt-3 flex flex-col items-stretch gap-3 sm:flex-row">
			<div class="relative flex-1">
				<Search class="absolute top-1/2 left-3 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
				<Input
					bind:value={searchTerm}
					placeholder={$t('pos.header.search_placeholder')}
					class="{inputClass} pl-9"
					onkeydown={(e) => e.key === 'Enter' && handleSearch()}
				/>
			</div>

			<Button class={btnPrimary} onclick={handleSearch}>
				<Search class="mr-2 h-4 w-4" />
				{$t('pos.header.search_btn')}
			</Button>
		</div>
	</div>

	<!-- ── Register grid ── -->
	<div class="mt-5 grid w-full grid-cols-1 gap-5 lg:grid-cols-2">
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
				{getLinePrice}
				onRemove={removeFromCart}
				onUpdateQuantity={updateQuantity}
				onUpdateUnit={updateUnit}
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
				bind:dueDate
				onSelectPayment={setPayment}
			/>

			{#if discountAvailable}
				<DiscountControl
					{subtotal}
					maxPercent={discountMaxPercent}
					maxAmount={discountMaxAmount}
					requireReason={discountRequireReason}
					amount={manualDiscountAmount}
					valueInvalid={discountValueInvalid}
					reasonMissing={discountReasonMissing}
					bind:discountType
					bind:discountValue
					bind:discountReason
				/>
			{/if}

			<CheckoutSummary
				{subtotal}
				manualDiscount={manualDiscountAmount}
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
