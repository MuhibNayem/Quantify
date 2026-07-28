<script lang="ts">
	import { toast } from 'svelte-sonner';
	import { TableCell } from '$lib/components/ui/table';
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
	class="grid gap-8 {auth.hasPermission('locations.write')
		? 'lg:grid-cols-[2fr,1fr]'
		: 'grid-cols-1'}"
>
	<div class="flex flex-col gap-6">
		<div
			class="flex flex-col gap-4 rounded-2xl border border-teal-100 bg-white/50 p-4 shadow-sm backdrop-blur sm:flex-row sm:items-center sm:justify-between"
		>
			<div class="space-y-1">
				<h2 class="text-lg font-semibold text-slate-800">{$t('catalog.locations.title')}</h2>
				<p class="text-sm text-slate-500">{$t('catalog.locations.subtitle')}</p>
			</div>
		</div>

		<EntityTable
			data={locations}
			columns={locationColumns}
			{loading}
			onRowClick={(location) => onViewDetails(location, 'locations')}
		>
			{#snippet children(location)}
				<TableCell class="font-medium text-slate-800">{location.Name}</TableCell>
				<TableCell class="text-slate-600">{location.Address ?? '—'}</TableCell>
				{#if auth.hasPermission('locations.write')}
					<TableCell class="text-right">
						<div class="flex items-center justify-end gap-1">
							<Button
								size="sm"
								variant="ghost"
								class="h-8 text-teal-600 hover:bg-teal-50 hover:text-teal-700"
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
								class="h-8 text-rose-600 hover:bg-rose-50 hover:text-rose-700"
								onclick={(event) => {
									event.stopPropagation();
									deleteLocation(location);
								}}
							>
								{$t('catalog.common.delete')}
							</Button>
						</div>
					</TableCell>
				{/if}
			{/snippet}
		</EntityTable>
	</div>

	{#if auth.hasPermission('locations.write')}
		<LocationForm editing={editingLocation} onsaved={() => loadLocations()} />
	{/if}
</div>
