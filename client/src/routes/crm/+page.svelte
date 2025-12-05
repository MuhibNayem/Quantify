<script lang="ts">
	import { onMount } from 'svelte';
	import { Input } from '$lib/components/ui/input';
	import { Button } from '$lib/components/ui/button';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import { Badge } from '$lib/components/ui/badge';
	import {
		UserPlus,
		Search,
		Star,
		Gift,
		Edit,
		Trash2,
		Users,
		Mail,
		Phone,
		Crown,
		Award,
		Zap,
		Filter,
		ChevronRight,
		ChevronLeft,
		Edit2
	} from 'lucide-svelte';
	import { toast } from 'svelte-sonner';
	import * as Dialog from '$lib/components/ui/dialog';
	import {
		Table,
		TableBody,
		TableCell,
		TableHead,
		TableHeader,
		TableRow
	} from '$lib/components/ui/table';
	import {
		Root,
		Content,
		Item,
		PrevButton,
		NextButton,
		Ellipsis,
		Link
	} from '$lib/components/ui/pagination';
	import DataTable from '$lib/components/ui/data-table/DataTable.svelte';
	import { crmApi } from '$lib/api/resources';
	import type { UserSummary } from '$lib/types';

	let customers = $state<UserSummary[]>([]);
	let filteredCustomers = $state<UserSummary[]>([]);
	let selectedCustomer = $state<UserSummary | null>(null);
	let searchTerm = $state('');
	let loading = $state(true);
	let isModalOpen = $state(false);
	let editingCustomer = $state<UserSummary | null>(null);
	let pointsToAdd = $state(0);
	let pointsToRedeem = $state(0);
	let showFilters = $state(false);
	let selectedTier = $state('all');

	let pagination = $state({
		currentPage: 1,
		totalPages: 1,
		totalItems: 0,
		itemsPerPage: 10
	});

	let customerForm = $state({
		name: '',
		email: '',
		phone: ''
	});

	const openModal = (customer: UserSummary | null = null) => {
		console.log('Opening modal for:', customer);
		if (customer) {
			editingCustomer = customer;
			customerForm.name = ((customer.FirstName || '') + ' ' + (customer.LastName || '')).trim();
			customerForm.email = customer.Email || '';
			customerForm.phone = customer.PhoneNumber || '';
		} else {
			editingCustomer = null;
			customerForm.name = '';
			customerForm.email = '';
			customerForm.phone = '';
		}
		isModalOpen = true;
	};

	const saveCustomer = async () => {
		const nameParts = customerForm.name.trim().split(' ');
		const firstName = nameParts[0] || '';
		const lastName = nameParts.slice(1).join(' ') || '';

		const payload = {
			username: customerForm.email, // Use email as username for now
			email: customerForm.email,
			firstName: firstName,
			lastName: lastName,
			phoneNumber: customerForm.phone,
			// For new customers, password is required by backend but not in this simple form
			// We can generate a random one or ask for it. For now, use a default implementation tweak or add password field
			// Actually backend CreateCustomer requires password. I'll use a default temp one.
			password: 'TempPassword123!'
		};

		try {
			if (editingCustomer) {
				await crmApi.updateCustomer(editingCustomer.ID, payload);
				toast.success('Customer updated');
			} else {
				await crmApi.createCustomer(payload);
				toast.success('Customer created');
			}
			isModalOpen = false;
			fetchCustomers();
		} catch (error) {
			console.error('Failed to save customer', error);
			toast.error('Failed to save customer');
		}
	};

	const deleteCustomer = async (customer: UserSummary) => {
		if (window.confirm(`Are you sure you want to delete ${customer.FirstName}?`)) {
			try {
				await crmApi.deleteCustomer(customer.ID);
				toast.success('Customer deleted');
				if (selectedCustomer?.ID === customer.ID) {
					selectedCustomer = null;
				}
				fetchCustomers();
			} catch (error) {
				console.error('Failed to delete customer', error);
				toast.error('Failed to delete customer');
			}
		}
	};

	const addPoints = async () => {
		if (selectedCustomer && pointsToAdd > 0) {
			try {
				const updatedAccount = await crmApi.addPoints(selectedCustomer.ID, pointsToAdd);
				if (selectedCustomer.loyalty) {
					selectedCustomer.loyalty = updatedAccount;
				}
				toast.success(`${pointsToAdd} points added`);
				pointsToAdd = 0;
				// Update list to reflect changes
				fetchCustomers();
			} catch (error) {
				console.error('Failed to add points', error);
				toast.error('Failed to add points');
			}
		}
	};

	const redeemPoints = async () => {
		if (selectedCustomer && pointsToRedeem > 0) {
			try {
				const updatedAccount = await crmApi.redeemPoints(selectedCustomer.ID, pointsToRedeem);
				if (selectedCustomer.loyalty) {
					selectedCustomer.loyalty = updatedAccount;
				}
				toast.success(`${pointsToRedeem} points redeemed`);
				pointsToRedeem = 0;
				fetchCustomers();
			} catch (error: any) {
				console.error('Failed to redeem points', error);
				toast.error(error.response?.data?.message || 'Failed to redeem points');
			}
		}
	};

	const fetchCustomers = async () => {
		loading = true;
		try {
			const response = await crmApi.listCustomers({
				page: pagination.currentPage,
				limit: pagination.itemsPerPage,
				q: searchTerm
			});
			customers = response.users || [];
			filteredCustomers = customers; // Server handles filtering mostly, but for Tier we might need client side or unsupported
			pagination.totalItems = response.totalItems;
			pagination.totalPages = response.totalPages;
		} catch (error) {
			console.error('Failed to fetch customers', error);
			toast.error('Failed to load customers');
		} finally {
			loading = false;
		}
	};

	const handleSearch = () => {
		pagination.currentPage = 1;
		fetchCustomers();
	};

	const handlePageChange = (page: number) => {
		pagination.currentPage = page;
		fetchCustomers();
	};

	const getTierColor = (tier: string) => {
		switch (tier) {
			case 'Bronze':
				return 'bg-orange-100 text-orange-800 border-orange-200';
			case 'Silver':
				return 'bg-indigo-100 text-indigo-800 border-indigo-200';
			case 'Gold':
				return 'bg-amber-400 text-amber-900 border-amber-500';
			default:
				return 'bg-gray-100 text-gray-800 border-gray-200';
		}
	};

	const getTierIcon = (tier: string) => {
		switch (tier) {
			case 'Bronze':
				return Award;
			case 'Silver':
				return Star;
			case 'Gold':
				return Crown;
			default:
				return Star;
		}
	};

	onMount(() => {
		fetchCustomers();
	});
