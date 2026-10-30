<script lang="ts">
	import type { ComponentType } from 'svelte';
	import { Button } from '$lib/components/ui/button';

	let {
		icon,
		title,
		description = '',
		actionLabel = '',
		onAction
	}: {
		icon?: ComponentType;
		title: string;
		description?: string;
		actionLabel?: string;
		onAction?: () => void;
	} = $props();
</script>

<div class="empty-state">
	{#if icon}
		{@const Icon = icon}
		<div class="icon-wrap"><Icon size={40} stroke-width={1.5} /></div>
	{/if}
	<h3>{title}</h3>
	{#if description}<p>{description}</p>{/if}
	{#if actionLabel && onAction}
		<Button onclick={onAction} variant="secondary" size="sm" class="mt-2 font-medium">
			{actionLabel}
		</Button>
	{/if}
</div>

<style>
	.empty-state {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		text-align: center;
		padding: 3rem 1.5rem;
		gap: 0.5rem;
	}

	.icon-wrap {
		color: hsl(var(--muted-foreground));
		opacity: 0.55;
		margin-bottom: 0.5rem;
		display: inline-flex;
		align-items: center;
		justify-content: center;
		width: 4rem;
		height: 4rem;
		border-radius: calc(var(--radius) - 2px);
		background: hsl(var(--muted) / 0.6);
		box-shadow: inset 0 0 0 1px hsl(var(--border) / 0.7);
	}

	h3 {
		font-family: var(--font-display);
		font-size: 1rem;
		font-weight: 600;
		letter-spacing: -0.015em;
		margin: 0;
		color: hsl(var(--foreground));
	}

	p {
		font-size: 0.85rem;
		color: hsl(var(--muted-foreground));
		margin: 0;
		max-width: 34ch;
		line-height: 1.55;
	}
</style>
