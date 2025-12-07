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

<div
	class="relative h-[calc(100vh-140px)] w-full overflow-hidden rounded-3xl bg-white/40 shadow-2xl ring-1 ring-white/60"
>
	<!-- Soft Light Background Effects -->

	<div class="relative z-10 flex h-full flex-col gap-6 p-4 backdrop-blur-xl lg:flex-row">
		<!-- Left Column: Role List -->
		<div
			class="flex w-full shrink-0 flex-col overflow-hidden rounded-2xl border border-white/60 bg-white/60 shadow-lg backdrop-blur-md transition-all hover:bg-white/70 lg:w-80"
		>
			<div class="flex items-center justify-between border-b border-slate-100 bg-white/50 p-5">
				<h3
					class="bg-gradient-to-r from-blue-600 to-purple-600 bg-clip-text text-lg font-bold text-transparent"
				>
					Roles & Access
				</h3>
				<Button
					variant="ghost"
					size="icon"
					class="h-8 w-8 rounded-full bg-blue-50 text-blue-600 hover:bg-blue-100"
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
							'group relative flex min-w-[160px] flex-col gap-1 rounded-xl p-3 text-left transition-all duration-300 lg:w-full lg:min-w-0',
							selectedRole?.ID === role.ID
								? 'border border-blue-100 bg-gradient-to-r from-blue-50 to-purple-50 shadow-sm'
								: 'border border-transparent hover:border-white/60 hover:bg-white/50'
						)}
						onclick={() => selectRole(role)}
					>
						<div class="flex w-full items-center justify-between">
							<span
								class={cn(
									'font-semibold transition-colors',
									selectedRole?.ID === role.ID
										? 'text-blue-700'
										: 'text-slate-700 group-hover:text-slate-900'
								)}
							>
								{role.Name}
							</span>
							{#if role.IsSystem}
								<ShieldCheck class="h-4 w-4 text-amber-500" />
							{/if}
						</div>
						<p class="line-clamp-1 text-xs text-slate-500 group-hover:text-slate-600">
							{role.Description || 'No description'}
						</p>
					</button>
				{/each}
			</div>
		</div>

		<!-- Right Column: Editor -->
		<div
			class="flex flex-1 flex-col overflow-hidden rounded-2xl border border-white/60 bg-white/60 shadow-xl backdrop-blur-md"
		>
			{#if selectedRole}
				<div in:fade={{ duration: 300 }} class="flex h-full flex-col">
					<!-- Header -->
					<div
						class="flex flex-col gap-6 border-b border-slate-100 bg-white/50 px-6 py-6 transition-all duration-300 md:px-8"
					>
						<!-- Title Section -->
						<div class="flex items-center gap-5">
							<div
								class={cn(
									'flex h-16 w-16 shrink-0 items-center justify-center rounded-2xl border border-white/60 shadow-sm transition-transform duration-500 hover:rotate-6',
									selectedRole.IsSystem ? 'bg-amber-50 text-amber-600' : 'bg-blue-50 text-blue-600'
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
								<h2 class="truncate text-2xl font-bold tracking-tight text-slate-800 md:text-3xl">
									{selectedRole.ID === 0 ? 'Create New Role' : selectedRole.Name}
								</h2>
								<div class="flex flex-wrap items-center gap-3 text-sm font-medium text-slate-500">
									{#if selectedRole.IsSystem}
										<span
											class="flex items-center gap-1.5 whitespace-nowrap rounded-full bg-amber-100/50 px-2.5 py-0.5 text-amber-700"
										>
											<AlertCircle class="h-3.5 w-3.5" /> System Managed
										</span>
									{:else}
										<span
											class="flex items-center gap-1.5 whitespace-nowrap rounded-full bg-emerald-100/50 px-2.5 py-0.5 text-emerald-700"
										>
											<Check class="h-3.5 w-3.5" /> Custom Role
										</span>
									{/if}
									<span class="hidden h-1 w-1 rounded-full bg-slate-300 sm:block"></span>
									<span class="whitespace-nowrap text-slate-600"
										>{selectedPermissionIds.length} Permissions Active</span
									>
								</div>
							</div>
						</div>

						<!-- Action Buttons (Separate Row) -->
						<div
							class="flex w-full items-center justify-end gap-3 border-t border-slate-100/50 pt-4"
						>
							<Button
								variant="ghost"
								onclick={handleReset}
								disabled={isSaving}
								class="rounded-xl font-medium text-slate-500 hover:bg-slate-100 hover:text-slate-700"
							>
								Reset
							</Button>
							{#if !selectedRole.IsSystem && selectedRole.ID !== 0}
								<Button
									variant="ghost"
									class="rounded-xl font-medium text-red-600 hover:bg-red-50 hover:text-red-700"
									onclick={handleDelete}
								>
									<Trash2 class="mr-2 h-4 w-4" /> Delete
								</Button>
							{/if}
							<Button
								class="transform rounded-xl bg-gradient-to-r from-blue-600 to-indigo-600 px-6 font-semibold text-white shadow-lg shadow-blue-500/30 transition-all duration-300 hover:-translate-y-0.5 hover:from-blue-700 hover:to-indigo-700 hover:shadow-blue-500/40 active:scale-95 disabled:opacity-70"
								onclick={handleSave}
								disabled={isSaving}
							>
								{#if isSaving}
									<div
										class="mr-2 h-4 w-4 animate-spin rounded-full border-2 border-white border-t-transparent"
									></div>
								{:else}
									<Save class="mr-2 h-4 w-4" />
								{/if}
								{isSaving ? 'Saving...' : 'Save Changes'}
							</Button>
						</div>
					</div>

					<!-- Scrollable Content -->
					<div
						class="scrollbar-thin scrollbar-thumb-slate-200 scrollbar-track-transparent flex-1 overflow-y-auto p-4 md:p-8"
					>
						<!-- Inputs -->
						<div class="mb-10 grid grid-cols-1 gap-6 md:gap-8 lg:grid-cols-2">
							<div class="space-y-3" in:fly={{ y: 20, duration: 400, delay: 100 }}>
								<Label class="text-base font-semibold text-slate-700">Role Name</Label>
								<div class="group relative">
									<Input
										bind:value={editName}
										placeholder="e.g. Sales Associate"
										disabled={selectedRole.IsSystem}
										class="h-12 rounded-xl border-slate-200 bg-white/50 text-base shadow-sm transition-all focus:border-blue-500 focus:bg-white focus:ring-4 focus:ring-blue-500/10 group-hover:bg-white/80"
									/>
									{#if selectedRole.IsSystem}
										<ShieldCheck
											class="absolute right-4 top-1/2 h-5 w-5 -translate-y-1/2 text-amber-500/50"
										/>
									{/if}
								</div>
							</div>
							<div class="space-y-3" in:fly={{ y: 20, duration: 400, delay: 200 }}>
								<Label class="text-base font-semibold text-slate-700">Description</Label>
								<Input
									bind:value={editDescription}
									placeholder="Describe the responsibilities..."
									class="h-12 rounded-xl border-slate-200 bg-white/50 text-base shadow-sm transition-all hover:bg-white/80 focus:border-blue-500 focus:bg-white focus:ring-4 focus:ring-blue-500/10"
								/>
							</div>
						</div>

						<!-- Permissions Grid -->
						<div class="space-y-6">
							<div
								class="flex items-center gap-3 border-b border-slate-100 pb-4"
								in:fly={{ y: 20, duration: 400, delay: 300 }}
							>
								<div class="flex h-10 w-10 items-center justify-center rounded-full bg-purple-50">
									<Sparkles class="h-5 w-5 text-purple-600" />
								</div>
								<div>
									<h3 class="text-lg font-bold text-slate-800">Capabilities</h3>
									<p class="text-sm text-slate-500">Fine-tune access controls for this role</p>
								</div>
							</div>

							<div class="grid grid-cols-1 gap-6 md:grid-cols-2 2xl:grid-cols-3">
								{#each sortedGroups as group, i}
									<div
										in:fly={{ y: 30, duration: 500, delay: 400 + i * 50 }}
										class="group/card relative overflow-hidden rounded-2xl border border-white/60 bg-gradient-to-br from-white/80 to-white/40 p-5 shadow-sm ring-1 ring-black/5 transition-all duration-300 hover:-translate-y-1 hover:border-white hover:shadow-xl hover:shadow-blue-500/5 md:p-6"
									>
										<!-- Card Glow -->
										<div
											class="absolute -right-20 -top-20 h-40 w-40 rounded-full bg-blue-100/30 blur-3xl transition-all duration-500 group-hover/card:bg-indigo-100/50"
										></div>

										<div
											class="relative z-10 mb-5 flex flex-wrap items-center justify-between gap-y-2"
										>
											<h4 class="flex items-center gap-2 text-lg font-bold text-slate-700">
												{group}
											</h4>
											<div class="flex items-center gap-2">
												<Label
													for="group-{group}"
													class="cursor-pointer text-xs font-semibold uppercase tracking-wider text-slate-400 transition-colors hover:text-blue-600"
												>
													Select All
												</Label>
												<Checkbox
													id="group-{group}"
													checked={groupSelectionStates[group]}
													onCheckedChange={(v) => toggleGroup(group, !!v)}
													class="h-5 w-5 rounded-md border-slate-300 data-[state=checked]:border-blue-600 data-[state=checked]:bg-blue-600 data-[state=checked]:text-white"
												/>
											</div>
										</div>

										<div class="grid grid-cols-1 gap-2.5">
											{#each groupedPermissions[group] as perm}
												<label
													class={cn(
														'relative flex cursor-pointer items-start gap-3 rounded-xl border p-3 transition-all duration-200',
														selectedPermissionIds.includes(perm.ID)
															? 'border-blue-200 bg-blue-50/60 shadow-inner'
															: 'border-transparent hover:border-slate-100 hover:bg-white/60 hover:shadow-sm'
													)}
												>
													<Checkbox
														checked={selectedPermissionIds.includes(perm.ID)}
														onCheckedChange={() => togglePermission(perm.ID)}
														class="h-4.5 w-4.5 mt-0.5 rounded border-slate-300 data-[state=checked]:border-blue-600 data-[state=checked]:bg-blue-600 data-[state=checked]:text-white"
													/>
													<div class="space-y-0.5">
														<span
															class={cn(
																'block text-sm font-semibold transition-colors',
																selectedPermissionIds.includes(perm.ID)
																	? 'text-blue-700'
																	: 'text-slate-600'
															)}
														>
															{perm.Name}
														</span>
														<span class="line-clamp-2 block text-xs leading-relaxed text-slate-500">
															{perm.Description}
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
							class="absolute inset-0 animate-pulse rounded-full bg-blue-100 blur-3xl transition-all duration-1000 group-hover:bg-blue-200"
						></div>
						<div
							class="relative flex h-32 w-32 items-center justify-center rounded-3xl bg-white/80 shadow-2xl backdrop-blur-xl transition-transform duration-500 group-hover:rotate-6 group-hover:scale-110"
						>
							<Shield
								class="h-16 w-16 text-slate-300 transition-colors duration-500 group-hover:text-blue-500"
							/>
						</div>
					</div>
					<h3 class="text-2xl font-bold text-slate-800">Security & Access Control</h3>
					<p class="mt-3 max-w-md text-lg text-slate-500">
						Select a role from the sidebar to configure permissions, or create a new custom role to
						delegate specific access capabilities.
					</p>
					<Button
						class="mt-10 h-12 rounded-xl bg-slate-900 px-8 font-semibold text-white shadow-xl shadow-slate-900/10 transition-all hover:-translate-y-1 hover:bg-blue-600 hover:shadow-blue-600/20"
						onclick={handleNewRole}
					>
						<Plus class="mr-2 h-5 w-5" /> Create First Role
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
