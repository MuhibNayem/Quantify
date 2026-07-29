<script lang="ts">
	import { Button } from '$lib/components/ui/button';
	import {
		Card,
		CardContent,
		CardDescription,
		CardFooter,
		CardHeader,
		CardTitle
	} from '$lib/components/ui/card';
	import { Zap, Loader2 } from 'lucide-svelte';
	import Kbd from '$lib/components/ui/kbd.svelte';
	import { formatCurrency } from '$lib/utils';
	import { t } from '$lib/i18n';

	type Props = {
		subtotal: number;
		tax: number;
		total: number;
		taxRate: number;
		cartLength: number;
		paymentMethod: string | null;
		selectedCustomer: any | null;
		potentialPoints: number;
		canCompleteOrder: boolean;
		isProcessing: boolean;
		onCompleteOrder: () => void;
	};

	let {
		subtotal,
		tax,
		total,
		taxRate,
		cartLength,
		paymentMethod,
		selectedCustomer,
		potentialPoints,
		canCompleteOrder,
		isProcessing,
		onCompleteOrder
	}: Props = $props();
</script>

<Card
	class="relative overflow-hidden rounded-2xl border-0 bg-slate-900 text-slate-50 shadow-xl"
	data-animate="fade-up"
	style="animation-delay:240ms"
>
	<div
		class="absolute inset-0 bg-[radial-gradient(circle_at_top,_rgba(129,140,248,0.35),_transparent_55%),_radial-gradient(circle_at_bottom,_rgba(56,189,248,0.25),_transparent_55%)] opacity-80"
	/>
	<div class="relative">
		<CardHeader class="pb-3">
			<CardTitle class="text-[0.9rem] text-slate-50/90">{$t('pos.summary.title')}</CardTitle>
			<CardDescription class="text-[0.7rem] text-slate-300">
				{$t('pos.summary.description')}
			</CardDescription>
		</CardHeader>
		<CardContent class="space-y-2 pt-0 text-[0.8rem]">
			<div class="flex justify-between text-slate-200">
				<span>{$t('pos.summary.subtotal')}</span>
				<span>{formatCurrency(subtotal)}</span>
			</div>
			<div class="flex justify-between text-[0.75rem] text-slate-300/90">
				<span>{$t('pos.summary.tax', { rate: Number((taxRate * 100).toFixed(2)) })}</span>
				<span>{formatCurrency(tax)}</span>
			</div>
			<div class="flex items-center justify-between border-t border-slate-700/70 pt-2">
				<span class="text-[0.85rem] font-medium text-slate-100">{$t('pos.summary.total')}</span>
				<span class="text-lg font-semibold tracking-tight text-slate-50">
					{formatCurrency(total)}
				</span>
			</div>

			<div class="flex items-center justify-between pt-1 text-[0.7rem] text-slate-300/90">
				<div>
					<span class="font-medium">{$t('pos.summary.payment')}</span>
					<span class="ml-1">
						{paymentMethod ? paymentMethod : $t('pos.summary.not_selected')}
					</span>
				</div>
				<div>
					<span class="font-medium">{$t('pos.summary.items')}</span>
					<span class="ml-1">{cartLength}</span>
				</div>
			</div>

			{#if selectedCustomer && potentialPoints > 0}
				<div
					class="flex items-center justify-between border-t border-slate-700/50 pt-2 text-[0.7rem] text-emerald-400"
				>
					<div class="flex items-center gap-1">
						<Zap class="h-3 w-3" />
						<span class="font-medium">{$t('pos.summary.loyalty_earnings')}</span>
					</div>
					<span class="font-bold">+{potentialPoints} pt</span>
				</div>
			{/if}
		</CardContent>
		<CardFooter class="pt-3">
			<Button
				size="lg"
				class="h-11 w-full rounded-xl bg-gradient-to-r from-emerald-400 via-sky-400 to-indigo-500 text-sm font-semibold tracking-wide shadow-lg shadow-indigo-500/40 hover:from-emerald-400 hover:via-sky-400 hover:to-indigo-500 disabled:cursor-not-allowed disabled:opacity-60"
				onclick={onCompleteOrder}
				disabled={!canCompleteOrder || isProcessing}
			>
				{#if isProcessing}
					<Loader2 class="mr-2 h-4 w-4 animate-spin" />
					{$t('pos.summary.processing_btn')}
				{:else if !cartLength}
					{$t('pos.summary.add_items_hint')}
				{:else if !paymentMethod}
					{$t('pos.summary.select_payment_hint')}
				{:else}
					{$t('pos.summary.complete_btn')}
				{/if}
				<Kbd class="ml-2 border-white/25 bg-white/15 text-white">F8</Kbd>
			</Button>
		</CardFooter>
	</div>
</Card>
