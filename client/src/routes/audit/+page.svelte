<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { auditApi } from '$lib/api/resources';
	import { auth } from '$lib/stores/auth';
	import { cn, formatDate } from '$lib/utils';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import TimeAgo from '$lib/components/ui/time-ago.svelte';
	import EmptyState from '$lib/components/ui/empty-state.svelte';
	import {
		ShieldCheck,
		Search,
		ChevronLeft,
		ChevronRight,
		ChevronDown,
		Loader2,
		Download,
		ArrowRight
	} from 'lucide-svelte';
	import { toast } from 'svelte-sonner';
	import { fly } from 'svelte/transition';
	import type { AuditLog } from '$lib/types';

	const inputClass =
		'glass-input h-10 w-full px-3 text-sm font-normal text-foreground placeholder:text-muted-foreground';
	const btnSecondary = 'shadow-none';

	let logs = $state<AuditLog[]>([]);
	let loading = $state(true);
	let total = $state(0);
	let page = $state(1);
	let limit = $state(25);

	let actionFilter = $state('');
	let entityFilter = $state('');
	let userSearch = $state('');
	let start = $state('');
	let end = $state('');

	let expandedId = $state<number | null>(null);

	const totalPages = $derived(Math.ceil(total / limit) || 1);

	const actionOptions = [
		{ value: '', label: 'All Actions' },
		{ value: 'CREATE', label: 'Create' },
		{ value: 'UPDATE', label: 'Update' },
		{ value: 'DELETE', label: 'Delete' },
		{ value: 'VOID', label: 'Void' },
		{ value: 'DISCOUNT', label: 'Discount' },
		{ value: 'SETTING_UPDATED', label: 'Setting Updated' },
		{ value: 'ROLE_PERMISSIONS_UPDATED', label: 'Role Permissions Updated' }
	];

	const entityOptions = [
		{ value: '', label: 'All Entities' },
		{ value: 'Product', label: 'Product' },
		{ value: 'Order', label: 'Order' },
		{ value: 'User', label: 'User' },
		{ value: 'Settings', label: 'Settings' },
		{ value: 'Role', label: 'Role' }
	];

	$effect(() => {
		if (!auth.hasPermission('reports.financial')) {
			toast.error('Access Denied', {
				description: 'You do not have permission to view the audit trail.'
			});
			goto('/');
		}
	});

	onMount(() => {
		loadLogs();
	});

	async function loadLogs() {
		loading = true;
		try {
			const params: Record<string, unknown> = { page, limit };
			if (actionFilter) params.action = actionFilter;
			if (entityFilter) params.entity = entityFilter;
			if (userSearch.trim()) params.userId = Number(userSearch) || undefined;
			if (start) params.start = start;
			if (end) params.end = end;

			const res = await auditApi.list(params);
			logs = res.data || res.items || [];
			total = res.total || 0;
		} catch (e: any) {
			console.error(e);
			toast.error('Failed to load audit logs', {
				description: e?.response?.data?.error || e.message
			});
		} finally {
			loading = false;
		}
	}

	function applyFilters() {
		page = 1;
		loadLogs();
	}

	function resetFilters() {
		actionFilter = '';
		entityFilter = '';
		userSearch = '';
		start = '';
		end = '';
		page = 1;
		loadLogs();
	}

	function goToPage(p: number) {
		if (p < 1 || p > totalPages) return;
		page = p;
		loadLogs();
	}

	function toggleExpand(id: number) {
		expandedId = expandedId === id ? null : id;
	}

	function getActionColor(action: string) {
		switch (action) {
			case 'CREATE':
				return 'badge-success';
			case 'UPDATE':
			case 'SETTING_UPDATED':
			case 'ROLE_PERMISSIONS_UPDATED':
				return 'badge-info';
			case 'DELETE':
				return 'badge-danger';
			case 'VOID':
			case 'DISCOUNT':
				return 'badge-warning';
			default:
				return 'badge-neutral';
		}
	}

	function parseChanges(json: string): { key: string; from?: unknown; to?: unknown }[] {
		if (!json) return [];
		try {
			const obj = JSON.parse(json);
			if (obj && typeof obj === 'object' && obj.old && obj.new) {
				return Object.keys(obj.new).map((key) => ({
					key,
					from: obj.old[key],
					to: obj.new[key]
				}));
			}
			return Object.entries(obj).map(([key, to]) => ({ key, to }));
		} catch {
			return [];
		}
	}

	function formatValue(value: unknown): string {
		if (value === null || value === undefined) return '—';
		if (typeof value === 'object') return JSON.stringify(value);
		return String(value);
	}

	function exportCsv() {
		if (logs.length === 0) {
			toast.error('Nothing to export');
			return;
		}
		const headers = [
			'ID',
			'Timestamp',
			'Action',
			'Entity',
			'EntityID',
			'User',
			'Email',
			'IPAddress',
			'UserAgent',
			'Changes'
		];
		const escape = (val: unknown) => {
			const s = val === null || val === undefined ? '' : String(val);
			return `"${s.replace(/"/g, '""')}"`;
		};
		const rows = logs.map((log) =>
			[
				log.ID,
				log.Timestamp,
				log.Action,
				log.Entity,
				log.EntityID,
				log.User?.Username || '',
				log.User?.Email || '',
				log.IPAddress,
				log.UserAgent,
				log.Changes
			]
				.map(escape)
				.join(',')
		);
		const csv = [headers.map(escape).join(','), ...rows].join('\n');
		const blob = new Blob([csv], { type: 'text/csv;charset=utf-8;' });
		const url = URL.createObjectURL(blob);
		const a = document.createElement('a');
		a.href = url;
		a.download = `audit-logs-${new Date().toISOString().slice(0, 10)}.csv`;
		a.click();
		URL.revokeObjectURL(url);
		toast.success('Audit logs exported');
	}
