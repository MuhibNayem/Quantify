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
	import {
		Card,
		CardContent,
		CardDescription,
		CardHeader,
		CardTitle
	} from '$lib/components/ui/card';
	import DescriptionList from '$lib/components/ui/description-list';
	import type { DescriptionItem } from '$lib/components/ui/description-list/description-list.svelte';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import {
		Table,
		TableBody,
		TableCell,
		TableHead,
		TableHeader,
		TableRow
	} from '$lib/components/ui/table';

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
			class="detail-drawer absolute right-0 top-0 flex h-full w-full flex-col rounded-l-lg border-l border-border bg-card shadow-lg sm:w-[600px] md:w-[700px]"
			transition:fly={{ x: 400, duration: 350, delay: 80, easing: cubicOut }}
		>
			<div class="flex items-start justify-between border-b border-border p-6">
				<div class="min-w-0">
					<h2 class="font-display break-words text-2xl font-semibold text-foreground">
						{title}
					</h2>
					{#if subtitle}
						<p class="mt-1 break-words text-sm text-muted-foreground">{subtitle}</p>
					{/if}
				</div>

				<div class="flex items-center gap-2">
					{#if showRefresh && resourceId}
						<Button
							variant="ghost"
							size="icon"
							class="rounded-md p-2 transition-all active:scale-95"
							onclick={fetchData}
							disabled={loading}
						>
							<RefreshCcw class={cn('h-4 w-4 text-muted-foreground', loading && 'animate-spin')} />
						</Button>
					{/if}
					<Button
						variant="ghost"
						size="icon"
						class="rounded-md p-2 transition-all"
						onclick={closeDrawer}
					>
						<X class="h-5 w-5 text-muted-foreground" />
					</Button>
				</div>
			</div>

			<!-- Body -->
			<div class="flex-1 space-y-6 overflow-y-auto p-6">
				{#if loading}
					<Skeleton class="h-8 w-3/4 rounded-md" />
					<Skeleton class="h-4 w-1/2 rounded-md" />
				{:else if error}
					<Card class="break-words border-destructive/25 bg-destructive/5 shadow-none">
						<CardHeader class="flex flex-row items-start gap-3">
							<div class="rounded-md bg-destructive/10 p-2 text-destructive">
								<AlertCircle class="h-5 w-5" />
							</div>
							<div>
								<CardTitle class="break-words text-destructive">Unable to load details</CardTitle>
								<CardDescription class="break-words text-destructive/80">{error}</CardDescription>
							</div>
						</CardHeader>
					</Card>
				{:else if data}
					{#each sections as section, index (index)}
						{#if section.type === 'summary'}
							<div class="grid gap-4 sm:grid-cols-2 md:grid-cols-3">
								{#each section.cards as card (card.title)}
									<Card
										class="min-w-0 break-words rounded-lg border-border bg-card shadow-sm transition-shadow duration-200 hover:shadow-md"
									>
										<CardHeader class="min-w-0 space-y-1">
											<CardDescription
												class="overline flex items-center gap-2 break-words"
											>
												{#if card.icon}
													<svelte:component
														this={card.icon}
														class={cn('h-4 w-4', resolveAccent(card.accent))}
													/>
												{/if}
												{card.title}
											</CardDescription>
											<CardTitle
												class="kpi-num min-w-0 overflow-hidden break-words text-[clamp(1.25rem,2.8vw,1.875rem)] text-foreground
													   [overflow-wrap:anywhere]"
											>
												{card.value ?? '—'}
											</CardTitle>
										</CardHeader>
										{#if card.hint}
											<CardContent class="max-w-full break-words text-sm text-muted-foreground"
												>{card.hint}</CardContent
											>
										{/if}
									</Card>
								{/each}
							</div>
						{:else if section.type === 'description'}
							<Card
								class="break-words rounded-lg border-border bg-card shadow-sm"
							>
								<CardHeader>
									<CardTitle class="break-words text-base font-semibold text-foreground"
										>{section.title ?? 'Overview'}</CardTitle
									>
									{#if section.description}
										<CardDescription class="break-words">{section.description}</CardDescription>
									{/if}
								</CardHeader>
								<CardContent>
									<DescriptionList items={section.items} columns={section.columns ?? 2} />
								</CardContent>
							</Card>
						{:else if section.type === 'table'}
							<Card
								class="break-words rounded-lg border-border bg-card shadow-sm"
							>
								<CardHeader>
									<CardTitle class="break-words text-base font-semibold text-foreground"
										>{section.title ?? 'Records'}</CardTitle
									>
									{#if section.description}
										<CardDescription class="break-words">{section.description}</CardDescription>
									{/if}
								</CardHeader>
								<CardContent class="overflow-x-auto p-0">
									<Table class="min-w-full">
										<TableHeader>
											<TableRow>
												{#each section.columns as column (column.key)}
													<TableHead
														class={cn(
															'break-words px-4 py-3 text-xs uppercase tracking-wide text-muted-foreground',
															alignClass(column.align)
														)}
													>
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
																'max-w-[1px] overflow-hidden text-ellipsis break-all px-4 py-3 text-sm text-foreground',
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
