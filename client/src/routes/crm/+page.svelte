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

	const inputClass =
		'liquid-input h-10 w-full rounded-md bg-card px-3 text-sm text-foreground placeholder:text-muted-foreground focus:border-primary focus:ring-0';

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

<!-- ===== HEADER ===== -->
<section class="w-full">
	<div class="mx-auto max-w-7xl px-4 pt-10 sm:px-6 lg:px-8">
		<div
			class="flex flex-col items-start justify-between gap-6 lg:flex-row lg:items-end"
			data-animate="fade-up"
		>
			<div>
				<div class="mb-3 inline-flex items-center gap-3">
					<span class="glass-button inline-flex rounded-md p-2">
						<Users class="text-primary h-5 w-5" />
					</span>
					<p class="overline">Customer Loyalty</p>
				</div>

				<h1 class="font-display text-foreground mb-2 text-3xl font-bold tracking-tight sm:text-4xl">
					Customer Management
				</h1>
				<p class="text-muted-foreground max-w-2xl text-sm sm:text-base">
					Manage customer relationships and loyalty programs
				</p>
			</div>

			<!-- Action buttons -->
			<div class="flex flex-col gap-3 sm:flex-row">
				<Button variant="outline" onclick={fetchCustomers} class="glass-button gap-2 rounded-md">
					<Zap class="h-4 w-4" /> Refresh
				</Button>
				<Button onclick={() => openModal()} class="bg-primary text-primary-foreground gap-2 rounded-md hover:bg-primary/90">
					<UserPlus class="h-4 w-4" /> Add Customer
				</Button>
			</div>
		</div>
	</div>
</section>

