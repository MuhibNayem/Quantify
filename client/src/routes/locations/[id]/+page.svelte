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

<div class="mx-auto w-full max-w-4xl px-6 py-8">
	<section class="space-y-8">
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
		<div class="flex items-center justify-between">
			<a
				href="/catalog"
				class="text-primary hover:text-foreground flex items-center text-sm font-medium transition-colors"
			>
				<ArrowLeft class="mr-2 h-4 w-4" />
				Back to Catalog
			</a>
		</div>

		{#if location}
			<div class="liquid-panel overflow-hidden rounded-lg">
				<div class="border-border space-y-1 border-b px-6 py-5">
					<h1 class="text-foreground flex items-center font-display text-xl font-semibold">
						<MapPin class="text-primary mr-2 h-5 w-5" />
						{location.Name}
					</h1>
					<p class="text-muted-foreground text-sm">Location Details</p>
				</div>
				<div class="p-6 text-sm">
					<div>
						<p class="overline">Address</p>
						<p class="text-foreground mt-0.5">{location.Address || 'N/A'}</p>
					</div>
				</div>
			</div>
		{:else}
			<!-- Loading State -->
			<Skeleton class="h-48 w-full" />
		{/if}
	</section>
</div>