</script>

<!-- ===== HERO SECTION ===== -->
<section class="relative isolate w-full overflow-hidden">
	<!-- Gradient background with motion -->
	<div
		class="animate-gradientShift absolute inset-0 -z-10 bg-gradient-to-r from-violet-400 via-purple-400 to-indigo-500 bg-[length:200%_200%]"
	></div>

	<!-- Floating glow blobs -->
	<div
		class="animate-pulseGlow absolute -left-24 -top-32 h-96 w-96 rounded-full bg-violet-300/50 blur-3xl"
	></div>
	<div
		class="animate-pulseGlow absolute -bottom-28 -right-24 h-80 w-80 rounded-full bg-indigo-300/40 blur-3xl delay-700"
	></div>

	<!-- Hero container -->
	<div class="relative mx-auto max-w-7xl px-4 pb-10 pt-16 sm:px-6 sm:pb-16 sm:pt-20 lg:px-8">
		<div class="flex flex-col items-start justify-between gap-6 lg:flex-row lg:items-center">
			<div>
				<div class="mb-3 inline-flex items-center gap-3">
					<span
						class="animate-cardFloat inline-flex rounded-2xl bg-gradient-to-br from-violet-500 to-indigo-600 p-2 shadow-md"
					>
						<Users class="h-6 w-6 text-white" />
					</span>
					<p class="text-xs font-semibold uppercase tracking-[0.18em] text-white sm:text-sm">
						Customer Loyalty
					</p>
				</div>

				<h1
					class="mb-3 bg-gradient-to-r from-white via-gray-100 to-gray-200 bg-clip-text text-3xl font-bold text-transparent sm:text-4xl lg:text-5xl"
				>
					Customer Management
				</h1>
				<p class="max-w-2xl text-sm text-white/90 sm:text-base">
					Manage customer relationships and loyalty programs
				</p>
			</div>

			<!-- Action buttons -->
			<div class="flex flex-col gap-3 sm:flex-row">
				<Button
					variant="secondary"
					onclick={fetchCustomers}
					class="rounded-xl bg-gradient-to-r from-violet-600 to-indigo-600 px-5 py-2.5 font-medium text-white shadow-lg transition-all duration-300 hover:scale-105 hover:from-violet-700 hover:to-indigo-700 hover:shadow-xl focus:ring-2 focus:ring-violet-300"
				>
					<Zap class="mr-2 h-4 w-4" /> Refresh
				</Button>
				<Button
					onclick={() => openModal()}
					class="rounded-xl border border-white/30 bg-white/20 px-5 py-2.5 font-medium text-white shadow-md transition-all duration-300 hover:scale-105 hover:bg-white/30 hover:shadow-lg focus:ring-2 focus:ring-white/50"
				>
					<UserPlus class="mr-2 h-4 w-4" /> Add Customer
				</Button>
			</div>
		</div>
	</div>
