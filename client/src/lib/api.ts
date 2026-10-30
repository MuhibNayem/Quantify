// client/src/lib/api.ts
import axios from 'axios';
import { browser } from '$app/environment';
import { auth } from './stores/auth';
import { enqueue, syncQueue, updateCount, type QueuedMutation } from './offline-queue';

// Default to the same hostname the SPA was loaded from so the session cookies
// are same-site (SameSite=Lax) during local development.
const defaultHost = typeof window !== 'undefined' ? window.location.hostname : 'localhost';
export const API_BASE_URL =
	import.meta.env.VITE_API_BASE_URL || `http://${defaultHost}:8080/api/v1`;
const AUTH_SERVICE_URL = import.meta.env.VITE_AUTH_SERVICE_URL || `${API_BASE_URL}/users`;

const api = axios.create({
	baseURL: API_BASE_URL,
	// Session cookies are httpOnly; the browser sends them automatically.
	withCredentials: true
});

function readCookie(name: string): string {
	if (!browser) return '';
	const match = document.cookie.match(new RegExp(`(?:^|; )${name}=([^;]*)`));
	return match ? decodeURIComponent(match[1]) : '';
}

// Memory fallback for cross-subdomain deployments where the CSRF cookie is
// scoped to the API host. Auth responses carry the token explicitly.
let csrfTokenMemory = '';

export function setCsrfToken(token?: string) {
	csrfTokenMemory = token || '';
}

const unsafeMethods = ['POST', 'PUT', 'PATCH', 'DELETE'];

// Attach the double-submit CSRF header. The token itself is a non-secret
// cookie readable by the SPA; the access/refresh tokens are httpOnly.
api.interceptors.request.use(
	(config) => {
		if (unsafeMethods.includes(config.method?.toUpperCase() ?? '')) {
			const csrf = csrfTokenMemory || readCookie('csrf_token');
			if (csrf) config.headers['X-CSRF-Token'] = csrf;
		}
		return config;
	},
	(error) => Promise.reject(error)
);

// Offline mutation queue: persist write operations while offline and replay them on reconnect
api.interceptors.request.use(
	async (config) => {
		const method = config.method?.toLowerCase() ?? '';
		const isMutation = ['post', 'put', 'patch', 'delete'].includes(method);
		if (browser && !navigator.onLine && isMutation) {
			await enqueue({
				method: config.method!.toUpperCase() as QueuedMutation['method'],
				url: config.url!,
				data: typeof config.data === 'string' ? JSON.parse(config.data) : config.data,
				headers: config.headers as unknown as Record<string, string>
			});
			const controller = new AbortController();
			config.signal = controller.signal;
			controller.abort('queued-offline');
		}
		return config;
	},
	(error) => Promise.reject(error)
);

const isAuthEndpoint = (url?: string) =>
	Boolean(
		url &&
			(url.includes('/users/login') ||
				url.includes('/users/refresh-token') ||
				url.includes('/users/logout'))
	);

api.interceptors.response.use(
	(response) => {
		if (response.data && typeof response.data.csrfToken === 'string') {
			setCsrfToken(response.data.csrfToken);
		}
		return response;
	},
	async (error) => {
		if (axios.isCancel(error) || error?.message === 'queued-offline') {
			return Promise.resolve({
				data: { queued: true, offline: true },
				status: 202,
				statusText: 'Queued Offline',
				headers: {},
				config: error.config
			});
		}

		const originalRequest = error.config;
		if (error.response?.status === 401 && originalRequest && !originalRequest._retry) {
			if (isAuthEndpoint(originalRequest.url)) {
				return Promise.reject(error);
			}

			originalRequest._retry = true;
			try {
				// Refresh rotates both cookies and the CSRF token. JWTs never
				// enter JavaScript; the CSRF token is non-secret.
				const refresh = await axios.post(
					`${AUTH_SERVICE_URL}/refresh-token`,
					{},
					{ withCredentials: true }
				);
				setCsrfToken(refresh.data?.csrfToken);
				return api(originalRequest);
			} catch (refreshError) {
				console.error('Session refresh failed:', refreshError);
				await auth.logout();
			}
		}

		return Promise.reject(error);
	}
);

// Replay queued mutations automatically when connectivity returns
if (browser) {
	window.addEventListener('online', async () => {
		await updateCount();
		const result = await syncQueue(api);
		if (result.synced > 0) {
			window.dispatchEvent(new CustomEvent('offline-sync', { detail: result }));
		}
	});
}

export default api;
