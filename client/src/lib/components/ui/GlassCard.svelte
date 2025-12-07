<script lang="ts">
	import { cn } from '$lib/utils';
	import { glassCard } from '$lib/styles/liquid-glass';
	import type { Snippet } from 'svelte';

	let {
		children,
		className = '',
		...rest
	}: { children: Snippet; className?: string; [key: string]: any } = $props();
</script>

<div class={cn(glassCard.base, className)} {...rest}>
	{@render children()}
</div>
