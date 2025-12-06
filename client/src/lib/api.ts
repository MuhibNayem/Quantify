// client/src/lib/api.ts
import axios from 'axios';
import { get } from 'svelte/store';
import { auth } from './stores/auth';

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

// Response interceptor for handling token refresh and unauthorized errors
api.interceptors.response.use(
  (response) => {
    return response;
  },
  async (error) => {
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

export default api;
