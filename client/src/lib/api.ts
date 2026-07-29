// client/src/lib/api.ts
import axios from 'axios';
import { get } from 'svelte/store';
import { browser } from '$app/environment';
import { auth } from './stores/auth';
import { enqueue, syncQueue, updateCount, type QueuedMutation } from './offline-queue';

export const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://127.0.0.1:8080/api/v1';
const AUTH_SERVICE_URL = import.meta.env.VITE_AUTH_SERVICE_URL || 'http://127.0.0.1:8080/api/v1/users';

const api = axios.create({
  baseURL: API_BASE_URL,
});

// Request interceptor for attaching access token and CSRF token
api.interceptors.request.use(
  async (config) => {
    const { accessToken, csrfToken } = get(auth);
    if (accessToken) {
      config.headers.Authorization = `Bearer ${accessToken}`;
    }
    if (csrfToken && ['POST', 'PUT', 'PATCH', 'DELETE'].includes(config.method?.toUpperCase() ?? '')) {
      config.headers['X-CSRF-Token'] = csrfToken;
    }
    return config;
  },
  (error) => {
    return Promise.reject(error);
  }
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
      // Abort the actual request; the response interceptor resolves a synthetic "queued" response
      const controller = new AbortController();
      config.signal = controller.signal;
      controller.abort('queued-offline');
    }
    return config;
  },
  (error) => {
    return Promise.reject(error);
  }
);

// Response interceptor for handling token refresh and unauthorized errors
api.interceptors.response.use(
  (response) => {
    return response;
  },
  async (error) => {
    // Requests aborted because they were queued offline: resolve a synthetic marker, not an error
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
    // If error is 401 Unauthorized and not a refresh token request
    if (error.response?.status === 401 && !originalRequest._retry) {
      originalRequest._retry = true; // Mark request as retried

      const { refreshToken } = get(auth);
      if (refreshToken) {
        try {
          const refreshResponse = await axios.post(`${AUTH_SERVICE_URL}/refresh-token`, { refreshToken });
          const { accessToken: newAccessToken, refreshToken: newRefreshToken, csrfToken: newCsrfToken } = refreshResponse.data;

          // Update auth store with new tokens
          const currentUser = get(auth).user;
          if (currentUser) {
            auth.login(newAccessToken, newRefreshToken, newCsrfToken, currentUser);
          }

          // Retry the original request with the new access token
          originalRequest.headers.Authorization = `Bearer ${newAccessToken}`;
          return api(originalRequest);
        } catch (refreshError) {
          console.error('Token refresh failed:', refreshError);
          auth.logout(); // Clear auth data and redirect to login
          // Optionally redirect to login page here if not handled by a router guard
        }
      }
    } else if (error.response?.status === 401) {
      auth.logout(); // No refresh token or refresh failed, clear auth data and redirect to login
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
