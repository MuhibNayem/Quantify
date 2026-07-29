<script lang="ts">
	import type { Snippet } from 'svelte';
	import '../app.css';
	import { ModeWatcher } from 'mode-watcher';
	import { Toaster } from '$lib/components/ui/sonner';
	import Sidebar from '$lib/components/layout/Sidebar.svelte';
	import Topbar from '$lib/components/layout/Topbar.svelte';
	import CommandPalette from '$lib/components/command-palette.svelte';
	import Onboarding from '$lib/components/onboarding.svelte';
	import OfflineBanner from '$lib/components/offline-banner.svelte';
	import { page } from '$app/stores';
	import { auth } from '$lib/stores/auth';
	import { settings } from '$lib/stores/settings';

	const { data, children } = $props<{
		data: { user: { Username?: string; Role?: string } | null };
		children?: Snippet;
	}>();

	const publicRoutes = ['/login', '/register'];
	let sidebarOpen = $state(false);
	let paletteOpen = $state(false);
	let onboardingOpen = $state(false);
	let onboardingRef = $state<Onboarding | null>(null);
	const isPublicRoute = $derived(publicRoutes.includes($page.url.pathname));
	const authState = $derived($auth);

	function isTyping(el: EventTarget | null): boolean {
		return el instanceof HTMLElement && ['INPUT', 'TEXTAREA', 'SELECT'].includes(el.tagName);
	}

	function handleGlobalKeydown(e: KeyboardEvent) {
		if ((e.metaKey || e.ctrlKey) && e.key === 'k') {
			e.preventDefault();
			paletteOpen = !paletteOpen;
		}
		if (e.key === '/' && !isTyping(e.target)) {
			e.preventDefault();
			document.querySelector<HTMLElement>('[data-search-input]')?.focus();
		}
		if (e.key === 'Escape') paletteOpen = false;
	}

	$effect(() => {
		if (authState.user) {
			settings.load();
			if (onboardingRef?.shouldShow()) {
				onboardingOpen = true;
			}
		}
	});
</script>

<ModeWatcher />
<OfflineBanner />
<svelte:window onkeydown={handleGlobalKeydown} />
<CommandPalette bind:open={paletteOpen} />
<Onboarding bind:this={onboardingRef} bind:open={onboardingOpen} />

{#if isPublicRoute}
	{@render children?.()}
{:else}
	<a href="#main-content" class="skip-link">Skip to main content</a>

	<!-- Fixed Height Layout -->
	<div class="relative z-0 flex h-screen overflow-hidden bg-background text-foreground">
		<!-- Sidebar (fixed height, non-scrolling page) -->
		<Sidebar bind:open={sidebarOpen} user={authState.user ?? data?.user} />

		<!-- Mobile overlay -->
		{#if sidebarOpen}
			<div
				class="fixed inset-0 z-30 bg-foreground/40 lg:hidden"
				onclick={() => (sidebarOpen = false)}
			></div>
		{/if}

		<!-- Main Section (scrolls independently) -->
		<div class="flex h-full min-w-0 flex-1 flex-col overflow-hidden">
			<Topbar user={authState.user ?? data?.user} />

		<main
			id="main-content"
			class="flex-1 overflow-y-auto scroll-smooth px-3 pb-12 pt-4 sm:px-6 sm:pt-6 lg:px-8 xl:px-12"
		>
				{@render children?.()}
			</main>
		</div>
	</div>
{/if}

<div aria-live="polite" aria-atomic="true">
	<Toaster />
</div>

<style lang="postcss">
	.skip-link {
		position: absolute;
		top: -100%;
		left: 1rem;
		z-index: 9999;
		padding: 0.75rem 1.5rem;
		border-radius: 0 0 var(--radius) var(--radius);
		background: hsl(var(--primary));
		color: hsl(var(--primary-foreground));
		font-weight: 600;
		text-decoration: none;
		transition: top 0.2s ease;
	}

	.skip-link:focus {
		top: 0;
	}

	/* Layout improvements */
	html,
	body {
		height: 100%;
		overflow: hidden; /* Prevent body scroll */
	}

	main {
		scrollbar-width: thin;
		scrollbar-color: hsl(var(--border)) transparent;
	}

	main::-webkit-scrollbar {
		width: 6px;
	}

	main::-webkit-scrollbar-thumb {
		background: hsl(var(--border));
		border-radius: 3px;
	}
	main::-webkit-scrollbar-thumb:hover {
		background: hsl(var(--muted-foreground) / 0.5);
	}

	/* Responsive adjustments */
	@media (max-width: 1024px) {
		main {
			padding-left: 1.5rem !important;
			padding-right: 1.5rem !important;
		}
	}

	@media (max-width: 768px) {
		main {
			padding: 1rem !important;
			padding-bottom: 4rem;
		}
		.flex.h-screen {
			flex-direction: column;
			height: 100vh;
		}
	}

	@media (max-width: 480px) {
		main {
			padding: 0.75rem !important;
		}
	}
</style>
