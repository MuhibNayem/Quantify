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
		'glass-input h-10 w-full px-3 text-sm text-foreground placeholder:text-muted-foreground';
	const btnPrimary = 'btn-soothe h-10 px-4 text-sm font-semibold';
	const btnSecondary = 'btn-soft h-10 px-4 text-sm font-medium';

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

<div class="panel overflow-hidden" data-animate="fade-up" style="animation-delay:180ms">
	<div class="panel-header flex-col items-start gap-1">
		<h3 class="section-title">
			{editing
				? $t('catalog.locations.form.update_title')
				: $t('catalog.locations.form.create_title')}
		</h3>
	</div>
	<div class="panel-body">
		<form method="POST" use:enhance class="space-y-4">
			<div class="space-y-1">
				<Input
					class={inputClass}
					placeholder={$t('catalog.locations.form.name')}
					bind:value={$form.name}
				/>
				{#if $errors.name}
					<p class="text-xs font-medium text-danger">{$errors.name}</p>
				{/if}
			</div>
			<div class="space-y-1">
				<Input
					class={inputClass}
					placeholder={$t('catalog.locations.form.address')}
					bind:value={$form.address}
				/>
				{#if $errors.address}
					<p class="text-xs font-medium text-danger">{$errors.address}</p>
				{/if}
			</div>
			<div class="flex flex-col gap-3 pt-1 sm:flex-row">
				<Button type="submit" class="{btnPrimary} w-full sm:flex-1"
					>{editing ? $t('catalog.common.update') : $t('catalog.common.create')}</Button
				>
				<Button
					type="button"
					variant="secondary"
					class="{btnSecondary} w-full sm:flex-1"
					onclick={() => reset()}>{$t('catalog.common.reset')}</Button
				>
			</div>
		</form>
	</div>
</div>
