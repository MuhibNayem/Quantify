<script lang="ts">
	import { onMount } from 'svelte';
	import { rolesApi, type Role, type Permission } from '$lib/api/roles';
	import { toast } from 'svelte-sonner';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import { Checkbox } from '$lib/components/ui/checkbox';
	import { Badge } from '$lib/components/ui/badge';
	import {
		Plus,
		Trash2,
		Save,
		Shield,
		ShieldCheck,
		Check,
		Sparkles,
		AlertCircle
	} from 'lucide-svelte';
	import { cn } from '$lib/utils';
	import { t } from '$lib/i18n';
	import { Textarea } from '$lib/components/ui/textarea';
	import { fade, fly, slide } from 'svelte/transition';

	let roles: Role[] = [];
	let permissions: Permission[] = [];
	let selectedRole: Role | null = null;
	let isLoading = false;
	let isSaving = false;

	// Permission State for Editing
	let selectedPermissionIds: number[] = [];

	// Create/Edit State
	let editName = '';
	let editDescription = '';

	onMount(async () => {
		await loadData();
	});

	async function loadData() {
		isLoading = true;
		try {
			const [rolesData, permissionsData] = await Promise.all([
				rolesApi.list(),
				rolesApi.listPermissions()
			]);
			roles = rolesData;
			permissions = permissionsData;
			// Select first role by default
			if (roles.length > 0 && !selectedRole) {
				selectRole(roles[0]);
			}
		} catch (error) {
			console.error('Failed to load roles data:', error);
			toast.error('Failed to load roles and permissions');
		} finally {
			isLoading = false;
		}
	}

	function selectRole(role: Role) {
		selectedRole = role;
		editName = role.Name;
		editDescription = role.Description;
		selectedPermissionIds = role.Permissions ? role.Permissions.map((p) => p.ID) : [];
	}

	function handleNewRole() {
		const newRole: Role = {
			ID: 0,
			Name: 'New Role',
			Description: '',
			IsSystem: false,
			Permissions: [],
			CreatedAt: '',
			UpdatedAt: '',
			DeletedAt: null
		};
		selectedRole = newRole;
		editName = '';
		editDescription = '';
		selectedPermissionIds = [];
	}

	function togglePermission(permId: number) {
		if (selectedPermissionIds.includes(permId)) {
			selectedPermissionIds = selectedPermissionIds.filter((id) => id !== permId);
		} else {
			selectedPermissionIds = [...selectedPermissionIds, permId];
		}
	}

	// Reactive Group Selection State
	$: groupSelectionStates = sortedGroups.reduce(
		(acc, group) => {
			const groupPerms = groupedPermissions[group] || [];
			acc[group] =
				groupPerms.length > 0 && groupPerms.every((p) => selectedPermissionIds.includes(p.ID));
			return acc;
		},
		{} as Record<string, boolean>
	);

	function toggleGroup(group: string, checked: boolean) {
		const groupPerms = permissions.filter((p) => p.Group === group);
		const groupIds = groupPerms.map((p) => p.ID);

		if (checked) {
			const missing = groupIds.filter((id) => !selectedPermissionIds.includes(id));
			selectedPermissionIds = [...selectedPermissionIds, ...missing];
		} else {
			selectedPermissionIds = selectedPermissionIds.filter((id) => !groupIds.includes(id));
		}
	}

	function handleReset() {
		if (selectedRole) selectRole(selectedRole);
	}

	async function handleSave() {
		if (!selectedRole) return;
		if (!editName.trim()) {
			toast.error('Role name is required');
			return;
		}
		isSaving = true;
		try {
			let savedRole: Role;
			if (selectedRole.ID === 0) {
				savedRole = await rolesApi.create({ name: editName, description: editDescription });
				toast.success('Role created successfully');
				roles = [...roles, savedRole];
			} else {
				savedRole = await rolesApi.update(selectedRole.ID, {
					name: editName,
					description: editDescription
				});
				toast.success('Role details updated');
				roles = roles.map((r) => (r.ID === savedRole.ID ? savedRole : r));
			}

			await rolesApi.updatePermissions(savedRole.ID, selectedPermissionIds);
			await loadData();
			const updated = roles.find((r) => r.ID === savedRole.ID);
			if (updated) selectRole(updated);
		} catch (error) {
			console.error('Failed to save role:', error);
			toast.error('Failed to save role');
		} finally {
			isSaving = false;
		}
	}

	async function handleDelete() {
		if (!selectedRole || selectedRole.IsSystem || selectedRole.ID === 0) return;
		if (!confirm(`Are you sure you want to delete role "${selectedRole.Name}"?`)) return;
		try {
			await rolesApi.delete(selectedRole.ID);
			toast.success('Role deleted');
			roles = roles.filter((r) => r.ID !== selectedRole!.ID);
			if (roles.length > 0) selectRole(roles[0]);
			else selectedRole = null;
		} catch (error) {
			console.error('Failed to delete role:', error);
			toast.error('Failed to delete role');
		}
	}

	// Group Permissions
	$: groupedPermissions = permissions.reduce(
		(acc, perm) => {
			if (!acc[perm.Group]) acc[perm.Group] = [];
			acc[perm.Group].push(perm);
			return acc;
		},
		{} as Record<string, Permission[]>
	);

	$: sortedGroups = Object.keys(groupedPermissions).sort();
