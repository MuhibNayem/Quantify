<script lang="ts">
	import { navSections } from '$lib/constants/navigation';
	import { page } from '$app/stores';
	import { cn } from '$lib/utils';
	import { Button } from '$lib/components/ui/button';
	import { auth } from '$lib/stores/auth';
	import { LogOut } from 'lucide-svelte';

	let { user = null, open = $bindable(false) } = $props<{
		user?: { Username?: string; Role?: string } | null;
		open?: boolean;
	}>();

	const handleLogout = () => {
		auth.logout();
		window.location.href = '/login';
	};

	const isActive = (href: string, pathname: string) => {
		if (href === '/') {
			return pathname === '/';
		}
		return pathname === href || pathname.startsWith(`${href}/`);
	};
</script>

<aside
	class={cn(
		'fixed inset-y-0 left-0 z-40 flex w-72 min-h-screen flex-col border-r border-border bg-card/95 backdrop-blur-xl transition-transform duration-300 ease-in-out lg:static lg:min-h-screen lg:translate-x-0',
		open ? 'translate-x-0 shadow-2xl' : '-translate-x-full'
	)}
>
	<div class="flex h-16 flex-shrink-0 items-center gap-3 border-b border-border/70 px-6">
		<div class="flex h-11 w-11 items-center justify-center rounded-2xl bg-primary/10 text-lg font-semibold text-primary">
			QF
		</div>
		<div>
			<p class="text-sm font-semibold text-foreground">Quantify Flow</p>
			<p class="text-xs text-muted-foreground">Inventory intelligence</p>
		</div>
	</div>
	<nav class="flex flex-1 flex-col justify-between px-4 py-6">
		<div class="space-y-6 overflow-y-auto pr-1">
			{#each navSections as section}
				<div>
					<p class="px-2 text-xs font-semibold uppercase tracking-wide text-muted-foreground">{section.title}</p>
					<div class="mt-3 space-y-1.5">
						{#each section.items as item}
							<a
								href={item.href}
								class={cn(
									'group flex items-start gap-3 rounded-2xl border px-3 py-2 transition',
									isActive(item.href, $page.url.pathname)
										? 'border-primary/60 bg-primary/5 text-foreground shadow-sm'
										: 'border-transparent text-muted-foreground hover:border-border hover:bg-muted/60'
								)}
								onclick={() => (open = false)}
							>
								<item.icon class="mt-0.5 h-4 w-4 text-primary" />
								<div class="flex flex-col">
									<span class="text-sm font-medium">{item.label}</span>
									{#if item.description}
										<span class="text-xs text-muted-foreground">{item.description}</span>
									{/if}
								</div>
							</a>
						{/each}
					</div>
				</div>
			{/each}
		</div>
		<div class="rounded-2xl border border-dashed border-border/70 bg-card/80 p-4">
			<p class="text-xs text-muted-foreground">Signed in as</p>
			<p class="text-base font-semibold">{user?.Username ?? 'Pending approval'}</p>
			<p class="text-xs text-muted-foreground capitalize">{user?.Role ?? 'role pending'}</p>
			<Button class="mt-4 w-full" variant="secondary" onclick={handleLogout}>
				<LogOut class="mr-2 h-4 w-4" /> Logout
			</Button>
		</div>
	</nav>
</aside>
