<!-- client/src/routes/locations/[id]/+page.svelte -->
<script lang="ts">
	import { page } from '$app/stores';
	import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '$lib/components/ui/card';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import { ArrowLeft, MapPin } from 'lucide-svelte';
	import type { PageData } from './$types';

	export let data: PageData;

	const { location } = data;
</script>

<div class="w-full max-w-4xl mx-auto py-8 px-6">
	<section class="space-y-8">
		<!-- HEADER -->
		<div class="flex items-center justify-between">
			<a href="/catalog" class="flex items-center text-cyan-600 hover:text-cyan-800 transition-colors">
				<ArrowLeft class="h-5 w-5 mr-2" />
				Back to Catalog
			</a>
		</div>

		{#if location}
			<Card class="rounded-2xl shadow-lg hover:shadow-xl transition-all duration-300 hover:scale-[1.01] overflow-hidden border-0 bg-gradient-to-br from-cyan-50 to-teal-100">
				<CardHeader class="space-y-1 bg-white/70 backdrop-blur px-6 py-5 border-b border-white/60">
					<CardTitle class="text-slate-800 flex items-center">
						<MapPin class="h-5 w-5 mr-2 text-cyan-600" />
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