</script>

<div class="panel overflow-hidden">
	<div class="flex h-[calc(100dvh-13rem)] min-h-[34rem] flex-col lg:flex-row">
		<!-- Left Column: Role List -->
		<aside
			class="flex w-full shrink-0 flex-col border-b border-border/70 lg:w-[19rem] lg:border-r lg:border-b-0"
		>
			<div class="flex items-center justify-between gap-2 px-4 py-3.5">
				<div class="min-w-0">
					<h3 class="section-title truncate">{$t('roles.title')}</h3>
					<p class="text-xs text-muted-foreground">
						<span class="num-mono">{roles.length}</span> role{roles.length === 1 ? '' : 's'}
					</p>
				</div>
				<Button
					variant="ghost"
					size="icon"
					class="h-8 w-8 shrink-0 rounded-md bg-accent text-accent-foreground transition-colors hover:bg-accent/70"
					onclick={handleNewRole}
					aria-label={$t('roles.create_new')}
					title={$t('roles.create_new')}
				>
					<Plus class="h-4 w-4" />
				</Button>
			</div>

			<div
				class="custom-scrollbar flex gap-2 overflow-x-auto px-3 pb-3 lg:flex-1 lg:flex-col lg:overflow-x-hidden lg:overflow-y-auto"
			>
				{#if isLoading && roles.length === 0}
					{#each Array(4) as _, i}
						<div
							class="h-14 animate-pulse rounded-md bg-muted"
							style={`animation-delay:${i * 80}ms`}
						></div>
					{/each}
				{/if}
				{#each roles as role (role.ID)}
					<button
						type="button"
						aria-current={selectedRole?.ID === role.ID ? 'true' : undefined}
						class={cn(
							'group relative flex min-w-[11rem] flex-col gap-0.5 rounded-md px-3 py-2.5 text-left transition-colors duration-150 lg:w-full lg:min-w-0',
							selectedRole?.ID === role.ID
								? 'bg-accent text-accent-foreground shadow-xs ring-1 ring-primary/20'
								: 'text-foreground hover:bg-muted'
						)}
						onclick={() => selectRole(role)}
					>
						<div class="flex w-full items-center justify-between gap-2">
							<span class="truncate text-sm font-semibold">
								{role.ID === 0
									? $t('roles.create_new')
									: $t(`roles.names.${role.Name.toLowerCase().replace(' ', '_')}`) !==
										  `roles.names.${role.Name.toLowerCase().replace(' ', '_')}`
										? $t(`roles.names.${role.Name.toLowerCase().replace(' ', '_')}`)
										: role.Name}
							</span>
							{#if role.IsSystem}
								<ShieldCheck class="h-3.5 w-3.5 shrink-0 text-warning" />
							{/if}
						</div>
						<p class="truncate text-xs text-muted-foreground">
							{role.Description || 'No description'}
						</p>
					</button>
				{/each}
			</div>
		</aside>

		<!-- Right Column: Editor -->
		<div class="flex min-w-0 flex-1 flex-col overflow-hidden">
			{#if selectedRole}
				<div in:fade={{ duration: 250 }} class="flex h-full min-h-0 flex-col">
					<!-- Header -->
					<div class="border-b border-border/70 px-5 py-4 sm:px-6">
						<div class="flex flex-col gap-4 xl:flex-row xl:items-center xl:justify-between">
							<div class="flex min-w-0 items-center gap-3.5">
								<div
									class={cn(
										'flex h-11 w-11 shrink-0 items-center justify-center rounded-md ring-1',
										selectedRole.IsSystem
											? 'bg-warning/10 text-warning ring-warning/20'
											: 'bg-accent text-accent-foreground ring-accent-foreground/10'
									)}
								>
									{#if selectedRole.IsSystem}
										<ShieldCheck class="h-5 w-5" />
									{:else if selectedRole.ID === 0}
										<Plus class="h-5 w-5" />
									{:else}
										<Shield class="h-5 w-5" />
									{/if}
								</div>
								<div class="min-w-0 space-y-1">
									<h2
										class="truncate font-display text-xl font-semibold tracking-tight text-foreground"
									>
										{selectedRole.ID === 0
											? $t('roles.create_new')
											: $t(`roles.names.${selectedRole.Name.toLowerCase().replace(' ', '_')}`) !==
												  `roles.names.${selectedRole.Name.toLowerCase().replace(' ', '_')}`
												? $t(`roles.names.${selectedRole.Name.toLowerCase().replace(' ', '_')}`)
												: selectedRole.Name}
									</h2>
									<div
										class="flex flex-wrap items-center gap-2 text-xs font-medium text-muted-foreground"
									>
										{#if selectedRole.IsSystem}
											<span class="badge-warning">
												<AlertCircle class="h-3 w-3" />
												{$t('roles.system_managed')}
											</span>
										{:else}
											<span class="badge-success">
												<Check class="h-3 w-3" />
												{$t('roles.custom_role')}
											</span>
										{/if}
										<span class="h-1 w-1 rounded-full bg-border"></span>
										<span class="whitespace-nowrap">
											<span class="num-mono">{selectedPermissionIds.length}</span>
											{$t('roles.active_permissions')}
										</span>
									</div>
								</div>
							</div>

							<!-- Actions -->
							<div class="flex shrink-0 flex-wrap items-center gap-2">
								<Button
									variant="outline"
									size="sm"
									onclick={handleReset}
									disabled={isSaving}
									class="font-medium"
								>
									{$t('roles.reset')}
								</Button>
								{#if !selectedRole.IsSystem && selectedRole.ID !== 0}
									<Button
										variant="ghost"
										size="sm"
										class="font-medium text-danger hover:bg-danger/10 hover:text-danger"
										onclick={handleDelete}
									>
										<Trash2 class="mr-1.5 h-3.5 w-3.5" />
										{$t('roles.delete')}
									</Button>
								{/if}
								<Button size="sm" class="btn-soothe px-4" onclick={handleSave} disabled={isSaving}>
									{#if isSaving}
										<div
											class="mr-1.5 h-3.5 w-3.5 animate-spin rounded-full border-2 border-primary-foreground/40 border-t-primary-foreground"
										></div>
									{:else}
										<Save class="mr-1.5 h-3.5 w-3.5" />
									{/if}
									{isSaving ? $t('roles.saving') : $t('roles.save_changes')}
								</Button>
							</div>
						</div>
					</div>

					<!-- Scrollable Content -->
					<div class="custom-scrollbar min-h-0 flex-1 overflow-y-auto px-5 py-5 sm:px-6">
						<!-- Inputs -->
						<div class="grid gap-4 sm:grid-cols-2">
							<div class="space-y-1.5">
								<Label class="ml-1" for="role-name">{$t('roles.role_name')}</Label>
								<div class="relative">
									<Input
										id="role-name"
										bind:value={editName}
										placeholder="e.g. Sales Associate"
										disabled={selectedRole.IsSystem}
										class="h-11 w-full rounded-md pr-9 text-sm"
									/>
									{#if selectedRole.IsSystem}
										<ShieldCheck
											class="absolute top-1/2 right-3 h-4 w-4 -translate-y-1/2 text-muted-foreground"
										/>
									{/if}
								</div>
							</div>
							<div class="space-y-1.5">
								<Label class="ml-1" for="role-description">{$t('roles.description')}</Label>
								<Input
									id="role-description"
									bind:value={editDescription}
									placeholder="Describe the responsibilities..."
									class="h-11 w-full rounded-md text-sm"
								/>
							</div>
						</div>

						<!-- Permissions -->
						<div class="mt-6 space-y-4">
							<div class="flex items-start gap-3">
								<span
									class="flex h-9 w-9 shrink-0 items-center justify-center rounded-md bg-accent text-accent-foreground ring-1 ring-accent-foreground/10"
								>
									<Sparkles class="h-4 w-4" />
								</span>
								<div class="space-y-0.5">
									<h3 class="section-title">{$t('roles.capabilities')}</h3>
									<p class="text-sm text-muted-foreground">{$t('roles.capabilities_desc')}</p>
								</div>
							</div>

							<div class="grid grid-cols-1 gap-4 md:grid-cols-2 2xl:grid-cols-3">
								{#each sortedGroups as group, i}
									<div
										in:fly={{ y: 12, duration: 350, delay: 60 + i * 40 }}
										class="rounded-lg border border-border/70 bg-muted/20 p-4"
									>
										<div class="mb-3 flex flex-wrap items-center justify-between gap-2">
											<h4 class="section-title">
												{$t(`permissions.groups.${group.toLowerCase()}`) !==
												`permissions.groups.${group.toLowerCase()}`
													? $t(`permissions.groups.${group.toLowerCase()}`)
													: group}
											</h4>
											<div class="flex items-center gap-2">
												<Label
													for="group-{group}"
													class="cursor-pointer text-[0.6875rem] font-semibold tracking-[0.08em] text-muted-foreground uppercase transition-colors hover:text-primary"
												>
													{$t('roles.select_all')}
												</Label>
												<Checkbox
													id="group-{group}"
													checked={groupSelectionStates[group]}
													onCheckedChange={(v) => toggleGroup(group, !!v)}
													class="h-4 w-4"
												/>
											</div>
										</div>

										<div class="grid gap-1.5">
											{#each groupedPermissions[group] as perm}
												<label
													class={cn(
														'relative flex cursor-pointer items-start gap-2.5 rounded-md px-2.5 py-2 transition-colors duration-150',
														selectedPermissionIds.includes(perm.ID)
															? 'bg-accent text-accent-foreground ring-1 ring-primary/20'
															: 'text-foreground hover:bg-muted'
													)}
												>
													<Checkbox
														checked={selectedPermissionIds.includes(perm.ID)}
														onCheckedChange={() => togglePermission(perm.ID)}
														class="mt-0.5 h-4 w-4 rounded"
													/>
													<div class="min-w-0 space-y-0.5">
														<span
															class={cn(
																'block text-sm font-semibold transition-colors',
																selectedPermissionIds.includes(perm.ID)
																	? 'text-accent-foreground'
																	: 'text-foreground'
															)}
														>
															{$t(`permissions.names.${perm.Name.replaceAll('.', '_')}`) !==
															`permissions.names.${perm.Name.replaceAll('.', '_')}`
																? $t(`permissions.names.${perm.Name.replaceAll('.', '_')}`)
																: perm.Name}
														</span>
														<span
															class="line-clamp-2 block text-xs leading-relaxed text-muted-foreground"
														>
															{$t(`permissions.descriptions.${perm.Name.replaceAll('.', '_')}`) !==
															`permissions.descriptions.${perm.Name.replaceAll('.', '_')}`
																? $t(`permissions.descriptions.${perm.Name.replaceAll('.', '_')}`)
																: perm.Description}
														</span>
													</div>
												</label>
											{/each}
										</div>
									</div>
								{/each}
							</div>
						</div>
					</div>
				</div>
			{:else}
				<div
					class="flex h-full flex-col items-center justify-center px-6 py-16 text-center"
					in:fade
				>
					<div
						class="mb-5 flex h-14 w-14 items-center justify-center rounded-lg bg-muted text-muted-foreground ring-1 ring-border"
					>
						<Shield class="h-7 w-7" />
					</div>
					<h3 class="font-display text-lg font-semibold tracking-tight text-foreground">
						{$t('roles.security_access_control')}
					</h3>
					<p class="mt-2 max-w-md text-sm text-muted-foreground">
						{$t('roles.security_desc')}
					</p>
					<Button class="btn-soothe mt-6 px-5" onclick={handleNewRole}>
						<Plus class="mr-1.5 h-4 w-4" />
						{$t('roles.create_first')}
					</Button>
				</div>
			{/if}
		</div>
	</div>
</div>
