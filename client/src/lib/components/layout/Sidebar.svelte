<script lang="ts">
	import { navSections } from '$lib/constants/navigation';
	import { page } from '$app/stores';
	import { cn } from '$lib/utils';
	import { auth } from '$lib/stores/auth';
	import { LogOut, BarChart3, Menu } from 'lucide-svelte';

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
</script>

<!-- Mobile toggle -->
<button
	class="fixed left-4 top-4 z-50 flex h-9 w-9 items-center justify-center rounded-md border border-border bg-card text-foreground shadow-sm lg:hidden"
	onclick={() => (open = !open)}
	aria-label="Toggle navigation"
>
	<Menu class="h-4 w-4" />
</button>

<!-- Mobile overlay -->
{#if open}
	<button
		class="fixed inset-0 z-40 bg-foreground/20 lg:hidden"
		onclick={() => (open = false)}
		aria-label="Close navigation"
	></button>
{/if}

<!-- Sidebar -->
<aside
	class={cn(
		'fixed inset-y-0 left-0 z-50 flex h-screen w-64 flex-col border-r border-border bg-card transition-transform duration-200 ease-out',
		open ? 'translate-x-0' : '-translate-x-full',
		'lg:static lg:z-auto lg:translate-x-0'
	)}
>
	<!-- Brand -->
	<div class="flex h-16 flex-shrink-0 items-center gap-3 border-b border-border px-5">
		<div class="flex h-8 w-8 items-center justify-center rounded-md bg-primary">
			<BarChart3 class="h-4.5 w-4.5 text-primary-foreground" />
		</div>
		<div class="flex flex-col leading-none">
			<span class="font-display text-[17px] font-semibold tracking-tight text-foreground">
				Quantify
			</span>
			<span class="overline mt-1">Inventory OS</span>
		</div>
	</div>

	<!-- Navigation -->
	<nav class="flex flex-1 flex-col overflow-hidden">
		<div class="custom-scrollbar flex-1 space-y-6 overflow-y-auto px-3 py-5">
			{#each filteredSections as section}
				<div>
					<p class="overline px-3">{section.title}</p>
					<div class="mt-2 space-y-0.5">
						{#each section.items as item}
							{@const active = isActive(item.href, $page.url.pathname)}
							<a
								href={item.href}
								class={cn(
									'group relative flex items-center gap-2.5 rounded-md px-3 py-2 text-sm transition-colors duration-150',
									active
										? 'bg-accent font-semibold text-accent-foreground'
										: 'text-muted-foreground hover:bg-muted hover:text-foreground'
								)}
								onclick={() => (open = false)}
							>
								{#if active}
									<span class="absolute left-0 top-1/2 h-4 w-0.5 -translate-y-1/2 rounded-full bg-primary"></span>
								{/if}
								<item.icon
									class={cn(
										'h-4 w-4 flex-shrink-0',
										active ? 'text-primary' : 'text-muted-foreground group-hover:text-foreground'
									)}
								/>
								<span class="truncate">{item.label}</span>
							</a>
						{/each}
					</div>
				</div>
			{/each}
		</div>

		<!-- User -->
		<div class="border-t border-border p-3">
			<div class="flex items-center gap-2.5 rounded-md px-2 py-2">
				<div
					class="flex h-8 w-8 flex-shrink-0 items-center justify-center rounded-md bg-primary/10 text-sm font-semibold text-primary"
				>
					{user?.Username ? user.Username.charAt(0).toUpperCase() : 'U'}
				</div>
				<div class="min-w-0 flex-1 leading-tight">
					<p class="truncate text-sm font-medium text-foreground">
						{user?.Username ?? 'Pending approval'}
					</p>
					<p class="truncate text-xs capitalize text-muted-foreground">{roleName}</p>
				</div>
				<button
					class="flex h-7 w-7 flex-shrink-0 items-center justify-center rounded-md text-muted-foreground transition-colors hover:bg-muted hover:text-foreground"
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
