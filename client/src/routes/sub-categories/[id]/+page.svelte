<!-- client/src/routes/sub-categories/[id]/+page.svelte -->
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
	import { ArrowLeft, Tag } from 'lucide-svelte';
	import type { PageData } from './$types';
	import { auth } from '$lib/stores/auth';
	import { goto } from '$app/navigation';
	import { toast } from 'svelte-sonner';

	$effect(() => {
		if (!auth.hasPermission('categories.read')) {
			toast.error('Access Denied', {
				description: 'You do not have permission to view sub-categories.'
			});
			goto('/');
		}
	});

	let { data }: { data: PageData } = $props();

	let subCategory = $derived(data.subCategory);
</script>

<div class="mx-auto w-full max-w-4xl px-6 py-8">
	<section class="space-y-8">
		<!-- HEADER -->
		<div class="flex items-center justify-between">
			<a
				href="/catalog"
				class="flex items-center text-amber-600 transition-colors hover:text-amber-800"
			>
				<ArrowLeft class="mr-2 h-5 w-5" />
				Back to Catalog
			</a>
		</div>

		{#if subCategory}
			<Card
				class="overflow-hidden rounded-2xl border-0 bg-gradient-to-br from-amber-50 to-orange-100 shadow-lg transition-all duration-300 hover:scale-[1.01] hover:shadow-xl"
			>
				<CardHeader class="space-y-1 border-b border-white/60 bg-white/70 px-6 py-5 backdrop-blur">
					<CardTitle class="flex items-center text-slate-800">
						<Tag class="mr-2 h-5 w-5 text-amber-600" />
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
