<script lang="ts">
	import { toast } from 'svelte-sonner';
	import { Button } from '$lib/components/ui/button';
	import EntityTable from './components/EntityTable.svelte';
	import LocationForm from './components/LocationForm.svelte';
	import { locationsApi } from '$lib/api/resources';
	import type { Location } from '$lib/types';
	import type { CatalogTab } from '$lib/catalog/detailSections';
	import { auth } from '$lib/stores/auth';
	import { t } from '$lib/i18n';

	type Props = {
		refreshToken?: number;
		onViewDetails: (resource: Record<string, any>, type: CatalogTab) => void;
	};

	let { refreshToken = 0, onViewDetails }: Props = $props();

	let loading = $state(false);
	let locations = $state<Location[]>([]);
	let editingLocation = $state<Location | null>(null);

	const locationColumns = $derived([
		{ header: $t('catalog.locations.form.name'), accessorKey: 'Name' },
		{ header: $t('catalog.locations.columns.address'), accessorKey: 'Address' },
		...(auth.hasPermission('locations.write')
			? [{ header: $t('catalog.common.actions'), accessorKey: 'id', class: 'text-right' }]
			: [])
	]);

	const loadLocations = async () => {
		loading = true;
		try {
			const locationList = await locationsApi.list();
			locations = Array.isArray(locationList) ? locationList : [locationList];
		} catch (error: any) {
			const errorMessage = error.response?.data?.error || 'Unable to load locations';
			toast.error('Failed to Load Catalog', { description: errorMessage });
		} finally {
			loading = false;
		}
	};

	const deleteLocation = async (location: Location) => {
		if (!window.confirm($t('catalog.toasts.confirm_delete', { name: location.Name }))) return;
		try {
			await locationsApi.remove(location.ID);
			toast.success($t('catalog.toasts.location_removed'));
			await loadLocations();
		} catch (error: any) {
			const errorMessage = error.response?.data?.error || 'Unable to delete location';
			toast.error($t('catalog.toasts.location_remove_fail'), { description: errorMessage });
		}
	};

	$effect(() => {
		if (refreshToken >= 0) loadLocations();
	});
</script>

<div
	class="grid gap-6 {auth.hasPermission('locations.write')
		? 'lg:grid-cols-[minmax(0,2fr)_minmax(0,1fr)]'
		: 'grid-cols-1'}"
>
	<div class="flex min-w-0 flex-col gap-4">
		<div class="panel p-4">
			<div class="min-w-0 space-y-1">
				<h2 class="section-title">{$t('catalog.locations.title')}</h2>
				<p class="text-sm text-muted-foreground">{$t('catalog.locations.subtitle')}</p>
			</div>
		</div>

		<EntityTable
			data={locations}
			columns={locationColumns}
			{loading}
			onRowClick={(location) => onViewDetails(location, 'locations')}
		>
			{#snippet children(location)}
				<td class="font-medium text-foreground">{location.Name}</td>
				<td class="text-muted-foreground">{location.Address ?? '—'}</td>
				{#if auth.hasPermission('locations.write')}
					<td class="text-right">
						<div class="flex items-center justify-end gap-1">
							<Button
								size="sm"
								variant="ghost"
								class="h-8 px-2.5 text-muted-foreground hover:bg-muted hover:text-foreground"
								onclick={(event) => {
									event.stopPropagation();
									editingLocation = location;
								}}
							>
								{$t('catalog.common.edit')}
							</Button>
							<Button
								size="sm"
								variant="ghost"
								class="h-8 px-2.5 text-danger hover:bg-danger/10 hover:text-danger"
								onclick={(event) => {
									event.stopPropagation();
									deleteLocation(location);
								}}
							>
								{$t('catalog.common.delete')}
							</Button>
						</div>
					</td>
				{/if}
			{/snippet}
		</EntityTable>
	</div>

	{#if auth.hasPermission('locations.write')}
		<LocationForm editing={editingLocation} onsaved={() => loadLocations()} />
	{/if}
</div>
