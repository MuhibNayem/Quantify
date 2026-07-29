<!-- client/src/routes/sub-categories/[id]/+page.svelte -->
<script lang="ts">
	import { page } from '$app/stores';
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
				class="text-primary hover:text-foreground flex items-center text-sm font-medium transition-colors"
			>
				<ArrowLeft class="mr-2 h-4 w-4" />
				Back to Catalog
			</a>
		</div>

		{#if subCategory}
			<div class="liquid-panel overflow-hidden rounded-lg">
				<div class="border-border space-y-1 border-b px-6 py-5">
					<h1 class="text-foreground flex items-center font-display text-xl font-semibold">
						<Tag class="text-primary mr-2 h-5 w-5" />
						{subCategory.Name}
					</h1>
					<p class="text-muted-foreground text-sm">Sub-Category Details</p>
				</div>
				<div class="p-6">
					<div class="text-sm">
						<p class="overline">Parent Category</p>
						<a
							href="/categories/{subCategory.CategoryID}"
							class="text-primary mt-0.5 inline-block hover:underline"
						>
							{subCategory.Category?.Name || 'N/A'}
						</a>
					</div>
				</div>
			</div>
		{:else}
			<!-- Loading State -->
			<Skeleton class="h-48 w-full" />
		{/if}
	</section>
</div>
