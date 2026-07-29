<script lang="ts">
	import { onMount } from 'svelte';
	import { History } from 'lucide-svelte';
	import { auditApi } from '$lib/api/resources';
	import type { AuditLog } from '$lib/types';
	import TimeAgo from '$lib/components/ui/time-ago.svelte';

	let {
		entity,
		entityId
	}: {
		entity: string;
		entityId: string | number;
	} = $props();

	let logs = $state<AuditLog[]>([]);
	let loading = $state(true);

	onMount(async () => {
		try {
			const res = await auditApi.list({ entity, limit: 20 });
			logs = (res.data || res.items || []).filter(
				(l: AuditLog) => String(l.EntityID) === String(entityId)
			);
		} catch {
			logs = [];
		} finally {
			loading = false;
		}
	});

	function actionColor(action: string): string {
		if (action.includes('CREATE')) return 'text-emerald-600';
		if (action.includes('DELETE')) return 'text-red-600';
		if (action.includes('UPDATE')) return 'text-blue-600';
		return 'text-amber-600';
	}

	function actionDot(action: string): string {
		if (action.includes('CREATE')) return 'bg-emerald-500 ring-emerald-100';
		if (action.includes('DELETE')) return 'bg-red-500 ring-red-100';
		if (action.includes('UPDATE')) return 'bg-blue-500 ring-blue-100';
		return 'bg-amber-500 ring-amber-100';
	}

	function changeSummary(changes: string): string | null {
		if (!changes) return null;
		try {
			const parsed = JSON.parse(changes);
			if (parsed && typeof parsed === 'object' && !Array.isArray(parsed)) {
				const keys = Object.keys(parsed);
				if (keys.length === 0) return null;
				const shown = keys.slice(0, 3).join(', ');
				return keys.length > 3 ? `${shown} +${keys.length - 3}` : shown;
			}
		} catch {
			// not JSON — fall through to raw preview
		}
		const trimmed = changes.trim();
		if (!trimmed) return null;
		return trimmed.length > 40 ? `${trimmed.slice(0, 40)}…` : trimmed;
	}
</script>

<div class="activity-history">
	<h4 class="flex items-center gap-2 text-sm font-semibold text-slate-800">
		<span
			class="flex h-7 w-7 items-center justify-center rounded-lg bg-slate-800/5 text-slate-500 ring-1 ring-slate-900/5"
		>
			<History size={14} />
		</span>
		Activity
		{#if !loading && logs.length > 0}
			<span
				class="ml-auto rounded-full bg-slate-800/5 px-2 py-0.5 text-[10px] font-bold tracking-wide text-slate-500 tabular-nums"
			>
				{logs.length}
			</span>
		{/if}
	</h4>

	{#if loading}
		<ul class="timeline" aria-hidden="true">
			{#each [72, 58, 64] as w, i}
				<li class="skeleton-row" style={`--i:${i}`}>
					<span class="skeleton-dot"></span>
					<span class="skeleton-bar" style={`width:${w}%`}></span>
				</li>
			{/each}
		</ul>
	{:else if logs.length === 0}
		<p class="empty">No recorded activity.</p>
	{:else}
		<ul class="timeline">
			{#each logs as log, i}
				<li class="entry" style={`--i:${i}`}>
					<span class={`dot ring-2 ${actionDot(log.Action)}`}></span>
					<div class="entry-body">
						<div class="entry-top">
							<span class={`action ${actionColor(log.Action)}`}>{log.Action}</span>
							<span class="when"><TimeAgo date={log.Timestamp} /></span>
						</div>
						<p class="meta">
							by {log.User?.Username || 'system'}
							{#if changeSummary(log.Changes)}
								<span class="changes" title={log.Changes}>· {changeSummary(log.Changes)}</span>
							{/if}
						</p>
					</div>
				</li>
			{/each}
		</ul>
	{/if}
</div>

<style>
	.activity-history {
		display: flex;
		flex-direction: column;
		gap: 0.75rem;
	}

	.timeline {
		list-style: none;
		margin: 0;
		padding: 0 0 0 0.25rem;
		display: flex;
		flex-direction: column;
		position: relative;
	}

	.timeline::before {
		content: '';
		position: absolute;
		left: 0.55rem;
		top: 0.4rem;
		bottom: 0.4rem;
		width: 1px;
		background: linear-gradient(to bottom, rgba(15, 23, 42, 0.12), rgba(15, 23, 42, 0.03));
	}

	.entry {
		position: relative;
		display: flex;
		gap: 0.6rem;
		align-items: flex-start;
		padding: 0.35rem 0.4rem 0.35rem 0;
		border-radius: 0.6rem;
		animation: rise 0.35s ease both;
		animation-delay: calc(var(--i) * 45ms);
		transition: background-color 0.15s ease;
	}

	.entry:hover {
		background: rgba(255, 255, 255, 0.55);
	}

	.dot {
		flex: none;
		width: 0.55rem;
		height: 0.55rem;
		margin-top: 0.3rem;
		border-radius: 9999px;
		transition: transform 0.15s ease;
	}

	.entry:hover .dot {
		transform: scale(1.25);
	}

	.entry-body {
		min-width: 0;
		flex: 1;
	}

	.entry-top {
		display: flex;
		align-items: baseline;
		justify-content: space-between;
		gap: 0.5rem;
	}

	.action {
		font-size: 0.72rem;
		font-weight: 700;
		letter-spacing: 0.02em;
	}

	.when {
		font-size: 0.68rem;
		color: rgba(51, 65, 85, 0.45);
		white-space: nowrap;
	}

	.meta {
		margin: 0.1rem 0 0;
		font-size: 0.72rem;
		color: rgba(51, 65, 85, 0.6);
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}

	.changes {
		color: rgba(51, 65, 85, 0.45);
		cursor: help;
	}

	.empty {
		margin: 0;
		padding: 0.6rem 0.75rem;
		font-size: 0.75rem;
		color: rgba(51, 65, 85, 0.5);
		border: 1px dashed rgba(15, 23, 42, 0.12);
		border-radius: 0.6rem;
		background: rgba(255, 255, 255, 0.35);
	}

	.skeleton-row {
		display: flex;
		align-items: center;
		gap: 0.6rem;
		padding: 0.45rem 0;
		animation: rise 0.3s ease both;
		animation-delay: calc(var(--i) * 60ms);
	}

	.skeleton-dot {
		flex: none;
		width: 0.55rem;
		height: 0.55rem;
		border-radius: 9999px;
		background: rgba(15, 23, 42, 0.08);
	}

	.skeleton-bar {
		height: 0.5rem;
		border-radius: 9999px;
		background: linear-gradient(
			90deg,
			rgba(15, 23, 42, 0.06) 25%,
			rgba(15, 23, 42, 0.12) 50%,
			rgba(15, 23, 42, 0.06) 75%
		);
		background-size: 200% 100%;
		animation: shimmer 1.2s linear infinite;
	}

	@keyframes rise {
		from {
			opacity: 0;
			transform: translateY(4px);
		}
		to {
			opacity: 1;
			transform: translateY(0);
		}
	}

	@keyframes shimmer {
		from {
			background-position: 200% 0;
		}
		to {
			background-position: -200% 0;
		}
	}

	@media (prefers-reduced-motion: reduce) {
		.entry,
		.skeleton-row {
			animation: none;
		}
		.skeleton-bar {
			animation: none;
		}
	}
</style>
