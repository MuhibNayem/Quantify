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

<div class="panel p-5">
	<div class="flex items-start gap-3">
		<span
			class="flex h-9 w-9 shrink-0 items-center justify-center rounded-[calc(var(--radius)_-_4px)] border border-border bg-primary/10 text-primary"
		>
			<ShoppingCart class="h-4 w-4" />
		</span>
		<div class="min-w-0">
			<h3 class="section-title flex items-center gap-1.5">
				{$t('reports.frequency.title')}
				<Tooltip.Root>
					<Tooltip.Trigger>
						<HelpCircle
							class="h-4 w-4 text-muted-foreground transition-colors hover:text-primary"
						/>
					</Tooltip.Trigger>
					<Tooltip.Content>
						<p class="max-w-xs font-normal text-foreground">
							{$t('reports.frequency.description')}
						</p>
					</Tooltip.Content>
				</Tooltip.Root>
			</h3>
		</div>
	</div>

	<div class="custom-scrollbar mt-4 max-h-[360px] space-y-2 overflow-y-auto pr-1">
		{#if basketAnalysis.length === 0}
			<p class="py-8 text-center text-xs text-muted-foreground">
				{$t('common.no_data_available')}
			</p>
		{/if}
		{#each basketAnalysis.slice(0, 10) as rule}
			<div
				class="rounded-[calc(var(--radius)_-_6px)] border border-border/70 bg-muted/30 p-3 transition-colors"
			>
				<div
					class="flex items-center justify-center gap-2 rounded-[calc(var(--radius)_-_8px)] border border-border/70 bg-card px-2 py-2 text-xs"
				>
					<span
						class="max-w-[45%] truncate rounded-md border border-border bg-muted/60 px-2.5 py-1 text-xs font-medium text-foreground"
					>
						{rule.ProductAName || $t('reports.frequency.unknown')}
					</span>
					<span class="text-base leading-none font-semibold text-primary">+</span>
					<span
						class="max-w-[45%] truncate rounded-md border border-border bg-muted/60 px-2.5 py-1 text-xs font-medium text-foreground"
					>
						{rule.ProductBName || $t('reports.frequency.unknown')}
					</span>
				</div>
				<div class="mt-2 flex items-center justify-between gap-2 px-0.5">
					<span class="num-mono truncate text-[10px] text-muted-foreground"
						>ID: {rule.ProductA}-{rule.ProductB}</span
					>
					<div class="flex shrink-0 items-center gap-1.5">
						<span class="h-1.5 w-1.5 animate-pulse rounded-full bg-primary"></span>
						<span class="num-mono text-xs font-semibold text-foreground">
							{rule.Frequency}
							{$t('reports.frequency.orders')}
						</span>
					</div>
				</div>
			</div>
		{/each}
	</div>
</div>
