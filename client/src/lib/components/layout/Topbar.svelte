<script lang="ts">
	import { createEventDispatcher } from 'svelte';
	import { Input } from '$lib/components/ui/input';
	import { Button } from '$lib/components/ui/button';
	import { Menu, RefreshCcw, Search } from 'lucide-svelte';
	import { toast } from 'svelte-sonner';
	import NotificationBell from '$lib/components/notifications/NotificationBell.svelte';

	const { user = null } = $props<{ user?: { Username?: string } | null }>();
	const dispatch = createEventDispatcher<{ toggleSidebar: void }>();

	let search = $state('');
	let searchFocused = $state(false);

	const runGlobalSearch = () => {
		if (!search.trim()) return;
		toast.message('Quick search', {
			description: `Filtering resources for "${search}"`
		});
	};
</script>

<header
	class="sticky top-0 z-40 flex h-16 items-center justify-between border-b border-white/40 bg-gradient-to-r from-white/80 via-indigo-50/70 to-sky-50/80 px-3 backdrop-blur-2xl transition-all duration-500 sm:px-6"
>
	<!-- Left section -->
	<div class="flex items-center gap-4">
		<Button
			variant="ghost"
			size="icon"
			class="h-9 w-9 rounded-2xl bg-gradient-to-r from-blue-600 to-purple-600 text-white shadow-lg shadow-purple-500/30 transition hover:scale-105 lg:hidden"
			onclick={() => dispatch('toggleSidebar')}
			aria-label="Open navigation"
		>
			<Menu class="h-5 w-5" />
		</Button>

		<!-- Search bar -->
		<div class="group relative hidden sm:block">
			<Input
				class="liquid-input peer h-10 w-64 pl-11 pr-4 text-sm text-slate-800 placeholder:text-transparent transition-all duration-300 focus:w-80 focus-visible:ring-0 focus-visible:ring-transparent focus-visible:border-transparent focus-visible:outline-none focus-visible:shadow-[0_30px_70px_-30px_rgba(59,130,246,0.45)]"
				placeholder="Search inventory"
				value={search}
				onfocus={() => (searchFocused = true)}
				onblur={() => (searchFocused = false)}
				oninput={(event) => (search = event.currentTarget.value)}
				onkeydown={(event) => event.key === 'Enter' && runGlobalSearch()}
			/>
			<div class="pointer-events-none absolute left-4 top-1/2 flex -translate-y-1/2 items-center gap-2">
				<Search
					class="h-4 w-4 text-slate-500 drop-shadow-[0_2px_8px_rgba(255,255,255,0.65)] transition-colors group-focus-within:text-emerald-500"
				/>
				{#if !search && !searchFocused}
					<span class="text-sm text-slate-500">Search inventory</span>
				{/if}
			</div>
		</div>
	</div>

	<!-- Right section -->
	<div class="flex items-center gap-1">
		<Button
			variant="ghost"
			size="icon"
			class="hidden h-9 w-9 rounded-2xl text-slate-500 hover:bg-white/70 hover:text-slate-800 sm:flex"
			onclick={() => window.location.reload()}
			aria-label="Refresh data"
		>
			<RefreshCcw class="h-4 w-4" />
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
