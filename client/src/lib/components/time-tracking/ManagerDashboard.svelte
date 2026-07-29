<!-- ManagerDashboard.svelte -->
<script lang="ts">
	import { Button } from '$lib/components/ui/button';
	import { Card, CardContent, CardHeader, CardTitle } from '$lib/components/ui/card';
	import {
		Clock,
		Users,
		TrendingUp,
		Calendar,
		BarChart3,
		PieChart,
		Activity,
		MapPin,
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

<div class="space-y-10">
	<!-- Header -->
	<div class="flex flex-col items-start justify-between gap-4 md:flex-row md:items-center">
		<div class="space-y-1" data-animate="fade-up">
			<p class="overline">{$t('time_tracking.manager.header.subtitle')}</p>
			<h1 class="font-display text-foreground text-3xl font-bold tracking-tight md:text-4xl">
				{$t('time_tracking.manager.header.title')}
			</h1>
			<p class="text-muted-foreground">{$t('time_tracking.manager.header.desc')}</p>
		</div>
		<div class="flex gap-3">
			<Button
				onclick={handleExportReport}
				class="bg-primary text-primary-foreground flex items-center gap-2 rounded-md px-5 py-3 font-semibold shadow-md hover:bg-primary/90"
			>
				<Download size={18} />
				{$t('time_tracking.manager.actions.export')}
			</Button>
			<Button
				variant="ghost"
				class="glass-button text-foreground flex items-center gap-2 rounded-md px-5 py-3"
			>
				<Filter size={18} />
				{$t('time_tracking.manager.actions.filter')}
			</Button>
		</div>
	</div>

	<!-- Team Stats Overview -->
	<div class="grid grid-cols-1 gap-6 md:grid-cols-3">
		<Card class="liquid-panel overflow-hidden rounded-lg" data-animate="fade-up">
			<CardHeader class="pb-2">
				<CardTitle class="text-muted-foreground flex items-center gap-3">
					<Clock size={20} />
					<span>{$t('time_tracking.manager.stats.total_hours')}</span>
				</CardTitle>
			</CardHeader>
			<CardContent>
				<p class="kpi-num text-foreground text-3xl">{teamStats.totalHours}</p>
				<div class="bg-muted mt-2 h-2 w-full rounded-md">
					<div
						class="bg-primary h-2 rounded-md transition-all duration-500"
						style="width: {teamStats.progress}%"
					></div>
				</div>
				<p class="text-muted-foreground num-mono mt-1 text-sm">
					{$t('time_tracking.manager.stats.weekly_target_percent', { percent: teamStats.progress })}
				</p>
			</CardContent>
		</Card>

		<Card
			class="liquid-panel overflow-hidden rounded-lg"
			data-animate="fade-up"
			style="animation-delay: 100ms;"
		>
			<CardHeader class="pb-2">
				<CardTitle class="text-muted-foreground flex items-center gap-3">
					<Users size={20} />
					<span>{$t('time_tracking.manager.stats.active_members')}</span>
				</CardTitle>
			</CardHeader>
			<CardContent>
				<p class="kpi-num text-foreground text-3xl">{teamStats.activeMembers}/5</p>
				<p class="text-muted-foreground text-sm">{$t('time_tracking.manager.stats.working')}</p>
			</CardContent>
		</Card>

		<Card
			class="liquid-panel overflow-hidden rounded-lg"
			data-animate="fade-up"
			style="animation-delay: 200ms;"
		>
			<CardHeader class="pb-2">
				<CardTitle class="text-muted-foreground flex items-center gap-3">
					<Target size={20} />
					<span>{$t('time_tracking.manager.stats.weekly_target')}</span>
				</CardTitle>
			</CardHeader>
			<CardContent>
				<p class="kpi-num text-foreground text-3xl">{teamStats.weeklyTarget}h</p>
				<p class="text-muted-foreground text-sm">{$t('time_tracking.manager.stats.team_goal')}</p>
			</CardContent>
		</Card>
	</div>

	<!-- Team Members and Projects -->
	<div class="grid grid-cols-1 gap-8">
		<!-- Team Members -->
		<Card class="liquid-panel overflow-hidden rounded-lg" data-animate="fade-up">
			<CardHeader>
				<CardTitle class="text-foreground flex items-center gap-3">
					<Users size={20} class="text-primary" />
					<span>{$t('time_tracking.manager.team.title')}</span>
				</CardTitle>
			</CardHeader>
			<CardContent>
				<div class="space-y-3">
					{#each teamMembers as member}
						<div class="liquid-hoverable flex items-center justify-between rounded-md p-4">
							<div class="flex items-center gap-3">
								<div
									class="bg-primary text-primary-foreground flex h-10 w-10 items-center justify-center rounded-md font-semibold"
								>
									{member.avatar}
								</div>
								<div>
									<p class="text-foreground font-semibold">{member.name}</p>
									<p class="text-muted-foreground text-sm">{member.role}</p>
								</div>
							</div>
							<div class="flex items-center gap-3">
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
								<p class="num-mono text-foreground font-medium">{member.hours}</p>
							</div>
						</div>
					{/each}
				</div>
			</CardContent>
		</Card>
	</div>

	<!-- Attendance and Recent Activities -->
	<div class="grid grid-cols-1 gap-8 lg:grid-cols-2">
		<!-- Attendance Overview -->
		<Card class="liquid-panel overflow-hidden rounded-lg" data-animate="fade-up">
			<CardHeader>
				<CardTitle class="text-foreground flex items-center gap-3">
					<Calendar size={20} class="text-success" />
					<span>{$t('time_tracking.manager.attendance.title')}</span>
				</CardTitle>
			</CardHeader>
			<CardContent>
				<div class="space-y-6">
					<div class="flex h-32 items-end justify-between">
						{#each attendanceData as day}
							<div class="flex flex-1 flex-col items-center">
								<div class="flex h-24 items-end justify-center gap-1">
									<div
										class="bg-success hover:bg-success/80 w-4 rounded-t-sm transition-colors duration-300"
										style="height: {(day.present / 10) * 60}px;"
									></div>
									{#if day.absent > 0}
										<div
											class="bg-danger hover:bg-danger/80 w-4 rounded-t-sm transition-colors duration-300"
											style="height: {(day.absent / 10) * 60}px;"
										></div>
									{/if}
									{#if day.late > 0}
										<div
											class="bg-warning hover:bg-warning/80 w-4 rounded-t-sm transition-colors duration-300"
											style="height: {(day.late / 10) * 60}px;"
										></div>
									{/if}
								</div>
								<span class="text-muted-foreground mt-2 text-xs">{day.day}</span>
							</div>
						{/each}
					</div>
					<div class="flex justify-center gap-6">
						<div class="flex items-center gap-2">
							<div class="bg-success h-3 w-3 rounded-sm"></div>
							<span class="text-muted-foreground text-xs">{$t('time_tracking.manager.attendance.present')}</span>
						</div>
						<div class="flex items-center gap-2">
							<div class="bg-warning h-3 w-3 rounded-sm"></div>
							<span class="text-muted-foreground text-xs">{$t('time_tracking.manager.attendance.late')}</span>
						</div>
						<div class="flex items-center gap-2">
							<div class="bg-danger h-3 w-3 rounded-sm"></div>
							<span class="text-muted-foreground text-xs">{$t('time_tracking.manager.attendance.absent')}</span>
						</div>
					</div>
				</div>
			</CardContent>
		</Card>

		<!-- Recent Activities -->
		<Card class="liquid-panel overflow-hidden rounded-lg" data-animate="fade-up">
			<CardHeader>
				<CardTitle class="text-foreground flex items-center gap-3">
					<Activity size={20} class="text-info" />
					<span>{$t('time_tracking.manager.recent_activity.title')}</span>
				</CardTitle>
			</CardHeader>
			<CardContent>
				<div class="space-y-3">
					{#each recentActivities as activity}
						<div class="liquid-hoverable flex items-start gap-3 rounded-md p-3">
							<div class="mt-1">
								{#if activity.action === 'clocked in'}
									<div class="bg-success/10 text-success rounded-md p-1">
										<CheckCircle size={14} />
									</div>
								{:else if activity.action === 'started break'}
									<div class="bg-warning/10 text-warning rounded-md p-1">
										<Coffee size={14} />
									</div>
								{:else}
									<div class="bg-info/10 text-info rounded-md p-1">
										<FileText size={14} />
									</div>
								{/if}
							</div>
							<div class="flex-1">
								<p class="text-sm">
									<span class="text-foreground font-semibold">{activity.user}</span>
									<span class="text-muted-foreground">{activity.action}</span>
								</p>
								<div class="mt-1 flex items-center justify-end">
									<p class="num-mono text-muted-foreground text-xs font-medium">{activity.time}</p>
								</div>
							</div>
						</div>
					{/each}
				</div>
			</CardContent>
		</Card>
	</div>

	<!-- Quick Actions -->
	<Card class="liquid-panel overflow-hidden rounded-lg" data-animate="fade-up">
		<CardHeader>
			<CardTitle class="text-foreground flex items-center gap-3">
				<Settings size={20} class="text-muted-foreground" />
				<span>{$t('time_tracking.manager.quick_actions.title')}</span>
			</CardTitle>
		</CardHeader>
		<CardContent>
			<div class="grid grid-cols-2 gap-4 md:grid-cols-4">
				<Button
					variant="ghost"
					class="glass-button text-foreground flex h-24 flex-col items-center justify-center gap-2 rounded-md"
				>
					<PieChart size={24} class="text-info" />
					<span class="text-sm font-medium">{$t('time_tracking.manager.quick_actions.reports')}</span>
				</Button>
				<Button
					variant="ghost"
					class="glass-button text-foreground flex h-24 flex-col items-center justify-center gap-2 rounded-md"
				>
					<Calendar size={24} class="text-success" />
					<span class="text-sm font-medium">{$t('time_tracking.manager.quick_actions.schedule')}</span>
				</Button>
				<Button
					variant="ghost"
					class="glass-button text-foreground flex h-24 flex-col items-center justify-center gap-2 rounded-md"
				>
					<DollarSign size={24} class="text-primary" />
					<span class="text-sm font-medium">{$t('time_tracking.manager.quick_actions.payroll')}</span>
				</Button>
				<Button
					variant="ghost"
					class="glass-button text-foreground flex h-24 flex-col items-center justify-center gap-2 rounded-md"
				>
					<AlertCircle size={24} class="text-warning" />
					<span class="text-sm font-medium">{$t('time_tracking.manager.quick_actions.reminders')}</span>
				</Button>
			</div>
		</CardContent>
	</Card>
</div>
