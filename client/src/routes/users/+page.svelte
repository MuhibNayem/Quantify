<script lang="ts">
	import { onMount } from 'svelte';
	import { usersApi } from '$lib/api/resources';
	import { rolesApi } from '$lib/api/roles';
	import type { Role, UserSummary } from '$lib/types';
	import { toast } from 'svelte-sonner';
	import { Input } from '$lib/components/ui/input';
	import { Button } from '$lib/components/ui/button';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import DetailsModal from '$lib/components/DetailsModal.svelte';
	import type { DetailBuilderContext, DetailSection } from '$lib/components/DetailsModal.svelte';
	import ConfirmDialog from '$lib/components/ui/confirm-dialog.svelte';
	import { UserCheck, Shield, ClipboardList, CheckCircle2, Pencil } from 'lucide-svelte';
	import { auth } from '$lib/stores/auth';
	import { t } from '$lib/i18n';
	import { goto } from '$app/navigation';

	const inputClass = 'h-11 w-full rounded-md text-sm';

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
			? { text: $t('users.status.approved'), variant: 'success' as const }
			: { text: $t('users.status.pending'), variant: 'warning' as const };

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
				title: $t('users.form.sections.personal'),
				items: [
					{ label: $t('users.form.fields.username'), value: user.Username },
					{ label: $t('users.form.fields.first_name'), value: user.FirstName || 'Not set' },
					{ label: $t('users.form.fields.last_name'), value: user.LastName || 'Not set' },
					{ label: $t('users.form.fields.email'), value: user.Email || 'Not set' },
					{ label: $t('users.form.fields.phone'), value: user.PhoneNumber || 'Not set' },
					{ label: $t('users.form.fields.address'), value: user.Address || 'Not set' },
					{ label: $t('users.table.role'), value: user.Role.Name },
					{
						label: $t('users.table.status'),
						value: user.IsActive ? $t('users.status.approved') : $t('users.status.pending'),
						badge: userStatusBadge(user.IsActive)
					},
					{ label: $t('users.table.id'), value: `#${user.ID}` }
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
				rolesApi
					.list()
					.then((data) => (roles = data))
					.catch((err) => {
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

	let deleteDialogOpen = $state(false);
	let deleteTargetUser = $state<{ id: number; username: string } | null>(null);

	const deleteUser = (userId?: number, username?: string) => {
		const targetId = userId ?? selectedUser?.ID;
		const targetUsername = username ?? selectedUser?.Username ?? '';
		if (!targetId) return;
		deleteTargetUser = { id: targetId, username: targetUsername };
		deleteDialogOpen = true;
	};

	const confirmDeleteUser = async () => {
		if (!deleteTargetUser) return;
		try {
			await usersApi.remove(deleteTargetUser.id);
			toast.success('User removed');
			if (selectedUser?.ID === deleteTargetUser.id) {
				selectedUser = null;
				applyFormFromUser(null);
			}
			await loadUsers();
		} catch (error: any) {
			const errorMessage = error.response?.data?.error || 'Unable to delete user';
			toast.error('Failed to Delete User', { description: errorMessage });
		}
		deleteTargetUser = null;
	};
</script>

<div class="mx-auto w-full max-w-7xl px-4 py-8 sm:px-6 lg:px-8 lg:py-10">
	<!-- Page header -->
	<header class="page-header" data-animate="fade-up">
		<div>
			<p class="flex items-center gap-2 overline">
				<span class="inline-flex rounded-md bg-primary/10 p-1.5 text-primary">
					<UserCheck class="h-3.5 w-3.5" />
				</span>
				Staff · Access
			</p>
			<h1 class="page-title mt-1">{$t('users.title')}</h1>
			<p class="page-sub">{$t('users.subtitle')}</p>
		</div>
		<div class="flex flex-wrap items-center gap-2">
			<span class="badge-neutral">{$t('users.badges.role_control')}</span>
			<span class="badge-neutral">{$t('users.badges.status_filters')}</span>
			<span class="badge-neutral">{$t('users.badges.inline_edits')}</span>
		</div>
	</header>

	<DetailsModal
		bind:open={detailModalOpen}
		resourceId={detailResourceId}
		endpoint="/users"
		title="User Details"
		subtitle={detailModalSubtitle}
		buildSections={buildUserSections}
	/>

	<!-- MAIN CONTENT -->
	<div class="mt-6 space-y-4">
		<!-- Search & Filter -->
		<section class="panel overflow-hidden" data-animate="fade-up" style="animation-delay: 80ms;">
			<div class="panel-body space-y-4">
				<div class="flex flex-col gap-3 sm:flex-row">
					<Input
						class="{inputClass} flex-1"
						placeholder={$t('users.filters.search_placeholder')}
						value={searchTerm}
						oninput={(event) => (searchTerm = event.currentTarget.value)}
						onkeydown={(event) => event.key === 'Enter' && handleSearch()}
					/>
					<Button class="btn-soothe h-11 shrink-0 px-5" onclick={handleSearch}>
						{$t('users.filters.search_btn')}
					</Button>
				</div>
				<div
					class="inline-flex w-full flex-wrap items-center gap-1 rounded-md border border-border/70 bg-muted/50 p-1 sm:w-auto"
					role="tablist"
				>
					{#each tabFilters as tab}
						<Button
							variant="ghost"
							size="sm"
							role="tab"
							aria-selected={tab.value === activeTab}
							class="flex-1 rounded-md px-4 text-sm font-medium sm:flex-none {tab.value ===
							activeTab
								? 'bg-card text-foreground shadow-xs ring-1 ring-border'
								: 'text-muted-foreground hover:text-foreground'}"
							onclick={() => (activeTab = tab.value)}
						>
							{$t(tab.label)}
						</Button>
					{/each}
				</div>
			</div>

			<div class="overflow-x-auto">
				<table class="ledger-table">
					<thead>
						<tr>
							<th>{$t('users.table.id')}</th>
							<th>{$t('users.table.username')}</th>
							<th>{$t('users.table.role')}</th>
							<th>{$t('users.table.status')}</th>
							<th class="text-right">{$t('users.table.actions')}</th>
						</tr>
					</thead>
					<tbody>
						{#if listLoading}
							{#each Array(4) as _, i}
								<tr>
									<td colspan="5">
										<Skeleton class="h-6 w-full bg-muted" />
									</td>
								</tr>
							{/each}
						{:else if users.length === 0}
							<tr>
								<td colspan="5" class="py-14 text-center">
									<div class="flex flex-col items-center gap-2">
										<span
											class="flex h-12 w-12 items-center justify-center rounded-lg bg-muted text-muted-foreground ring-1 ring-border"
										>
											<UserCheck class="h-5 w-5" />
										</span>
										<p class="text-sm font-medium text-foreground">{$t('users.table.empty')}</p>
										<p class="text-xs text-muted-foreground">
											Try a different filter or search term.
										</p>
									</div>
								</td>
							</tr>
						{:else}
							{#each users as item}
								<tr
									class={selectedUser?.ID === item.ID
										? 'cursor-pointer bg-muted/60 shadow-[inset_2px_0_0_0_hsl(var(--primary))]'
										: 'cursor-pointer'}
									onclick={() => openUserDetails(item)}
								>
									<td class="num-mono text-xs text-muted-foreground">#{item.ID}</td>
									<td class="font-medium text-foreground">{item.Username}</td>
									<td>
										<span class="badge-info">{item.Role.Name}</span>
									</td>
									<td>
										<span class={item.IsActive ? 'badge-success' : 'badge-warning'}>
											{item.IsActive ? $t('users.status.approved') : $t('users.status.pending')}
										</span>
									</td>
									<td>
										<div class="flex items-center justify-end gap-1">
											<Button
												size="sm"
												variant="ghost"
												disabled={item.IsActive}
												class="h-8 px-2.5 text-xs font-semibold text-success hover:bg-success/10 hover:text-success"
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
												class="h-8 px-2.5 text-xs font-medium text-muted-foreground hover:text-foreground"
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
												class="h-8 px-2.5 text-xs font-medium text-danger hover:bg-danger/10 hover:text-danger"
												onclick={(event: MouseEvent) => {
													event.stopPropagation();
													deleteUser(item.ID, item.Username);
												}}
											>
												{$t('users.table.delete')}
											</Button>
										</div>
									</td>
								</tr>
							{/each}
						{/if}
					</tbody>
				</table>
			</div>
		</section>

		<!-- User Details -->
		<section
			id="user-edit-form"
			class="panel overflow-hidden"
			data-animate="fade-up"
			style="animation-delay: 140ms;"
		>
			<div class="panel-header items-start">
				<div class="space-y-0.5">
					<h2 class="section-title">{$t('users.form.title')}</h2>
					<p class="text-sm text-muted-foreground">{$t('users.form.subtitle')}</p>
				</div>
				{#if selectedUser}
					<span class="num-mono text-xs text-muted-foreground">
						{$t('users.form.editing')} #{selectedUser.ID}
					</span>
				{/if}
			</div>
			{#if !selectedUser}
				<div class="panel-body">
					<p class="text-sm text-muted-foreground">{$t('users.form.select_prompt')}</p>
				</div>
			{:else}
				<div class="panel-body space-y-6">
					<!-- Credentials -->
					<div class="space-y-3">
						<p class="overline">{$t('users.form.sections.credentials')}</p>
						<div class="grid gap-3 sm:grid-cols-2">
							<Input
								class={inputClass}
								placeholder={$t('users.form.fields.username')}
								bind:value={form.username}
							/>
							<Input
								type="password"
								class={inputClass}
								placeholder={$t('users.form.fields.password_placeholder')}
								bind:value={form.password}
							/>
						</div>
						<select bind:value={form.role} aria-label={$t('users.form.fields.select_role')}>
							<option value="" disabled>{$t('users.form.fields.select_role')}</option>
							{#each roles as role}
								<option value={role.Name}>{role.Name}</option>
							{/each}
						</select>
					</div>

					<!-- Personal Info -->
					<div class="space-y-3">
						<p class="overline">{$t('users.form.sections.personal')}</p>
						<div class="grid gap-3 sm:grid-cols-2">
							<Input
								class={inputClass}
								placeholder={$t('users.form.fields.first_name')}
								bind:value={form.firstName}
							/>
							<Input
								class={inputClass}
								placeholder={$t('users.form.fields.last_name')}
								bind:value={form.lastName}
							/>
							<Input
								type="email"
								class={inputClass}
								placeholder={$t('users.form.fields.email')}
								bind:value={form.email}
							/>
							<Input
								type="tel"
								class={inputClass}
								placeholder={$t('users.form.fields.phone')}
								bind:value={form.phoneNumber}
							/>
							<Input
								class="{inputClass} sm:col-span-2"
								placeholder={$t('users.form.fields.address')}
								bind:value={form.address}
							/>
						</div>
					</div>

					<div class="flex flex-col gap-2.5 border-t border-border/70 pt-4 sm:flex-row">
						{#if auth.hasPermission('users.manage')}
							<Button class="btn-soothe h-11 flex-1 font-semibold" onclick={updateUser}>
								{$t('users.form.buttons.save')}
							</Button>
							<Button
								variant="outline"
								class="h-11 flex-1 font-semibold"
								onclick={() => approveUser()}
								disabled={selectedUser.IsActive}
							>
								{$t('users.form.buttons.approve')}
							</Button>
							<Button
								variant="ghost"
								class="h-11 flex-1 border border-danger/25 font-semibold text-danger hover:bg-danger/10 hover:text-danger"
								onclick={() => deleteUser()}
							>
								{$t('users.form.buttons.delete')}
							</Button>
						{:else}
							<div
								class="w-full rounded-md border border-warning/25 bg-warning/10 p-4 text-center text-sm text-warning"
							>
								{$t('users.form.read_only')}
							</div>
						{/if}
					</div>
				</div>
			{/if}
		</section>
	</div>
</div>

<ConfirmDialog
	bind:open={deleteDialogOpen}
	title="Delete user"
	description={deleteTargetUser
		? `This will permanently delete the user "${deleteTargetUser.username}". This action cannot be undone.`
		: ''}
	confirmText={deleteTargetUser?.username ?? ''}
	confirmLabel="Delete"
	onConfirm={confirmDeleteUser}
/>
