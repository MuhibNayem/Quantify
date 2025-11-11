<script lang="ts">
	import { onMount } from 'svelte';
	import { usersApi } from '$lib/api/resources';
	import type { UserSummary } from '$lib/types';
	import { toast } from 'svelte-sonner';
	import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '$lib/components/ui/card';
	import { Input } from '$lib/components/ui/input';
	import { Button } from '$lib/components/ui/button';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import DetailsModal from '$lib/components/DetailsModal.svelte';
	import type { DetailBuilderContext, DetailSection } from '$lib/components/DetailsModal.svelte';
import { UserCheck, Shield, ClipboardList, CheckCircle2 } from 'lucide-svelte';

	const tabFilters = [
		{ value: 'all', label: 'All users' },
		{ value: 'approved', label: 'Approved' },
		{ value: 'pending', label: 'Pending' }
	] as const;

	type TabValue = (typeof tabFilters)[number]['value'];

	let activeTab = $state<TabValue>('all');
	let searchTerm = $state('');
	let users = $state<UserSummary[]>([]);
	let listLoading = $state(false);

	let selectedUser = $state<UserSummary | null>(null);
	const form = $state({ username: '', password: '', role: '' });
	let detailModalOpen = $state(false);
	let detailResourceId = $state<number | null>(null);
	let detailModalSubtitle = $state<string | null>(null);

	const userStatusBadge = (isActive: boolean) =>
		isActive ? { text: 'Approved', variant: 'success' as const } : { text: 'Pending', variant: 'warning' as const };

	const buildUserSections = ({ data }: DetailBuilderContext): DetailSection[] => {
		const user = data as UserSummary;
		return [
			{
				type: 'summary',
				cards: [
					{
						title: 'Status',
						value: user.IsActive ? 'Approved' : 'Pending',
						hint: user.IsActive ? 'Active access' : 'Awaiting approval',
						icon: CheckCircle2,
						accent: user.IsActive ? 'emerald' : 'amber',
					},
					{
						title: 'Role',
						value: user.Role,
						hint: 'Current privilege',
						icon: Shield,
						accent: 'sky',
					},
					{
						title: 'User ID',
						value: `#${user.ID}`,
						hint: 'Primary identifier',
						icon: ClipboardList,
						accent: 'slate',
					},
				],
			},
			{
				type: 'description',
				title: 'Account Profile',
				items: [
					{ label: 'Username', value: user.Username },
					{ label: 'Role', value: user.Role },
					{ label: 'Status', value: user.IsActive ? 'Approved' : 'Pending', badge: userStatusBadge(user.IsActive) },
					{ label: 'User ID', value: `#${user.ID}` },
				],
			},
		];
	};

	const applyFormFromUser = (user: UserSummary | null) => {
		if (!user) {
			form.username = '';
			form.password = '';
			form.role = '';
			return;
		}
		form.username = user.Username;
		form.password = '';
		form.role = user.Role;
	};

	const loadUsers = async () => {
		listLoading = true;
		try {
			const params: Record<string, unknown> = {};
			if (activeTab !== 'all') params.status = activeTab;
			if (searchTerm.trim()) params.q = searchTerm.trim();
			const data = await usersApi.list(params);
			users = data;
			if (data.length === 0) {
				selectedUser = null;
				applyFormFromUser(null);
			} else if (!selectedUser || !data.find((u) => u.ID === selectedUser?.ID)) {
				selectedUser = data[0];
				applyFormFromUser(data[0]);
			}
		} catch (error: any) {
			const errorMessage = error.response?.data?.error || 'Unable to load users';
			toast.error('Failed to Load Users', { description: errorMessage });
		} finally {
			listLoading = false;
		}
	};

	$effect(() => {
		activeTab;
		loadUsers();
	});

	const handleSearch = () => {
		loadUsers();
	};

	const selectUser = (user: UserSummary) => {
		selectedUser = user;
		applyFormFromUser(user);
	};

	const openUserDetails = (user: UserSummary) => {
		selectUser(user);
		detailResourceId = user.ID;
		detailModalSubtitle = user.Username;
		detailModalOpen = true;
	};

	const updateUser = async () => {
		if (!selectedUser) return;
		try {
			await usersApi.update(selectedUser.ID, {
				username: form.username,
				password: form.password || undefined,
				role: form.role
			});
			toast.success('User updated');
			await loadUsers();
		} catch (error: any) {
			const errorMessage = error.response?.data?.error || 'Unable to update user';
			toast.error('Failed to Update User', { description: errorMessage });
		}
	};

	const approveUser = async (userId?: number) => {
		const targetId = userId ?? selectedUser?.ID;
		if (!targetId) return;
		try {
			await usersApi.approve(targetId);
			toast.success('User approved');
			await loadUsers();
		} catch (error: any) {
			const errorMessage = error.response?.data?.error || 'Unable to approve user';
			toast.error('Failed to Approve User', { description: errorMessage });
		}
	};

	const deleteUser = async (userId?: number, username?: string) => {
		const targetId = userId ?? selectedUser?.ID;
		if (!targetId) return;

		toast.confirm(`Are you sure you want to delete ${username ?? selectedUser?.Username}?`, {
			onConfirm: async () => {
				try {
					await usersApi.remove(targetId);
					toast.success('User removed');
					if (selectedUser?.ID === targetId) {
						selectedUser = null;
						applyFormFromUser(null);
					}
					await loadUsers();
				} catch (error: any) {
					const errorMessage = error.response?.data?.error || 'Unable to delete user';
					toast.error('Failed to Delete User', { description: errorMessage });
				}
			}
		});
	};

	// --- Parallax Hero Motion ---
	onMount(() => {
		const hero = document.querySelector('.parallax-hero') as HTMLElement | null;
		if (!hero) return;
		const handleScroll = () => {
			const y = window.scrollY / 8;
			hero.style.transform = `translateY(${y}px)`;
		};
		window.addEventListener('scroll', handleScroll, { passive: true });
		return () => window.removeEventListener('scroll', handleScroll);
	});
