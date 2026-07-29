<script lang="ts">
	import { onMount } from 'svelte';
	import { ordersApi } from '$lib/api/orders';
	import { returnsApi } from '$lib/api/returns';
	import { auth } from '$lib/stores/auth';
	import * as Tabs from '$lib/components/ui/tabs';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Undo2, Package, History, ShieldCheck, Search, Filter, Loader2 } from 'lucide-svelte';
	import { fade, fly } from 'svelte/transition';
	import { cn, orderNumber } from '$lib/utils';
	import { toast } from 'svelte-sonner';
	import TimeAgo from '$lib/components/ui/time-ago.svelte';
	import EmptyState from '$lib/components/ui/empty-state.svelte';
	import ReturnRequestModal from '$lib/components/returns/ReturnRequestModal.svelte';
	import { goto } from '$app/navigation';
	import AdminReturnList from '$lib/components/returns/AdminReturnList.svelte';

	let activeTab = 'request';
	let orders: any[] = [];
	let myReturns: any[] = [];
	let adminReturns: any[] = [];
	let selectedOrder: any = null;
	let isRequestModalOpen = false;

	let searchQuery = '';
	let isSearching = false;

	const inputClass =
		'liquid-input h-10 w-full rounded-md bg-card bg-none px-3 text-sm font-normal text-foreground shadow-none ring-0 backdrop-blur-none backdrop-saturate-100 from-card to-card placeholder:text-muted-foreground focus:scale-100 focus:border-primary focus:ring-0 md:text-sm';
	const btnPrimary =
		'rounded-md bg-primary bg-none px-4 text-sm font-medium text-primary-foreground shadow-none hover:bg-primary/90 hover:bg-none';
	const tabTriggerClass =
		'data-[state=active]:bg-card data-[state=active]:text-foreground data-[state=active]:shadow-sm rounded-md';

	$effect(() => {
		if (!auth.hasPermission('returns.manage') && !auth.hasPermission('returns.request')) {
			toast.error('Access Denied', {
				description: 'You do not have permission to access returns.'
			});
			goto('/');
		}
	});

	onMount(async () => {
		if (!auth.hasPermission('returns.manage') && !auth.hasPermission('returns.request')) return;

		loadOrders();
		loadMyReturns();
		if (auth.hasPermission('returns.manage')) {
			loadAdminReturns();
		}
	});

	async function loadOrders() {
		try {
			const data = await ordersApi.listOrders();
			orders = data.data || data.items || (data as any).orders || [];
		} catch (e) {
			console.error(e);
			toast.error('Failed to load orders');
		}
	}

	async function searchOrder() {
		if (!searchQuery.trim()) {
			loadOrders();
			return;
		}

		isSearching = true;
		try {
			const data = await ordersApi.getOrder(searchQuery.trim());
			// API returns { order: ... }
			if (data && data.order) {
				orders = [data.order];
				toast.success('Order found');
			} else {
				orders = [];
				toast.warning('Order not found');
			}
		} catch (e) {
			console.error(e);
			orders = []; // Clear list if not found
			toast.error('Order not found or access denied');
		} finally {
			isSearching = false;
		}
	}

	async function loadMyReturns() {
		// TODO: Implement listMyReturns in API if needed, or filter listReturns
		// For now, let's assume listReturns returns all for admin, but maybe we need a "mine" filter?
		// The current API `listReturns` calls `GET /returns`.
		// Backend `ListReturns` usually returns all.
		// I might need to update backend to filter by user if not admin.
		// For now, I'll skip this or just show empty.
	}

	async function loadAdminReturns() {
		try {
			const data = await returnsApi.listReturns('PENDING'); // Default to pending
			adminReturns = data.returns || [];
		} catch (e) {
			console.error(e);
			toast.error('Failed to load returns');
		}
	}

	function openRequestModal(order: any) {
		selectedOrder = order;
		isRequestModalOpen = true;
	}

	function handleReturnSubmitted() {
		isRequestModalOpen = false;
		if (searchQuery) {
			searchOrder(); // Refresh search result
		} else {
			loadOrders(); // Refresh list
		}
		loadMyReturns();
		toast.success('Return requested successfully');
	}
</script>

