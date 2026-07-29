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
		'liquid-input h-10 w-full rounded-md bg-card bg-none px-3 text-sm font-normal text-foreground shadow-none ring-0 backdrop-blur-none backdrop-saturate-100 from-card to-card placeholder:text-muted-foreground focus:scale-100 focus:border-primary focus:ring-0 md:text-sm';
	const btnSecondary =
		'rounded-md border border-border bg-card bg-none text-foreground shadow-none backdrop-blur-none hover:bg-muted hover:bg-none';

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

<div class="relative min-h-screen overflow-hidden bg-background p-6 lg:p-10">
	<div class="mx-auto max-w-7xl space-y-8">
		<div class="flex flex-col gap-4 md:flex-row md:items-center md:justify-between">
			<div class="space-y-1">
				<h1 class="font-display text-foreground text-3xl font-bold tracking-tight">
					Audit Trail
				</h1>
				<p class="text-muted-foreground">Track every action and change across the system</p>
			</div>
			<Button
				variant="outline"
				class="{btnSecondary} gap-2"
				onclick={exportCsv}
				disabled={loading || logs.length === 0}
			>
				<Download class="h-4 w-4" />
				Export CSV
			</Button>
		</div>

		<div
			class="liquid-panel flex flex-col gap-4 rounded-lg p-4 lg:flex-row lg:items-center"
		>
			<div class="w-full lg:w-52">
				<div class="select-wrapper">
					<select bind:value={actionFilter}>
						{#each actionOptions as opt}
							<option value={opt.value}>{opt.label}</option>
						{/each}
					</select>
				</div>
			</div>
			<div class="w-full lg:w-48">
				<div class="select-wrapper">
					<select bind:value={entityFilter}>
						{#each entityOptions as opt}
							<option value={opt.value}>{opt.label}</option>
						{/each}
					</select>
				</div>
			</div>
			<div class="w-full lg:w-40">
				<Input type="date" bind:value={start} placeholder="Start date" class="{inputClass} num-mono" />
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
				<Button variant="outline" size="sm" onclick={applyFilters} class="{btnSecondary} gap-1">
					<Search class="h-3 w-3" />
					Filter
				</Button>
				<Button variant="ghost" size="sm" onclick={resetFilters} class="rounded-md">Reset</Button>
			</div>
		</div>

		{#if loading}
			<div class="flex h-64 items-center justify-center">
				<Loader2 class="text-primary h-8 w-8 animate-spin" />
			</div>
		{:else if logs.length === 0}
			<div class="liquid-panel rounded-lg">
				<EmptyState
					icon={ShieldCheck}
					title="No audit logs found"
					description="Try adjusting your filters to see more activity."
					actionLabel="Reset Filters"
					onAction={resetFilters}
				/>
			</div>
		{:else}
			<div in:fly={{ y: 20, duration: 300 }} class="liquid-panel overflow-hidden rounded-lg">
				<table class="ledger-table">
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
											'text-muted-foreground h-4 w-4 transition-transform',
											expandedId === log.ID && 'rotate-180'
										)}
									/>
								</td>
								<td class="text-muted-foreground whitespace-nowrap num-mono">
									<TimeAgo date={log.Timestamp} />
								</td>
								<td>
									<div class="flex flex-col">
										<span class="text-foreground font-semibold">
											{log.User?.Username || `User #${log.UserID}`}
										</span>
										{#if log.User?.Email}
											<span class="text-muted-foreground text-xs">{log.User.Email}</span>
										{/if}
									</div>
								</td>
								<td>
									<span class={cn(getActionColor(log.Action), 'uppercase tracking-wide')}>
										{log.Action}
									</span>
								</td>
								<td>
									<div class="flex flex-col">
										<span class="text-foreground font-medium">{log.Entity}</span>
										<span class="text-muted-foreground text-xs num-mono">#{log.EntityID}</span>
									</div>
								</td>
								<td class="text-muted-foreground">
									{#if changes.length > 0}
										<span class="text-xs font-medium">
											{changes.length} field{changes.length === 1 ? '' : 's'} changed
										</span>
									{:else}
										<span class="text-muted-foreground text-xs">—</span>
									{/if}
								</td>
								<td class="text-muted-foreground text-xs num-mono">
									{log.IPAddress || '—'}
								</td>
							</tr>

							{#if expandedId === log.ID}
								<tr class="bg-muted/50">
									<td colspan={7} class="p-0">
										<div in:fly={{ y: -8, duration: 200 }} class="space-y-4 px-6 py-4">
											{#if changes.length > 0}
												<div class="space-y-2">
													<p class="overline">
														Changes
													</p>
													<div class="grid gap-2 sm:grid-cols-2 lg:grid-cols-3">
														{#each changes as change}
															<div
																class="border-border bg-card rounded-md border px-3 py-2"
															>
																<p class="text-muted-foreground text-xs font-semibold num-mono">
																	{change.key}
																</p>
																<div class="mt-1 flex flex-wrap items-center gap-2 text-sm">
																	{#if change.from !== undefined}
																		<span
																			class="bg-danger/10 text-danger rounded px-1.5 py-0.5 text-xs line-through num-mono"
																		>
																			{formatValue(change.from)}
																		</span>
																		<ArrowRight class="text-muted-foreground h-3 w-3" />
																	{/if}
																	<span
																		class="bg-success/10 text-success rounded px-1.5 py-0.5 text-xs num-mono"
																	>
																		{formatValue(change.to)}
																	</span>
																</div>
															</div>
														{/each}
													</div>
												</div>
											{:else}
												<p class="text-muted-foreground text-sm">No structured changes recorded.</p>
											{/if}

											<div class="text-muted-foreground grid gap-3 text-xs sm:grid-cols-2">
												<div>
													<span class="text-foreground font-semibold">Full timestamp:</span>
													{formatDate(log.Timestamp)}
												</div>
												{#if log.UserAgent}
													<div class="truncate">
														<span class="text-foreground font-semibold">User agent:</span>
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
				<div class="flex items-center justify-between">
					<p class="text-muted-foreground text-sm">
						Showing {logs.length} of {total} events
					</p>
					<div class="flex items-center gap-2">
						<Button
							variant="outline"
							size="sm"
							disabled={page <= 1}
							onclick={() => goToPage(page - 1)}
							class={btnSecondary}
						>
							<ChevronLeft class="h-4 w-4" />
						</Button>
						<span class="text-foreground text-sm font-medium num-mono">
							Page {page} of {totalPages}
						</span>
						<Button
							variant="outline"
							size="sm"
							disabled={page >= totalPages}
							onclick={() => goToPage(page + 1)}
							class={btnSecondary}
						>
							<ChevronRight class="h-4 w-4" />
						</Button>
					</div>
				</div>
			{/if}
		{/if}
	</div>
</div>
