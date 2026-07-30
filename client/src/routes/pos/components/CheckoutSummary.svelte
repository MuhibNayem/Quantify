<script lang="ts">
	import { Button } from '$lib/components/ui/button';
	import { Zap, Loader2 } from 'lucide-svelte';
	import Kbd from '$lib/components/ui/kbd.svelte';
	import { formatCurrency } from '$lib/utils';
	import { t } from '$lib/i18n';

	type Props = {
		subtotal: number;
		manualDiscount: number;
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
		manualDiscount,
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

<div
	class="liquid-panel overflow-hidden rounded-lg"
	data-animate="fade-up"
	style="animation-delay:240ms"
>
	<div class="border-border bg-muted/50 border-b px-4 py-3">
		<h3 class="font-display text-foreground text-sm font-semibold">{$t('pos.summary.title')}</h3>
		<p class="text-muted-foreground text-[0.7rem]">
			{$t('pos.summary.description')}
		</p>
	</div>

	<div class="space-y-2 p-4 text-[0.8rem]">
		<div class="text-foreground flex justify-between">
			<span>{$t('pos.summary.subtotal')}</span>
			<span class="num-mono">{formatCurrency(subtotal)}</span>
		</div>
		{#if manualDiscount > 0}
			<div class="text-success flex justify-between text-[0.75rem]">
				<span>{$t('pos.summary.manual_discount')}</span>
				<span class="num-mono">− {formatCurrency(manualDiscount)}</span>
			</div>
		{/if}
		<div class="text-muted-foreground flex justify-between text-[0.75rem]">
			<span>{$t('pos.summary.tax', { rate: Number((taxRate * 100).toFixed(2)) })}</span>
			<span class="num-mono">{formatCurrency(tax)}</span>
		</div>
		<div class="border-border flex items-center justify-between border-t pt-2">
			<span class="text-foreground text-[0.85rem] font-medium">{$t('pos.summary.total')}</span>
			<span class="kpi-num text-foreground text-2xl">
				{formatCurrency(total)}
			</span>
		</div>

		<div class="text-muted-foreground flex items-center justify-between pt-1 text-[0.7rem]">
			<div>
				<span class="font-medium">{$t('pos.summary.payment')}</span>
				<span class="ml-1">
					{paymentMethod ? paymentMethod : $t('pos.summary.not_selected')}
				</span>
			</div>
			<div>
				<span class="font-medium">{$t('pos.summary.items')}</span>
				<span class="num ml-1">{cartLength}</span>
			</div>
		</div>

		{#if selectedCustomer && potentialPoints > 0}
			<div
				class="border-border flex items-center justify-between border-t pt-2 text-[0.7rem]"
			>
				<div class="text-success flex items-center gap-1">
					<Zap class="h-3 w-3" />
					<span class="font-medium">{$t('pos.summary.loyalty_earnings')}</span>
				</div>
				<span class="num-mono text-success font-bold">+{potentialPoints} pt</span>
			</div>
		{/if}
	</div>

	<div class="px-4 pb-4">
		<Button
			size="lg"
			class="bg-primary text-primary-foreground hover:bg-primary/90 h-11 w-full rounded-md bg-none text-sm font-semibold tracking-wide shadow-sm hover:bg-none disabled:cursor-not-allowed disabled:opacity-60"
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
			<Kbd class="border-primary-foreground/30 bg-primary-foreground/15 text-primary-foreground ml-2">F8</Kbd>
		</Button>
	</div>
</div>
