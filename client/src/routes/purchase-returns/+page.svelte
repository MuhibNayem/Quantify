<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { purchaseReturnsApi, purchaseOrdersApi, suppliersApi } from '$lib/api/resources';
	import { auth } from '$lib/stores/auth';
	import { formatCurrency, cn, returnNumber, poNumber } from '$lib/utils';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import TimeAgo from '$lib/components/ui/time-ago.svelte';
	import EmptyState from '$lib/components/ui/empty-state.svelte';
	import Select from '$lib/components/ui/select/select.svelte';
	import * as Dialog from '$lib/components/ui/dialog';
	import { Label } from '$lib/components/ui/label';
	import {
		Undo2,
		Plus,
		Search,
		CheckCircle2,
		PackageCheck,
		XCircle,
		ChevronLeft,
		ChevronRight,
		Loader2
	} from 'lucide-svelte';
	import { toast } from 'svelte-sonner';
	import { fly } from 'svelte/transition';
	import type { PurchaseReturn, PurchaseOrder, Supplier } from '$lib/types';

	let returns = $state<PurchaseReturn[]>([]);
	let suppliers = $state<Supplier[]>([]);
	let purchaseOrders = $state<PurchaseOrder[]>([]);
	let loading = $state(true);
	let total = $state(0);
	let page = $state(1);
	let limit = $state(20);
	let statusFilter = $state('');
	let supplierFilter = $state('');
	let actionLoading = $state<number | null>(null);

	let isCreateModalOpen = $state(false);
	let createForm = $state({
		purchaseOrderId: '' as string,
		reason: '',
		items: [] as { productId: number; quantity: number; unitPrice: number }[]
	});
	let creating = $state(false);

	const totalPages = $derived(Math.ceil(total / limit) || 1);

	const statusOptions = [
		{ value: '', label: 'All Statuses' },
		{ value: 'PENDING', label: 'Pending' },
		{ value: 'APPROVED', label: 'Approved' },
		{ value: 'COMPLETED', label: 'Completed' },
		{ value: 'CANCELLED', label: 'Cancelled' }
	];

	const supplierOptions = $derived([
		{ value: '', label: 'All Suppliers' },
		...suppliers.map((s) => ({ value: String(s.ID), label: s.Name }))
	]);

	$effect(() => {
		if (!auth.hasPermission('inventory.view')) {
			toast.error('Access Denied', {
				description: 'You do not have permission to view purchase returns.'
			});
			goto('/');
		}
	});

	onMount(async () => {
		await Promise.all([loadSuppliers(), loadReturns()]);
	});

	async function loadSuppliers() {
		try {
			suppliers = await suppliersApi.list();
		} catch {
			suppliers = [];
		}
	}

	async function loadReturns() {
		loading = true;
		try {
			const params: Record<string, unknown> = { page, limit };
			if (statusFilter) params.status = statusFilter;
			if (supplierFilter) params.supplierId = Number(supplierFilter);

			const res = await purchaseReturnsApi.list(params);
			returns = res.data || res.items || [];
			total = res.total || 0;
		} catch (e: any) {
			console.error(e);
			toast.error('Failed to load purchase returns', {
				description: e?.response?.data?.error || e.message
			});
		} finally {
			loading = false;
		}
	}

	function applyFilters() {
		page = 1;
		loadReturns();
	}

	function goToPage(p: number) {
		if (p < 1 || p > totalPages) return;
		page = p;
		loadReturns();
	}

	function getStatusColor(status: string) {
		switch (status) {
			case 'APPROVED':
				return 'badge-info';
			case 'COMPLETED':
				return 'badge-success';
			case 'PENDING':
				return 'badge-warning';
			case 'CANCELLED':
				return 'badge-danger';
			default:
				return 'badge-neutral';
		}
	}

	async function handleApprove(ret: PurchaseReturn) {
		actionLoading = ret.ID;
		try {
			await purchaseReturnsApi.approve(ret.ID);
			toast.success(`${returnNumber(ret.ID)} approved`);
			await loadReturns();
		} catch (e: any) {
			toast.error('Failed to approve', { description: e?.response?.data?.error || e.message });
		} finally {
			actionLoading = null;
		}
	}

	async function handleComplete(ret: PurchaseReturn) {
		actionLoading = ret.ID;
		try {
			await purchaseReturnsApi.complete(ret.ID);
			toast.success(`${returnNumber(ret.ID)} completed`);
			await loadReturns();
		} catch (e: any) {
			toast.error('Failed to complete', { description: e?.response?.data?.error || e.message });
		} finally {
			actionLoading = null;
		}
	}

	async function handleCancel(ret: PurchaseReturn) {
		if (!confirm(`Cancel ${returnNumber(ret.ID)}? This cannot be undone.`)) return;
		actionLoading = ret.ID;
		try {
			await purchaseReturnsApi.cancel(ret.ID);
			toast.success(`${returnNumber(ret.ID)} cancelled`);
			await loadReturns();
		} catch (e: any) {
			toast.error('Failed to cancel', { description: e?.response?.data?.error || e.message });
		} finally {
			actionLoading = null;
		}
	}

	async function openCreateModal() {
		isCreateModalOpen = true;
		createForm = { purchaseOrderId: '', reason: '', items: [] };
		try {
			const res = await purchaseOrdersApi.list({ limit: 100, status: 'RECEIVED' });
			purchaseOrders = res.data || res.items || [];
		} catch {
			purchaseOrders = [];
		}
	}

	function onPoSelect() {
		const po = purchaseOrders.find((p) => String(p.ID) === createForm.purchaseOrderId);
		if (po && po.PurchaseOrderItems) {
			createForm.items = po.PurchaseOrderItems.map((item) => ({
				productId: item.ProductID,
				quantity: item.ReceivedQuantity || item.OrderedQuantity,
				unitPrice: item.UnitPrice
			}));
		} else {
			createForm.items = [];
		}
	}

	async function handleCreate() {
		if (!createForm.purchaseOrderId) {
			toast.error('Please select a purchase order');
			return;
		}
		if (!createForm.reason.trim()) {
			toast.error('Please provide a reason');
			return;
		}
		if (createForm.items.length === 0) {
			toast.error('No items to return');
			return;
		}

		creating = true;
		try {
			await purchaseReturnsApi.create({
				purchaseOrderId: Number(createForm.purchaseOrderId),
				reason: createForm.reason,
				items: createForm.items.map((i) => ({
					productId: i.productId,
					quantity: i.quantity,
					unitPrice: i.unitPrice
				}))
			});
			toast.success('Purchase return created');
			isCreateModalOpen = false;
			await loadReturns();
		} catch (e: any) {
			toast.error('Failed to create return', {
				description: e?.response?.data?.error || e.message
			});
		} finally {
			creating = false;
		}
	}
