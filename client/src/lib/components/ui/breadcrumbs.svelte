<script lang="ts">
	import { ChevronRight, Home } from 'lucide-svelte';

	interface Crumb {
		label: string;
		href?: string;
	}

	let { items = [] }: { items: Crumb[] } = $props();
</script>

<nav aria-label="Breadcrumb" class="breadcrumbs">
	<ol>
		<li>
			<a href="/"><Home size={14} /><span class="sr-only">Home</span></a>
		</li>
		{#each items as item, i}
			<li aria-current={i === items.length - 1 ? 'page' : undefined}>
				<ChevronRight size={14} class="separator" />
				{#if item.href && i < items.length - 1}
					<a href={item.href}>{item.label}</a>
				{:else}
					<span>{item.label}</span>
				{/if}
			</li>
		{/each}
	</ol>
</nav>

<style>
	.breadcrumbs ol {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		list-style: none;
		padding: 0;
		margin: 0;
		font-size: 0.8rem;
	}
	.breadcrumbs li {
		display: flex;
		align-items: center;
		gap: 0.5rem;
	}
	.breadcrumbs a {
		color: var(--text-secondary, #64748b);
		text-decoration: none;
		display: flex;
		align-items: center;
		transition: color 0.15s;
	}
	.breadcrumbs a:hover {
		color: var(--text-primary, #0f172a);
	}
	.breadcrumbs [aria-current='page'] span {
		color: var(--text-primary, #0f172a);
		font-weight: 600;
	}
	:global(.separator) {
		opacity: 0.4;
	}
</style>
