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

<div class="liquid-panel col-span-2 rounded-lg p-6 lg:col-span-3">
	<div class="mb-6 flex items-center justify-between">
		<div class="flex items-center gap-3">
			<div
				class="flex h-10 w-10 items-center justify-center rounded-md border border-border bg-primary/10 text-primary"
			>
				<Users class="h-5 w-5" />
			</div>
			<div>
				<h3 class="font-display text-foreground flex items-center gap-2 text-lg font-semibold">
					{$t('reports.customers.title')}
					<Tooltip.Root>
						<Tooltip.Trigger>
							<HelpCircle class="text-muted-foreground hover:text-primary h-4 w-4 transition-colors" />
						</Tooltip.Trigger>
						<Tooltip.Content>
							<p class="text-foreground max-w-xs font-normal">
								{$t('reports.customers.description')}
							</p>
						</Tooltip.Content>
					</Tooltip.Root>
				</h3>
				<p class="text-muted-foreground text-sm">
					{$t('reports.customers.subtitle')}
				</p>
			</div>
		</div>
	</div>

	<!-- Headers -->
	<div
		class="text-muted-foreground mb-2 hidden grid-cols-12 gap-4 px-4 text-[11px] font-semibold tracking-wider uppercase lg:grid"
	>
		<div class="col-span-3">{$t('reports.customers.headers.customer')}</div>
		<div class="col-span-3">{$t('reports.customers.headers.contact')}</div>
		<div class="col-span-2 text-center">{$t('reports.customers.headers.orders')}</div>
		<div class="col-span-2">{$t('reports.customers.headers.last_order')}</div>
		<div class="col-span-2 text-right">
			{$t('reports.customers.headers.total_spent')}
		</div>
	</div>

	<div class="custom-scrollbar relative z-10 max-h-[400px] space-y-2 overflow-y-auto p-1 pr-2">
		{#each customerInsights.slice(0, 50) as cust}
			<div
				class="group border-border bg-card hover:bg-muted rounded-md border p-4 transition-colors"
			>
				<div class="grid grid-cols-1 gap-4 lg:grid-cols-12 lg:items-center">
					<!-- Customer Info -->
					<div class="col-span-3 flex items-center gap-3">
						<div
							class="bg-muted text-muted-foreground border-border flex h-9 w-9 shrink-0 items-center justify-center rounded-full border font-semibold"
						>
							{(cust.FullName || cust.Username || '?').charAt(0)}
						</div>
						<div class="min-w-0">
							<p class="text-foreground truncate text-sm font-semibold">
								{cust.FullName || $t('reports.customers.unknown')}
							</p>
							<p class="text-muted-foreground text-[10px] num-mono">ID: {cust.UserID}</p>
						</div>
					</div>

					<!-- Contact (Email/Username) -->
					<div class="col-span-3 min-w-0">
						<div class="flex items-center gap-2">
							<Mail class="text-muted-foreground h-3.5 w-3.5" />
							<span class="text-muted-foreground truncate text-sm">{cust.Username || '-'}</span>
						</div>
					</div>

					<!-- Orders & Activity -->
					<div class="col-span-2 flex flex-col items-center justify-center gap-1">
						<span class="text-foreground text-sm font-semibold num-mono"
							>{cust.OrderCount}
							<span class="text-muted-foreground text-[10px] font-normal"
								>{$t('reports.customers.orders_suffix')}</span
							></span
						>
					</div>

					<!-- Last Order Date -->
					<div class="col-span-2">
						<div class="flex flex-col">
							<span class="text-foreground text-xs font-semibold num-mono">
								{new Date(cust.LastOrderDate).toLocaleDateString()}
							</span>
							<div class="flex items-center gap-1">
								<span class="text-muted-foreground text-[10px] num-mono"
									>{cust.DaysSinceLastOrder} {$t('reports.customers.days_ago')}</span
								>
								{#if cust.DaysSinceLastOrder > 90}
									<span class="badge-danger text-[9px]">
										{$t('reports.customers.lost')}
									</span>
								{/if}
							</div>
						</div>
					</div>

					<!-- Total Spent -->
					<div class="col-span-2 text-right">
						<span class="text-primary block text-sm font-semibold num-mono">
							{formatCurrency(cust.TotalSpent)}
						</span>
						<span class="text-muted-foreground text-[10px]"
							>{$t('reports.customers.lifetime_value')}</span
						>
					</div>
				</div>
			</div>
		{/each}
	</div>
</div>
