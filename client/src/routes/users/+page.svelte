<script lang="ts">
	import { onMount } from 'svelte';
	import { usersApi } from '$lib/api/resources';
	import { rolesApi } from '$lib/api/roles';
	import type { Role, UserSummary } from '$lib/types';
	import { toast } from 'svelte-sonner';
	import {
		Card,
		CardContent,
		CardDescription,
		CardHeader,
		CardTitle
	} from '$lib/components/ui/card';
	import { Input } from '$lib/components/ui/input';
	import { Button } from '$lib/components/ui/button';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import DetailsModal from '$lib/components/DetailsModal.svelte';
	import type { DetailBuilderContext, DetailSection } from '$lib/components/DetailsModal.svelte';
	import { UserCheck, Shield, ClipboardList, CheckCircle2, Pencil } from 'lucide-svelte';
	import { auth } from '$lib/stores/auth';
	import { t } from '$lib/i18n';
	import { goto } from '$app/navigation';

	$effect(() => {
		if (!auth.hasPermission('users.view')) {
			toast.error($t('common.access_denied'), { description: $t('common.no_permission_settings') });
			goto('/');
		}
	});

	const tabFilters = [
		{ value: 'all', label: 'users.filters.all' },
		{ value: 'approved', label: 'users.filters.approved' },
		{ value: 'pending', label: 'users.filters.pending' }
	] as const;

	type TabValue = (typeof tabFilters)[number]['value'];

	let activeTab = $state<TabValue>('all');
	let searchTerm = $state('');
	let users = $state<UserSummary[]>([]);
	let listLoading = $state(false);

	let selectedUser = $state<UserSummary | null>(null);
	const form = $state({
		username: '',
		password: '',
		role: '',
		firstName: '',
		lastName: '',
		email: '',
		phoneNumber: '',
		address: ''
	});
	let detailModalOpen = $state(false);
	let detailResourceId = $state<number | null>(null);
	let detailModalSubtitle = $state<string | null>(null);
	let roles = $state<Role[]>([]);

	const userStatusBadge = (isActive: boolean) =>
		isActive
			? { text: 'users.status.approved', variant: 'success' as const }
			: { text: 'users.status.pending', variant: 'warning' as const };

	const buildUserSections = ({ data }: DetailBuilderContext): DetailSection[] => {
		const user = data as unknown as UserSummary;
		return [
			{
				type: 'summary',
				cards: [
					{
						title: 'Status',
						value: user.IsActive ? 'Approved' : 'Pending',
						hint: user.IsActive ? 'Active access' : 'Awaiting approval',
						icon: CheckCircle2,
						accent: user.IsActive ? 'emerald' : 'amber'
					},
					{
						title: 'Role',
						value: user.Role.Name,
						hint: 'Current privilege',
						icon: Shield,
						accent: 'sky'
					},
					{
						title: 'User ID',
						value: `#${user.ID}`,
						hint: 'Primary identifier',
						icon: ClipboardList,
						accent: 'slate'
					}
				]
			},
			{
				type: 'description',
				title: 'users.form.sections.personal',
				items: [
					{ label: 'users.form.fields.username', value: user.Username },
					{ label: 'users.form.fields.first_name', value: user.FirstName || 'Not set' },
					{ label: 'users.form.fields.last_name', value: user.LastName || 'Not set' },
					{ label: 'users.form.fields.email', value: user.Email || 'Not set' },
					{ label: 'users.form.fields.phone', value: user.PhoneNumber || 'Not set' },
					{ label: 'users.form.fields.address', value: user.Address || 'Not set' },
					{ label: 'users.table.role', value: user.Role.Name },
					{
						label: 'users.table.status',
						value: user.IsActive ? 'users.status.approved' : 'users.status.pending',
						badge: userStatusBadge(user.IsActive)
					},
					{ label: 'users.table.id', value: `#${user.ID}` }
				]
			}
		];
	};

	const applyFormFromUser = (user: UserSummary | null) => {
		if (!user) {
			form.username = '';
			form.password = '';
			form.role = '';
			form.firstName = '';
			form.lastName = '';
			form.email = '';
			form.phoneNumber = '';
			form.address = '';
			return;
		}
		form.username = user.Username;
		form.password = '';
		form.role = user.Role?.Name ?? '';
		form.firstName = user.FirstName || '';
		form.lastName = user.LastName || '';
		form.email = user.Email || '';
		form.phoneNumber = user.PhoneNumber || '';
		form.address = user.Address || '';
	};

	const loadUsers = async () => {
		listLoading = true;
		try {
			const params: Record<string, unknown> = {};
			if (activeTab !== 'all') params.status = activeTab;
			if (searchTerm.trim()) params.q = searchTerm.trim();
			const response = await usersApi.list(params);
			const data = response.users || [];
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

	// Load users when activeTab changes
	let initialLoadDone = false;
	
	onMount(async () => {
		try {
			await Promise.all([
				loadUsers(),
				rolesApi.list().then(data => roles = data).catch(err => {
					console.error('Failed to load roles:', err);
					toast.error('Failed to load roles');
				})
			]);
		} finally {
			initialLoadDone = true;
		}
	});

	$effect(() => {
		if (initialLoadDone) {
			activeTab;
			loadUsers();
		}
	});

	const handleSearch = () => {
		loadUsers();
	};

	const selectUser = (user: UserSummary) => {
		selectedUser = user;
		applyFormFromUser(user);
	};

	const editUser = (user: UserSummary) => {
		selectUser(user);
		// Scroll to form
		const formElement = document.getElementById('user-edit-form');
		if (formElement) {
			formElement.scrollIntoView({ behavior: 'smooth', block: 'center' });
		}
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
				role: form.role,
				firstName: form.firstName || undefined,
				lastName: form.lastName || undefined,
				email: form.email || undefined,
				phoneNumber: form.phoneNumber || undefined,
				address: form.address || undefined
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

		if (confirm(`Are you sure you want to delete ${username ?? selectedUser?.Username}?`)) {
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
<section
	class="animate-gradientShift relative w-full overflow-hidden bg-gradient-to-r from-teal-50 via-sky-50 to-indigo-100 px-6 py-16 sm:py-20"
>
	<div class="absolute inset-0 bg-white/40 backdrop-blur-[2px]"></div>

	<!-- floating blobs -->
	<div
		class="animate-floatSlow pointer-events-none absolute -right-20 -top-24 h-56 w-56 rounded-full bg-gradient-to-br from-sky-300/50 to-teal-300/50 blur-3xl"
	></div>
	<div
		class="animate-floatSlow pointer-events-none absolute -bottom-24 -left-20 h-56 w-56 rounded-full bg-gradient-to-br from-indigo-300/50 to-sky-300/50 blur-3xl delay-500"
	></div>

	<!-- hero content -->
	<div
		class="parallax-hero relative z-10 mx-auto flex max-w-5xl flex-col items-center gap-4 text-center will-change-transform"
	>
		<div
			class="animate-pulseGlow rounded-2xl bg-gradient-to-br from-sky-500 to-indigo-600 p-4 shadow-xl sm:p-5"
		>
			<UserCheck class="h-8 w-8 text-white sm:h-9 sm:w-9" />
		</div>
		<h1
			class="animate-fadeUp bg-gradient-to-r from-sky-700 via-indigo-700 to-teal-700 bg-clip-text text-3xl font-extrabold tracking-tight text-transparent sm:text-5xl"
		>
			{$t('users.title')}
		</h1>
		<p class="animate-fadeUp max-w-2xl text-slate-700 delay-150">
			{$t('users.subtitle')}
		</p>
		<div class="animate-fadeUp mt-2 flex flex-wrap items-center justify-center gap-2 delay-200">
			<span
				class="rounded-full border border-sky-200 bg-white/70 px-3 py-1.5 text-xs text-sky-700 shadow-sm sm:text-sm"
				>{$t('users.badges.role_control')}</span
			>
			<span
				class="rounded-full border border-indigo-200 bg-white/70 px-3 py-1.5 text-xs text-indigo-700 shadow-sm sm:text-sm"
				>{$t('users.badges.status_filters')}</span
			>
			<span
				class="rounded-full border border-teal-200 bg-white/70 px-3 py-1.5 text-xs text-teal-700 shadow-sm sm:text-sm"
				>{$t('users.badges.inline_edits')}</span
			>
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
<section class="mx-auto max-w-7xl space-y-10 bg-white px-4 py-12 sm:px-6 sm:py-14">
	<!-- Search & Filter -->
	<Card
		class="rounded-2xl border-0 bg-gradient-to-br from-sky-50 to-indigo-100 shadow-lg transition-all duration-300 hover:scale-[1.01] hover:shadow-xl"
	>
		<CardHeader
			class="space-y-4 rounded-t-2xl border-b border-white/60 bg-white/80 px-6 py-5 backdrop-blur"
		>
			<div class="flex flex-col gap-3 sm:flex-row">
				<Input
					class="flex-1 rounded-xl border-sky-200 bg-white/90 focus:ring-2 focus:ring-sky-400"
					placeholder={$t('users.filters.search_placeholder')}
					value={searchTerm}
					oninput={(event) => (searchTerm = event.currentTarget.value)}
					onkeydown={(event) => event.key === 'Enter' && handleSearch()}
				/>
				<Button
					class="rounded-xl bg-gradient-to-r from-sky-500 to-indigo-600 text-white shadow-md transition-all hover:from-sky-600 hover:to-indigo-700 hover:shadow-lg"
					onclick={handleSearch}
				>
					{$t('users.filters.search_btn')}
				</Button>
			</div>
			<div class="flex flex-wrap gap-2">
				{#each tabFilters as tab}
					<Button
						variant={tab.value === activeTab ? 'default' : 'secondary'}
						class={`flex-1 rounded-xl ${tab.value === activeTab ? 'bg-gradient-to-r from-sky-500 to-indigo-600 text-white' : 'border border-sky-200 bg-white/70 text-sky-700 hover:bg-sky-50'}`}
						onclick={() => (activeTab = tab.value)}
					>
						{$t(tab.label)}
					</Button>
				{/each}
			</div>
		</CardHeader>

		<CardContent class="overflow-hidden rounded-b-2xl p-0">
			<div class="overflow-x-auto">
				<table class="min-w-full divide-y divide-sky-200 text-sm">
					<thead
						class="bg-gradient-to-r from-sky-100 to-indigo-100 text-left text-xs uppercase tracking-wide text-slate-600"
					>
						<tr>
							<th class="px-4 py-3 font-medium">{$t('users.table.id')}</th>
							<th class="px-4 py-3 font-medium">{$t('users.table.username')}</th>
							<th class="px-4 py-3 font-medium">{$t('users.table.role')}</th>
							<th class="px-4 py-3 font-medium">{$t('users.table.status')}</th>
							<th class="px-4 py-3 text-right font-medium">{$t('users.table.actions')}</th>
						</tr>
					</thead>
					<tbody class="[&>tr:nth-child(even)]:bg-white/70 [&>tr:nth-child(odd)]:bg-white/50">
						{#if listLoading}
							{#each Array(4) as _, i}
								<tr
									><td colspan="5" class="px-4 py-3"><Skeleton class="h-6 w-full bg-white/60" /></td
									></tr
								>
							{/each}
						{:else if users.length === 0}
							<tr>
								<td colspan="5" class="px-4 py-6 text-center text-sm text-slate-600">
									{$t('users.table.empty')}
								</td>
							</tr>
						{:else}
							{#each users as item}
								<tr
									class={selectedUser?.ID === item.ID
										? 'border-l-4 border-sky-400 bg-sky-50'
										: 'cursor-pointer hover:bg-white/90'}
									on:click={() => openUserDetails(item)}
								>
									<td class="px-4 py-3 font-mono text-xs text-slate-700">#{item.ID}</td>
									<td class="px-4 py-3 text-slate-800">{item.Username}</td>
									<td class="px-4 py-3 text-slate-700">{item.Role.Name}</td>
									<td class="px-4 py-3">
										<span
											class={`rounded-full px-2.5 py-0.5 text-xs capitalize shadow-sm ${
												item.IsActive
													? 'border border-sky-200 bg-sky-100 text-sky-700'
													: 'border border-amber-200 bg-amber-100 text-amber-800'
											}`}
										>
											{item.IsActive ? $t('users.status.approved') : $t('users.status.pending')}
										</span>
									</td>
									<td class="space-x-1 px-4 py-3 text-right">
										<Button
											size="sm"
											variant="ghost"
											disabled={item.IsActive}
											class="rounded-lg px-3 py-1.5 text-sky-700 hover:bg-sky-100"
											onclick={(event: MouseEvent) => {
												event.stopPropagation();
												approveUser(item.ID);
											}}
										>
											Approve
										</Button>
										<Button
											size="sm"
											variant="ghost"
											class="rounded-lg px-3 py-1.5 text-indigo-700 hover:bg-indigo-100"
											onclick={(event: MouseEvent) => {
												event.stopPropagation();
												editUser(item);
											}}
										>
											<Pencil class="mr-1 h-3.5 w-3.5" />
											{$t('users.table.edit')}
										</Button>
										<Button
											size="sm"
											variant="ghost"
											class="rounded-lg px-3 py-1.5 text-rose-700 hover:bg-rose-100"
											onclick={(event: MouseEvent) => {
												event.stopPropagation();
												deleteUser(item.ID, item.Username);
											}}
										>
											{$t('users.table.delete')}
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
	<Card
		id="user-edit-form"
		class="rounded-2xl border-0 bg-gradient-to-br from-teal-50 to-sky-100 shadow-lg transition-all duration-300 hover:scale-[1.01] hover:shadow-xl"
	>
		<CardHeader class="rounded-t-2xl border-b border-white/60 bg-white/80 px-6 py-5 backdrop-blur">
			<CardTitle class="text-slate-800">{$t('users.form.title')}</CardTitle>
			<CardDescription class="text-slate-600"
				>{$t('users.form.subtitle')}</CardDescription
			>
		</CardHeader>
		{#if !selectedUser}
			<CardContent class="p-6">
				<p class="text-sm text-slate-600">{$t('users.form.select_prompt')}</p>
			</CardContent>
		{:else}
			<CardContent class="space-y-4 p-6">
				<p class="text-xs text-slate-500">{$t('users.form.editing')} #{selectedUser.ID}</p>

				<!-- Basic Info -->
				<div class="space-y-3">
					<p class="border-b border-sky-200 pb-2 text-sm font-semibold text-slate-700">
						{$t('users.form.sections.credentials')}
					</p>
					<Input
						class="rounded-xl border-sky-200 bg-white/90 focus:ring-2 focus:ring-sky-400"
						placeholder={$t('users.form.fields.username')}
						bind:value={form.username}
					/>
					<Input
						type="password"
						class="rounded-xl border-sky-200 bg-white/90 focus:ring-2 focus:ring-sky-400"
						placeholder={$t('users.form.fields.password_placeholder')}
						bind:value={form.password}
					/>
					<select
						class="w-full rounded-xl border border-sky-200 bg-white/90 px-3 py-2.5 text-sm focus:ring-2 focus:ring-sky-400"
						bind:value={form.role}
					>
						<option value="" disabled>{$t('users.form.fields.select_role')}</option>
						{#each roles as role}
							<option value={role.Name}>{role.Name}</option>
						{/each}
					</select>
				</div>

				<!-- Personal Info -->
				<div class="space-y-3">
					<p class="border-b border-sky-200 pb-2 text-sm font-semibold text-slate-700">
						{$t('users.form.sections.personal')}
					</p>
					<div class="grid grid-cols-1 gap-3 sm:grid-cols-2">
						<Input
							class="rounded-xl border-sky-200 bg-white/90 focus:ring-2 focus:ring-sky-400"
							placeholder={$t('users.form.fields.first_name')}
							bind:value={form.firstName}
						/>
						<Input
							class="rounded-xl border-sky-200 bg-white/90 focus:ring-2 focus:ring-sky-400"
							placeholder={$t('users.form.fields.last_name')}
							bind:value={form.lastName}
						/>
					</div>
					<Input
						type="email"
						class="rounded-xl border-sky-200 bg-white/90 focus:ring-2 focus:ring-sky-400"
						placeholder={$t('users.form.fields.email')}
						bind:value={form.email}
					/>
					<Input
						type="tel"
						class="rounded-xl border-sky-200 bg-white/90 focus:ring-2 focus:ring-sky-400"
						placeholder={$t('users.form.fields.phone')}
						bind:value={form.phoneNumber}
					/>
					<Input
						class="rounded-xl border-sky-200 bg-white/90 focus:ring-2 focus:ring-sky-400"
						placeholder={$t('users.form.fields.address')}
						bind:value={form.address}
					/>
				</div>

				<div class="flex flex-col gap-3 pt-2 sm:flex-row">
					{#if auth.hasPermission('users.manage')}
						<Button
							class="flex-1 rounded-xl bg-gradient-to-r from-sky-500 to-indigo-600 font-semibold text-white shadow-md transition-all hover:scale-105 hover:from-sky-600 hover:to-indigo-700 hover:shadow-lg"
							onclick={updateUser}
						>
							{$t('users.form.buttons.save')}
						</Button>
						<Button
							class="flex-1 rounded-xl border border-sky-200 text-sky-700 hover:bg-sky-50"
							onclick={() => approveUser()}
							disabled={selectedUser.IsActive}
						>
							{$t('users.form.buttons.approve')}
						</Button>
						<Button
							class="flex-1 rounded-xl border border-rose-200 text-rose-700 hover:bg-rose-50"
							onclick={() => deleteUser()}
						>
							{$t('users.form.buttons.delete')}
						</Button>
					{:else}
						<div
							class="w-full rounded-xl border border-amber-200 bg-amber-50 p-4 text-center text-sm text-amber-700"
						>
							{$t('users.form.read_only')}
						</div>
					{/if}
				</div>
			</CardContent>
		{/if}
	</Card>
</section>

<style lang="postcss">
	@keyframes gradientShift {
		0% {
			background-position: 0% 50%;
		}
		50% {
			background-position: 100% 50%;
		}
		100% {
			background-position: 0% 50%;
		}
	}
	.animate-gradientShift {
		background-size: 200% 200%;
		animation: gradientShift 20s ease-in-out infinite;
	}

	@keyframes pulseGlow {
		0%,
		100% {
			transform: scale(1);
			box-shadow: 0 0 15px rgba(59, 130, 246, 0.3);
		}
		50% {
			transform: scale(1.08);
			box-shadow: 0 0 25px rgba(59, 130, 246, 0.5);
		}
	}
	.animate-pulseGlow {
		animation: pulseGlow 8s ease-in-out infinite;
	}

	@keyframes fadeUp {
		from {
			opacity: 0;
			transform: translateY(20px);
		}
		to {
			opacity: 1;
			transform: translateY(0);
		}
	}
	.animate-fadeUp {
		animation: fadeUp 1.3s ease forwards;
	}
	.animate-fadeUp.delay-150 {
		animation-delay: 150ms;
	}
	.animate-fadeUp.delay-200 {
		animation-delay: 200ms;
	}

	@keyframes floatSlow {
		0%,
		100% {
			transform: translateY(0px) scale(1);
		}
		50% {
			transform: translateY(-10px) scale(1.03);
		}
	}
	.animate-floatSlow {
		animation: floatSlow 10s ease-in-out infinite;
	}

	* {
		transition-property:
			color, background-color, border-color, box-shadow, transform, filter, backdrop-filter;
		transition-timing-function: cubic-bezier(0.4, 0, 0.2, 1);
		transition-duration: 300ms;
	}
</style>
