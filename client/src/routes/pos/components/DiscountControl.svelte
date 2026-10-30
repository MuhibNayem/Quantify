<script lang="ts">
	import { Label } from '$lib/components/ui/label';
	import { Percent, BadgePercent, X } from 'lucide-svelte';
	import { formatCurrency } from '$lib/utils';
	import { t } from '$lib/i18n';

	type Props = {
		subtotal: number;
		maxPercent: number;
		maxAmount: number;
		requireReason: boolean;
		amount: number;
		valueInvalid: boolean;
		reasonMissing: boolean;
		discountType: string;
		discountValue: number;
		discountReason: string;
	};

	let {
		subtotal,
		maxPercent,
		maxAmount,
		requireReason,
		amount,
		valueInvalid,
		reasonMissing,
		discountType = $bindable(),
		discountValue = $bindable(),
		discountReason = $bindable()
	}: Props = $props();

	const inputClass = 'glass-input h-9 w-full rounded-lg px-3 text-sm font-normal';

	const segments = $derived([
		{ id: 'percent', label: $t('pos.discount.type_percent'), icon: Percent },
		{ id: 'fixed', label: $t('pos.discount.type_fixed'), icon: BadgePercent }
	]);

	function clearDiscount() {
		discountValue = 0;
		discountReason = '';
	}
</script>

<div class="panel overflow-hidden" data-animate="fade-up" style="animation-delay:220ms">
	<div class="panel-header">
		<div>
			<h3 class="section-title">
				{$t('pos.discount.title')}
			</h3>
			<p class="mt-0.5 text-xs text-muted-foreground">{$t('pos.discount.description')}</p>
		</div>
		{#if discountValue > 0 || discountReason}
			<button
				type="button"
				onclick={clearDiscount}
				class="inline-flex items-center gap-1 rounded-lg px-2 py-1 text-[0.7rem] font-medium text-muted-foreground transition-colors hover:bg-danger/10 hover:text-danger"
			>
				<X class="h-3 w-3" />
				{$t('pos.discount.clear_btn')}
			</button>
		{/if}
	</div>

	<div class="space-y-3 p-4">
		<div class="flex rounded-lg bg-muted p-0.5">
			{#each segments as seg}
				<button
					type="button"
					onclick={() => (discountType = seg.id)}
					class={`flex flex-1 items-center justify-center gap-1.5 rounded-md px-3 py-1.5 text-xs font-semibold transition-all ${
						discountType === seg.id
							? 'bg-card text-foreground shadow-xs'
							: 'text-muted-foreground hover:text-foreground'
					}`}
				>
					<seg.icon class="h-3.5 w-3.5" />
					{seg.label}
				</button>
			{/each}
		</div>

		<div>
			<div class="flex items-baseline justify-between gap-2">
				<Label class="text-xs font-medium text-muted-foreground">
					{$t('pos.discount.value_label')}
				</Label>
				<span class="num-mono text-[0.65rem] text-muted-foreground">
					{#if discountType === 'percent'}
						{$t('pos.discount.max_percent_hint', { max: maxPercent })}
					{:else if maxAmount > 0}
						{$t('pos.discount.max_amount_hint', { max: formatCurrency(maxAmount) })}
					{:else}
						{$t('pos.discount.no_cap_hint')}
					{/if}
				</span>
			</div>
			<input
				type="number"
				min="0"
				step="any"
				bind:value={discountValue}
				placeholder="0"
				class="{inputClass} num-mono mt-1 text-right {valueInvalid
					? 'border-danger focus:border-danger'
					: ''}"
			/>
			{#if valueInvalid}
				<p class="mt-1 text-[0.65rem] text-danger">
					{#if discountType === 'percent'}
						{$t('pos.discount.error_exceed_percent', { max: maxPercent })}
					{:else}
						{$t('pos.discount.error_exceed_amount', { max: formatCurrency(maxAmount) })}
					{/if}
				</p>
			{/if}
		</div>

		{#if requireReason}
			<div>
				<Label class="text-xs font-medium text-muted-foreground">
					{$t('pos.discount.reason_label')}
				</Label>
				<input
					type="text"
					bind:value={discountReason}
					placeholder={$t('pos.discount.reason_placeholder')}
					class="{inputClass} mt-1 {reasonMissing ? 'border-danger focus:border-danger' : ''}"
				/>
				{#if reasonMissing}
					<p class="mt-1 text-[0.65rem] text-danger">
						{$t('pos.discount.error_reason_required')}
					</p>
				{/if}
			</div>
		{/if}

		{#if amount > 0}
			<div class="flex items-center justify-between border-t border-border/80 pt-2 text-xs">
				<span class="font-medium text-muted-foreground">{$t('pos.discount.preview')}</span>
				<span class="num-mono font-semibold text-success">− {formatCurrency(amount)}</span>
			</div>
		{/if}
	</div>
</div>
