<script lang="ts">
	import { CalendarRange, Sparkles } from 'lucide-svelte';
	import { t } from '$lib/i18n';

	type Props = {
		start: string;
		end: string;
		onRefresh: () => void;
	};

	let { start = $bindable(), end = $bindable(), onRefresh }: Props = $props();

	const dateInputClass =
		'num-mono w-[8.5rem] cursor-pointer border-0 bg-transparent p-0 text-[13px] font-medium tracking-tight text-foreground outline-none focus:ring-0 focus:outline-none';
</script>

<div class="panel flex items-center gap-2 p-1.5">
	<div
		class="flex h-9 items-center gap-2 rounded-[calc(var(--radius)_-_6px)] border border-border bg-muted/50 px-3"
	>
		<CalendarRange class="h-4 w-4 shrink-0 text-muted-foreground" />
		<input type="date" bind:value={start} class={dateInputClass} aria-label="Start date" />
		<span class="text-xs text-muted-foreground/60">—</span>
		<input type="date" bind:value={end} class={dateInputClass} aria-label="End date" />
	</div>
	<button
		type="button"
		onclick={onRefresh}
		title={$t('reports.actions.refresh')}
		aria-label={$t('reports.actions.refresh')}
		class="btn-soothe inline-flex h-9 w-9 items-center justify-center p-0 whitespace-nowrap"
	>
		<Sparkles class="h-4 w-4" />
	</button>
</div>
