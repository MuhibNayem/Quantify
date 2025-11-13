<!-- ManagerDashboard.svelte -->
<script lang="ts">
    import { Button } from '$lib/components/ui/button';
    import { Card, CardContent, CardHeader, CardTitle } from '$lib/components/ui/card';
    import { 
        Clock, Users, TrendingUp, Calendar, BarChart3, PieChart, 
        Activity, MapPin, Coffee, FileText, Settings, AlertCircle,
        CheckCircle, DollarSign, Target, Download, Filter
    } from 'lucide-svelte';
    import { onMount } from 'svelte';
    import { toast } from 'svelte-sonner';

    // Team data
    let teamMembers = [
        { id: 1, name: 'Alice Johnson', role: 'Frontend Developer', status: 'active', project: 'Website Redesign', hours: '6h 15m', productivity: 92, avatar: 'AJ' },
        { id: 2, name: 'Bob Smith', role: 'Backend Developer', status: 'break', project: 'Mobile App', hours: '5h 45m', productivity: 87, avatar: 'BS' },
        { id: 3, name: 'Carol Davis', role: 'UI/UX Designer', status: 'active', project: 'Website Redesign', hours: '7h 30m', productivity: 95, avatar: 'CD' },
        { id: 4, name: 'David Wilson', role: 'Project Manager', status: 'active', project: 'API Development', hours: '8h 00m', productivity: 89, avatar: 'DW' },
        { id: 5, name: 'Emma Brown', role: 'QA Engineer', status: 'offline', project: 'Mobile App', hours: '4h 20m', productivity: 84, avatar: 'EB' }
    ];

    // Stats data
    let teamStats = {
        totalHours: '156h 45m',
        activeMembers: 4,
        avgProductivity: 89,
        weeklyTarget: 200,
        progress: 78.4
    };

    // Project data
    let projects = [
        { name: 'Website Redesign', hours: '68h 30m', progress: 75, team: 3, status: 'on-track' },
        { name: 'Mobile App', hours: '52h 15m', progress: 60, team: 2, status: 'on-track' },
        { name: 'API Development', hours: '36h 00m', progress: 90, team: 2, status: 'ahead' }
    ];

    // Attendance data
    let attendanceData = [
        { day: 'Mon', present: 8, absent: 1, late: 1 },
        { day: 'Tue', present: 9, absent: 0, late: 1 },
        { day: 'Wed', present: 7, absent: 2, late: 1 },
        { day: 'Thu', present: 8, absent: 1, late: 1 },
        { day: 'Fri', present: 9, absent: 0, late: 1 },
        { day: 'Sat', present: 3, absent: 0, late: 0 },
        { day: 'Sun', present: 1, absent: 0, late: 0 }
    ];

    // Recent activities
    let recentActivities = [
        { user: 'Alice Johnson', action: 'clocked in', time: '8:05 AM', project: 'Website Redesign' },
        { user: 'Bob Smith', action: 'started break', time: '10:30 AM', project: 'Mobile App' },
        { user: 'Carol Davis', action: 'completed task', time: '11:15 AM', project: 'Website Redesign' },
        { user: 'David Wilson', action: 'clocked in', time: '8:00 AM', project: 'API Development' }
    ];

    // Filter state
    let statusFilter = 'all';
    let projectFilter = 'all';

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

    onMount(() => {
        // Initialize with mock data
    });
</script>

