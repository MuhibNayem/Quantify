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

<div class="liquid-panel group relative col-span-2 overflow-hidden rounded-lg p-6 lg:col-span-2">
	<div class="mb-6 flex items-center justify-between">
		<div class="flex items-center gap-3">
			<div
				class="flex h-10 w-10 items-center justify-center rounded-md border border-border bg-primary/10 text-primary"
			>
				<Clock class="h-5 w-5" />
			</div>
			<div>
				<h3 class="font-display text-foreground flex items-center gap-2 text-lg font-semibold">
					{$t('reports.heatmap.title')}
					<Tooltip.Root>
						<Tooltip.Trigger>
							<HelpCircle class="text-muted-foreground hover:text-primary h-4 w-4 transition-colors" />
						</Tooltip.Trigger>
						<Tooltip.Content>
							<p class="text-foreground max-w-xs font-normal">
								{$t('reports.heatmap.description')}
							</p>
						</Tooltip.Content>
					</Tooltip.Root>
				</h3>
				<p class="text-muted-foreground text-sm">
					{$t('reports.heatmap.subtitle')}
				</p>
			</div>
		</div>
	</div>

	{#if loading}
		<Skeleton class="bg-muted h-[300px] w-full rounded-md" />
	{:else}
		<div class="relative">
			<!-- X-Axis Labels -->
			<div class="mb-3 flex justify-between pl-10">
				{#each hours as h}
					<div
						class="text-muted-foreground border-border w-full border-l pl-2 text-left text-[10px] font-semibold num-mono"
					>
						{h}:00
					</div>
				{/each}
			</div>

			<div class="space-y-1.5">
				{#each days as day, i}
					<div class="flex items-center gap-3">
						<span class="text-muted-foreground w-8 text-right text-[11px] font-semibold tracking-wide"
							>{day.toUpperCase()}</span
						>
						<div
							class="grid-cols-24 border-border bg-muted/40 grid flex-1 gap-[2px] rounded-md border p-1"
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
											class="bg-foreground text-background rounded-md px-3 py-1.5 text-[10px] font-semibold whitespace-nowrap shadow-md num-mono"
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
