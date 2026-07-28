<script lang="ts">
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import {
		Card,
		CardContent,
		CardDescription,
		CardHeader,
		CardTitle
	} from '$lib/components/ui/card';
	import { Search, UserPlus, Users, Mail, Phone } from 'lucide-svelte';
	import * as Dialog from '$lib/components/ui/dialog';
	import { t } from '$lib/i18n';

	type Props = {
		selectedCustomer: any | null;
		customerSearchTerm: string;
		onSearchCustomer: () => void;
		onCreateCustomer: (form: { name: string; email: string; phone: string }) => Promise<boolean>;
	};

	let {
		selectedCustomer,
		customerSearchTerm = $bindable(),
		onSearchCustomer,
		onCreateCustomer
	}: Props = $props();

	let isNewCustomerModalOpen = $state(false);
	let newCustomerForm = $state({ name: '', email: '', phone: '' });

	const saveNewCustomer = async () => {
		const ok = await onCreateCustomer(newCustomerForm);
		if (ok) {
			isNewCustomerModalOpen = false;
			newCustomerForm = { name: '', email: '', phone: '' };
		}
	};
</script>

<Card
	class="overflow-hidden rounded-2xl border-0 bg-white/90 shadow-lg backdrop-blur-xl transition-all duration-300 hover:scale-[1.01] hover:shadow-xl"
	data-animate="fade-up"
	style="animation-delay:160ms"
