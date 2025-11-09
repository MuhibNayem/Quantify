<script lang="ts">
	import { createEventDispatcher, onMount } from 'svelte';
	import { Input } from '$lib/components/ui/input';
	import { Button } from '$lib/components/ui/button';
	import { Bell, Menu, RefreshCcw, Search, Sun, Moon } from 'lucide-svelte';
	import { toast } from 'svelte-sonner';

	const { user = null } = $props<{ user?: { Username?: string } | null }>();
	const dispatch = createEventDispatcher<{ toggleSidebar: void }>();

	let search = $state('');
	let isDark = $state(false);

	onMount(() => {
		isDark = document.documentElement.classList.contains('dark');
	});

	const toggleTheme = () => {
		isDark = !isDark;
		document.documentElement.classList.toggle('dark', isDark);
		localStorage.setItem('theme', isDark ? 'dark' : 'light');
	};

	const runGlobalSearch = () => {
		if (!search.trim()) return;
		toast.message('Quick search', {
			description: `Filtering resources for "${search}"`,
		});
	};
</script>

<header
	class="sticky top-0 z-40 flex h-20 items-center justify-between
	border-b border-sky-200/30
	bg-gradient-to-r from-[#b3d4ff]/70 via-[#d9d6ff]/80 to-[#0f172a]/90
	backdrop-blur-2xl shadow-[0_4px_30px_rgba(15,23,42,0.15)]
	transition-all duration-500
	dark:from-[#0f172a]/95 dark:via-[#1e293b]/95 dark:to-[#020617]/95"
>
	<!-- Left section -->
	<div class="flex items-center gap-2">
		<Button
			variant="ghost"
			size="icon"
			class="lg:hidden hover:bg-sky-200/50 dark:hover:bg-slate-800/50 transition-colors"
			onclick={() => dispatch('toggleSidebar')}
		>
			<Menu class="h-5 w-5 text-sky-700 dark:text-sky-300" />
		</Button>

		<!-- Search bar -->
		<div class="relative hidden sm:block">
			<Search class="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-sky-500/80" />
			<Input
				class="pl-9 pr-24 rounded-2xl border border-sky-200/60 bg-white/80 backdrop-blur-sm
					text-slate-800 placeholder-slate-400
					shadow-inner focus:ring-2 focus:ring-sky-200/60 focus:border-sky-400/70
					transition-all duration-300
					dark:bg-slate-800/80 dark:text-sky-100 dark:border-slate-600/70"
				placeholder="Search catalog, suppliers, alerts..."
				value={search}
				oninput={(event) => (search = event.currentTarget.value)}
				onkeydown={(event) => event.key === 'Enter' && runGlobalSearch()}
			/>
			<Button
				size="sm"
				class="absolute right-1 top-1/2 -translate-y-1/2 px-3 rounded-xl
					bg-gradient-to-r from-sky-500 to-indigo-500 hover:from-sky-600 hover:to-indigo-600
					text-white font-semibold shadow-md hover:shadow-lg transition-all"
				onclick={runGlobalSearch}
			>
				Go
			</Button>
		</div>
	</div>

	<!-- Right section -->
	<div class="flex items-center gap-2">
		<Button
			variant="ghost"
			size="icon"
			class="hidden sm:flex hover:bg-sky-200/40 dark:hover:bg-slate-700/40 transition-colors"
			onclick={() => window.location.reload()}
		>
			<RefreshCcw class="h-4 w-4 text-sky-700 dark:text-sky-300" />
		</Button>

		<Button
			variant="ghost"
			size="icon"
			class="relative hover:bg-sky-200/40 dark:hover:bg-slate-700/40 transition-colors"
		>
			<Bell class="h-4 w-4 text-sky-700 dark:text-sky-300" />
			<div
				class="absolute -top-1 -right-1 h-2 w-2
				bg-gradient-to-r from-pink-500 to-rose-500 rounded-full animate-pulse"
			></div>
		</Button>

		<Button
			variant="ghost"
			size="icon"
			class="hover:bg-sky-200/40 dark:hover:bg-slate-700/40 transition-colors"
			onclick={toggleTheme}
		>
			{#if isDark}
				<Sun class="h-4 w-4 text-amber-500" />
			{:else}
				<Moon class="h-4 w-4 text-indigo-700" />
			{/if}
		</Button>

		<!-- User badge -->
		<div
			class="flex items-center gap-2 rounded-2xl border border-sky-200/50
				bg-gradient-to-br from-white/80 to-sky-50/70
				px-3 py-1 shadow-md hover:shadow-lg transition-all
				dark:from-slate-800/80 dark:to-slate-900/80 dark:border-slate-700/50"
		>
			<div
				class="flex h-9 w-9 items-center justify-center rounded-full
					bg-gradient-to-br from-sky-500 to-indigo-600 text-sm font-semibold text-white shadow-md"
			>
				{user?.Username?.slice(0, 2)?.toUpperCase() ?? '??'}
			</div>
			<div class="hidden text-left text-sm leading-tight sm:block">
				<p class="font-semibold text-slate-800 dark:text-sky-100">
					{user?.Username ?? 'Pending'}
				</p>
				<p class="text-xs text-sky-600/80 dark:text-sky-400/80">
					{user ? 'Online' : 'Awaiting approval'}
				</p>
			</div>
		</div>
	</div>
</header>

<style>
	/* Optional subtle glow motion for a soft ambient effect */
	@keyframes ambientGlow {
		0%, 100% {
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
