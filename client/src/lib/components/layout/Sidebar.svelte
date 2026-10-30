<!-- client/src/lib/components/layout/Sidebar.svelte
     Enterprise sidebar: porcelain surface, precise border, quiet
     active state with an indigo indicator. -->
<script lang="ts">
	import { navSections } from '$lib/constants/navigation';
	import { page } from '$app/stores';
	import { cn } from '$lib/utils';
	import { auth } from '$lib/stores/auth';
	import { LogOut, Menu, Sparkles } from 'lucide-svelte';

	let { user = null, open = $bindable(false) } = $props<{
		user?: { Username?: string; Role?: string | { Name?: string } } | null;
		open?: boolean;
	}>();

	const handleLogout = async () => {
		await auth.logout();
		window.location.href = '/login';
	};

	const isActive = (href: string, pathname: string) => {
		if (href === '/') return pathname === '/';
		return pathname === href || pathname.startsWith(`${href}/`);
	};

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

	const roleName = $derived(
		user?.Role
			? (typeof user.Role === 'string' ? user.Role : user.Role.Name).toLowerCase()
			: 'role pending'
	);

	const userInitial = $derived(user?.Username ? user.Username.charAt(0).toUpperCase() : 'U');
	const userName = $derived(user?.Username ?? 'Pending approval');
</script>

<!-- Mobile toggle -->
<button
	class="fixed top-4 left-4 z-50 flex h-10 w-10 items-center justify-center rounded-xl border border-border bg-card text-foreground shadow-sm transition-transform hover:scale-105 lg:hidden"
	onclick={() => (open = !open)}
	aria-label="Toggle navigation"
>
	<Menu class="h-4 w-4" />
</button>

<!-- Mobile overlay -->
{#if open}
	<button
		class="fixed inset-0 z-40 bg-slate-900/20 backdrop-blur-sm lg:hidden"
		onclick={() => (open = false)}
		aria-label="Close navigation"
	></button>
{/if}

<!-- Sidebar -->
<aside
	class={cn(
		'glass-chrome fixed inset-y-0 left-0 z-50 flex h-screen w-72 flex-col border-r transition-transform duration-300 ease-out',
		open ? 'translate-x-0' : '-translate-x-full',
		'lg:static lg:z-auto lg:translate-x-0'
	)}
>
	<!-- Brand -->
	<div class="flex h-20 flex-shrink-0 items-center gap-3 border-b border-border/70 px-6">
		<div
			class="flex h-10 w-10 items-center justify-center rounded-xl shadow-sm"
			style="background: var(--gradient-primary); box-shadow: 0 6px 18px -8px hsl(var(--primary) / 0.65);"
		>
			<Sparkles class="h-4 w-4 text-white" />
		</div>
		<div class="flex flex-col leading-none">
			<span class="font-display text-[18px] font-semibold tracking-tight text-foreground">
				Quantify
			</span>
			<span
				class="mt-1.5 text-[0.625rem] font-semibold tracking-[0.16em] text-muted-foreground uppercase"
			>
				Inventory OS
			</span>
		</div>
	</div>

	<!-- Navigation -->
	<nav class="flex flex-1 flex-col overflow-hidden">
		<div class="custom-scrollbar flex-1 space-y-6 overflow-y-auto px-3 py-5">
			{#each filteredSections as section}
				<div>
					<p class="px-3 overline">{section.title}</p>
					<div class="mt-2 space-y-0.5">
						{#each section.items as item}
							{@const active = isActive(item.href, $page.url.pathname)}
							<a
								href={item.href}
								class={cn(
									'group relative flex items-center gap-3 rounded-lg px-3 py-2.5 text-[13.5px] transition-colors duration-150',
									active
										? 'bg-primary/[0.07] font-semibold text-primary'
										: 'text-muted-foreground hover:bg-muted/70 hover:text-foreground'
								)}
								onclick={() => (open = false)}
							>
								{#if active}
									<span
										class="absolute top-1/2 -left-3 h-5 w-[3px] -translate-y-1/2 rounded-r-full bg-primary"
										aria-hidden="true"
									></span>
								{/if}
								<item.icon
									class={cn(
										'h-[16px] w-[16px] flex-shrink-0 transition-colors',
										active ? 'text-primary' : 'text-muted-foreground/80 group-hover:text-foreground'
									)}
								/>
								<span class="truncate tracking-[-0.005em]">{item.label}</span>
							</a>
						{/each}
					</div>
				</div>
			{/each}
		</div>

		<!-- User -->
		<div class="border-t border-border/70 p-3">
			<div
				class="flex items-center gap-2.5 rounded-xl border border-border bg-muted/40 px-2.5 py-2.5"
			>
				<div
					class="flex h-9 w-9 flex-shrink-0 items-center justify-center rounded-full text-sm font-semibold text-white shadow-sm"
					style="background: var(--gradient-primary);"
				>
					{userInitial}
				</div>
				<div class="min-w-0 flex-1 leading-tight">
					<p class="truncate text-sm font-semibold tracking-[-0.005em] text-foreground">
						{userName}
					</p>
					<p class="truncate text-[11px] text-muted-foreground capitalize">{roleName}</p>
				</div>
				<button
					class="flex h-8 w-8 flex-shrink-0 items-center justify-center rounded-lg text-muted-foreground transition-colors hover:bg-card hover:text-foreground"
					onclick={handleLogout}
					aria-label="Sign out"
					title="Sign out"
				>
					<LogOut class="h-3.5 w-3.5" />
				</button>
			</div>
		</div>
	</nav>
</aside>
