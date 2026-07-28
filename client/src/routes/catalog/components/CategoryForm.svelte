<script lang="ts">
	import { superForm } from 'sveltekit-superforms';
	import { zod4 } from 'sveltekit-superforms/adapters';
	import { categorySchema } from '$lib/schemas/category';
	import { categoriesApi } from '$lib/api/resources';
	import type { Category } from '$lib/types';
	import { toast } from 'svelte-sonner';
	import { t } from '$lib/i18n';
	import { Card, CardContent, CardHeader, CardTitle } from '$lib/components/ui/card';
	import { Input } from '$lib/components/ui/input';
	import { Button } from '$lib/components/ui/button';

	let { editing = null, onsaved }: { editing: Category | null; onsaved: () => void } = $props();

	const { form, errors, enhance, reset } = superForm(
		{ name: '' },
		{
			validators: zod4(categorySchema),
			onSubmit: async ({ cancel }) => {
				cancel();
				const data = $form;
				try {
					if (editing) {
						await categoriesApi.update(editing.ID, { name: data.name });
					} else {
						await categoriesApi.create({ name: data.name });
					}
					toast.success($t('catalog.toasts.category_saved'));
					reset();
					onsaved();
				} catch (error) {
					const errorMessage = (error as any)?.response?.data?.error || 'Unable to save category';
					toast.error($t('catalog.toasts.category_save_fail'), { description: errorMessage });
				}
			}
		}
	);

	$effect(() => {
		if (editing) {
			$form.name = editing.Name;
		} else {
			reset();
		}
	});
</script>

<Card
	class="overflow-hidden rounded-2xl border-0 bg-gradient-to-br from-green-50 to-emerald-100 shadow-lg transition-all duration-300 hover:scale-[1.01] hover:shadow-xl"
	data-animate="fade-up"
	style="animation-delay:180ms"
>
	<CardHeader class="space-y-1 border-b border-white/60 bg-white/70 px-6 py-5 backdrop-blur">
		<CardTitle class="text-slate-800"
			>{editing
				? $t('catalog.categories.form.update_title')
				: $t('catalog.categories.form.create_title')}</CardTitle
		>
	</CardHeader>
	<CardContent class="p-6">
		<form method="POST" use:enhance class="space-y-4">
			<div class="space-y-1">
				<Input
					class="w-full rounded-xl border border-emerald-200 bg-white/90 px-3.5 py-2.5 text-sm focus:ring-2 focus:ring-emerald-400"
					placeholder={$t('catalog.categories.form.name')}
					bind:value={$form.name}
				/>
				{#if $errors.name}
					<p class="text-xs font-medium text-rose-600">{$errors.name}</p>
				{/if}
			</div>
			<div class="flex flex-col gap-3 pt-1 sm:flex-row">
				<Button
					type="submit"
					class="w-full rounded-xl bg-gradient-to-r from-emerald-500 to-green-600 py-2.5 text-white shadow-lg transition-all duration-300 hover:scale-105 hover:from-emerald-600 hover:to-green-700 hover:shadow-xl sm:w-1/2"
					>{editing ? $t('catalog.common.update') : $t('catalog.common.create')}</Button
				>
				<Button
					type="button"
					class="w-full rounded-xl border border-emerald-200 py-2.5 text-emerald-700 transition hover:bg-emerald-50 sm:w-1/2"
					onclick={() => reset()}>{$t('catalog.common.reset')}</Button
				>
			</div>
		</form>
	</CardContent>
</Card>
