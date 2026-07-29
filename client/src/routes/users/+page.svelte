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

	const inputClass =
		'liquid-input h-10 w-full rounded-md bg-card px-3 text-sm text-foreground placeholder:text-muted-foreground focus:border-primary focus:ring-0';

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

<!-- HEADER -->
<section class="w-full">
	<div class="mx-auto max-w-7xl px-4 pt-10 sm:px-6 lg:px-8">
		<div class="space-y-3" data-animate="fade-up">
			<div class="flex items-center gap-3">
				<span class="glass-button inline-flex rounded-md p-2">
					<UserCheck class="text-primary h-5 w-5" />
				</span>
				<p class="overline">Staff · Access</p>
			</div>
			<h1 class="font-display text-foreground text-3xl font-bold tracking-tight sm:text-4xl">
				{$t('users.title')}
			</h1>
			<p class="text-muted-foreground max-w-2xl">
				{$t('users.subtitle')}
			</p>
			<div class="flex flex-wrap items-center gap-2 pt-1">
				<span class="badge-neutral">{$t('users.badges.role_control')}</span>
				<span class="badge-neutral">{$t('users.badges.status_filters')}</span>
				<span class="badge-neutral">{$t('users.badges.inline_edits')}</span>
			</div>
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
<section class="mx-auto max-w-7xl space-y-8 px-4 py-10 sm:px-6">
	<!-- Search & Filter -->
	<div class="liquid-panel overflow-hidden rounded-lg" data-animate="fade-up" style="animation-delay: 100ms;">
		<div class="border-border bg-muted/40 space-y-4 border-b px-6 py-5">
			<div class="flex flex-col gap-3 sm:flex-row">
				<Input
					class="{inputClass} flex-1"
					placeholder={$t('users.filters.search_placeholder')}
					value={searchTerm}
					oninput={(event) => (searchTerm = event.currentTarget.value)}
					onkeydown={(event) => event.key === 'Enter' && handleSearch()}
				/>
				<Button
					class="bg-primary text-primary-foreground rounded-md hover:bg-primary/90"
					onclick={handleSearch}
				>
					{$t('users.filters.search_btn')}
				</Button>
			</div>
			<div class="flex flex-wrap gap-2">
				{#each tabFilters as tab}
					<Button
						variant={tab.value === activeTab ? 'default' : 'secondary'}
						class={tab.value === activeTab
							? 'glass-toggle-active flex-1 rounded-md'
							: 'glass-button text-muted-foreground flex-1 rounded-md'}
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
								<td colspan="5"><Skeleton class="bg-muted h-6 w-full" /></td>
							</tr>
						{/each}
					{:else if users.length === 0}
						<tr>
							<td colspan="5" class="text-muted-foreground py-6 text-center text-sm">
								{$t('users.table.empty')}
							</td>
						</tr>
					{:else}
						{#each users as item}
							<tr
								class={selectedUser?.ID === item.ID
									? 'border-l-2 border-primary bg-muted/60'
									: 'cursor-pointer'}
								onclick={() => openUserDetails(item)}
							>
								<td class="num-mono text-muted-foreground text-xs">#{item.ID}</td>
								<td class="text-foreground font-medium">{item.Username}</td>
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
											class="bg-success text-primary-foreground hover:bg-success/90 rounded-md px-3 py-1.5"
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
											class="text-foreground hover:bg-muted rounded-md px-3 py-1.5"
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
											class="text-danger hover:bg-danger/10 hover:text-danger rounded-md px-3 py-1.5"
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
	</div>

	<!-- User Details -->
	<div
		id="user-edit-form"
		class="liquid-panel overflow-hidden rounded-lg"
		data-animate="fade-up"
		style="animation-delay: 180ms;"
	>
		<div class="border-border bg-muted/40 border-b px-6 py-5">
			<h2 class="font-display text-foreground text-lg font-bold">{$t('users.form.title')}</h2>
			<p class="text-muted-foreground text-sm">{$t('users.form.subtitle')}</p>
		</div>
		{#if !selectedUser}
			<div class="p-6">
				<p class="text-muted-foreground text-sm">{$t('users.form.select_prompt')}</p>
			</div>
		{:else}
			<div class="space-y-4 p-6">
				<p class="text-muted-foreground num-mono text-xs">{$t('users.form.editing')} #{selectedUser.ID}</p>

				<!-- Basic Info -->
				<div class="space-y-3">
					<p class="overline border-border border-b pb-2">
						{$t('users.form.sections.credentials')}
					</p>
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
					<div class="select-wrapper">
						<select bind:value={form.role}>
							<option value="" disabled>{$t('users.form.fields.select_role')}</option>
							{#each roles as role}
								<option value={role.Name}>{role.Name}</option>
							{/each}
						</select>
					</div>
				</div>

				<!-- Personal Info -->
				<div class="space-y-3">
					<p class="overline border-border border-b pb-2">
						{$t('users.form.sections.personal')}
					</p>
					<div class="grid grid-cols-1 gap-3 sm:grid-cols-2">
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
					</div>
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
						class={inputClass}
						placeholder={$t('users.form.fields.address')}
						bind:value={form.address}
					/>
				</div>

				<div class="flex flex-col gap-3 pt-2 sm:flex-row">
					{#if auth.hasPermission('users.manage')}
						<Button
							class="bg-primary text-primary-foreground flex-1 rounded-md font-semibold hover:bg-primary/90"
							onclick={updateUser}
						>
							{$t('users.form.buttons.save')}
						</Button>
						<Button
							class="bg-success text-primary-foreground flex-1 rounded-md font-semibold hover:bg-success/90"
							onclick={() => approveUser()}
							disabled={selectedUser.IsActive}
						>
							{$t('users.form.buttons.approve')}
						</Button>
						<Button
							class="text-danger hover:bg-danger/10 hover:text-danger flex-1 rounded-md border border-danger/25 font-semibold"
							onclick={() => deleteUser()}
						>
							{$t('users.form.buttons.delete')}
						</Button>
					{:else}
						<div
							class="text-warning w-full rounded-md border border-warning/25 bg-warning/10 p-4 text-center text-sm"
						>
							{$t('users.form.read_only')}
						</div>
					{/if}
				</div>
			</div>
		{/if}
	</div>
</section>

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
