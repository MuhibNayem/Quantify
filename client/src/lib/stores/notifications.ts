import { browser } from '$app/environment';
import { notificationsApi } from '$lib/api/resources';
import type { Notification } from '$lib/types';
import { get, writable } from 'svelte/store';
import { auth } from './auth';

type NotificationStoreState = {
	items: Notification[];
	unreadCount: number;
	loading: boolean;
	connected: boolean;
	error: string | null;
};

const deriveDefaultWsUrl = () => {
	const explicit = import.meta.env.VITE_WS_BASE_URL;
	if (explicit) return explicit;

	const apiBase = import.meta.env.VITE_API_BASE_URL || 'http://127.0.0.1:8080/api/v1';
	try {
		const url = new URL(apiBase);
		url.pathname = '/ws';
		url.search = '';
		url.hash = '';
		url.protocol = url.protocol === 'https:' ? 'wss:' : 'ws:';
		return url.toString();
	} catch {
		// fallback to localhost secure guess
		return 'ws://127.0.0.1:8080/ws';
	}
};

const WS_BASE_URL = deriveDefaultWsUrl();
const MAX_NOTIFICATIONS = 25;

const initialState: NotificationStoreState = {
	items: [],
	unreadCount: 0,
	loading: false,
	connected: false,
	error: null,
};

function createNotificationStore() {
	const { subscribe, set, update } = writable<NotificationStoreState>(initialState);

	let currentUserId: number | null = null;
	let token: string | null = null;
	let socket: WebSocket | null = null;
	let reconnectHandle: ReturnType<typeof setTimeout> | null = null;
	let initializing = false;

	const cleanupSocket = () => {
		if (reconnectHandle) {
			clearTimeout(reconnectHandle);
			reconnectHandle = null;
		}
		if (socket) {
			socket.onclose = null;
			socket.onerror = null;
			socket.onopen = null;
			socket.onmessage = null;
			socket.close();
			socket = null;
		}
	};

	const resetState = () => {
		cleanupSocket();
		set(initialState);
	};

	const fetchNotifications = async () => {
		if (!browser || !currentUserId) return;
		if (initializing) return;
		initializing = true;
		update((state) => ({ ...state, loading: true, error: null }));
		try {
			const [items, count] = await Promise.all([
				notificationsApi.list(currentUserId, { limit: MAX_NOTIFICATIONS }),
				notificationsApi.unreadCount(currentUserId),
			]);
			update((state) => ({
				...state,
				items,
				unreadCount: count,
				loading: false,
				error: null,
			}));
		} catch (error: any) {
			const message = error?.response?.data?.error || error?.message || 'Unable to load notifications';
			update((state) => ({ ...state, loading: false, error: message }));
		} finally {
			initializing = false;
		}
	};

	const buildWsUrl = () => {
		if (!browser || !token) return null;
		try {
			const url = new URL(WS_BASE_URL, WS_BASE_URL.startsWith('ws') ? undefined : window.location.origin);
			url.searchParams.set('token', token);
			return url.toString();
		} catch {
			const separator = WS_BASE_URL.includes('?') ? '&' : '?';
			return `${WS_BASE_URL}${separator}token=${encodeURIComponent(token)}`;
		}
	};

	const connectSocket = () => {
		if (!browser || !currentUserId || !token) return;
		const wsUrl = buildWsUrl();
		console.log({ wsUrl })
		if (!wsUrl) return;

		cleanupSocket();

		socket = new WebSocket(wsUrl);
		socket.onopen = (e) => {
			console.log('opened ws connection', e)
			update((state) => ({ ...state, connected: true }));
		};
		socket.onclose = (e) => {
			console.log('closed ws connection', e)
			update((state) => ({ ...state, connected: false }));
			if (browser && currentUserId && token) {
				reconnectHandle = setTimeout(connectSocket, 5000);
			}
		};
		socket.onerror = (e) => {
			console.log('error on ws connection', e)
			update((state) => ({ ...state, connected: false }));
		};
		socket.onmessage = (event) => {
			try {
				const payload = JSON.parse(event.data);
				if (browser && payload?.event === 'BULK_JOB_STATUS') {
					window.dispatchEvent(new CustomEvent('bulk-job-status', { detail: payload }));
					return;
				}
				// Handle Return Updates (Real-time sync)
				if (browser && (payload?.event === 'RETURN_UPDATED' || payload?.event === 'RETURN_REQUESTED')) {
					window.dispatchEvent(new CustomEvent('return-updated', { detail: payload }));
					// We might also want to notify the user via a toast/notification, 
					// but let's assume the event dispatch is enough for the page to refresh.
					// If it's a notification object too, we fall through?
					// The payload from backend is simple gin.H, so it lacks Notification fields.
					// So we should return here.
					return;
				}

				const notification = payload as Notification;
				update((state) => {
					const nextItems = [notification, ...state.items.filter((item) => item.ID !== notification.ID)].slice(
						0,
						MAX_NOTIFICATIONS
					);
					const unreadCount = notification.IsRead ? state.unreadCount : state.unreadCount + 1;
					console.log({ nextItems, unreadCount })
					return { ...state, items: nextItems, unreadCount };
				});
			} catch (error) {
				console.error('Failed to parse websocket payload', error);
			}
		};
	};

	const ensureSession = () => {
		if (!browser) return;
		const authState = get(auth);
		const nextUserId = authState.user?.ID ?? null;
		const nextToken = authState.accessToken ?? null;
		const hasPermission = authState.permissions?.includes('notifications.read');

		if (!nextUserId || !nextToken || !hasPermission) {
			currentUserId = null;
			token = null;
			resetState();
			return;
		}

		const userChanged = nextUserId !== currentUserId;
		const tokenChanged = nextToken !== token;

		currentUserId = nextUserId;
		token = nextToken;

		if (userChanged || tokenChanged) {
			fetchNotifications();
			connectSocket();
		}
	};

	if (browser) {
		auth.subscribe(() => {
			ensureSession();
		});
		ensureSession();
	}

	const markNotificationRead = async (notificationId: number) => {
		if (!currentUserId) throw new Error('User not authenticated');
		await notificationsApi.markRead(currentUserId, notificationId);
		update((state) => {
			let decrement = 0;
			const items = state.items.map((item) => {
				if (item.ID === notificationId && !item.IsRead) {
					decrement = 1;
					return { ...item, IsRead: true, ReadAt: new Date().toISOString() };
				}
				return item;
			});
			return { ...state, items, unreadCount: Math.max(0, state.unreadCount - decrement) };
		});
	};

	const markAllAsRead = async () => {
		if (!currentUserId) throw new Error('User not authenticated');
		await notificationsApi.markAllRead(currentUserId);
		update((state) => ({
			...state,
			items: state.items.map((item) =>
				item.IsRead ? item : { ...item, IsRead: true, ReadAt: new Date().toISOString() }
			),
			unreadCount: 0,
		}));
	};

	const refresh = async () => {
		await fetchNotifications();
	};

	return {
		subscribe,
		markNotificationRead,
		markAllAsRead,
		refresh,
	};
}

export const notifications = createNotificationStore();
