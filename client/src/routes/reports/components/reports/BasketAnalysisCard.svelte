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

<div
	class="rounded-[2.5rem] border border-white/60 bg-white/40 p-8 shadow-2xl shadow-pink-900/5 backdrop-blur-3xl"
>
	<h3 class="mb-6 flex items-center gap-2 text-lg font-bold text-slate-700">
		<ShoppingCart class="h-5 w-5 text-pink-500" />
		{$t('reports.frequency.title')}
		<Tooltip.Root>
			<Tooltip.Trigger>
				<HelpCircle class="h-4 w-4 text-slate-400 transition-colors hover:text-pink-500" />
			</Tooltip.Trigger>
			<Tooltip.Content>
				<p class="max-w-xs">{$t('reports.frequency.description')}</p>
			</Tooltip.Content>
		</Tooltip.Root>
	</h3>
	<div class="custom-scrollbar max-h-[350px] space-y-2 overflow-y-auto p-1 pr-2">
		{#each basketAnalysis.slice(0, 10) as rule}
			<div
				class="group rounded-[1.25rem] border border-pink-100/50 bg-pink-50/40 p-3 backdrop-blur-sm transition-all hover:scale-[1.01] hover:bg-pink-50/80 hover:shadow-md"
			>
				<div
					class="mb-4 flex items-center justify-center gap-3 rounded-xl border border-white/40 bg-white/40 p-2 text-xs text-slate-500"
				>
					<span
						class="rounded-full border border-slate-100 bg-white px-3 py-1 font-bold text-slate-700 shadow-sm"
						>{rule.ProductAName || $t('reports.frequency.unknown')}</span
					>
					<span class="text-lg font-bold text-pink-400">+</span>
					<span
						class="rounded-full border border-slate-100 bg-white px-3 py-1 font-bold text-slate-700 shadow-sm"
						>{rule.ProductBName || $t('reports.frequency.unknown')}</span
					>
				</div>
				<div class="flex items-center justify-between px-1">
					<span class="font-mono text-[10px] font-medium text-pink-400/80"
						>ID: {rule.ProductA}-{rule.ProductB}</span
					>
					<div class="flex items-center gap-1.5">
						<span class="h-1.5 w-1.5 animate-pulse rounded-full bg-pink-500"></span>
						<span class="text-xs font-bold text-pink-700">
							{rule.Frequency}
							{$t('reports.frequency.orders')}
						</span>
					</div>
				</div>
			</div>
		{/each}
	</div>
</div>
