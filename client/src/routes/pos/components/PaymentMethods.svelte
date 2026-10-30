<script lang="ts">
	import { Label } from '$lib/components/ui/label';
	import Kbd from '$lib/components/ui/kbd.svelte';
	import {
		CreditCard,
		Banknote,
		QrCode,
		Wallet,
		Receipt,
		Check,
		AlertTriangle
	} from 'lucide-svelte';
	import { settings } from '$lib/stores/settings';
	import { formatCurrency } from '$lib/utils';
	import { t } from '$lib/i18n';

	type Props = {
		paymentMethod: string | null;
		selectedCustomer: any | null;
		pointsToRedeem: number;
		dueDate: string;
		onSelectPayment: (method: string) => void;
	};

	let {
		paymentMethod,
		selectedCustomer,
		pointsToRedeem = $bindable(),
		dueDate = $bindable(),
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
			id: 'CREDIT',
			label: $t('pos.payment.methods.credit'),
			icon: Receipt,
			color: 'text-warning',
			sub: $t('pos.payment.sub.credit'),
			shortcut: null
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

<div class="panel overflow-hidden" data-animate="fade-up" style="animation-delay:200ms">
	<div class="panel-header">
		<div>
			<h3 class="section-title">{$t('pos.payment.title')}</h3>
			<p class="mt-0.5 text-xs text-muted-foreground">
				{$t('pos.payment.description')}
			</p>
		</div>
	</div>
	<div class="p-4">
		<div class="grid grid-cols-2 gap-3 sm:grid-cols-3">
			{#each methods as method}
				<button
					type="button"
					onclick={() => onSelectPayment(method.id)}
					class={`group relative flex flex-col items-center justify-center gap-2 rounded-lg border p-3 transition-all duration-200 ${
						paymentMethod === method.id
							? 'border-primary/45 bg-accent shadow-xs'
							: 'border-border bg-card hover:border-primary/35 hover:bg-muted/60'
					}`}
				>
					{#if paymentMethod === method.id}
						<div
							class="absolute top-2 right-2 rounded-full bg-primary p-0.5 text-primary-foreground"
						>
							<Check class="h-3 w-3" />
						</div>
					{/if}

					{#if method.shortcut}
						<Kbd class="absolute top-1.5 left-1.5 border-border bg-card text-muted-foreground">
							{method.shortcut}
						</Kbd>
					{/if}

					<div
						class={`rounded-lg p-2.5 transition-colors ${
							paymentMethod === method.id
								? 'bg-primary text-primary-foreground'
								: 'bg-muted ' + method.color
						}`}
					>
						<svelte:component this={method.icon} class="h-5 w-5" />
					</div>

					<div class="text-center">
						<span
							class={`block text-xs font-semibold ${
								paymentMethod === method.id ? 'text-accent-foreground' : 'text-foreground'
							}`}
						>
							{method.label}
						</span>
						<span
							class={`block text-[0.65rem] ${
								paymentMethod === method.id ? 'text-accent-foreground/70' : 'text-muted-foreground'
							}`}
						>
							{method.sub}
						</span>
					</div>
				</button>
			{/each}
		</div>

		<!-- Credit (unpaid) sale options -->
		{#if paymentMethod === 'CREDIT'}
			<div class="mt-4 border-t border-border/80 pt-4">
				{#if !selectedCustomer}
					<div
						class="flex items-start gap-2 rounded-lg border border-warning/30 bg-warning/10 px-3 py-2.5 text-xs"
					>
						<AlertTriangle class="mt-0.5 h-4 w-4 shrink-0 text-warning" />
						<span class="font-medium text-foreground"
							>{$t('pos.payment.credit.customer_required')}</span
						>
					</div>
				{/if}

				<div class="mt-3 flex items-center justify-between gap-3">
					<Label class="text-xs font-medium text-muted-foreground">
						{$t('pos.payment.credit.due_date')}
					</Label>
					<input
						type="date"
						bind:value={dueDate}
						class="glass-input num-mono h-9 rounded-lg px-2 text-right text-sm"
					/>
				</div>
			</div>
		{/if}

		<!-- Loyalty Redemption -->
		{#if selectedCustomer && selectedCustomer.loyalty && selectedCustomer.loyalty.Points > 0}
			<div class="mt-4 border-t border-border/80 pt-4">
				<div class="flex items-center justify-between gap-2">
					<Label class="text-xs font-medium text-muted-foreground"
						>{$t('pos.loyalty.redeem_label')}</Label
					>
					<span class="text-xs font-medium text-success">
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
						class="glass-input num-mono h-9 w-24 rounded-lg px-2 text-right text-sm"
					/>
					<span class="text-xs text-muted-foreground">{$t('pos.loyalty.points')}</span>

					{#if pointsToRedeem > 0}
						<span class="num-mono ml-auto text-sm font-semibold text-success">
							- {formatCurrency(
								pointsToRedeem * ($settings.loyalty_points_redemption_rate || 0.01)
							)}
						</span>
					{/if}
				</div>
				{#if pointsToRedeem > selectedCustomer.loyalty.Points}
					<p class="mt-1 text-[0.65rem] text-danger">{$t('pos.loyalty.error_exceed')}</p>
				{/if}
			</div>
		{/if}
	</div>
</div>
