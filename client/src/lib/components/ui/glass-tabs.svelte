<script lang="ts">
	import { Button } from '$lib/components/ui/button';
	import { fade } from 'svelte/transition';
	import { cn } from '$lib/utils';

	type Tab = {
		value: string;
		label: string;
	};

	let {
		tabs,
		value = $bindable(),
		class: className
	} = $props<{
		tabs: Tab[];
		value: string;
		class?: string;
	}>();
</script>

<div
	class={cn(
		'flex w-fit flex-wrap items-center justify-center gap-2 rounded-[28px] border border-white/40 bg-white/70 px-2 py-2 text-slate-600 shadow-[0_25px_80px_-40px_rgba(15,23,42,0.35)] backdrop-blur-xl',
		className
	)}
>
	{#each tabs as tab}
		<div class="relative flex items-center">
			<Button
				size="lg"
				variant="ghost"
				onclick={() => (value = tab.value)}
				class={cn(
					'relative isolate rounded-2xl border border-transparent px-6 py-2 text-sm font-semibold transition-colors duration-200 hover:text-slate-900',
					value === tab.value ? 'text-slate-900' : 'text-slate-500'
				)}
			>
				{#if value === tab.value}
					<div
						class="glass-toggle-active absolute inset-0 -z-10 rounded-2xl"
						in:fade={{ duration: 200 }}
						out:fade={{ duration: 200 }}
					></div>
				{/if}
				<span class="relative z-10">{tab.label}</span>
			</Button>
		</div>
	{/each}
</div>
