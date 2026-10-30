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

<div class="panel overflow-hidden" data-animate="fade-up" style="animation-delay:240ms">
	<div class="panel-header">
		<div>
			<h3 class="section-title">{$t('pos.summary.title')}</h3>
			<p class="mt-0.5 text-xs text-muted-foreground">
				{$t('pos.summary.description')}
			</p>
		</div>
	</div>

	<div class="space-y-2 p-4 text-[0.8125rem]">
		<div class="flex justify-between text-foreground">
			<span>{$t('pos.summary.subtotal')}</span>
			<span class="num-mono">{formatCurrency(subtotal)}</span>
		</div>
		{#if manualDiscount > 0}
			<div class="flex justify-between text-xs text-success">
				<span>{$t('pos.summary.manual_discount')}</span>
				<span class="num-mono">− {formatCurrency(manualDiscount)}</span>
			</div>
		{/if}
		<div class="flex justify-between text-xs text-muted-foreground">
			<span>{$t('pos.summary.tax', { rate: Number((taxRate * 100).toFixed(2)) })}</span>
			<span class="num-mono">{formatCurrency(tax)}</span>
		</div>
		<div class="flex items-center justify-between border-t border-border/80 pt-3">
			<span class="text-sm font-medium text-foreground">{$t('pos.summary.total')}</span>
			<span class="kpi-num text-2xl text-foreground">
				{formatCurrency(total)}
			</span>
		</div>

		<div class="flex items-center justify-between pt-1 text-[0.7rem] text-muted-foreground">
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
			<div class="flex items-center justify-between border-t border-border/80 pt-2 text-[0.7rem]">
				<div class="flex items-center gap-1 text-success">
					<Zap class="h-3 w-3" />
					<span class="font-medium">{$t('pos.summary.loyalty_earnings')}</span>
				</div>
				<span class="num-mono font-semibold text-success">+{potentialPoints} pt</span>
			</div>
		{/if}
	</div>

	<div class="px-4 pb-4">
		<Button
			size="lg"
			class="btn-soothe h-11 w-full text-sm font-semibold tracking-wide"
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
			<Kbd
				class="ml-2 border-primary-foreground/30 bg-primary-foreground/15 text-primary-foreground"
				>F8</Kbd
			>
		</Button>
	</div>
</div>
