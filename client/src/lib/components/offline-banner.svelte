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
		<WifiOff size={16} />
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
		padding: 0.5rem 1rem;
		background: #f59e0b;
		color: #1a1a1a;
		font-size: 0.85rem;
		font-weight: 500;
	}
</style>
