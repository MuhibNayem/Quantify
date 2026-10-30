<!-- client/src/lib/components/layout/Topbar.svelte
     Muse topbar: no hairline, no border. Slight bg shift and
     a soft shadow on the bottom edge separate it from content. -->
<script lang="ts">
	import { Input } from '$lib/components/ui/input';
	import {
		Search,
		Loader2,
		Package,
		Users,
		Truck,
		FolderTree,
		Sun,
		Moon,
		Rows2,
		Rows3,
		RefreshCcw
	} from 'lucide-svelte';
	import { toggleMode } from 'mode-watcher';
	import { goto } from '$app/navigation';
	import { searchApi } from '$lib/api/resources';
	import NotificationBell from '$lib/components/notifications/NotificationBell.svelte';
	import { density } from '$lib/stores/density';

	const { user = null } = $props<{ user?: { Username?: string } | null }>();

	let search = $state('');
	let searchFocused = $state(false);
	let searchResults = $state<any[]>([]);
	let searching = $state(false);
	let hasSearched = $state(false);

	let debounceTimer: ReturnType<typeof setTimeout>;

	const typeConfig: Record<
		string,
		{ label: string; icon: typeof Package; route: (entity: any) => string }
	> = {
		product: {
			label: 'Products',
			icon: Package,
			// SKU is the human-facing product identifier; numeric ID is only a
			// fallback for legacy records.
			route: (entity) =>
				`/products/${encodeURIComponent(entity.SKU || String(entity.ID ?? entity.id))}`
		},
		user: { label: 'Users', icon: Users, route: () => `/users` },
		supplier: {
			label: 'Suppliers',
			icon: Truck,
			route: (entity) => (entity.ID ? `/suppliers/${entity.ID}` : '/suppliers')
		},
		category: { label: 'Categories', icon: FolderTree, route: () => '/catalog' }
	};

	const normalizeHit = (hit: any) => hit.entity ?? hit;

	const groupedResults = $derived.by(() => {
		const groups: Record<string, any[]> = {};
		for (const hit of searchResults) {
			const type = (hit.entity_type || hit.type || 'other').toLowerCase();
			if (!groups[type]) groups[type] = [];
			groups[type].push(hit);
		}
		return groups;
	});

	const runGlobalSearch = async () => {
		const q = search.trim();
		if (!q) {
			searchResults = [];
			hasSearched = false;
			return;
		}
		searching = true;
		hasSearched = true;
		try {
			const data = await searchApi.query(q);
			searchResults = Array.isArray(data) ? data : data.results || data.hits || [];
		} catch {
			searchResults = [];
		} finally {
			searching = false;
		}
	};

	const onInput = (event: Event) => {
		search = (event.currentTarget as HTMLInputElement).value;
		clearTimeout(debounceTimer);
		debounceTimer = setTimeout(runGlobalSearch, 300);
	};

	const navigateToResult = (hit: any) => {
		const type = (hit.entity_type || hit.type || '').toLowerCase();
		const config = typeConfig[type];
		const entity = normalizeHit(hit);
		if (config) {
			goto(config.route(entity));
		}
		searchResults = [];
		searchFocused = false;
		search = '';
	};

	const getDisplayName = (hit: any) => {
		const entity = normalizeHit(hit);
		return (
			entity.Name ||
			entity.Username ||
			entity.name ||
			entity.title ||
			hit.content ||
			hit.matchedContent ||
			`#${entity.ID ?? hit.entity_id ?? hit.id}`
		);
	};
</script>

<header
	class="glass-chrome sticky top-0 z-40 flex h-16 items-center justify-between gap-4 border-b px-4 sm:px-6"
