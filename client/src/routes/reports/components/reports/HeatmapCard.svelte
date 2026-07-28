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

<div
	class="group relative col-span-2 overflow-hidden rounded-[2.5rem] border border-white/60 bg-white/40 p-8 shadow-2xl shadow-blue-900/5 backdrop-blur-3xl transition-all duration-500 hover:bg-white/50 lg:col-span-2"
>
	<div
		class="pointer-events-none absolute inset-0 bg-gradient-to-b from-white/40 to-transparent"
	></div>
	<div class="relative z-10">
		<div class="mb-8 flex items-center justify-between">
			<div class="flex items-center gap-4">
				<div
					class="flex h-12 w-12 items-center justify-center rounded-2xl border border-white/60 bg-gradient-to-br from-blue-50 to-indigo-50/50 text-blue-600 shadow-inner"
				>
					<Clock class="h-6 w-6" />
				</div>
				<div>
					<h3 class="flex items-center gap-2 text-xl font-bold tracking-tight text-slate-800">
						{$t('reports.heatmap.title')}
						<Tooltip.Root>
							<Tooltip.Trigger>
								<HelpCircle class="h-4 w-4 text-slate-400 transition-colors hover:text-blue-600" />
							</Tooltip.Trigger>
							<Tooltip.Content>
								<p class="max-w-xs font-normal text-slate-700">
									{$t('reports.heatmap.description')}
								</p>
							</Tooltip.Content>
						</Tooltip.Root>
					</h3>
					<p class="text-sm font-medium text-slate-500">
						{$t('reports.heatmap.subtitle')}
					</p>
				</div>
			</div>
		</div>

		{#if loading}
			<Skeleton class="h-[300px] w-full rounded-[2rem] bg-white/50" />
		{:else}
			<div class="relative">
				<!-- X-Axis Labels -->
				<div class="mb-3 flex justify-between pl-10">
					{#each hours as h}
						<div
							class="w-full border-l border-white/20 pl-2 text-left text-[10px] font-bold text-slate-400"
						>
							{h}:00
						</div>
					{/each}
				</div>

				<div class="space-y-1.5">
					{#each days as day, i}
						<div class="flex items-center gap-3">
							<span class="w-8 text-right text-[11px] font-bold tracking-wide text-slate-400"
								>{day.toUpperCase()}</span
							>
							<div
								class="grid flex-1 grid-cols-24 gap-[2px] rounded-xl border border-white/20 bg-white/20 p-1"
							>
								{#each grid.filter((c) => c.day === i) as cell}
									<div
										class="group/cell relative h-8 cursor-help rounded-[4px] transition-all duration-300 hover:z-20 hover:scale-[1.3] hover:shadow-lg"
										style="background-color: rgba(37, 99, 235, {Math.max(cell.intensity, 0.05)})"
									>
										<div
											class="pointer-events-none absolute bottom-full left-1/2 z-50 mb-2 hidden -translate-x-1/2 group-hover/cell:block"
										>
											<div
												class="rounded-xl bg-slate-900/90 px-3 py-1.5 text-[10px] font-bold whitespace-nowrap text-white shadow-xl backdrop-blur-md"
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
</div>