<div class="relative min-h-screen overflow-hidden p-6 lg:p-10">
	<div class="relative z-10 mx-auto max-w-7xl space-y-8">
		<!-- Header -->
		<div class="flex flex-col gap-2" data-animate="fade-up">
			<h1
				class="font-display text-foreground text-3xl font-bold tracking-tight"
			>
				Returns Center
			</h1>
			<p class="text-muted-foreground font-medium">
				Manage your orders, request returns, and track refunds.
			</p>
		</div>

		<Tabs.Root
			value={activeTab}
			class="w-full space-y-8"
			onValueChange={(val) => (activeTab = val)}
		>
			<Tabs.List
				class="border-border bg-muted inline-flex h-auto w-full rounded-lg border p-1.5 shadow-none backdrop-blur-none"
			>
				<Tabs.Trigger
					value="request"
					class="flex-1 py-3 text-sm font-medium transition-all duration-300 {tabTriggerClass}"
				>
					<div class="flex items-center justify-center gap-2">
						<Package size={18} /> Request Return
					</div>
				</Tabs.Trigger>
				<Tabs.Trigger
					value="history"
					class="flex-1 py-3 text-sm font-medium transition-all duration-300 {tabTriggerClass}"
				>
					<div class="flex items-center justify-center gap-2">
						<History size={18} /> My Returns
					</div>
				</Tabs.Trigger>
				{#if auth.hasPermission('returns.manage')}
					<Tabs.Trigger
						value="admin"
						class="flex-1 py-3 text-sm font-medium transition-all duration-300 {tabTriggerClass}"
					>
						<div class="flex items-center justify-center gap-2">
							<ShieldCheck size={18} /> Manage Returns
						</div>
					</Tabs.Trigger>
				{/if}
			</Tabs.List>

			<!-- Request Return Tab -->
			<Tabs.Content value="request" class="space-y-6 pt-2 outline-none">
				<!-- Search -->
				<div class="liquid-panel flex items-center gap-4 rounded-lg p-4">
					<div class="relative flex-1">
						<Search class="text-muted-foreground absolute top-1/2 left-3 h-4 w-4 -translate-y-1/2" />
						<Input
							placeholder="Search by Order Number (e.g., ORD-12345...)"
							class="{inputClass} pl-9"
							bind:value={searchQuery}
							onkeydown={(e) => e.key === 'Enter' && searchOrder()}
						/>
					</div>
					<Button
						variant="default"
						class={btnPrimary}
						onclick={searchOrder}
						disabled={isSearching}
					>
						{#if isSearching}
							<Loader2 class="mr-2 h-4 w-4 animate-spin" /> Searching...
						{:else}
							Find Order
						{/if}
					</Button>
					{#if searchQuery}
						<Button
							variant="ghost"
							class="text-muted-foreground hover:text-foreground"
							onclick={() => {
								searchQuery = '';
								loadOrders();
							}}
						>
							Clear
						</Button>
					{/if}
				</div>

				<div in:fly={{ y: 20, duration: 300 }} class="grid gap-6">
					<!-- Orders List -->
					{#each orders as order}
						<div
							class="liquid-panel group relative overflow-hidden rounded-lg p-6 transition-all hover:border-primary/35 hover:shadow-md"
						>
							<div class="flex flex-col justify-between gap-4 md:flex-row md:items-center">
								<div>
									<div class="flex items-center gap-3">
										<h3 class="num-mono text-foreground text-lg font-bold">{orderNumber(order)}</h3>
										<span class={order.Status === 'COMPLETED' ? 'badge-success' : 'badge-neutral'}
											>{order.Status}</span
										>
									</div>
									<p class="text-muted-foreground text-sm">
										{#if order.Customer}
											{order.Customer.FirstName} {order.Customer.LastName} •
										{/if}
										Placed <TimeAgo date={order.OrderDate} /> • <span class="num">{order.Items?.length || 0}</span> Items
									</p>
								</div>
								<div class="flex items-center gap-4">
									<div class="text-right">
										<p class="overline">Total</p>
										<p class="num-mono text-foreground text-xl font-bold">${order.TotalAmount.toFixed(2)}</p>
									</div>
									<Button
										class="bg-primary text-primary-foreground hover:bg-primary/90 rounded-md bg-none shadow-sm hover:bg-none"
										onclick={() => openRequestModal(order)}
										disabled={order.Status !== 'COMPLETED'}
									>
										<Undo2 size={18} class="mr-2" /> Return Items
									</Button>
								</div>
							</div>
						</div>
					{:else}
						<div class="border-border bg-muted/40 rounded-lg border border-dashed">
							<EmptyState
								icon={Package}
								title="No orders found"
								description="Go make some purchases, then come back to request a return."
							/>
						</div>
					{/each}
				</div>
			</Tabs.Content>

			<!-- My Returns Tab -->
			<Tabs.Content value="history" class="space-y-6 pt-2 outline-none">
				<div in:fly={{ y: 20, duration: 300 }}>
					<p class="text-muted-foreground text-center">Return history coming soon...</p>
				</div>
			</Tabs.Content>

			<!-- Admin Tab -->
			<Tabs.Content value="admin" class="space-y-6 pt-2 outline-none">
				<div in:fly={{ y: 20, duration: 300 }}>
					<AdminReturnList />
				</div>
			</Tabs.Content>
		</Tabs.Root>
	</div>
</div>

{#if selectedOrder}
	<ReturnRequestModal
		bind:open={isRequestModalOpen}
		order={selectedOrder}
		on:submit={handleReturnSubmitted}
	/>
{/if}
