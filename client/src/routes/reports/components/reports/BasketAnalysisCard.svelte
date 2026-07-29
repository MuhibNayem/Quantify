<script lang="ts">
	import * as Tooltip from '$lib/components/ui/tooltip';
	import { ShoppingCart, HelpCircle } from 'lucide-svelte';
	import type { BasketAnalysisItem } from '$lib/types';
	import { t } from '$lib/i18n';

	type Props = {
		basketAnalysis: BasketAnalysisItem[];
	};

	let { basketAnalysis }: Props = $props();
</script>

<div class="liquid-panel rounded-lg p-6">
	<h3 class="font-display text-foreground mb-4 flex items-center gap-2 text-lg font-semibold">
		<ShoppingCart class="text-primary h-5 w-5" />
		{$t('reports.frequency.title')}
		<Tooltip.Root>
			<Tooltip.Trigger>
				<HelpCircle class="text-muted-foreground hover:text-primary h-4 w-4 transition-colors" />
			</Tooltip.Trigger>
			<Tooltip.Content>
				<p class="max-w-xs">{$t('reports.frequency.description')}</p>
			</Tooltip.Content>
		</Tooltip.Root>
	</h3>
	<div class="custom-scrollbar max-h-[350px] space-y-2 overflow-y-auto p-1 pr-2">
		{#each basketAnalysis.slice(0, 10) as rule}
			<div
				class="group border-border bg-card hover:bg-muted rounded-md border p-3 transition-colors"
			>
				<div
					class="border-border bg-muted/40 text-muted-foreground mb-3 flex items-center justify-center gap-3 rounded-md border p-2 text-xs"
				>
					<span
						class="border-border bg-card text-foreground rounded-md border px-3 py-1 font-semibold"
					>
						{rule.ProductAName || $t('reports.frequency.unknown')}
					</span>
					<span class="text-primary text-lg font-bold">+</span>
					<span
						class="border-border bg-card text-foreground rounded-md border px-3 py-1 font-semibold"
					>
						{rule.ProductBName || $t('reports.frequency.unknown')}
					</span>
				</div>
				<div class="flex items-center justify-between px-1">
					<span class="text-muted-foreground text-[10px] num-mono"
						>ID: {rule.ProductA}-{rule.ProductB}</span
					>
					<div class="flex items-center gap-1.5">
						<span class="bg-primary h-1.5 w-1.5 animate-pulse rounded-full"></span>
						<span class="text-foreground text-xs font-semibold num-mono">
							{rule.Frequency}
							{$t('reports.frequency.orders')}
						</span>
					</div>
				</div>
			</div>
		{/each}
	</div>
</div>
