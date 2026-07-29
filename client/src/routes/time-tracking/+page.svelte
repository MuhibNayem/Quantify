<script lang="ts">
	import StaffDashboard from '$lib/components/time-tracking/StaffDashboard.svelte';
	import ManagerDashboard from '$lib/components/time-tracking/ManagerDashboard.svelte';
	import GlassTabs from '$lib/components/ui/glass-tabs.svelte';
	import { Clock } from 'lucide-svelte';
	import { auth } from '$lib/stores/auth';
	import { goto } from '$app/navigation';
	import { toast } from 'svelte-sonner';
	import { fade } from 'svelte/transition';

	import { t } from '$lib/i18n';

	const tabsClass = 'rounded-lg border-border bg-card shadow-sm';

	$effect(() => {
		if (!auth.hasPermission('users.view')) {
			toast.error($t('time_tracking.toasts.access_denied'), {
				description: $t('time_tracking.toasts.access_denied_desc')
			});
			goto('/');
		}
	});

	let currentRole: 'Staff' | 'Manager' = $state('Staff');
</script>

<section class="w-full">
	<div class="mx-auto max-w-7xl px-4 pt-10 sm:px-6 lg:px-8">
		<div class="space-y-3" data-animate="fade-up">
			<div class="flex items-center gap-3">
				<span class="glass-button inline-flex rounded-md p-2">
					<Clock class="text-primary h-5 w-5" />
				</span>
				<p class="overline">{$t('time_tracking.hero.label')}</p>
			</div>
			<h1 class="font-display text-foreground text-3xl font-bold tracking-tight sm:text-4xl">
				{$t('time_tracking.hero.title')}
			</h1>
			<p class="text-muted-foreground max-w-3xl text-base">
				{$t('time_tracking.hero.subtitle')}
			</p>
		</div>
	</div>
</section>

<div class="mx-auto mt-8 max-w-7xl px-4 pb-12 sm:px-6 lg:px-8">
	<GlassTabs
		bind:value={currentRole}
		tabs={[
			{ value: 'Staff', label: $t('time_tracking.role_toggle.staff') },
			{ value: 'Manager', label: $t('time_tracking.role_toggle.manager') }
		]}
		class="{tabsClass} mx-auto mb-8"
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
