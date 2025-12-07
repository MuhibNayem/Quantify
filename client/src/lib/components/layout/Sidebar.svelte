<script lang="ts">
	import { navSections } from '$lib/constants/navigation';
	import { page } from '$app/stores';
	import { cn } from '$lib/utils';
	import { Button } from '$lib/components/ui/button';
	import { auth } from '$lib/stores/auth';
	import { LogOut, Sparkles, Menu } from 'lucide-svelte';

	let { user = null, open = $bindable(false) } = $props<{
		user?: { Username?: string; Role?: string } | null;
		open?: boolean;
	}>();

	const handleLogout = () => {
		auth.logout();
		window.location.href = '/login';
	};

	const isActive = (href: string, pathname: string) => {
		if (href === '/') return pathname === '/';
		return pathname === href || pathname.startsWith(`${href}/`);
	};
	/* Reactive filtered navigation based on permissions */
	const permissions = $derived($auth.permissions || []);

	const filteredSections = $derived(
		navSections
			.map((section) => ({
				...section,
				items: section.items.filter((item) => {
					if (!item.permission) return true;
					return permissions.includes(item.permission);
				})
			}))
			.filter((section) => section.items.length > 0)
	);
</script>

<!-- Mobile toggle button -->
<Button
	class="fixed left-4 top-4 z-50 rounded-2xl bg-gradient-to-r from-blue-600 to-purple-600 text-white shadow-lg shadow-purple-500/30 transition-transform duration-300 hover:scale-105 lg:hidden"
	onclick={() => (open = !open)}
	aria-label="Toggle navigation"
>
	<Menu class="h-5 w-5" />
</Button>

<!-- Mobile overlay -->
{#if open}
	<div
		class="fixed inset-0 z-40 bg-slate-900/20 backdrop-blur-lg transition-all lg:hidden"
		onclick={() => (open = false)}
	/>
{/if}

<!-- Sidebar -->
<aside
	class={cn(
		'liquid-panel fixed inset-y-0 left-0 z-50 flex h-screen w-80 flex-col overflow-hidden border border-white/40 bg-gradient-to-b from-sky-50/90 via-blue-50/85 to-purple-50/90 px-0 text-slate-800 shadow-[0_25px_80px_-40px_rgba(59,7,100,0.45)] transition-transform duration-500 ease-out',
		open ? 'translate-x-0' : '-translate-x-full',
		'lg:static lg:z-auto lg:translate-x-0'
	)}
>
	<!-- Header -->
	<div class="flex h-20 flex-shrink-0 items-center gap-3 border-b border-white/30 bg-gradient-to-r from-blue-500/5 via-purple-500/10 to-pink-500/5 px-6 backdrop-blur-sm">
		<div class="relative flex h-12 w-12 items-center justify-center rounded-[1.4rem] bg-gradient-to-br from-blue-500 to-purple-600 shadow-lg shadow-purple-500/30">
			<Sparkles class="h-6 w-6 text-white drop-shadow-[0_0_15px_rgba(255,255,255,0.65)]" />
			<div class="pointer-events-none absolute inset-0 rounded-[1.4rem] border border-white/30"></div>
		</div>
		<div class="flex flex-col">
			<p class="bg-gradient-to-r from-blue-600 to-purple-600 bg-clip-text text-lg font-semibold tracking-tight text-transparent">
				Quantify Flow
			</p>
			<p class="text-xs font-medium text-purple-600/80">Inventory intelligence</p>
		</div>
	</div>

	<!-- Navigation -->
	<nav class="flex flex-1 flex-col justify-between overflow-hidden">
		<!-- Scrollable nav items -->
		<div class="scrollbar-thin scrollbar-track-transparent flex-1 space-y-6 overflow-y-auto px-5 py-6">
			{#each filteredSections as section}
				<div class="group">
					<p
						class="px-3 text-xs font-semibold uppercase tracking-wider text-purple-600/70 transition-colors duration-300 group-hover:text-purple-700"
					>
						{section.title}
					</p>
					<div class="mt-4 space-y-2.5">
						{#each section.items as item}
							<a
								href={item.href}
								class={cn(
									'liquid-hoverable group relative flex items-start gap-3 rounded-[1.4rem] border border-transparent px-4 py-3 transition-all duration-500 ease-[cubic-bezier(0.4,0,0.2,1)]',
									isActive(item.href, $page.url.pathname)
										? 'border-white/30 bg-gradient-to-r from-blue-500/15 to-purple-500/15 text-blue-700 shadow-lg shadow-blue-500/10'
										: 'text-slate-600 hover:bg-white/55 hover:text-slate-900 hover:shadow-[0_12px_45px_-35px_rgba(15,23,42,0.6)]'
								)}
								onclick={() => (open = false)}
							>
								{#if isActive(item.href, $page.url.pathname)}
									<div
										class="absolute -left-2 top-1/2 h-9 w-1.5 -translate-y-1/2 rounded-full bg-gradient-to-b from-blue-500 to-purple-500 shadow-lg shadow-blue-500/40"
									></div>
								{/if}

								<item.icon
									class={cn(
										'mt-0.5 h-5 w-5 transition-all duration-300',
										isActive(item.href, $page.url.pathname)
											? 'text-blue-600'
											: 'text-purple-500/70 group-hover:text-purple-600'
									)}
								/>
								<div class="flex flex-col">
									<span class="text-sm font-semibold transition-colors duration-300"
										>{item.label}</span
									>
									{#if item.description}
										<span
											class={cn(
												'text-xs transition-colors duration-300',
												isActive(item.href, $page.url.pathname)
													? 'text-blue-600/80'
													: 'text-slate-500 group-hover:text-slate-600'
											)}>{item.description}</span
										>
									{/if}
								</div>
							</a>
						{/each}
					</div>
				</div>
			{/each}
		</div>

		<!-- User Card -->
		<div class="liquid-panel m-4 space-y-3 rounded-[1.6rem] border-transparent bg-gradient-to-br from-white/90 to-white/60 p-5 shadow-xl shadow-purple-500/10">
			<div class="flex items-center gap-3">
				<div class="flex h-10 w-10 items-center justify-center rounded-full bg-gradient-to-br from-blue-400 to-purple-500 text-sm font-bold text-white shadow-lg">
					{user?.Username ? user.Username.charAt(0).toUpperCase() : 'U'}
				</div>
				<div class="min-w-0 flex-1">
					<p class="truncate text-sm font-semibold text-slate-700">
						{user?.Username ?? 'Pending approval'}
					</p>
					<p class="text-xs font-medium capitalize text-purple-600/80">
						{user?.Role
							? (typeof user.Role === 'string' ? user.Role : user.Role.Name).toLowerCase()
							: 'role pending'}
					</p>
				</div>
			</div>

			<Button
				class="mt-2 w-full rounded-2xl border-0 bg-gradient-to-r from-slate-700 to-slate-600 text-white shadow-lg shadow-slate-500/20 transition-all duration-300 hover:scale-105"
				onclick={handleLogout}
			>
				<LogOut class="mr-2 h-4 w-4" />
				Sign Out
			</Button>
		</div>
	</nav>
</aside>

<style>
	/* Custom scroll styling for nav */
	.scrollbar-thin::-webkit-scrollbar {
		width: 4px;
	}
	.scrollbar-thin::-webkit-scrollbar-track {
		background: transparent;
	}
	.scrollbar-thin::-webkit-scrollbar-thumb {
		background: rgba(125, 211, 252, 0.35);
		border-radius: 10px;
	}
	.scrollbar-thin::-webkit-scrollbar-thumb:hover {
		background: rgba(125, 211, 252, 0.55);
	}
</style>