>
	<!-- Search -->
	<div class="relative hidden w-full max-w-sm sm:block">
		<Search
			class="pointer-events-none absolute top-1/2 left-3.5 h-3.5 w-3.5 -translate-y-1/2 text-muted-foreground"
		/>
		{#if searching}
			<Loader2
				class="absolute top-1/2 left-3.5 h-3.5 w-3.5 -translate-y-1/2 animate-spin text-primary"
			/>
		{/if}
		<Input
			data-search-input
			class="glass-input h-9 w-full rounded-lg pr-12 pl-9 text-sm"
			placeholder="Search products, suppliers, users…"
			value={search}
			onfocus={() => (searchFocused = true)}
			onblur={() => setTimeout(() => (searchFocused = false), 150)}
			oninput={onInput}
			onkeydown={(event) => event.key === 'Enter' && runGlobalSearch()}
		/>
		<kbd
			class="pointer-events-none absolute top-1/2 right-3 hidden -translate-y-1/2 rounded-full bg-muted/60 px-2 py-0.5 font-mono text-[10px] text-muted-foreground sm:block"
		>
			⌘K
		</kbd>

		{#if searchFocused && hasSearched && !searching}
			<div class="glass-float absolute top-12 left-0 z-50 w-full overflow-hidden rounded-xl">
				{#if searchResults.length === 0}
					<div class="px-4 py-6 text-center text-sm text-muted-foreground">
						No results for “{search}”
					</div>
				{:else}
					<div class="custom-scrollbar max-h-80 overflow-y-auto p-1.5">
						{#each Object.entries(groupedResults) as [type, hits]}
							{@const config = typeConfig[type]}
							<div class="mb-1">
								<div class="flex items-center gap-1.5 px-3 py-1.5 overline">
									{#if config}
										{@const Icon = config.icon}
										<Icon class="h-3 w-3" />
									{/if}
									{config?.label || type}
								</div>
								{#each hits as hit}
									<button
										class="flex w-full items-center gap-2.5 rounded-xl px-3 py-2 text-left text-sm text-foreground transition-colors hover:bg-muted"
										onmousedown={() => navigateToResult(hit)}
									>
										<span class="truncate font-medium">{getDisplayName(hit)}</span>
										{#if hit.matchedContent}
											<span class="ml-auto truncate text-xs text-muted-foreground"
												>{hit.matchedContent}</span
											>
										{/if}
									</button>
								{/each}
							</div>
						{/each}
					</div>
				{/if}
			</div>
		{/if}
	</div>

	<!-- Actions -->
	<div class="ml-auto flex items-center gap-0.5">
		<button
			class="hidden h-9 w-9 items-center justify-center rounded-xl text-muted-foreground transition-colors hover:bg-muted hover:text-foreground sm:flex"
			onclick={() => window.location.reload()}
			aria-label="Refresh data"
			title="Refresh"
		>
			<RefreshCcw class="h-3.5 w-3.5" />
		</button>

		<button
			class="hidden h-9 w-9 items-center justify-center rounded-xl text-muted-foreground transition-colors hover:bg-muted hover:text-foreground sm:flex"
			onclick={() => density.toggle()}
			aria-label="Toggle table density"
			title="Toggle density"
		>
			{#if $density === 'comfortable'}
				<Rows3 class="h-3.5 w-3.5" />
			{:else}
				<Rows2 class="h-3.5 w-3.5" />
			{/if}
		</button>

		<button
			class="flex h-9 w-9 items-center justify-center rounded-xl text-muted-foreground transition-colors hover:bg-muted hover:text-foreground"
			onclick={toggleMode}
			aria-label="Toggle dark mode"
			title="Toggle theme"
		>
			<Sun class="h-3.5 w-3.5 dark:hidden" />
			<Moon class="hidden h-3.5 w-3.5 dark:block" />
		</button>

		<NotificationBell />

		<!-- User -->
		<div
			class="ml-2 flex items-center gap-2.5 rounded-full border border-border bg-muted/40 py-1 pr-3 pl-1"
		>
			<div
				class="flex h-8 w-8 items-center justify-center rounded-full text-xs font-semibold text-white shadow-sm"
				style="background: var(--gradient-clay);"
			>
				{user?.Username?.slice(0, 2)?.toUpperCase() ?? '??'}
			</div>
			<div class="hidden leading-tight md:block">
				<p class="text-sm font-medium tracking-[-0.005em] text-foreground">
					{user?.Username ?? 'Guest'}
				</p>
				<p class="flex items-center gap-1 text-[11px] text-muted-foreground">
					<span class="live-dot" style="width:5px;height:5px"></span> Online
				</p>
			</div>
		</div>
	</div>
</header>
