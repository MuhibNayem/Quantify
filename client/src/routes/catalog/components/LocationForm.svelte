<script lang="ts">
	import { superForm } from 'sveltekit-superforms';
	import { zod4 } from 'sveltekit-superforms/adapters';
	import { locationSchema } from '$lib/schemas/location';
	import { locationsApi } from '$lib/api/resources';
	import type { Location } from '$lib/types';
	import { toast } from 'svelte-sonner';
	import { t } from '$lib/i18n';
	import { Input } from '$lib/components/ui/input';
	import { Button } from '$lib/components/ui/button';

	let { editing = null, onsaved }: { editing: Location | null; onsaved: () => void } = $props();

	const inputClass =
		'liquid-input h-10 w-full rounded-md bg-card bg-none px-3 text-sm font-normal text-foreground shadow-none ring-0 backdrop-blur-none backdrop-saturate-100 from-card to-card placeholder:text-muted-foreground focus:scale-100 focus:border-primary focus:ring-0 md:text-sm';
	const btnPrimary =
		'rounded-md bg-primary bg-none px-4 text-sm font-medium text-primary-foreground shadow-none hover:bg-primary/90 hover:bg-none';
	const btnSecondary =
		'rounded-md border border-border bg-card bg-none text-foreground shadow-none backdrop-blur-none hover:bg-muted hover:bg-none';

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

<div
	class="liquid-panel overflow-hidden rounded-lg"
	data-animate="fade-up"
	style="animation-delay:180ms"
>
	<div class="border-border space-y-1 border-b px-6 py-5">
		<h3 class="text-foreground font-display text-lg font-semibold"
			>{editing
				? $t('catalog.locations.form.update_title')
				: $t('catalog.locations.form.create_title')}</h3
		>
	</div>
	<div class="p-6">
		<form method="POST" use:enhance class="space-y-4">
			<div class="space-y-1">
				<Input
					class={inputClass}
					placeholder={$t('catalog.locations.form.name')}
					bind:value={$form.name}
				/>
				{#if $errors.name}
					<p class="text-danger text-xs font-medium">{$errors.name}</p>
				{/if}
			</div>
			<div class="space-y-1">
				<Input
					class={inputClass}
					placeholder={$t('catalog.locations.form.address')}
					bind:value={$form.address}
				/>
				{#if $errors.address}
					<p class="text-danger text-xs font-medium">{$errors.address}</p>
				{/if}
			</div>
			<div class="flex flex-col gap-3 pt-1 sm:flex-row">
				<Button type="submit" class="{btnPrimary} w-full py-2.5 sm:w-1/2"
					>{editing ? $t('catalog.common.update') : $t('catalog.common.create')}</Button
				>
				<Button type="button" class="{btnSecondary} w-full py-2.5 sm:w-1/2" onclick={() => reset()}
					>{$t('catalog.common.reset')}</Button
				>
			</div>
		</form>
	</div>
</div>
