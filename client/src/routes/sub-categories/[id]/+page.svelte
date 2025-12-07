<!-- client/src/routes/sub-categories/[id]/+page.svelte -->
<script lang="ts">
	import { page } from '$app/stores';
	import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '$lib/components/ui/card';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import { ArrowLeft, Tag } from 'lucide-svelte';
	import type { PageData } from './$types';

	export let data: PageData;

	const { subCategory } = data;
</script>

<div class="w-full max-w-4xl mx-auto py-8 px-6">
	<section class="space-y-8">
		<!-- HEADER -->
		<div class="flex items-center justify-between">
			<a href="/catalog" class="flex items-center text-amber-600 hover:text-amber-800 transition-colors">
				<ArrowLeft class="h-5 w-5 mr-2" />
				Back to Catalog
			</a>
		</div>

		{#if subCategory}
			<Card class="rounded-2xl shadow-lg hover:shadow-xl transition-all duration-300 hover:scale-[1.01] overflow-hidden border-0 bg-gradient-to-br from-amber-50 to-orange-100">
				<CardHeader class="space-y-1 bg-white/70 backdrop-blur px-6 py-5 border-b border-white/60">
					<CardTitle class="text-slate-800 flex items-center">
						<Tag class="h-5 w-5 mr-2 text-amber-600" />
						{subCategory.Name}
					</CardTitle>
					<CardDescription class="text-slate-600">Sub-Category Details</CardDescription>
				</CardHeader>
				<CardContent class="p-6">
					<div class="text-sm">
						<p class="font-medium text-slate-500">Parent Category</p>
						<a href="/categories/{subCategory.CategoryID}" class="text-slate-800 hover:underline">
							{subCategory.Category?.Name || 'N/A'}
						</a>
					</div>
				</CardContent>
			</Card>
		{:else}
			<!-- Loading State -->
			<Skeleton class="h-48 w-full" />
		{/if}
	</section>
</div>
