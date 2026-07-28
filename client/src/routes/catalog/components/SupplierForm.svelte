<script lang="ts">
	import { superForm } from 'sveltekit-superforms';
	import { zod4 } from 'sveltekit-superforms/adapters';
	import { supplierSchema } from '$lib/schemas/supplier';
	import { suppliersApi } from '$lib/api/resources';
	import type { Supplier } from '$lib/types';
	import { toast } from 'svelte-sonner';
	import { t } from '$lib/i18n';
	import { Card, CardContent, CardHeader, CardTitle } from '$lib/components/ui/card';
	import { Input } from '$lib/components/ui/input';
	import { Button } from '$lib/components/ui/button';

	let { editing = null, onsaved }: { editing: Supplier | null; onsaved: () => void } = $props();

	const { form, errors, enhance, reset } = superForm(
		{ name: '', contactPerson: '', email: '', phone: '', address: '' },
		{
			validators: zod4(supplierSchema),
			onSubmit: async ({ cancel }) => {
				cancel();
				const data = $form;
				try {
					const payload = {
						name: data.name,
						contactPerson: data.contactPerson,
						email: data.email,
						phone: data.phone,
						address: data.address
					};
					if (editing) {
						await suppliersApi.update(editing.ID, payload);
					} else {
						await suppliersApi.create(payload);
					}
					toast.success($t('catalog.toasts.supplier_saved'));
					reset();
					onsaved();
				} catch (error) {
					const errorMessage = (error as any)?.response?.data?.error || 'Unable to save supplier';
					toast.error($t('catalog.toasts.supplier_save_fail'), { description: errorMessage });
				}
			}
		}
	);

	$effect(() => {
		if (editing) {
			$form.name = editing.Name;
			$form.contactPerson = editing.ContactPerson ?? '';
			$form.email = editing.Email ?? '';
			$form.phone = editing.Phone ?? '';
			$form.address = editing.Address ?? '';
		} else {
			reset();
		}
	});
</script>

<Card
	class="overflow-hidden rounded-2xl border-0 bg-gradient-to-br from-violet-50 to-purple-100 shadow-lg transition-all duration-300 hover:scale-[1.01] hover:shadow-xl"
	data-animate="fade-up"
	style="animation-delay:180ms"
>
	<CardHeader class="space-y-1 border-b border-white/60 bg-white/70 px-6 py-5 backdrop-blur">
		<CardTitle class="text-slate-800"
			>{editing
				? $t('catalog.suppliers.form.update_title')
				: $t('catalog.suppliers.form.create_title')}</CardTitle
		>
	</CardHeader>
	<CardContent class="p-6">
		<form method="POST" use:enhance class="space-y-4">
			<div class="space-y-1">
				<Input
					class="w-full rounded-xl border border-violet-200 bg-white/90 px-3.5 py-2.5 text-sm focus:ring-2 focus:ring-violet-400"
					placeholder={$t('catalog.suppliers.form.name')}
					bind:value={$form.name}
				/>
				{#if $errors.name}
					<p class="text-xs font-medium text-rose-600">{$errors.name}</p>
				{/if}
			</div>
			<Input
				class="w-full rounded-xl border border-violet-200 bg-white/90 px-3.5 py-2.5 text-sm focus:ring-2 focus:ring-violet-400"
				placeholder={$t('catalog.suppliers.form.contact_person')}
				bind:value={$form.contactPerson}
			/>
			<div class="space-y-1">
				<Input
					class="w-full rounded-xl border border-violet-200 bg-white/90 px-3.5 py-2.5 text-sm focus:ring-2 focus:ring-violet-400"
					placeholder={$t('catalog.suppliers.form.email')}
					bind:value={$form.email}
				/>
				{#if $errors.email}
					<p class="text-xs font-medium text-rose-600">{$errors.email}</p>
				{/if}
			</div>
			<Input
				class="w-full rounded-xl border border-violet-200 bg-white/90 px-3.5 py-2.5 text-sm focus:ring-2 focus:ring-violet-400"
				placeholder={$t('catalog.suppliers.form.phone')}
				bind:value={$form.phone}
			/>
			<Input
				class="w-full rounded-xl border border-violet-200 bg-white/90 px-3.5 py-2.5 text-sm focus:ring-2 focus:ring-violet-400"
				placeholder={$t('catalog.suppliers.form.address')}
				bind:value={$form.address}
			/>
			<div class="flex flex-col gap-3 pt-1 sm:flex-row">
				<Button
					type="submit"
					class="w-full rounded-xl bg-gradient-to-r from-violet-500 to-purple-600 py-2.5 text-white shadow-lg transition-all duration-300 hover:scale-105 hover:from-violet-600 hover:to-purple-700 hover:shadow-xl sm:w-1/2"
					>{editing ? $t('catalog.common.update') : $t('catalog.common.create')}</Button
				>
				<Button
					type="button"
					class="w-full rounded-xl border border-violet-200 py-2.5 text-violet-700 transition hover:bg-violet-50 sm:w-1/2"
					onclick={() => reset()}>{$t('catalog.common.reset')}</Button
				>
			</div>
		</form>
	</CardContent>
</Card>
