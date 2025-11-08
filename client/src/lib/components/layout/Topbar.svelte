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
		if (!search.trim()) {
			return;
		}
		toast.message('Quick search', {
			description: `Filtering resources for "${search}"`,
		});
	};
</script>

<header class="sticky top-0 z-30 flex h-16 items-center justify-between border-b border-border/60 bg-background/80 px-4 backdrop-blur-xl">
	<div class="flex items-center gap-2">
		<Button variant="ghost" size="icon" class="lg:hidden" onclick={() => dispatch('toggleSidebar')}>
			<Menu class="h-5 w-5" />
		</Button>
		<div class="relative hidden sm:block">
			<Search class="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
			<Input
				class="pl-9 pr-24"
				placeholder="Search catalog, suppliers, alerts..."
				value={search}
				oninput={(event) => (search = event.currentTarget.value)}
				onkeydown={(event) => event.key === 'Enter' && runGlobalSearch()}
			/>
			<Button size="sm" class="absolute right-1 top-1/2 -translate-y-1/2 px-3" onclick={runGlobalSearch}>
				Go
			</Button>
		</div>
	</div>
	<div class="flex items-center gap-2">
		<Button variant="ghost" size="icon" class="hidden sm:flex" onclick={() => window.location.reload()}>
			<RefreshCcw class="h-4 w-4" />
		</Button>
		<Button variant="ghost" size="icon">
			<Bell class="h-4 w-4" />
		</Button>
		<Button variant="ghost" size="icon" onclick={toggleTheme}>
			{#if isDark}
				<Sun class="h-4 w-4" />
			{:else}
				<Moon class="h-4 w-4" />
			{/if}
		</Button>
		<div class="flex items-center gap-2 rounded-full border border-border/70 bg-card px-3 py-1">
			<div class="flex h-9 w-9 items-center justify-center rounded-full bg-primary/10 text-sm font-semibold text-primary">
				{user?.Username?.slice(0, 2)?.toUpperCase() ?? '??'}
			</div>
			<div class="hidden text-left text-sm leading-tight sm:block">
				<p class="font-semibold text-foreground">{user?.Username ?? 'Pending'}</p>
				<p class="text-xs text-muted-foreground">{user ? 'Online' : 'Awaiting approval'}</p>
			</div>
		</div>
	</div>
</header>
