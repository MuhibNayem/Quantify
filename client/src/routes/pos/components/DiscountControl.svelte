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

	const inputClass =
		'liquid-input h-9 w-full rounded-md bg-card bg-none px-3 text-sm font-normal text-foreground shadow-none ring-0 placeholder:text-muted-foreground focus:border-primary focus:ring-0';

	const segments = $derived([
		{ id: 'percent', label: $t('pos.discount.type_percent'), icon: Percent },
		{ id: 'fixed', label: $t('pos.discount.type_fixed'), icon: BadgePercent }
	]);

	function clearDiscount() {
		discountValue = 0;
		discountReason = '';
	}
</script>

<div
	class="liquid-panel overflow-hidden rounded-lg"
	data-animate="fade-up"
	style="animation-delay:220ms"
>
	<div class="border-border bg-muted/50 flex items-center justify-between border-b px-4 py-3">
		<div>
			<h3 class="font-display text-foreground text-sm font-semibold">
				{$t('pos.discount.title')}
			</h3>
			<p class="text-muted-foreground text-[0.75rem]">{$t('pos.discount.description')}</p>
		</div>
		{#if discountValue > 0 || discountReason}
			<button
				type="button"
				onclick={clearDiscount}
				class="text-muted-foreground hover:text-danger border-border bg-card hover:border-danger/40 inline-flex items-center gap-1 rounded-md border px-2 py-1 text-[0.7rem] font-medium transition-colors"
			>
				<X class="h-3 w-3" />
				{$t('pos.discount.clear_btn')}
			</button>
		{/if}
	</div>

	<div class="space-y-3 p-4 pt-3">
		<div class="border-border bg-muted flex rounded-md border p-0.5">
			{#each segments as seg}
				<button
					type="button"
					onclick={() => (discountType = seg.id)}
					class={`flex flex-1 items-center justify-center gap-1.5 rounded-md px-3 py-1.5 text-xs font-semibold transition-all ${
						discountType === seg.id
							? 'bg-card text-foreground border-border shadow-sm border'
							: 'text-muted-foreground hover:text-foreground border border-transparent'
					}`}
				>
					<seg.icon class="h-3.5 w-3.5" />
					{seg.label}
				</button>
			{/each}
		</div>

		<div>
			<div class="flex items-center justify-between">
				<Label class="text-muted-foreground text-[0.7rem] font-medium">
					{$t('pos.discount.value_label')}
				</Label>
				<span class="text-muted-foreground num-mono text-[0.65rem]">
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
				class="{inputClass} num-mono mt-1 text-right {valueInvalid ? 'border-danger focus:border-danger' : ''}"
			/>
			{#if valueInvalid}
				<p class="text-danger mt-1 text-[0.65rem]">
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
				<Label class="text-muted-foreground text-[0.7rem] font-medium">
					{$t('pos.discount.reason_label')}
				</Label>
				<input
					type="text"
					bind:value={discountReason}
					placeholder={$t('pos.discount.reason_placeholder')}
					class="{inputClass} mt-1 {reasonMissing ? 'border-danger focus:border-danger' : ''}"
				/>
				{#if reasonMissing}
					<p class="text-danger mt-1 text-[0.65rem]">
						{$t('pos.discount.error_reason_required')}
					</p>
				{/if}
			</div>
		{/if}

		{#if amount > 0}
			<div
				class="border-border flex items-center justify-between border-t pt-2 text-[0.75rem]"
			>
				<span class="text-muted-foreground font-medium">{$t('pos.discount.preview')}</span>
				<span class="num-mono text-success font-bold">− {formatCurrency(amount)}</span>
			</div>
		{/if}
	</div>
</div>
