<!-- ManagerDashboard.svelte -->
<script lang="ts">
	import { Button } from '$lib/components/ui/button';
	import {
		Clock,
		Users,
		Calendar,
		PieChart,
		Activity,
		Coffee,
		FileText,
		Settings,
		AlertCircle,
		CheckCircle,
		DollarSign,
		Target,
		Download,
		Filter
	} from 'lucide-svelte';
	import { onMount } from 'svelte';
	import { toast } from 'svelte-sonner';
	import { timeTrackingApi } from '$lib/api/resources';
	import { t } from '$lib/i18n';

	// Team data
	let teamMembers: any[] = [];

	// Stats data
	let teamStats = {
		totalHours: '0h 00m',
		activeMembers: 0,
		weeklyTarget: 200,
		progress: 0
	};

	// Project data (Mock)

	// Attendance data (Mock)
	let attendanceData: any[] = [];

	// Recent activities
	let recentActivities: any[] = [];

	// Filter state
	let statusFilter = 'all';

	const handleExportReport = () => {
		toast.success($t('time_tracking.toasts.report_exported'), {
			description: $t('time_tracking.toasts.report_exported_desc')
		});
	};

	const handleSendReminder = (memberName: string) => {
		toast.info($t('time_tracking.toasts.reminder_sent'), {
			description: $t('time_tracking.toasts.reminder_sent_desc', { name: memberName })
		});
	};

	async function loadData() {
		try {
			const statusData = await timeTrackingApi.getTeamStatus();
			teamMembers = statusData;

			const activeCount = statusData.filter(
				(m: any) => m.status === 'active' || m.status === 'break'
			).length;
			teamStats.activeMembers = activeCount;
			// update progress based on activeCount? mock it.
			teamStats.progress = Math.min(100, ((activeCount * 8 * 5) / teamStats.weeklyTarget) * 100);

			const activityData = await timeTrackingApi.getRecentActivities();
			recentActivities = activityData.map((a: any) => ({
				user: a.User?.FirstName ? `${a.User.FirstName} ${a.User.LastName}` : `User #${a.UserID}`,
				action:
					a.Status === 'CLOCKED_IN'
						? 'clocked in'
						: a.Status === 'ON_BREAK'
							? 'started break'
							: 'clocked out',
				time: new Date(a.ClockIn).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
			}));

			const overview = await timeTrackingApi.getTeamOverview();
			if (overview) {
				teamStats.totalHours = overview.totalHours;
				teamStats.progress = overview.progress;
				attendanceData = overview.attendance;
			}
		} catch (e) {
			console.error('Failed to load manager dashboard', e);
			toast.error($t('time_tracking.toasts.load_fail'));
		}
	}

	onMount(() => {
		loadData();
	});
</script>

