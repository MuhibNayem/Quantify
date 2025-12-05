<script lang="ts">
	import { onMount } from 'svelte';
	import { notifications } from '$lib/stores/notifications';
	import { Popover, PopoverTrigger, PopoverContent } from '$lib/components/ui/popover';
	import { Button } from '$lib/components/ui/button';
	import { Badge } from '$lib/components/ui/badge';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import { Bell, CheckCheck, Loader2, RefreshCcw } from 'lucide-svelte';
	import { toast } from 'svelte-sonner';

	let popoverOpen = $state(false);

	const handleMarkAll = async () => {
		try {
			await notifications.markAllAsRead();
			toast.success('Notifications cleared');
		} catch (error: any) {
			const message =
				error?.response?.data?.error || error?.message || 'Unable to mark notifications';
			toast.error('Action failed', { description: message });
		}
	};

	const handleMarkRead = async (notificationId: number) => {
		try {
			await notifications.markNotificationRead(notificationId);
		} catch (error: any) {
			const message =
				error?.response?.data?.error || error?.message || 'Unable to mark notification';
			toast.error('Action failed', { description: message });
		}
	};

	import { goto } from '$app/navigation'; // Import goto

	const handleRefresh = async () => {
		try {
			await notifications.refresh();
			toast.success('Notifications synced');
		} catch (error: any) {
			const message =
				error?.response?.data?.error || error?.message || 'Unable to refresh notifications';
			toast.error('Refresh failed', { description: message });
		}
	};

	const handleNotificationClick = async (notification: any) => {
		// 1. Mark as read
		if (!notification.IsRead) {
			await handleMarkRead(notification.ID);
		}

		// 2. Deep link if route exists
		if (notification.Payload) {
			try {
				const payload = JSON.parse(notification.Payload);
				if (payload.route) {
					// Note: lowercase 'route' from JSON tag
					popoverOpen = false; // Close popover
					await goto(payload.route);
				}
			} catch (e) {
				console.error('Failed to parse notification payload', e);
			}
		}
	};

	const formatRelativeTime = (timestamp?: string | null) => {
		if (!timestamp) return '—';
		const eventDate = new Date(timestamp);
		if (Number.isNaN(eventDate.getTime())) return '—';
		const now = new Date();
		const diffMs = eventDate.getTime() - now.getTime();
		const divisions: [Intl.RelativeTimeFormatUnit, number][] = [
			['year', 1000 * 60 * 60 * 24 * 365],
			['month', 1000 * 60 * 60 * 24 * 30],
			['day', 1000 * 60 * 60 * 24],
			['hour', 1000 * 60 * 60],
			['minute', 1000 * 60],
			['second', 1000]
		];
		for (const [unit, amount] of divisions) {
			if (Math.abs(diffMs) >= amount || unit === 'second') {
				const formatter = new Intl.RelativeTimeFormat(undefined, { numeric: 'auto' });
				return formatter.format(Math.round(diffMs / amount), unit);
			}
		}
		return 'Just now';
	};

	const notificationAccent = (type: string) => {
		switch (type) {
			case 'ALERT':
				return 'bg-amber-100/70 text-amber-700 border border-amber-200';
			case 'SYSTEM':
				return 'bg-sky-100/70 text-sky-700 border border-sky-200';
			default:
				return 'bg-slate-100/70 text-slate-600 border border-slate-200';
		}
	};

	onMount(() => {
		// ensures store is initialised when component mounts
		notifications.refresh().catch(() => {
			// refresh is already invoked via auth trigger; ignore errors here
		});
	});
</script>

