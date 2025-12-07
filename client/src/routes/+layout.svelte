<script lang="ts">
	import type { Snippet } from 'svelte';
	import '../app.css';
	import { Toaster } from '$lib/components/ui/sonner';
	import Sidebar from '$lib/components/layout/Sidebar.svelte';
	import Topbar from '$lib/components/layout/Topbar.svelte';
	import { page } from '$app/stores';
	import { auth } from '$lib/stores/auth';
	import { settings } from '$lib/stores/settings';

	const { data, children } = $props<{
		data: { user: { Username?: string; Role?: string } | null };
		children?: Snippet;
	}>();

	const publicRoutes = ['/login', '/register'];
	let sidebarOpen = $state(false);
	const isPublicRoute = $derived(publicRoutes.includes($page.url.pathname));
	const authState = $derived($auth);

	$effect(() => {
		if (authState.user) {
			settings.load();
		}
	});
</script>

{#if isPublicRoute}
	{@render children?.()}
{:else}
	<!-- Background Glows -->
	<div class="pointer-events-none absolute inset-0 -z-10 overflow-hidden">
		<div
			class="animate-pulseGlow absolute -right-32 -top-40 h-64 w-64 rounded-full bg-gradient-to-r from-sky-300 to-blue-400 opacity-50 blur-3xl sm:-right-48 sm:-top-60 sm:h-96 sm:w-96"
		></div>
		<div
			class="animate-pulseGlow absolute -bottom-40 -left-32 h-64 w-64 rounded-full bg-gradient-to-r from-violet-300 to-purple-400 opacity-40 blur-3xl delay-700 sm:-bottom-60 sm:-left-48 sm:h-96 sm:w-96"
		></div>
		<div
			class="animate-pulseGlow absolute left-1/2 top-1/2 h-72 w-72 -translate-x-1/2 -translate-y-1/2 rounded-full bg-gradient-to-r from-emerald-300 to-green-400 opacity-30 blur-3xl delay-500 sm:h-[28rem] sm:w-[28rem]"
		></div>
	</div>

	<!-- Fixed Height Layout -->
	<div
		class="relative z-0 flex h-screen overflow-hidden bg-gradient-to-br from-sky-100 via-blue-50 to-indigo-100/40 text-slate-800"
	>
		<!-- Sidebar (fixed height, non-scrolling page) -->
		<Sidebar bind:open={sidebarOpen} user={authState.user ?? data?.user} />

		<!-- Mobile overlay -->
		{#if sidebarOpen}
			<div
				class="fixed inset-0 z-30 bg-slate-900/30 backdrop-blur-md lg:hidden"
				onclick={() => (sidebarOpen = false)}
			></div>
		{/if}

		<!-- Main Section (scrolls independently) -->
		<div class="flex h-full min-w-0 flex-1 flex-col overflow-hidden">
			<Topbar
				user={authState.user ?? data?.user}
				on:toggleSidebar={() => (sidebarOpen = !sidebarOpen)}
			/>

			<main
				class="flex-1 overflow-y-auto scroll-smooth px-3 pb-12 pt-4 sm:px-6 sm:pt-6 lg:px-8 xl:px-12"
			>
				{@render children?.()}
			</main>
		</div>
	</div>
{/if}

<Toaster />

<style lang="postcss">
	/* Smooth transitions for all layout elements */
	* {
		transition-property:
			color, background-color, border-color, fill, stroke, opacity, box-shadow, transform, filter,
			backdrop-filter;
		transition-timing-function: cubic-bezier(0.4, 0, 0.2, 1);
		transition-duration: 300ms;
	}

	/* Background glow animation */
	@keyframes pulseGlow {
		0%,
		100% {
			transform: scale(1);
			opacity: 0.45;
			filter: blur(80px);
		}
		50% {
			transform: scale(1.1);
			opacity: 0.6;
			filter: blur(90px);
		}
	}
	.animate-pulseGlow {
		animation: pulseGlow 10s ease-in-out infinite;
	}

	/* Layout improvements */
	html,
	body {
		height: 100%;
		overflow: hidden; /* Prevent body scroll */
	}

	main {
		scrollbar-width: thin;
		scrollbar-color: rgba(139, 92, 246, 0.3) transparent;
	}

	main::-webkit-scrollbar {
		width: 6px;
	}

	main::-webkit-scrollbar-thumb {
		background: rgba(139, 92, 246, 0.3);
		border-radius: 10px;
	}
	main::-webkit-scrollbar-thumb:hover {
		background: rgba(139, 92, 246, 0.5);
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
