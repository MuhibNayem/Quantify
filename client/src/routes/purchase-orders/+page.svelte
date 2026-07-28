<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { purchaseOrdersApi, suppliersApi } from '$lib/api/resources';
	import { auth } from '$lib/stores/auth';
	import { formatDate, formatCurrency, cn } from '$lib/utils';
	import { Button } from '$lib/components/ui/button';
	import { Badge } from '$lib/components/ui/badge';
	import { Input } from '$lib/components/ui/input';
	import Select from '$lib/components/ui/select/select.svelte';
	import {
		Table,
		TableBody,
		TableCell,
		TableHead,
		TableHeader,
		TableRow
	} from '$lib/components/ui/table';
	import {
		ShoppingCart,
		Plus,
		Search,
		ArrowRight,
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

	function getStatusVariant(status: string): 'default' | 'secondary' | 'destructive' | 'outline' {
		switch (status) {
			case 'APPROVED':
			case 'RECEIVED':
				return 'default';
			case 'SENT':
			case 'PARTIALLY_RECEIVED':
				return 'secondary';
			case 'CANCELLED':
				return 'destructive';
			default:
				return 'outline';
		}
	}

	function getStatusColor(status: string) {
		switch (status) {
			case 'APPROVED':
			case 'RECEIVED':
				return 'bg-emerald-100 text-emerald-700 border-emerald-200';
			case 'SENT':
				return 'bg-blue-100 text-blue-700 border-blue-200';
			case 'PARTIALLY_RECEIVED':
				return 'bg-amber-100 text-amber-700 border-amber-200';
			case 'CANCELLED':
				return 'bg-red-100 text-red-700 border-red-200';
			default:
				return 'bg-slate-100 text-slate-700 border-slate-200';
		}
	}

	function getPoTotal(po: PurchaseOrder): number {
		if (!po.PurchaseOrderItems) return 0;
		return po.PurchaseOrderItems.reduce((sum, item) => sum + item.UnitPrice * item.OrderedQuantity, 0);
	}

	async function handleApprove(po: PurchaseOrder) {
		actionLoading = po.ID;
		try {
			await purchaseOrdersApi.approve(po.ID);
			toast.success(`PO #${po.ID} approved`);
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
			toast.success(`PO #${po.ID} sent`);
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
			toast.success(`PO #${po.ID} received`);
			await loadPurchaseOrders();
		} catch (e: any) {
			toast.error('Failed to receive', { description: e?.response?.data?.error || e.message });
		} finally {
			actionLoading = null;
		}
	}

	async function handleCancel(po: PurchaseOrder) {
		if (!confirm(`Cancel PO #${po.ID}? This cannot be undone.`)) return;
		actionLoading = po.ID;
		try {
			await purchaseOrdersApi.cancel(po.ID);
			toast.success(`PO #${po.ID} cancelled`);
			await loadPurchaseOrders();
		} catch (e: any) {
			toast.error('Failed to cancel', { description: e?.response?.data?.error || e.message });
		} finally {
			actionLoading = null;
		}
	}
</script>

<div class="relative min-h-screen overflow-hidden bg-slate-50/50 p-6 lg:p-10">
	<div
		class="absolute inset-0 -z-10 bg-[radial-gradient(ellipse_at_top,_var(--tw-gradient-stops))] from-purple-100/20 via-slate-50/20 to-white/20"
	></div>

	<div class="mx-auto max-w-7xl space-y-8">
		<div class="flex flex-col gap-4 md:flex-row md:items-center md:justify-between">
			<div class="space-y-1">
				<h1
					class="bg-gradient-to-r from-slate-900 via-purple-900 to-slate-900 bg-clip-text text-3xl font-bold tracking-tight text-transparent"
				>
					Purchase Orders
				</h1>
				<p class="text-slate-500">Manage replenishment purchase orders</p>
			</div>
			<Button
				class="gap-2 bg-gradient-to-r from-purple-600 to-indigo-600 text-white shadow-md hover:from-purple-500 hover:to-indigo-500"
				onclick={() => goto('/intelligence')}
			>
				<Plus class="h-4 w-4" />
				Create PO
			</Button>
		</div>

		<div
			class="liquid-panel flex flex-col gap-4 rounded-2xl bg-gradient-to-br from-white/40 via-white/20 to-white/5 p-4 backdrop-blur-sm md:flex-row md:items-center"
		>
			<div class="relative max-w-xs flex-1">
				<Search class="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-slate-400" />
				<Input
					bind:value={searchQuery}
					placeholder="Search POs..."
					class="border-slate-200 bg-white pl-9 shadow-none focus:border-purple-500 focus:ring-purple-500/20"
					onkeydown={(e) => e.key === 'Enter' && applyFilters()}
				/>
			</div>
			<div class="w-48">
				<Select
					bind:value={statusFilter}
					options={statusOptions}
					placeholder="All Statuses"
				/>
			</div>
			<div class="w-48">
				<Select
					bind:value={supplierFilter}
					options={supplierOptions}
					placeholder="All Suppliers"
				/>
			</div>
			<Button variant="outline" size="sm" onclick={applyFilters} class="gap-1">
				<Search class="h-3 w-3" />
				Filter
			</Button>
		</div>

		{#if loading}
			<div class="flex h-64 items-center justify-center">
				<Loader2 class="h-8 w-8 animate-spin text-purple-500" />
			</div>
		{:else if purchaseOrders.length === 0}
			<div class="flex h-64 flex-col items-center justify-center text-slate-500">
				<ShoppingCart class="mb-4 h-12 w-12 text-slate-300" />
				<p class="font-medium">No purchase orders found</p>
			</div>
		{:else}
			<div in:fly={{ y: 20, duration: 300 }} class="overflow-hidden rounded-2xl border border-white/60 bg-white/70 shadow-lg backdrop-blur-xl">
				<Table>
					<TableHeader>
						<TableRow class="border-slate-100 bg-slate-50/50">
							<TableHead class="font-bold text-slate-600">PO #</TableHead>
							<TableHead class="font-bold text-slate-600">Supplier</TableHead>
							<TableHead class="font-bold text-slate-600">Status</TableHead>
							<TableHead class="font-bold text-slate-600">Total</TableHead>
							<TableHead class="font-bold text-slate-600">Created</TableHead>
							<TableHead class="text-right font-bold text-slate-600">Actions</TableHead>
						</TableRow>
					</TableHeader>
					<TableBody>
						{#each purchaseOrders as po (po.ID)}
							<TableRow class="border-slate-50 transition-colors hover:bg-purple-50/30">
								<TableCell class="font-mono font-semibold text-slate-800">
									#{po.ID}
								</TableCell>
								<TableCell class="text-slate-700">
									{(po as any).Supplier?.Name || `Supplier #${po.SupplierID}`}
								</TableCell>
								<TableCell>
									<span
										class={cn(
											'inline-flex items-center rounded-full border px-2.5 py-0.5 text-xs font-bold uppercase tracking-wide',
											getStatusColor(po.Status)
										)}
									>
										{po.Status.replace('_', ' ')}
									</span>
								</TableCell>
								<TableCell class="font-semibold text-slate-800">
									{formatCurrency(getPoTotal(po))}
								</TableCell>
								<TableCell class="text-slate-500">
									{formatDate(po.OrderDate)}
								</TableCell>
								<TableCell>
									<div class="flex items-center justify-end gap-1">
										{#if actionLoading === po.ID}
											<Loader2 class="h-4 w-4 animate-spin text-purple-500" />
										{:else}
											<Button
												variant="ghost"
												size="sm"
												class="h-7 w-7 p-0 text-slate-500 hover:text-purple-600"
												title="View"
												onclick={() => goto(`/purchase-orders/${po.ID}`)}
											>
												<Eye class="h-3.5 w-3.5" />
											</Button>
											{#if po.Status === 'DRAFT'}
												<Button
													variant="ghost"
													size="sm"
													class="h-7 w-7 p-0 text-slate-500 hover:text-emerald-600"
													title="Approve"
													onclick={() => handleApprove(po)}
												>
													<CheckCircle2 class="h-3.5 w-3.5" />
												</Button>
											{/if}
											{#if po.Status === 'APPROVED'}
												<Button
													variant="ghost"
													size="sm"
													class="h-7 w-7 p-0 text-slate-500 hover:text-blue-600"
													title="Send"
													onclick={() => handleSend(po)}
												>
													<Send class="h-3.5 w-3.5" />
												</Button>
											{/if}
											{#if po.Status === 'SENT' || po.Status === 'PARTIALLY_RECEIVED'}
												<Button
													variant="ghost"
													size="sm"
													class="h-7 w-7 p-0 text-slate-500 hover:text-amber-600"
													title="Receive"
													onclick={() => handleReceive(po)}
												>
													<PackageCheck class="h-3.5 w-3.5" />
												</Button>
											{/if}
											{#if !['CANCELLED', 'RECEIVED'].includes(po.Status)}
												<Button
													variant="ghost"
													size="sm"
													class="h-7 w-7 p-0 text-slate-500 hover:text-red-600"
													title="Cancel"
													onclick={() => handleCancel(po)}
												>
													<XCircle class="h-3.5 w-3.5" />
												</Button>
											{/if}
										{/if}
									</div>
								</TableCell>
							</TableRow>
						{/each}
					</TableBody>
				</Table>
			</div>

			{#if totalPages > 1}
				<div class="flex items-center justify-between">
					<p class="text-sm text-slate-500">
						Showing {purchaseOrders.length} of {total} orders
					</p>
					<div class="flex items-center gap-2">
						<Button
							variant="outline"
							size="sm"
							disabled={page <= 1}
							onclick={() => goToPage(page - 1)}
						>
							<ChevronLeft class="h-4 w-4" />
						</Button>
						<span class="text-sm font-medium text-slate-700">
							Page {page} of {totalPages}
						</span>
						<Button
							variant="outline"
							size="sm"
							disabled={page >= totalPages}
							onclick={() => goToPage(page + 1)}
						>
							<ChevronRight class="h-4 w-4" />
						</Button>
					</div>
				</div>
			{/if}
		{/if}
	</div>
</div>
