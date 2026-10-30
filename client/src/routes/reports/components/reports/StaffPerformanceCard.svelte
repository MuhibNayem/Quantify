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

<div class="panel col-span-2 flex flex-col p-5 lg:col-span-1">
	<div class="flex items-start gap-3">
		<span
			class="flex h-9 w-9 shrink-0 items-center justify-center rounded-[calc(var(--radius)_-_4px)] border border-border bg-primary/10 text-primary"
		>
			<Users class="h-4 w-4" />
		</span>
		<div class="min-w-0">
			<h3 class="section-title flex items-center gap-1.5">
				{$t('reports.staff.title')}
				<Tooltip.Root>
					<Tooltip.Trigger>
						<HelpCircle
							class="h-4 w-4 text-muted-foreground transition-colors hover:text-primary"
						/>
					</Tooltip.Trigger>
					<Tooltip.Content>
						<p class="max-w-xs font-normal text-foreground">
							{$t('reports.staff.description')}
						</p>
					</Tooltip.Content>
				</Tooltip.Root>
			</h3>
			<p class="mt-0.5 text-xs text-muted-foreground">{$t('reports.staff.subtitle')}</p>
		</div>
	</div>

	<div class="custom-scrollbar mt-4 max-h-[420px] flex-1 space-y-1 overflow-y-auto pr-1">
		{#if employeeSales.length === 0}
			<p class="py-8 text-center text-xs text-muted-foreground">
				{$t('common.no_data_available')}
			</p>
		{/if}
		{#each employeeSales.slice(0, 50) as emp}
			<div
				class="flex w-full items-center justify-between gap-3 rounded-[calc(var(--radius)_-_6px)] px-2.5 py-2 transition-colors hover:bg-muted/50"
			>
				<div class="flex min-w-0 items-center gap-2.5">
					<div
						class="flex h-8 w-8 shrink-0 items-center justify-center rounded-full border border-border bg-muted text-xs font-semibold text-muted-foreground"
					>
						{emp.name.charAt(0)}
					</div>
					<div class="min-w-0">
						<p class="truncate text-sm font-medium text-foreground">
							{emp.name}
						</p>
						<p class="num-mono text-[11px] text-muted-foreground">
							{emp.count}
							{$t('reports.staff.transactions')}
						</p>
					</div>
				</div>
				<span class="num-mono shrink-0 text-sm font-semibold whitespace-nowrap text-foreground"
					>{formatCurrency(emp.sales)}</span
				>
			</div>
		{/each}
	</div>
</div>
