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
		toast.success('Report Exported', {
			description: 'Weekly time report has been exported successfully.'
		});
	};

	const handleSendReminder = (memberName: string) => {
		toast.info('Reminder Sent', {
			description: `Reminder sent to ${memberName} to complete their timesheet.`
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
			toast.error('Failed to load data');
		}
	}

	onMount(() => {
		loadData();
	});
</script>

<div class="space-y-10">
	<!-- Header -->
	<div class="flex flex-col items-start justify-between gap-4 md:flex-row md:items-center">
		<div>
			<p class="text-xs font-semibold uppercase tracking-[0.3em] text-slate-500">Leadership</p>
			<h1 class="mt-2 text-3xl font-semibold text-slate-900 md:text-4xl">Team Dashboard</h1>
			<p class="text-slate-500">Monitor attendance, live shifts, and weekly momentum.</p>
		</div>
		<div class="flex gap-3">
			<Button
				onclick={handleExportReport}
				class="glass-button flex items-center gap-2 rounded-2xl bg-gradient-to-r from-blue-500/90 to-indigo-500/90 px-5 py-3 font-semibold text-white shadow-lg"
			>
				<Download size={18} />
				Export Report
			</Button>
			<Button
				variant="ghost"
				class="glass-button flex items-center gap-2 rounded-2xl px-5 py-3 text-slate-700"
			>
				<Filter size={18} />
				Filter
			</Button>
		</div>
	</div>

	<!-- Team Stats Overview -->
	<div class="grid grid-cols-1 gap-6 md:grid-cols-3">
		<Card
			class="liquid-panel overflow-hidden rounded-[28px] bg-gradient-to-br from-white/40 via-white/20 to-white/5 text-slate-900 shadow-[0_35px_90px_-60px_rgba(59,130,246,0.55)] transition-all duration-300 hover:scale-[1.01]"
		>
			<CardHeader class="pb-2">
				<CardTitle class="flex items-center gap-3 text-slate-600">
					<Clock size={20} />
					<span>Total Hours</span>
				</CardTitle>
			</CardHeader>
			<CardContent>
				<p class="text-3xl font-bold text-slate-900">{teamStats.totalHours}</p>
				<div class="mt-2 h-2 w-full rounded-full bg-slate-200/70">
					<div
						class="h-2 rounded-full bg-gradient-to-r from-blue-400 to-indigo-500 transition-all duration-500"
						style="width: {teamStats.progress}%"
					></div>
				</div>
				<p class="mt-1 text-sm text-slate-500">{teamStats.progress}% of weekly target</p>
			</CardContent>
		</Card>

		<Card
			class="liquid-panel overflow-hidden rounded-[28px] bg-gradient-to-br from-white/40 via-white/20 to-white/5 text-slate-900 shadow-[0_35px_90px_-60px_rgba(16,185,129,0.5)] transition-all duration-300 hover:scale-[1.01]"
		>
			<CardHeader class="pb-2">
				<CardTitle class="flex items-center gap-3 text-slate-600">
					<Users size={20} />
					<span>Active Members</span>
				</CardTitle>
			</CardHeader>
			<CardContent>
				<p class="text-3xl font-bold text-slate-900">{teamStats.activeMembers}/5</p>
				<p class="text-sm text-slate-500">Currently working</p>
			</CardContent>
		</Card>

		<Card
			class="liquid-panel overflow-hidden rounded-[28px] bg-gradient-to-br from-white/40 via-white/20 to-white/5 text-slate-900 shadow-[0_35px_90px_-60px_rgba(249,115,22,0.45)] transition-all duration-300 hover:scale-[1.01]"
		>
			<CardHeader class="pb-2">
				<CardTitle class="flex items-center gap-3 text-slate-600">
					<Target size={20} />
					<span>Weekly Target</span>
				</CardTitle>
			</CardHeader>
			<CardContent>
				<p class="text-3xl font-bold text-slate-900">{teamStats.weeklyTarget}h</p>
				<p class="text-sm text-slate-500">Team goal</p>
			</CardContent>
		</Card>
	</div>

	<!-- Team Members and Projects -->
	<div class="grid grid-cols-1 gap-8">
		<!-- Team Members -->
		<Card
			class="liquid-panel overflow-hidden rounded-[28px] bg-gradient-to-br from-white/40 via-white/20 to-white/5 shadow-[0_35px_90px_-60px_rgba(191,219,254,0.6)] transition-all duration-300 hover:scale-[1.01]"
		>
			<CardHeader>
				<CardTitle class="flex items-center gap-3 text-slate-600">
					<Users size={20} class="text-blue-500/80" />
					<span>Team Members</span>
				</CardTitle>
			</CardHeader>
			<CardContent>
				<div class="space-y-3">
					{#each teamMembers as member}
						<div
							class="liquid-hoverable flex items-center justify-between rounded-xl bg-gradient-to-br from-white/50 via-white/30 to-white/10 p-4 transition-all duration-200 hover:bg-white/60"
						>
							<div class="flex items-center gap-3">
								<div
									class="flex h-10 w-10 items-center justify-center rounded-full bg-gradient-to-br from-blue-400 to-indigo-500 font-semibold text-white shadow-lg shadow-blue-500/20"
								>
									{member.avatar}
								</div>
								<div>
									<p class="font-semibold text-slate-800">{member.name}</p>
									<p class="text-sm text-slate-500">{member.role}</p>
								</div>
							</div>
							<div class="text-right">
								<div class="mb-1 flex items-center gap-2">
									<div
										class="h-2 w-2 rounded-full ring-2 ring-white/50 {member.status === 'active'
											? 'animate-pulse bg-green-400 shadow-[0_0_8px_rgba(74,222,128,0.6)]'
											: member.status === 'break'
												? 'bg-yellow-400 shadow-[0_0_8px_rgba(250,204,21,0.5)]'
												: 'bg-slate-400'}"
									></div>
									<span class="text-xs font-medium text-slate-600">
										{member.status === 'active'
											? 'Working'
											: member.status === 'break'
												? 'On Break'
												: 'Offline'}
									</span>
								</div>
								<p class="font-mono font-medium text-slate-700">{member.hours}</p>
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
		<Card
			class="liquid-panel overflow-hidden rounded-[28px] bg-gradient-to-br from-white/40 via-white/20 to-white/5 shadow-[0_35px_90px_-60px_rgba(52,211,153,0.45)] transition-all duration-300 hover:scale-[1.01]"
		>
			<CardHeader>
				<CardTitle class="flex items-center gap-3 text-slate-800">
					<Calendar size={20} class="text-green-500" />
					<span>Weekly Attendance</span>
				</CardTitle>
			</CardHeader>
			<CardContent>
				<div class="space-y-6">
					<div class="flex h-32 items-end justify-between">
						{#each attendanceData as day}
							<div class="flex flex-1 flex-col items-center">
								<div class="flex h-24 items-end justify-center gap-1">
									<div
										class="w-4 rounded-t-lg bg-gradient-to-t from-green-400 to-green-600 transition-all duration-300 hover:from-green-500 hover:to-green-700"
										style="height: {(day.present / 10) * 60}px;"
									></div>
									{#if day.absent > 0}
										<div
											class="w-4 rounded-t-lg bg-gradient-to-t from-rose-400 to-rose-600 transition-all duration-300 hover:from-rose-500 hover:to-rose-700"
											style="height: {(day.absent / 10) * 60}px;"
										></div>
									{/if}
									{#if day.late > 0}
										<div
											class="w-4 rounded-t-lg bg-gradient-to-t from-amber-400 to-amber-600 transition-all duration-300 hover:from-amber-500 hover:to-amber-700"
											style="height: {(day.late / 10) * 60}px;"
										></div>
									{/if}
								</div>
								<span class="mt-2 text-xs text-slate-500">{day.day}</span>
							</div>
						{/each}
					</div>
					<div class="flex justify-center gap-6">
						<div class="flex items-center gap-2">
							<div class="h-3 w-3 rounded-full bg-green-500"></div>
							<span class="text-xs text-slate-600">Present</span>
						</div>
						<div class="flex items-center gap-2">
							<div class="h-3 w-3 rounded-full bg-amber-500"></div>
							<span class="text-xs text-slate-600">Late</span>
						</div>
						<div class="flex items-center gap-2">
							<div class="h-3 w-3 rounded-full bg-rose-500"></div>
							<span class="text-xs text-slate-600">Absent</span>
						</div>
					</div>
				</div>
			</CardContent>
		</Card>

		<!-- Recent Activities -->
		<Card
			class="liquid-panel overflow-hidden rounded-[28px] bg-gradient-to-br from-white/40 via-white/20 to-white/5 shadow-[0_35px_90px_-60px_rgba(165,180,252,0.5)] transition-all duration-300 hover:scale-[1.01]"
		>
			<CardHeader>
				<CardTitle class="flex items-center gap-3 text-slate-800">
					<Activity size={20} class="text-indigo-500" />
					<span>Recent Activities</span>
				</CardTitle>
			</CardHeader>
			<CardContent>
				<div class="space-y-3">
					{#each recentActivities as activity}
						<div
							class="liquid-hoverable flex items-start gap-3 rounded-xl bg-gradient-to-br from-white/50 via-white/30 to-white/10 p-3 transition-all duration-200 hover:bg-white/60"
						>
							<div class="mt-1">
								{#if activity.action === 'clocked in'}
									<div class="rounded-full bg-green-400/20 p-1 text-green-600">
										<CheckCircle size={14} />
									</div>
								{:else if activity.action === 'started break'}
									<div class="rounded-full bg-amber-400/20 p-1 text-amber-600">
										<Coffee size={14} />
									</div>
								{:else}
									<div class="rounded-full bg-blue-400/20 p-1 text-blue-600">
										<FileText size={14} />
									</div>
								{/if}
							</div>
							<div class="flex-1">
								<p class="text-sm text-slate-800">
									<span class="font-semibold">{activity.user}</span>
									<span class="text-slate-600">{activity.action}</span>
								</p>
								<div class="mt-1 flex items-center justify-end">
									<p class="font-mono text-xs font-medium text-slate-500">{activity.time}</p>
								</div>
							</div>
						</div>
					{/each}
				</div>
			</CardContent>
		</Card>
	</div>

	<!-- Quick Actions -->
	<Card
		class="liquid-panel overflow-hidden rounded-[28px] bg-gradient-to-br from-white/40 via-white/20 to-white/5 shadow-[0_35px_90px_-60px_rgba(148,163,184,0.5)] transition-all duration-300 hover:scale-[1.01]"
	>
		<CardHeader>
			<CardTitle class="flex items-center gap-3 text-slate-800">
				<Settings size={20} class="text-slate-500" />
				<span>Quick Actions</span>
			</CardTitle>
		</CardHeader>
		<CardContent>
			<div class="grid grid-cols-2 gap-4 md:grid-cols-4">
				<Button
					variant="ghost"
					class="glass-button flex h-24 flex-col items-center justify-center gap-2 rounded-2xl text-slate-700"
				>
					<PieChart size={24} class="text-blue-500" />
					<span class="text-sm font-medium">Reports</span>
				</Button>
				<Button
					variant="ghost"
					class="glass-button flex h-24 flex-col items-center justify-center gap-2 rounded-2xl text-slate-700"
				>
					<Calendar size={24} class="text-green-500" />
					<span class="text-sm font-medium">Schedule</span>
				</Button>
				<Button
					variant="ghost"
					class="glass-button flex h-24 flex-col items-center justify-center gap-2 rounded-2xl text-slate-700"
				>
					<DollarSign size={24} class="text-purple-500" />
					<span class="text-sm font-medium">Payroll</span>
				</Button>
				<Button
					variant="ghost"
					class="glass-button flex h-24 flex-col items-center justify-center gap-2 rounded-2xl text-slate-700"
				>
					<AlertCircle size={24} class="text-amber-500" />
					<span class="text-sm font-medium">Reminders</span>
				</Button>
			</div>
		</CardContent>
	</Card>
</div>