</script>

<!-- HERO -->
<section class="relative w-full overflow-hidden bg-gradient-to-r from-teal-50 via-sky-50 to-indigo-100 animate-gradientShift py-16 sm:py-20 px-6">
	<div class="absolute inset-0 bg-white/40 backdrop-blur-[2px]"></div>

	<!-- floating blobs -->
	<div class="pointer-events-none absolute -top-24 -right-20 h-56 w-56 rounded-full bg-gradient-to-br from-sky-300/50 to-teal-300/50 blur-3xl animate-floatSlow"></div>
	<div class="pointer-events-none absolute -bottom-24 -left-20 h-56 w-56 rounded-full bg-gradient-to-br from-indigo-300/50 to-sky-300/50 blur-3xl animate-floatSlow delay-500"></div>

	<!-- hero content -->
	<div class="relative z-10 max-w-5xl mx-auto flex flex-col items-center text-center gap-4 parallax-hero will-change-transform">
		<div class="p-4 sm:p-5 bg-gradient-to-br from-sky-500 to-indigo-600 rounded-2xl shadow-xl animate-pulseGlow">
			<UserCheck class="h-8 w-8 sm:h-9 sm:w-9 text-white" />
		</div>
		<h1 class="text-3xl sm:text-5xl font-extrabold tracking-tight bg-gradient-to-r from-sky-700 via-indigo-700 to-teal-700 bg-clip-text text-transparent animate-fadeUp">
			User Access Management
		</h1>
		<p class="max-w-2xl text-slate-700 animate-fadeUp delay-150">
			Approve, edit, or revoke workspace access — with live filters and secure updates.
		</p>
		<div class="mt-2 flex flex-wrap items-center justify-center gap-2 animate-fadeUp delay-200">
			<span class="px-3 py-1.5 text-xs sm:text-sm rounded-full border border-sky-200 bg-white/70 text-sky-700 shadow-sm">Role-based control</span>
			<span class="px-3 py-1.5 text-xs sm:text-sm rounded-full border border-indigo-200 bg-white/70 text-indigo-700 shadow-sm">Status filters</span>
			<span class="px-3 py-1.5 text-xs sm:text-sm rounded-full border border-teal-200 bg-white/70 text-teal-700 shadow-sm">Inline edits</span>
		</div>
	</div>
</section>

<DetailsModal
	bind:open={detailModalOpen}
	resourceId={detailResourceId}
	endpoint="/users"
	title="User Details"
	subtitle={detailModalSubtitle}
	buildSections={buildUserSections}
/>

