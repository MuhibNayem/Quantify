<script lang="ts">
	import StaffDashboard from '$lib/components/time-tracking/StaffDashboard.svelte';
	import ManagerDashboard from '$lib/components/time-tracking/ManagerDashboard.svelte';
	import { Button } from '$lib/components/ui/button';
	import GlassTabs from '$lib/components/ui/glass-tabs.svelte';
	import { Clock } from 'lucide-svelte';
	import { auth } from '$lib/stores/auth';
	import { goto } from '$app/navigation';
	import { toast } from 'svelte-sonner';
	import { fade } from 'svelte/transition';

	$effect(() => {
		if (!auth.hasPermission('users.view')) {
			toast.error('Access Denied', {
				description: 'You do not have permission to access time tracking.'
			});
			goto('/');
		}
	});

	let currentRole: 'Staff' | 'Manager' = $state('Staff');
</script>

<section
	class="relative isolate w-full overflow-hidden rounded-[32px] border border-white/30 bg-gradient-to-br from-slate-50/90 via-sky-50/70 to-indigo-50/80 shadow-[0_40px_120px_-80px_rgba(15,23,42,0.9)]"
>
	<div class="absolute inset-0 -z-10 opacity-80">
		<div class="absolute -left-36 top-10 h-80 w-80 rounded-full bg-sky-200/70 blur-[120px]"></div>
		<div class="absolute right-0 top-0 h-72 w-72 rounded-full bg-emerald-200/60 blur-[110px]"></div>
		<div
			class="absolute -bottom-24 left-1/4 h-64 w-64 rounded-full bg-indigo-200/60 blur-[150px]"
		></div>
	</div>

	<div
		class="relative mx-auto flex max-w-4xl flex-col items-center px-6 py-20 text-center sm:px-8 lg:px-10"
	>
		<div class="mb-6 flex items-center gap-3 text-slate-600">
			<span
				class="glass-button flex h-12 w-12 items-center justify-center rounded-2xl text-sky-600 shadow-[0_12px_35px_-20px_rgba(14,165,233,0.9)]"
			>
				<Clock class="h-5 w-5" />
			</span>
			<p class="text-xs font-semibold uppercase tracking-[0.28em]">Time Intelligence</p>
		</div>
		<h1 class="text-balance text-3xl font-semibold text-slate-900 sm:text-4xl lg:text-5xl">
			Time Tracking Control Center
		</h1>
		<p class="mx-auto mt-6 max-w-3xl text-base text-slate-600">
			Stay on top of shifts, breaks, and approvals with a calm workspace designed to feel invisible.
			Switch between personal and manager views without losing the Apple-inspired polish.
		</p>
	</div>
</section>

<div class="mx-auto mt-12 max-w-7xl px-4 sm:px-6 lg:px-8">
	<GlassTabs
		bind:value={currentRole}
		tabs={[
			{ value: 'Staff', label: 'Staff View' },
			{ value: 'Manager', label: 'Manager View' }
		]}
		class="mx-auto mb-8"
	/>

	{#if currentRole === 'Staff'}
		<div in:fade={{ duration: 300, delay: 150 }} out:fade={{ duration: 150 }}>
			<StaffDashboard />
		</div>
	{:else}
		<div in:fade={{ duration: 300, delay: 150 }} out:fade={{ duration: 150 }}>
			<ManagerDashboard />
		</div>
	{/if}
</div>

<style lang="postcss">
	/* Smooth transitions globally - Removed to prevent jarring tab switches */
	/* * {
		transition-property:
			color, background-color, border-color, text-decoration-color, fill, stroke, opacity,
			box-shadow, transform, filter, backdrop-filter;
		transition-timing-function: cubic-bezier(0.4, 0, 0.2, 1);
		transition-duration: 300ms;
	} */

	@keyframes pulseGlow {
		0%,
		100% {
			transform: scale(1);
			opacity: 0.35;
			filter: blur(70px);
		}
		50% {
			transform: scale(1.08);
			opacity: 0.55;
			filter: blur(90px);
		}
	}
	.animate-pulseGlow {
		animation: pulseGlow 12s ease-in-out infinite;
	}
</style>
