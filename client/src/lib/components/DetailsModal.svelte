<script lang="ts" context="module">
	type Accent = 'sky' | 'violet' | 'emerald' | 'amber' | 'slate' | 'rose';

	export type SummaryCard = {
		title: string;
		value: string | number | null;
		hint?: string;
		icon?: ComponentType | null;
		accent?: Accent;
	};

	export type DescriptionSection = {
		type: 'description';
		title?: string;
		description?: string;
		columns?: 1 | 2 | 3;
		items: DescriptionItem[];
	};

	export type TableColumn = {
		key: string;
		label: string;
		align?: 'left' | 'center' | 'right';
		formatter?: (value: unknown, row: Record<string, unknown>) => string | number | null;
	};

	export type TableSection = {
		type: 'table';
		title?: string;
		description?: string;
		columns: TableColumn[];
		rows: Record<string, unknown>[];
		emptyText?: string;
	};

	export type SummarySection = {
		type: 'summary';
		cards: SummaryCard[];
	};

	export type DetailSection = SummarySection | DescriptionSection | TableSection;

	export type DetailBuilderContext = {
		data: Record<string, unknown>;
		extras: Record<string, unknown>;
	};

	export type DetailExtraFetcher = {
		key: string;
		request: (resourceId: string | number) => Promise<unknown>;
	};
</script>

<script lang="ts">
	import type { ComponentType } from 'svelte';
	import { createEventDispatcher, onMount } from 'svelte';
	import { AlertCircle, RefreshCcw, X } from 'lucide-svelte';
	import { fade, fly } from 'svelte/transition';
	import { cubicOut } from 'svelte/easing';
	import api from '$lib/api';
	import { cn } from '$lib/utils';
	import { Button } from '$lib/components/ui/button';
	import DescriptionList from '$lib/components/ui/description-list';
	import type { DescriptionItem } from '$lib/components/ui/description-list/description-list.svelte';
	import { Skeleton } from '$lib/components/ui/skeleton';

	export let open = false;
	export let resourceId: string | number | null = null;
	export let endpoint: string;
	export let title = 'Details';
	export let subtitle: string | null = null;
	export let extraFetchers: DetailExtraFetcher[] = [];
	export let buildSections: (ctx: DetailBuilderContext) => DetailSection[] = () => [];
	export let showRefresh = true;

	const dispatch = createEventDispatcher<{ close: void }>();

	let loading = false;
	let error: string | null = null;
	let data: Record<string, unknown> | null = null;
	let extras: Record<string, unknown> = {};
	let sections: DetailSection[] = [];
	let requestToken: symbol | null = null;

	const cardAccents: Record<Accent, string> = {
		sky: 'text-info',
		violet: 'text-primary',
		emerald: 'text-success',
		amber: 'text-warning',
		slate: 'text-muted-foreground',
		rose: 'text-danger'
	};

	function resolveAccent(accent?: Accent) {
		return cardAccents[accent ?? 'sky'];
	}

	function alignClass(align?: 'left' | 'center' | 'right') {
		switch (align) {
			case 'center':
				return 'text-center';
			case 'right':
				return 'text-right';
			default:
				return 'text-left';
		}
	}

	function resolveCellValue(row: Record<string, unknown>, column: TableColumn) {
		const value =
			column.key.split('.').reduce<unknown>((acc, key) => {
				if (acc && typeof acc === 'object' && key in acc) {
					return (acc as Record<string, unknown>)[key];
				}
				return undefined;
			}, row) ?? null;

		const formatted = column.formatter ? column.formatter(value, row) : value;
		return formatted ?? '—';
	}

	function resetState() {
		data = null;
		extras = {};
		sections = [];
		error = null;
		requestToken = null;
	}

	function closeDrawer() {
		open = false;
		resetState();
		resourceId = null;
		dispatch('close');
	}

	async function fetchData() {
		if (!open || !resourceId || !endpoint) return;
		const id = resourceId;
		const token = Symbol('details-fetch');
		requestToken = token;
		loading = true;
		error = null;

		try {
			const { data: payload } = await api.get(`${endpoint}/${id}`);
			if (requestToken !== token) return;

			const extraEntries = await Promise.all(
				extraFetchers.map(async (fetcher) => {
					try {
						const value = await fetcher.request(id);
						return [fetcher.key, value] as const;
					} catch {
						return [fetcher.key, null] as const;
					}
				})
			);

			data = payload ?? null;
			extras = Object.fromEntries(extraEntries);
		} catch (err: any) {
			if (requestToken !== token) return;
			error = err?.response?.data?.error ?? err?.message ?? 'Unable to load details';
			data = null;
			extras = {};
		} finally {
			if (requestToken === token) loading = false;
		}
	}

	onMount(() => {
		const onKey = (e: KeyboardEvent) => {
			if (e.key === 'Escape' && open) closeDrawer();
		};
		window.addEventListener('keydown', onKey);
		return () => window.removeEventListener('keydown', onKey);
	});

	$: if (open && resourceId && endpoint) fetchData();
	$: if (data) {
		try {
			sections = buildSections({ data, extras }) ?? [];
		} catch {
			sections = [];
		}
	} else sections = [];
</script>

