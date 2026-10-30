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
	role="tablist"
	class={cn(
		'flex w-fit flex-wrap items-center justify-center gap-1 rounded-lg border border-border bg-muted/60 p-1 shadow-xs',
		className
	)}
>
	{#each tabs as tab}
		<div class="relative flex items-center">
			<Button
				size="sm"
				variant="ghost"
				role="tab"
				aria-selected={value === tab.value}
				onclick={() => (value = tab.value)}
				class={cn(
					'relative isolate h-9 rounded-md border border-transparent px-4 py-2 text-sm font-medium transition-colors duration-200',
					value === tab.value ? 'text-foreground' : 'text-muted-foreground hover:text-foreground'
				)}
			>
				{#if value === tab.value}
					<div
						class="segmented-pill absolute inset-0 -z-10 rounded-md"
						in:fade={{ duration: 200 }}
						out:fade={{ duration: 200 }}
					></div>
				{/if}
				<span class="relative z-10">{tab.label}</span>
			</Button>
		</div>
	{/each}
</div>

<style>
	/* Selected segment: opaque porcelain key sitting on the muted track. */
	.segmented-pill {
		background: hsl(var(--card));
		border: 1px solid hsl(var(--border) / 0.9);
		box-shadow: var(--surface-inset), var(--shadow-xs);
	}
</style>
