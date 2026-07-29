<script lang="ts">
	import {
		Command,
		CommandInput,
		CommandList,
		CommandEmpty,
		CommandGroup,
		CommandItem
	} from '$lib/components/ui/command';
	import { goto } from '$app/navigation';
	import { auth } from '$lib/stores/auth';
	import { toggleMode } from 'mode-watcher';
	import { getRecent } from '$lib/recent';
	import {
		LayoutDashboard,
		ShoppingCart,
		Package,
		ClipboardList,
		BarChart3,
		Bell,
		Settings,
		Plus,
		MoonStar,
		LogOut,
		History
	} from 'lucide-svelte';

	let { open = $bindable(false) } = $props();

	let recent = $state(getRecent());

	$effect(() => {
		if (open) recent = getRecent();
	});

	const navigationCommands = [
		{ label: 'Go to Dashboard', shortcut: 'G D', icon: LayoutDashboard, action: () => goto('/') },
		{ label: 'Go to POS', shortcut: 'G P', icon: ShoppingCart, action: () => goto('/pos') },
		{ label: 'Go to Catalog', shortcut: 'G C', icon: Package, action: () => goto('/catalog') },
		{ label: 'Go to Orders', shortcut: 'G O', icon: ClipboardList, action: () => goto('/orders') },
		{ label: 'Go to Reports', shortcut: 'G R', icon: BarChart3, action: () => goto('/reports') },
		{ label: 'Go to Alerts', icon: Bell, action: () => goto('/alerts') },
		{ label: 'Go to Settings', icon: Settings, action: () => goto('/settings') }
	];

	const actionCommands = [
		{ label: 'New Product', icon: Plus, action: () => goto('/catalog?tab=products&new=1') },
		{ label: 'Toggle Dark Mode', icon: MoonStar, action: () => toggleMode() },
		{ label: 'Logout', icon: LogOut, action: () => auth.logout() }
	];

	function runCommand(cmd: { action: () => void }) {
		open = false;
		cmd.action();
	}
</script>

{#if open}
	<!-- svelte-ignore a11y_interactive_supports_focus -->
	<div
		class="fixed inset-0 z-50 flex items-start justify-center bg-foreground/40 pt-[15vh]"
		role="dialog"
		aria-modal="true"
		onclick={() => (open = false)}
		onkeydown={(e) => e.key === 'Escape' && (open = false)}
	>
		<!-- svelte-ignore a11y_click_events_have_key_events a11y_no_noninteractive_element_interactions -->
		<div
			class="w-full max-w-lg overflow-hidden rounded-lg border border-border bg-card shadow-lg"
			onclick={(e) => e.stopPropagation()}
			role="document"
		>
			<Command class="w-full bg-transparent">
				<CommandInput placeholder="Type a command or search…" />
				<CommandList class="max-h-72 overflow-y-auto">
					<CommandEmpty>No results found.</CommandEmpty>
					{#if recent.length > 0}
						<CommandGroup heading="Recent">
							{#each recent as item}
								<CommandItem onSelect={() => runCommand({ action: () => goto(item.href) })}>
									<History class="h-4 w-4 text-muted-foreground" />
									<span class="flex-1">{item.label}</span>
									<span
										class="text-[0.65rem] font-medium text-muted-foreground capitalize">{item.type}</span
									>
								</CommandItem>
							{/each}
						</CommandGroup>
					{/if}
					<CommandGroup heading="Navigate">
						{#each navigationCommands as cmd}
							<CommandItem onSelect={() => runCommand(cmd)}>
								<cmd.icon class="h-4 w-4 text-muted-foreground" />
								<span class="flex-1">{cmd.label}</span>
								{#if cmd.shortcut}
									<kbd
										class="rounded-md border border-border bg-muted px-1.5 py-0.5 text-[0.65rem] font-medium text-muted-foreground"
									>
										{cmd.shortcut}
									</kbd>
								{/if}
							</CommandItem>
						{/each}
					</CommandGroup>
					<CommandGroup heading="Actions">
						{#each actionCommands as cmd}
							<CommandItem onSelect={() => runCommand(cmd)}>
								<cmd.icon class="h-4 w-4 text-muted-foreground" />
								<span class="flex-1">{cmd.label}</span>
							</CommandItem>
						{/each}
					</CommandGroup>
				</CommandList>
			</Command>
		</div>
	</div>
{/if}
