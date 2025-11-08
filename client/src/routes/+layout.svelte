<script lang="ts">
	import type { Snippet } from 'svelte';
	import '../app.css';
	import { ModeWatcher } from 'mode-watcher';
	import { Toaster } from '$lib/components/ui/sonner';
	import Sidebar from '$lib/components/layout/Sidebar.svelte';
	import Topbar from '$lib/components/layout/Topbar.svelte';
	import { page } from '$app/stores';
	import { auth } from '$lib/stores/auth';

	const { data, children } = $props<{
		data: { user: { Username?: string; Role?: string } | null };
		children?: Snippet;
	}>();

	const publicRoutes = ['/login', '/register'];
	let sidebarOpen = $state(false);
	const isPublicRoute = $derived(publicRoutes.includes($page.url.pathname));
	const authState = $derived($auth);
</script>

<ModeWatcher />
{#if isPublicRoute}
	{@render children?.()}
{:else}
	<div class="flex min-h-screen bg-muted/30 text-foreground">
		<Sidebar bind:open={sidebarOpen} user={authState.user ?? data?.user} />
		{#if sidebarOpen}
			<div class="fixed inset-0 z-30 bg-background/60 backdrop-blur-sm lg:hidden" onclick={() => (sidebarOpen = false)}></div>
		{/if}
		<div class="flex flex-1 flex-col">
			<Topbar user={authState.user ?? data?.user} on:toggleSidebar={() => (sidebarOpen = !sidebarOpen)} />
			<main class="flex-1 px-4 pb-8 pt-4 sm:px-8">
				{@render children?.()}
			</main>
		</div>
	</div>
{/if}
<Toaster />
