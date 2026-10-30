<!-- client/src/routes/suppliers/[id]/+page.svelte -->
<script lang="ts">
	import { page } from '$app/stores';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import { ArrowLeft, User, BarChart2 } from 'lucide-svelte';
	import Breadcrumbs from '$lib/components/ui/breadcrumbs.svelte';
	import type { PageData } from './$types';
	import { auth } from '$lib/stores/auth';
	import { goto } from '$app/navigation';
	import { toast } from 'svelte-sonner';
	import { addRecent } from '$lib/recent';

	$effect(() => {
		if (!auth.hasPermission('suppliers.read')) {
			toast.error('Access Denied', {
				description: 'You do not have permission to view suppliers.'
			});
			goto('/');
		}
	});

	let { data }: { data: PageData } = $props();

	let supplier = $derived(data.supplier);
	let performance = $derived(data.performance);

	$effect(() => {
		if (supplier?.ID) {
			addRecent({
				type: 'supplier',
				id: supplier.ID,
				label: supplier.Name,
				href: `/suppliers/${supplier.ID}`
			});
		}
	});
</script>

<div class="mx-auto w-full max-w-6xl px-4 py-8 sm:px-6 lg:px-8">
	<div class="space-y-6">
		{#if supplier}
			<Breadcrumbs
				items={[
					{ label: 'Catalog', href: '/catalog' },
					{ label: 'Suppliers', href: '/catalog' },
					{ label: supplier.Name }
				]}
			/>
		{/if}

		<!-- HEADER -->
		<div class="page-header">
			<div class="min-w-0 space-y-2">
				<p class="overline">Catalog · Supplier</p>
				<h1 class="page-title truncate">{supplier?.Name ?? 'Supplier'}</h1>
				<p class="page-sub">Supplier details and delivery performance.</p>
			</div>
			<div class="flex flex-wrap items-center gap-2">
				<a
					href="/catalog"
					class="btn-soft inline-flex h-10 items-center gap-2 px-4 text-sm font-medium"
				>
					<ArrowLeft class="h-4 w-4" />
					Back to Catalog
				</a>
			</div>
		</div>

		{#if supplier && performance}
			<div class="grid gap-6 lg:grid-cols-3">
				<!-- Supplier Details -->
				<div class="min-w-0 lg:col-span-1">
					<div class="panel overflow-hidden">
						<div class="panel-header">
							<div class="flex items-center gap-2">
								<User class="h-4 w-4 text-primary" />
								<h2 class="section-title">Contact</h2>
							</div>
						</div>
						<div class="panel-body">
							<dl class="space-y-4 text-sm">
								<div class="min-w-0">
									<dt class="overline">Contact Person</dt>
									<dd class="mt-1 text-foreground">{supplier.ContactPerson || 'N/A'}</dd>
								</div>
								<div class="min-w-0">
									<dt class="overline">Email</dt>
									<dd class="mt-1 break-words text-foreground">{supplier.Email || 'N/A'}</dd>
								</div>
								<div class="min-w-0">
									<dt class="overline">Phone</dt>
									<dd class="num-mono mt-1 text-foreground">{supplier.Phone || 'N/A'}</dd>
								</div>
								<div class="min-w-0">
									<dt class="overline">Address</dt>
									<dd class="mt-1 text-foreground">{supplier.Address || 'N/A'}</dd>
								</div>
							</dl>
						</div>
					</div>
				</div>

				<!-- Performance Report -->
				<div class="min-w-0 lg:col-span-2">
					<div class="panel overflow-hidden">
						<div class="panel-header">
							<div class="flex items-center gap-2">
								<BarChart2 class="h-4 w-4 text-primary" />
								<h2 class="section-title">Performance Report</h2>
							</div>
							<span class="text-xs text-muted-foreground">Key performance indicators</span>
						</div>
						<div class="panel-body">
							<div class="grid grid-cols-1 gap-4 sm:grid-cols-2">
								<div class="surface-1 p-4">
									<p class="overline">On-Time Delivery Rate</p>
									<p class="kpi-num mt-2 text-3xl text-success">
										{(performance.onTimeDeliveryRate * 100).toFixed(1)}%
									</p>
								</div>
								<div class="surface-1 p-4">
									<p class="overline">Average Lead Time</p>
									<p class="kpi-num mt-2 text-3xl text-info">
										{performance.averageLeadTimeDays.toFixed(1)} days
									</p>
								</div>
							</div>
						</div>
					</div>
				</div>
			</div>
		{:else}
			<!-- Loading State -->
			<div class="grid gap-6 lg:grid-cols-3">
				<div class="lg:col-span-1">
					<Skeleton class="h-64 w-full rounded-xl" />
				</div>
				<div class="lg:col-span-2">
					<Skeleton class="h-64 w-full rounded-xl" />
				</div>
			</div>
		{/if}
	</div>
</div>
