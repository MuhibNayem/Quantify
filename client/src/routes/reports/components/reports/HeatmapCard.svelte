<script lang="ts">
	import { Skeleton } from '$lib/components/ui/skeleton';
	import * as Tooltip from '$lib/components/ui/tooltip';
	import { Clock, HelpCircle } from 'lucide-svelte';
	import { formatCurrency } from '$lib/utils';
	import { t } from '$lib/i18n';

	type Props = {
		grid: { day: number; hour: number; sales: number; intensity: number }[];
		loading: boolean;
	};

	let { grid, loading }: Props = $props();

	const days = ['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat'];
	const hours = [0, 4, 8, 12, 16, 20, 23];
</script>

<div class="panel col-span-2 p-5">
	<div class="flex items-start gap-3">
		<span
			class="flex h-9 w-9 shrink-0 items-center justify-center rounded-[calc(var(--radius)_-_4px)] border border-border bg-primary/10 text-primary"
		>
			<Clock class="h-4 w-4" />
		</span>
		<div class="min-w-0">
			<h3 class="section-title flex items-center gap-1.5">
				{$t('reports.heatmap.title')}
				<Tooltip.Root>
					<Tooltip.Trigger>
						<HelpCircle
							class="h-4 w-4 text-muted-foreground transition-colors hover:text-primary"
						/>
					</Tooltip.Trigger>
					<Tooltip.Content>
						<p class="max-w-xs font-normal text-foreground">
							{$t('reports.heatmap.description')}
						</p>
					</Tooltip.Content>
				</Tooltip.Root>
			</h3>
			<p class="mt-0.5 text-xs text-muted-foreground">
				{$t('reports.heatmap.subtitle')}
			</p>
		</div>
	</div>

	{#if loading}
		<Skeleton class="mt-4 h-[268px] w-full rounded-md bg-muted" />
	{:else}
		<div class="mt-4">
			<!-- X-Axis Labels -->
			<div class="mb-2 flex justify-between pl-10">
				{#each hours as h}
					<div
						class="num-mono w-full border-l border-border pl-1.5 text-left text-[10px] font-medium text-muted-foreground"
					>
						{h}:00
					</div>
				{/each}
			</div>

			<div class="space-y-1">
				{#each days as day, i}
					<div class="flex items-center gap-3">
						<span
							class="w-8 text-right text-[10px] font-semibold tracking-[0.08em] text-muted-foreground"
							>{day.toUpperCase()}</span
						>
						<div
							class="grid flex-1 grid-cols-24 gap-[2px] rounded-[calc(var(--radius)_-_6px)] border border-border bg-muted/40 p-1"
						>
							{#each grid.filter((c) => c.day === i) as cell}
								<div
									class="group/cell relative h-8 cursor-help rounded-[3px] transition-transform duration-200 hover:z-20 hover:scale-110"
									style="background-color: hsl(var(--primary) / {Math.max(cell.intensity, 0.05)})"
								>
									<div
										class="pointer-events-none absolute bottom-full left-1/2 z-50 mb-2 hidden -translate-x-1/2 group-hover/cell:block"
									>
										<div
											class="glass-float num-mono px-2.5 py-1 text-[11px] font-semibold whitespace-nowrap text-foreground"
										>
											{formatCurrency(cell.sales)}
										</div>
									</div>
								</div>
							{/each}
						</div>
					</div>
				{/each}
			</div>
		</div>
	{/if}
</div>
