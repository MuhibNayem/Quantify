<script lang="ts">
	import { Check, ChevronsUpDown, Loader2, Search } from 'lucide-svelte';
	import { cn } from '$lib/utils';
	import { Button } from '$lib/components/ui/button';
	import {
		Command,
		CommandEmpty,
		CommandGroup,
		CommandInput,
		CommandItem,
		CommandList
	} from '$lib/components/ui/command';
	import { Popover, PopoverContent, PopoverTrigger } from '$lib/components/ui/popover';
	import { usersApi } from '$lib/api/resources';
	import type { UserSummary } from '$lib/types';
	import { onMount } from 'svelte';

	let {
		value = $bindable(''),
		placeholder = 'Select user...',
		className = '',
		onSelect
	} = $props<{
		value?: string;
		placeholder?: string;
		className?: string;
		onSelect?: (user: UserSummary) => void;
	}>();

	let open = $state(false);
	let users = $state<UserSummary[]>([]);
	let loading = $state(false);
	let selectedUser = $state<UserSummary | null>(null);
	let searchQuery = $state('');
	let debounceTimer: NodeJS.Timeout;

	async function fetchUsers(search: string = '') {
		loading = true;
		try {
			// Assuming list accepts generic params which might be handled as query params
			const res = await usersApi.list({ search, limit: 50 });
			users = res.users || [];
		} catch (e) {
			console.error('Failed to fetch users', e);
			users = [];
		} finally {
			loading = false;
		}
	}

	function handleSearch(val: string) {
		searchQuery = val;
		clearTimeout(debounceTimer);
		debounceTimer = setTimeout(() => {
			fetchUsers(val);
		}, 300);
	}

	// Initial fetch
	onMount(() => {
		fetchUsers();
	});

	// Update selected label when value changes externally
	$effect(() => {
		if (value && !selectedUser) {
			const found = users.find((u) => String(u.ID) === String(value));
			if (found) selectedUser = found;
			else if (value) {
				// Fetch single user to get label
				usersApi
					.get(Number(value))
					.then((u) => (selectedUser = u))
					.catch(() => {});
			}
		}
	});

	function getUserLabel(user: UserSummary) {
		if (user.FirstName && user.LastName) {
			return `${user.FirstName} ${user.LastName} (${user.Username})`;
		}
		return user.Username;
	}
</script>

<Popover bind:open>
	<PopoverTrigger>
		{#snippet child({ props })}
			<Button
				variant="outline"
				role="combobox"
				aria-expanded={open}
				class={cn(
					'h-12 w-full justify-between rounded-2xl border-white/60 bg-white/50 shadow-sm backdrop-blur-md transition-all hover:bg-white/80',
					className
				)}
				{...props}
			>
				{#if selectedUser}
					<span class="truncate">{getUserLabel(selectedUser)}</span>
				{:else}
					<span class="text-muted-foreground">{placeholder}</span>
				{/if}
				<ChevronsUpDown class="ml-2 h-4 w-4 shrink-0 opacity-50" />
			</Button>
		{/snippet}
	</PopoverTrigger>
	<PopoverContent
		class="z-50 w-[300px] rounded-2xl border border-white/60 bg-white/90 p-0 shadow-[0_40px_100px_-20px_rgba(50,60,90,0.2),inset_0_1px_0_0_rgba(255,255,255,0.8)] outline-none backdrop-blur-3xl backdrop-saturate-[180%]"
		align="start"
	>
		<Command shouldFilter={false}>
			<div class="flex items-center border-b border-white/20 px-4 py-2" data-cmdk-input-wrapper="">
				<Search class="mr-2 h-4 w-4 shrink-0 opacity-50" />
				<input
					class="flex h-12 w-full rounded-md bg-transparent py-3 text-base font-medium text-slate-900 outline-none placeholder:text-slate-500/80 disabled:cursor-not-allowed disabled:opacity-50"
					placeholder="Search by name..."
					value={searchQuery}
					oninput={(e) => handleSearch(e.currentTarget.value)}
				/>
			</div>
			<CommandList>
				{#if loading}
					<div class="text-muted-foreground flex justify-center py-6 text-center text-sm">
						<Loader2 class="mr-2 h-4 w-4 animate-spin" /> Loading...
					</div>
				{:else if users.length === 0}
					<CommandEmpty>No user found.</CommandEmpty>
				{:else}
					<CommandGroup>
						{#each users as user}
							<CommandItem
								value={String(user.ID)}
								onSelect={() => {
									value = String(user.ID);
									selectedUser = user;
									open = false;
									onSelect?.(user);
								}}
							>
								<Check
									class={cn(
										'mr-2 h-4 w-4',
										value === String(user.ID) ? 'opacity-100' : 'opacity-0'
									)}
								/>
								<div class="flex flex-col">
									<span class="font-medium text-slate-900">{getUserLabel(user)}</span>
									<span class="text-xs text-slate-500">{user.Email}</span>
								</div>
							</CommandItem>
						{/each}
					</CommandGroup>
				{/if}
			</CommandList>
		</Command>
	</PopoverContent>
</Popover>
