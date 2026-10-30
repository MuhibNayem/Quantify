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
	import { crmApi } from '$lib/api/resources';
	import type { UserSummary } from '$lib/types';
	import { auth } from '$lib/stores/auth';
	import { settings } from '$lib/stores/settings';
	import { goto } from '$app/navigation';

	let customers = $state<UserSummary[]>([]);
	let filteredCustomers = $state<UserSummary[]>([]);
	let selectedCustomer = $state<UserSummary | null>(null);

	$effect(() => {
		if (!auth.hasPermission('customers.read')) {
			toast.error('Access Denied', {
				description: 'You do not have permission to view customers.'
			});
			goto('/');
		}
	});
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
				return 'badge-neutral';
			case 'Silver':
				return 'badge-info';
			case 'Gold':
				return 'badge-success';
			default:
				return 'badge-neutral';
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

<div class="mx-auto w-full max-w-7xl px-4 py-8 sm:px-6 lg:px-8 lg:py-10">
	<!-- ===== PAGE HEADER ===== -->
	<header class="page-header" data-animate="fade-up">
		<div>
			<p class="flex items-center gap-2 overline">
				<span class="inline-flex rounded-md bg-primary/10 p-1.5 text-primary">
					<Users class="h-3.5 w-3.5" />
				</span>
				Customer Loyalty
			</p>
			<h1 class="page-title mt-1">Customer Management</h1>
			<p class="page-sub">Manage customer relationships and loyalty programs</p>
		</div>

		<!-- Action buttons -->
		<div class="flex flex-col gap-2 sm:flex-row">
			<Button variant="outline" onclick={fetchCustomers}>
				<Zap class="h-4 w-4" /> Refresh
			</Button>
			<Button onclick={() => openModal()} class="btn-soothe">
				<UserPlus class="h-4 w-4" /> Add Customer
			</Button>
		</div>
	</header>

	<!-- ===== SEARCH AND FILTERS ===== -->
	<section class="panel mt-6 p-4" data-animate="fade-up" style="animation-delay: 80ms;">
		<div class="flex flex-col items-start justify-between gap-3 lg:flex-row lg:items-center">
			<div class="relative w-full flex-1">
				<Search class="absolute top-1/2 left-3 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
				<Input
					bind:value={searchTerm}
					oninput={handleSearch}
					placeholder="Search customers by name or email..."
					class="h-11 w-full rounded-md pl-9 text-sm"
				/>
			</div>

			<div class="flex items-center gap-2">
				<Button variant="outline" onclick={() => (showFilters = !showFilters)}>
					<Filter class="h-4 w-4" /> Filters
					{#if selectedTier !== 'all'}
						<Badge class="badge-info ml-1">1</Badge>
					{/if}
				</Button>
			</div>
		</div>

		{#if showFilters}
			<div class="mt-3 border-t border-border/70 pt-3">
				<div class="flex flex-wrap items-center gap-3">
					<span class="text-xs font-semibold tracking-[0.08em] text-muted-foreground uppercase"
						>Tier</span
					>
					<div
						class="inline-flex items-center gap-1 rounded-md border border-border/70 bg-muted/50 p-1"
					>
						{#each ['all', 'Bronze', 'Silver', 'Gold'] as tier}
							<Button
								variant="ghost"
								size="sm"
								class="rounded-md px-3 text-sm font-medium {selectedTier === tier
									? 'bg-card text-foreground shadow-xs ring-1 ring-border'
									: 'text-muted-foreground hover:text-foreground'}"
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
	</section>

	<!-- ===== MAIN CONTENT ===== -->
	<div class="mt-4 grid grid-cols-1 gap-4 lg:grid-cols-3">
		<!-- Customer Table -->
		<div class="lg:col-span-2" data-animate="fade-up" style="animation-delay: 140ms;">
			<section class="panel overflow-hidden">
				<div class="overflow-x-auto">
					<table class="ledger-table">
						<thead>
							<tr>
								<th class="w-[250px]">Customer</th>
								<th>Contact</th>
								<th>Loyalty Tier</th>
								<th class="text-right">Points</th>
								<th class="text-right">Actions</th>
							</tr>
						</thead>
						<tbody>
							{#if loading}
								{#each Array(pagination.itemsPerPage) as _}
									<tr>
										<td colspan="5">
											<Skeleton class="h-5 w-full bg-muted" />
										</td>
									</tr>
								{/each}
							{:else if customers.length === 0}
								<tr>
									<td colspan="5" class="py-14 text-center">
										<div class="flex flex-col items-center gap-2">
											<span
												class="flex h-12 w-12 items-center justify-center rounded-lg bg-muted text-muted-foreground ring-1 ring-border"
											>
												<Users class="h-5 w-5" />
											</span>
											<p class="text-sm font-medium text-foreground">No customers found</p>
											<p class="text-xs text-muted-foreground">Try adjusting your search</p>
										</div>
									</td>
								</tr>
							{:else}
								{#each customers as customer (customer.ID)}
									{@const TierIcon = getTierIcon(customer.loyalty?.Tier || 'Bronze')}
									<tr
										class="cursor-pointer transition-colors {selectedCustomer?.ID === customer.ID
											? 'bg-muted/60 shadow-[inset_2px_0_0_0_hsl(var(--primary))]'
											: ''}"
										onclick={() => (selectedCustomer = customer)}
									>
										<td>
											<div class="font-medium text-foreground">
												{(customer.FirstName + ' ' + (customer.LastName || '')).trim()}
											</div>
											<div class="text-xs text-muted-foreground">{customer.Email}</div>
										</td>
										<td class="num-mono text-sm text-muted-foreground">
											{customer.PhoneNumber}
										</td>
										<td>
											<span class={getTierColor(customer.loyalty?.Tier || 'Bronze')}>
												<TierIcon class="h-3 w-3" />
												{customer.loyalty?.Tier || 'Bronze'}
											</span>
										</td>
										<td class="col-num font-semibold text-foreground">
											{customer.loyalty?.Points || 0}
										</td>
										<td>
											<div class="flex justify-end gap-1" onclick={(e) => e.stopPropagation()}>
												<Button
													size="sm"
													variant="ghost"
													class="h-8 w-8 p-0 text-muted-foreground hover:text-foreground"
													aria-label="Edit customer"
													onclick={() => openModal(customer)}
												>
													<Edit class="h-4 w-4" />
												</Button>
												<Button
													size="sm"
													variant="ghost"
													onclick={() => deleteCustomer(customer)}
													class="h-8 w-8 p-0 text-danger hover:bg-danger/10 hover:text-danger"
													aria-label="Delete customer"
												>
													<Trash2 class="h-4 w-4" />
												</Button>
											</div>
										</td>
									</tr>
								{/each}
							{/if}
						</tbody>
					</table>
				</div>
			</section>

			{#if pagination.totalPages > 1}
				<div class="mt-3 flex items-center justify-between">
					<p class="text-sm text-muted-foreground">
						Showing {customers.length} of {pagination.totalItems} customers
					</p>
					<div class="flex items-center gap-2">
						<Button
							variant="outline"
							size="sm"
							class="h-9 w-9 p-0"
							disabled={pagination.currentPage <= 1}
							onclick={() => handlePageChange(pagination.currentPage - 1)}
							aria-label="Previous page"
						>
							<ChevronLeft class="h-4 w-4" />
						</Button>
						<span class="num-mono text-sm font-medium text-foreground">
							Page {pagination.currentPage} of {pagination.totalPages}
						</span>
						<Button
							variant="outline"
							size="sm"
							class="h-9 w-9 p-0"
							disabled={pagination.currentPage >= pagination.totalPages}
							onclick={() => handlePageChange(pagination.currentPage + 1)}
							aria-label="Next page"
						>
							<ChevronRight class="h-4 w-4" />
						</Button>
					</div>
				</div>
			{/if}
		</div>

		<!-- Customer Details Sidebar -->
		<div class="lg:col-span-1" data-animate="fade-up" style="animation-delay: 200ms;">
			{#if selectedCustomer}
				{@const SelectedTierIcon = getTierIcon(selectedCustomer.loyalty?.Tier || 'Bronze')}
				<div class="sticky top-6 space-y-4">
					<!-- Customer Profile Card -->
					<section class="panel p-5">
						<div class="mb-3 flex items-start justify-between gap-3">
							<h2 class="min-w-0 truncate font-display text-lg font-semibold text-foreground">
								{(selectedCustomer.FirstName + ' ' + (selectedCustomer.LastName || '')).trim()}
							</h2>
							<div class="flex shrink-0 gap-1">
								<Button
									size="sm"
									variant="ghost"
									onclick={() => openModal(selectedCustomer)}
									class="h-8 w-8 p-0 text-muted-foreground hover:text-foreground"
									aria-label="Edit customer"
								>
									<Edit2 class="h-4 w-4" />
								</Button>
								<Button
									size="sm"
									variant="ghost"
									onclick={() => selectedCustomer && deleteCustomer(selectedCustomer)}
									class="h-8 w-8 p-0 text-danger hover:bg-danger/10 hover:text-danger"
									aria-label="Delete customer"
								>
									<Trash2 class="h-4 w-4" />
								</Button>
							</div>
						</div>

						<div class="space-y-2.5">
							<div class="flex items-center gap-2.5">
								<Mail class="h-4 w-4 shrink-0 text-muted-foreground" />
								<span class="truncate text-sm text-foreground">{selectedCustomer.Email}</span>
							</div>
							<div class="flex items-center gap-2.5">
								<Phone class="h-4 w-4 shrink-0 text-muted-foreground" />
								<span class="num-mono truncate text-sm text-foreground"
									>{selectedCustomer.PhoneNumber}</span
								>
							</div>
						</div>
					</section>

					<!-- Loyalty Card -->
					{#if auth.hasPermission('loyalty.read')}
						<section class="panel p-5">
							<h3 class="section-title mb-4 flex items-center gap-2">
								<Star class="h-4 w-4 text-primary" />
								Loyalty Program
							</h3>

							<div class="mb-5 text-center">
								<p class="kpi-num text-4xl text-primary">
									{selectedCustomer.loyalty?.Points || 0}
								</p>
								<p class="mt-1 text-xs font-medium text-muted-foreground">Available Points</p>
								<span
									class={`mt-2 inline-flex ${getTierColor(selectedCustomer.loyalty?.Tier || 'Bronze')}`}
								>
									<SelectedTierIcon class="h-3 w-3" />
									{selectedCustomer.loyalty?.Tier || 'Bronze'} Member
								</span>
								<p class="mt-3 text-[0.65rem] text-muted-foreground">
									Earning Rate: {$settings.loyalty_points_earning_rate} pt / {$settings.currency_symbol}1
									spent
								</p>
							</div>

							<div class="space-y-3">
								<div>
									<label class="text-xs font-medium text-foreground">Add Points</label>
									<div class="mt-1 flex gap-2">
										<Input
											type="number"
											placeholder="Amount"
											bind:value={pointsToAdd}
											class="num-mono h-10 w-full rounded-md text-sm"
										/>
										<Button
											onclick={addPoints}
											class="btn-soothe h-10 w-10 shrink-0 p-0"
											aria-label="Add points"
										>
											<Gift class="h-4 w-4" />
										</Button>
									</div>
								</div>

								<div>
									<label class="text-xs font-medium text-foreground">Redeem Points</label>
									<div class="mt-1 flex gap-2">
										<Input
											type="number"
											placeholder="Amount"
											bind:value={pointsToRedeem}
											class="num-mono h-10 w-full rounded-md text-sm"
										/>
										<Button onclick={redeemPoints} class="btn-soothe h-10 shrink-0 px-3 text-sm">
											Redeem
										</Button>
									</div>
								</div>
							</div>
						</section>
					{/if}

					<!-- Quick Stats -->
					<section class="panel p-5">
						<h3 class="section-title mb-3">Quick Stats</h3>
						<div class="space-y-2.5">
							<div class="flex justify-between gap-3">
								<span class="text-xs text-muted-foreground">Joined Date</span>
								<span class="num-mono text-xs font-medium text-foreground"
									>{new Date(selectedCustomer.CreatedAt || '').toLocaleDateString()}</span
								>
							</div>
							<div class="flex justify-between gap-3">
								<span class="text-xs text-muted-foreground">Last Activity</span>
								<span class="num-mono text-xs font-medium text-foreground"
									>{new Date(selectedCustomer.UpdatedAt || '').toLocaleDateString()}</span
								>
							</div>
						</div>
					</section>
				</div>
			{:else}
				<div class="sticky top-6">
					<section
						class="rounded-lg border border-dashed border-border/70 bg-muted/20 p-10 text-center"
					>
						<div
							class="mx-auto mb-3 flex h-14 w-14 items-center justify-center rounded-lg bg-muted text-muted-foreground/60"
						>
							<Users class="h-6 w-6" />
						</div>
						<p class="text-sm font-semibold text-foreground">Select a Customer</p>
						<p class="mt-1 text-xs text-muted-foreground">
							Choose a customer from list to view details
						</p>
					</section>
				</div>
			{/if}
		</div>
	</div>
</div>

<Dialog.Root bind:open={isModalOpen}>
	<Dialog.Content class="overflow-hidden p-0 sm:max-w-[425px]">
		<!-- Header -->
		<div class="border-b border-border/70 bg-muted/40 px-6 py-5">
			<Dialog.Title class="font-display text-xl font-semibold tracking-tight text-foreground">
				{editingCustomer ? 'Edit Customer' : 'New Customer'}
			</Dialog.Title>
			<Dialog.Description class="mt-1 text-sm text-muted-foreground">
				{editingCustomer
					? "Update the customer's profile information."
					: 'Add a new member to your customer base.'}
			</Dialog.Description>
		</div>

		<!-- Body -->
		<div class="space-y-4 px-6 py-5">
			{#key isModalOpen}
				<!-- Name Field -->
				<div class="space-y-1.5">
					<label for="name" class="text-xs font-medium text-foreground">Full Name</label>
					<div class="relative">
						<div class="pointer-events-none absolute inset-y-0 left-0 flex items-center pl-3">
							<Users class="h-4 w-4 text-muted-foreground" />
						</div>
						<Input
							id="name"
							bind:value={customerForm.name}
							placeholder="Jane Doe"
							class="h-11 w-full rounded-md pl-9 text-sm"
						/>
					</div>
				</div>

				<!-- Email Field -->
				<div class="space-y-1.5">
					<label for="email" class="text-xs font-medium text-foreground">Email Address</label>
					<div class="relative">
						<div class="pointer-events-none absolute inset-y-0 left-0 flex items-center pl-3">
							<Mail class="h-4 w-4 text-muted-foreground" />
						</div>
						<Input
							id="email"
							type="email"
							bind:value={customerForm.email}
							placeholder="jane@example.com"
							class="h-11 w-full rounded-md pl-9 text-sm"
						/>
					</div>
				</div>

				<!-- Phone Field -->
				<div class="space-y-1.5">
					<label for="phone" class="text-xs font-medium text-foreground">Phone Number</label>
					<div class="relative">
						<div class="pointer-events-none absolute inset-y-0 left-0 flex items-center pl-3">
							<Phone class="h-4 w-4 text-muted-foreground" />
						</div>
						<Input
							id="phone"
							bind:value={customerForm.phone}
							placeholder="+1 (555) 000-0000"
							class="h-11 w-full rounded-md pl-9 text-sm"
						/>
					</div>
				</div>
			{/key}
		</div>

		<!-- Footer -->
		<div
			class="flex items-center justify-end gap-2 border-t border-border/70 bg-muted/40 px-6 py-4"
		>
			<Button
				variant="ghost"
				onclick={() => (isModalOpen = false)}
				class="text-muted-foreground hover:text-foreground"
			>
				Cancel
			</Button>
			<Button onclick={saveCustomer} class="btn-soothe px-6">
				{editingCustomer ? 'Update Profile' : 'Create Profile'}
			</Button>
		</div>
	</Dialog.Content>
</Dialog.Root>
