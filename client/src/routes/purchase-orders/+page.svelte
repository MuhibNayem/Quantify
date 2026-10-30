<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { purchaseOrdersApi, suppliersApi } from '$lib/api/resources';
	import { auth } from '$lib/stores/auth';
	import { formatCurrency, cn, poNumber } from '$lib/utils';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import TimeAgo from '$lib/components/ui/time-ago.svelte';
	import EmptyState from '$lib/components/ui/empty-state.svelte';
	import Select from '$lib/components/ui/select/select.svelte';
	import {
		ShoppingCart,
		Plus,
		Search,
		CheckCircle2,
		Send,
		PackageCheck,
		XCircle,
		Eye,
		ChevronLeft,
		ChevronRight,
		Loader2
	} from 'lucide-svelte';
	import { toast } from 'svelte-sonner';
	import { fade, fly } from 'svelte/transition';
	import type { PurchaseOrder, Supplier } from '$lib/types';

	let purchaseOrders = $state<PurchaseOrder[]>([]);
	let suppliers = $state<Supplier[]>([]);
	let loading = $state(true);
	let total = $state(0);
	let page = $state(1);
	let limit = $state(20);
	let statusFilter = $state('');
	let supplierFilter = $state('');
	let searchQuery = $state('');
	let actionLoading = $state<number | null>(null);

	const totalPages = $derived(Math.ceil(total / limit) || 1);

	const statusOptions = [
		{ value: '', label: 'All Statuses' },
		{ value: 'DRAFT', label: 'Draft' },
		{ value: 'APPROVED', label: 'Approved' },
		{ value: 'SENT', label: 'Sent' },
		{ value: 'PARTIALLY_RECEIVED', label: 'Partially Received' },
		{ value: 'RECEIVED', label: 'Received' },
		{ value: 'CANCELLED', label: 'Cancelled' }
	];

	const supplierOptions = $derived([
		{ value: '', label: 'All Suppliers' },
		...suppliers.map((s) => ({ value: String(s.ID), label: s.Name }))
	]);

	$effect(() => {
		if (!auth.hasPermission('replenishment.read') && !auth.hasPermission('inventory.view')) {
			toast.error('Access Denied', {
				description: 'You do not have permission to view purchase orders.'
			});
			goto('/');
		}
	});

	$effect(() => {
		statusFilter;
		supplierFilter;
		if (!loading) {
			page = 1;
			loadPurchaseOrders();
		}
	});

	onMount(async () => {
		await loadSuppliers();
		await loadPurchaseOrders();
	});

	async function loadSuppliers() {
		try {
			suppliers = await suppliersApi.list();
		} catch {
			suppliers = [];
		}
	}

	async function loadPurchaseOrders() {
		loading = true;
		try {
			const params: Record<string, unknown> = { page, limit };
			if (statusFilter) params.status = statusFilter;
			if (supplierFilter) params.supplierId = Number(supplierFilter);
			if (searchQuery) params.search = searchQuery;

			const res = await purchaseOrdersApi.list(params);
			purchaseOrders = res.data || res.items || [];
			total = res.total || 0;
		} catch (e: any) {
			console.error(e);
			toast.error('Failed to load purchase orders', {
				description: e?.response?.data?.error || e.message
			});
		} finally {
			loading = false;
		}
	}

	function applyFilters() {
		page = 1;
		loadPurchaseOrders();
	}

	function goToPage(p: number) {
		if (p < 1 || p > totalPages) return;
		page = p;
		loadPurchaseOrders();
	}

	function getStatusColor(status: string) {
		switch (status) {
			case 'APPROVED':
			case 'SENT':
				return 'badge-info';
			case 'PARTIALLY_RECEIVED':
				return 'badge-warning';
			case 'RECEIVED':
				return 'badge-success';
			case 'CANCELLED':
				return 'badge-danger';
			default:
				return 'badge-neutral';
		}
	}

	function getPoTotal(po: PurchaseOrder): number {
		if (!po.PurchaseOrderItems) return 0;
		return po.PurchaseOrderItems.reduce(
			(sum, item) => sum + item.UnitPrice * item.OrderedQuantity,
			0
		);
	}

	async function handleApprove(po: PurchaseOrder) {
		actionLoading = po.ID;
		try {
			await purchaseOrdersApi.approve(po.ID);
			toast.success(`${poNumber(po.ID)} approved`);
			await loadPurchaseOrders();
		} catch (e: any) {
			toast.error('Failed to approve', { description: e?.response?.data?.error || e.message });
		} finally {
			actionLoading = null;
		}
	}

	async function handleSend(po: PurchaseOrder) {
		actionLoading = po.ID;
		try {
			await purchaseOrdersApi.send(po.ID);
			toast.success(`${poNumber(po.ID)} sent`);
			await loadPurchaseOrders();
		} catch (e: any) {
			toast.error('Failed to send', { description: e?.response?.data?.error || e.message });
		} finally {
			actionLoading = null;
		}
	}

	async function handleReceive(po: PurchaseOrder) {
		actionLoading = po.ID;
		try {
			const items = (po.PurchaseOrderItems || []).map((item) => ({
				productId: item.ProductID,
				quantity: item.OrderedQuantity - item.ReceivedQuantity
			}));
			await purchaseOrdersApi.receive(po.ID, { items });
			toast.success(`${poNumber(po.ID)} received`);
			await loadPurchaseOrders();
		} catch (e: any) {
			toast.error('Failed to receive', { description: e?.response?.data?.error || e.message });
		} finally {
			actionLoading = null;
		}
	}

	async function handleCancel(po: PurchaseOrder) {
		if (!confirm(`Cancel ${poNumber(po.ID)}? This cannot be undone.`)) return;
		actionLoading = po.ID;
		try {
			await purchaseOrdersApi.cancel(po.ID);
			toast.success(`${poNumber(po.ID)} cancelled`);
			await loadPurchaseOrders();
		} catch (e: any) {
			toast.error('Failed to cancel', { description: e?.response?.data?.error || e.message });
		} finally {
			actionLoading = null;
		}
	}
