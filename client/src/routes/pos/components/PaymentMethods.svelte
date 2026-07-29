<script lang="ts">
	import { Input } from '$lib/components/ui/input';
	import {
		Card,
		CardContent,
		CardDescription,
		CardHeader,
		CardTitle
	} from '$lib/components/ui/card';
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
			color: 'text-emerald-500',
			sub: $t('pos.payment.sub.physical'),
			shortcut: 'F4'
		},
		{
			id: 'CARD',
			label: $t('pos.payment.methods.card'),
			icon: CreditCard,
			color: 'text-violet-500',
			sub: $t('pos.payment.sub.terminal'),
			shortcut: 'F5'
		},
		{
			id: 'BKASH',
			label: $t('pos.payment.methods.bkash'),
			icon: QrCode,
			color: 'text-pink-500',
			sub: $t('pos.payment.sub.mobile'),
			shortcut: 'F6'
		},
		{
			id: 'OTHER',
			label: $t('pos.payment.methods.other'),
			icon: Wallet,
			color: 'text-orange-500',
			sub: $t('pos.payment.sub.check_due'),
			shortcut: null
		}
	]);
</script>

<Card
	class="overflow-hidden rounded-2xl border-0 bg-white/90 shadow-lg backdrop-blur-xl transition-all duration-300 hover:scale-[1.01] hover:shadow-xl"
	data-animate="fade-up"
	style="animation-delay:200ms"
>
	<CardHeader class="border-b border-slate-100/80 bg-white/80 pb-3">
		<CardTitle class="text-sm text-slate-900">{$t('pos.payment.title')}</CardTitle>
		<CardDescription class="text-[0.75rem] text-slate-500">
			{$t('pos.payment.description')}
		</CardDescription>
	</CardHeader>
	<CardContent class="pt-3">
		<div class="grid grid-cols-2 gap-3 sm:grid-cols-4">
			{#each methods as method}
				<button
					type="button"
					onclick={() => onSelectPayment(method.id)}
					class={`group relative flex flex-col items-center justify-center gap-2 rounded-2xl border p-3 transition-all duration-200 hover:shadow-md ${
						paymentMethod === method.id
							? 'scale-[1.02] border-transparent bg-gradient-to-br from-indigo-600 to-violet-600 shadow-lg ring-2 shadow-indigo-500/25 ring-indigo-500/10'
							: 'border-slate-100 bg-white/50 text-slate-600 hover:border-indigo-100 hover:bg-white'
					}`}
				>
				{#if paymentMethod === method.id}
					<div class="absolute top-2 right-2 rounded-full bg-white/20 p-0.5">
						<Check class="h-3 w-3 text-white" />
					</div>
				{/if}

				{#if method.shortcut}
					<Kbd
						class="absolute top-1.5 left-1.5 {paymentMethod === method.id
							? 'border-white/25 bg-white/15 text-white'
							: 'border-slate-200 bg-slate-100/80 text-slate-400'}"
					>
						{method.shortcut}
					</Kbd>
				{/if}

					<div
						class={`rounded-xl p-2.5 transition-colors ${
							paymentMethod === method.id
								? 'bg-white/10 text-white'
								: 'bg-slate-50 ' + method.color + ' group-hover:bg-indigo-50/50'
						}`}
					>
						<svelte:component this={method.icon} class="h-5 w-5" />
					</div>

					<div class="text-center">
						<span
							class={`block text-xs font-semibold ${
								paymentMethod === method.id ? 'text-white' : 'text-slate-700'
							}`}
						>
							{method.label}
						</span>
						<span
							class={`block text-[0.65rem] ${
								paymentMethod === method.id ? 'text-indigo-100' : 'text-slate-400'
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
			<div class="mt-4 border-t border-slate-100/80 pt-4">
				<div class="flex items-center justify-between">
					<Label class="text-[0.7rem] font-medium text-slate-500"
						>{$t('pos.loyalty.redeem_label')}</Label
					>
					<span class="text-[0.7rem] font-medium text-emerald-600">
						{$t('pos.loyalty.available', {
							points: selectedCustomer.loyalty.Points,
							value: formatCurrency(
								selectedCustomer.loyalty.Points * ($settings.loyalty_points_redemption_rate || 0.01)
							)
						})}
					</span>
				</div>
				<div class="mt-2 flex items-center gap-3">
					<Input
						type="number"
						min="0"
						max={selectedCustomer.loyalty.Points}
						bind:value={pointsToRedeem}
						class="h-9 w-24 rounded-lg border-slate-200 bg-slate-50 text-right text-sm"
					/>
					<span class="text-xs text-slate-400">{$t('pos.loyalty.points')}</span>

					{#if pointsToRedeem > 0}
						<span class="ml-auto text-sm font-semibold text-emerald-500">
							- {formatCurrency(
								pointsToRedeem * ($settings.loyalty_points_redemption_rate || 0.01)
							)}
						</span>
					{/if}
				</div>
				{#if pointsToRedeem > selectedCustomer.loyalty.Points}
					<p class="mt-1 text-[0.65rem] text-rose-500">{$t('pos.loyalty.error_exceed')}</p>
				{/if}
			</div>
		{/if}
	</CardContent>
</Card>
