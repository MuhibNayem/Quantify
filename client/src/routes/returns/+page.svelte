<script lang="ts">
	import { onMount } from 'svelte';
	import { ordersApi } from '$lib/api/orders';
	import { returnsApi } from '$lib/api/returns';
	import { auth } from '$lib/stores/auth';
	import * as Tabs from '$lib/components/ui/tabs';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Badge } from '$lib/components/ui/badge';
	import { Undo2, Package, History, ShieldCheck, Search, Filter, Loader2 } from 'lucide-svelte';
	import { fade, fly } from 'svelte/transition';
	import { cn } from '$lib/utils';
	import { toast } from 'svelte-sonner';
	import ReturnRequestModal from '$lib/components/returns/ReturnRequestModal.svelte';
	import AdminReturnList from '$lib/components/returns/AdminReturnList.svelte';

	let activeTab = 'request';
	let orders: any[] = [];
	let myReturns: any[] = [];
	let adminReturns: any[] = [];
	let selectedOrder: any = null;
	let isRequestModalOpen = false;

	let searchQuery = '';
	let isSearching = false;

	onMount(async () => {
		loadOrders();
		loadMyReturns();
		if (auth.hasPermission('returns.manage')) {
			loadAdminReturns();
		}
	});

	async function loadOrders() {
		try {
			const data = await ordersApi.listOrders();
			orders = data.orders || [];
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

<div class="relative min-h-screen overflow-hidden bg-slate-50 p-6 lg:p-10">
	<!-- Background -->
	<div class="absolute left-0 top-0 -z-10 h-full w-full overflow-hidden bg-white/50">
		<div
			class="animate-blob absolute -left-[10%] top-[20%] h-[50%] w-[50%] rounded-full bg-blue-100/60 blur-[100px]"
		></div>
		<div
			class="animate-blob animation-delay-2000 absolute -right-[10%] -top-[10%] h-[60%] w-[60%] rounded-full bg-purple-100/60 blur-[100px]"
		></div>
	</div>

	<div class="relative z-10 mx-auto max-w-7xl space-y-8">
		<!-- Header -->
		<div class="flex flex-col gap-2 backdrop-blur-sm">
			<h1
				class="bg-gradient-to-r from-blue-600 via-purple-600 to-pink-600 bg-clip-text text-4xl font-bold tracking-tight text-transparent drop-shadow-sm"
			>
				Returns Center
			</h1>
			<p class="font-medium text-slate-500">
				Manage your orders, request returns, and track refunds.
			</p>
		</div>

		<Tabs.Root
			value={activeTab}
			class="w-full space-y-8"
			onValueChange={(val) => (activeTab = val)}
		>
			<Tabs.List
				class="inline-flex h-auto w-full rounded-2xl border border-white/60 bg-white/40 p-1.5 shadow-lg backdrop-blur-xl"
			>
				<Tabs.Trigger
					value="request"
					class="flex-1 rounded-xl py-3 text-sm font-medium transition-all duration-300 data-[state=active]:bg-white data-[state=active]:text-blue-600 data-[state=active]:shadow-md"
				>
					<div class="flex items-center justify-center gap-2">
						<Package size={18} /> Request Return
					</div>
				</Tabs.Trigger>
				<Tabs.Trigger
					value="history"
					class="flex-1 rounded-xl py-3 text-sm font-medium transition-all duration-300 data-[state=active]:bg-white data-[state=active]:text-blue-600 data-[state=active]:shadow-md"
				>
					<div class="flex items-center justify-center gap-2">
						<History size={18} /> My Returns
					</div>
				</Tabs.Trigger>
				{#if auth.hasPermission('returns.manage')}
					<Tabs.Trigger
						value="admin"
						class="flex-1 rounded-xl py-3 text-sm font-medium transition-all duration-300 data-[state=active]:bg-white data-[state=active]:text-purple-600 data-[state=active]:shadow-md"
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
				<div class="flex items-center gap-4 rounded-2xl bg-white/60 p-4 shadow-sm backdrop-blur-md">
					<div class="relative flex-1">
						<Search class="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-slate-400" />
						<Input
							placeholder="Search by Order Number (e.g., ORD-12345...)"
							class="border-slate-200 bg-white/50 pl-9 transition-all focus:bg-white"
							bind:value={searchQuery}
							onkeydown={(e) => e.key === 'Enter' && searchOrder()}
						/>
					</div>
					<Button
						variant="default"
						class="bg-blue-600 shadow-md transition-all hover:bg-blue-700 hover:shadow-lg"
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
							class="text-slate-500 hover:text-slate-700"
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
							class="group relative overflow-hidden rounded-3xl border border-white/60 bg-white/60 p-6 shadow-xl backdrop-blur-2xl transition-all hover:bg-white/80"
						>
							<div class="flex flex-col justify-between gap-4 md:flex-row md:items-center">
								<div>
									<div class="flex items-center gap-3">
										<h3 class="text-lg font-bold text-slate-800">{order.OrderNumber}</h3>
										<Badge variant={order.Status === 'COMPLETED' ? 'default' : 'secondary'}
											>{order.Status}</Badge
										>
									</div>
									<p class="text-sm text-slate-500">
										Placed on {new Date(order.OrderDate).toLocaleDateString()} • {order.Items
											?.length || 0} Items
									</p>
								</div>
								<div class="flex items-center gap-4">
									<div class="text-right">
										<p class="text-sm font-medium text-slate-500">Total</p>
										<p class="text-xl font-bold text-slate-800">${order.TotalAmount.toFixed(2)}</p>
									</div>
									<Button
										class="rounded-xl bg-blue-600 text-white shadow-lg hover:bg-blue-700"
										onclick={() => openRequestModal(order)}
										disabled={order.Status !== 'COMPLETED'}
									>
										<Undo2 size={18} class="mr-2" /> Return Items
									</Button>
								</div>
							</div>
						</div>
					{:else}
						<div
							class="flex h-64 flex-col items-center justify-center rounded-3xl border border-dashed border-slate-300 bg-white/40 text-center"
						>
							<Package size={48} class="mb-4 text-slate-300" />
							<p class="text-lg font-medium text-slate-600">No orders found</p>
							<p class="text-slate-400">Go make some purchases!</p>
						</div>
					{/each}
				</div>
			</Tabs.Content>

			<!-- My Returns Tab -->
			<Tabs.Content value="history" class="space-y-6 pt-2 outline-none">
				<div in:fly={{ y: 20, duration: 300 }}>
					<p class="text-center text-slate-500">Return history coming soon...</p>
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

<style>
	@keyframes blob {
		0%,
		100% {
			transform: translate(0, 0) scale(1);
		}
		33% {
			transform: translate(30px, -50px) scale(1.1);
		}
		66% {
			transform: translate(-20px, 20px) scale(0.9);
		}
	}
	.animate-blob {
		animation: blob 15s infinite;
	}
	.animation-delay-2000 {
		animation-delay: 2s;
	}
</style>
