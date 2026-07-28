<script lang="ts">
	import * as Tooltip from '$lib/components/ui/tooltip';
	import { Users, HelpCircle } from 'lucide-svelte';
	import { formatCurrency } from '$lib/utils';
	import { t } from '$lib/i18n';

	type Props = {
		employeeSales: { name: string; sales: number; count: number }[];
	};

	let { employeeSales }: Props = $props();
</script>

<div
	class="relative col-span-2 flex flex-col overflow-hidden rounded-[2.5rem] border border-white/60 bg-white/40 p-8 shadow-2xl shadow-purple-900/5 backdrop-blur-3xl transition-all duration-500 hover:bg-white/50 lg:col-span-1"
>
	<div class="mb-8 flex shrink-0 items-center gap-4">
		<div
			class="flex h-12 w-12 items-center justify-center rounded-2xl border border-white/60 bg-gradient-to-br from-purple-50 to-pink-50/50 text-purple-600 shadow-inner"
		>
			<Users class="h-6 w-6" />
		</div>
		<div>
			<h3 class="flex items-center gap-2 text-xl font-bold tracking-tight text-slate-800">
				{$t('reports.staff.title')}
				<Tooltip.Root>
					<Tooltip.Trigger>
						<HelpCircle class="h-4 w-4 text-slate-400 transition-colors hover:text-purple-600" />
					</Tooltip.Trigger>
					<Tooltip.Content>
						<p class="max-w-xs font-normal text-slate-700">
							{$t('reports.staff.description')}
						</p>
					</Tooltip.Content>
				</Tooltip.Root>
			</h3>
			<p class="text-sm font-medium text-slate-500">{$t('reports.staff.subtitle')}</p>
		</div>
	</div>

	<div class="custom-scrollbar relative z-10 flex-1 space-y-2 overflow-y-auto p-1 pr-2">
		{#each employeeSales.slice(0, 50) as emp, i}
			<div
				class="group flex w-full cursor-default items-center justify-between gap-3 overflow-x-auto rounded-[1.25rem] border border-white/40 bg-white/40 p-3 transition-all duration-300 hover:scale-[1.01] hover:bg-white/80 hover:shadow-lg"
			>
				<div class="flex shrink-0 items-center gap-4">
					<div
						class="flex h-10 w-10 items-center justify-center rounded-full border border-white bg-gradient-to-br from-slate-100 to-slate-200 font-bold text-slate-600 shadow-inner"
					>
						{emp.name.charAt(0)}
					</div>
					<div>
						<p
							class="text-sm font-bold whitespace-nowrap text-slate-700 transition-colors group-hover:text-blue-900"
						>
							{emp.name}
						</p>
						<p
							class="text-[11px] font-medium whitespace-nowrap text-slate-400 group-hover:text-slate-500"
						>
							{emp.count}
							{$t('reports.staff.transactions')}
						</p>
					</div>
				</div>
				<span
					class="shrink-0 rounded-full border border-purple-100/50 bg-purple-50/50 px-3 py-1 text-sm font-bold whitespace-nowrap text-purple-600"
					>{formatCurrency(emp.sales)}</span
				>
			</div>
		{/each}
	</div>
</div>
