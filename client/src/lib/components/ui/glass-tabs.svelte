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
		'flex w-fit flex-wrap items-center justify-center gap-1 rounded-lg border border-border bg-muted p-1 shadow-xs',
		className
	)}
>
	{#each tabs as tab}
		<div class="relative flex items-center">
			<Button
				size="sm"
				variant="ghost"
				onclick={() => (value = tab.value)}
				class={cn(
					'relative isolate h-9 rounded-md border border-transparent px-4 py-2 text-sm font-medium transition-colors duration-200',
					value === tab.value
						? 'text-accent-foreground'
						: 'text-muted-foreground hover:text-foreground'
				)}
			>
				{#if value === tab.value}
					<div
						class="glass-toggle-active absolute inset-0 -z-10 rounded-md"
						in:fade={{ duration: 200 }}
						out:fade={{ duration: 200 }}
					></div>
				{/if}
				<span class="relative z-10">{tab.label}</span>
			</Button>
		</div>
	{/each}
</div>
