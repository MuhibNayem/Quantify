<script lang="ts">
	import { cn } from '$lib/utils';
	import type { Snippet } from 'svelte';

	let {
		children,
		className = '',
		...rest
	}: { children: Snippet; className?: string; [key: string]: any } = $props();
</script>

<div class={cn('panel relative overflow-hidden p-6', className)} {...rest}>
	{@render children()}
</div>
