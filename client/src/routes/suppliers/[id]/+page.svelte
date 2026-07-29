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

<div class="mx-auto w-full max-w-6xl px-6 py-8">
	<section class="space-y-8">
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
		<div class="flex items-center justify-between">
			<a
				href="/catalog"
				class="text-primary hover:text-foreground flex items-center text-sm font-medium transition-colors"
			>
				<ArrowLeft class="mr-2 h-4 w-4" />
				Back to Catalog
			</a>
		</div>

		{#if supplier && performance}
			<div class="grid gap-8 lg:grid-cols-3">
				<!-- Supplier Details -->
				<div class="lg:col-span-1">
					<div class="liquid-panel overflow-hidden rounded-lg">
						<div class="border-border space-y-1 border-b px-6 py-5">
							<h1 class="text-foreground flex items-center font-display text-xl font-semibold">
								<User class="text-primary mr-2 h-5 w-5" />
								{supplier.Name}
							</h1>
							<p class="text-muted-foreground text-sm">Supplier Details</p>
						</div>
						<div class="space-y-4 p-6 text-sm">
							<div>
								<p class="overline">Contact Person</p>
								<p class="text-foreground mt-0.5">{supplier.ContactPerson || 'N/A'}</p>
							</div>
							<div>
								<p class="overline">Email</p>
								<p class="text-foreground mt-0.5">{supplier.Email || 'N/A'}</p>
							</div>
							<div>
								<p class="overline">Phone</p>
								<p class="num-mono text-foreground mt-0.5">{supplier.Phone || 'N/A'}</p>
							</div>
							<div>
								<p class="overline">Address</p>
								<p class="text-foreground mt-0.5">{supplier.Address || 'N/A'}</p>
							</div>
						</div>
					</div>
				</div>

				<!-- Performance Report -->
				<div class="lg:col-span-2">
					<div class="liquid-panel overflow-hidden rounded-lg">
						<div class="border-border space-y-1 border-b px-6 py-5">
							<h2 class="text-foreground flex items-center font-display text-xl font-semibold">
								<BarChart2 class="text-primary mr-2 h-5 w-5" />
								Performance Report
							</h2>
							<p class="text-muted-foreground text-sm">Key performance indicators</p>
						</div>
						<div class="grid grid-cols-1 gap-6 p-6 sm:grid-cols-2">
							<div class="border-border bg-muted/50 rounded-lg border p-4">
								<p class="overline">On-Time Delivery Rate</p>
								<p class="kpi-num text-success mt-1 text-3xl">
									{(performance.onTimeDeliveryRate * 100).toFixed(1)}%
								</p>
							</div>
							<div class="border-border bg-muted/50 rounded-lg border p-4">
								<p class="overline">Average Lead Time</p>
								<p class="kpi-num text-info mt-1 text-3xl">
									{performance.averageLeadTimeDays.toFixed(1)} days
								</p>
							</div>
						</div>
					</div>
				</div>
			</div>
		{:else}
			<!-- Loading State -->
			<div class="grid gap-8 lg:grid-cols-3">
				<div class="lg:col-span-1">
					<Skeleton class="h-64 w-full" />
				</div>
				<div class="lg:col-span-2">
					<Skeleton class="h-64 w-full" />
				</div>
			</div>
		{/if}
	</section>
</div>
