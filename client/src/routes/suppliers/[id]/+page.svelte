<!-- client/src/routes/suppliers/[id]/+page.svelte -->
<script lang="ts">
	import { page } from '$app/stores';
	import {
		Card,
		CardContent,
		CardDescription,
		CardHeader,
		CardTitle
	} from '$lib/components/ui/card';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import { ArrowLeft, User, BarChart2 } from 'lucide-svelte';
	import type { PageData } from './$types';
	import { auth } from '$lib/stores/auth';
	import { goto } from '$app/navigation';
	import { toast } from 'svelte-sonner';

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
</script>

<div class="mx-auto w-full max-w-6xl px-6 py-8">
	<section class="space-y-8">
		<!-- HEADER -->
		<div class="flex items-center justify-between">
			<a
				href="/catalog"
				class="flex items-center text-violet-600 transition-colors hover:text-violet-800"
			>
				<ArrowLeft class="mr-2 h-5 w-5" />
				Back to Catalog
			</a>
		</div>

		{#if supplier && performance}
			<div class="grid gap-8 lg:grid-cols-3">
				<!-- Supplier Details -->
				<div class="lg:col-span-1">
					<Card
						class="overflow-hidden rounded-2xl border-0 bg-gradient-to-br from-violet-50 to-purple-100 shadow-lg transition-all duration-300 hover:scale-[1.01] hover:shadow-xl"
					>
						<CardHeader
							class="space-y-1 border-b border-white/60 bg-white/70 px-6 py-5 backdrop-blur"
						>
							<CardTitle class="flex items-center text-slate-800">
								<User class="mr-2 h-5 w-5 text-violet-600" />
								{supplier.Name}
							</CardTitle>
							<CardDescription class="text-slate-600">Supplier Details</CardDescription>
						</CardHeader>
						<CardContent class="space-y-4 p-6 text-sm">
							<div>
								<p class="font-medium text-slate-500">Contact Person</p>
								<p class="text-slate-800">{supplier.ContactPerson || 'N/A'}</p>
							</div>
							<div>
								<p class="font-medium text-slate-500">Email</p>
								<p class="text-slate-800">{supplier.Email || 'N/A'}</p>
							</div>
							<div>
								<p class="font-medium text-slate-500">Phone</p>
								<p class="text-slate-800">{supplier.Phone || 'N/A'}</p>
							</div>
							<div>
								<p class="font-medium text-slate-500">Address</p>
								<p class="text-slate-800">{supplier.Address || 'N/A'}</p>
							</div>
						</CardContent>
					</Card>
				</div>

				<!-- Performance Report -->
				<div class="lg:col-span-2">
					<Card
						class="overflow-hidden rounded-2xl border-0 bg-gradient-to-br from-teal-50 to-cyan-100 shadow-lg transition-all duration-300 hover:scale-[1.01] hover:shadow-xl"
					>
						<CardHeader
							class="space-y-1 border-b border-white/60 bg-white/70 px-6 py-5 backdrop-blur"
						>
							<CardTitle class="flex items-center text-slate-800">
								<BarChart2 class="mr-2 h-5 w-5 text-teal-600" />
								Performance Report
							</CardTitle>
							<CardDescription class="text-slate-600">Key performance indicators</CardDescription>
						</CardHeader>
						<CardContent class="grid grid-cols-1 gap-6 p-6 sm:grid-cols-2">
							<div class="rounded-lg bg-white/50 p-4">
								<p class="font-medium text-slate-500">On-Time Delivery Rate</p>
								<p class="text-3xl font-bold text-teal-600">
									{(performance.onTimeDeliveryRate * 100).toFixed(1)}%
								</p>
							</div>
							<div class="rounded-lg bg-white/50 p-4">
								<p class="font-medium text-slate-500">Average Lead Time</p>
								<p class="text-3xl font-bold text-cyan-600">
									{performance.averageLeadTimeDays.toFixed(1)} days
								</p>
							</div>
						</CardContent>
					</Card>
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
