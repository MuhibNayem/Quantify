<script lang="ts">
	import { Button } from '$lib/components/ui/button';
	import { Card, CardContent, CardHeader, CardTitle } from '$lib/components/ui/card';
	import {
		Clock,
		LogIn,
		LogOut,
		Timer,
		Coffee,
		Target,
		TrendingUp,
		Calendar,
		Activity,
		CheckCircle,
		FileText
	} from 'lucide-svelte';
	import { onMount, onDestroy } from 'svelte';
	import { toast } from 'svelte-sonner';
	import { timeTrackingApi } from '$lib/api/resources';
	import { page } from '$app/stores';
	import type { TimeClock } from '$lib/types';
	import { t } from '$lib/i18n';

	// Core state
	let clockedIn = false;
	let shiftStartTime: Date | null = null;
	let elapsedTime = '00:00:00';
	let timerInterval: any;

	// Enhanced features
	let onBreak = false;
	let breakStartTime: Date | null = null;
	let breakTime = '00:00:00';
	let currentTask = '';
	// let productivityScore = 87; // Removed
	let weeklyTarget = 40;
	let todayTarget = 8;
	let user = $page.data.user;

	// Data
	let todayHours = '0h 00m';
	let weeklyHours = '0h 00m';
	let weeklyProgress = 0;

	let recentShifts: any[] = [];

	let dailyStats = {
		goals: [
			{ name: 'Focus Time', current: 0, target: 8, unit: 'h' },
			{ name: 'Breaks Taken', current: 0, target: 4, unit: '' }
		]
	};

	async function loadData() {
		if (!user?.id) return;

		try {
			// 1. Get Status
			const lastEntry = await timeTrackingApi.getLastEntry(user.id);
			if (lastEntry) {
				if (lastEntry.Status === 'CLOCKED_IN') {
					clockedIn = true;
					shiftStartTime = new Date(lastEntry.ClockIn);
					startTimer();
				} else if (lastEntry.Status === 'ON_BREAK') {
					clockedIn = true;
					shiftStartTime = new Date(lastEntry.ClockIn);
					startTimer();

					onBreak = true;
					if (lastEntry.BreakStart) {
						breakStartTime = new Date(lastEntry.BreakStart);
						startBreakTimer();
					}
				}
			}

			// 2. Get History
			const history = await timeTrackingApi.getHistory(user.id);
			if (history && history.length > 0) {
				recentShifts = history.map((h) => {
					const start = new Date(h.ClockIn);
					const end = h.ClockOut ? new Date(h.ClockOut) : new Date(); // Approximate if active
					const durationMs = end.getTime() - start.getTime();
					const hours = Math.floor(durationMs / (1000 * 60 * 60));
					const minutes = Math.floor((durationMs % (1000 * 60 * 60)) / (1000 * 60));

					return {
						date: start.toLocaleDateString('en-US', {
							weekday: 'long',
							month: 'short',
							day: 'numeric'
						}),
						time: `${start.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })} - ${h.ClockOut ? new Date(h.ClockOut).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' }) : 'Now'}`,
					duration: `${hours}h ${minutes}m`
				};
				});
			}

			// 3. Get Weekly Summary
			const summary = await timeTrackingApi.getWeeklySummary(user.id);
			if (summary) {
				weeklyHours = summary.weeklyHours;
				weeklyProgress = summary.weeklyProgress;

				// Update Daily Goals
				// Parse "Xh Ym" from summary.todayHours to float for progress bar?
				// The API returns strings like "5h 30m" for display.
				// I need float for progress.
				// API also returns summary.todayProgress (percentage).

				// Extract hours from todayHours string or use progress
				// Let's use progress.

				dailyStats.goals[0].current = parseFloat(summary.todayHours.split('h')[0]); // Approx

				// Count breaks from history or separate endpoint?
				// Using history for today breaks count
				const todayDate = new Date().toDateString();
				const breaks = history.filter(
					(h) => h.Status === 'ON_BREAK' && new Date(h.ClockIn).toDateString() === todayDate
				).length;
				// Wait, history items are shifts/breaks? My TimeClock model mixes them.
				// Status=ON_BREAK means it IS a break record?
				// Yes, if I look at StartBreak implementation, it creates a TimeClock with status ON_BREAK.

				dailyStats.goals[1].current = history.filter(
					(h) => h.Status === 'ON_BREAK' && new Date(h.ClockIn).toDateString() === todayDate
				).length;

				todayHours = summary.todayHours;
			}
		} catch (error) {
			console.error('Failed to load time tracking data', error);
		}
	}

	const handleClockToggle = async () => {
		if (!user?.id) return;
		try {
			if (!clockedIn) {
				const entry = await timeTrackingApi.clockIn(
					user.id,
					currentTask ? `Task: ${currentTask}` : ''
				);
				clockedIn = true;
				shiftStartTime = new Date(entry.ClockIn);
				startTimer();
				toast.success($t('time_tracking.toasts.clock_in_success'), {
					description: $t('time_tracking.toasts.clock_in_desc')
				});
				loadData();
			} else {
				await timeTrackingApi.clockOut(user.id, currentTask ? `Task: ${currentTask}` : '');
				clockedIn = false;
				stopTimer();
				if (onBreak) {
					stopBreakTimer();
					onBreak = false;
				}
				toast.info($t('time_tracking.toasts.clock_out_info'), {
					description: $t('time_tracking.toasts.clock_out_desc', { time: elapsedTime })
				});
				shiftStartTime = null;
				loadData(); // Refresh history
			}
		} catch (e) {
			toast.error($t('time_tracking.toasts.op_fail'), { description: (e as Error).message });
		}
	};

	const handleBreakToggle = async () => {
		if (!user?.id) return;
		try {
			if (!onBreak) {
				const entry = await timeTrackingApi.startBreak(user.id);
				onBreak = true;
				breakStartTime = new Date(entry.BreakStart!);
				startBreakTimer();
				toast.info($t('time_tracking.toasts.break_start'), {
					description: $t('time_tracking.toasts.break_start_desc')
				});
			} else {
				await timeTrackingApi.endBreak(user.id);
				onBreak = false;
				stopBreakTimer();
				toast.success($t('time_tracking.toasts.break_end'), {
					description: $t('time_tracking.toasts.break_end_desc')
				});
				// breakStartTime = null; // Keep it for display? No, reset.
				breakStartTime = null;
			}
		} catch (e) {
			toast.error($t('time_tracking.toasts.op_fail'), { description: (e as Error).message });
		}
	};

	const startTimer = () => {
		if (timerInterval) clearInterval(timerInterval);
		timerInterval = setInterval(() => {
			if (shiftStartTime) {
				// removed && !onBreak condition to keep tracking shift duration even during break (or decided otherwise?)
				// User usually wants to see elapsed time since clock in.
				const now = new Date();
				const diff = now.getTime() - shiftStartTime.getTime();
				const hours = Math.floor(diff / (1000 * 60 * 60));
				const minutes = Math.floor((diff % (1000 * 60 * 60)) / (1000 * 60));
				const seconds = Math.floor((diff % (1000 * 60)) / 1000);
				elapsedTime = `${String(hours).padStart(2, '0')}:${String(minutes).padStart(2, '0')}:${String(seconds).padStart(2, '0')}`;
			}
		}, 1000);
	};

	const stopTimer = () => {
		clearInterval(timerInterval);
	};

	let breakInterval: any;
	const startBreakTimer = () => {
		if (breakInterval) clearInterval(breakInterval);
		breakInterval = setInterval(() => {
			if (breakStartTime) {
				const now = new Date();
				const diff = now.getTime() - breakStartTime.getTime();
				const hours = Math.floor(diff / (1000 * 60 * 60));
				const minutes = Math.floor((diff % (1000 * 60 * 60)) / (1000 * 60));
				const seconds = Math.floor((diff % (1000 * 60)) / 1000);
				breakTime = `${String(hours).padStart(2, '0')}:${String(minutes).padStart(2, '0')}:${String(seconds).padStart(2, '0')}`;
			}
		}, 1000);
	};

	const stopBreakTimer = () => {
		clearInterval(breakInterval);
	};

	const updateTask = () => {
		if (currentTask) {
			toast.success($t('time_tracking.toasts.task_updated'), {
				description: $t('time_tracking.toasts.task_updated_desc', { task: currentTask })
			});
			// In real implementation, could update Notes of current TimeClock entry
		}
	};

	onMount(() => {
		loadData();
	});

	onDestroy(() => {
		stopTimer();
		stopBreakTimer();
	});
