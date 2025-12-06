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

<div class="min-h-screen bg-gradient-to-br from-slate-50 via-blue-50 to-purple-50 p-4 md:p-8">
	<div class="mx-auto max-w-7xl space-y-8">
		<!-- Header -->
		<div class="mb-6 flex flex-col items-start justify-between gap-4 md:flex-row md:items-center">
			<div>
				<h1
					class="mb-2 bg-gradient-to-r from-blue-600 via-purple-600 to-indigo-600 bg-clip-text text-3xl font-bold text-transparent md:text-4xl"
				>
					Team Dashboard
				</h1>
				<p class="text-slate-600">Monitor your team's productivity and time tracking</p>
			</div>
			<div class="flex gap-3">
				<Button
					on:click={handleExportReport}
					class="flex items-center gap-2 rounded-xl bg-gradient-to-r from-blue-500 to-indigo-600 px-4 py-2 font-semibold text-white transition-all duration-300 hover:scale-105 hover:from-blue-600 hover:to-indigo-700"
				>
					<Download size={18} />
					Export Report
				</Button>
				<Button
					variant="outline"
					class="flex items-center gap-2 rounded-xl border-slate-200 px-4 py-2 font-semibold text-slate-700 transition-all duration-300 hover:border-blue-300 hover:bg-blue-50"
				>
					<Filter size={18} />
					Filter
				</Button>
			</div>
		</div>

		<!-- Team Stats Overview -->
		<div class="grid grid-cols-1 gap-6 md:grid-cols-3">
			<Card
				class="overflow-hidden rounded-2xl border-0 bg-gradient-to-br from-blue-400 to-indigo-500 text-white shadow-lg transition-all duration-300 hover:scale-[1.02] hover:shadow-xl"
			>
				<CardHeader class="pb-2">
					<CardTitle class="flex items-center gap-3 text-white/90">
						<Clock size={20} />
						<span>Total Hours</span>
					</CardTitle>
				</CardHeader>
				<CardContent>
					<p class="text-3xl font-bold text-white">{teamStats.totalHours}</p>
					<div class="mt-2 h-2 w-full rounded-full bg-white/20">
						<div
							class="h-2 rounded-full bg-white transition-all duration-500"
							style="width: {teamStats.progress}%"
						></div>
					</div>
					<p class="mt-1 text-sm text-white/80">{teamStats.progress}% of weekly target</p>
				</CardContent>
			</Card>

			<Card
				class="overflow-hidden rounded-2xl border-0 bg-gradient-to-br from-emerald-400 to-green-500 text-white shadow-lg transition-all duration-300 hover:scale-[1.02] hover:shadow-xl"
			>
				<CardHeader class="pb-2">
					<CardTitle class="flex items-center gap-3 text-white/90">
						<Users size={20} />
						<span>Active Members</span>
					</CardTitle>
				</CardHeader>
				<CardContent>
					<p class="text-3xl font-bold text-white">{teamStats.activeMembers}/5</p>
					<p class="text-sm text-white/80">Currently working</p>
				</CardContent>
			</Card>

			<Card
				class="overflow-hidden rounded-2xl border-0 bg-gradient-to-br from-amber-400 to-orange-500 text-white shadow-lg transition-all duration-300 hover:scale-[1.02] hover:shadow-xl"
			>
				<CardHeader class="pb-2">
					<CardTitle class="flex items-center gap-3 text-white/90">
						<Target size={20} />
						<span>Weekly Target</span>
					</CardTitle>
				</CardHeader>
				<CardContent>
					<p class="text-3xl font-bold text-white">{teamStats.weeklyTarget}h</p>
					<p class="text-sm text-white/80">Team goal</p>
				</CardContent>
			</Card>
		</div>

		<!-- Team Members and Projects -->
		<div class="grid grid-cols-1 gap-8">
			<!-- Team Members -->
			<Card
				class="overflow-hidden rounded-2xl border-0 bg-gradient-to-br from-white to-slate-50 shadow-lg transition-all duration-300 hover:scale-[1.01] hover:shadow-xl"
			>
				<CardHeader>
					<CardTitle class="flex items-center gap-3 text-slate-800">
						<Users size={20} class="text-blue-500" />
						<span>Team Members</span>
					</CardTitle>
				</CardHeader>
				<CardContent>
					<div class="space-y-3">
						{#each teamMembers as member}
							<div
								class="flex items-center justify-between rounded-xl border border-slate-200/80 bg-slate-50/50 p-4 transition-all duration-200 hover:bg-slate-100"
							>
								<div class="flex items-center gap-3">
									<div
										class="flex h-10 w-10 items-center justify-center rounded-full bg-gradient-to-br from-blue-400 to-indigo-500 font-semibold text-white"
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
											class="h-2 w-2 rounded-full {member.status === 'active'
												? 'animate-pulse bg-green-400'
												: member.status === 'break'
													? 'bg-yellow-400'
													: 'bg-slate-400'}"
										></div>
										<span class="text-xs text-slate-500">
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
				class="overflow-hidden rounded-2xl border-0 bg-gradient-to-br from-white to-slate-50 shadow-lg transition-all duration-300 hover:scale-[1.01] hover:shadow-xl"
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
				class="overflow-hidden rounded-2xl border-0 bg-gradient-to-br from-white to-slate-50 shadow-lg transition-all duration-300 hover:scale-[1.01] hover:shadow-xl"
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
								class="flex items-start gap-3 rounded-xl border border-slate-200/80 bg-slate-50/50 p-3 transition-all duration-200 hover:bg-slate-100"
							>
								<div class="mt-1">
									{#if activity.action === 'clocked in'}
										<CheckCircle size={16} class="text-green-500" />
									{:else if activity.action === 'started break'}
										<Coffee size={16} class="text-amber-500" />
									{:else}
										<FileText size={16} class="text-blue-500" />
									{/if}
								</div>
								<div class="flex-1">
									<p class="text-sm text-slate-800">
										<span class="font-semibold">{activity.user}</span>
										{activity.action}
									</p>
									<div class="mt-1 flex items-center justify-end">
										<p class="text-xs text-slate-500">{activity.time}</p>
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
			class="overflow-hidden rounded-2xl border-0 bg-gradient-to-br from-white to-slate-50 shadow-lg transition-all duration-300 hover:scale-[1.01] hover:shadow-xl"
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
						variant="outline"
						class="flex h-20 flex-col items-center justify-center gap-2 border-slate-200 transition-all duration-200 hover:border-blue-300 hover:bg-blue-50"
					>
						<PieChart size={24} class="text-blue-500" />
						<span class="text-sm font-medium">Reports</span>
					</Button>
					<Button
						variant="outline"
						class="flex h-20 flex-col items-center justify-center gap-2 border-slate-200 transition-all duration-200 hover:border-green-300 hover:bg-green-50"
					>
						<Calendar size={24} class="text-green-500" />
						<span class="text-sm font-medium">Schedule</span>
					</Button>
					<Button
						variant="outline"
						class="flex h-20 flex-col items-center justify-center gap-2 border-slate-200 transition-all duration-200 hover:border-purple-300 hover:bg-purple-50"
					>
						<DollarSign size={24} class="text-purple-500" />
						<span class="text-sm font-medium">Payroll</span>
					</Button>
					<Button
						variant="outline"
						class="flex h-20 flex-col items-center justify-center gap-2 border-slate-200 transition-all duration-200 hover:border-amber-300 hover:bg-amber-50"
					>
						<AlertCircle size={24} class="text-amber-500" />
						<span class="text-sm font-medium">Reminders</span>
					</Button>
				</div>
			</CardContent>
		</Card>
	</div>
</div>
