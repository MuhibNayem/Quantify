<script lang="ts">
	import StaffDashboard from '$lib/components/time-tracking/StaffDashboard.svelte';
	import ManagerDashboard from '$lib/components/time-tracking/ManagerDashboard.svelte';
	import { Button } from '$lib/components/ui/button';
	import { Clock } from 'lucide-svelte';
	import { onMount } from 'svelte';

	let currentRole: 'Staff' | 'Manager' = 'Staff';

	onMount(() => {
		let ticking = false;
		const hero = document.querySelector('.parallax-hero') as HTMLElement | null;

		const handleScroll = () => {
			if (!hero) return;
			if (!ticking) {
				window.requestAnimationFrame(() => {
					const scrollY = window.scrollY || 0;
					const translate = Math.min(scrollY * 0.25, 60);
					const blur = Math.min(scrollY * 0.02, 6);
					hero.style.transform = `translateY(${translate}px)`;
					hero.style.filter = `blur(${blur}px)`;
					ticking = false;
				});
				ticking = true;
			}
		};

		window.addEventListener('scroll', handleScroll, { passive: true });

		return () => {
			window.removeEventListener('scroll', handleScroll);
		};
	});
</script>

<section class="relative isolate w-full overflow-hidden">
	<!-- Gradient background with motion -->
	<div
		class="animate-gradientShift absolute inset-0 -z-10 bg-gradient-to-r from-teal-50 via-cyan-50 to-emerald-100 bg-[length:200%_200%]"
	></div>

	<!-- Floating glow blobs -->
	<div
		class="animate-pulseGlow absolute -left-24 -top-32 h-96 w-96 rounded-full bg-teal-200/40 blur-3xl"
	></div>
	<div
		class="animate-pulseGlow absolute -bottom-28 -right-24 h-80 w-80 rounded-full bg-cyan-200/30 blur-3xl delay-700"
	></div>

	<!-- Hero container -->
	<div
		class="parallax-hero relative mx-auto max-w-7xl px-4 pb-10 pt-16 text-center sm:px-6 sm:pb-16 sm:pt-20 lg:px-8"
	>
		<div class="mb-3 inline-flex items-center justify-center gap-3">
			<span
				class="animate-cardFloat inline-flex rounded-2xl bg-gradient-to-br from-teal-500 to-cyan-600 p-2 shadow-md"
			>
				<Clock class="h-6 w-6 text-white" />
			</span>
			<p class="text-xs font-semibold uppercase tracking-[0.18em] text-teal-700 sm:text-sm">
				Time Management
			</p>
		</div>

		<h1
			class="mb-3 bg-gradient-to-r from-teal-700 via-cyan-700 to-emerald-700 bg-clip-text text-3xl font-bold text-transparent sm:text-4xl lg:text-5xl"
		>
			Time Tracking Dashboard
		</h1>
		<p class="mx-auto max-w-2xl text-sm text-slate-600 sm:text-base">
			Monitor and manage employee work hours with ease and precision.
		</p>
	</div>
</section>

<div class="mx-auto mt-6 max-w-7xl px-4 sm:px-6 lg:px-8">
	<div class="mb-8 flex justify-end">
		<div class="flex items-center gap-2 rounded-xl bg-white/60 p-1 backdrop-blur">
			<Button
				variant={currentRole === 'Staff' ? 'default' : 'ghost'}
				onclick={() => (currentRole = 'Staff')}
				class="rounded-lg {currentRole === 'Staff' ? 'bg-teal-600 text-white' : ''}"
			>
				Staff View
			</Button>
			<Button
				variant={currentRole === 'Manager' ? 'default' : 'ghost'}
				onclick={() => (currentRole = 'Manager')}
				class="rounded-lg {currentRole === 'Manager' ? 'bg-cyan-600 text-white' : ''}"
			>
				Manager View
			</Button>
		</div>
	</div>

	{#if currentRole === 'Staff'}
		<StaffDashboard />
	{:else}
		<ManagerDashboard />
	{/if}
</div>

<style lang="postcss">
	/* Smooth transitions globally */
	* {
		transition-property:
			color, background-color, border-color, text-decoration-color, fill, stroke, opacity,
			box-shadow, transform, filter, backdrop-filter;
		transition-timing-function: cubic-bezier(0.4, 0, 0.2, 1);
		transition-duration: 300ms;
	}

	/* Hero gradient animation */
	@keyframes gradientShift {
		0% {
			background-position: 0% 50%;
		}
		50% {
			background-position: 100% 50%;
		}
		100% {
			background-position: 0% 50%;
		}
	}
	.animate-gradientShift {
		background-size: 200% 200%;
		animation: gradientShift 16s ease-in-out infinite;
	}

	/* Soft glowing blobs */
	@keyframes pulseGlow {
		0%,
		100% {
			transform: scale(1);
			opacity: 0.45;
			filter: blur(80px);
		}
		50% {
			transform: scale(1.08);
			opacity: 0.6;
			filter: blur(90px);
		}
	}
	.animate-pulseGlow {
		animation: pulseGlow 10s ease-in-out infinite;
	}

	/* Card float micro-motion */
	@keyframes cardFloat {
		0%,
		100% {
			transform: translateY(0);
		}
		50% {
			transform: translateY(-4px);
		}
	}
	.animate-cardFloat {
		animation: cardFloat 4s ease-in-out infinite;
	}

	.parallax-hero {
		transform: translateY(0);
		will-change: transform, filter;
		transition:
			transform 0.1s ease-out,
			filter 0.2s ease-out;
	}
</style>
