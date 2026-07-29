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

<div class="border-border bg-card relative h-[calc(100vh-140px)] w-full overflow-hidden rounded-lg border shadow-md">
	<div class="relative z-10 flex h-full flex-col gap-6 p-4 lg:flex-row">
		<!-- Left Column: Role List -->
		<div
			class="border-border bg-card flex w-full shrink-0 flex-col overflow-hidden rounded-md border shadow-xs lg:w-80"
		>
			<div class="border-border bg-muted/40 flex items-center justify-between border-b p-5">
				<h3 class="font-display text-foreground text-lg font-bold">
					{$t('roles.title')}
				</h3>
				<Button
					variant="ghost"
					size="icon"
					class="bg-accent text-primary hover:bg-accent/70 h-8 w-8 rounded-md"
					onclick={handleNewRole}
				>
					<Plus class="h-4 w-4" />
				</Button>
			</div>

			<div
				class="scrollbar-hide flex gap-2 overflow-x-auto p-3 lg:flex-1 lg:flex-col lg:overflow-y-auto lg:overflow-x-hidden"
			>
				{#each roles as role (role.ID)}
					<button
						class={cn(
							'group relative flex min-w-[160px] flex-col gap-1 rounded-md p-3 text-left transition-colors duration-200 lg:w-full lg:min-w-0',
							selectedRole?.ID === role.ID
								? 'border-primary/40 bg-accent border shadow-xs'
								: 'border-border hover:bg-muted border'
						)}
						onclick={() => selectRole(role)}
					>
						<div class="flex w-full items-center justify-between">
							<span
								class={cn(
									'font-semibold transition-colors',
									selectedRole?.ID === role.ID
										? 'text-accent-foreground'
										: 'text-foreground group-hover:text-foreground'
								)}
							>
								{role.ID === 0 ? $t('roles.create_new') : $t(`roles.names.${role.Name.toLowerCase().replace(' ', '_')}`) !== `roles.names.${role.Name.toLowerCase().replace(' ', '_')}` ? $t(`roles.names.${role.Name.toLowerCase().replace(' ', '_')}`) : role.Name}
							</span>
							{#if role.IsSystem}
								<ShieldCheck class="text-warning h-4 w-4" />
							{/if}
						</div>
						<p class="text-muted-foreground line-clamp-1 text-xs">
							{role.Description || 'No description'}
						</p>
					</button>
				{/each}
			</div>
		</div>

		<!-- Right Column: Editor -->
		<div class="border-border bg-card flex flex-1 flex-col overflow-hidden rounded-md border shadow-xs">
			{#if selectedRole}
				<div in:fade={{ duration: 300 }} class="flex h-full flex-col">
					<!-- Header -->
					<div class="border-border bg-muted/40 flex flex-col gap-6 border-b px-6 py-6 md:px-8">
						<!-- Title Section -->
						<div class="flex items-center gap-5">
							<div
								class={cn(
									'flex h-16 w-16 shrink-0 items-center justify-center rounded-md border shadow-xs',
									selectedRole.IsSystem
										? 'border-warning/30 bg-warning/10 text-warning'
										: 'bg-accent text-primary border-primary/20'
								)}
							>
								{#if selectedRole.IsSystem}
									<ShieldCheck class="h-8 w-8" />
								{:else if selectedRole.ID === 0}
									<Plus class="h-8 w-8" />
								{:else}
									<Shield class="h-8 w-8" />
								{/if}
							</div>
							<div class="space-y-1 overflow-hidden">
								<h2 class="font-display text-foreground truncate text-2xl font-bold tracking-tight md:text-3xl">
									{selectedRole.ID === 0 ? $t('roles.create_new') : $t(`roles.names.${selectedRole.Name.toLowerCase().replace(' ', '_')}`) !== `roles.names.${selectedRole.Name.toLowerCase().replace(' ', '_')}` ? $t(`roles.names.${selectedRole.Name.toLowerCase().replace(' ', '_')}`) : selectedRole.Name}
								</h2>
								<div class="text-muted-foreground flex flex-wrap items-center gap-3 text-sm font-medium">
									{#if selectedRole.IsSystem}
										<span class="badge-warning">
											<AlertCircle class="h-3.5 w-3.5" /> {$t('roles.system_managed')}
										</span>
									{:else}
										<span class="badge-success">
											<Check class="h-3.5 w-3.5" /> {$t('roles.custom_role')}
										</span>
									{/if}
									<span class="bg-border hidden h-1 w-1 rounded-full sm:block"></span>
									<span class="text-muted-foreground whitespace-nowrap"
										><span class="num-mono">{selectedPermissionIds.length}</span>
										{$t('roles.active_permissions')}</span
									>
								</div>
							</div>
						</div>

						<!-- Action Buttons (Separate Row) -->
						<div class="border-border flex w-full items-center justify-end gap-3 border-t pt-4">
							<Button
								variant="ghost"
								onclick={handleReset}
								disabled={isSaving}
								class="text-muted-foreground hover:bg-muted hover:text-foreground rounded-md font-medium"
							>
								{$t('roles.reset')}
							</Button>
							{#if !selectedRole.IsSystem && selectedRole.ID !== 0}
								<Button
									variant="ghost"
									class="text-danger hover:bg-danger/10 hover:text-danger rounded-md font-medium"
									onclick={handleDelete}
								>
									<Trash2 class="mr-2 h-4 w-4" /> {$t('roles.delete')}
								</Button>
							{/if}
							<Button
								class="bg-primary text-primary-foreground rounded-md px-6 font-semibold shadow-md transition-all hover:bg-primary/90 disabled:opacity-70"
								onclick={handleSave}
								disabled={isSaving}
							>
								{#if isSaving}
									<div
										class="border-primary-foreground/40 mr-2 h-4 w-4 animate-spin rounded-full border-2 border-t-transparent"
									></div>
								{:else}
									<Save class="mr-2 h-4 w-4" />
								{/if}
								{isSaving ? $t('roles.saving') : $t('roles.save_changes')}
							</Button>
						</div>
					</div>

					<!-- Scrollable Content -->
					<div class="custom-scrollbar flex-1 overflow-y-auto p-4 md:p-8">
						<!-- Inputs -->
						<div class="mb-10 grid grid-cols-1 gap-6 md:gap-8 lg:grid-cols-2">
							<div class="space-y-3" in:fly={{ y: 20, duration: 400, delay: 100 }}>
								<Label class="text-foreground text-base font-semibold">{$t('roles.role_name')}</Label>
								<div class="group relative">
									<Input
										bind:value={editName}
										placeholder="e.g. Sales Associate"
										disabled={selectedRole.IsSystem}
										class="liquid-input text-foreground h-12 w-full rounded-md px-3 text-base placeholder:text-muted-foreground"
									/>
									{#if selectedRole.IsSystem}
										<ShieldCheck
											class="text-warning/60 absolute right-4 top-1/2 h-5 w-5 -translate-y-1/2"
										/>
									{/if}
								</div>
							</div>
							<div class="space-y-3" in:fly={{ y: 20, duration: 400, delay: 200 }}>
								<Label class="text-foreground text-base font-semibold">{$t('roles.description')}</Label>
								<Input
									bind:value={editDescription}
									placeholder="Describe the responsibilities..."
									class="liquid-input text-foreground h-12 w-full rounded-md px-3 text-base placeholder:text-muted-foreground"
								/>
							</div>
						</div>

						<!-- Permissions Grid -->
						<div class="space-y-6">
							<div
								class="border-border flex items-center gap-3 border-b pb-4"
								in:fly={{ y: 20, duration: 400, delay: 300 }}
							>
								<div class="bg-accent flex h-10 w-10 items-center justify-center rounded-md">
									<Sparkles class="text-primary h-5 w-5" />
								</div>
								<div>
									<h3 class="font-display text-foreground text-lg font-bold">{$t('roles.capabilities')}</h3>
									<p class="text-muted-foreground text-sm">{$t('roles.capabilities_desc')}</p>
								</div>
							</div>

							<div class="grid grid-cols-1 gap-6 md:grid-cols-2 2xl:grid-cols-3">
								{#each sortedGroups as group, i}
									<div
										in:fly={{ y: 30, duration: 500, delay: 400 + i * 50 }}
										class="border-border bg-card relative overflow-hidden rounded-md border p-5 shadow-xs transition-shadow duration-200 hover:shadow-md md:p-6"
									>
										<div class="relative z-10 mb-5 flex flex-wrap items-center justify-between gap-y-2">
											<h4 class="text-foreground flex items-center gap-2 text-lg font-bold">
												{$t(`permissions.groups.${group.toLowerCase()}`) !== `permissions.groups.${group.toLowerCase()}` ? $t(`permissions.groups.${group.toLowerCase()}`) : group}
											</h4>
											<div class="flex items-center gap-2">
												<Label
													for="group-{group}"
													class="text-muted-foreground hover:text-primary cursor-pointer text-xs font-semibold uppercase tracking-wider transition-colors"
												>
													{$t('roles.select_all')}
												</Label>
												<Checkbox
													id="group-{group}"
													checked={groupSelectionStates[group]}
													onCheckedChange={(v) => toggleGroup(group, !!v)}
													class="border-input data-[state=checked]:border-primary data-[state=checked]:bg-primary data-[state=checked]:text-primary-foreground h-5 w-5 rounded-md"
												/>
											</div>
										</div>

										<div class="grid grid-cols-1 gap-2.5">
											{#each groupedPermissions[group] as perm}
												<label
													class={cn(
														'relative flex cursor-pointer items-start gap-3 rounded-md border p-3 transition-colors duration-150',
														selectedPermissionIds.includes(perm.ID)
															? 'border-primary/30 bg-accent'
															: 'border-border hover:bg-muted'
													)}
												>
													<Checkbox
														checked={selectedPermissionIds.includes(perm.ID)}
														onCheckedChange={() => togglePermission(perm.ID)}
														class="border-input data-[state=checked]:border-primary data-[state=checked]:bg-primary data-[state=checked]:text-primary-foreground mt-0.5 h-4.5 w-4.5 rounded"
													/>
													<div class="space-y-0.5">
														<span
															class={cn(
																'block text-sm font-semibold transition-colors',
																selectedPermissionIds.includes(perm.ID)
																	? 'text-accent-foreground'
																	: 'text-foreground'
															)}
														>
															{$t(`permissions.names.${perm.Name.replaceAll('.', '_')}`) !== `permissions.names.${perm.Name.replaceAll('.', '_')}` ? $t(`permissions.names.${perm.Name.replaceAll('.', '_')}`) : perm.Name}
														</span>
														<span class="text-muted-foreground line-clamp-2 block text-xs leading-relaxed">
															{$t(`permissions.descriptions.${perm.Name.replaceAll('.', '_')}`) !== `permissions.descriptions.${perm.Name.replaceAll('.', '_')}` ? $t(`permissions.descriptions.${perm.Name.replaceAll('.', '_')}`) : perm.Description}
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
				<div class="flex h-full flex-col items-center justify-center p-8 text-center" in:fade>
					<div class="group relative mb-8 cursor-default">
						<div
							class="bg-card border-border relative flex h-32 w-32 items-center justify-center rounded-lg border shadow-md transition-transform duration-300 group-hover:scale-105"
						>
							<Shield class="text-muted-foreground/50 h-16 w-16 transition-colors duration-300 group-hover:text-primary" />
						</div>
					</div>
					<h3 class="font-display text-foreground text-2xl font-bold">{$t('roles.security_access_control')}</h3>
					<p class="text-muted-foreground mt-3 max-w-md text-lg">
						{$t('roles.security_desc')}
					</p>
					<Button
						class="bg-primary text-primary-foreground mt-10 h-12 rounded-md px-8 font-semibold shadow-md transition-all hover:-translate-y-0.5 hover:bg-primary/90"
						onclick={handleNewRole}
					>
						<Plus class="mr-2 h-5 w-5" /> {$t('roles.create_first')}
					</Button>
				</div>
			{/if}
		</div>
	</div>
</div>

<style>
	/* Hide scrollbar for cleaner look */
	.scrollbar-hide::-webkit-scrollbar {
		display: none;
	}
	.scrollbar-hide {
		-ms-overflow-style: none;
		scrollbar-width: none;
	}
</style>
