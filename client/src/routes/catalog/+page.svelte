<script lang="ts">
	import { toast } from 'svelte-sonner';
	import { Button } from '$lib/components/ui/button';
	import GlassTabs from '$lib/components/ui/glass-tabs.svelte';
	import DetailsModal from '$lib/components/DetailsModal.svelte';
	import ProductsTab from './ProductsTab.svelte';
	import CategoriesTab from './CategoriesTab.svelte';
	import SubCategoriesTab from './SubCategoriesTab.svelte';
	import SuppliersTab from './SuppliersTab.svelte';
	import LocationsTab from './LocationsTab.svelte';
	import { getDetailConfig, type CatalogTab, type DetailConfig } from '$lib/catalog/detailSections';
	import { PlusCircle, RefreshCcw, Zap } from 'lucide-svelte';
	import { auth } from '$lib/stores/auth';
	import { goto } from '$app/navigation';
	import { t } from '$lib/i18n';

	let activeTab = $state<CatalogTab>('products');
	let refreshToken = $state(0);

	let selectedResourceId = $state<number | null>(null);
	let isModalOpen = $state(false);
	let modalConfig = $state<DetailConfig | null>(null);

	$effect(() => {
		if (!auth.hasPermission('products.read')) {
			toast.error($t('catalog.toasts.access_denied'), {
				description: $t('catalog.toasts.access_denied_desc')
			});
			goto('/');
		}
	});

	const viewDetails = (resource: Record<string, any>, type: CatalogTab) => {
		selectedResourceId = resource.ID;
		modalConfig = getDetailConfig(type, resource);
		isModalOpen = true;
	};
</script>

