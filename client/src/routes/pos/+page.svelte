<script lang="ts">
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
	import { crmApi } from '$lib/api/resources';
	import { toast } from 'svelte-sonner';
	import * as Dialog from '$lib/components/ui/dialog';
	import { auth } from '$lib/stores/auth';
	import { goto } from '$app/navigation';

	$effect(() => {
		if (!auth.hasPermission('pos.view')) {
			toast.error('Access Denied', { description: 'You do not have permission to access POS.' });
			goto('/');
		}
	});

	// Runes state
	let products = $state<any[]>([]);
	let cart = $state<any[]>([]);
	let searchTerm = $state('');
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

	const currencyFormatter = new Intl.NumberFormat('en-US', {
		style: 'currency',
		currency: 'USD'
	});

	const formatCurrency = (value?: number | null) => {
		if (value === null || value === undefined || Number.isNaN(value)) return '$0.00';
		return currencyFormatter.format(value);
	};

	const fetchProducts = async (search = '') => {
		try {
			// Efficiently fetch products with stock in one go
			const response = await api.get('/sales/products');
			let productsData = response.data.products;

			// Client-side filtering for search (since the optimized endpoint returns all active products for POS cache)
			// For a large catalog, we should add server-side search to the sales/products endpoint.
			// Assuming simpler POS requirement for now, or we can filter if the endpoint supports it.
			// The current implementations of ListProducts doesn't accept search query, so we filter here.
			if (search) {
				const lowerSearch = search.toLowerCase();
				productsData = productsData.filter(
					(p: any) =>
						p.Name.toLowerCase().includes(lowerSearch) ||
						p.SKU.toLowerCase().includes(lowerSearch) ||
						p.ID.toString().includes(search)
				);
			}

			// Map to expected structure (or update usage in template)
			// Flattening structure for easier access
			products = productsData.map((p: any) => ({
				...p,
				stock: { currentQuantity: p.StockQuantity } // Adapter for existing template usage
			}));
		} catch (error) {
			console.error('Error fetching products:', error);
			// toast.error('Failed to load product catalog');
		}
	};

	onMount(() => {
		fetchProducts();

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
					toast.error('Customer not found');
				}
			}
		} catch (error) {
			console.error('Error fetching customer:', error);
			selectedCustomer = null;
			toast.error('Error searching for customer');
		}
	};

	const addToCart = (product: any) => {
		const existingItem = cart.find((item) => item.id === product.ID);
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
			item.quantity = quantity;
			cart = [...cart];
		}
	};

	const clearCart = () => {
		cart = [];
	};

	const subtotal = $derived(
		cart.reduce((acc, item) => acc + (item.SellingPrice || 0) * item.quantity, 0)
	);
	const taxRate = 0.1; // 10% tax
	const tax = $derived(subtotal * taxRate);
	const total = $derived(subtotal + tax);
	const canCompleteOrder = $derived(cart.length > 0 && !!paymentMethod);

	const setPayment = (method: string) => {
		paymentMethod = method;
	};

	const completeOrder = async () => {
		if (!canCompleteOrder || isProcessing) return;
		isProcessing = true;

		const toastId = toast.loading('Processing transaction...');

		try {
			const payload = {
				items: cart.map((item) => ({
					productId: item.ID,
					quantity: item.quantity
				})),
				customerId: selectedCustomer ? selectedCustomer.ID : null,
				paymentMethod: paymentMethod
			};

			await api.post('/sales/checkout', payload);

			toast.success('Order completed successfully!', { id: toastId });
			cart = [];
			paymentMethod = null;
			await fetchProducts(); // Refresh stock
		} catch (error: any) {
			console.error('Error completing order:', error);
			const msg = error.response?.data?.error || 'Failed to complete order';
			toast.error(`Transaction Failed: ${msg}`, { id: toastId });
		} finally {
			isProcessing = false;
		}
	};

	const saveNewCustomer = async () => {
		const nameParts = newCustomerForm.name.trim().split(' ');
		const firstName = nameParts[0] || '';
		const lastName = nameParts.slice(1).join(' ') || '';

		if (!firstName) {
			toast.error('Name is required');
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
			toast.success('Customer created and selected!');
			newCustomerForm = { name: '', email: '', phone: '' };
		} catch (error) {
			console.error('Error creating customer:', error);
			toast.error('Failed to create customer');
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
					Point of Sale
				</p>
				<p class="text-[0.65rem] text-slate-500 sm:text-[0.7rem]">
					Live checkout canvas for counter teams
				</p>
			</div>
		</div>

		<h1
			class="mb-3 bg-gradient-to-r from-slate-900 via-indigo-700 to-sky-700 bg-clip-text text-3xl font-bold text-transparent sm:text-4xl lg:text-5xl"
		>
			Unified Checkout Console
		</h1>
		<p class="mx-auto max-w-2xl text-sm text-slate-600 sm:mx-0 sm:text-base">
			Scan, search, and complete orders with a low-friction flow that stays in sync with your
			catalog.
		</p>

		<div class="mt-6 flex flex-col justify-center gap-3 sm:flex-row sm:justify-start">
			<Button
				variant="secondary"
				onclick={clearCart}
				class="w-full rounded-xl bg-gradient-to-r from-indigo-500 to-sky-500 px-5 py-2.5 font-medium text-white shadow-lg transition-all duration-300 hover:scale-105 hover:from-indigo-600 hover:to-sky-600 hover:shadow-xl focus:ring-2 focus:ring-indigo-300 sm:w-auto"
			>
				<Zap class="mr-2 h-4 w-4" />
				New walk-in sale
			</Button>
			<Button
				variant="outline"
				onclick={() => fetchProducts(searchTerm.trim())}
				class="w-full rounded-xl border border-indigo-100 bg-white/80 px-5 py-2.5 font-medium text-indigo-700 shadow-md transition-all duration-300 hover:scale-105 hover:bg-indigo-50 hover:shadow-lg focus:ring-2 focus:ring-indigo-200 sm:w-auto"
			>
				<Search class="mr-2 h-4 w-4" />
				Refresh catalog
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
						Point of Sale
					</CardTitle>
					<CardDescription class="text-[0.75rem] text-slate-500">
						Tap products to build the cart, review below, then confirm on the right.
					</CardDescription>
				</div>
				<div class="hidden items-center gap-2 text-[0.7rem] text-slate-500 sm:flex">
					<span class="rounded-full border border-slate-200 bg-slate-50 px-2 py-0.5"
						>Super shop mode</span
					>
				</div>
			</CardHeader>
			<CardContent class="pb-3 pt-0">
				<div class="flex flex-col items-stretch gap-3 sm:flex-row">
					<div class="relative flex-1">
						<Search class="absolute left-2.5 top-1/2 h-4 w-4 -translate-y-1/2 text-slate-400" />
						<Input
							bind:value={searchTerm}
							placeholder="Search by name, barcode, or SKU..."
							class="rounded-xl border-slate-200 bg-slate-50/80 pl-8 text-sm focus-visible:ring-indigo-300"
							onkeydown={(e) => e.key === 'Enter' && handleSearch()}
						/>
					</div>
					<Button
						class="rounded-xl bg-gradient-to-r from-indigo-500 to-sky-500 text-white shadow-md hover:from-indigo-600 hover:to-sky-600"
						onclick={handleSearch}
					>
						<Search class="mr-2 h-4 w-4" />
						Search
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
						<div class="flex items-center justify-between gap-2">
							<div>
								<CardTitle class="text-sm text-slate-800">Products</CardTitle>
								<CardDescription class="text-[0.75rem] text-slate-500">
									Tap a tile to add it to the active cart.
								</CardDescription>
							</div>
							<div class="text-[0.7rem] text-slate-500">
								<span class="font-medium text-slate-700">{products.length}</span> results
							</div>
						</div>
					</CardHeader>
					<CardContent class="max-h-[22rem] overflow-y-auto p-3 pt-3">
						{#if products.length === 0}
							<div class="flex h-full items-center justify-center text-[0.8rem] text-slate-500">
								No products found. Try adjusting your search.
							</div>
						{:else}
							<div class="grid grid-cols-2 gap-3 xl:grid-cols-3">
								{#each products as product}
									<button
										type="button"
										onclick={() => addToCart(product)}
										class="group flex flex-col justify-between rounded-xl border border-indigo-50 bg-white/80 p-3 text-left shadow-sm transition-all duration-200 hover:border-indigo-200 hover:bg-indigo-50/80 hover:shadow-md"
									>
										<div class="space-y-1.5">
											<div class="line-clamp-2 text-[0.8rem] font-medium text-slate-900">
												{product.Name}
											</div>
											<div class="mt-0.5 flex items-center justify-between">
												<div class="text-[0.8rem] font-semibold text-indigo-600">
													{formatCurrency(product.SellingPrice)}
												</div>
												<div class="flex items-center gap-1 text-[0.7rem] text-slate-500">
													<span
														class="h-1.5 w-1.5 rounded-full {product.stock &&
														product.stock.currentQuantity > 0
															? 'bg-emerald-500'
															: 'bg-rose-400'}"
													/>
													<span>
														{product.stock ? product.stock.currentQuantity : 'N/A'} in stock
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
												to add
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
							<CardTitle class="text-sm text-slate-900">Cart</CardTitle>
							<CardDescription class="text-[0.75rem] text-slate-500">
								{cart.length === 0
									? 'No items added yet.'
									: `${cart.length} item${cart.length > 1 ? 's' : ''} in cart`}
							</CardDescription>
						</div>

						{#if cart.length > 0}
							<Button
								variant="ghost"
								size="sm"
								class="rounded-xl px-3 py-1.5 text-[0.75rem] text-slate-500 hover:bg-rose-50 hover:text-rose-500"
								onclick={clearCart}
							>
								Clear cart
							</Button>
						{/if}
					</CardHeader>

					<CardContent class="max-h-[20rem] overflow-y-auto pt-0">
						{#if cart.length === 0}
							<div class="py-6 text-center text-[0.8rem] text-slate-400">
								Add products from the grid above to start a new order.
							</div>
						{:else}
							<!-- ⭐ FIX: force proper word wrapping and no horizontal scroll -->
							<Table class="w-full table-fixed">
								<TableHeader>
									<TableRow class="border-slate-100 bg-slate-50/80">
										<TableHead class="w-1/3 text-[0.7rem] text-slate-500">Product</TableHead>
										<TableHead class="text-[0.7rem] text-slate-500">Price</TableHead>
										<TableHead class="text-[0.7rem] text-slate-500">Qty</TableHead>
										<TableHead class="text-right text-[0.7rem] text-slate-500">Total</TableHead>
										<TableHead />
									</TableRow>
								</TableHeader>

								<TableBody>
									{#each cart as item}
										<TableRow class="border-slate-100 hover:bg-slate-50/60">
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
												</div>
											</TableCell>

											<TableCell class="align-top text-[0.8rem] text-slate-800">
												{formatCurrency(item.SellingPrice)}
											</TableCell>

											<!-- Quantity input -->
											<TableCell class="min-w-[4rem] align-top">
												<Input
													type="number"
													class="h-8 w-full rounded-lg border-slate-200 bg-slate-50/80 px-2 text-[0.8rem]"
													min="1"
													value={item.quantity}
													onchange={(e) => updateQuantity(item.id, parseInt(e.currentTarget.value))}
												/>
											</TableCell>

											<TableCell
												class="text-right align-top text-[0.8rem] font-semibold text-slate-900"
											>
												{formatCurrency((item.SellingPrice || 0) * item.quantity)}
											</TableCell>

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
						<CardTitle class="text-sm text-slate-900">Customer</CardTitle>
						<CardDescription class="text-[0.75rem] text-slate-500">
							Attach a customer by ID, username, email, or phone. Optional for walk-ins.
						</CardDescription>
					</CardHeader>
					<CardContent class="space-y-3 pt-3">
						<div class="flex items-center gap-2">
							<div class="relative flex-1">
								<Search class="absolute left-2.5 top-1/2 h-4 w-4 -translate-y-1/2 text-slate-400" />
								<Input
									bind:value={customerSearchTerm}
									placeholder="Search by ID, username, email, phone"
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
								<span class="font-medium">New</span>
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
											{selectedCustomer.loyalty.Points} pts
										</span>
										<span
											class="rounded-sm bg-emerald-100 px-1 text-[0.65rem] uppercase text-emerald-800"
										>
											{selectedCustomer.loyalty.Tier}
										</span>
									</div>
								{/if}
							</div>
						{:else}
							<div class="text-[0.75rem] text-slate-400">
								No customer selected. You can still complete a walk-in sale.
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
						<CardTitle class="text-sm text-slate-900">Payment</CardTitle>
						<CardDescription class="text-[0.75rem] text-slate-500">
							Choose how the customer is paying for this order.
						</CardDescription>
					</CardHeader>
					<CardContent class="pt-3">
						<div class="grid grid-cols-2 gap-3 sm:grid-cols-4">
							{#each [{ id: 'CASH', label: 'Cash', icon: Banknote, color: 'text-emerald-500', sub: 'Physical' }, { id: 'CARD', label: 'Card', icon: CreditCard, color: 'text-violet-500', sub: 'Terminal' }, { id: 'BKASH', label: 'bKash', icon: QrCode, color: 'text-pink-500', sub: 'Mobile' }, { id: 'OTHER', label: 'Other', icon: Wallet, color: 'text-orange-500', sub: 'Check/Due' }] as method}
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
							<CardTitle class="text-[0.9rem] text-slate-50/90">Order Summary</CardTitle>
							<CardDescription class="text-[0.7rem] text-slate-300">
								Review totals and payment before confirming the sale.
							</CardDescription>
						</CardHeader>
						<CardContent class="space-y-2 pt-0 text-[0.8rem]">
							<div class="flex justify-between text-slate-200">
								<span>Subtotal</span>
								<span>{formatCurrency(subtotal)}</span>
							</div>
							<div class="flex justify-between text-[0.75rem] text-slate-300/90">
								<span>Tax ({(taxRate * 100).toFixed(0)}%)</span>
								<span>{formatCurrency(tax)}</span>
							</div>
							<div class="flex items-center justify-between border-t border-slate-700/70 pt-2">
								<span class="text-[0.85rem] font-medium text-slate-100">Total</span>
								<span class="text-lg font-semibold tracking-tight text-slate-50">
									{formatCurrency(total)}
								</span>
							</div>

							<div class="flex items-center justify-between pt-1 text-[0.7rem] text-slate-300/90">
								<div>
									<span class="font-medium">Payment:</span>
									<span class="ml-1">
										{paymentMethod ? paymentMethod : 'Not selected'}
									</span>
								</div>
								<div>
									<span class="font-medium">Items:</span>
									<span class="ml-1">{cart.length}</span>
								</div>
							</div>
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
									Processing...
								{:else if !cart.length}
									Add items to cart to continue
								{:else if !paymentMethod}
									Select a payment method to complete
								{:else}
									Complete Order
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
						New Customer
					</Dialog.Title>
					<Dialog.Description class="mt-1.5 text-violet-100">
						Add a new member to your customer base.
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
							Full Name
						</label>
						<div class="relative">
							<div class="pointer-events-none absolute inset-y-0 left-0 flex items-center pl-3">
								<Users class="h-4 w-4 text-slate-400 group-focus-within:text-violet-500" />
							</div>
							<Input
								id="name"
								bind:value={newCustomerForm.name}
								placeholder="Jane Doe"
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
							Email Address
						</label>
						<div class="relative">
							<div class="pointer-events-none absolute inset-y-0 left-0 flex items-center pl-3">
								<Mail class="h-4 w-4 text-slate-400 group-focus-within:text-violet-500" />
							</div>
							<Input
								id="email"
								type="email"
								bind:value={newCustomerForm.email}
								placeholder="jane@example.com"
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
							Phone Number
						</label>
						<div class="relative">
							<div class="pointer-events-none absolute inset-y-0 left-0 flex items-center pl-3">
								<Phone class="h-4 w-4 text-slate-400 group-focus-within:text-violet-500" />
							</div>
							<Input
								id="phone"
								bind:value={newCustomerForm.phone}
								placeholder="+1 (555) 000-0000"
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
					Cancel
				</Button>
				<Button
					onclick={saveNewCustomer}
					class="rounded-xl bg-gradient-to-r from-violet-600 to-indigo-600 px-8 font-semibold text-white shadow-lg shadow-violet-500/25 transition-all duration-300 hover:scale-[1.02] hover:shadow-xl hover:shadow-violet-500/35 active:scale-[0.98]"
				>
					Create Customer
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
