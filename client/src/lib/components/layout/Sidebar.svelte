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
					// If no permission is specified, show the item
					if (!item.permission) return true;
					// Otherwise, check if user has the required permission
					return permissions.includes(item.permission);
				})
			}))
			.filter((section) => section.items.length > 0)
	);
</script>

<!-- Mobile toggle button -->
<Button
	class="fixed left-4 top-4 z-50 rounded-xl bg-gradient-to-r from-blue-600 to-purple-600 text-white shadow-lg transition-transform duration-300 hover:scale-105 lg:hidden"
	onclick={() => (open = !open)}
>
	<Menu class="h-5 w-5" />
</Button>

<!-- Mobile overlay -->
{#if open}
	<div
		class="fixed inset-0 z-40 bg-black/30 backdrop-blur-sm lg:hidden"
		onclick={() => (open = false)}
	/>
{/if}

<!-- Sidebar -->
<aside
	class={cn(
		'fixed inset-y-0 left-0 z-50 flex h-screen w-80 flex-col border-r border-white/10 bg-gradient-to-b from-blue-50/90 via-purple-50/80 to-pink-50/90 shadow-2xl shadow-purple-500/10 backdrop-blur-2xl transition-transform duration-500 ease-out',
		open ? 'translate-x-0' : '-translate-x-full',
		'lg:static lg:z-auto lg:translate-x-0'
	)}
>
	<!-- Header -->
	<div
		class="flex h-20 flex-shrink-0 items-center gap-3 border-b border-white/20 bg-gradient-to-r from-blue-500/10 via-purple-500/10 to-pink-500/10 px-6 backdrop-blur-sm"
	>
		<div
			class="flex h-12 w-12 items-center justify-center rounded-2xl bg-gradient-to-br from-blue-500 to-purple-600 shadow-lg shadow-purple-500/25"
		>
			<Sparkles class="h-6 w-6 text-white" />
		</div>
		<div class="flex flex-col">
			<p
				class="bg-gradient-to-r from-blue-600 to-purple-600 bg-clip-text text-lg font-bold text-transparent"
			>
				Quantify Flow
			</p>
			<p class="text-xs font-medium text-purple-600/70">Inventory intelligence</p>
		</div>
	</div>

	<!-- Navigation -->
	<nav class="flex flex-1 flex-col justify-between overflow-hidden">
		<!-- Scrollable nav items -->
		<div
			class="scrollbar-thin scrollbar-thumb-purple-300 scrollbar-track-transparent flex-1 space-y-6 overflow-y-auto px-5 py-6"
		>
			{#each filteredSections as section}
				<div class="group">
					<p
						class="px-3 text-xs font-semibold uppercase tracking-wider text-purple-600/60 transition-colors duration-300 group-hover:text-purple-700/80"
					>
						{section.title}
					</p>
					<div class="mt-4 space-y-2.5">
						{#each section.items as item}
							<a
								href={item.href}
								class={cn(
									'group relative flex items-start gap-3 rounded-2xl border border-transparent p-3 transition-all duration-300 hover:border-white/30 hover:shadow-lg',
									isActive(item.href, $page.url.pathname)
										? 'border-white/20 bg-gradient-to-r from-blue-500/15 to-purple-500/15 text-blue-700 shadow-lg shadow-blue-500/10'
										: 'text-slate-600 hover:scale-105 hover:bg-white/50 hover:text-slate-800'
								)}
								onclick={() => (open = false)}
							>
								{#if isActive(item.href, $page.url.pathname)}
									<div
										class="absolute -left-2 top-1/2 h-8 w-1 -translate-y-1/2 transform rounded-full bg-gradient-to-b from-blue-500 to-purple-500 shadow-lg shadow-blue-500/50"
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
		<div
			class="m-4 space-y-3 rounded-2xl border border-white/50 bg-gradient-to-br from-white/80 to-white/40 p-5 shadow-xl shadow-purple-500/10 backdrop-blur-sm"
		>
			<div class="flex items-center gap-3">
				<div
					class="flex h-10 w-10 items-center justify-center rounded-full bg-gradient-to-br from-blue-400 to-purple-500 shadow-lg"
				>
					<span class="text-sm font-bold text-white">
						{user?.Username ? user.Username.charAt(0).toUpperCase() : 'U'}
					</span>
				</div>
				<div class="min-w-0 flex-1">
					<p class="truncate text-sm font-semibold text-slate-800">
						{user?.Username ?? 'Pending approval'}
					</p>
					<p class="text-xs font-medium capitalize text-purple-600/70">
						{user?.Role
							? (typeof user.Role === 'string' ? user.Role : user.Role.Name).toLowerCase()
							: 'role pending'}
					</p>
				</div>
			</div>

			<Button
				class="mt-2 w-full border-0 bg-gradient-to-r from-slate-700 to-slate-600 text-white shadow-lg shadow-slate-500/20 transition-all duration-300 hover:scale-105 hover:from-slate-600 hover:to-slate-500 hover:shadow-slate-500/30"
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
		background: rgba(139, 92, 246, 0.3);
		border-radius: 10px;
	}
	.scrollbar-thin::-webkit-scrollbar-thumb:hover {
		background: rgba(139, 92, 246, 0.5);
	}
</style>
