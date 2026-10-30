<script lang="ts">
	import { onMount } from 'svelte';
	import { notifications } from '$lib/stores/notifications';
	import { auth } from '$lib/stores/auth';
	import { notificationsApi } from '$lib/api/resources';
	import { optimistic } from '$lib/optimistic';
	import { Popover, PopoverTrigger, PopoverContent } from '$lib/components/ui/popover';
	import { Button } from '$lib/components/ui/button';
	import { Badge } from '$lib/components/ui/badge';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import { Bell, CheckCheck, Loader2, RefreshCcw } from 'lucide-svelte';
	import { toast } from 'svelte-sonner';
	import TimeAgo from '$lib/components/ui/time-ago.svelte';
	import EmptyState from '$lib/components/ui/empty-state.svelte';

	let popoverOpen = $state(false);

	const handleMarkAll = async () => {
		if (!$auth.permissions.includes('notifications.write')) {
			toast.error('Permission denied', { description: 'You cannot manage notifications.' });
			return;
		}
		const userId = $auth.user?.ID;
		if (!userId) return;

		const previousItems = $notifications.items.map((i) => ({ ...i }));
		const previousUnread = $notifications.unreadCount;

		try {
			await optimistic(
				() => {
					notifications.update((state) => ({
						...state,
						items: state.items.map((item) =>
							item.IsRead ? item : { ...item, IsRead: true, ReadAt: new Date().toISOString() }
						),
						unreadCount: 0
					}));
				},
				() => {
					notifications.update((state) => ({
						...state,
						items: previousItems,
						unreadCount: previousUnread
					}));
				},
				() => notificationsApi.markAllRead(userId),
				{ success: 'Notifications cleared', error: 'Failed to mark notifications' }
			);
		} catch {
			// optimistic() already rolled back and toasted
		}
	};

	const handleMarkRead = async (notificationId: number) => {
		if (!$auth.permissions.includes('notifications.write')) return;
		const userId = $auth.user?.ID;
		if (!userId) return;

		const target = $notifications.items.find((n) => n.ID === notificationId);
		const wasUnread = target ? !target.IsRead : false;

		try {
			await optimistic(
				() => {
					notifications.update((state) => {
						const items = state.items.map((item) =>
							item.ID === notificationId && !item.IsRead
								? { ...item, IsRead: true, ReadAt: new Date().toISOString() }
								: item
						);
						return {
							...state,
							items,
							unreadCount: Math.max(0, state.unreadCount - (wasUnread ? 1 : 0))
						};
					});
				},
				() => {
					if (!wasUnread) return;
					notifications.update((state) => {
						const items = state.items.map((item) =>
							item.ID === notificationId ? { ...item, IsRead: false, ReadAt: null } : item
						);
						return { ...state, items, unreadCount: state.unreadCount + 1 };
					});
				},
				() => notificationsApi.markRead(userId, notificationId),
				{ success: 'Marked as read', error: 'Failed to mark notification' }
			);
		} catch {
			// optimistic() already rolled back and toasted
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

	const notificationAccent = (type: string) => {
		switch (type) {
			case 'ALERT':
				return 'badge-warning';
			case 'SYSTEM':
				return 'badge-info';
			default:
				return 'badge-neutral';
		}
	};

	onMount(() => {
		// ensures store is initialised when component mounts
		notifications.refresh().catch(() => {
			// refresh is already invoked via auth trigger; ignore errors here
		});
	});
</script>

{#if $auth.permissions.includes('notifications.read')}
	<Popover bind:open={popoverOpen}>
		<PopoverTrigger>
			{#snippet child({ props })}
				<Button
					{...props}
					variant="ghost"
					size="icon"
					class="relative transition-colors"
					aria-label="Notifications"
				>
					<Bell class="h-4 w-4 text-muted-foreground" />
					{#if $notifications.unreadCount > 0}
						<span
							class="num absolute -top-1 -right-1 flex h-5 min-w-[1.35rem] items-center justify-center rounded-full bg-danger px-1 text-[0.65rem] font-semibold text-destructive-foreground shadow-sm"
						>
							{$notifications.unreadCount > 9 ? '9+' : $notifications.unreadCount}
						</span>
					{/if}
				</Button>
			{/snippet}
		</PopoverTrigger>
		<PopoverContent class="w-[360px] rounded-xl p-0">
			<div class="flex items-center justify-between border-b border-border px-4 py-3">
				<div>
					<p class="text-sm font-semibold text-foreground">Notifications</p>
					<p class="text-xs text-muted-foreground">Real-time updates tailored for you</p>
				</div>
				<div class="flex items-center gap-2">
					<Button
						variant="ghost"
						size="icon"
						class="h-8 w-8 rounded-md"
						onclick={handleRefresh}
						aria-label="Refresh notifications"
					>
						{#if $notifications.loading}
							<Loader2 class="h-4 w-4 animate-spin text-muted-foreground" />
						{:else}
							<RefreshCcw class="h-4 w-4 text-muted-foreground" />
						{/if}
					</Button>
					{#if $auth.permissions.includes('notifications.write')}
						<Button
							variant="ghost"
							size="sm"
							class="rounded-md px-3 text-muted-foreground"
							onclick={handleMarkAll}
							disabled={$notifications.unreadCount === 0 || $notifications.loading}
						>
							<CheckCheck class="mr-1 h-4 w-4 text-success" /> Mark all
						</Button>
					{/if}
				</div>
			</div>

			<div class="custom-scrollbar max-h-[24rem] divide-y divide-border overflow-y-auto">
				{#if $notifications.loading && $notifications.items.length === 0}
					{#each Array(3) as _, index}
						<div class="space-y-2 p-4" aria-hidden="true">
							<Skeleton class="h-3 w-1/3" />
							<Skeleton class="h-4 w-4/5" />
							<Skeleton class="h-4 w-1/2" />
						</div>
					{/each}
				{:else if $notifications.items.length === 0}
					<EmptyState
						icon={Bell}
						title="You're all caught up"
						description="No new notifications at the moment."
					/>
				{:else}
					{#each $notifications.items as notification}
						<button
							type="button"
							class={`w-full p-4 text-left transition-colors hover:bg-muted ${
								notification.IsRead ? 'bg-card' : 'bg-muted/40'
							}`}
							onclick={() => handleNotificationClick(notification)}
						>
							<div class="flex items-start justify-between gap-2">
								<div class="flex items-center gap-2">
									<Badge class={notificationAccent(notification.Type)}>{notification.Type}</Badge>
									<span class="text-xs text-muted-foreground"
										><TimeAgo date={notification.TriggeredAt} /></span
									>
								</div>
								{#if !notification.IsRead}
									<span class="mt-1 h-2 w-2 rounded-full bg-danger"></span>
								{/if}
							</div>
							<p class="mt-2 text-sm font-semibold text-foreground">{notification.Title}</p>
							<p class="line-clamp-2 text-sm text-muted-foreground">{notification.Message}</p>
						</button>
					{/each}
				{/if}
			</div>

			{#if $notifications.error}
				<div class="border-t border-danger/25 bg-danger/10 px-4 py-2 text-xs text-danger">
					{$notifications.error}
				</div>
			{/if}
		</PopoverContent>
	</Popover>
{/if}
