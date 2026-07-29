<script lang="ts">
	import type { Component } from 'svelte';
	import { Button } from '$lib/components/ui/button';

	let {
		icon,
		title,
		description = '',
		actionLabel = '',
		onAction
	}: {
		icon?: Component;
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
		<Button onclick={onAction} variant="outline" size="sm">{actionLabel}</Button>
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
		gap: 0.75rem;
	}
	.icon-wrap {
		opacity: 0.35;
		margin-bottom: 0.5rem;
	}
	h3 {
		font-size: 1rem;
		font-weight: 600;
		margin: 0;
	}
	p {
		font-size: 0.85rem;
		opacity: 0.6;
		margin: 0;
		max-width: 320px;
	}
</style>
