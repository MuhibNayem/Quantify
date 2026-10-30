<script lang="ts">
	import { network } from '$lib/stores/network';
	import { WifiOff } from 'lucide-svelte';
	import { pendingCount } from '$lib/offline-queue';

	let online = $state(true);
	let pending = $state(0);

	$effect(() => {
		const unsubOnline = network.isOnline.subscribe((v) => (online = v));
		const unsubPending = pendingCount.subscribe((v) => (pending = v));
		return () => {
			unsubOnline();
			unsubPending();
		};
	});
</script>

{#if !online}
	<div class="offline-banner" role="status" aria-live="polite">
		<WifiOff size={16} class="offline-icon" />
		<span>
			You're offline.
			{#if pending > 0}
				{pending} change{pending === 1 ? '' : 's'} queued — will sync when reconnected.
			{:else}
				Changes will sync when you reconnect.
			{/if}
		</span>
	</div>
{/if}

<style>
	.offline-banner {
		position: fixed;
		top: 0;
		left: 0;
		right: 0;
		z-index: 100;
		display: flex;
		align-items: center;
		justify-content: center;
		gap: 0.5rem;
		padding: 0.55rem 1rem;
		background: hsl(var(--warning) / 0.14);
		color: hsl(var(--foreground));
		border-bottom: 1px solid hsl(var(--warning) / 0.32);
		backdrop-filter: blur(22px) saturate(180%);
		-webkit-backdrop-filter: blur(22px) saturate(180%);
		box-shadow: var(--shadow-xs);
		font-size: 0.8125rem;
		font-weight: 500;
		letter-spacing: -0.005em;
	}

	.offline-icon {
		flex: none;
		color: hsl(var(--warning));
	}

	@supports not ((backdrop-filter: blur(1px)) or (-webkit-backdrop-filter: blur(1px))) {
		.offline-banner {
			background: hsl(var(--card));
			backdrop-filter: none;
			-webkit-backdrop-filter: none;
		}
	}
</style>
