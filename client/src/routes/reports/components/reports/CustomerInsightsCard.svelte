<script lang="ts">
	import * as Tooltip from '$lib/components/ui/tooltip';
	import { Users, HelpCircle, Mail } from 'lucide-svelte';
	import type { CustomerInsight } from '$lib/types';
	import { formatCurrency } from '$lib/utils';
	import { t } from '$lib/i18n';

	type Props = {
		customerInsights: CustomerInsight[];
	};

	let { customerInsights }: Props = $props();
</script>

<div
	class="relative col-span-2 overflow-hidden rounded-[2.5rem] border border-white/60 bg-white/40 p-8 shadow-2xl shadow-sky-900/5 backdrop-blur-3xl transition-all duration-500 hover:bg-white/50 lg:col-span-3"
>
	<div class="mb-8 flex items-center justify-between">
		<div class="flex items-center gap-4">
			<div
				class="flex h-12 w-12 items-center justify-center rounded-2xl border border-white/60 bg-gradient-to-br from-sky-50 to-cyan-50/50 text-sky-600 shadow-inner"
			>
				<Users class="h-6 w-6" />
			</div>
			<div>
				<h3 class="flex items-center gap-2 text-xl font-bold tracking-tight text-slate-800">
					{$t('reports.customers.title')}
					<Tooltip.Root>
						<Tooltip.Trigger>
							<HelpCircle class="h-4 w-4 text-slate-400 transition-colors hover:text-sky-600" />
						</Tooltip.Trigger>
						<Tooltip.Content>
							<p class="max-w-xs font-normal text-slate-700">
								{$t('reports.customers.description')}
							</p>
						</Tooltip.Content>
					</Tooltip.Root>
				</h3>
				<p class="text-sm font-medium text-slate-500">
					{$t('reports.customers.subtitle')}
				</p>
			</div>
		</div>
	</div>

	<!-- Headers -->
	<div
		class="mb-2 hidden grid-cols-12 gap-4 px-4 text-[11px] font-bold tracking-wider text-slate-400 lg:grid"
	>
		<div class="col-span-3">{$t('reports.customers.headers.customer')}</div>
		<div class="col-span-3">{$t('reports.customers.headers.contact')}</div>
		<div class="col-span-2 text-center">{$t('reports.customers.headers.orders')}</div>
		<div class="col-span-2">{$t('reports.customers.headers.last_order')}</div>
		<div class="col-span-2 text-right">
			{$t('reports.customers.headers.total_spent')}
		</div>
	</div>

	<div class="custom-scrollbar relative z-10 max-h-[400px] space-y-3 overflow-y-auto p-1 pr-2">
		{#each customerInsights.slice(0, 50) as cust}
			<div
				class="group relative overflow-hidden rounded-[1.5rem] border border-white/40 bg-white/40 p-4 transition-all duration-300 hover:scale-[1.01] hover:bg-white/80 hover:shadow-lg"
			>
				<div class="grid grid-cols-1 gap-4 lg:grid-cols-12 lg:items-center">
					<!-- Customer Info -->
					<div class="col-span-3 flex items-center gap-3">
						<div
							class="flex h-10 w-10 shrink-0 items-center justify-center rounded-full border border-white bg-gradient-to-br from-sky-100 to-sky-200 font-bold text-sky-700 shadow-inner"
						>
							{(cust.FullName || cust.Username || '?').charAt(0)}
						</div>
						<div class="min-w-0">
							<p
								class="truncate text-sm font-bold text-slate-700 transition-colors group-hover:text-sky-900"
							>
								{cust.FullName || $t('reports.customers.unknown')}
							</p>
							<p class="text-[10px] font-medium text-slate-400">ID: {cust.UserID}</p>
						</div>
					</div>

					<!-- Contact (Email/Username) -->
					<div class="col-span-3 min-w-0">
						<div class="flex items-center gap-2">
							<Mail class="h-3.5 w-3.5 text-slate-400" />
							<span class="truncate text-sm font-medium text-slate-600">{cust.Username || '-'}</span
							>
						</div>
					</div>

					<!-- Orders & Activity -->
					<div class="col-span-2 flex flex-col items-center justify-center gap-1">
						<span class="text-sm font-bold text-slate-700"
							>{cust.OrderCount}
							<span class="text-[10px] font-normal text-slate-400"
								>{$t('reports.customers.orders_suffix')}</span
							></span
						>
					</div>

					<!-- Last Order Date -->
					<div class="col-span-2">
						<div class="flex flex-col">
							<span class="text-xs font-semibold text-slate-600">
								{new Date(cust.LastOrderDate).toLocaleDateString()}
							</span>
							<div class="flex items-center gap-1">
								<span class="text-[10px] text-slate-400"
									>{cust.DaysSinceLastOrder} {$t('reports.customers.days_ago')}</span
								>
								{#if cust.DaysSinceLastOrder > 90}
									<span
										class="rounded-full bg-rose-100 px-1.5 py-0.5 text-[9px] font-bold text-rose-600"
									>
										{$t('reports.customers.lost')}
									</span>
								{/if}
							</div>
						</div>
					</div>

					<!-- Total Spent -->
					<div class="col-span-2 text-right">
						<span class="block text-sm font-black text-sky-600">
							{formatCurrency(cust.TotalSpent)}
						</span>
						<span class="text-[10px] font-medium text-sky-600/60"
							>{$t('reports.customers.lifetime_value')}</span
						>
					</div>
				</div>
			</div>
		{/each}
	</div>
</div>
