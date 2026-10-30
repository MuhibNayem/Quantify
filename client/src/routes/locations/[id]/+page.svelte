<!-- client/src/routes/locations/[id]/+page.svelte -->
<script lang="ts">
	import { page } from '$app/stores';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import { ArrowLeft, MapPin } from 'lucide-svelte';
	import Breadcrumbs from '$lib/components/ui/breadcrumbs.svelte';
	import type { PageData } from './$types';

	let { data }: { data: PageData } = $props();

	let location = $derived(data.location);
</script>

<div class="mx-auto w-full max-w-4xl px-4 py-8 sm:px-6 lg:px-8">
	<div class="space-y-6">
		{#if location}
			<Breadcrumbs
				items={[
					{ label: 'Catalog', href: '/catalog' },
					{ label: 'Locations', href: '/catalog' },
					{ label: location.Name }
				]}
			/>
		{/if}

		<!-- HEADER -->
		<div class="page-header">
			<div class="min-w-0 space-y-2">
				<p class="overline">Catalog · Location</p>
				<h1 class="page-title truncate">{location?.Name ?? 'Location'}</h1>
				<p class="page-sub">Stocking location details.</p>
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

		{#if location}
			<div class="panel overflow-hidden">
				<div class="panel-header">
					<div class="flex items-center gap-2">
						<MapPin class="h-4 w-4 text-primary" />
						<h2 class="section-title">Location Details</h2>
					</div>
				</div>
				<div class="panel-body">
					<dl class="grid grid-cols-1 gap-4 text-sm">
						<div class="min-w-0">
							<dt class="overline">Address</dt>
							<dd class="mt-1 text-foreground">{location.Address || 'N/A'}</dd>
						</div>
					</dl>
				</div>
			</div>
		{:else}
			<!-- Loading State -->
			<Skeleton class="h-48 w-full rounded-xl" />
		{/if}
	</div>
</div>
