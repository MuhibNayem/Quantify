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

<div class="min-h-screen bg-gradient-to-br from-slate-50 via-blue-50 to-purple-50 p-4 md:p-8">
	<div class="mx-auto max-w-6xl space-y-8">
		<!-- Header -->
		<div class="mb-6 text-center">
			<h1
				class="mb-2 bg-gradient-to-r from-blue-600 via-purple-600 to-indigo-600 bg-clip-text text-3xl font-bold text-transparent md:text-4xl"
			>
				My Time Tracker
			</h1>
			<p class="text-slate-600">Track your productivity and achieve your daily goals</p>
		</div>

		<!-- Main Clock Card -->
		<Card
			class="overflow-hidden rounded-3xl border-0 bg-gradient-to-br from-blue-500 via-purple-500 to-indigo-600 text-white shadow-xl transition-all duration-500 hover:scale-[1.02] hover:shadow-2xl"
			data-animate="fade-up"
		>
			<CardHeader class="pb-4">
				<CardTitle class="flex items-center justify-between text-white/90">
					<span class="text-xl font-semibold">My Status</span>
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
								class="flex items-center gap-2 rounded-full bg-yellow-400/20 px-3 py-1 text-sm font-medium"
							>
								<Coffee size={16} />
								On Break
							</span>
						{/if}
					</div>
				</CardTitle>
			</CardHeader>
			<CardContent class="p-8 text-center">
				<div
					class="mb-6 bg-gradient-to-r from-white to-blue-100 bg-clip-text font-mono text-6xl font-bold tracking-tighter text-transparent md:text-7xl"
				>
					{elapsedTime}
				</div>

				{#if onBreak}
					<div class="mb-6">
						<div class="mb-2 text-lg font-medium text-yellow-200">Break Time</div>
						<div class="font-mono text-2xl text-white">{breakTime}</div>
					</div>
				{/if}

				<div class="mb-6 flex flex-col gap-4 md:flex-row">
					<div class="flex-1 rounded-2xl bg-white/10 p-4 backdrop-blur-sm">
						<div class="text-lg font-bold">{todayHours}</div>
						<div class="text-sm opacity-80">Today's Total</div>
					</div>
				</div>

				<div class="flex flex-col gap-4 md:flex-row">
					<Button
						on:click={handleClockToggle}
						class="flex-1 rounded-2xl py-4 text-lg shadow-lg transition-all duration-300 hover:scale-105 hover:shadow-xl {clockedIn
							? 'bg-gradient-to-r from-rose-500 to-pink-600 hover:from-rose-600 hover:to-pink-700'
							: 'bg-gradient-to-r from-emerald-500 to-green-600 hover:from-emerald-600 hover:to-green-700'} font-semibold text-white"
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
							on:click={handleBreakToggle}
							class="flex-1 rounded-2xl py-4 text-lg shadow-lg transition-all duration-300 hover:scale-105 hover:shadow-xl {onBreak
								? 'bg-gradient-to-r from-orange-500 to-amber-600 hover:from-orange-600 hover:to-amber-700'
								: 'bg-gradient-to-r from-blue-500 to-cyan-600 hover:from-blue-600 hover:to-cyan-700'} font-semibold text-white"
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
				class="overflow-hidden rounded-2xl border-0 bg-gradient-to-br from-emerald-400 to-green-500 text-white shadow-lg transition-all duration-300 hover:scale-[1.01] hover:shadow-xl"
				data-animate="fade-up"
				style="animation-delay: 150ms;"
			>
				<CardHeader class="pb-2">
					<CardTitle class="flex items-center gap-3 text-white/90">
						<Timer size={20} />
						<span>Today's Hours</span>
					</CardTitle>
				</CardHeader>
				<CardContent>
					<p class="mb-2 text-4xl font-bold text-white">{todayHours}</p>
					<div class="mb-2 h-3 w-full rounded-full bg-white/20">
						<div
							class="h-3 rounded-full bg-white transition-all duration-500"
							style="width: {(parseFloat(todayHours) / todayTarget) * 100}%"
						></div>
					</div>
					<p class="text-sm text-white/80">out of {todayTarget} hours target</p>
				</CardContent>
			</Card>

			<Card
				class="overflow-hidden rounded-2xl border-0 bg-gradient-to-br from-blue-400 to-indigo-500 text-white shadow-lg transition-all duration-300 hover:scale-[1.01] hover:shadow-xl"
				data-animate="fade-up"
				style="animation-delay: 300ms;"
			>
				<CardHeader class="pb-2">
					<CardTitle class="flex items-center gap-3 text-white/90">
						<Calendar size={20} />
						<span>Weekly Hours</span>
					</CardTitle>
				</CardHeader>
				<CardContent>
					<p class="mb-2 text-4xl font-bold text-white">{weeklyHours}</p>
					<div class="mb-2 h-3 w-full rounded-full bg-white/20">
						<div
							class="h-3 rounded-full bg-white transition-all duration-500"
							style="width: {weeklyProgress}%"
						></div>
					</div>
					<p class="text-sm text-white/80">out of {weeklyTarget} hours target</p>
				</CardContent>
			</Card>
		</div>

		<!-- Current Task & Daily Goals -->
		<div class="grid grid-cols-1 gap-8 lg:grid-cols-2">
			<!-- Current Task -->
			<Card
				class="overflow-hidden rounded-2xl border-0 bg-gradient-to-br from-white to-slate-50 shadow-lg transition-all duration-300 hover:scale-[1.01] hover:shadow-xl"
				data-animate="fade-up"
				style="animation-delay: 450ms;"
			>
				<CardHeader>
					<CardTitle class="flex items-center gap-3 text-slate-800">
						<Target size={20} class="text-purple-500" />
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
								class="w-full rounded-xl border border-slate-200 px-4 py-3 transition-all duration-200 focus:border-purple-500 focus:ring-2 focus:ring-purple-200"
							/>
						</div>
						<Button
							on:click={updateTask}
							class="w-full rounded-xl bg-gradient-to-r from-purple-500 to-indigo-600 py-3 font-semibold text-white transition-all duration-300 hover:scale-105 hover:from-purple-600 hover:to-indigo-700"
						>
							Update Task
						</Button>
					</div>
				</CardContent>
			</Card>

			<!-- Daily Goals -->
			<Card
				class="overflow-hidden rounded-2xl border-0 bg-gradient-to-br from-white to-slate-50 shadow-lg transition-all duration-300 hover:scale-[1.01] hover:shadow-xl"
				data-animate="fade-up"
				style="animation-delay: 600ms;"
			>
				<CardHeader>
					<CardTitle class="flex items-center gap-3 text-slate-800">
						<TrendingUp size={20} class="text-green-500" />
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
								<div class="h-2 w-full rounded-full bg-slate-200">
									<div
										class="h-2 rounded-full bg-gradient-to-r from-green-400 to-green-600 transition-all duration-500"
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
			class="overflow-hidden rounded-2xl border-0 bg-gradient-to-br from-white to-slate-50 shadow-lg transition-all duration-300 hover:scale-[1.01] hover:shadow-xl"
			data-animate="fade-up"
			style="animation-delay: 750ms;"
		>
			<CardHeader>
				<CardTitle class="flex items-center gap-3 text-slate-800">
					<Activity size={20} class="text-blue-500" />
					<span>Recent Shifts</span>
				</CardTitle>
			</CardHeader>
			<CardContent>
				<div class="space-y-3">
					{#each recentShifts as shift}
						<div
							class="flex items-center justify-between rounded-xl border border-slate-200/80 bg-slate-50/50 p-4 transition-all duration-200 hover:bg-slate-100"
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
</div>
