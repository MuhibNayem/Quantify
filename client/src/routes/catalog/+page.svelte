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

	const btnPrimary = 'btn-soothe h-10 px-5 text-sm font-semibold';
	const btnSecondary = 'btn-soft h-10 px-5 text-sm font-medium';
	const tabsClass = 'mb-8 overflow-x-auto rounded-xl';

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

<!-- ===== HEADER ===== -->
<section class="w-full">
	<div class="mx-auto max-w-7xl px-4 pt-10 pb-6 sm:px-6 sm:pt-14 lg:px-8">
		<div class="page-header" data-animate="fade-up">
			<div class="min-w-0 space-y-2">
				<p class="flex items-center gap-2 overline">
					<span class="inline-flex rounded-md bg-primary/10 p-1.5 text-primary">
						<Zap class="h-4 w-4" />
					</span>
					{$t('catalog.hero.subtitle')}
				</p>
				<h1 class="page-title">{$t('catalog.hero.title')}</h1>
				<p class="page-sub">{$t('catalog.hero.description')}</p>
			</div>

			<div class="flex flex-col gap-3 sm:flex-row sm:items-center">
				<Button onclick={() => refreshToken++} class={btnPrimary}>
					<RefreshCcw class="h-4 w-4" />
					{$t('catalog.hero.sync_data')}
				</Button>
				<Button href="/bulk" variant="secondary" class={btnSecondary}>
					<PlusCircle class="h-4 w-4" />
					{$t('catalog.hero.bulk_import')}
				</Button>
			</div>
		</div>
	</div>
</section>

<!-- ===== TABS ===== -->
<div class="mx-auto mt-2 max-w-7xl px-4 sm:px-6 lg:px-8">
	<GlassTabs
		bind:value={activeTab}
		tabs={[
			{ value: 'products', label: $t('catalog.tabs.products') },
			{ value: 'categories', label: $t('catalog.tabs.categories') },
			{ value: 'sub-categories', label: $t('catalog.tabs.sub_categories') },
			{ value: 'suppliers', label: $t('catalog.tabs.suppliers') },
			{ value: 'locations', label: $t('catalog.tabs.locations') }
		]}
		class={tabsClass}
	/>

	<section class="space-y-8">
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
