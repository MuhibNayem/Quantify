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

	const tabsClass = 'mx-auto mb-6';

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

<div class="mx-auto w-full max-w-7xl px-4 py-8 sm:px-6 lg:px-8 lg:py-10">
	<!-- Page header -->
	<header class="page-header" data-animate="fade-up">
		<div>
			<p class="flex items-center gap-2 overline">
				<span class="inline-flex rounded-md bg-primary/10 p-1.5 text-primary">
					<Clock class="h-3.5 w-3.5" />
				</span>
				{$t('time_tracking.hero.label')}
			</p>
			<h1 class="page-title mt-1">{$t('time_tracking.hero.title')}</h1>
			<p class="page-sub">{$t('time_tracking.hero.subtitle')}</p>
		</div>
	</header>

	<div class="mt-6">
		<GlassTabs
			bind:value={currentRole}
			tabs={[
				{ value: 'Staff', label: $t('time_tracking.role_toggle.staff') },
				{ value: 'Manager', label: $t('time_tracking.role_toggle.manager') }
			]}
			class={tabsClass}
		/>

		{#if currentRole === 'Staff'}
			<div in:fade={{ duration: 250, delay: 100 }} out:fade={{ duration: 120 }}>
				<StaffDashboard />
			</div>
		{:else}
			<div in:fade={{ duration: 250, delay: 100 }} out:fade={{ duration: 120 }}>
				<ManagerDashboard />
			</div>
		{/if}
	</div>
</div>