<div class="min-h-screen bg-gradient-to-br from-slate-50 via-blue-50 to-purple-50 p-4 md:p-8">
    <div class="max-w-7xl mx-auto space-y-8">
        <!-- Header -->
        <div class="flex flex-col md:flex-row justify-between items-start md:items-center gap-4 mb-6">
            <div>
                <h1 class="text-3xl md:text-4xl font-bold bg-gradient-to-r from-blue-600 via-purple-600 to-indigo-600 bg-clip-text text-transparent mb-2">
                    Team Dashboard
                </h1>
                <p class="text-slate-600">Monitor your team's productivity and time tracking</p>
            </div>
            <div class="flex gap-3">
                <Button
                    on:click={handleExportReport}
                    class="bg-gradient-to-r from-blue-500 to-indigo-600 hover:from-blue-600 hover:to-indigo-700 text-white font-semibold py-2 px-4 rounded-xl transition-all duration-300 hover:scale-105 flex items-center gap-2"
                >
                    <Download size={18} />
                    Export Report
                </Button>
                <Button
                    variant="outline"
                    class="border-slate-200 hover:border-blue-300 hover:bg-blue-50 text-slate-700 font-semibold py-2 px-4 rounded-xl transition-all duration-300 flex items-center gap-2"
                >
                    <Filter size={18} />
                    Filter
                </Button>
            </div>
        </div>

        <!-- Team Stats Overview -->
        <div class="grid grid-cols-1 md:grid-cols-4 gap-6">
            <Card
                class="rounded-2xl shadow-lg hover:shadow-xl transition-all duration-300 hover:scale-[1.02] overflow-hidden border-0 bg-gradient-to-br from-blue-400 to-indigo-500 text-white"
            >
                <CardHeader class="pb-2">
                    <CardTitle class="flex items-center gap-3 text-white/90">
                        <Clock size={20} />
                        <span>Total Hours</span>
                    </CardTitle>
                </CardHeader>
                <CardContent>
                    <p class="text-3xl font-bold text-white">{teamStats.totalHours}</p>
                    <div class="w-full bg-white/20 rounded-full h-2 mt-2">
                        <div 
                            class="bg-white rounded-full h-2 transition-all duration-500" 
                            style="width: {teamStats.progress}%"
                        ></div>
                    </div>
                    <p class="text-sm text-white/80 mt-1">{teamStats.progress}% of weekly target</p>
                </CardContent>
            </Card>

            <Card
                class="rounded-2xl shadow-lg hover:shadow-xl transition-all duration-300 hover:scale-[1.02] overflow-hidden border-0 bg-gradient-to-br from-emerald-400 to-green-500 text-white"
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
                class="rounded-2xl shadow-lg hover:shadow-xl transition-all duration-300 hover:scale-[1.02] overflow-hidden border-0 bg-gradient-to-br from-purple-400 to-pink-500 text-white"
            >
                <CardHeader class="pb-2">
                    <CardTitle class="flex items-center gap-3 text-white/90">
                        <TrendingUp size={20} />
                        <span>Avg. Productivity</span>
                    </CardTitle>
                </CardHeader>
                <CardContent>
                    <p class="text-3xl font-bold text-white">{teamStats.avgProductivity}%</p>
                    <p class="text-sm text-white/80">Team average</p>
                </CardContent>
            </Card>

            <Card
                class="rounded-2xl shadow-lg hover:shadow-xl transition-all duration-300 hover:scale-[1.02] overflow-hidden border-0 bg-gradient-to-br from-amber-400 to-orange-500 text-white"
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
        <div class="grid grid-cols-1 lg:grid-cols-2 gap-8">
            <!-- Team Members -->
            <Card
                class="rounded-2xl shadow-lg hover:shadow-xl transition-all duration-300 hover:scale-[1.01] overflow-hidden border-0 bg-gradient-to-br from-white to-slate-50"
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
                                class="flex items-center justify-between p-4 bg-slate-50/50 rounded-xl border border-slate-200/80 hover:bg-slate-100 transition-all duration-200"
                            >
                                <div class="flex items-center gap-3">
                                    <div class="w-10 h-10 rounded-full bg-gradient-to-br from-blue-400 to-indigo-500 flex items-center justify-center text-white font-semibold">
                                        {member.avatar}
                                    </div>
                                    <div>
                                        <p class="font-semibold text-slate-800">{member.name}</p>
                                        <p class="text-sm text-slate-500">{member.role}</p>
                                    </div>
                                </div>
                                <div class="text-right">
                                    <div class="flex items-center gap-2 mb-1">
                                        <div
                                            class="w-2 h-2 rounded-full {member.status === 'active'
                                                ? 'bg-green-400 animate-pulse'
                                                : member.status === 'break'
                                                ? 'bg-yellow-400'
                                                : 'bg-slate-400'}"
                                        ></div>
                                        <span class="text-xs text-slate-500">
                                            {member.status === 'active' ? 'Working' : member.status === 'break' ? 'On Break' : 'Offline'}
                                        </span>
                                    </div>
                                    <p class="font-mono text-slate-700 font-medium">{member.hours}</p>
                                    <div class="flex items-center gap-1 mt-1">
                                        {#each Array(5) as _, i}
                                            <div
                                                class="w-1.5 h-1.5 rounded-full {i < Math.floor(member.productivity / 20)
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

            <!-- Projects -->
            <Card
                class="rounded-2xl shadow-lg hover:shadow-xl transition-all duration-300 hover:scale-[1.01] overflow-hidden border-0 bg-gradient-to-br from-white to-slate-50"
            >
                <CardHeader>
                    <CardTitle class="flex items-center gap-3 text-slate-800">
                        <BarChart3 size={20} class="text-purple-500" />
                        <span>Projects</span>
                    </CardTitle>
                </CardHeader>
                <CardContent>
                    <div class="space-y-4">
                        {#each projects as project}
                            <div class="space-y-2">
                                <div class="flex justify-between items-center">
                                    <div>
                                        <p class="font-semibold text-slate-800">{project.name}</p>
                                        <div class="flex items-center gap-2 mt-1">
                                            <span class="text-xs text-slate-500">{project.team} team members</span>
                                            <span class="text-xs px-2 py-1 rounded-full {project.status === 'on-track'
                                                ? 'bg-blue-100 text-blue-700'
                                                : 'bg-green-100 text-green-700'}">
                                                {project.status === 'on-track' ? 'On Track' : 'Ahead'}
                                            </span>
                                        </div>
                                    </div>
                                    <p class="font-mono text-slate-700 font-medium">{project.hours}</p>
                                </div>
                                <div class="w-full bg-slate-200 rounded-full h-2">
                                    <div 
                                        class="bg-gradient-to-r from-purple-400 to-purple-600 rounded-full h-2 transition-all duration-500" 
                                        style="width: {project.progress}%"
                                    ></div>
                                </div>
                                <div class="flex justify-between text-xs text-slate-500">
                                    <span>Progress</span>
                                    <span>{project.progress}%</span>
                                </div>
                            </div>
                        {/each}
                    </div>
                </CardContent>
            </Card>
        </div>

        <!-- Attendance and Recent Activities -->
        <div class="grid grid-cols-1 lg:grid-cols-2 gap-8">
            <!-- Attendance Overview -->
            <Card
                class="rounded-2xl shadow-lg hover:shadow-xl transition-all duration-300 hover:scale-[1.01] overflow-hidden border-0 bg-gradient-to-br from-white to-slate-50"
            >
                <CardHeader>
                    <CardTitle class="flex items-center gap-3 text-slate-800">
                        <Calendar size={20} class="text-green-500" />
                        <span>Weekly Attendance</span>
                    </CardTitle>
                </CardHeader>
                <CardContent>
                    <div class="space-y-6">
                        <div class="flex justify-between items-end h-32">
                            {#each attendanceData as day}
                                <div class="flex flex-col items-center flex-1">
                                    <div class="flex items-end justify-center gap-1 h-24">
                                        <div
                                            class="w-4 bg-gradient-to-t from-green-400 to-green-600 rounded-t-lg transition-all duration-300 hover:from-green-500 hover:to-green-700"
                                            style="height: {(day.present / 10) * 60}px;"
                                        ></div>
                                        {#if day.absent > 0}
                                            <div
                                                class="w-4 bg-gradient-to-t from-rose-400 to-rose-600 rounded-t-lg transition-all duration-300 hover:from-rose-500 hover:to-rose-700"
                                                style="height: {(day.absent / 10) * 60}px;"
                                            ></div>
                                        {/if}
                                        {#if day.late > 0}
                                            <div
                                                class="w-4 bg-gradient-to-t from-amber-400 to-amber-600 rounded-t-lg transition-all duration-300 hover:from-amber-500 hover:to-amber-700"
                                                style="height: {(day.late / 10) * 60}px;"
                                            ></div>
                                        {/if}
                                    </div>
                                    <span class="text-xs text-slate-500 mt-2">{day.day}</span>
                                </div>
                            {/each}
                        </div>
                        <div class="flex justify-center gap-6">
                            <div class="flex items-center gap-2">
                                <div class="w-3 h-3 rounded-full bg-green-500"></div>
                                <span class="text-xs text-slate-600">Present</span>
                            </div>
                            <div class="flex items-center gap-2">
                                <div class="w-3 h-3 rounded-full bg-amber-500"></div>
                                <span class="text-xs text-slate-600">Late</span>
                            </div>
                            <div class="flex items-center gap-2">
                                <div class="w-3 h-3 rounded-full bg-rose-500"></div>
                                <span class="text-xs text-slate-600">Absent</span>
                            </div>
                        </div>
                    </div>
                </CardContent>
            </Card>

            <!-- Recent Activities -->
            <Card
                class="rounded-2xl shadow-lg hover:shadow-xl transition-all duration-300 hover:scale-[1.01] overflow-hidden border-0 bg-gradient-to-br from-white to-slate-50"
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
                                class="flex items-start gap-3 p-3 bg-slate-50/50 rounded-xl border border-slate-200/80 hover:bg-slate-100 transition-all duration-200"
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
                                        <span class="font-semibold">{activity.user}</span> {activity.action}
                                    </p>
                                    <div class="flex justify-between items-center mt-1">
                                        <p class="text-xs text-slate-500">{activity.project}</p>
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
            class="rounded-2xl shadow-lg hover:shadow-xl transition-all duration-300 hover:scale-[1.01] overflow-hidden border-0 bg-gradient-to-br from-white to-slate-50"
        >
            <CardHeader>
                <CardTitle class="flex items-center gap-3 text-slate-800">
                    <Settings size={20} class="text-slate-500" />
                    <span>Quick Actions</span>
                </CardTitle>
            </CardHeader>
            <CardContent>
                <div class="grid grid-cols-2 md:grid-cols-4 gap-4">
                    <Button
                        variant="outline"
                        class="h-20 flex flex-col items-center justify-center gap-2 border-slate-200 hover:border-blue-300 hover:bg-blue-50 transition-all duration-200"
                    >
                        <PieChart size={24} class="text-blue-500" />
                        <span class="text-sm font-medium">Reports</span>
                    </Button>
                    <Button
                        variant="outline"
                        class="h-20 flex flex-col items-center justify-center gap-2 border-slate-200 hover:border-green-300 hover:bg-green-50 transition-all duration-200"
                    >
                        <Calendar size={24} class="text-green-500" />
                        <span class="text-sm font-medium">Schedule</span>
                    </Button>
                    <Button
                        variant="outline"
                        class="h-20 flex flex-col items-center justify-center gap-2 border-slate-200 hover:border-purple-300 hover:bg-purple-50 transition-all duration-200"
                    >
                        <DollarSign size={24} class="text-purple-500" />
                        <span class="text-sm font-medium">Payroll</span>
                    </Button>
                    <Button
                        variant="outline"
                        class="h-20 flex flex-col items-center justify-center gap-2 border-slate-200 hover:border-amber-300 hover:bg-amber-50 transition-all duration-200"
                    >
                        <AlertCircle size={24} class="text-amber-500" />
                        <span class="text-sm font-medium">Reminders</span>
                    </Button>
                </div>
            </CardContent>
        </Card>
    </div>
</div>