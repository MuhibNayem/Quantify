<script lang="ts">
	import { createEventDispatcher } from 'svelte';
	import { Input } from '$lib/components/ui/input';
	import { Button } from '$lib/components/ui/button';
	import { Menu, RefreshCcw, Search, Loader2, Package, Users, Truck, FolderTree, Sun, Moon } from 'lucide-svelte';
	import { toggleMode } from 'mode-watcher';
	import { toast } from 'svelte-sonner';
	import { goto } from '$app/navigation';
	import { searchApi } from '$lib/api/resources';
	import NotificationBell from '$lib/components/notifications/NotificationBell.svelte';

	const { user = null } = $props<{ user?: { Username?: string } | null }>();
	const dispatch = createEventDispatcher<{ toggleSidebar: void }>();

	let search = $state('');
	let searchFocused = $state(false);
	let searchResults = $state<any[]>([]);
	let searching = $state(false);
	let hasSearched = $state(false);

	let debounceTimer: ReturnType<typeof setTimeout>;

	const typeConfig: Record<string, { label: string; icon: typeof Package; route: (id: any) => string }> = {
		product: { label: 'Products', icon: Package, route: (id) => `/products/${id}` },
		user: { label: 'Users', icon: Users, route: (id) => `/users` },
		supplier: { label: 'Suppliers', icon: Truck, route: (id) => `/suppliers` },
		category: { label: 'Categories', icon: FolderTree, route: (id) => `/catalog` },
	};

	const groupedResults = $derived.by(() => {
		const groups: Record<string, any[]> = {};
		for (const hit of searchResults) {
			const type = (hit.type || hit.entityType || 'other').toLowerCase();
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
		const type = (hit.type || hit.entityType || '').toLowerCase();
		const config = typeConfig[type];
		const id = hit.id || hit.ID || hit.entityId;
		if (config && id) {
			goto(config.route(id));
		}
		searchResults = [];
		searchFocused = false;
		search = '';
	};

	const getDisplayName = (hit: any) => hit.name || hit.Name || hit.title || hit.matchedContent || `#${hit.id || hit.ID || hit.entityId}`;
</script>

<header
	class="sticky top-0 z-40 flex h-16 items-center justify-between border-b border-white/40 bg-gradient-to-r from-white/80 via-indigo-50/70 to-sky-50/80 px-3 backdrop-blur-2xl transition-all duration-500 dark:border-white/10 dark:from-slate-900/80 dark:via-slate-900/70 dark:to-slate-800/80 sm:px-6"
>
	<!-- Left section -->
	<div class="lg:ml-16">
		<!-- Search bar -->
		<div class="group relative hidden sm:block">
			<Input
				class="liquid-input peer h-10 w-64 pl-11 pr-4 text-sm text-slate-800 placeholder:text-transparent transition-all duration-300 focus:w-80 focus-visible:ring-0 focus-visible:ring-transparent focus-visible:border-transparent focus-visible:outline-none focus-visible:shadow-[0_30px_70px_-30px_rgba(59,130,246,0.45)]"
				placeholder="Search inventory"
				value={search}
				onfocus={() => (searchFocused = true)}
				onblur={() => setTimeout(() => (searchFocused = false), 150)}
				oninput={onInput}
				onkeydown={(event) => event.key === 'Enter' && runGlobalSearch()}
			/>
			<div class="pointer-events-none absolute left-4 top-1/2 flex -translate-y-1/2 items-center gap-2">
				{#if searching}
					<Loader2 class="h-4 w-4 animate-spin text-emerald-500" />
				{:else}
					<Search
						class="h-4 w-4 text-slate-500 drop-shadow-[0_2px_8px_rgba(255,255,255,0.65)] transition-colors group-focus-within:text-emerald-500"
					/>
				{/if}
				{#if !search && !searchFocused}
					<span class="text-sm text-slate-500">Search inventory</span>
				{/if}
			</div>

			{#if searchFocused && hasSearched && !searching}
				<div class="absolute left-0 top-12 z-50 w-80 overflow-hidden rounded-2xl border border-white/60 bg-white/95 shadow-2xl shadow-slate-900/10 backdrop-blur-xl">
					{#if searchResults.length === 0}
						<div class="px-4 py-6 text-center text-sm text-slate-500">
							No results found for "{search}"
						</div>
					{:else}
						<div class="max-h-80 overflow-y-auto p-2">
							{#each Object.entries(groupedResults) as [type, hits]}
								{@const config = typeConfig[type]}
								<div class="mb-1">
									<div class="flex items-center gap-2 px-3 py-1.5 text-[11px] font-bold uppercase tracking-wider text-slate-400">
									{#if config}
										{@const Icon = config.icon}
										<Icon class="h-3 w-3" />
									{/if}
										{config?.label || type}
									</div>
									{#each hits as hit}
										<button
											class="flex w-full items-center gap-3 rounded-xl px-3 py-2 text-left text-sm text-slate-700 transition-colors hover:bg-blue-50 hover:text-blue-700"
											onmousedown={() => navigateToResult(hit)}
										>
											<span class="truncate font-medium">{getDisplayName(hit)}</span>
											{#if hit.matchedContent}
												<span class="ml-auto truncate text-xs text-slate-400">{hit.matchedContent}</span>
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
	</div>

	<!-- Right section -->
	<div class="flex items-center gap-1">
		<Button
			variant="ghost"
			size="icon"
			class="hidden h-9 w-9 rounded-2xl text-slate-500 hover:bg-white/70 hover:text-slate-800 dark:text-slate-400 dark:hover:bg-slate-700/70 dark:hover:text-slate-200 sm:flex"
			onclick={() => window.location.reload()}
			aria-label="Refresh data"
		>
			<RefreshCcw class="h-4 w-4" />
		</Button>

		<Button
			variant="ghost"
			size="icon"
			class="h-9 w-9 rounded-2xl text-slate-500 hover:bg-white/70 hover:text-slate-800 dark:text-slate-400 dark:hover:bg-slate-700/70 dark:hover:text-slate-200"
			onclick={toggleMode}
			aria-label="Toggle dark mode"
		>
			<Sun class="h-4 w-4 dark:hidden" />
			<Moon class="hidden h-4 w-4 dark:block" />
		</Button>

		<NotificationBell />

		<div class="ml-2 h-6 w-px bg-white/60"></div>

		<!-- User badge -->
		<div class="ml-2 flex cursor-pointer items-center gap-3 rounded-full border border-white/40 bg-white/55 py-1 pl-1 pr-3 shadow-[0_8px_30px_-20px_rgba(15,23,42,0.8)] transition hover:bg-white/75">
			<div class="flex h-9 w-9 items-center justify-center rounded-full bg-gradient-to-br from-blue-500 via-purple-500 to-pink-500 text-xs font-medium text-white shadow-[0_8px_18px_-8px_rgba(147,51,234,0.65)] ring-2 ring-white">
				{user?.Username?.slice(0, 2)?.toUpperCase() ?? '??'}
			</div>
			<div class="hidden text-left sm:block">
				<p class="text-xs font-medium leading-none text-slate-700">
					{user?.Username ?? 'Guest'}
				</p>
				<p class="mt-0.5 text-[10px] text-slate-500">
					{user ? 'Online' : 'Offline'}
				</p>
			</div>
		</div>
	</div>
</header>

<style>
	/* Optional subtle glow motion for a soft ambient effect */
	@keyframes ambientGlow {
		0%,
		100% {
			box-shadow: 0 0 12px rgba(147, 197, 253, 0.15);
		}
		50% {
			box-shadow: 0 0 24px rgba(147, 197, 253, 0.35);
		}
	}
	.animate-ambientGlow {
		animation: ambientGlow 6s ease-in-out infinite;
	}
</style>