</script>

<div class="space-y-10">
	<!-- Header -->
	<div class="space-y-1" data-animate="fade-up">
		<p class="overline">{$t('time_tracking.staff.header.subtitle')}</p>
		<h1 class="font-display text-foreground text-3xl font-bold tracking-tight md:text-4xl">
			{$t('time_tracking.staff.header.title')}
		</h1>
		<p class="text-muted-foreground">{$t('time_tracking.staff.header.desc')}</p>
	</div>

	<!-- Main Clock Card -->
	<Card class="liquid-panel overflow-hidden rounded-lg" data-animate="fade-up">
		<CardHeader class="pb-4">
			<CardTitle class="text-foreground flex items-center justify-between">
				<span class="text-xl font-semibold">{$t('time_tracking.staff.status_card.title')}</span>
				<div class="flex items-center gap-3">
					{#if onBreak}
						<span class="badge-warning">
							<Coffee size={14} />
							{$t('time_tracking.staff.status_card.on_break')}
						</span>
					{:else if clockedIn}
						<span class="badge-success">
							<span class="live-dot" style="width: 0.4rem; height: 0.4rem;"></span>
							{$t('time_tracking.staff.status_card.clocked_in')}
						</span>
					{:else}
						<span class="badge-neutral">
							{$t('time_tracking.staff.status_card.clocked_out')}
						</span>
					{/if}
				</div>
			</CardTitle>
		</CardHeader>
		<CardContent class="p-8 text-center">
			<div class="num-mono text-foreground mb-6 text-6xl font-bold tracking-tighter md:text-7xl">
				{elapsedTime}
			</div>

			{#if onBreak}
				<div class="mb-6">
					<div class="text-warning mb-2 text-lg font-medium">
						{$t('time_tracking.staff.status_card.break_time')}
					</div>
					<div class="num-mono text-foreground text-2xl">{breakTime}</div>
				</div>
			{/if}

			<div class="mb-6 flex flex-col gap-4 md:flex-row">
				<div class="bg-muted flex-1 rounded-md p-4">
					<div class="num-mono text-foreground text-lg font-bold">{todayHours}</div>
					<div class="text-muted-foreground text-sm font-medium">
						{$t('time_tracking.staff.status_card.today_total')}
					</div>
				</div>
			</div>

			<div class="flex flex-col gap-4 md:flex-row">
				<Button
					onclick={handleClockToggle}
					class="bg-primary text-primary-foreground flex-1 rounded-md py-4 text-lg font-semibold shadow-md transition-all hover:bg-primary/90"
				>
					{#if clockedIn}
						<LogOut class="mr-2" size={20} />
						{$t('time_tracking.staff.status_card.clock_out_btn')}
					{:else}
						<LogIn class="mr-2" size={20} />
						{$t('time_tracking.staff.status_card.clock_in_btn')}
					{/if}
				</Button>
				{#if clockedIn}
					<Button
						onclick={handleBreakToggle}
						class={onBreak
							? 'glass-button text-foreground flex-1 rounded-md py-4 text-lg font-semibold'
							: 'bg-warning text-primary-foreground flex-1 rounded-md py-4 text-lg font-semibold shadow-md hover:bg-warning/90'}
					>
						<Coffee class="mr-2" size={20} />
						{onBreak ? $t('time_tracking.staff.status_card.end_break_btn') : $t('time_tracking.staff.status_card.start_break_btn')}
					</Button>
				{/if}
			</div>
		</CardContent>
	</Card>

	<!-- Hours Overview -->
	<div class="grid grid-cols-1 gap-8 md:grid-cols-2">
		<Card
			class="liquid-panel overflow-hidden rounded-lg"
			data-animate="fade-up"
			style="animation-delay: 150ms;"
		>
			<CardHeader class="pb-2">
				<CardTitle class="text-muted-foreground flex items-center gap-3">
					<Timer size={20} />
					<span>{$t('time_tracking.staff.stats.today_hours')}</span>
				</CardTitle>
			</CardHeader>
			<CardContent>
				<p class="kpi-num text-foreground mb-2 text-4xl">{todayHours}</p>
				<div class="bg-muted mb-2 h-3 w-full rounded-md">
					<div
						class="bg-primary h-3 rounded-md transition-all duration-500"
						style="width: {(parseFloat(todayHours) / todayTarget) * 100}%"
					></div>
				</div>
				<p class="text-muted-foreground text-sm">{$t('time_tracking.staff.stats.target_label', { target: todayTarget })}</p>
			</CardContent>
		</Card>

		<Card
			class="liquid-panel overflow-hidden rounded-lg"
			data-animate="fade-up"
			style="animation-delay: 300ms;"
		>
			<CardHeader class="pb-2">
				<CardTitle class="text-muted-foreground flex items-center gap-3">
					<Calendar size={20} />
					<span>{$t('time_tracking.staff.stats.weekly_hours')}</span>
				</CardTitle>
			</CardHeader>
			<CardContent>
				<p class="kpi-num text-foreground mb-2 text-4xl">{weeklyHours}</p>
				<div class="bg-muted mb-2 h-3 w-full rounded-md">
					<div
						class="bg-info h-3 rounded-md transition-all duration-500"
						style="width: {weeklyProgress}%"
					></div>
				</div>
				<p class="text-muted-foreground text-sm">{$t('time_tracking.staff.stats.target_label', { target: weeklyTarget })}</p>
			</CardContent>
		</Card>
	</div>

	<!-- Current Task & Daily Goals -->
	<div class="grid grid-cols-1 gap-8 lg:grid-cols-2">
		<!-- Current Task -->
		<Card
			class="liquid-panel overflow-hidden rounded-lg"
			data-animate="fade-up"
			style="animation-delay: 450ms;"
		>
			<CardHeader>
				<CardTitle class="text-foreground flex items-center gap-3">
					<Target size={20} class="text-primary" />
					<span>{$t('time_tracking.staff.task.title')}</span>
				</CardTitle>
			</CardHeader>
			<CardContent>
				<div class="space-y-4">
					<div>
						<label class="text-foreground mb-2 block text-sm font-medium"
							>{$t('time_tracking.staff.task.label')}</label
						>
						<input
							bind:value={currentTask}
							type="text"
							placeholder={$t('time_tracking.staff.task.placeholder')}
							class="liquid-input text-foreground w-full px-4 py-3 text-sm placeholder:text-muted-foreground"
						/>
					</div>
					<Button
						onclick={updateTask}
						class="bg-primary text-primary-foreground w-full rounded-md py-3 font-semibold hover:bg-primary/90"
					>
						{$t('time_tracking.staff.task.button')}
					</Button>
				</div>
			</CardContent>
		</Card>

		<!-- Daily Goals -->
		<Card
			class="liquid-panel overflow-hidden rounded-lg"
			data-animate="fade-up"
			style="animation-delay: 600ms;"
		>
			<CardHeader>
				<CardTitle class="text-foreground flex items-center gap-3">
					<TrendingUp size={20} class="text-success" />
					<span>{$t('time_tracking.staff.goals.title')}</span>
				</CardTitle>
			</CardHeader>
			<CardContent>
				<div class="space-y-4">
					{#each dailyStats.goals as goal}
						<div class="space-y-2">
							<div class="flex items-center justify-between">
								<span class="text-foreground text-sm font-medium">{goal.name}</span>
								<span class="text-muted-foreground num-mono text-sm"
									>{goal.current}{goal.unit} / {goal.target}{goal.unit}</span
								>
							</div>
							<div class="bg-muted h-2 w-full rounded-md">
								<div
									class="bg-success h-2 rounded-md transition-all duration-500"
									style="width: {(goal.current / goal.target) * 100}%"
								></div>
							</div>
						</div>
					{/each}
				</div>
			</CardContent>
		</Card>
	</div>

	<!-- Recent Shifts -->
	<Card
		class="liquid-panel overflow-hidden rounded-lg"
		data-animate="fade-up"
		style="animation-delay: 750ms;"
	>
		<CardHeader>
			<CardTitle class="text-foreground flex items-center gap-3">
				<Activity size={20} class="text-info" />
				<span>{$t('time_tracking.staff.recent_shifts.title')}</span>
			</CardTitle>
		</CardHeader>
		<CardContent>
			<div class="space-y-3">
				{#each recentShifts as shift}
					<div class="liquid-hoverable flex items-center justify-between rounded-md p-4">
						<div class="flex-1">
							<div class="mb-1 flex items-center gap-3">
								<p class="text-foreground font-semibold">{shift.date}</p>
							</div>
							<p class="text-muted-foreground num-mono text-sm">{shift.time}</p>
						</div>
						<div class="text-right">
							<p class="num-mono text-foreground text-lg font-semibold">{shift.duration}</p>
						</div>
					</div>
				{/each}
			</div>
		</CardContent>
	</Card>
</div>
