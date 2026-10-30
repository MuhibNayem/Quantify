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

<div class="panel col-span-full p-5">
	<div class="flex items-start gap-3">
		<span
			class="flex h-9 w-9 shrink-0 items-center justify-center rounded-[calc(var(--radius)_-_4px)] border border-border bg-primary/10 text-primary"
		>
			<Users class="h-4 w-4" />
		</span>
		<div class="min-w-0">
			<h3 class="section-title flex items-center gap-1.5">
				{$t('reports.customers.title')}
				<Tooltip.Root>
					<Tooltip.Trigger>
						<HelpCircle
							class="h-4 w-4 text-muted-foreground transition-colors hover:text-primary"
						/>
					</Tooltip.Trigger>
					<Tooltip.Content>
						<p class="max-w-xs font-normal text-foreground">
							{$t('reports.customers.description')}
						</p>
					</Tooltip.Content>
				</Tooltip.Root>
			</h3>
			<p class="mt-0.5 text-xs text-muted-foreground">
				{$t('reports.customers.subtitle')}
			</p>
		</div>
	</div>

	<div class="custom-scrollbar mt-4 max-h-[420px] overflow-y-auto">
		<!-- Column headers -->
		<div
			class="hidden grid-cols-12 gap-3 border-b border-border/70 px-3 pb-2 text-[10px] font-semibold tracking-[0.08em] text-muted-foreground uppercase lg:grid"
		>
			<div class="col-span-3">{$t('reports.customers.headers.customer')}</div>
			<div class="col-span-3">{$t('reports.customers.headers.contact')}</div>
			<div class="col-span-2 text-center">{$t('reports.customers.headers.orders')}</div>
			<div class="col-span-2">{$t('reports.customers.headers.last_order')}</div>
			<div class="col-span-2 text-right">
				{$t('reports.customers.headers.total_spent')}
			</div>
		</div>

		{#if customerInsights.length === 0}
			<p class="py-8 text-center text-xs text-muted-foreground">
				{$t('common.no_data_available')}
			</p>
		{/if}

		<div class="divide-y divide-border/60">
			{#each customerInsights.slice(0, 50) as cust}
				<div class="px-3 py-3 transition-colors hover:bg-muted/40">
					<div class="grid grid-cols-1 gap-3 lg:grid-cols-12 lg:items-center">
						<!-- Customer Info -->
						<div class="col-span-3 flex items-center gap-2.5">
							<div
								class="flex h-8 w-8 shrink-0 items-center justify-center rounded-full border border-border bg-muted text-xs font-semibold text-muted-foreground"
							>
								{(cust.FullName || cust.Username || '?').charAt(0)}
							</div>
							<div class="min-w-0">
								<p class="truncate text-sm font-medium text-foreground">
									{cust.FullName || $t('reports.customers.unknown')}
								</p>
								<p class="num-mono text-[10px] text-muted-foreground">ID: {cust.UserID}</p>
							</div>
						</div>

						<!-- Contact (Email/Username) -->
						<div class="col-span-3 min-w-0">
							<div class="flex items-center gap-1.5">
								<Mail class="h-3.5 w-3.5 shrink-0 text-muted-foreground" />
								<span class="truncate text-sm text-muted-foreground">{cust.Username || '-'}</span>
							</div>
						</div>

						<!-- Orders & Activity -->
						<div class="col-span-2 flex flex-col items-center justify-center gap-1">
							<span class="num-mono text-sm font-semibold text-foreground"
								>{cust.OrderCount}
								<span class="text-[10px] font-normal text-muted-foreground"
									>{$t('reports.customers.orders_suffix')}</span
								></span
							>
						</div>

						<!-- Last Order Date -->
						<div class="col-span-2">
							<div class="flex flex-col">
								<span class="num-mono text-xs font-semibold text-foreground">
									{new Date(cust.LastOrderDate).toLocaleDateString()}
								</span>
								<div class="flex items-center gap-1.5">
									<span class="num-mono text-[10px] text-muted-foreground"
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
							<span class="num-mono block text-sm font-semibold text-foreground">
								{formatCurrency(cust.TotalSpent)}
							</span>
							<span class="text-[10px] text-muted-foreground"
								>{$t('reports.customers.lifetime_value')}</span
							>
						</div>
					</div>
				</div>
			{/each}
		</div>
	</div>
</div>
