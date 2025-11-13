<script lang="ts">
    import { Button } from '$lib/components/ui/button';
    import { Card, CardContent, CardHeader, CardTitle } from '$lib/components/ui/card';
    import { Clock, LogIn, LogOut, Timer, Coffee, Target, TrendingUp, Calendar, Activity } from 'lucide-svelte';
    import { onMount, onDestroy } from 'svelte';
    import { toast } from 'svelte-sonner';

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
    let currentProject = '';
    let productivityScore = 87;
    let weeklyTarget = 40;
    let todayTarget = 8;

    // Mock data
    let todayHours = '4h 32m';
    let weeklyHours = '32h 15m';
    let weeklyProgress = 80.4; // percentage

    let recentShifts = [
        { date: 'Monday, Nov 10', time: '8:01 AM - 5:03 PM', duration: '9h 2m', productivity: 92 },
        { date: 'Friday, Nov 7', time: '8:05 AM - 4:55 PM', duration: '8h 50m', productivity: 88 },
        { date: 'Thursday, Nov 6', time: '8:15 AM - 5:15 PM', duration: '9h 0m', productivity: 85 },
        { date: 'Wednesday, Nov 5', time: '7:55 AM - 4:45 PM', duration: '8h 50m', productivity: 90 }
    ];

    let dailyStats = {
        goals: [
            { name: 'Focus Time', current: 6, target: 7, unit: 'h' },
            { name: 'Tasks Completed', current: 12, target: 15, unit: '' },
            { name: 'Breaks Taken', current: 3, target: 4, unit: '' }
        ]
    };

    const handleClockToggle = () => {
        if (!clockedIn) {
            clockedIn = true;
            shiftStartTime = new Date();
            startTimer();
            toast.success('Clocked In Successfully', {
                description: 'Your shift has started. Have a productive day!'
            });
        } else {
            clockedIn = false;
            stopTimer();
            if (onBreak) {
                handleBreakToggle();
            }
            toast.info('Clocked Out', {
                description: `Great work today! You completed ${elapsedTime} of focused time.`
            });
            shiftStartTime = null;
        }
    };

    const handleBreakToggle = () => {
        if (!onBreak) {
            onBreak = true;
            breakStartTime = new Date();
            startBreakTimer();
            toast.info('Break Started', {
                description: 'Take a well-deserved break! Your timer is paused.'
            });
        } else {
            onBreak = false;
            stopBreakTimer();
            toast.success('Break Ended', {
                description: 'Welcome back! Ready to continue your productive day?'
            });
            breakStartTime = null;
        }
    };

    const startTimer = () => {
        timerInterval = setInterval(() => {
            if (shiftStartTime && !onBreak) {
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
        }
    };

    onMount(() => {
        // Initialize with mock data
    });

    onDestroy(() => {
        stopTimer();
        stopBreakTimer();
    });
</script>

<div class="min-h-screen bg-gradient-to-br from-slate-50 via-blue-50 to-purple-50 p-4 md:p-8">
    <div class="max-w-6xl mx-auto space-y-8">
        <!-- Header -->
        <div class="text-center mb-6">
            <h1 class="text-3xl md:text-4xl font-bold bg-gradient-to-r from-blue-600 via-purple-600 to-indigo-600 bg-clip-text text-transparent mb-2">
                My Time Tracker
            </h1>
            <p class="text-slate-600">Track your productivity and achieve your daily goals</p>
        </div>

        <!-- Main Clock Card -->
        <Card
            class="rounded-3xl shadow-xl hover:shadow-2xl transition-all duration-500 hover:scale-[1.02] overflow-hidden border-0 bg-gradient-to-br from-blue-500 via-purple-500 to-indigo-600 text-white"
            data-animate="fade-up"
        >
            <CardHeader class="pb-4">
                <CardTitle class="flex items-center justify-between text-white/90">
                    <span class="text-xl font-semibold">My Status</span>
                    <div class="flex items-center gap-4">
                        <span class="text-sm font-medium flex items-center gap-2">
                            <div
                                class="w-4 h-4 rounded-full {clockedIn
                                    ? 'bg-green-300 animate-pulse'
                                    : 'bg-slate-300'}"
                            ></div>
                            {clockedIn ? 'Clocked In' : 'Clocked Out'}
                        </span>
                        {#if onBreak}
                            <span class="text-sm font-medium flex items-center gap-2 bg-yellow-400/20 px-3 py-1 rounded-full">
                                <Coffee size={16} />
                                On Break
                            </span>
                        {/if}
                    </div>
                </CardTitle>
            </CardHeader>
            <CardContent class="text-center p-8">
                <div class="text-6xl md:text-7xl font-bold font-mono tracking-tighter mb-6 bg-gradient-to-r from-white to-blue-100 bg-clip-text text-transparent">
                    {elapsedTime}
                </div>
                
                {#if onBreak}
                    <div class="mb-6">
                        <div class="text-lg font-medium text-yellow-200 mb-2">Break Time</div>
                        <div class="text-2xl font-mono text-white">{breakTime}</div>
                    </div>
                {/if}

                <div class="flex flex-col md:flex-row gap-4 mb-6">
                    <div class="bg-white/10 backdrop-blur-sm rounded-2xl p-4 flex-1">
                        <div class="text-lg font-bold">{productivityScore}%</div>
                        <div class="text-sm opacity-80">Productivity</div>
                    </div>
                    <div class="bg-white/10 backdrop-blur-sm rounded-2xl p-4 flex-1">
                        <div class="text-lg font-bold">{todayHours}</div>
                        <div class="text-sm opacity-80">Today's Total</div>
                    </div>
                </div>

                <div class="flex flex-col md:flex-row gap-4">
                    <Button
                        on:click={handleClockToggle}
                        class="flex-1 text-lg py-4 rounded-2xl shadow-lg hover:shadow-xl transition-all duration-300 hover:scale-105 {clockedIn
                            ? 'bg-gradient-to-r from-rose-500 to-pink-600 hover:from-rose-600 hover:to-pink-700'
                            : 'bg-gradient-to-r from-emerald-500 to-green-600 hover:from-emerald-600 hover:to-green-700'} text-white font-semibold"
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
                            class="flex-1 text-lg py-4 rounded-2xl shadow-lg hover:shadow-xl transition-all duration-300 hover:scale-105 {onBreak
                                ? 'bg-gradient-to-r from-orange-500 to-amber-600 hover:from-orange-600 hover:to-amber-700'
                                : 'bg-gradient-to-r from-blue-500 to-cyan-600 hover:from-blue-600 hover:to-cyan-700'} text-white font-semibold"
                        >
                            <Coffee class="mr-2" size={20} />
                            {onBreak ? 'End Break' : 'Start Break'}
                        </Button>
                    {/if}
                </div>
            </CardContent>
        </Card>

        <!-- Hours Overview -->
        <div class="grid grid-cols-1 md:grid-cols-2 gap-8">
            <Card
                class="rounded-2xl shadow-lg hover:shadow-xl transition-all duration-300 hover:scale-[1.01] overflow-hidden border-0 bg-gradient-to-br from-emerald-400 to-green-500 text-white"
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
                    <p class="text-4xl font-bold text-white mb-2">{todayHours}</p>
                    <div class="w-full bg-white/20 rounded-full h-3 mb-2">
                        <div 
                            class="bg-white rounded-full h-3 transition-all duration-500" 
                            style="width: {(parseFloat(todayHours) / todayTarget) * 100}%"
                        ></div>
                    </div>
                    <p class="text-sm text-white/80">out of {todayTarget} hours target</p>
                </CardContent>
            </Card>

            <Card
                class="rounded-2xl shadow-lg hover:shadow-xl transition-all duration-300 hover:scale-[1.01] overflow-hidden border-0 bg-gradient-to-br from-blue-400 to-indigo-500 text-white"
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
                    <p class="text-4xl font-bold text-white mb-2">{weeklyHours}</p>
                    <div class="w-full bg-white/20 rounded-full h-3 mb-2">
                        <div 
                            class="bg-white rounded-full h-3 transition-all duration-500" 
                            style="width: {weeklyProgress}%"
                        ></div>
                    </div>
                    <p class="text-sm text-white/80">out of {weeklyTarget} hours target</p>
                </CardContent>
            </Card>
        </div>

        <!-- Current Task & Daily Goals -->
        <div class="grid grid-cols-1 lg:grid-cols-2 gap-8">
            <!-- Current Task -->
            <Card
                class="rounded-2xl shadow-lg hover:shadow-xl transition-all duration-300 hover:scale-[1.01] overflow-hidden border-0 bg-gradient-to-br from-white to-slate-50"
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
                            <label class="block text-sm font-medium text-slate-700 mb-2">What are you working on?</label>
                            <input
                                bind:value={currentTask}
                                type="text"
                                placeholder="Enter your current task..."
                                class="w-full px-4 py-3 rounded-xl border border-slate-200 focus:border-purple-500 focus:ring-2 focus:ring-purple-200 transition-all duration-200"
                            />
                        </div>
                        <div>
                            <label class="block text-sm font-medium text-slate-700 mb-2">Project</label>
                            <input
                                bind:value={currentProject}
                                type="text"
                                placeholder="Project name..."
                                class="w-full px-4 py-3 rounded-xl border border-slate-200 focus:border-purple-500 focus:ring-2 focus:ring-purple-200 transition-all duration-200"
                            />
                        </div>
                        <Button
                            on:click={updateTask}
                            class="w-full bg-gradient-to-r from-purple-500 to-indigo-600 hover:from-purple-600 hover:to-indigo-700 text-white font-semibold py-3 rounded-xl transition-all duration-300 hover:scale-105"
                        >
                            Update Task
                        </Button>
                    </div>
                </CardContent>
            </Card>

            <!-- Daily Goals -->
            <Card
                class="rounded-2xl shadow-lg hover:shadow-xl transition-all duration-300 hover:scale-[1.01] overflow-hidden border-0 bg-gradient-to-br from-white to-slate-50"
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
                                <div class="flex justify-between items-center">
                                    <span class="text-sm font-medium text-slate-700">{goal.name}</span>
                                    <span class="text-sm text-slate-500">{goal.current}{goal.unit} / {goal.target}{goal.unit}</span>
                                </div>
                                <div class="w-full bg-slate-200 rounded-full h-2">
                                    <div 
                                        class="bg-gradient-to-r from-green-400 to-green-600 rounded-full h-2 transition-all duration-500" 
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
            class="rounded-2xl shadow-lg hover:shadow-xl transition-all duration-300 hover:scale-[1.01] overflow-hidden border-0 bg-gradient-to-br from-white to-slate-50"
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
                            class="flex justify-between items-center p-4 bg-slate-50/50 rounded-xl border border-slate-200/80 hover:bg-slate-100 transition-all duration-200"
                        >
                            <div class="flex-1">
                                <div class="flex items-center gap-3 mb-1">
                                    <p class="font-semibold text-slate-800">{shift.date}</p>
                                    <span class="text-xs bg-blue-100 text-blue-700 px-2 py-1 rounded-full">
                                        {shift.productivity}% productive
                                    </span>
                                </div>
                                <p class="text-sm text-slate-500">{shift.time}</p>
                            </div>
                            <div class="text-right">
                                <p class="font-mono text-slate-700 text-lg font-semibold">{shift.duration}</p>
                                <div class="flex items-center gap-1 mt-1">
                                    {#each Array(5) as _, i}
                                        <div
                                            class="w-2 h-2 rounded-full {i < Math.floor(shift.productivity / 20)
                                                ? 'bg-green-400'
                                                : 'bg-slate-300'}"
                                        ></div>
                                    {/each}
                                </div>
                            </div>
                        </div>
                    {/each}
                </div>
            </CardContent>
        </Card>
    </div>
</div>