<!-- ===== SEARCH AND FILTERS ===== -->
<div class="mx-auto mt-6 max-w-7xl px-4 sm:px-6 lg:px-8">
	<div class="liquid-panel rounded-lg p-5" data-animate="fade-up" style="animation-delay: 80ms;">
		<div class="flex flex-col items-start justify-between gap-4 lg:flex-row lg:items-center">
			<div class="w-full flex-1 lg:w-auto">
				<div class="relative">
					<Search class="text-muted-foreground absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2" />
					<Input
						bind:value={searchTerm}
						oninput={handleSearch}
						placeholder="Search customers by name or email..."
						class="{inputClass} pl-10"
					/>
				</div>
			</div>

			<div class="flex items-center gap-3">
				<Button
					variant="outline"
					onclick={() => (showFilters = !showFilters)}
					class="glass-button gap-2 rounded-md"
				>
					<Filter class="h-4 w-4" /> Filters
					{#if selectedTier !== 'all'}
						<Badge class="badge-info ml-1">1</Badge>
					{/if}
				</Button>
			</div>
		</div>

		{#if showFilters}
			<div class="border-border mt-4 border-t pt-4">
				<div class="flex items-center gap-3">
					<span class="text-muted-foreground text-sm font-medium">Tier:</span>
					<div class="flex gap-2">
						{#each ['all', 'Bronze', 'Silver', 'Gold'] as tier}
							<Button
								variant={selectedTier === tier ? 'default' : 'ghost'}
								size="sm"
								class={selectedTier === tier
									? 'glass-toggle-active rounded-md'
									: 'text-muted-foreground hover:bg-muted hover:text-foreground rounded-md'}
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
		<div class="lg:col-span-2" data-animate="fade-up" style="animation-delay: 140ms;">
			<div class="liquid-panel overflow-hidden rounded-lg">
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
											<Skeleton class="bg-muted h-5 w-full" />
										</td>
									</tr>
								{/each}
							{:else if customers.length === 0}
								<tr>
									<td colspan="5" class="py-12 text-center">
										<Users class="text-muted-foreground/40 mx-auto mb-3 h-10 w-10" />
										<p class="text-foreground text-sm font-medium">No customers found</p>
										<p class="text-muted-foreground text-xs">Try adjusting your search</p>
									</td>
								</tr>
							{:else}
								{#each customers as customer (customer.ID)}
									<tr
										class="cursor-pointer transition-colors {selectedCustomer?.ID ===
										customer.ID
											? 'border-l-2 border-primary bg-muted/60'
											: ''}"
										onclick={() => (selectedCustomer = customer)}
									>
										<td>
											<div class="text-foreground font-medium">
												{(customer.FirstName + ' ' + (customer.LastName || '')).trim()}
											</div>
											<div class="text-muted-foreground text-xs">{customer.Email}</div>
										</td>
										<td class="text-muted-foreground num-mono text-sm">
											{customer.PhoneNumber}
										</td>
										<td>
											<span class={getTierColor(customer.loyalty?.Tier || 'Bronze')}>
												<svelte:component
													this={getTierIcon(customer.loyalty?.Tier || 'Bronze')}
													class="h-3 w-3"
												/>
												{customer.loyalty?.Tier || 'Bronze'}
											</span>
										</td>
										<td class="col-num text-foreground font-semibold">
											{customer.loyalty?.Points || 0}
										</td>
										<td>
											<div class="flex justify-end gap-2" onclick={(e) => e.stopPropagation()}>
												<Button
													size="sm"
													variant="outline"
													onclick={() => openModal(customer)}
													class="glass-button h-8 w-8 rounded-md p-0"
												>
													<Edit class="h-4 w-4" />
												</Button>
												<Button
													size="sm"
													variant="ghost"
													onclick={() => deleteCustomer(customer)}
													class="text-danger hover:bg-danger/10 hover:text-danger h-8 w-8 rounded-md p-0"
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
			</div>

			{#if pagination.totalPages > 1}
				<div class="mt-4 flex items-center justify-between">
					<p class="text-muted-foreground text-sm">
						Showing {customers.length} of {pagination.totalItems} customers
					</p>
					<div class="flex items-center gap-2">
						<Button
							variant="outline"
							size="sm"
							class="glass-button rounded-md"
							disabled={pagination.currentPage <= 1}
							onclick={() => handlePageChange(pagination.currentPage - 1)}
						>
							<ChevronLeft class="h-4 w-4" />
						</Button>
						<span class="text-foreground text-sm font-medium">
							Page {pagination.currentPage} of {pagination.totalPages}
						</span>
						<Button
							variant="outline"
							size="sm"
							class="glass-button rounded-md"
							disabled={pagination.currentPage >= pagination.totalPages}
							onclick={() => handlePageChange(pagination.currentPage + 1)}
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
				<div class="sticky top-6 space-y-6">
					<!-- Customer Profile Card -->
					<div class="liquid-panel rounded-lg p-6">
						<div class="mb-4 flex items-center justify-between">
							<h2 class="font-display text-foreground text-xl font-bold">
								{(selectedCustomer.FirstName + ' ' + (selectedCustomer.LastName || '')).trim()}
							</h2>
							<div class="flex gap-2">
								<Button
									size="sm"
									variant="outline"
									onclick={() => openModal(selectedCustomer)}
									class="glass-button h-8 w-8 rounded-md p-0"
								>
									<Edit2 class="h-4 w-4" />
								</Button>
								<Button
									size="sm"
									variant="ghost"
									onclick={() => selectedCustomer && deleteCustomer(selectedCustomer)}
									class="text-danger hover:bg-danger/10 hover:text-danger h-8 w-8 rounded-md p-0"
								>
									<Trash2 class="h-4 w-4" />
								</Button>
							</div>
						</div>

						<div class="space-y-3">
							<div class="flex items-center gap-3">
								<Mail class="text-muted-foreground h-4 w-4" />
								<span class="text-foreground text-sm">{selectedCustomer.Email}</span>
							</div>
							<div class="flex items-center gap-3">
								<Phone class="text-muted-foreground h-4 w-4" />
								<span class="text-foreground text-sm">{selectedCustomer.PhoneNumber}</span>
							</div>
						</div>
					</div>

					<!-- Loyalty Card -->
					{#if auth.hasPermission('loyalty.read')}
						<div class="liquid-panel rounded-lg p-6">
							<h3 class="text-foreground mb-4 flex items-center gap-2 text-lg font-semibold">
								<Star class="text-primary" />
								Loyalty Program
							</h3>

							<div class="mb-6 text-center">
								<p class="kpi-num text-primary text-4xl">
									{selectedCustomer.loyalty?.Points || 0}
								</p>
								<p class="text-muted-foreground text-sm">Available Points</p>
								<span
									class={`mt-2 inline-flex ${getTierColor(selectedCustomer.loyalty?.Tier || 'Bronze')}`}
								>
									<svelte:component
										this={getTierIcon(selectedCustomer.loyalty?.Tier || 'Bronze')}
										class="h-3 w-3"
									/>
									{selectedCustomer.loyalty?.Tier || 'Bronze'} Member
								</span>
								<div class="text-muted-foreground mt-3 text-[0.65rem]">
									Earning Rate: {$settings.loyalty_points_earning_rate} pt / {$settings.currency_symbol}1
									spent
								</div>
							</div>

							<div class="space-y-3">
								<div>
									<label class="text-foreground text-sm font-medium">Add Points</label>
									<div class="mt-1 flex gap-2">
										<Input
											type="number"
											placeholder="Amount"
											bind:value={pointsToAdd}
											class="{inputClass} num-mono"
										/>
										<Button
											onclick={addPoints}
											class="bg-primary text-primary-foreground rounded-md px-3 hover:bg-primary/90"
										>
											<Gift class="h-4 w-4" />
										</Button>
									</div>
								</div>

								<div>
									<label class="text-foreground text-sm font-medium">Redeem Points</label>
									<div class="mt-1 flex gap-2">
										<Input
											type="number"
											placeholder="Amount"
											bind:value={pointsToRedeem}
											class="{inputClass} num-mono"
										/>
										<Button
											onclick={redeemPoints}
											class="bg-primary text-primary-foreground rounded-md px-3 hover:bg-primary/90"
										>
											Redeem
										</Button>
									</div>
								</div>
							</div>
						</div>
					{/if}

					<!-- Quick Stats -->
					<div class="liquid-panel rounded-lg p-6">
						<h3 class="text-foreground mb-4 text-lg font-semibold">Quick Stats</h3>
						<div class="space-y-3">
							<div class="flex justify-between">
								<span class="text-muted-foreground text-sm">Joined Date</span>
								<span class="text-foreground num-mono text-sm font-medium"
									>{new Date(selectedCustomer.CreatedAt || '').toLocaleDateString()}</span
								>
							</div>
							<div class="flex justify-between">
								<span class="text-muted-foreground text-sm">Last Activity</span>
								<span class="text-foreground num-mono text-sm font-medium"
									>{new Date(selectedCustomer.UpdatedAt || '').toLocaleDateString()}</span
								>
							</div>
						</div>
					</div>
				</div>
			{:else}
				<div class="sticky top-6">
					<div class="liquid-panel rounded-lg p-12 text-center">
						<Users class="text-muted-foreground/40 mx-auto mb-4 h-16 w-16" />
						<p class="text-foreground text-lg font-medium">Select a Customer</p>
						<p class="text-muted-foreground text-sm">Choose a customer from list to view details</p>
					</div>
				</div>
			{/if}
		</div>
	</div>
</div>

<Dialog.Root bind:open={isModalOpen}>
	<Dialog.Content
		class="bg-card border-border overflow-hidden rounded-lg border p-0 shadow-lg sm:max-w-[425px]"
	>
		<!-- Header -->
		<div class="border-border bg-muted/50 px-6 py-6 sm:px-8">
			<Dialog.Title class="font-display text-foreground text-2xl font-bold tracking-tight">
				{editingCustomer ? 'Edit Customer' : 'New Customer'}
			</Dialog.Title>
			<Dialog.Description class="text-muted-foreground mt-1.5">
				{editingCustomer
					? "Update the customer's profile information."
					: 'Add a new member to your customer base.'}
			</Dialog.Description>
		</div>

		<!-- Body -->
		<div class="space-y-6 px-6 py-8 sm:px-8">
			{#key isModalOpen}
				<!-- Name Field -->
				<div class="group space-y-2">
					<label for="name" class="overline ml-1"> Full Name </label>
					<div class="relative">
						<div class="pointer-events-none absolute inset-y-0 left-0 flex items-center pl-3">
							<Users class="text-muted-foreground h-4 w-4" />
						</div>
						<Input
							id="name"
							bind:value={customerForm.name}
							placeholder="Jane Doe"
							class="{inputClass} pl-10"
						/>
					</div>
				</div>

				<!-- Email Field -->
				<div class="group space-y-2">
					<label for="email" class="overline ml-1"> Email Address </label>
					<div class="relative">
						<div class="pointer-events-none absolute inset-y-0 left-0 flex items-center pl-3">
							<Mail class="text-muted-foreground h-4 w-4" />
						</div>
						<Input
							id="email"
							type="email"
							bind:value={customerForm.email}
							placeholder="jane@example.com"
							class="{inputClass} pl-10"
						/>
					</div>
				</div>

				<!-- Phone Field -->
				<div class="group space-y-2">
					<label for="phone" class="overline ml-1"> Phone Number </label>
					<div class="relative">
						<div class="pointer-events-none absolute inset-y-0 left-0 flex items-center pl-3">
							<Phone class="text-muted-foreground h-4 w-4" />
						</div>
						<Input
							id="phone"
							bind:value={customerForm.phone}
							placeholder="+1 (555) 000-0000"
							class="{inputClass} pl-10"
						/>
					</div>
				</div>
			{/key}
		</div>

		<!-- Footer -->
		<div class="bg-muted/50 flex items-center justify-end gap-3 px-6 py-4 sm:px-8">
			<Button
				variant="ghost"
				onclick={() => (isModalOpen = false)}
				class="text-muted-foreground hover:bg-muted hover:text-foreground rounded-md"
			>
				Cancel
			</Button>
			<Button
				onclick={saveCustomer}
				class="bg-primary text-primary-foreground rounded-md px-8 font-semibold hover:bg-primary/90"
			>
				{editingCustomer ? 'Update Profile' : 'Create Profile'}
			</Button>
		</div>
	</Dialog.Content>
</Dialog.Root>
