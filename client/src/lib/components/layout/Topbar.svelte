<script lang="ts">
	import { createEventDispatcher, onMount } from 'svelte';
	import { Input } from '$lib/components/ui/input';
	import { Button } from '$lib/components/ui/button';
	import { Menu, RefreshCcw, Search, Sun, Moon } from 'lucide-svelte';
	import { toast } from 'svelte-sonner';
	import NotificationBell from '$lib/components/notifications/NotificationBell.svelte';

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
			description: `Filtering resources for "${search}"`
		});
	};
</script>

<header
	class="sticky top-0 z-40 flex h-16 items-center justify-between border-b
	border-slate-200/60 bg-white/80 backdrop-blur-xl transition-all duration-300
	dark:border-slate-800/60 dark:bg-slate-950/80"
>
	<!-- Left section -->
	<div class="flex items-center gap-4 px-4">
		<Button
			variant="ghost"
			size="icon"
			class="text-slate-500 hover:bg-slate-100 hover:text-slate-700 lg:hidden dark:text-slate-400 dark:hover:bg-slate-800 dark:hover:text-slate-200"
			onclick={() => dispatch('toggleSidebar')}
		>
			<Menu class="h-5 w-5" />
		</Button>

		<!-- Search bar -->
		<div class="group relative hidden sm:block">
			<Search
				class="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-slate-400 transition-colors group-focus-within:text-sky-500"
			/>
			<Input
				class="h-9 w-64 rounded-full border-slate-200 bg-slate-50 pl-9 pr-4 text-sm
					text-slate-600 transition-all duration-200 placeholder:text-slate-400
					focus:w-80 focus:border-sky-500 focus:ring-1 focus:ring-sky-500
					dark:border-slate-800 dark:bg-slate-900/50 dark:text-slate-300 dark:focus:border-sky-400"
				placeholder="Search..."
				value={search}
				oninput={(event) => (search = event.currentTarget.value)}
				onkeydown={(event) => event.key === 'Enter' && runGlobalSearch()}
			/>
		</div>
	</div>

	<!-- Right section -->
	<div class="flex items-center gap-1 pr-4">
		<Button
			variant="ghost"
			size="icon"
			class="hidden h-9 w-9 text-slate-500 hover:bg-slate-100 hover:text-slate-700 sm:flex dark:text-slate-400 dark:hover:bg-slate-800 dark:hover:text-slate-200"
			onclick={() => window.location.reload()}
		>
			<RefreshCcw class="h-4 w-4" />
		</Button>

		<NotificationBell />

		<Button
			variant="ghost"
			size="icon"
			class="h-9 w-9 text-slate-500 hover:bg-slate-100 hover:text-slate-700 dark:text-slate-400 dark:hover:bg-slate-800 dark:hover:text-slate-200"
			onclick={toggleTheme}
		>
			{#if isDark}
				<Sun class="h-4 w-4" />
			{:else}
				<Moon class="h-4 w-4" />
			{/if}
		</Button>

		<div class="ml-2 h-6 w-px bg-slate-200 dark:bg-slate-800"></div>

		<!-- User badge -->
		<div
			class="ml-2 flex cursor-pointer items-center gap-3 rounded-full border
				border-transparent py-1 pl-1 pr-3 transition-colors hover:bg-slate-100 dark:hover:bg-slate-800"
		>
			<div
				class="flex h-8 w-8 items-center justify-center rounded-full
					bg-gradient-to-br from-indigo-500 to-purple-600 text-xs font-medium text-white shadow-sm ring-2 ring-white dark:ring-slate-950"
			>
				{user?.Username?.slice(0, 2)?.toUpperCase() ?? '??'}
			</div>
			<div class="hidden text-left sm:block">
				<p class="text-xs font-medium leading-none text-slate-700 dark:text-slate-200">
					{user?.Username ?? 'Guest'}
				</p>
				<p class="mt-0.5 text-[10px] text-slate-500 dark:text-slate-400">
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
