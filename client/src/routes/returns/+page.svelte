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

<div class="mx-auto w-full max-w-7xl px-4 py-8 sm:px-6 lg:px-8 lg:py-10">
	<!-- Page header -->
	<header class="page-header" data-animate="fade-up">
		<div>
			<p class="overline">After-sales · Returns</p>
			<h1 class="page-title mt-1">Returns Center</h1>
			<p class="page-sub">Manage your orders, request returns, and track refunds.</p>
		</div>
	</header>

	<Tabs.Root
		value={activeTab}
		class="mt-6 w-full space-y-4"
		onValueChange={(val) => (activeTab = val)}
	>
		<Tabs.List
			class="inline-flex h-11 w-full items-center gap-1 rounded-lg border border-border/70 bg-muted/60 p-1 sm:w-auto"
		>
			<Tabs.Trigger
				value="request"
				class="flex flex-1 items-center justify-center gap-2 rounded-md px-4 py-2 text-sm font-medium text-muted-foreground transition-colors data-[state=active]:bg-card data-[state=active]:text-foreground data-[state=active]:shadow-xs sm:flex-none"
			>
				<Package size={16} /> Request Return
			</Tabs.Trigger>
			<Tabs.Trigger
				value="history"
				class="flex flex-1 items-center justify-center gap-2 rounded-md px-4 py-2 text-sm font-medium text-muted-foreground transition-colors data-[state=active]:bg-card data-[state=active]:text-foreground data-[state=active]:shadow-xs sm:flex-none"
			>
				<History size={16} /> My Returns
			</Tabs.Trigger>
			{#if auth.hasPermission('returns.manage')}
				<Tabs.Trigger
					value="admin"
					class="flex flex-1 items-center justify-center gap-2 rounded-md px-4 py-2 text-sm font-medium text-muted-foreground transition-colors data-[state=active]:bg-card data-[state=active]:text-foreground data-[state=active]:shadow-xs sm:flex-none"
				>
					<ShieldCheck size={16} /> Manage Returns
				</Tabs.Trigger>
			{/if}
		</Tabs.List>

		<!-- Request Return Tab -->
		<Tabs.Content value="request" class="space-y-4 pt-1 outline-none">
			<!-- Search -->
			<section class="panel flex flex-col gap-3 p-4 sm:flex-row sm:items-center">
				<div class="relative flex-1">
					<Search class="absolute top-1/2 left-3 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
					<Input
						placeholder="Search by Order Number (e.g., ORD-12345...)"
						class="h-11 w-full rounded-md pl-9 text-sm"
						bind:value={searchQuery}
						onkeydown={(e) => e.key === 'Enter' && searchOrder()}
					/>
				</div>
				<div class="flex items-center gap-2">
					<Button class="btn-soothe shrink-0" onclick={searchOrder} disabled={isSearching}>
						{#if isSearching}
							<Loader2 class="mr-2 h-4 w-4 animate-spin" /> Searching...
						{:else}
							Find Order
						{/if}
					</Button>
					{#if searchQuery}
						<Button
							variant="ghost"
							class="shrink-0 text-muted-foreground hover:text-foreground"
							onclick={() => {
								searchQuery = '';
								loadOrders();
							}}
						>
							Clear
						</Button>
					{/if}
				</div>
			</section>

			<div in:fly={{ y: 16, duration: 300 }} class="grid gap-3">
				<!-- Orders List -->
				{#each orders as order}
					<div
						class="panel group relative overflow-hidden p-5 transition-colors duration-200 hover:border-primary/30"
					>
						<div class="flex flex-col justify-between gap-4 md:flex-row md:items-center">
							<div class="min-w-0">
								<div class="flex flex-wrap items-center gap-2.5">
									<h3 class="num-mono text-base font-bold text-foreground">{orderNumber(order)}</h3>
									<span class={order.Status === 'COMPLETED' ? 'badge-success' : 'badge-neutral'}
										>{order.Status}</span
									>
								</div>
								<p class="mt-1 text-sm text-muted-foreground">
									{#if order.Customer}
										{order.Customer.FirstName} {order.Customer.LastName} •
									{/if}
									Placed <TimeAgo date={order.OrderDate} /> •
									<span class="num">{order.Items?.length || 0}</span>
									Items
								</p>
							</div>
							<div class="flex items-center justify-between gap-5 md:justify-end">
								<div class="text-right">
									<p class="overline">Total</p>
									<p class="num-mono mt-0.5 text-lg font-bold text-foreground">
										${order.TotalAmount.toFixed(2)}
									</p>
								</div>
								<Button
									size="sm"
									class="btn-soothe shrink-0"
									onclick={() => openRequestModal(order)}
									disabled={order.Status !== 'COMPLETED'}
								>
									<Undo2 size={16} class="mr-1.5" /> Return Items
								</Button>
							</div>
						</div>
					</div>
				{:else}
					<div class="rounded-lg border border-dashed border-border/70 bg-muted/25">
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
		<Tabs.Content value="history" class="space-y-6 pt-1 outline-none">
			<div in:fly={{ y: 16, duration: 300 }} class="panel p-12">
				<p class="text-center text-sm text-muted-foreground">Return history coming soon...</p>
			</div>
		</Tabs.Content>

		<!-- Admin Tab -->
		<Tabs.Content value="admin" class="space-y-6 pt-1 outline-none">
			<div in:fly={{ y: 16, duration: 300 }}>
				<AdminReturnList />
			</div>
		</Tabs.Content>
	</Tabs.Root>
</div>

{#if selectedOrder}
	<ReturnRequestModal
		bind:open={isRequestModalOpen}
		order={selectedOrder}
		on:submit={handleReturnSubmitted}
	/>
{/if}
