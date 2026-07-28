<script lang="ts">
	import { superForm } from 'sveltekit-superforms';
	import { zod4 } from 'sveltekit-superforms/adapters';
	import { locationSchema } from '$lib/schemas/location';
	import { locationsApi } from '$lib/api/resources';
	import type { Location } from '$lib/types';
	import { toast } from 'svelte-sonner';
	import { t } from '$lib/i18n';
	import { Card, CardContent, CardHeader, CardTitle } from '$lib/components/ui/card';
	import { Input } from '$lib/components/ui/input';
	import { Button } from '$lib/components/ui/button';

	let { editing = null, onsaved }: { editing: Location | null; onsaved: () => void } = $props();

	const { form, errors, enhance, reset } = superForm(
		{ name: '', address: '' },
		{
			validators: zod4(locationSchema),
			onSubmit: async ({ cancel }) => {
				cancel();
				const data = $form;
				try {
					if (editing) {
						await locationsApi.update(editing.ID, { name: data.name, address: data.address });
					} else {
						await locationsApi.create({ name: data.name, address: data.address });
					}
					toast.success($t('catalog.toasts.location_saved'));
					reset();
					onsaved();
				} catch (error) {
					const errorMessage = (error as any)?.response?.data?.error || 'Unable to save location';
					toast.error($t('catalog.toasts.location_save_fail'), { description: errorMessage });
				}
			}
		}
	);

	$effect(() => {
		if (editing) {
			$form.name = editing.Name;
			$form.address = editing.Address ?? '';
		} else {
			reset();
		}
	});
</script>

<Card
	class="overflow-hidden rounded-2xl border-0 bg-gradient-to-br from-teal-50 to-emerald-100 shadow-lg transition-all duration-300 hover:scale-[1.01] hover:shadow-xl"
	data-animate="fade-up"
	style="animation-delay:180ms"
>
	<CardHeader class="space-y-1 border-b border-white/60 bg-white/70 px-6 py-5 backdrop-blur">
		<CardTitle class="text-slate-800"
			>{editing
				? $t('catalog.locations.form.update_title')
				: $t('catalog.locations.form.create_title')}</CardTitle
		>
	</CardHeader>
	<CardContent class="p-6">
		<form method="POST" use:enhance class="space-y-4">
			<div class="space-y-1">
				<Input
					class="w-full rounded-xl border border-teal-200 bg-white/90 px-3.5 py-2.5 text-sm focus:ring-2 focus:ring-teal-400"
					placeholder={$t('catalog.locations.form.name')}
					bind:value={$form.name}
				/>
				{#if $errors.name}
					<p class="text-xs font-medium text-rose-600">{$errors.name}</p>
				{/if}
			</div>
			<div class="space-y-1">
				<Input
					class="w-full rounded-xl border border-teal-200 bg-white/90 px-3.5 py-2.5 text-sm focus:ring-2 focus:ring-teal-400"
					placeholder={$t('catalog.locations.form.address')}
					bind:value={$form.address}
				/>
				{#if $errors.address}
					<p class="text-xs font-medium text-rose-600">{$errors.address}</p>
				{/if}
			</div>
			<div class="flex flex-col gap-3 pt-1 sm:flex-row">
				<Button
					type="submit"
					class="w-full rounded-xl bg-gradient-to-r from-teal-500 to-emerald-600 py-2.5 text-white shadow-lg transition-all duration-300 hover:scale-105 hover:from-teal-600 hover:to-emerald-700 hover:shadow-xl sm:w-1/2"
					>{editing ? $t('catalog.common.update') : $t('catalog.common.create')}</Button
				>
				<Button
					type="button"
					class="w-full rounded-xl border border-teal-200 py-2.5 text-teal-700 transition hover:bg-teal-50 sm:w-1/2"
					onclick={() => reset()}>{$t('catalog.common.reset')}</Button
				>
			</div>
		</form>
	</CardContent>
</Card>