<Popover bind:open={popoverOpen}>
	<PopoverTrigger>
		{#snippet child({ props })}
			<Button
				{...props}
				variant="ghost"
				size="icon"
				class="relative transition-colors hover:bg-sky-200/40 dark:hover:bg-slate-700/40"
				aria-label="Notifications"
			>
				<Bell class="h-4 w-4 text-sky-700 dark:text-sky-300" />
				{#if $notifications.unreadCount > 0}
					<span
						class="absolute -right-1 -top-1 flex h-5 min-w-[1.35rem] items-center justify-center rounded-full bg-gradient-to-r from-rose-500 to-pink-500 px-1 text-[0.65rem] font-semibold text-white shadow-lg"
					>
						{$notifications.unreadCount > 9 ? '9+' : $notifications.unreadCount}
					</span>
				{/if}
				{#if $notifications.connected}
					<span
						class="absolute -bottom-1 -right-1 h-2.5 w-2.5 rounded-full bg-emerald-400 shadow-md"
					></span>
				{/if}
			</Button>
		{/snippet}
	</PopoverTrigger>
	<PopoverContent
		class="w-[360px] rounded-2xl border border-slate-100 bg-white/95 p-0 shadow-2xl backdrop-blur"
	>
		<div class="flex items-center justify-between border-b border-slate-100 px-4 py-3">
			<div>
				<p class="text-sm font-semibold text-slate-800">Notifications</p>
				<p class="text-xs text-slate-500">Real-time updates tailored for you</p>
			</div>
			<div class="flex items-center gap-2">
				<Button
					variant="ghost"
					size="icon"
					class="h-8 w-8 rounded-full hover:bg-slate-100"
					onclick={handleRefresh}
					aria-label="Refresh notifications"
				>
					{#if $notifications.loading}
						<Loader2 class="h-4 w-4 animate-spin text-slate-500" />
					{:else}
						<RefreshCcw class="h-4 w-4 text-slate-500" />
					{/if}
				</Button>
				<Button
					variant="ghost"
					size="sm"
					class="rounded-full px-3 text-slate-600 hover:bg-slate-100"
					onclick={handleMarkAll}
					disabled={$notifications.unreadCount === 0 || $notifications.loading}
				>
					<CheckCheck class="mr-1 h-4 w-4 text-emerald-600" /> Mark all
				</Button>
			</div>
		</div>

		<div class="max-h-[24rem] divide-y divide-slate-100 overflow-y-auto">
			{#if $notifications.loading && $notifications.items.length === 0}
				{#each Array(3) as _, index}
					<div class="space-y-2 p-4" aria-hidden="true">
						<Skeleton class="h-3 w-1/3" />
						<Skeleton class="h-4 w-4/5" />
						<Skeleton class="h-4 w-1/2" />
					</div>
				{/each}
			{:else if $notifications.items.length === 0}
				<div class="p-6 text-center text-sm text-slate-500">
					<p class="font-semibold text-slate-700">You're all caught up 🎉</p>
					<p class="text-xs text-slate-500">No new notifications at the moment.</p>
				</div>
			{:else}
				{#each $notifications.items as notification}
					<button
						type="button"
						class={`w-full p-4 text-left transition-colors hover:bg-slate-100 ${
							notification.IsRead ? 'bg-white' : 'bg-slate-50/80'
						}`}
						onclick={() => handleNotificationClick(notification)}
					>
						<div class="flex items-start justify-between gap-2">
							<div class="flex items-center gap-2">
								<Badge class={notificationAccent(notification.Type)}>{notification.Type}</Badge>
								<span class="text-xs text-slate-500"
									>{formatRelativeTime(notification.TriggeredAt)}</span
								>
							</div>
							{#if !notification.IsRead}
								<span class="mt-1 h-2 w-2 rounded-full bg-emerald-400"></span>
							{/if}
						</div>
						<p class="mt-2 text-sm font-semibold text-slate-800">{notification.Title}</p>
						<p class="line-clamp-2 text-sm text-slate-600">{notification.Message}</p>
					</button>
				{/each}
			{/if}
		</div>

		{#if $notifications.error}
			<div class="border-t border-rose-100 bg-rose-50 px-4 py-2 text-xs text-rose-600">
				{$notifications.error}
			</div>
		{/if}
	</PopoverContent>
</Popover>
