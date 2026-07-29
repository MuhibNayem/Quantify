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

<div class="liquid-panel col-span-2 flex flex-col rounded-lg p-6 lg:col-span-1">
	<div class="mb-6 flex shrink-0 items-center gap-3">
		<div
			class="flex h-10 w-10 items-center justify-center rounded-md border border-border bg-primary/10 text-primary"
		>
			<Users class="h-5 w-5" />
		</div>
		<div>
			<h3 class="font-display text-foreground flex items-center gap-2 text-lg font-semibold">
				{$t('reports.staff.title')}
				<Tooltip.Root>
					<Tooltip.Trigger>
						<HelpCircle class="text-muted-foreground hover:text-primary h-4 w-4 transition-colors" />
					</Tooltip.Trigger>
					<Tooltip.Content>
						<p class="text-foreground max-w-xs font-normal">
							{$t('reports.staff.description')}
						</p>
					</Tooltip.Content>
				</Tooltip.Root>
			</h3>
			<p class="text-muted-foreground text-sm">{$t('reports.staff.subtitle')}</p>
		</div>
	</div>

	<div class="custom-scrollbar relative z-10 flex-1 space-y-2 overflow-y-auto p-1 pr-2">
		{#each employeeSales.slice(0, 50) as emp, i}
			<div
				class="group border-border bg-card hover:bg-muted flex w-full cursor-default items-center justify-between gap-3 overflow-x-auto rounded-md border p-3 transition-colors"
			>
				<div class="flex shrink-0 items-center gap-3">
					<div
						class="bg-muted text-muted-foreground border-border flex h-9 w-9 items-center justify-center rounded-full border font-semibold"
					>
						{emp.name.charAt(0)}
					</div>
					<div>
						<p class="text-foreground text-sm font-semibold whitespace-nowrap">
							{emp.name}
						</p>
						<p class="text-muted-foreground text-[11px] whitespace-nowrap num-mono">
							{emp.count}
							{$t('reports.staff.transactions')}
						</p>
					</div>
				</div>
				<span
					class="bg-muted text-foreground shrink-0 rounded-md px-3 py-1 text-sm font-semibold whitespace-nowrap num-mono"
					>{formatCurrency(emp.sales)}</span
				>
			</div>
		{/each}
	</div>
</div>
