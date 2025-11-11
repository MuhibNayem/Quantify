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
	import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '$lib/components/ui/card';
	import DescriptionList from '$lib/components/ui/description-list';
	import type { DescriptionItem } from '$lib/components/ui/description-list/description-list.svelte';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '$lib/components/ui/table';

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
		sky: 'from-sky-50 via-blue-50 to-cyan-50 border-sky-100 hover:shadow-sky-100/50',
		violet: 'from-violet-50 via-purple-50 to-fuchsia-50 border-violet-100 hover:shadow-violet-100/50',
		emerald: 'from-emerald-50 via-green-50 to-lime-50 border-emerald-100 hover:shadow-emerald-100/50',
		amber: 'from-amber-50 via-orange-50 to-yellow-50 border-amber-100 hover:shadow-amber-100/50',
		slate: 'from-slate-50 via-gray-50 to-slate-50 border-slate-100 hover:shadow-slate-100/50',
		rose: 'from-rose-50 via-pink-50 to-rose-50 border-rose-100 hover:shadow-rose-100/50'
	};

	function resolveAccent(accent?: Accent) {
		return cardAccents[accent ?? 'sky'];
	}

	function alignClass(align?: 'left' | 'center' | 'right') {
		switch (align) {
			case 'center': return 'text-center';
			case 'right': return 'text-right';
			default: return 'text-left';
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
		const token = Symbol('details-fetch');
		requestToken = token;
		loading = true;
		error = null;

		try {
			const { data: payload } = await api.get(`${endpoint}/${resourceId}`);
			if (requestToken !== token) return;

			const extraEntries = await Promise.all(
				extraFetchers.map(async (fetcher) => {
					try {
						const value = await fetcher.request(resourceId);
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
		class="fixed inset-0 z-40 bg-slate-900/30 backdrop-blur-sm"
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
			class="absolute right-0 top-0 h-full w-full sm:w-[600px] md:w-[700px] bg-gradient-to-br from-white/90 via-sky-50/80 to-violet-50/70 border-l border-slate-200/70 shadow-2xl rounded-l-2xl flex flex-col detail-drawer"
			transition:fly={{ x: 400, duration: 350, delay: 80, easing: cubicOut }}
		>
			<div class="flex items-start justify-between p-6 border-b border-slate-100/70">
				<div class="min-w-0">
					<h2 class="text-2xl font-semibold bg-clip-text text-transparent bg-gradient-to-r from-violet-600 to-sky-600 break-words">
						{title}
					</h2>
					{#if subtitle}
						<p class="text-sm text-slate-600 mt-1 break-words">{subtitle}</p>
					{/if}
				</div>

				<div class="flex items-center gap-2">
					{#if showRefresh && resourceId}
						<Button
							variant="ghost"
							size="icon"
							class="rounded-full p-2 hover:bg-sky-100/60 active:scale-95 transition-all"
							onclick={fetchData}
							disabled={loading}
						>
							<RefreshCcw class={cn('h-4 w-4 text-sky-600', loading && 'animate-spin')} />
						</Button>
					{/if}
					<Button
						variant="ghost"
						size="icon"
						class="rounded-full p-2 hover:bg-rose-100/70 transition-all"
						onclick={closeDrawer}
					>
						<X class="h-5 w-5 text-rose-500" />
					</Button>
				</div>
			</div>

			<!-- Body -->
			<div class="p-6 overflow-y-auto flex-1 space-y-6">
				{#if loading}
					<Skeleton class="h-8 w-3/4 rounded-xl" />
					<Skeleton class="h-4 w-1/2 rounded-xl" />
				{:else if error}
					<Card class="border border-rose-100 bg-rose-50/70 shadow-none break-words">
						<CardHeader class="flex flex-row items-start gap-3">
							<div class="rounded-full bg-rose-100 p-2 text-rose-600">
								<AlertCircle class="h-5 w-5" />
							</div>
							<div>
								<CardTitle class="text-rose-700 break-words">Unable to load details</CardTitle>
								<CardDescription class="text-rose-600 break-words">{error}</CardDescription>
							</div>
						</CardHeader>
					</Card>
				{:else if data}
					{#each sections as section, index (index)}
						{#if section.type === 'summary'}
							<div class="grid gap-4 sm:grid-cols-2 md:grid-cols-3">
								{#each section.cards as card (card.title)}
									<Card
										class={cn(
											'min-w-0 border rounded-2xl bg-gradient-to-br transition-all duration-300 hover:-translate-y-1 hover:shadow-lg break-words',
											resolveAccent(card.accent)
										)}
									>
										<CardHeader class="space-y-1 min-w-0">
											<CardDescription class="text-xs uppercase tracking-wide text-slate-600 flex items-center gap-2 break-words">
												{#if card.icon}
													<svelte:component this={card.icon} class="h-4 w-4 text-slate-500" />
												{/if}
												{card.title}
											</CardDescription>
											<CardTitle
												class="font-bold text-slate-900 leading-tight min-w-0 overflow-hidden break-words [overflow-wrap:anywhere]
													   text-[clamp(1.25rem,2.8vw,1.875rem)]"
											>
												{card.value ?? '—'}
											</CardTitle>
										</CardHeader>
										{#if card.hint}
											<CardContent class="text-sm text-slate-600 break-words max-w-full">{card.hint}</CardContent>
										{/if}
									</Card>
								{/each}
							</div>
						{:else if section.type === 'description'}
							<Card class="border border-slate-100 shadow-none hover:shadow-md transition-all rounded-2xl break-words">
								<CardHeader>
									<CardTitle class="text-base font-semibold text-slate-900 break-words">{section.title ?? 'Overview'}</CardTitle>
									{#if section.description}
										<CardDescription class="break-words">{section.description}</CardDescription>
									{/if}
								</CardHeader>
								<CardContent>
									<DescriptionList items={section.items} columns={section.columns ?? 2} />
								</CardContent>
							</Card>
						{:else if section.type === 'table'}
							<Card class="border border-slate-100 shadow-none rounded-2xl hover:shadow-md transition-all break-words">
								<CardHeader>
									<CardTitle class="text-base font-semibold text-slate-900 break-words">{section.title ?? 'Records'}</CardTitle>
									{#if section.description}
										<CardDescription class="break-words">{section.description}</CardDescription>
									{/if}
								</CardHeader>
								<CardContent class="p-0 overflow-x-auto">
									<Table class="min-w-full">
										<TableHeader>
											<TableRow>
												{#each section.columns as column (column.key)}
													<TableHead class={cn('px-4 py-3 text-xs uppercase tracking-wide text-slate-600 break-words', alignClass(column.align))}>
														{column.label}
													</TableHead>
												{/each}
											</TableRow>
										</TableHeader>
										<TableBody>
											{#each section.rows as row, rowIndex (rowIndex)}
												<TableRow>
													{#each section.columns as column (column.key)}
														<TableCell
															class={cn(
																'px-4 py-3 text-sm text-slate-700 break-all max-w-[1px] overflow-hidden text-ellipsis',
																alignClass(column.align)
															)}
														>
															{resolveCellValue(row, column)}
														</TableCell>
													{/each}
												</TableRow>
											{/each}
										</TableBody>
									</Table>
								</CardContent>
							</Card>
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
