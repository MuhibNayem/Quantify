<script lang="ts">
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
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

	const inputClass = 'glass-input h-10 w-full rounded-lg px-3 text-sm font-normal';
	const dialogClass = 'overflow-hidden rounded-xl bg-card p-0 backdrop-blur-none';
	const labelClass = 'text-muted-foreground mb-1.5 block text-xs font-medium';

	const saveNewCustomer = async () => {
		const ok = await onCreateCustomer(newCustomerForm);
		if (ok) {
			isNewCustomerModalOpen = false;
			newCustomerForm = { name: '', email: '', phone: '' };
		}
	};
</script>

<div class="panel overflow-hidden" data-animate="fade-up" style="animation-delay:160ms">
	<div class="panel-header">
		<div>
			<h3 class="section-title">{$t('pos.customer.title')}</h3>
			<p class="mt-0.5 text-xs text-muted-foreground">
				{$t('pos.customer.description')}
			</p>
		</div>
	</div>
	<div class="space-y-3 p-4">
		<div class="flex items-center gap-2">
			<div class="relative flex-1">
				<Search class="absolute top-1/2 left-3 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
				<Input
					bind:value={customerSearchTerm}
					placeholder={$t('pos.customer.search_placeholder')}
					class="{inputClass} pl-9"
					onkeydown={(e) => e.key === 'Enter' && onSearchCustomer()}
				/>
			</div>
			<Button variant="outline" size="icon" class="h-10 w-10 rounded-lg" onclick={onSearchCustomer}>
				<Search class="h-4 w-4" />
			</Button>
			<Button
				variant="secondary"
				class="h-10 gap-1.5 rounded-lg px-3 text-xs font-medium"
				onclick={() => (isNewCustomerModalOpen = true)}
			>
				<UserPlus class="h-4 w-4" />
				<span>{$t('pos.customer.new_btn')}</span>
			</Button>
		</div>

		{#if selectedCustomer}
			<div
				class="flex flex-col gap-1 rounded-lg border border-success/25 bg-success/5 px-3 py-3 text-xs text-foreground"
			>
				<div class="flex items-center justify-between gap-2 font-medium">
					<span>
						{selectedCustomer.FirstName}
						{selectedCustomer.LastName}
					</span>
					<span class="badge-success px-1.5 py-0 text-[0.65rem]">
						{selectedCustomer.Username}
					</span>
				</div>
				{#if selectedCustomer.Email}
					<div class="text-muted-foreground">{selectedCustomer.Email}</div>
				{/if}
				{#if selectedCustomer.PhoneNumber}
					<div class="num-mono text-muted-foreground">{selectedCustomer.PhoneNumber}</div>
				{/if}
				{#if selectedCustomer.loyalty}
					<div class="mt-1 flex items-center gap-2 border-t border-success/20 pt-1.5">
						<span class="font-semibold text-success">
							{$t('pos.customer.loyalty_pts', { points: selectedCustomer.loyalty.Points })}
						</span>
						<span
							class="rounded-md bg-success/10 px-1.5 py-0.5 text-[0.65rem] font-semibold text-success uppercase"
						>
							{$t('pos.customer.tier', { tier: selectedCustomer.loyalty.Tier })}
						</span>
					</div>
				{/if}
			</div>
		{:else}
			<div class="text-xs text-muted-foreground">
				{$t('pos.customer.no_selected')}
			</div>
		{/if}
	</div>
</div>

<Dialog.Root bind:open={isNewCustomerModalOpen}>
	<Dialog.Content class="{dialogClass} sm:max-w-[425px]">
		<!-- Header -->
		<div class="border-b border-border/80 bg-muted/40 px-6 py-5">
			<Dialog.Title class="font-display text-lg font-semibold tracking-tight text-foreground">
				{$t('pos.new_customer_modal.title')}
			</Dialog.Title>
			<Dialog.Description class="mt-1 text-sm text-muted-foreground">
				{$t('pos.new_customer_modal.description')}
			</Dialog.Description>
		</div>

		<!-- Body -->
		<div class="space-y-4 px-6 py-6">
			{#key isNewCustomerModalOpen}
				<!-- Name Field -->
				<div class="group space-y-1.5">
					<label for="name" class={labelClass}>
						{$t('pos.new_customer_modal.name_label')}
					</label>
					<div class="relative">
						<div class="pointer-events-none absolute inset-y-0 left-3 flex items-center">
							<Users class="h-4 w-4 text-muted-foreground group-focus-within:text-primary" />
						</div>
						<Input
							id="name"
							bind:value={newCustomerForm.name}
							placeholder={$t('pos.new_customer_modal.name_placeholder')}
							class="{inputClass} pl-10"
						/>
					</div>
				</div>

				<!-- Email Field -->
				<div class="group space-y-1.5">
					<label for="email" class={labelClass}>
						{$t('pos.new_customer_modal.email_label')}
					</label>
					<div class="relative">
						<div class="pointer-events-none absolute inset-y-0 left-3 flex items-center">
							<Mail class="h-4 w-4 text-muted-foreground group-focus-within:text-primary" />
						</div>
						<Input
							id="email"
							type="email"
							bind:value={newCustomerForm.email}
							placeholder={$t('pos.new_customer_modal.email_placeholder')}
							class="{inputClass} pl-10"
						/>
					</div>
				</div>

				<!-- Phone Field -->
				<div class="group space-y-1.5">
					<label for="phone" class={labelClass}>
						{$t('pos.new_customer_modal.phone_label')}
					</label>
					<div class="relative">
						<div class="pointer-events-none absolute inset-y-0 left-3 flex items-center">
							<Phone class="h-4 w-4 text-muted-foreground group-focus-within:text-primary" />
						</div>
						<Input
							id="phone"
							bind:value={newCustomerForm.phone}
							placeholder={$t('pos.new_customer_modal.phone_placeholder')}
							class="{inputClass} pl-10"
						/>
					</div>
				</div>
			{/key}
		</div>

		<!-- Footer -->
		<div
			class="flex items-center justify-end gap-3 border-t border-border/80 bg-muted/40 px-6 py-4"
		>
			<Button
				variant="ghost"
				onclick={() => (isNewCustomerModalOpen = false)}
				class="h-10 rounded-lg px-4 text-muted-foreground hover:bg-muted hover:text-foreground"
			>
				{$t('pos.new_customer_modal.cancel_btn')}
			</Button>
			<Button onclick={saveNewCustomer} class="btn-soothe h-10 px-6 font-semibold">
				{$t('pos.new_customer_modal.create_btn')}
			</Button>
		</div>
	</Dialog.Content>
</Dialog.Root>
