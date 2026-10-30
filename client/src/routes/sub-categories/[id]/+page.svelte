<!-- client/src/routes/sub-categories/[id]/+page.svelte -->
<script lang="ts">
	import { page } from '$app/stores';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import { ArrowLeft, Tag } from 'lucide-svelte';
	import Breadcrumbs from '$lib/components/ui/breadcrumbs.svelte';
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

<div class="mx-auto w-full max-w-4xl px-4 py-8 sm:px-6 lg:px-8">
	<div class="space-y-6">
		{#if subCategory}
			<Breadcrumbs
				items={[
					{ label: 'Catalog', href: '/catalog' },
					{ label: 'Sub-Categories', href: '/catalog' },
					{ label: subCategory.Name }
				]}
			/>
		{/if}

		<!-- HEADER -->
		<div class="page-header">
			<div class="min-w-0 space-y-2">
				<p class="overline">Catalog · Sub-Category</p>
				<h1 class="page-title truncate">{subCategory?.Name ?? 'Sub-Category'}</h1>
				<p class="page-sub">Grouping details for this sub-category.</p>
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

		{#if subCategory}
			<div class="panel overflow-hidden">
				<div class="panel-header">
					<div class="flex items-center gap-2">
						<Tag class="h-4 w-4 text-primary" />
						<h2 class="section-title">Sub-Category Details</h2>
					</div>
				</div>
				<div class="panel-body">
					<dl class="text-sm">
						<div class="min-w-0">
							<dt class="overline">Parent Category</dt>
							<dd class="mt-1">
								<a
									href="/categories/{subCategory.CategoryID}"
									class="font-medium text-primary underline-offset-4 hover:underline"
								>
									{subCategory.Category?.Name || 'N/A'}
								</a>
							</dd>
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
