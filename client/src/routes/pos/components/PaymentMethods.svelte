<script lang="ts">
	import { Label } from '$lib/components/ui/label';
	import Kbd from '$lib/components/ui/kbd.svelte';
	import { CreditCard, Banknote, QrCode, Wallet, Check } from 'lucide-svelte';
	import { settings } from '$lib/stores/settings';
	import { formatCurrency } from '$lib/utils';
	import { t } from '$lib/i18n';

	type Props = {
		paymentMethod: string | null;
		selectedCustomer: any | null;
		pointsToRedeem: number;
		onSelectPayment: (method: string) => void;
	};

	let {
		paymentMethod,
		selectedCustomer,
		pointsToRedeem = $bindable(),
		onSelectPayment
	}: Props = $props();

	const methods = $derived([
		{
			id: 'CASH',
			label: $t('pos.payment.methods.cash'),
			icon: Banknote,
			color: 'text-success',
			sub: $t('pos.payment.sub.physical'),
			shortcut: 'F4'
		},
		{
			id: 'CARD',
			label: $t('pos.payment.methods.card'),
			icon: CreditCard,
			color: 'text-primary',
			sub: $t('pos.payment.sub.terminal'),
			shortcut: 'F5'
		},
		{
			id: 'BKASH',
			label: $t('pos.payment.methods.bkash'),
			icon: QrCode,
			color: 'text-info',
			sub: $t('pos.payment.sub.mobile'),
			shortcut: 'F6'
		},
		{
			id: 'OTHER',
			label: $t('pos.payment.methods.other'),
			icon: Wallet,
			color: 'text-warning',
			sub: $t('pos.payment.sub.check_due'),
			shortcut: null
		}
	]);
</script>

<div
	class="liquid-panel overflow-hidden rounded-lg"
	data-animate="fade-up"
	style="animation-delay:200ms"
>
	<div class="border-border bg-muted/50 border-b px-4 py-3">
		<h3 class="font-display text-foreground text-sm font-semibold">{$t('pos.payment.title')}</h3>
		<p class="text-muted-foreground text-[0.75rem]">
			{$t('pos.payment.description')}
		</p>
	</div>
	<div class="p-4 pt-3">
		<div class="grid grid-cols-2 gap-3 sm:grid-cols-4">
			{#each methods as method}
				<button
					type="button"
					onclick={() => onSelectPayment(method.id)}
					class={`group relative flex flex-col items-center justify-center gap-2 rounded-lg border p-3 transition-all duration-200 hover:shadow-md ${
						paymentMethod === method.id
							? 'border-primary bg-primary text-primary-foreground shadow-md'
							: 'border-border bg-card text-muted-foreground hover:border-primary/40 hover:bg-muted'
					}`}
				>
				{#if paymentMethod === method.id}
					<div class="bg-primary-foreground/20 absolute top-2 right-2 rounded-md p-0.5">
						<Check class="text-primary-foreground h-3 w-3" />
					</div>
				{/if}

				{#if method.shortcut}
					<Kbd
						class="absolute top-1.5 left-1.5 {paymentMethod === method.id
							? 'border-primary-foreground/30 bg-primary-foreground/15 text-primary-foreground'
							: 'border-border bg-muted text-muted-foreground'}"
					>
						{method.shortcut}
					</Kbd>
				{/if}

					<div
						class={`rounded-md p-2.5 transition-colors ${
							paymentMethod === method.id
								? 'bg-primary-foreground/15 text-primary-foreground'
								: 'bg-muted ' + method.color
						}`}
					>
						<svelte:component this={method.icon} class="h-5 w-5" />
					</div>

					<div class="text-center">
						<span
							class={`block text-xs font-semibold ${
								paymentMethod === method.id ? 'text-primary-foreground' : 'text-foreground'
							}`}
						>
							{method.label}
						</span>
						<span
							class={`block text-[0.65rem] ${
								paymentMethod === method.id ? 'text-primary-foreground/75' : 'text-muted-foreground'
							}`}
						>
							{method.sub}
						</span>
					</div>
				</button>
			{/each}
		</div>

		<!-- Loyalty Redemption -->
		{#if selectedCustomer && selectedCustomer.loyalty && selectedCustomer.loyalty.Points > 0}
			<div class="border-border mt-4 border-t pt-4">
				<div class="flex items-center justify-between">
					<Label class="text-muted-foreground text-[0.7rem] font-medium"
						>{$t('pos.loyalty.redeem_label')}</Label
					>
					<span class="text-success text-[0.7rem] font-medium">
						{$t('pos.loyalty.available', {
							points: selectedCustomer.loyalty.Points,
							value: formatCurrency(
								selectedCustomer.loyalty.Points * ($settings.loyalty_points_redemption_rate || 0.01)
							)
						})}
					</span>
				</div>
				<div class="mt-2 flex items-center gap-3">
					<input
						type="number"
						min="0"
						max={selectedCustomer.loyalty.Points}
						bind:value={pointsToRedeem}
						class="liquid-input num-mono h-9 w-24 rounded-md px-2 text-right text-sm"
					/>
					<span class="text-muted-foreground text-xs">{$t('pos.loyalty.points')}</span>

					{#if pointsToRedeem > 0}
						<span class="text-success num-mono ml-auto text-sm font-semibold">
							- {formatCurrency(
								pointsToRedeem * ($settings.loyalty_points_redemption_rate || 0.01)
							)}
						</span>
					{/if}
				</div>
				{#if pointsToRedeem > selectedCustomer.loyalty.Points}
					<p class="text-danger mt-1 text-[0.65rem]">{$t('pos.loyalty.error_exceed')}</p>
				{/if}
			</div>
		{/if}
	</div>
</div>
