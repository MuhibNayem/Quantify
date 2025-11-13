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

<section class="relative w-full isolate overflow-hidden">
	<!-- Gradient background with motion -->
	<div
		class="absolute inset-0 -z-10 animate-gradientShift bg-gradient-to-r from-teal-50 via-cyan-50 to-emerald-100 bg-[length:200%_200%]"
	></div>

	<!-- Floating glow blobs -->
	<div
		class="absolute -top-32 -left-24 w-96 h-96 rounded-full bg-teal-200/40 blur-3xl animate-pulseGlow"
	></div>
	<div
		class="absolute -bottom-28 -right-24 w-80 h-80 rounded-full bg-cyan-200/30 blur-3xl animate-pulseGlow delay-700"
	></div>

	<!-- Hero container -->
	<div
		class="parallax-hero relative mx-auto max-w-7xl px-4 sm:px-6 lg:px-8 pt-16 sm:pt-20 pb-10 sm:pb-16 text-center"
	>
		<div class="inline-flex items-center gap-3 justify-center mb-3">
			<span
				class="inline-flex p-2 rounded-2xl shadow-md bg-gradient-to-br from-teal-500 to-cyan-600 animate-cardFloat"
			>
				<Clock class="h-6 w-6 text-white" />
			</span>
			<p class="text-xs sm:text-sm uppercase tracking-[0.18em] text-teal-700 font-semibold">
				Time Management
			</p>
		</div>

		<h1
			class="text-3xl sm:text-4xl lg:text-5xl font-bold bg-gradient-to-r from-teal-700 via-cyan-700 to-emerald-700 bg-clip-text text-transparent mb-3"
		>
			Time Tracking Dashboard
		</h1>
		<p class="text-slate-600 text-sm sm:text-base max-w-2xl mx-auto">
			Monitor and manage employee work hours with ease and precision.
		</p>
	</div>
</section>

<div class="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8 mt-6">
	<div class="flex justify-end mb-8">
		<div class="flex items-center gap-2 p-1 bg-white/60 backdrop-blur rounded-xl">
			<Button
				variant="{currentRole === 'Staff' ? 'default' : 'ghost'}"
				onclick="{() => (currentRole = 'Staff')}"
				class="rounded-lg {currentRole === 'Staff' ? 'bg-teal-600 text-white' : ''}"
			>
				Staff View
			</Button>
			<Button
				variant="{currentRole === 'Manager' ? 'default' : 'ghost'}"
				onclick="{() => (currentRole = 'Manager')}"
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
		transition-property: color, background-color, border-color, text-decoration-color, fill, stroke,
			opacity, box-shadow, transform, filter, backdrop-filter;
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
		transition: transform 0.1s ease-out, filter 0.2s ease-out;
	}
</style>

