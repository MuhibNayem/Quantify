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
	import Select from '$lib/components/ui/select/select.svelte';
	import {
		Table,
		TableBody,
		TableCell,
		TableHead,
		TableHeader,
		TableRow
	} from '$lib/components/ui/table';
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
				return 'bg-emerald-100 text-emerald-700 border-emerald-200';
			case 'UPDATE':
			case 'SETTING_UPDATED':
			case 'ROLE_PERMISSIONS_UPDATED':
				return 'bg-blue-100 text-blue-700 border-blue-200';
			case 'DELETE':
				return 'bg-red-100 text-red-700 border-red-200';
			case 'VOID':
			case 'DISCOUNT':
				return 'bg-amber-100 text-amber-700 border-amber-200';
			default:
				return 'bg-slate-100 text-slate-700 border-slate-200';
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

<div class="relative min-h-screen overflow-hidden bg-slate-50/50 p-6 lg:p-10">
	<div
		class="absolute inset-0 -z-10 bg-[radial-gradient(ellipse_at_top,_var(--tw-gradient-stops))] from-indigo-100/20 via-slate-50/20 to-white/20"
	></div>

	<div class="mx-auto max-w-7xl space-y-8">
		<div class="flex flex-col gap-4 md:flex-row md:items-center md:justify-between">
			<div class="space-y-1">
				<h1
					class="bg-gradient-to-r from-slate-900 via-indigo-900 to-slate-900 bg-clip-text text-3xl font-bold tracking-tight text-transparent"
				>
					Audit Trail
				</h1>
				<p class="text-slate-500">Track every action and change across the system</p>
			</div>
			<Button
				variant="outline"
				class="gap-2"
				onclick={exportCsv}
				disabled={loading || logs.length === 0}
			>
				<Download class="h-4 w-4" />
				Export CSV
			</Button>
		</div>

		<div
			class="liquid-panel flex flex-col gap-4 rounded-2xl bg-gradient-to-br from-white/40 via-white/20 to-white/5 p-4 backdrop-blur-sm lg:flex-row lg:items-center"
		>
			<div class="w-full lg:w-52">
				<Select bind:value={actionFilter} options={actionOptions} placeholder="All Actions" />
			</div>
			<div class="w-full lg:w-48">
				<Select bind:value={entityFilter} options={entityOptions} placeholder="All Entities" />
			</div>
			<div class="w-full lg:w-40">
				<Input
					type="date"
					bind:value={start}
					placeholder="Start date"
					class="h-12 text-sm font-semibold"
				/>
			</div>
			<div class="w-full lg:w-40">
				<Input
					type="date"
					bind:value={end}
					placeholder="End date"
					class="h-12 text-sm font-semibold"
				/>
			</div>
			<div class="w-full lg:w-40">
				<Input
					type="number"
					bind:value={userSearch}
					placeholder="User ID"
					class="h-12 text-sm font-semibold"
				/>
			</div>
			<div class="flex items-center gap-2">
				<Button variant="outline" size="sm" onclick={applyFilters} class="gap-1">
					<Search class="h-3 w-3" />
					Filter
				</Button>
				<Button variant="ghost" size="sm" onclick={resetFilters}>Reset</Button>
			</div>
		</div>

		{#if loading}
			<div class="flex h-64 items-center justify-center">
				<Loader2 class="h-8 w-8 animate-spin text-indigo-500" />
			</div>
		{:else if logs.length === 0}
			<div class="rounded-2xl border border-white/60 bg-white/70 shadow-lg backdrop-blur-xl">
				<EmptyState
					icon={ShieldCheck}
					title="No audit logs found"
					description="Try adjusting your filters to see more activity."
					actionLabel="Reset Filters"
					onAction={resetFilters}
				/>
			</div>
		{:else}
			<div
				in:fly={{ y: 20, duration: 300 }}
				class="overflow-hidden rounded-2xl border border-white/60 bg-white/70 shadow-lg backdrop-blur-xl"
			>
				<Table>
					<TableHeader>
						<TableRow class="border-slate-100 bg-slate-50/50">
							<TableHead class="w-8"></TableHead>
							<TableHead class="font-bold text-slate-600">Timestamp</TableHead>
							<TableHead class="font-bold text-slate-600">User</TableHead>
							<TableHead class="font-bold text-slate-600">Action</TableHead>
							<TableHead class="font-bold text-slate-600">Entity</TableHead>
							<TableHead class="font-bold text-slate-600">Changes</TableHead>
							<TableHead class="font-bold text-slate-600">IP Address</TableHead>
						</TableRow>
					</TableHeader>
					<TableBody>
						{#each logs as log (log.ID)}
							{@const changes = parseChanges(log.Changes)}
							<TableRow
								class="cursor-pointer border-slate-50 transition-colors hover:bg-indigo-50/30"
								onclick={() => toggleExpand(log.ID)}
							>
								<TableCell>
									<ChevronDown
										class={cn(
											'h-4 w-4 text-slate-400 transition-transform',
											expandedId === log.ID && 'rotate-180'
										)}
									/>
								</TableCell>
								<TableCell class="whitespace-nowrap text-slate-500">
									<TimeAgo date={log.Timestamp} />
								</TableCell>
								<TableCell>
									<div class="flex flex-col">
										<span class="font-semibold text-slate-800">
											{log.User?.Username || `User #${log.UserID}`}
										</span>
										{#if log.User?.Email}
											<span class="text-xs text-slate-400">{log.User.Email}</span>
										{/if}
									</div>
								</TableCell>
								<TableCell>
									<span
										class={cn(
											'inline-flex items-center rounded-full border px-2.5 py-0.5 text-xs font-bold tracking-wide uppercase',
											getActionColor(log.Action)
										)}
									>
										{log.Action}
									</span>
								</TableCell>
								<TableCell>
									<div class="flex flex-col">
										<span class="font-medium text-slate-700">{log.Entity}</span>
										<span class="font-mono text-xs text-slate-400">#{log.EntityID}</span>
									</div>
								</TableCell>
								<TableCell class="text-slate-500">
									{#if changes.length > 0}
										<span class="text-xs font-medium">
											{changes.length} field{changes.length === 1 ? '' : 's'} changed
										</span>
									{:else}
										<span class="text-xs text-slate-400">—</span>
									{/if}
								</TableCell>
								<TableCell class="font-mono text-xs text-slate-400">
									{log.IPAddress || '—'}
								</TableCell>
							</TableRow>

							{#if expandedId === log.ID}
								<TableRow class="bg-slate-50/60 hover:bg-slate-50/60">
									<TableCell colspan={7} class="p-0">
										<div
											in:fly={{ y: -8, duration: 200 }}
											class="space-y-4 px-6 py-4"
										>
											{#if changes.length > 0}
												<div class="space-y-2">
													<p class="text-xs font-bold tracking-wide text-slate-500 uppercase">
														Changes
													</p>
													<div
														class="grid gap-2 sm:grid-cols-2 lg:grid-cols-3"
													>
														{#each changes as change}
															<div
																class="rounded-lg border border-slate-200 bg-white/80 px-3 py-2"
															>
																<p class="font-mono text-xs font-bold text-slate-600">
																	{change.key}
																</p>
																<div class="mt-1 flex flex-wrap items-center gap-2 text-sm">
																	{#if change.from !== undefined}
																		<span
																			class="rounded bg-red-50 px-1.5 py-0.5 font-mono text-xs text-red-600 line-through"
																		>
																			{formatValue(change.from)}
																		</span>
																		<ArrowRight class="h-3 w-3 text-slate-400" />
																	{/if}
																	<span
																		class="rounded bg-emerald-50 px-1.5 py-0.5 font-mono text-xs text-emerald-700"
																	>
																		{formatValue(change.to)}
																	</span>
																</div>
															</div>
														{/each}
													</div>
												</div>
											{:else}
												<p class="text-sm text-slate-400">No structured changes recorded.</p>
											{/if}

											<div class="grid gap-3 text-xs text-slate-500 sm:grid-cols-2">
												<div>
													<span class="font-bold text-slate-600">Full timestamp:</span>
													{formatDate(log.Timestamp)}
												</div>
												{#if log.UserAgent}
													<div class="truncate">
														<span class="font-bold text-slate-600">User agent:</span>
														{log.UserAgent}
													</div>
												{/if}
											</div>
										</div>
									</TableCell>
								</TableRow>
							{/if}
						{/each}
					</TableBody>
				</Table>
			</div>

			{#if totalPages > 1}
				<div class="flex items-center justify-between">
					<p class="text-sm text-slate-500">
						Showing {logs.length} of {total} events
					</p>
					<div class="flex items-center gap-2">
						<Button
							variant="outline"
							size="sm"
							disabled={page <= 1}
							onclick={() => goToPage(page - 1)}
						>
							<ChevronLeft class="h-4 w-4" />
						</Button>
						<span class="text-sm font-medium text-slate-700">
							Page {page} of {totalPages}
						</span>
						<Button
							variant="outline"
							size="sm"
							disabled={page >= totalPages}
							onclick={() => goToPage(page + 1)}
						>
							<ChevronRight class="h-4 w-4" />
						</Button>
					</div>
				</div>
			{/if}
		{/if}
	</div>
</div>