<div class="space-y-4">
	<!-- Dashboard header -->
	<div
		class="flex flex-col gap-3 md:flex-row md:items-end md:justify-between"
		data-animate="fade-up"
	>
		<div>
			<p class="overline">{$t('time_tracking.manager.header.subtitle')}</p>
			<h2
				class="mt-1 font-display text-xl font-semibold tracking-tight text-foreground sm:text-2xl"
			>
				{$t('time_tracking.manager.header.title')}
			</h2>
			<p class="mt-1 text-sm text-muted-foreground">{$t('time_tracking.manager.header.desc')}</p>
		</div>
		<div class="flex flex-wrap gap-2">
			<Button onclick={handleExportReport} class="btn-soothe">
				<Download size={16} />
				{$t('time_tracking.manager.actions.export')}
			</Button>
			<Button variant="outline">
				<Filter size={16} />
				{$t('time_tracking.manager.actions.filter')}
			</Button>
		</div>
	</div>

	<!-- Team Stats Overview -->
	<div class="grid grid-cols-1 gap-4 md:grid-cols-3">
		<section class="panel p-5" data-animate="fade-up">
			<div class="flex items-center gap-2 text-muted-foreground">
				<Clock size={16} />
				<span class="text-xs font-medium">{$t('time_tracking.manager.stats.total_hours')}</span>
			</div>
			<p class="kpi-num mt-2 text-2xl text-foreground">{teamStats.totalHours}</p>
			<div class="mt-3 h-1.5 w-full overflow-hidden rounded-full bg-muted">
				<div
					class="h-full rounded-full bg-primary transition-all duration-500"
					style="width: {teamStats.progress}%"
				></div>
			</div>
			<p class="num-mono mt-2 text-xs text-muted-foreground">
				{$t('time_tracking.manager.stats.weekly_target_percent', { percent: teamStats.progress })}
			</p>
		</section>

		<section class="panel p-5" data-animate="fade-up" style="animation-delay: 80ms;">
			<div class="flex items-center gap-2 text-muted-foreground">
				<Users size={16} />
				<span class="text-xs font-medium">{$t('time_tracking.manager.stats.active_members')}</span>
			</div>
			<p class="kpi-num mt-2 text-2xl text-foreground">{teamStats.activeMembers}/5</p>
			<p class="mt-2 text-xs text-muted-foreground">{$t('time_tracking.manager.stats.working')}</p>
		</section>

		<section class="panel p-5" data-animate="fade-up" style="animation-delay: 160ms;">
			<div class="flex items-center gap-2 text-muted-foreground">
				<Target size={16} />
				<span class="text-xs font-medium">{$t('time_tracking.manager.stats.weekly_target')}</span>
			</div>
			<p class="kpi-num mt-2 text-2xl text-foreground">{teamStats.weeklyTarget}h</p>
			<p class="mt-2 text-xs text-muted-foreground">
				{$t('time_tracking.manager.stats.team_goal')}
			</p>
		</section>
	</div>

	<!-- Team Members -->
	<section class="panel overflow-hidden" data-animate="fade-up">
		<div class="panel-header">
			<h3 class="section-title flex items-center gap-2">
				<Users size={16} class="text-primary" />
				{$t('time_tracking.manager.team.title')}
			</h3>
		</div>
		<div class="panel-body space-y-1.5">
			{#each teamMembers as member}
				<div
					class="flex items-center justify-between gap-3 rounded-md px-2.5 py-2.5 transition-colors hover:bg-muted/60"
				>
					<div class="flex min-w-0 items-center gap-3">
						<div
							class="flex h-9 w-9 shrink-0 items-center justify-center rounded-md bg-primary/10 text-sm font-semibold text-primary ring-1 ring-primary/15"
						>
							{member.avatar}
						</div>
						<div class="min-w-0">
							<p class="truncate text-sm font-semibold text-foreground">{member.name}</p>
							<p class="truncate text-xs text-muted-foreground">{member.role}</p>
						</div>
					</div>
					<div class="flex shrink-0 items-center gap-3">
						<span
							class={member.status === 'active'
								? 'badge-success'
								: member.status === 'break'
									? 'badge-warning'
									: 'badge-neutral'}
						>
							{#if member.status === 'active'}
								<span class="live-dot" style="width: 0.4rem; height: 0.4rem;"></span>
							{/if}
							{member.status === 'active'
								? $t('time_tracking.manager.team.status.working')
								: member.status === 'break'
									? $t('time_tracking.manager.team.status.on_break')
									: $t('time_tracking.manager.team.status.offline')}
						</span>
						<p class="num-mono w-16 text-right text-sm font-medium text-foreground">
							{member.hours}
						</p>
					</div>
				</div>
			{:else}
				<p class="text-muted-foreground px-2.5 py-6 text-center text-sm">
					No team members to display.
				</p>
			{/each}
		</div>
	</section>

	<!-- Attendance and Recent Activities -->
	<div class="grid grid-cols-1 gap-4 lg:grid-cols-2">
		<!-- Attendance Overview -->
		<section class="panel overflow-hidden" data-animate="fade-up">
			<div class="panel-header">
				<h3 class="section-title flex items-center gap-2">
					<Calendar size={16} class="text-success" />
					{$t('time_tracking.manager.attendance.title')}
				</h3>
			</div>
			<div class="panel-body space-y-5">
				<div class="flex h-32 items-end justify-between gap-2">
					{#each attendanceData as day}
						<div class="flex flex-1 flex-col items-center">
							<div class="flex h-24 items-end justify-center gap-1">
								<div
									class="w-3.5 rounded-t-sm bg-success/85 transition-colors duration-300 hover:bg-success"
									style="height: {(day.present / 10) * 60}px;"
								></div>
								{#if day.absent > 0}
									<div
										class="w-3.5 rounded-t-sm bg-danger/85 transition-colors duration-300 hover:bg-danger"
										style="height: {(day.absent / 10) * 60}px;"
									></div>
								{/if}
								{#if day.late > 0}
									<div
										class="w-3.5 rounded-t-sm bg-warning/85 transition-colors duration-300 hover:bg-warning"
										style="height: {(day.late / 10) * 60}px;"
									></div>
								{/if}
							</div>
							<span class="mt-2 text-[0.65rem] font-medium text-muted-foreground">{day.day}</span>
						</div>
					{/each}
				</div>
				<div class="flex flex-wrap justify-center gap-4">
					<div class="flex items-center gap-1.5">
						<div class="h-2.5 w-2.5 rounded-sm bg-success"></div>
						<span class="text-xs text-muted-foreground"
							>{$t('time_tracking.manager.attendance.present')}</span
						>
					</div>
					<div class="flex items-center gap-1.5">
						<div class="h-2.5 w-2.5 rounded-sm bg-warning"></div>
						<span class="text-xs text-muted-foreground"
							>{$t('time_tracking.manager.attendance.late')}</span
						>
					</div>
					<div class="flex items-center gap-1.5">
						<div class="h-2.5 w-2.5 rounded-sm bg-danger"></div>
						<span class="text-xs text-muted-foreground"
							>{$t('time_tracking.manager.attendance.absent')}</span
						>
					</div>
				</div>
			</div>
		</section>

		<!-- Recent Activities -->
		<section class="panel overflow-hidden" data-animate="fade-up">
			<div class="panel-header">
				<h3 class="section-title flex items-center gap-2">
					<Activity size={16} class="text-info" />
					{$t('time_tracking.manager.recent_activity.title')}
				</h3>
			</div>
			<div class="panel-body space-y-1.5">
				{#each recentActivities as activity}
					<div
						class="flex items-start gap-3 rounded-md px-2 py-2 transition-colors hover:bg-muted/60"
					>
						<div class="mt-0.5 shrink-0">
							{#if activity.action === 'clocked in'}
								<div class="rounded-md bg-success/10 p-1 text-success">
									<CheckCircle size={13} />
								</div>
							{:else if activity.action === 'started break'}
								<div class="rounded-md bg-warning/10 p-1 text-warning">
									<Coffee size={13} />
								</div>
							{:else}
								<div class="rounded-md bg-info/10 p-1 text-info">
									<FileText size={13} />
								</div>
							{/if}
						</div>
						<div class="min-w-0 flex-1">
							<p class="text-sm">
								<span class="font-semibold text-foreground">{activity.user}</span>
								<span class="text-muted-foreground">{activity.action}</span>
							</p>
							<p class="num-mono mt-0.5 text-right text-xs font-medium text-muted-foreground">
								{activity.time}
							</p>
						</div>
					</div>
				{:else}
					<p class="text-muted-foreground px-2 py-6 text-center text-sm">No recent activity.</p>
				{/each}
			</div>
		</section>
	</div>

	<!-- Quick Actions -->
	<section class="panel overflow-hidden" data-animate="fade-up">
		<div class="panel-header">
			<h3 class="section-title flex items-center gap-2">
				<Settings size={16} class="text-muted-foreground" />
				{$t('time_tracking.manager.quick_actions.title')}
			</h3>
		</div>
		<div class="panel-body grid grid-cols-2 gap-3 md:grid-cols-4">
			<Button
				variant="outline"
				class="flex h-24 flex-col items-center justify-center gap-2 rounded-lg text-foreground"
			>
				<PieChart size={22} class="text-info" />
				<span class="text-sm font-medium">{$t('time_tracking.manager.quick_actions.reports')}</span>
			</Button>
			<Button
				variant="outline"
				class="flex h-24 flex-col items-center justify-center gap-2 rounded-lg text-foreground"
			>
				<Calendar size={22} class="text-success" />
				<span class="text-sm font-medium">{$t('time_tracking.manager.quick_actions.schedule')}</span
				>
			</Button>
			<Button
				variant="outline"
				class="flex h-24 flex-col items-center justify-center gap-2 rounded-lg text-foreground"
			>
				<DollarSign size={22} class="text-primary" />
				<span class="text-sm font-medium">{$t('time_tracking.manager.quick_actions.payroll')}</span>
			</Button>
			<Button
				variant="outline"
				class="flex h-24 flex-col items-center justify-center gap-2 rounded-lg text-foreground"
			>
				<AlertCircle size={22} class="text-warning" />
				<span class="text-sm font-medium"
					>{$t('time_tracking.manager.quick_actions.reminders')}</span
				>
			</Button>
		</div>
	</section>
</div>
