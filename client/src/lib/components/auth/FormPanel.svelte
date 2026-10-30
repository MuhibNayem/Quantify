<!-- client/src/lib/components/auth/FormPanel.svelte
     Porcelain form side: quiet ambient wash, generous spacing. -->
<script lang="ts">
	import type { Snippet } from 'svelte';

	interface Props {
		maxWidth?: string;
		children?: Snippet;
	}
	let { maxWidth = 'max-w-sm', children }: Props = $props();
</script>

<div
	class="app-canvas relative flex w-full flex-1 items-center justify-center overflow-hidden px-6 py-12 sm:px-10"
>
	<div class="hero-orbs" aria-hidden="true" style="opacity: 0.22;"></div>

	<div class="relative w-full {maxWidth}">
		{@render children?.()}
	</div>
</div>
