<script lang="ts">
	import { onMount } from 'svelte';
	import { usersApi } from '$lib/api/resources';
	import type { UserSummary } from '$lib/types';
	import { toast } from 'svelte-sonner';
	import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '$lib/components/ui/card';
	import { Input } from '$lib/components/ui/input';
	import { Button } from '$lib/components/ui/button';
	import { Skeleton } from '$lib/components/ui/skeleton';

	const tabFilters = [
		{ value: 'all', label: 'All users' },
		{ value: 'approved', label: 'Approved' },
		{ value: 'pending', label: 'Pending' },
	] as const;

	type TabValue = (typeof tabFilters)[number]['value'];

	let activeTab = $state<TabValue>('all');
	let searchTerm = $state('');
	let users = $state<UserSummary[]>([]);
	let listLoading = $state(false);

	let selectedUser = $state<UserSummary | null>(null);
	const form = $state({ username: '', password: '', role: '' });

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
		} catch (error) {
			const errorMessage = error.response?.data?.error || 'Unable to load users';
			toast.error('Failed to Load Users', {
				description: errorMessage,
			});
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

	const updateUser = async () => {
		if (!selectedUser) return;
		try {
			await usersApi.update(selectedUser.ID, {
				username: form.username,
				password: form.password || undefined,
				role: form.role,
			});
			toast.success('User updated');
			await loadUsers();
		} catch (error) {
			const errorMessage = error.response?.data?.error || 'Unable to update user';
			toast.error('Failed to Update User', {
				description: errorMessage,
			});
		}
	};

	const approveUser = async (userId?: number) => {
		const targetId = userId ?? selectedUser?.ID;
		if (!targetId) return;
		try {
			await usersApi.approve(targetId);
			toast.success('User approved');
			await loadUsers();
		} catch (error) {
			const errorMessage = error.response?.data?.error || 'Unable to approve user';
			toast.error('Failed to Approve User', {
				description: errorMessage,
			});
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
					toast.error('Failed to Delete User', {
						description: errorMessage,
					});
				}
			},
		});
	};
</script>

<section class="space-y-8">
	<header class="space-y-2">
		<p class="text-sm uppercase tracking-wide text-muted-foreground">Access control</p>
		<h1 class="text-3xl font-semibold">Approve, edit, or revoke workspace access</h1>
		<p class="text-sm text-muted-foreground">Search by username or ID, then review by status.</p>
	</header>

	<Card>
		<CardHeader class="space-y-4">
			<div class="flex flex-wrap gap-2">
				<Input
					class="flex-1"
					placeholder="Search by username or ID"
					value={searchTerm}
					oninput={(event) => (searchTerm = event.currentTarget.value)}
					onkeydown={(event) => event.key === 'Enter' && handleSearch()}
				/>
				<Button variant="secondary" onclick={handleSearch}>Search</Button>
			</div>
			<div class="flex gap-2">
				{#each tabFilters as tab}
					<Button
						varient={undefined}
						variant={tab.value === activeTab ? 'default' : 'secondary'}
						class="flex-1"
						onclick={() => (activeTab = tab.value)}
					>
						{tab.label}
					</Button>
				{/each}
			</div>
		</CardHeader>
		<CardContent>
			<div class="overflow-x-auto rounded-2xl border border-border/70">
				<table class="min-w-full divide-y divide-border/70 text-sm">
					<thead class="bg-muted/50 text-left text-xs uppercase tracking-wide text-muted-foreground">
						<tr>
							<th class="px-4 py-3 font-medium">ID</th>
							<th class="px-4 py-3 font-medium">Username</th>
							<th class="px-4 py-3 font-medium">Role</th>
							<th class="px-4 py-3 font-medium">Status</th>
							<th class="px-4 py-3 font-medium text-right">Actions</th>
						</tr>
					</thead>
					<tbody class="divide-y divide-border/60 bg-background">
						{#if listLoading}
							{#each Array(4) as _, i}
								<tr>
									<td colspan="5" class="px-4 py-3">
										<Skeleton class="h-6 w-full" />
									</td>
								</tr>
							{/each}
						{:else if users.length === 0}
							<tr>
								<td colspan="5" class="px-4 py-6 text-center text-sm text-muted-foreground">
									No users match this filter
								</td>
							</tr>
						{:else}
							{#each users as item}
								<tr
									class={selectedUser?.ID === item.ID ? 'bg-primary/5' : 'hover:bg-muted/50 cursor-pointer'}
									onclick={() => selectUser(item)}
								>
									<td class="px-4 py-3 font-mono text-xs">#{item.ID}</td>
									<td class="px-4 py-3">{item.Username}</td>
									<td class="px-4 py-3">{item.Role}</td>
									<td class="px-4 py-3">
										<span class={`rounded-full px-2 py-0.5 text-xs ${item.IsActive ? 'bg-primary/10 text-primary' : 'bg-amber-100 text-amber-800'}`}>
											{item.IsActive ? 'Approved' : 'Pending'}
										</span>
									</td>
									<td class="px-4 py-3 text-right space-x-1">
										<Button
											size="sm"
											variant="ghost"
											disabled={item.IsActive}
											onclick={(event) => {
												event.stopPropagation();
												approveUser(item.ID);
											}}
										>
											Approve
										</Button>
										<Button
											size="sm"
											variant="ghost"
											class="text-destructive"
											onclick={(event) => {
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

	<Card>
		<CardHeader>
			<CardTitle>User details</CardTitle>
			<CardDescription>Update role or credentials for the selected user.</CardDescription>
		</CardHeader>
		{#if !selectedUser}
			<CardContent>
				<p class="text-sm text-muted-foreground">Select a user from the table to edit access.</p>
			</CardContent>
		{:else}
			<CardContent class="space-y-3">
				<p class="text-xs text-muted-foreground">Editing #{selectedUser.ID}</p>
				<Input
					placeholder="Username"
					value={form.username}
					oninput={(event) => (form.username = event.currentTarget.value)}
				/>
				<Input
					type="password"
					placeholder="Reset password (optional)"
					value={form.password}
					oninput={(event) => (form.password = event.currentTarget.value)}
				/>
				<select
					class="w-full rounded-md border border-border bg-background px-3 py-2 text-sm"
					value={form.role}
					onchange={(event) => (form.role = event.currentTarget.value)}
				>
					<option value="Admin">Admin</option>
					<option value="Manager">Manager</option>
					<option value="Staff">Staff</option>
				</select>
				<div class="flex flex-wrap gap-2">
					<Button class="flex-1" onclick={updateUser}>Save changes</Button>
					<Button class="flex-1" variant="secondary" onclick={() => approveUser()} disabled={selectedUser.IsActive}>
						Approve
					</Button>
					<Button class="flex-1" variant="destructive" onclick={() => deleteUser()}>
						Delete
					</Button>
				</div>
			</CardContent>
		{/if}
	</Card>
</section>
