<script lang="ts">
	import { CalendarRange, Sparkles } from 'lucide-svelte';
	import { t } from '$lib/i18n';

	type Props = {
		start: string;
		end: string;
		onRefresh: () => void;
	};

	let { start = $bindable(), end = $bindable(), onRefresh }: Props = $props();
</script>

<div
	class="liquid-panel flex items-center gap-3 rounded-lg p-2"
>
	<div
		class="border-border bg-muted flex items-center gap-3 rounded-md border px-4 py-2"
	>
		<CalendarRange class="text-primary h-4 w-4" />
		<input
			type="date"
			bind:value={start}
			class="text-foreground focus:ring-0 num-mono cursor-pointer border-none bg-transparent p-0 text-xs font-semibold tracking-wide focus:outline-none"
		/>
		<span class="text-muted-foreground font-light">|</span>
		<input
			type="date"
			bind:value={end}
			class="text-foreground focus:ring-0 num-mono cursor-pointer border-none bg-transparent p-0 text-xs font-semibold tracking-wide focus:outline-none"
		/>
	</div>
	<button
		onclick={onRefresh}
		class="group bg-primary text-primary-foreground hover:bg-primary/90 rounded-md p-2.5 shadow-sm transition-colors active:scale-95"
		title={$t('reports.actions.refresh')}
	>
		<Sparkles class="h-4 w-4 transition-transform group-active:skew-12" />
	</button>
</div>