>
	<CardHeader class="border-b border-slate-100/80 bg-white/80 pb-3">
		<CardTitle class="text-sm text-slate-900">{$t('pos.customer.title')}</CardTitle>
		<CardDescription class="text-[0.75rem] text-slate-500">
			{$t('pos.customer.description')}
		</CardDescription>
	</CardHeader>
	<CardContent class="space-y-3 pt-3">
		<div class="flex items-center gap-2">
			<div class="relative flex-1">
				<Search class="absolute top-1/2 left-2.5 h-4 w-4 -translate-y-1/2 text-slate-400" />
				<Input
					bind:value={customerSearchTerm}
					placeholder={$t('pos.customer.search_placeholder')}
					class="rounded-xl border-slate-200 bg-slate-50/80 pl-8 text-sm focus-visible:ring-emerald-300"
					onkeydown={(e) => e.key === 'Enter' && onSearchCustomer()}
				/>
			</div>
			<Button
				variant="outline"
				size="icon"
				class="rounded-xl border-slate-200 bg-white/80 hover:bg-slate-50"
				onclick={onSearchCustomer}
			>
				<Search class="h-4 w-4" />
			</Button>
			<Button
				variant="secondary"
				class="flex items-center gap-1 rounded-xl border border-indigo-100 bg-indigo-50 px-3 py-2 text-[0.75rem] text-indigo-700 hover:bg-indigo-100"
				onclick={() => (isNewCustomerModalOpen = true)}
			>
				<UserPlus class="h-4 w-4" />
				<span class="font-medium">{$t('pos.customer.new_btn')}</span>
			</Button>
		</div>

		{#if selectedCustomer}
			<div
				class="flex flex-col gap-0.5 rounded-xl border border-emerald-100 bg-emerald-50/70 px-3 py-2.5 text-[0.75rem] text-slate-700"
			>
				<div class="flex items-center justify-between gap-2 font-medium text-slate-900">
					<span>
						{selectedCustomer.FirstName}
						{selectedCustomer.LastName}
					</span>
					<span
						class="rounded-full border border-emerald-100 bg-white/80 px-1.5 py-0.5 text-[0.65rem] text-emerald-600"
					>
						{selectedCustomer.Username}
					</span>
				</div>
				{#if selectedCustomer.Email}
					<div>{selectedCustomer.Email}</div>
				{/if}
				{#if selectedCustomer.PhoneNumber}
					<div>{selectedCustomer.PhoneNumber}</div>
				{/if}
				{#if selectedCustomer.loyalty}
					<div class="mt-1 flex items-center gap-2 border-t border-emerald-100/50 pt-1">
						<span class="font-semibold text-emerald-700">
							{$t('pos.customer.loyalty_pts', { points: selectedCustomer.loyalty.Points })}
						</span>
						<span class="rounded-sm bg-emerald-100 px-1 text-[0.65rem] text-emerald-800 uppercase">
							{$t('pos.customer.tier', { tier: selectedCustomer.loyalty.Tier })}
						</span>
					</div>
				{/if}
			</div>
		{:else}
			<div class="text-[0.75rem] text-slate-400">
				{$t('pos.customer.no_selected')}
			</div>
		{/if}
	</CardContent>
</Card>

<Dialog.Root bind:open={isNewCustomerModalOpen}>
	<Dialog.Content
		class="overflow-hidden rounded-3xl border-0 bg-white/95 p-0 shadow-2xl backdrop-blur-xl transition-all duration-300 sm:max-w-[425px]"
	>
		<!-- Header with gradient and pattern -->
		<div
			class="relative overflow-hidden bg-gradient-to-br from-violet-600 to-indigo-600 px-6 py-6 sm:px-10 sm:py-10"
		>
			<div class="absolute -top-10 -right-10 h-40 w-40 rounded-full bg-white/10 blur-3xl"></div>
			<div
				class="absolute right-0 bottom-0 h-32 w-32 translate-x-12 translate-y-12 rounded-full bg-indigo-500/30 blur-2xl"
			></div>

			<div class="relative z-10">
				<Dialog.Title class="text-2xl font-bold tracking-tight text-white">
					{$t('pos.new_customer_modal.title')}
				</Dialog.Title>
				<Dialog.Description class="mt-1.5 text-violet-100">
					{$t('pos.new_customer_modal.description')}
				</Dialog.Description>
			</div>
		</div>

		<!-- Body -->
		<div class="space-y-6 px-6 py-8 sm:px-10">
			{#key isNewCustomerModalOpen}
				<!-- Name Field -->
				<div class="group space-y-2">
					<label
						for="name"
						class="ml-1 text-xs font-semibold tracking-wider text-slate-500 uppercase transition-colors group-focus-within:text-violet-600"
					>
						{$t('pos.new_customer_modal.name_label')}
					</label>
					<div class="relative">
						<div class="pointer-events-none absolute inset-y-0 left-0 flex items-center pl-3">
							<Users class="h-4 w-4 text-slate-400 group-focus-within:text-violet-500" />
						</div>
						<Input
							id="name"
							bind:value={newCustomerForm.name}
							placeholder={$t('pos.new_customer_modal.name_placeholder')}
							class="rounded-xl border-slate-200 bg-slate-50 pl-10 transition-all duration-300 group-hover:border-violet-300 focus:border-violet-500 focus:bg-white focus:ring-4 focus:ring-violet-500/10"
						/>
					</div>
				</div>

				<!-- Email Field -->
				<div class="group space-y-2">
					<label
						for="email"
						class="ml-1 text-xs font-semibold tracking-wider text-slate-500 uppercase transition-colors group-focus-within:text-violet-600"
					>
						{$t('pos.new_customer_modal.email_label')}
					</label>
					<div class="relative">
						<div class="pointer-events-none absolute inset-y-0 left-0 flex items-center pl-3">
							<Mail class="h-4 w-4 text-slate-400 group-focus-within:text-violet-500" />
						</div>
						<Input
							id="email"
							type="email"
							bind:value={newCustomerForm.email}
							placeholder={$t('pos.new_customer_modal.email_placeholder')}
							class="rounded-xl border-slate-200 bg-slate-50 pl-10 transition-all duration-300 group-hover:border-violet-300 focus:border-violet-500 focus:bg-white focus:ring-4 focus:ring-violet-500/10"
						/>
					</div>
				</div>

				<!-- Phone Field -->
				<div class="group space-y-2">
					<label
						for="phone"
						class="ml-1 text-xs font-semibold tracking-wider text-slate-500 uppercase transition-colors group-focus-within:text-violet-600"
					>
						{$t('pos.new_customer_modal.phone_label')}
					</label>
					<div class="relative">
						<div class="pointer-events-none absolute inset-y-0 left-0 flex items-center pl-3">
							<Phone class="h-4 w-4 text-slate-400 group-focus-within:text-violet-500" />
						</div>
						<Input
							id="phone"
							bind:value={newCustomerForm.phone}
							placeholder={$t('pos.new_customer_modal.phone_placeholder')}
							class="rounded-xl border-slate-200 bg-slate-50 pl-10 transition-all duration-300 group-hover:border-violet-300 focus:border-violet-500 focus:bg-white focus:ring-4 focus:ring-violet-500/10"
						/>
					</div>
				</div>
			{/key}
		</div>

		<!-- Footer -->
		<div
			class="flex items-center justify-end gap-3 bg-slate-50/50 px-6 py-4 backdrop-blur sm:px-10"
		>
			<Button
				variant="ghost"
				onclick={() => (isNewCustomerModalOpen = false)}
				class="rounded-xl text-slate-600 hover:bg-slate-100 hover:text-slate-900"
			>
				{$t('pos.new_customer_modal.cancel_btn')}
			</Button>
			<Button
				onclick={saveNewCustomer}
				class="rounded-xl bg-gradient-to-r from-violet-600 to-indigo-600 px-8 font-semibold text-white shadow-lg shadow-violet-500/25 transition-all duration-300 hover:scale-[1.02] hover:shadow-xl hover:shadow-violet-500/35 active:scale-[0.98]"
			>
				{$t('pos.new_customer_modal.create_btn')}
			</Button>
		</div>
	</Dialog.Content>
</Dialog.Root>
