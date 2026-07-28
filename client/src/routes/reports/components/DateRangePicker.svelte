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
	class="flex items-center gap-3 rounded-[1.5rem] border border-white/40 bg-white/60 p-2 shadow-lg shadow-slate-200/20 backdrop-blur-2xl transition-all hover:scale-[1.01] hover:bg-white/70 hover:shadow-xl"
>
	<div
		class="flex items-center gap-3 rounded-[1.2rem] border border-white/60 bg-gradient-to-b from-white/80 to-white/40 px-4 py-2 shadow-sm"
	>
		<CalendarRange class="h-4 w-4 text-blue-600" />
		<input
			type="date"
			bind:value={start}
			class="cursor-pointer border-none bg-transparent p-0 font-mono text-xs font-bold tracking-wide text-slate-700 focus:ring-0"
		/>
		<span class="font-light text-slate-300">|</span>
		<input
			type="date"
			bind:value={end}
			class="cursor-pointer border-none bg-transparent p-0 font-mono text-xs font-bold tracking-wide text-slate-700 focus:ring-0"
		/>
	</div>
	<button
		onclick={onRefresh}
		class="group rounded-[1.2rem] bg-gradient-to-br from-blue-600 to-blue-700 p-2.5 text-white shadow-lg shadow-blue-500/25 transition-all hover:from-blue-500 hover:to-blue-600 active:scale-95"
		title={$t('reports.actions.refresh')}
	>
		<Sparkles class="h-4 w-4 transition-transform group-active:skew-12" />
	</button>
</div>