{#if modalConfig}
	<DetailsModal
		bind:open={isModalOpen}
		resourceId={selectedResourceId}
		endpoint={modalConfig.endpoint}
		title={modalConfig.title}
		subtitle={modalConfig.subtitle}
		extraFetchers={modalConfig.extraFetchers}
		buildSections={modalConfig.buildSections}
	/>
{/if}

<!-- ===== FIXED HERO (responsive, clean parallax, correct layering) ===== -->
<section class="relative isolate w-full overflow-hidden">
	<div
		class="animate-gradientShift absolute inset-0 -z-10 bg-gradient-to-r from-sky-50 via-blue-50 to-cyan-100 bg-[length:200%_200%]"
	></div>

	<div
		class="animate-pulseGlow absolute -top-32 -left-24 h-96 w-96 rounded-full bg-sky-200/40 blur-3xl"
	></div>
	<div
		class="animate-pulseGlow absolute -right-24 -bottom-28 h-80 w-80 rounded-full bg-cyan-200/30 blur-3xl delay-700"
	></div>

	<div
		class="parallax-hero relative mx-auto max-w-7xl px-4 pt-16 pb-10 text-center sm:px-6 sm:pt-20 sm:pb-16 sm:text-left lg:px-8"
	>
		<div class="mb-3 inline-flex items-center justify-center gap-3 sm:justify-start">
			<span
				class="animate-cardFloat inline-flex rounded-2xl bg-gradient-to-br from-sky-500 to-blue-600 p-2 shadow-md"
			>
				<Zap class="h-6 w-6 text-white" />
			</span>
			<p class="text-xs font-semibold tracking-[0.18em] text-sky-700 uppercase sm:text-sm">
				{$t('catalog.hero.subtitle')}
			</p>
		</div>

		<h1
			class="mb-3 bg-gradient-to-r from-sky-700 via-blue-700 to-cyan-700 bg-clip-text text-3xl font-bold text-transparent sm:text-4xl lg:text-5xl"
		>
			{$t('catalog.hero.title')}
		</h1>
		<p class="mx-auto max-w-2xl text-sm text-slate-600 sm:mx-0 sm:text-base">
			{$t('catalog.hero.description')}
		</p>

		<div class="mt-6 flex flex-col justify-center gap-3 sm:flex-row sm:justify-start">
			<Button
				variant="secondary"
				onclick={() => refreshToken++}
				class="w-full rounded-xl bg-gradient-to-r from-sky-500 to-blue-600 px-5 py-2.5 font-medium text-white shadow-lg transition-all duration-300 hover:scale-105 hover:from-sky-600 hover:to-blue-700 hover:shadow-xl focus:ring-2 focus:ring-sky-300 sm:w-auto"
			>
				<RefreshCcw class="mr-2 h-4 w-4" />
				{$t('catalog.hero.sync_data')}
			</Button>
			<Button
				href="/bulk"
				variant="outline"
				class="w-full rounded-xl border border-sky-200 bg-white/80 px-5 py-2.5 font-medium text-sky-700 shadow-md transition-all duration-300 hover:scale-105 hover:bg-sky-50 hover:shadow-lg focus:ring-2 focus:ring-sky-200 sm:w-auto"
			>
				<PlusCircle class="mr-2 h-4 w-4" />
				{$t('catalog.hero.bulk_import')}
			</Button>
		</div>
	</div>
</section>

<!-- ===== TABS ===== -->
<div class="mx-auto mt-6 max-w-7xl px-4 sm:px-6 lg:px-8">
	<GlassTabs
		bind:value={activeTab}
		tabs={[
			{ value: 'products', label: $t('catalog.tabs.products') },
			{ value: 'categories', label: $t('catalog.tabs.categories') },
			{ value: 'sub-categories', label: $t('catalog.tabs.sub_categories') },
			{ value: 'suppliers', label: $t('catalog.tabs.suppliers') },
			{ value: 'locations', label: $t('catalog.tabs.locations') }
		]}
		class="mb-8 overflow-x-auto"
	/>

	<section class="space-y-10">
		{#key activeTab}
			{#if activeTab === 'products'}
				<ProductsTab
					{refreshToken}
					onViewDetails={viewDetails}
					onNavigate={(tab) => (activeTab = tab)}
				/>
			{:else if activeTab === 'categories'}
				<CategoriesTab {refreshToken} onViewDetails={viewDetails} />
			{:else if activeTab === 'sub-categories'}
				<SubCategoriesTab {refreshToken} onViewDetails={viewDetails} />
			{:else if activeTab === 'suppliers'}
				<SuppliersTab {refreshToken} onViewDetails={viewDetails} />
			{:else if activeTab === 'locations'}
				<LocationsTab {refreshToken} onViewDetails={viewDetails} />
			{/if}
		{/key}
	</section>
</div>

<style lang="postcss">
	@keyframes gradientShift {
		0% {
			background-position: 0% 50%;
		}
		50% {
			background-position: 100% 50%;
		}
		100% {
			background-position: 0% 50%;
		}
	}
	.animate-gradientShift {
		background-size: 200% 200%;
		animation: gradientShift 18s ease-in-out infinite;
	}

	@keyframes pulseGlow {
		0%,
		100% {
			transform: scale(1);
			opacity: 0.45;
		}
		50% {
			transform: scale(1.08);
			opacity: 0.7;
		}
	}
	.animate-pulseGlow {
		animation: pulseGlow 12s ease-in-out infinite;
	}

	@keyframes cardFloat {
		0%,
		100% {
			transform: translateY(0);
		}
		50% {
			transform: translateY(-4px);
		}
	}
	.animate-cardFloat {
		animation: cardFloat 4s ease-in-out infinite;
	}

	::-webkit-scrollbar {
		width: 8px;
		height: 8px;
	}
	::-webkit-scrollbar-track {
		background: transparent;
	}
	::-webkit-scrollbar-thumb {
		background: rgba(14, 165, 233, 0.25);
		border-radius: 9999px;
	}
	::-webkit-scrollbar-thumb:hover {
		background: rgba(14, 165, 233, 0.35);
	}

	.parallax-hero {
		transform: translateY(0);
		will-change: transform, filter;
		transition:
			transform 0.1s ease-out,
			filter 0.2s ease-out;
	}

	@media (max-width: 640px) {
		.parallax-hero {
			padding-top: 5rem;
			padding-bottom: 3rem;
		}
	}
</style>
