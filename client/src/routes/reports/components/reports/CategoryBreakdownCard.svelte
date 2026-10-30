<script lang="ts">
	import * as Tooltip from '$lib/components/ui/tooltip';
	import { PieChart, HelpCircle } from 'lucide-svelte';
	import type { CategoryPerformance } from '$lib/types';
	import { formatCurrency, formatPercent } from '$lib/utils';
	import { t } from '$lib/i18n';

	type Props = {
		categoryPerformance: CategoryPerformance[];
	};

	let { categoryPerformance }: Props = $props();
</script>

<div class="panel p-5">
	<div class="flex items-start gap-3">
		<span
			class="flex h-9 w-9 shrink-0 items-center justify-center rounded-[calc(var(--radius)_-_4px)] border border-border bg-primary/10 text-primary"
		>
			<PieChart class="h-4 w-4" />
		</span>
		<div class="min-w-0">
			<h3 class="section-title flex items-center gap-1.5">
				{$t('reports.category.title')}
				<Tooltip.Root>
					<Tooltip.Trigger>
						<HelpCircle
							class="h-4 w-4 text-muted-foreground transition-colors hover:text-primary"
						/>
					</Tooltip.Trigger>
					<Tooltip.Content>
						<p class="max-w-xs font-normal text-foreground">{$t('reports.category.description')}</p>
					</Tooltip.Content>
				</Tooltip.Root>
			</h3>
		</div>
	</div>

	<div class="custom-scrollbar mt-4 max-h-[360px] space-y-2 overflow-y-auto pr-1">
		{#if categoryPerformance.length === 0}
			<p class="py-8 text-center text-xs text-muted-foreground">
				{$t('common.no_data_available')}
			</p>
		{/if}
		{#each categoryPerformance as cat}
			<div
				class="rounded-[calc(var(--radius)_-_6px)] border border-border/70 bg-muted/30 p-3 transition-colors"
			>
				<div class="flex items-center justify-between gap-2">
					<span class="truncate text-sm font-semibold text-foreground">{cat.CategoryName}</span>
					<span class="badge-neutral num-mono">
						{cat.ItemCount || cat.TotalUnits || 0}
						{$t('reports.category.items')}
					</span>
				</div>
				<div class="mt-2.5 grid grid-cols-2 gap-2">
					<div
						class="rounded-[calc(var(--radius)_-_8px)] border border-border/70 bg-card px-2.5 py-2"
					>
						<p class="overline">{$t('reports.category.sales')}</p>
						<p class="num-mono mt-1 text-sm font-semibold text-foreground">
							{formatCurrency(cat.TotalSales)}
						</p>
					</div>
					<div
						class="rounded-[calc(var(--radius)_-_8px)] border border-border/70 bg-card px-2.5 py-2"
					>
						<p class="overline">{$t('reports.category.margin')}</p>
						<p class="num-mono mt-1 flex items-baseline gap-1.5 text-sm font-semibold text-success">
							{formatCurrency(cat.GrossMargin)}
							<span class="text-[10px] font-medium">
								{formatPercent(cat.MarginPercent / 100)}
							</span>
						</p>
					</div>
				</div>
			</div>
		{/each}
	</div>
</div>
