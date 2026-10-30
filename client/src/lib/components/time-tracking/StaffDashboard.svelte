<script lang="ts">
	import { Button } from '$lib/components/ui/button';
	import {
		LogIn,
		LogOut,
		Timer,
		Coffee,
		Target,
		TrendingUp,
		Calendar,
		Activity
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

<div class="space-y-4">
	<!-- Dashboard header -->
	<div data-animate="fade-up">
		<p class="overline">{$t('time_tracking.staff.header.subtitle')}</p>
		<h2 class="mt-1 font-display text-xl font-semibold tracking-tight text-foreground sm:text-2xl">
			{$t('time_tracking.staff.header.title')}
		</h2>
		<p class="mt-1 text-sm text-muted-foreground">{$t('time_tracking.staff.header.desc')}</p>
	</div>

	<!-- Main Clock Card -->
	<section class="panel overflow-hidden" data-animate="fade-up">
		<div class="panel-header flex-wrap items-start">
			<div class="min-w-0">
				<h3 class="section-title text-base">{$t('time_tracking.staff.status_card.title')}</h3>
				<p class="mt-0.5 text-xs text-muted-foreground">
					{$t('time_tracking.staff.status_card.today_total')}:
					<span class="num-mono font-medium text-foreground">{todayHours}</span>
				</p>
			</div>
			<div class="flex shrink-0 items-center gap-2">
				{#if onBreak}
					<span class="badge-warning">
						<Coffee size={13} />
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
		</div>
		<div class="panel-body flex flex-col items-center py-8 text-center">
			<div class="num-mono text-5xl font-bold tracking-tighter text-foreground sm:text-6xl">
				{elapsedTime}
			</div>

			{#if onBreak}
				<div class="mt-5">
					<div
						class="mb-1 flex items-center justify-center gap-1.5 text-xs font-semibold tracking-wide text-warning uppercase"
					>
						<Coffee size={13} />
						{$t('time_tracking.staff.status_card.break_time')}
					</div>
					<div class="num-mono text-xl text-foreground">{breakTime}</div>
				</div>
			{/if}

			<div class="mt-7 flex w-full max-w-md flex-col gap-2">
				<Button onclick={handleClockToggle} class="btn-soothe h-12 w-full text-base font-semibold">
					{#if clockedIn}
						<LogOut class="mr-2" size={18} />
						{$t('time_tracking.staff.status_card.clock_out_btn')}
					{:else}
						<LogIn class="mr-2" size={18} />
						{$t('time_tracking.staff.status_card.clock_in_btn')}
					{/if}
				</Button>
				{#if clockedIn}
					<Button
						variant="outline"
						onclick={handleBreakToggle}
						class="h-12 w-full text-base font-semibold {onBreak
							? 'text-foreground'
							: 'text-warning hover:text-warning'}"
					>
						<Coffee class="mr-2" size={18} />
						{onBreak
							? $t('time_tracking.staff.status_card.end_break_btn')
							: $t('time_tracking.staff.status_card.start_break_btn')}
					</Button>
				{/if}
			</div>
		</div>
	</section>

	<!-- Hours Overview -->
	<div class="grid grid-cols-1 gap-4 md:grid-cols-2">
		<section class="panel p-5" data-animate="fade-up" style="animation-delay: 100ms;">
			<div class="flex items-center gap-2 text-muted-foreground">
				<Timer size={16} />
				<span class="text-xs font-medium">{$t('time_tracking.staff.stats.today_hours')}</span>
			</div>
			<p class="kpi-num mt-2 text-3xl text-foreground">{todayHours}</p>
			<div class="mt-3 h-1.5 w-full overflow-hidden rounded-full bg-muted">
				<div
					class="h-full rounded-full bg-primary transition-all duration-500"
					style="width: {(parseFloat(todayHours) / todayTarget) * 100}%"
				></div>
			</div>
			<p class="mt-2 text-xs text-muted-foreground">
				{$t('time_tracking.staff.stats.target_label', { target: todayTarget })}
			</p>
		</section>

		<section class="panel p-5" data-animate="fade-up" style="animation-delay: 180ms;">
			<div class="flex items-center gap-2 text-muted-foreground">
				<Calendar size={16} />
				<span class="text-xs font-medium">{$t('time_tracking.staff.stats.weekly_hours')}</span>
			</div>
			<p class="kpi-num mt-2 text-3xl text-foreground">{weeklyHours}</p>
			<div class="mt-3 h-1.5 w-full overflow-hidden rounded-full bg-muted">
				<div
					class="h-full rounded-full bg-info transition-all duration-500"
					style="width: {weeklyProgress}%"
				></div>
			</div>
			<p class="mt-2 text-xs text-muted-foreground">
				{$t('time_tracking.staff.stats.target_label', { target: weeklyTarget })}
			</p>
		</section>
	</div>

	<!-- Current Task & Daily Goals -->
	<div class="grid grid-cols-1 gap-4 lg:grid-cols-2">
		<!-- Current Task -->
		<section class="panel overflow-hidden" data-animate="fade-up" style="animation-delay: 240ms;">
			<div class="panel-header">
				<h3 class="section-title flex items-center gap-2">
					<Target size={16} class="text-primary" />
					{$t('time_tracking.staff.task.title')}
				</h3>
			</div>
			<div class="panel-body space-y-3">
				<div>
					<label class="mb-1.5 block text-sm font-medium text-foreground" for="current-task"
						>{$t('time_tracking.staff.task.label')}</label
					>
					<input
						id="current-task"
						bind:value={currentTask}
						type="text"
						placeholder={$t('time_tracking.staff.task.placeholder')}
						class="glass-input h-11 w-full rounded-md px-3 text-sm"
					/>
				</div>
				<Button onclick={updateTask} class="btn-soothe w-full">
					{$t('time_tracking.staff.task.button')}
				</Button>
			</div>
		</section>

		<!-- Daily Goals -->
		<section class="panel overflow-hidden" data-animate="fade-up" style="animation-delay: 300ms;">
			<div class="panel-header">
				<h3 class="section-title flex items-center gap-2">
					<TrendingUp size={16} class="text-success" />
					{$t('time_tracking.staff.goals.title')}
				</h3>
			</div>
			<div class="panel-body space-y-4">
				{#each dailyStats.goals as goal}
					<div class="space-y-1.5">
						<div class="flex items-center justify-between gap-3">
							<span class="text-sm font-medium text-foreground">{goal.name}</span>
							<span class="num-mono text-xs text-muted-foreground"
								>{goal.current}{goal.unit} / {goal.target}{goal.unit}</span
							>
						</div>
						<div class="h-1.5 w-full overflow-hidden rounded-full bg-muted">
							<div
								class="h-full rounded-full bg-success transition-all duration-500"
								style="width: {(goal.current / goal.target) * 100}%"
							></div>
						</div>
					</div>
				{/each}
			</div>
		</section>
	</div>

	<!-- Recent Shifts -->
	<section class="panel overflow-hidden" data-animate="fade-up" style="animation-delay: 360ms;">
		<div class="panel-header">
			<h3 class="section-title flex items-center gap-2">
				<Activity size={16} class="text-info" />
				{$t('time_tracking.staff.recent_shifts.title')}
			</h3>
		</div>
		<div class="panel-body space-y-1.5">
			{#each recentShifts as shift}
				<div
					class="flex items-center justify-between gap-3 rounded-md px-2.5 py-2.5 transition-colors hover:bg-muted/60"
				>
					<div class="min-w-0">
						<p class="truncate text-sm font-semibold text-foreground">{shift.date}</p>
						<p class="num-mono mt-0.5 text-xs text-muted-foreground">{shift.time}</p>
					</div>
					<p class="num-mono shrink-0 text-sm font-semibold text-foreground">{shift.duration}</p>
				</div>
			{:else}
				<p class="text-muted-foreground px-2 py-6 text-center text-sm">No shifts recorded yet.</p>
			{/each}
		</div>
	</section>
</div>