</section>

<!-- ===== SEARCH AND FILTERS ===== -->
<div class="mx-auto mt-6 max-w-7xl px-4 sm:px-6 lg:px-8">
	<div class="rounded-2xl border border-violet-200/50 bg-white/90 p-6 shadow-lg backdrop-blur">
		<div class="flex flex-col items-start justify-between gap-4 lg:flex-row lg:items-center">
			<div class="w-full flex-1 lg:w-auto">
				<div class="relative">
					<Search class="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-gray-400" />
					<Input
						bind:value={searchTerm}
						oninput={handleSearch}
						placeholder="Search customers by name or email..."
						class="w-full rounded-xl border-violet-300 bg-white/90 px-3.5 py-2.5 pl-10 text-sm focus:ring-2 focus:ring-violet-500"
					/>
				</div>
			</div>

			<div class="flex items-center gap-3">
				<Button
					variant="outline"
					onclick={() => (showFilters = !showFilters)}
					class="rounded-xl border-violet-300 px-4 py-2.5 text-violet-700 hover:bg-violet-50"
				>
					<Filter class="mr-2 h-4 w-4" /> Filters
					{#if selectedTier !== 'all'}
						<Badge class="ml-2 bg-violet-100 text-violet-700">1</Badge>
					{/if}
				</Button>
			</div>
		</div>

		{#if showFilters}
			<div class="mt-4 border-t border-violet-200/50 pt-4">
				<div class="flex items-center gap-3">
					<span class="text-sm font-medium text-slate-700">Tier:</span>
					<div class="flex gap-2">
						{#each ['all', 'Bronze', 'Silver', 'Gold'] as tier}
							<Button
								variant={selectedTier === tier ? 'default' : 'ghost'}
								size="sm"
								class={selectedTier === tier
									? 'bg-violet-600 text-white'
									: 'text-slate-600 hover:bg-violet-50 hover:text-violet-700'}
								onclick={() => {
									selectedTier = tier;
									handleSearch();
								}}
							>
								{tier === 'all' ? 'All Tiers' : tier}
							</Button>
						{/each}
					</div>
				</div>
			</div>
		{/if}
	</div>
</div>

<!-- ===== MAIN CONTENT ===== -->
<div class="mx-auto mb-12 mt-6 max-w-7xl px-4 sm:px-6 lg:px-8">
	<div class="grid grid-cols-1 gap-8 lg:grid-cols-3">
		<!-- Customer Table -->
		<div class="lg:col-span-2">
			<DataTable
				data={customers}
				columns={[
					{ header: 'Customer', class: 'w-[250px]' },
					{ header: 'Contact' },
					{ header: 'Loyalty Tier' },
					{ header: 'Points', class: 'text-right' },
					{ header: 'Actions', class: 'text-right' }
				]}
				totalItems={pagination.totalItems}
				pageSize={pagination.itemsPerPage}
				currentPage={pagination.currentPage}
				onPageChange={handlePageChange}
				{loading}
				onRowClick={(customer) => (selectedCustomer = customer)}
			>
				{#snippet children(customer)}
					<TableCell>
						<div class="font-medium">
							{(customer.FirstName + ' ' + (customer.LastName || '')).trim()}
						</div>
						<div class="text-muted-foreground text-sm">{customer.Email}</div>
					</TableCell>
					<TableCell>{customer.PhoneNumber}</TableCell>
					<TableCell>
						<Badge class={getTierColor(customer.loyalty?.Tier || 'Bronze')}>
							<svelte:component
								this={getTierIcon(customer.loyalty?.Tier || 'Bronze')}
								class="mr-1 h-3 w-3"
							/>
							{customer.loyalty?.Tier || 'Bronze'}
						</Badge>
					</TableCell>
					<TableCell class="text-right font-mono">{customer.loyalty?.Points || 0}</TableCell>
					<TableCell class="text-right">
						<div class="flex justify-end gap-2" onclick={(e) => e.stopPropagation()}>
							<Button
								size="sm"
								variant="outline"
								onclick={() => openModal(customer)}
								class="h-8 w-8 rounded-lg border-violet-300 p-0 text-violet-700 hover:bg-violet-50"
							>
								<Edit class="h-4 w-4" />
							</Button>
							<Button
								size="sm"
								variant="destructive"
								onclick={() => deleteCustomer(customer)}
								class="h-8 w-8 rounded-lg bg-red-500/20 p-0 text-red-500 hover:bg-red-500/30"
							>
								<Trash2 class="h-4 w-4 text-red-500" />
							</Button>
						</div>
					</TableCell>
				{/snippet}
			</DataTable>
		</div>

		<!-- Customer Details Sidebar -->
		<div class="lg:col-span-1">
			{#if selectedCustomer}
				<div class="sticky top-6 space-y-6">
					<!-- Customer Profile Card -->
					<div
						class="rounded-2xl bg-gradient-to-br from-violet-500 to-indigo-600 p-6 text-white shadow-xl"
					>
						<div class="mb-4 flex items-center justify-between">
							<h2 class="text-xl font-bold">
								{(selectedCustomer.FirstName + ' ' + (selectedCustomer.LastName || '')).trim()}
							</h2>
							<div class="flex gap-2">
								<Button
									size="sm"
									variant="outline"
									onclick={() => openModal(selectedCustomer)}
									class="rounded-lg border-white/30 text-white hover:bg-white/20"
								>
									<Edit2 class="h-4 w-4" />
								</Button>
								<Button
									size="sm"
									variant="destructive"
									onclick={() => deleteCustomer(selectedCustomer)}
									class="rounded-lg bg-red-500/20 text-red-500 hover:bg-red-500/30"
								>
									<Trash2 class="h-4 w-4 text-red-500" />
								</Button>
							</div>
						</div>

						<div class="space-y-3">
							<div class="flex items-center gap-3">
								<Mail class="h-4 w-4 text-white/80" />
								<span class="text-sm text-white/90">{selectedCustomer.Email}</span>
							</div>
							<div class="flex items-center gap-3">
								<Phone class="h-4 w-4 text-white/80" />
								<span class="text-sm text-white/90">{selectedCustomer.PhoneNumber}</span>
							</div>
						</div>
					</div>

					<!-- Loyalty Card -->
					<div
						class="rounded-2xl border border-violet-200/50 bg-white/90 p-6 shadow-lg backdrop-blur"
					>
						<h3 class="mb-4 flex items-center gap-2 text-lg font-semibold text-slate-800">
							<Star class="text-violet-600" />
							Loyalty Program
						</h3>

						<div class="mb-6 text-center">
							<p class="text-3xl font-bold text-violet-600">
								{selectedCustomer.loyalty?.Points || 0}
							</p>
							<p class="text-sm text-slate-500">Available Points</p>
							<Badge class={`mt-2 ${getTierColor(selectedCustomer.loyalty?.Tier || 'Bronze')}`}>
								<svelte:component
									this={getTierIcon(selectedCustomer.loyalty?.Tier || 'Bronze')}
									class="mr-1 h-3 w-3"
								/>
								{selectedCustomer.loyalty?.Tier || 'Bronze'} Member
							</Badge>
						</div>

						<div class="space-y-3">
							<div>
								<label class="text-sm font-medium text-slate-700">Add Points</label>
								<div class="mt-1 flex gap-2">
									<Input
										type="number"
										placeholder="Amount"
										bind:value={pointsToAdd}
										class="rounded-lg border-violet-300 bg-white/90 px-3 py-2 text-sm focus:ring-2 focus:ring-violet-500"
									/>
									<Button
										onclick={addPoints}
										class="rounded-lg bg-violet-600 px-3 py-2 text-white hover:bg-violet-700"
									>
										<Gift class="h-4 w-4" />
									</Button>
								</div>
							</div>

							<div>
								<label class="text-sm font-medium text-slate-700">Redeem Points</label>
								<div class="mt-1 flex gap-2">
									<Input
										type="number"
										placeholder="Amount"
										bind:value={pointsToRedeem}
										class="rounded-lg border-violet-300 bg-white/90 px-3 py-2 text-sm focus:ring-2 focus:ring-violet-500"
									/>
									<Button
										onclick={redeemPoints}
										class="rounded-lg bg-indigo-600 px-3 py-2 text-white hover:bg-indigo-700"
									>
										Redeem
									</Button>
								</div>
							</div>
						</div>
					</div>

					<!-- Quick Stats -->
					<div
						class="rounded-2xl border border-violet-200/50 bg-white/90 p-6 shadow-lg backdrop-blur"
					>
						<h3 class="mb-4 text-lg font-semibold text-slate-800">Quick Stats</h3>
						<div class="space-y-3">
							<div class="flex justify-between">
								<span class="text-sm text-slate-600">Joined Date</span>
								<span class="text-sm font-medium text-slate-800"
									>{new Date(selectedCustomer.CreatedAt || '').toLocaleDateString()}</span
								>
							</div>
							<div class="flex justify-between">
								<span class="text-sm text-slate-600">Last Activity</span>
								<span class="text-sm font-medium text-slate-800"
									>{new Date(selectedCustomer.UpdatedAt || '').toLocaleDateString()}</span
								>
							</div>
						</div>
					</div>
				</div>
			{:else}
				<div class="sticky top-6">
					<div
						class="rounded-2xl border border-violet-200/50 bg-white/90 p-12 text-center backdrop-blur"
					>
						<Users class="mx-auto mb-4 h-16 w-16 text-violet-300" />
						<p class="text-lg font-medium text-slate-600">Select a Customer</p>
						<p class="text-sm text-slate-500">Choose a customer from list to view details</p>
					</div>
				</div>
			{/if}
		</div>
	</div>
</div>

<Dialog.Root bind:open={isModalOpen}>
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
					{editingCustomer ? 'Edit Customer' : 'New Customer'}
				</Dialog.Title>
				<Dialog.Description class="mt-1.5 text-violet-100">
					{editingCustomer
						? "Update the customer's profile information."
						: 'Add a new member to your customer base.'}
				</Dialog.Description>
			</div>
		</div>

		<!-- Body -->
		<div class="space-y-6 px-6 py-8 sm:px-10">
			{#key isModalOpen}
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
							bind:value={customerForm.name}
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
							bind:value={customerForm.email}
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
							bind:value={customerForm.phone}
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
				onclick={() => (isModalOpen = false)}
				class="rounded-xl text-slate-600 hover:bg-slate-100 hover:text-slate-900"
			>
				Cancel
			</Button>
			<Button
				onclick={saveCustomer}
				class="rounded-xl bg-gradient-to-r from-violet-600 to-indigo-600 px-8 font-semibold text-white shadow-lg shadow-violet-500/25 transition-all duration-300 hover:scale-[1.02] hover:shadow-xl hover:shadow-violet-500/35 active:scale-[0.98]"
			>
				{editingCustomer ? 'Update Profile' : 'Create Profile'}
			</Button>
		</div>
	</Dialog.Content>
</Dialog.Root>

<style lang="postcss">
	/* Smooth transitions globally */
	* {
		transition-property:
			color, background-color, border-color, text-decoration-color, fill, stroke, opacity,
			box-shadow, transform, filter, backdrop-filter;
		transition-timing-function: cubic-bezier(0.4, 0, 0.2, 1);
		transition-duration: 300ms;
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
		animation: gradientShift 16s ease-in-out infinite;
	}

	/* Soft glowing blobs */
	@keyframes pulseGlow {
		0%,
		100% {
			transform: scale(1);
			opacity: 0.45;
			filter: blur(80px);
		}
		50% {
			transform: scale(1.08);
			opacity: 0.6;
			filter: blur(90px);
		}
	}
	.animate-pulseGlow {
		animation: pulseGlow 10s ease-in-out infinite;
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
	.animate-fadeUp {
		animation: fadeUp 500ms var(--ease, cubic-bezier(0.4, 0, 0.2, 1)) forwards;
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
		background: rgba(139, 92, 246, 0.25);
		border-radius: 9999px;
	}
	::-webkit-scrollbar-thumb:hover {
		background: rgba(139, 92, 246, 0.35);
	}
</style>