{#if open}
	<div
		class="fixed inset-0 z-40 bg-foreground/40"
		role="button"
		tabindex="0"
		aria-label="Close details"
		on:click={(e) => e.target === e.currentTarget && closeDrawer()}
		on:keydown={(e) => {
			if (e.target === e.currentTarget && (e.key === 'Enter' || e.key === ' ')) {
				e.preventDefault();
				closeDrawer();
			}
		}}
		transition:fade={{ duration: 250, easing: cubicOut }}
	>
		<div
			class="detail-drawer absolute top-0 right-0 flex h-full w-full flex-col rounded-l-lg border-l border-border/80 bg-card shadow-xl sm:w-[600px] md:w-[700px]"
			transition:fly={{ x: 400, duration: 350, delay: 80, easing: cubicOut }}
		>
			<div class="flex items-start justify-between gap-4 border-b border-border/80 p-5 sm:p-6">
				<div class="min-w-0">
					<h2
						class="font-display text-xl font-semibold tracking-tight break-words text-foreground sm:text-2xl"
					>
						{title}
					</h2>
					{#if subtitle}
						<p class="mt-1 text-sm break-words text-muted-foreground">{subtitle}</p>
					{/if}
				</div>

				<div class="flex shrink-0 items-center gap-1">
					{#if showRefresh && resourceId}
						<Button
							variant="ghost"
							size="icon"
							class="h-9 w-9 rounded-md text-muted-foreground transition-all active:scale-95"
							onclick={fetchData}
							disabled={loading}
							aria-label="Refresh details"
						>
							<RefreshCcw class={cn('h-4 w-4', loading && 'animate-spin')} />
						</Button>
					{/if}
					<Button
						variant="ghost"
						size="icon"
						class="h-9 w-9 rounded-md text-muted-foreground transition-all hover:text-foreground"
						onclick={closeDrawer}
						aria-label="Close details"
					>
						<X class="h-5 w-5" />
					</Button>
				</div>
			</div>

			<!-- Body -->
			<div class="flex-1 space-y-5 overflow-y-auto p-5 sm:p-6">
				{#if loading}
					<Skeleton class="h-8 w-3/4 rounded-md" />
					<Skeleton class="h-4 w-1/2 rounded-md" />
					<Skeleton class="h-40 w-full rounded-lg" />
				{:else if error}
					<div class="flex items-start gap-3 rounded-lg border border-danger/25 bg-danger/5 p-4">
						<div class="rounded-md bg-danger/10 p-2 text-danger">
							<AlertCircle class="h-5 w-5" />
						</div>
						<div class="min-w-0">
							<p class="section-title break-words text-danger">Unable to load details</p>
							<p class="mt-1 text-sm break-words text-muted-foreground">{error}</p>
						</div>
					</div>
				{:else if data}
					{#each sections as section, index (index)}
						{#if section.type === 'summary'}
							<div class="grid gap-3 sm:grid-cols-2 md:grid-cols-3">
								{#each section.cards as card (card.title)}
									<div
										class="surface-1 min-w-0 p-4 break-words transition-shadow duration-200 hover:shadow-md"
									>
										<p class="flex items-center gap-2 break-words overline">
											{#if card.icon}
												<svelte:component
													this={card.icon}
													class={cn('h-4 w-4 shrink-0', resolveAccent(card.accent))}
												/>
											{/if}
											<span class="truncate">{card.title}</span>
										</p>
										<p
											class="kpi-num mt-2 min-w-0 overflow-hidden text-[clamp(1.15rem,2.6vw,1.625rem)] [overflow-wrap:anywhere] break-words text-foreground"
										>
											{card.value ?? '—'}
										</p>
										{#if card.hint}
											<p class="mt-1 text-xs break-words text-muted-foreground">{card.hint}</p>
										{/if}
									</div>
								{/each}
							</div>
						{:else if section.type === 'description'}
							<section class="rounded-lg bg-muted/35 p-4 sm:p-5">
								<h3 class="section-title break-words">{section.title ?? 'Overview'}</h3>
								{#if section.description}
									<p class="mt-1 text-sm break-words text-muted-foreground">
										{section.description}
									</p>
								{/if}
								<div class="mt-3">
									<DescriptionList items={section.items} columns={section.columns ?? 2} />
								</div>
							</section>
						{:else if section.type === 'table'}
							<section class="overflow-hidden rounded-lg bg-muted/35">
								<div class="px-4 pt-4 sm:px-5">
									<h3 class="section-title break-words">{section.title ?? 'Records'}</h3>
									{#if section.description}
										<p class="mt-1 text-sm break-words text-muted-foreground">
											{section.description}
										</p>
									{/if}
								</div>
								<div class="mt-3 overflow-x-auto">
									<table class="ledger-table min-w-full">
										<thead>
											<tr>
												{#each section.columns as column (column.key)}
													<th class={cn('break-words', alignClass(column.align))}>
														{column.label}
													</th>
												{/each}
											</tr>
										</thead>
										<tbody>
											{#each section.rows as row, rowIndex (rowIndex)}
												<tr>
													{#each section.columns as column (column.key)}
														<td
															class={cn(
																'max-w-[1px] overflow-hidden text-sm break-all text-ellipsis text-foreground',
																alignClass(column.align)
															)}
														>
															{resolveCellValue(row, column)}
														</td>
													{/each}
												</tr>
											{/each}
										</tbody>
									</table>
								</div>
							</section>
						{/if}
					{/each}
				{/if}
			</div>
		</div>
	</div>
{/if}

<style>
	.detail-drawer * {
		word-wrap: break-word;
		overflow-wrap: anywhere;
		min-width: 0;
	}
</style>