</script>

<div class="mx-auto w-full max-w-7xl px-4 py-8 sm:px-6 lg:px-8 lg:py-10">
	<!-- Page header -->
	<header class="page-header" data-animate="fade-up">
		<div>
			<p class="overline">Procurement · Replenishment</p>
			<h1 class="page-title mt-1">Purchase Orders</h1>
			<p class="page-sub">Manage replenishment purchase orders</p>
		</div>
		<Button class="btn-soothe" onclick={() => goto('/intelligence')}>
			<Plus class="h-4 w-4" />
			Create PO
		</Button>
	</header>

	<!-- Filters -->
	<section
		class="panel mt-6 flex flex-col gap-3 p-4 md:flex-row md:items-center"
		data-animate="fade-up"
		style="animation-delay: 80ms;"
	>
		<div class="relative w-full max-w-xs flex-1">
			<Search class="absolute top-1/2 left-3 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
			<Input
				bind:value={searchQuery}
				placeholder="Search POs..."
				class="h-11 w-full rounded-md pl-9 text-sm"
				onkeydown={(e) => e.key === 'Enter' && applyFilters()}
			/>
		</div>
		<div class="w-full md:w-48">
			<Select bind:value={statusFilter} options={statusOptions} placeholder="All Statuses" />
		</div>
		<div class="w-full md:w-48">
			<Select bind:value={supplierFilter} options={supplierOptions} placeholder="All Suppliers" />
		</div>
		<Button variant="outline" size="sm" onclick={applyFilters} class="h-9 shrink-0 gap-1.5">
			<Search class="h-3.5 w-3.5" />
			Filter
		</Button>
	</section>

	<div class="mt-4">
		{#if loading}
			<div class="panel flex h-64 items-center justify-center">
				<Loader2 class="h-7 w-7 animate-spin text-primary" />
			</div>
		{:else if purchaseOrders.length === 0}
			<div class="rounded-lg border border-dashed border-border/70 bg-muted/25">
				<EmptyState
					icon={ShoppingCart}
					title="No purchase orders found"
					description="Create a purchase order to replenish your stock."
					actionLabel="Create PO"
					onAction={() => goto('/intelligence')}
				/>
			</div>
		{:else}
			<section in:fly={{ y: 16, duration: 300 }} class="panel overflow-hidden">
				<div class="overflow-x-auto">
					<table class="ledger-table">
						<thead>
							<tr>
								<th>PO #</th>
								<th>Supplier</th>
								<th>Status</th>
								<th class="text-right">Total</th>
								<th>Created</th>
								<th class="text-right">Actions</th>
							</tr>
						</thead>
						<tbody>
							{#each purchaseOrders as po (po.ID)}
								<tr>
									<td class="num-mono font-semibold text-foreground">
										{poNumber(po.ID)}
									</td>
									<td class="text-foreground">
										{(po as any).Supplier?.Name || `Supplier #${po.SupplierID}`}
									</td>
									<td>
										<span class={cn('uppercase', getStatusColor(po.Status))}>
											{po.Status.replace('_', ' ')}
										</span>
									</td>
									<td class="col-num font-semibold text-foreground">
										{formatCurrency(getPoTotal(po))}
									</td>
									<td class="text-sm text-muted-foreground">
										<TimeAgo date={po.OrderDate} />
									</td>
									<td>
										<div class="flex items-center justify-end gap-0.5">
											{#if actionLoading === po.ID}
												<Loader2 class="h-4 w-4 animate-spin text-primary" />
											{:else}
												<Button
													variant="ghost"
													size="sm"
													class="h-8 w-8 p-0 text-muted-foreground hover:text-foreground"
													title="View"
													aria-label="View purchase order"
													onclick={() => goto(`/purchase-orders/${po.ID}`)}
												>
													<Eye class="h-3.5 w-3.5" />
												</Button>
												{#if po.Status === 'DRAFT'}
													<Button
														variant="ghost"
														size="sm"
														class="h-8 w-8 p-0 text-muted-foreground hover:text-success"
														title="Approve"
														aria-label="Approve purchase order"
														onclick={() => handleApprove(po)}
													>
														<CheckCircle2 class="h-3.5 w-3.5" />
													</Button>
												{/if}
												{#if po.Status === 'APPROVED'}
													<Button
														variant="ghost"
														size="sm"
														class="h-8 w-8 p-0 text-muted-foreground hover:text-info"
														title="Send"
														aria-label="Send purchase order"
														onclick={() => handleSend(po)}
													>
														<Send class="h-3.5 w-3.5" />
													</Button>
												{/if}
												{#if po.Status === 'SENT' || po.Status === 'PARTIALLY_RECEIVED'}
													<Button
														variant="ghost"
														size="sm"
														class="h-8 w-8 p-0 text-muted-foreground hover:text-warning"
														title="Receive"
														aria-label="Receive purchase order"
														onclick={() => handleReceive(po)}
													>
														<PackageCheck class="h-3.5 w-3.5" />
													</Button>
												{/if}
												{#if !['CANCELLED', 'RECEIVED'].includes(po.Status)}
													<Button
														variant="ghost"
														size="sm"
														class="h-8 w-8 p-0 text-muted-foreground hover:text-danger"
														title="Cancel"
														aria-label="Cancel purchase order"
														onclick={() => handleCancel(po)}
													>
														<XCircle class="h-3.5 w-3.5" />
													</Button>
												{/if}
											{/if}
										</div>
									</td>
								</tr>
							{/each}
						</tbody>
					</table>
				</div>
			</section>

			{#if totalPages > 1}
				<div class="mt-3 flex items-center justify-between">
					<p class="text-sm text-muted-foreground">
						Showing {purchaseOrders.length} of {total} orders
					</p>
					<div class="flex items-center gap-2">
						<Button
							variant="outline"
							size="sm"
							class="h-9 w-9 p-0"
							disabled={page <= 1}
							onclick={() => goToPage(page - 1)}
							aria-label="Previous page"
						>
							<ChevronLeft class="h-4 w-4" />
						</Button>
						<span class="num-mono text-sm font-medium text-foreground">
							Page {page} of {totalPages}
						</span>
						<Button
							variant="outline"
							size="sm"
							class="h-9 w-9 p-0"
							disabled={page >= totalPages}
							onclick={() => goToPage(page + 1)}
							aria-label="Next page"
						>
							<ChevronRight class="h-4 w-4" />
						</Button>
					</div>
				</div>
			{/if}
		{/if}
	</div>
</div>
