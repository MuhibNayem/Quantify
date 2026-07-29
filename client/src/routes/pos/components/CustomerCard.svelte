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

	const inputClass =
		'liquid-input h-10 w-full rounded-md bg-card bg-none px-3 text-sm font-normal text-foreground shadow-none ring-0 backdrop-blur-none backdrop-saturate-100 from-card to-card placeholder:text-muted-foreground focus:scale-100 focus:border-primary focus:ring-0 md:text-sm';
	const dialogClass =
		'rounded-lg border-border bg-card bg-none shadow-lg backdrop-blur-none backdrop-saturate-100 from-card to-card';
	const labelClass =
		'ml-1 text-xs font-semibold tracking-wider text-muted-foreground uppercase transition-colors group-focus-within:text-primary';

	const saveNewCustomer = async () => {
		const ok = await onCreateCustomer(newCustomerForm);
		if (ok) {
			isNewCustomerModalOpen = false;
			newCustomerForm = { name: '', email: '', phone: '' };
		}
	};
</script>

<div
	class="liquid-panel overflow-hidden rounded-lg"
	data-animate="fade-up"
	style="animation-delay:160ms"
>
	<div class="border-border bg-muted/50 border-b px-4 py-3">
		<h3 class="font-display text-foreground text-sm font-semibold">{$t('pos.customer.title')}</h3>
		<p class="text-muted-foreground text-[0.75rem]">
			{$t('pos.customer.description')}
		</p>
	</div>
	<div class="space-y-3 p-4 pt-3">
		<div class="flex items-center gap-2">
			<div class="relative flex-1">
				<Search class="text-muted-foreground absolute top-1/2 left-2.5 h-4 w-4 -translate-y-1/2" />
				<Input
					bind:value={customerSearchTerm}
					placeholder={$t('pos.customer.search_placeholder')}
					class="{inputClass} pl-8"
					onkeydown={(e) => e.key === 'Enter' && onSearchCustomer()}
				/>
			</div>
			<Button
				variant="outline"
				size="icon"
				class="glass-button rounded-md border-border bg-card shadow-none backdrop-blur-none"
				onclick={onSearchCustomer}
			>
				<Search class="h-4 w-4" />
			</Button>
			<Button
				variant="secondary"
				class="border-primary/30 bg-accent text-accent-foreground hover:bg-accent/70 flex items-center gap-1 rounded-md border bg-none px-3 py-2 text-[0.75rem] shadow-none"
				onclick={() => (isNewCustomerModalOpen = true)}
			>
				<UserPlus class="h-4 w-4" />
				<span class="font-medium">{$t('pos.customer.new_btn')}</span>
			</Button>
		</div>

		{#if selectedCustomer}
			<div
				class="text-foreground flex flex-col gap-0.5 rounded-md border border-success/25 bg-success/10 px-3 py-2.5 text-[0.75rem]"
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
					<div class="text-muted-foreground num-mono">{selectedCustomer.PhoneNumber}</div>
				{/if}
				{#if selectedCustomer.loyalty}
					<div class="mt-1 flex items-center gap-2 border-t border-success/20 pt-1">
						<span class="text-success font-semibold">
							{$t('pos.customer.loyalty_pts', { points: selectedCustomer.loyalty.Points })}
						</span>
						<span class="text-success rounded-sm bg-success/15 px-1 text-[0.65rem] uppercase">
							{$t('pos.customer.tier', { tier: selectedCustomer.loyalty.Tier })}
						</span>
					</div>
				{/if}
			</div>
		{:else}
			<div class="text-muted-foreground text-[0.75rem]">
				{$t('pos.customer.no_selected')}
			</div>
		{/if}
	</div>
</div>

<Dialog.Root bind:open={isNewCustomerModalOpen}>
	<Dialog.Content class="{dialogClass} overflow-hidden p-0 sm:max-w-[425px]">
		<!-- Header -->
		<div class="bg-primary text-primary-foreground px-6 py-6 sm:px-10 sm:py-8">
			<Dialog.Title class="font-display text-2xl font-bold tracking-tight">
				{$t('pos.new_customer_modal.title')}
			</Dialog.Title>
			<Dialog.Description class="text-primary-foreground/80 mt-1.5">
				{$t('pos.new_customer_modal.description')}
			</Dialog.Description>
		</div>

		<!-- Body -->
		<div class="space-y-6 px-6 py-8 sm:px-10">
			{#key isNewCustomerModalOpen}
				<!-- Name Field -->
				<div class="group space-y-2">
					<label for="name" class={labelClass}>
						{$t('pos.new_customer_modal.name_label')}
					</label>
					<div class="relative">
						<div class="pointer-events-none absolute inset-y-0 left-0 flex items-center pl-3">
							<Users class="text-muted-foreground group-focus-within:text-primary h-4 w-4" />
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
				<div class="group space-y-2">
					<label for="email" class={labelClass}>
						{$t('pos.new_customer_modal.email_label')}
					</label>
					<div class="relative">
						<div class="pointer-events-none absolute inset-y-0 left-0 flex items-center pl-3">
							<Mail class="text-muted-foreground group-focus-within:text-primary h-4 w-4" />
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
				<div class="group space-y-2">
					<label for="phone" class={labelClass}>
						{$t('pos.new_customer_modal.phone_label')}
					</label>
					<div class="relative">
						<div class="pointer-events-none absolute inset-y-0 left-0 flex items-center pl-3">
							<Phone class="text-muted-foreground group-focus-within:text-primary h-4 w-4" />
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
			class="border-border bg-muted/50 flex items-center justify-end gap-3 border-t px-6 py-4 sm:px-10"
		>
			<Button
				variant="ghost"
				onclick={() => (isNewCustomerModalOpen = false)}
				class="text-muted-foreground hover:text-foreground hover:bg-muted rounded-md"
			>
				{$t('pos.new_customer_modal.cancel_btn')}
			</Button>
			<Button
				onclick={saveNewCustomer}
				class="bg-primary text-primary-foreground hover:bg-primary/90 rounded-md bg-none px-8 font-semibold shadow-sm hover:bg-none active:scale-[0.98]"
			>
				{$t('pos.new_customer_modal.create_btn')}
			</Button>
		</div>
	</Dialog.Content>
</Dialog.Root>
