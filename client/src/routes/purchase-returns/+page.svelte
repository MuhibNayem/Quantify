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
		Table,
		TableBody,
		TableCell,
		TableHead,
		TableHeader,
		TableRow
	} from '$lib/components/ui/table';
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
			case 'COMPLETED':
				return 'bg-emerald-100 text-emerald-700 border-emerald-200';
			case 'PENDING':
				return 'bg-amber-100 text-amber-700 border-amber-200';
			case 'CANCELLED':
				return 'bg-red-100 text-red-700 border-red-200';
			default:
				return 'bg-slate-100 text-slate-700 border-slate-200';
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

<div class="relative min-h-screen overflow-hidden bg-slate-50/50 p-6 lg:p-10">
	<div
		class="absolute inset-0 -z-10 bg-[radial-gradient(ellipse_at_top,_var(--tw-gradient-stops))] from-orange-100/20 via-slate-50/20 to-white/20"
	></div>

	<div class="mx-auto max-w-7xl space-y-8">
		<div class="flex flex-col gap-4 md:flex-row md:items-center md:justify-between">
			<div class="space-y-1">
				<h1
					class="bg-gradient-to-r from-slate-900 via-orange-900 to-slate-900 bg-clip-text text-3xl font-bold tracking-tight text-transparent"
				>
					Purchase Returns
				</h1>
				<p class="text-slate-500">Manage supplier returns and refunds</p>
			</div>
			<Button
				class="gap-2 bg-gradient-to-r from-orange-600 to-amber-600 text-white shadow-md hover:from-orange-500 hover:to-amber-500"
				onclick={openCreateModal}
			>
				<Plus class="h-4 w-4" />
				New Return
			</Button>
		</div>

		<div
			class="liquid-panel flex flex-col gap-4 rounded-2xl bg-gradient-to-br from-white/40 via-white/20 to-white/5 p-4 backdrop-blur-sm md:flex-row md:items-center"
		>
			<div class="w-48">
				<Select bind:value={statusFilter} options={statusOptions} placeholder="All Statuses" />
			</div>
			<div class="w-48">
				<Select bind:value={supplierFilter} options={supplierOptions} placeholder="All Suppliers" />
			</div>
			<Button variant="outline" size="sm" onclick={applyFilters} class="gap-1">
				<Search class="h-3 w-3" />
				Filter
			</Button>
		</div>

		{#if loading}
			<div class="flex h-64 items-center justify-center">
				<Loader2 class="h-8 w-8 animate-spin text-orange-500" />
			</div>
		{:else if returns.length === 0}
			<div class="rounded-2xl border border-white/60 bg-white/70 shadow-lg backdrop-blur-xl">
				<EmptyState
					icon={Undo2}
					title="No purchase returns found"
					description="Return damaged or unwanted stock to suppliers."
					actionLabel="New Return"
					onAction={openCreateModal}
				/>
			</div>
		{:else}
			<div
				in:fly={{ y: 20, duration: 300 }}
				class="overflow-hidden rounded-2xl border border-white/60 bg-white/70 shadow-lg backdrop-blur-xl"
			>
				<Table>
					<TableHeader>
						<TableRow class="border-slate-100 bg-slate-50/50">
							<TableHead class="font-bold text-slate-600">Return #</TableHead>
							<TableHead class="font-bold text-slate-600">PO #</TableHead>
							<TableHead class="font-bold text-slate-600">Supplier</TableHead>
							<TableHead class="font-bold text-slate-600">Status</TableHead>
							<TableHead class="font-bold text-slate-600">Amount</TableHead>
							<TableHead class="font-bold text-slate-600">Reason</TableHead>
							<TableHead class="font-bold text-slate-600">Created</TableHead>
							<TableHead class="text-right font-bold text-slate-600">Actions</TableHead>
						</TableRow>
					</TableHeader>
					<TableBody>
						{#each returns as ret (ret.ID)}
							<TableRow class="border-slate-50 transition-colors hover:bg-orange-50/30">
								<TableCell class="font-mono font-semibold text-slate-800">
									{returnNumber(ret.ID)}
								</TableCell>
								<TableCell class="font-mono text-slate-600">
									{poNumber(ret.PurchaseOrderID)}
								</TableCell>
								<TableCell class="text-slate-700">
									{ret.Supplier?.Name || `Supplier #${ret.SupplierID}`}
								</TableCell>
								<TableCell>
									<span
										class={cn(
											'inline-flex items-center rounded-full border px-2.5 py-0.5 text-xs font-bold tracking-wide uppercase',
											getStatusColor(ret.Status)
										)}
									>
										{ret.Status}
									</span>
								</TableCell>
								<TableCell class="font-semibold text-slate-800">
									{formatCurrency(ret.TotalAmount)}
								</TableCell>
								<TableCell class="max-w-[12rem] truncate text-slate-500">
									{ret.Reason}
								</TableCell>
								<TableCell class="text-slate-500">
									<TimeAgo date={ret.CreatedAt} />
								</TableCell>
								<TableCell>
									<div class="flex items-center justify-end gap-1">
										{#if actionLoading === ret.ID}
											<Loader2 class="h-4 w-4 animate-spin text-orange-500" />
										{:else}
											{#if ret.Status === 'PENDING'}
												<Button
													variant="ghost"
													size="sm"
													class="h-7 w-7 p-0 text-slate-500 hover:text-emerald-600"
													title="Approve"
													onclick={() => handleApprove(ret)}
												>
													<CheckCircle2 class="h-3.5 w-3.5" />
												</Button>
											{/if}
											{#if ret.Status === 'APPROVED'}
												<Button
													variant="ghost"
													size="sm"
													class="h-7 w-7 p-0 text-slate-500 hover:text-blue-600"
													title="Complete"
													onclick={() => handleComplete(ret)}
												>
													<PackageCheck class="h-3.5 w-3.5" />
												</Button>
											{/if}
											{#if ret.Status === 'PENDING' || ret.Status === 'APPROVED'}
												<Button
													variant="ghost"
													size="sm"
													class="h-7 w-7 p-0 text-slate-500 hover:text-red-600"
													title="Cancel"
													onclick={() => handleCancel(ret)}
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
						Showing {returns.length} of {total} returns
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
				<Label>Purchase Order</Label>
				<select
					bind:value={createForm.purchaseOrderId}
					onchange={onPoSelect}
					class="h-9 w-full rounded-md border border-slate-200 bg-white px-3 text-sm shadow-sm focus:border-orange-500 focus:ring-1 focus:ring-orange-500 focus:outline-none"
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
				<Label>Reason</Label>
				<Input
					bind:value={createForm.reason}
					placeholder="e.g. Damaged goods, wrong items, quality issue..."
				/>
			</div>

			{#if createForm.items.length > 0}
				<div class="space-y-2">
					<Label>Items</Label>
					<div class="rounded-lg border border-slate-200">
						<Table>
							<TableHeader>
								<TableRow>
									<TableHead class="text-xs">Product</TableHead>
									<TableHead class="text-xs">Qty</TableHead>
									<TableHead class="text-xs">Unit Price</TableHead>
								</TableRow>
							</TableHeader>
							<TableBody>
								{#each createForm.items as item, i}
									<TableRow>
										<TableCell class="text-sm">
											Product #{item.productId}
										</TableCell>
										<TableCell>
											<Input
												type="number"
												min="1"
												bind:value={item.quantity}
												class="h-7 w-20 text-sm"
											/>
										</TableCell>
										<TableCell class="text-sm">
											{formatCurrency(item.unitPrice)}
										</TableCell>
									</TableRow>
								{/each}
							</TableBody>
						</Table>
					</div>
				</div>
			{/if}
		</div>

		<Dialog.Footer>
			<Button variant="outline" onclick={() => (isCreateModalOpen = false)}>Cancel</Button>
			<Button
				onclick={handleCreate}
				disabled={creating}
				class="bg-gradient-to-r from-orange-600 to-amber-600 text-white hover:from-orange-500 hover:to-amber-500"
			>
				{#if creating}
					<Loader2 class="mr-2 h-4 w-4 animate-spin" />
				{/if}
				Create Return
			</Button>
		</Dialog.Footer>
	</Dialog.Content>
</Dialog.Root>
