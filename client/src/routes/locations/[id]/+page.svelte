<!-- client/src/routes/locations/[id]/+page.svelte -->
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
	import { ArrowLeft, MapPin } from 'lucide-svelte';
	import type { PageData } from './$types';

	let { data }: { data: PageData } = $props();

	let location = $derived(data.location);
</script>

<div class="mx-auto w-full max-w-4xl px-6 py-8">
	<section class="space-y-8">
		<!-- HEADER -->
		<div class="flex items-center justify-between">
			<a
				href="/catalog"
				class="flex items-center text-cyan-600 transition-colors hover:text-cyan-800"
			>
				<ArrowLeft class="mr-2 h-5 w-5" />
				Back to Catalog
			</a>
		</div>

		{#if location}
			<Card
				class="overflow-hidden rounded-2xl border-0 bg-gradient-to-br from-cyan-50 to-teal-100 shadow-lg transition-all duration-300 hover:scale-[1.01] hover:shadow-xl"
			>
				<CardHeader class="space-y-1 border-b border-white/60 bg-white/70 px-6 py-5 backdrop-blur">
					<CardTitle class="flex items-center text-slate-800">
						<MapPin class="mr-2 h-5 w-5 text-cyan-600" />
						{location.Name}
					</CardTitle>
					<CardDescription class="text-slate-600">Location Details</CardDescription>
				</CardHeader>
				<CardContent class="p-6 text-sm">
					<div>
						<p class="font-medium text-slate-500">Address</p>
						<p class="text-slate-800">{location.Address || 'N/A'}</p>
					</div>
				</CardContent>
			</Card>
		{:else}
			<!-- Loading State -->
			<Skeleton class="h-48 w-full" />
		{/if}
	</section>
</div>
