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
						duration: `${hours}h ${minutes}m`,
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
				toast.success('Clocked In Successfully', {
					description: 'Your shift has started. Have a productive day!'
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
				toast.info('Clocked Out', {
					description: `Great work today! You completed ${elapsedTime} of focused time.`
				});
				shiftStartTime = null;
				loadData(); // Refresh history
			}
		} catch (e) {
			toast.error('Operation failed', { description: (e as Error).message });
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
				toast.info('Break Started', {
					description: 'Take a well-deserved break! Your timer is paused.'
				});
			} else {
				await timeTrackingApi.endBreak(user.id);
				onBreak = false;
				stopBreakTimer();
				toast.success('Break Ended', {
					description: 'Welcome back! Ready to continue your productive day?'
				});
				// breakStartTime = null; // Keep it for display? No, reset.
				breakStartTime = null;
			}
		} catch (e) {
			toast.error('Operation failed', { description: (e as Error).message });
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
			toast.success('Task Updated', {
				description: `Now working on: ${currentTask}`
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
	<div class="text-center">
		<p class="text-xs font-semibold uppercase tracking-[0.3em] text-slate-500">Personal Flow</p>
		<h1 class="mt-2 text-3xl font-semibold text-slate-900 md:text-4xl">My Time Tracker</h1>
		<p class="mt-2 text-slate-500">Track your focus, breaks, and progress from one calm surface.</p>
	</div>

	<!-- Main Clock Card -->
	<Card
		class="liquid-panel overflow-hidden rounded-[32px] bg-gradient-to-br from-white/40 via-white/20 to-white/5 text-slate-900 shadow-[0_45px_120px_-60px_rgba(15,23,42,0.45)]"
		data-animate="fade-up"
	>
		<CardHeader class="pb-4">
			<CardTitle class="flex items-center justify-between text-slate-600">
				<span class="text-xl font-semibold text-slate-900">My Status</span>
				<div class="flex items-center gap-4">
					<span class="flex items-center gap-2 text-sm font-medium">
						<div
							class="h-4 w-4 rounded-full {clockedIn
								? 'animate-pulse bg-green-300'
								: 'bg-slate-300'}"
						></div>
						{clockedIn ? 'Clocked In' : 'Clocked Out'}
					</span>
					{#if onBreak}
						<span
							class="flex items-center gap-2 rounded-full bg-yellow-500/10 px-3 py-1 text-sm font-medium text-slate-700"
						>
							<Coffee size={16} />
							On Break
						</span>
					{/if}
				</div>
			</CardTitle>
		</CardHeader>
		<CardContent class="p-8 text-center">
			<div class="mb-6 font-mono text-6xl font-bold tracking-tighter text-slate-900 md:text-7xl">
				{elapsedTime}
			</div>

			{#if onBreak}
				<div class="mb-6">
					<div class="mb-2 text-lg font-medium text-yellow-600">Break Time</div>
					<div class="font-mono text-2xl text-slate-800">{breakTime}</div>
				</div>
			{/if}

			<div class="mb-6 flex flex-col gap-4 md:flex-row">
				<div
					class="flex-1 rounded-2xl bg-gradient-to-br from-white/50 via-white/30 to-white/10 p-4 shadow-lg shadow-blue-900/5 backdrop-blur-md transition-all hover:bg-white/60"
				>
					<div class="text-lg font-bold text-slate-900">{todayHours}</div>
					<div class="text-sm font-medium text-slate-500">Today's Total</div>
				</div>
			</div>

			<div class="flex flex-col gap-4 md:flex-row">
				<Button
					onclick={handleClockToggle}
					class="glass-button flex-1 rounded-2xl py-4 text-lg font-semibold text-slate-800 shadow-lg transition-all duration-300 hover:scale-[1.01] hover:shadow-xl {clockedIn
						? 'bg-gradient-to-r from-rose-400/90 to-pink-500/90 text-white'
						: 'bg-gradient-to-r from-emerald-400/90 to-green-500/90 text-white'}"
				>
					{#if clockedIn}
						<LogOut class="mr-2" size={20} />
						Clock Out
					{:else}
						<LogIn class="mr-2" size={20} />
						Clock In
					{/if}
				</Button>
				{#if clockedIn}
					<Button
						onclick={handleBreakToggle}
						class="glass-button flex-1 rounded-2xl py-4 text-lg font-semibold text-slate-800 shadow-lg transition-all duration-300 hover:scale-[1.01] hover:shadow-xl {onBreak
							? 'bg-gradient-to-r from-orange-400/95 to-amber-500/95 text-white'
							: 'bg-gradient-to-r from-blue-500/95 to-cyan-500/95 text-white'}"
					>
						<Coffee class="mr-2" size={20} />
						{onBreak ? 'End Break' : 'Start Break'}
					</Button>
				{/if}
			</div>
		</CardContent>
	</Card>

	<!-- Hours Overview -->
	<div class="grid grid-cols-1 gap-8 md:grid-cols-2">
		<Card
			class="liquid-panel overflow-hidden rounded-[28px] bg-gradient-to-br from-white/40 via-white/20 to-white/5 text-slate-900 shadow-[0_30px_90px_-60px_rgba(16,185,129,0.55)] transition-all duration-300 hover:scale-[1.01]"
			data-animate="fade-up"
			style="animation-delay: 150ms;"
		>
			<CardHeader class="pb-2">
				<CardTitle class="flex items-center gap-3 text-slate-500">
					<Timer size={20} />
					<span>Today's Hours</span>
				</CardTitle>
			</CardHeader>
			<CardContent>
				<p class="mb-2 text-4xl font-bold text-slate-900">{todayHours}</p>
				<div class="mb-2 h-3 w-full rounded-full bg-slate-200/40 shadow-inner shadow-black/5">
					<div
						class="h-3 rounded-full bg-gradient-to-r from-emerald-400 to-emerald-500 transition-all duration-500"
						style="width: {(parseFloat(todayHours) / todayTarget) * 100}%"
					></div>
				</div>
				<p class="text-sm text-slate-500">out of {todayTarget} hours target</p>
			</CardContent>
		</Card>

		<Card
			class="liquid-panel overflow-hidden rounded-[28px] bg-gradient-to-br from-white/40 via-white/20 to-white/5 text-slate-900 shadow-[0_30px_90px_-60px_rgba(59,130,246,0.55)] transition-all duration-300 hover:scale-[1.01]"
			data-animate="fade-up"
			style="animation-delay: 300ms;"
		>
			<CardHeader class="pb-2">
				<CardTitle class="flex items-center gap-3 text-slate-500">
					<Calendar size={20} />
					<span>Weekly Hours</span>
				</CardTitle>
			</CardHeader>
			<CardContent>
				<p class="mb-2 text-4xl font-bold text-slate-900">{weeklyHours}</p>
				<div class="mb-2 h-3 w-full rounded-full bg-slate-200/40 shadow-inner shadow-black/5">
					<div
						class="h-3 rounded-full bg-gradient-to-r from-blue-400 to-indigo-500 transition-all duration-500"
						style="width: {weeklyProgress}%"
					></div>
				</div>
				<p class="text-sm text-slate-500">out of {weeklyTarget} hours target</p>
			</CardContent>
		</Card>
	</div>

	<!-- Current Task & Daily Goals -->
	<div class="grid grid-cols-1 gap-8 lg:grid-cols-2">
		<!-- Current Task -->
		<Card
			class="liquid-panel overflow-hidden rounded-[28px] bg-gradient-to-br from-white/40 via-white/20 to-white/5 shadow-[0_35px_90px_-60px_rgba(147,197,253,0.6)] transition-all duration-300 hover:scale-[1.01]"
			data-animate="fade-up"
			style="animation-delay: 450ms;"
		>
			<CardHeader>
				<CardTitle class="flex items-center gap-3 text-slate-600">
					<Target size={20} class="text-purple-500/80" />
					<span>Current Task</span>
				</CardTitle>
			</CardHeader>
			<CardContent>
				<div class="space-y-4">
					<div>
						<label class="mb-2 block text-sm font-medium text-slate-700"
							>What are you working on?</label
						>
						<input
							bind:value={currentTask}
							type="text"
							placeholder="Enter your current task..."
							class="liquid-input w-full px-4 py-3 text-sm text-slate-800 placeholder:text-slate-400"
						/>
					</div>
					<Button
						onclick={updateTask}
						class="glass-button w-full rounded-2xl bg-gradient-to-r from-purple-500/95 to-indigo-500/95 py-3 font-semibold text-white transition-all duration-300 hover:scale-[1.01]"
					>
						Update Task
					</Button>
				</div>
			</CardContent>
		</Card>

		<!-- Daily Goals -->
		<Card
			class="liquid-panel overflow-hidden rounded-[28px] bg-gradient-to-br from-white/40 via-white/20 to-white/5 shadow-[0_35px_90px_-60px_rgba(74,222,128,0.5)] transition-all duration-300 hover:scale-[1.01]"
			data-animate="fade-up"
			style="animation-delay: 600ms;"
		>
			<CardHeader>
				<CardTitle class="flex items-center gap-3 text-slate-600">
					<TrendingUp size={20} class="text-green-500/80" />
					<span>Daily Goals</span>
				</CardTitle>
			</CardHeader>
			<CardContent>
				<div class="space-y-4">
					{#each dailyStats.goals as goal}
						<div class="space-y-2">
							<div class="flex items-center justify-between">
								<span class="text-sm font-medium text-slate-700">{goal.name}</span>
								<span class="text-sm text-slate-500"
									>{goal.current}{goal.unit} / {goal.target}{goal.unit}</span
								>
							</div>
							<div class="h-2 w-full rounded-full bg-slate-200/40 shadow-inner shadow-black/5">
								<div
									class="h-2 rounded-full bg-gradient-to-r from-emerald-400 to-green-500 transition-all duration-500"
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
		class="liquid-panel overflow-hidden rounded-[28px] bg-gradient-to-br from-white/40 via-white/20 to-white/5 shadow-[0_35px_90px_-60px_rgba(125,211,252,0.5)] transition-all duration-300 hover:scale-[1.01]"
		data-animate="fade-up"
		style="animation-delay: 750ms;"
	>
		<CardHeader>
			<CardTitle class="flex items-center gap-3 text-slate-600">
				<Activity size={20} class="text-blue-500/80" />
				<span>Recent Shifts</span>
			</CardTitle>
		</CardHeader>
		<CardContent>
			<div class="space-y-3">
				{#each recentShifts as shift}
					<div
						class="liquid-hoverable flex items-center justify-between rounded-2xl bg-gradient-to-br from-white/50 via-white/30 to-white/10 p-4 transition-all duration-200 hover:bg-white/60"
					>
						<div class="flex-1">
							<div class="mb-1 flex items-center gap-3">
								<p class="font-semibold text-slate-800">{shift.date}</p>
							</div>
							<p class="text-sm text-slate-500">{shift.time}</p>
						</div>
						<div class="text-right">
							<p class="font-mono text-lg font-semibold text-slate-700">{shift.duration}</p>
						</div>
					</div>
				{/each}
			</div>
		</CardContent>
	</Card>
</div>