</script>

<div class="min-h-screen bg-background p-6 lg:p-8">
	<div class="mx-auto max-w-7xl space-y-6">
		<header class="page-header">
			<div class="min-w-0">
				<h1 class="page-title">Audit Trail</h1>
				<p class="page-sub">Track every action and change across the system</p>
			</div>
			<Button
				variant="outline"
				class="{btnSecondary} h-10 gap-2 px-4 text-sm"
				onclick={exportCsv}
				disabled={loading || logs.length === 0}
			>
				<Download class="h-4 w-4" />
				Export CSV
			</Button>
		</header>

		<div class="panel flex flex-col gap-3 p-4 lg:flex-row lg:items-center">
			<div class="w-full lg:w-52">
				<div class="select-wrapper">
					<select bind:value={actionFilter} class="h-10">
						{#each actionOptions as opt}
							<option value={opt.value}>{opt.label}</option>
						{/each}
					</select>
				</div>
			</div>
			<div class="w-full lg:w-48">
				<div class="select-wrapper">
					<select bind:value={entityFilter} class="h-10">
						{#each entityOptions as opt}
							<option value={opt.value}>{opt.label}</option>
						{/each}
					</select>
				</div>
			</div>
			<div class="w-full lg:w-40">
				<Input
					type="date"
					bind:value={start}
					placeholder="Start date"
					class="{inputClass} num-mono"
				/>
			</div>
			<div class="w-full lg:w-40">
				<Input type="date" bind:value={end} placeholder="End date" class="{inputClass} num-mono" />
			</div>
			<div class="w-full lg:w-40">
				<Input
					type="number"
					bind:value={userSearch}
					placeholder="User ID"
					class="{inputClass} num-mono"
				/>
			</div>
			<div class="flex items-center gap-2">
				<Button
					variant="outline"
					class="{btnSecondary} h-10 gap-2 px-4 text-sm"
					onclick={applyFilters}
				>
					<Search class="h-3.5 w-3.5" />
					Filter
				</Button>
				<Button variant="ghost" class="h-10 px-3 text-sm" onclick={resetFilters}>Reset</Button>
			</div>
		</div>

		{#if loading}
			<div class="panel flex h-64 items-center justify-center">
				<Loader2 class="h-6 w-6 animate-spin text-primary" />
			</div>
		{:else if logs.length === 0}
			<div class="panel">
				<EmptyState
					icon={ShieldCheck}
					title="No audit logs found"
					description="Try adjusting your filters to see more activity."
					actionLabel="Reset Filters"
					onAction={resetFilters}
				/>
			</div>
		{:else}
			<div in:fly={{ y: 20, duration: 300 }} class="panel overflow-x-auto">
				<table class="ledger-table min-w-[64rem]">
					<thead>
						<tr>
							<th class="w-8"></th>
							<th>Timestamp</th>
							<th>User</th>
							<th>Action</th>
							<th>Entity</th>
							<th>Changes</th>
							<th>IP Address</th>
						</tr>
					</thead>
					<tbody>
						{#each logs as log (log.ID)}
							{@const changes = parseChanges(log.Changes)}
							<tr class="cursor-pointer" onclick={() => toggleExpand(log.ID)}>
								<td>
									<ChevronDown
										class={cn(
											'h-4 w-4 text-muted-foreground transition-transform',
											expandedId === log.ID && 'rotate-180'
										)}
									/>
								</td>
								<td class="num-mono whitespace-nowrap text-muted-foreground">
									<TimeAgo date={log.Timestamp} />
								</td>
								<td>
									<div class="flex flex-col">
										<span class="font-semibold text-foreground">
											{log.User?.Username || `User #${log.UserID}`}
										</span>
										{#if log.User?.Email}
											<span class="text-xs text-muted-foreground">{log.User.Email}</span>
										{/if}
									</div>
								</td>
								<td>
									<span class={cn(getActionColor(log.Action), 'tracking-wide uppercase')}>
										{log.Action}
									</span>
								</td>
								<td>
									<div class="flex flex-col">
										<span class="font-medium text-foreground">{log.Entity}</span>
										<span class="num-mono text-xs text-muted-foreground">#{log.EntityID}</span>
									</div>
								</td>
								<td class="text-muted-foreground">
									{#if changes.length > 0}
										<span class="text-xs font-medium">
											{changes.length} field{changes.length === 1 ? '' : 's'} changed
										</span>
									{:else}
										<span class="text-xs text-muted-foreground">—</span>
									{/if}
								</td>
								<td class="num-mono text-xs text-muted-foreground">
									{log.IPAddress || '—'}
								</td>
							</tr>

							{#if expandedId === log.ID}
								<tr class="bg-muted/50">
									<td colspan={7} class="p-0">
										<div in:fly={{ y: -8, duration: 200 }} class="space-y-4 px-6 py-4">
											{#if changes.length > 0}
												<div class="space-y-2">
													<p class="overline">Changes</p>
													<div class="grid gap-2 sm:grid-cols-2 lg:grid-cols-3">
														{#each changes as change}
															<div class="rounded-md border border-border bg-card px-3 py-2">
																<p class="num-mono text-xs font-semibold text-muted-foreground">
																	{change.key}
																</p>
																<div class="mt-1 flex flex-wrap items-center gap-2 text-sm">
																	{#if change.from !== undefined}
																		<span
																			class="num-mono rounded bg-danger/10 px-1.5 py-0.5 text-xs text-danger line-through"
																		>
																			{formatValue(change.from)}
																		</span>
																		<ArrowRight class="h-3 w-3 text-muted-foreground" />
																	{/if}
																	<span
																		class="num-mono rounded bg-success/10 px-1.5 py-0.5 text-xs text-success"
																	>
																		{formatValue(change.to)}
																	</span>
																</div>
															</div>
														{/each}
													</div>
												</div>
											{:else}
												<p class="text-sm text-muted-foreground">No structured changes recorded.</p>
											{/if}

											<div class="grid gap-3 text-xs text-muted-foreground sm:grid-cols-2">
												<div>
													<span class="font-semibold text-foreground">Full timestamp:</span>
													{formatDate(log.Timestamp)}
												</div>
												{#if log.UserAgent}
													<div class="truncate">
														<span class="font-semibold text-foreground">User agent:</span>
														{log.UserAgent}
													</div>
												{/if}
											</div>
										</div>
									</td>
								</tr>
							{/if}
						{/each}
					</tbody>
				</table>
			</div>

			{#if totalPages > 1}
				<div class="flex flex-wrap items-center justify-between gap-3">
					<p class="num-mono text-xs text-muted-foreground">
						Showing {logs.length} of {total} events
					</p>
					<div class="flex items-center gap-2">
						<Button
							variant="outline"
							size="icon"
							disabled={page <= 1}
							onclick={() => goToPage(page - 1)}
							class="{btnSecondary} size-9"
							aria-label="Previous page"
						>
							<ChevronLeft class="h-4 w-4" />
						</Button>
						<span class="num-mono text-xs font-medium text-foreground">
							Page {page} of {totalPages}
						</span>
						<Button
							variant="outline"
							size="icon"
							disabled={page >= totalPages}
							onclick={() => goToPage(page + 1)}
							class="{btnSecondary} size-9"
							aria-label="Next page"
						>
							<ChevronRight class="h-4 w-4" />
						</Button>
					</div>
				</div>
			{/if}
		{/if}
	</div>
</div>