<!-- MAIN CONTENT -->
<section class="max-w-7xl mx-auto py-12 sm:py-14 px-4 sm:px-6 bg-white space-y-10">
	<!-- Search & Filter -->
	<Card class="rounded-2xl border-0 shadow-lg hover:shadow-xl transition-all duration-300 hover:scale-[1.01] bg-gradient-to-br from-sky-50 to-indigo-100">
		<CardHeader class="space-y-4 bg-white/80 backdrop-blur rounded-t-2xl border-b border-white/60 px-6 py-5">
			<div class="flex flex-col sm:flex-row gap-3">
				<Input
					class="flex-1 rounded-xl border-sky-200 bg-white/90 focus:ring-2 focus:ring-sky-400"
					placeholder="Search by username or ID"
					value={searchTerm}
					oninput={(event) => (searchTerm = event.currentTarget.value)}
					onkeydown={(event) => event.key === 'Enter' && handleSearch()}
				/>
				<Button
					class="bg-gradient-to-r from-sky-500 to-indigo-600 hover:from-sky-600 hover:to-indigo-700 text-white rounded-xl shadow-md hover:shadow-lg transition-all"
					onclick={handleSearch}
				>
					Search
				</Button>
			</div>
			<div class="flex flex-wrap gap-2">
				{#each tabFilters as tab}
					<Button
						variant={tab.value === activeTab ? 'default' : 'secondary'}
						class={`flex-1 rounded-xl ${tab.value === activeTab ? 'bg-gradient-to-r from-sky-500 to-indigo-600 text-white' : 'bg-white/70 border border-sky-200 text-sky-700 hover:bg-sky-50'}`}
						onclick={() => (activeTab = tab.value)}
					>
						{tab.label}
					</Button>
				{/each}
			</div>
		</CardHeader>

		<CardContent class="p-0 overflow-hidden rounded-b-2xl">
			<div class="overflow-x-auto">
				<table class="min-w-full divide-y divide-sky-200 text-sm">
					<thead class="bg-gradient-to-r from-sky-100 to-indigo-100 text-left text-xs uppercase tracking-wide text-slate-600">
						<tr>
							<th class="px-4 py-3 font-medium">ID</th>
							<th class="px-4 py-3 font-medium">Username</th>
							<th class="px-4 py-3 font-medium">Role</th>
							<th class="px-4 py-3 font-medium">Status</th>
							<th class="px-4 py-3 font-medium text-right">Actions</th>
						</tr>
					</thead>
					<tbody class="[&>tr:nth-child(even)]:bg-white/70 [&>tr:nth-child(odd)]:bg-white/50">
						{#if listLoading}
							{#each Array(4) as _, i}
								<tr><td colspan="5" class="px-4 py-3"><Skeleton class="h-6 w-full bg-white/60" /></td></tr>
							{/each}
						{:else if users.length === 0}
							<tr>
								<td colspan="5" class="px-4 py-6 text-center text-sm text-slate-600">
									No users match this filter
								</td>
							</tr>
						{:else}
							{#each users as item}
								<tr
									class={selectedUser?.ID === item.ID ? 'bg-sky-50 border-l-4 border-sky-400' : 'hover:bg-white/90 cursor-pointer'}
									on:click={() => openUserDetails(item)}
								>
									<td class="px-4 py-3 font-mono text-xs text-slate-700">#{item.ID}</td>
									<td class="px-4 py-3 text-slate-800">{item.Username}</td>
									<td class="px-4 py-3 text-slate-700">{item.Role}</td>
									<td class="px-4 py-3">
										<span class={`rounded-full px-2.5 py-0.5 text-xs capitalize shadow-sm ${
											item.IsActive
												? 'bg-sky-100 text-sky-700 border border-sky-200'
												: 'bg-amber-100 text-amber-800 border border-amber-200'
										}`}>
											{item.IsActive ? 'Approved' : 'Pending'}
										</span>
									</td>
									<td class="px-4 py-3 text-right space-x-1">
										<Button
											size="sm"
											variant="ghost"
											disabled={item.IsActive}
											class="text-sky-700 hover:bg-sky-100 rounded-lg px-3 py-1.5"
											on:click={(event) => {
												event.stopPropagation();
												approveUser(item.ID);
											}}
										>
											Approve
										</Button>
										<Button
											size="sm"
											variant="ghost"
											class="text-rose-700 hover:bg-rose-100 rounded-lg px-3 py-1.5"
											on:click={(event) => {
												event.stopPropagation();
												deleteUser(item.ID, item.Username);
											}}
										>
											Delete
										</Button>
									</td>
								</tr>
							{/each}
						{/if}
					</tbody>
				</table>
			</div>
		</CardContent>
	</Card>

	<!-- User Details -->
	<Card class="rounded-2xl border-0 shadow-lg hover:shadow-xl transition-all duration-300 hover:scale-[1.01] bg-gradient-to-br from-teal-50 to-sky-100">
		<CardHeader class="bg-white/80 backdrop-blur rounded-t-2xl border-b border-white/60 px-6 py-5">
			<CardTitle class="text-slate-800">User Details</CardTitle>
			<CardDescription class="text-slate-600">Update role or credentials for the selected user</CardDescription>
		</CardHeader>
		{#if !selectedUser}
			<CardContent class="p-6">
				<p class="text-sm text-slate-600">Select a user from the table to edit access.</p>
			</CardContent>
		{:else}
			<CardContent class="space-y-3 p-6">
				<p class="text-xs text-slate-500">Editing #{selectedUser.ID}</p>
				<Input class="rounded-xl border-sky-200 bg-white/90 focus:ring-2 focus:ring-sky-400" placeholder="Username" bind:value={form.username} />
				<Input type="password" class="rounded-xl border-sky-200 bg-white/90 focus:ring-2 focus:ring-sky-400" placeholder="Reset password (optional)" bind:value={form.password} />
				<select class="w-full rounded-xl border border-sky-200 bg-white/90 px-3 py-2.5 text-sm focus:ring-2 focus:ring-sky-400" bind:value={form.role}>
					<option value="Admin">Admin</option>
					<option value="Manager">Manager</option>
					<option value="Staff">Staff</option>
				</select>
				<div class="flex flex-col sm:flex-row gap-3 pt-1">
					<Button class="flex-1 bg-gradient-to-r from-sky-500 to-indigo-600 hover:from-sky-600 hover:to-indigo-700 text-white font-semibold rounded-xl shadow-md hover:shadow-lg hover:scale-105 transition-all" onclick={updateUser}>
						Save changes
					</Button>
					<Button class="flex-1 border border-sky-200 text-sky-700 hover:bg-sky-50 rounded-xl" onclick={() => approveUser()} disabled={selectedUser.IsActive}>
						Approve
					</Button>
					<Button class="flex-1 border border-rose-200 text-rose-700 hover:bg-rose-50 rounded-xl" onclick={() => deleteUser()}>
						Delete
					</Button>
				</div>
			</CardContent>
		{/if}
	</Card>
</section>

<style lang="postcss">
	@keyframes gradientShift {
		0% { background-position: 0% 50%; }
		50% { background-position: 100% 50%; }
		100% { background-position: 0% 50%; }
	}
	.animate-gradientShift {
		background-size: 200% 200%;
		animation: gradientShift 20s ease-in-out infinite;
	}

	@keyframes pulseGlow {
		0%, 100% { transform: scale(1); box-shadow: 0 0 15px rgba(59, 130, 246, 0.3); }
		50% { transform: scale(1.08); box-shadow: 0 0 25px rgba(59, 130, 246, 0.5); }
	}
	.animate-pulseGlow { animation: pulseGlow 8s ease-in-out infinite; }

	@keyframes fadeUp {
		from { opacity: 0; transform: translateY(20px); }
		to { opacity: 1; transform: translateY(0); }
	}
	.animate-fadeUp { animation: fadeUp 1.3s ease forwards; }
	.animate-fadeUp.delay-150 { animation-delay: 150ms; }
	.animate-fadeUp.delay-200 { animation-delay: 200ms; }

	@keyframes floatSlow {
		0%, 100% { transform: translateY(0px) scale(1); }
		50% { transform: translateY(-10px) scale(1.03); }
	}
	.animate-floatSlow { animation: floatSlow 10s ease-in-out infinite; }

	* {
		transition-property: color, background-color, border-color, box-shadow, transform, filter, backdrop-filter;
		transition-timing-function: cubic-bezier(0.4, 0, 0.2, 1);
		transition-duration: 300ms;
	}
</style>