</script>

<div class="mx-auto w-full max-w-7xl px-4 py-8 sm:px-6 lg:px-8 lg:py-10">
	<!-- Page header -->
	<header class="page-header" data-animate="fade-up">
		<div>
			<p class="overline">Procurement · Vendor Returns</p>
			<h1 class="page-title mt-1">Purchase Returns</h1>
			<p class="page-sub">Manage supplier returns and refunds</p>
		</div>
		<Button class="btn-soothe" onclick={openCreateModal}>
			<Plus class="h-4 w-4" />
			New Return
		</Button>
	</header>

	<!-- Filters -->
	<section
		class="panel mt-6 flex flex-col gap-3 p-4 md:flex-row md:items-center"
		data-animate="fade-up"
		style="animation-delay: 80ms;"
	>
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
		{:else if returns.length === 0}
			<div class="rounded-lg border border-dashed border-border/70 bg-muted/25">
				<EmptyState
					icon={Undo2}
					title="No purchase returns found"
					description="Return damaged or unwanted stock to suppliers."
					actionLabel="New Return"
					onAction={openCreateModal}
				/>
			</div>
		{:else}
			<section in:fly={{ y: 16, duration: 300 }} class="panel overflow-hidden">
				<div class="overflow-x-auto">
					<table class="ledger-table">
						<thead>
							<tr>
								<th>Return #</th>
								<th>PO #</th>
								<th>Supplier</th>
								<th>Status</th>
								<th class="text-right">Amount</th>
								<th>Reason</th>
								<th>Created</th>
								<th class="text-right">Actions</th>
							</tr>
						</thead>
						<tbody>
							{#each returns as ret (ret.ID)}
								<tr>
									<td class="num-mono font-semibold text-foreground">
										{returnNumber(ret.ID)}
									</td>
									<td class="num-mono text-xs text-muted-foreground">
										{poNumber(ret.PurchaseOrderID)}
									</td>
									<td class="text-foreground">
										{ret.Supplier?.Name || `Supplier #${ret.SupplierID}`}
									</td>
									<td>
										<span class={cn('uppercase', getStatusColor(ret.Status))}>
											{ret.Status}
										</span>
									</td>
									<td class="col-num font-semibold text-foreground">
										{formatCurrency(ret.TotalAmount)}
									</td>
									<td class="max-w-[12rem] truncate text-sm text-muted-foreground">
										{ret.Reason}
									</td>
									<td class="text-sm text-muted-foreground">
										<TimeAgo date={ret.CreatedAt} />
									</td>
									<td>
										<div class="flex items-center justify-end gap-0.5">
											{#if actionLoading === ret.ID}
												<Loader2 class="h-4 w-4 animate-spin text-primary" />
											{:else}
												{#if ret.Status === 'PENDING'}
													<Button
														variant="ghost"
														size="sm"
														class="h-8 w-8 p-0 text-muted-foreground hover:text-success"
														title="Approve"
														aria-label="Approve purchase return"
														onclick={() => handleApprove(ret)}
													>
														<CheckCircle2 class="h-3.5 w-3.5" />
													</Button>
												{/if}
												{#if ret.Status === 'APPROVED'}
													<Button
														variant="ghost"
														size="sm"
														class="h-8 w-8 p-0 text-muted-foreground hover:text-info"
														title="Complete"
														aria-label="Complete purchase return"
														onclick={() => handleComplete(ret)}
													>
														<PackageCheck class="h-3.5 w-3.5" />
													</Button>
												{/if}
												{#if ret.Status === 'PENDING' || ret.Status === 'APPROVED'}
													<Button
														variant="ghost"
														size="sm"
														class="h-8 w-8 p-0 text-muted-foreground hover:text-danger"
														title="Cancel"
														aria-label="Cancel purchase return"
														onclick={() => handleCancel(ret)}
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
						Showing {returns.length} of {total} returns
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

<Dialog.Root bind:open={isCreateModalOpen}>
	<Dialog.Content class="max-h-[85vh] overflow-y-auto sm:max-w-[550px]">
		<Dialog.Header>
			<Dialog.Title>Create Purchase Return</Dialog.Title>
			<Dialog.Description>
				Select a received purchase order and specify return details.
			</Dialog.Description>
		</Dialog.Header>

		<div class="space-y-4 py-4">
			<div class="space-y-2">
				<Label for="pr-po">Purchase Order</Label>
				<select
					id="pr-po"
					bind:value={createForm.purchaseOrderId}
					onchange={onPoSelect}
					class="h-11 text-sm"
				>
					<option value="">Select a PO...</option>
					{#each purchaseOrders as po}
						<option value={String(po.ID)}>
							{poNumber(po.ID)} — {(po as any).Supplier?.Name || `Supplier #${po.SupplierID}`}
						</option>
					{/each}
				</select>
			</div>

			<div class="space-y-2">
				<Label for="pr-reason">Reason</Label>
				<Input
					id="pr-reason"
					bind:value={createForm.reason}
					placeholder="e.g. Damaged goods, wrong items, quality issue..."
					class="h-11 w-full rounded-md text-sm"
				/>
			</div>

			{#if createForm.items.length > 0}
				<div class="space-y-2">
					<Label>Items</Label>
					<div class="overflow-hidden rounded-lg ring-1 ring-border">
						<table class="ledger-table">
							<thead>
								<tr>
									<th>Product</th>
									<th class="text-right">Qty</th>
									<th class="text-right">Unit Price</th>
								</tr>
							</thead>
							<tbody>
								{#each createForm.items as item, i}
									<tr>
										<td class="num-mono text-sm">
											Product #{item.productId}
										</td>
										<td class="text-right">
											<Input
												type="number"
												min="1"
												bind:value={item.quantity}
												class="num-mono ml-auto h-8 w-20 rounded-md px-2 text-right text-sm"
											/>
										</td>
										<td class="col-num">
											{formatCurrency(item.unitPrice)}
										</td>
									</tr>
								{/each}
							</tbody>
						</table>
					</div>
				</div>
			{/if}
		</div>

		<Dialog.Footer>
			<Button variant="outline" onclick={() => (isCreateModalOpen = false)}>Cancel</Button>
			<Button onclick={handleCreate} disabled={creating} class="btn-soothe">
				{#if creating}
					<Loader2 class="mr-2 h-4 w-4 animate-spin" />
				{/if}
				Create Return
			</Button>
		</Dialog.Footer>
	</Dialog.Content>
</Dialog.Root>
