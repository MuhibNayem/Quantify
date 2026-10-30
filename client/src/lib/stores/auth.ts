import { writable } from 'svelte/store';
import api from '$lib/api';

interface User {
	ID: number;
	Username: string;
	FirstName?: string;
	LastName?: string;
	Email?: string;
	PhoneNumber?: string;
	Address?: string;
	Role: { Name: string; Permissions?: { Name: string }[] } | string;
	IsActive: boolean;
	Permissions?: string[];
	LegacyRole?: string;
}

interface AuthState {
	user: User | null;
	isAuthenticated: boolean;
	permissions: string[];
}

const createAuthStore = () => {
	const { subscribe, set } = writable<AuthState>(getInitialAuthState());

	function extractPermissions(user: User, explicitPermissions?: string[]): string[] {
		if (explicitPermissions && explicitPermissions.length > 0) {
			return explicitPermissions;
		}
		if (user && typeof user.Role === 'object' && Array.isArray(user.Role.Permissions)) {
			return user.Role.Permissions.map((p) => p.Name);
		}
		return [];
	}

	function getInitialAuthState(): AuthState {
		if (typeof window === 'undefined') {
			return { user: null, isAuthenticated: false, permissions: [] };
		}

		// Only non-sensitive identity is persisted. Access/refresh tokens live
		// in httpOnly cookies managed by the backend and are never readable by JS.
		const userStr = localStorage.getItem('user');
		const permissionsStr = localStorage.getItem('permissions');

		if (userStr) {
			try {
				const parsedUser: User = JSON.parse(userStr);
				let parsedPermissions: string[] = permissionsStr ? JSON.parse(permissionsStr) : [];
				if (parsedPermissions.length === 0) {
					parsedPermissions = extractPermissions(parsedUser);
				}
				return { user: parsedUser, isAuthenticated: true, permissions: parsedPermissions };
			} catch (e) {
				console.error('Failed to parse user from storage', e);
				clearAuthData();
			}
		}

		return { user: null, isAuthenticated: false, permissions: [] };
	}

	function setAuthData(user: User, permissions: string[] = []) {
		const finalPermissions = extractPermissions(user, permissions);

		if (typeof window !== 'undefined') {
			localStorage.setItem('user', JSON.stringify(user));
			localStorage.setItem('permissions', JSON.stringify(finalPermissions));
			// Remove legacy token keys from older sessions.
			localStorage.removeItem('accessToken');
			localStorage.removeItem('refreshToken');
			localStorage.removeItem('csrfToken');
		}

		set({ user, isAuthenticated: true, permissions: finalPermissions });
	}

	function clearAuthData() {
		if (typeof window !== 'undefined') {
			localStorage.removeItem('user');
			localStorage.removeItem('permissions');
			localStorage.removeItem('accessToken');
			localStorage.removeItem('refreshToken');
			localStorage.removeItem('csrfToken');
		}
		set({ user: null, isAuthenticated: false, permissions: [] });
	}

	function hasPermission(permission: string): boolean {
		let currentPermissions: string[] = [];
		subscribe((state) => {
			currentPermissions = state.permissions;
		})();
		return currentPermissions.includes(permission);
	}

	return {
		subscribe,
		login: (user: User, permissions: string[] = []) => setAuthData(user, permissions),
		/** Validate the cookie session and refresh identity/permissions after a reload. */
		verify: async (): Promise<boolean> => {
			let currentUser: User | null = null;
			subscribe((state) => {
				currentUser = state.user;
			})();
			if (!currentUser) return false;

			try {
				const { data } = await api.get('/users/me');
				setAuthData(data.user, data.permissions || []);
				return true;
			} catch (error: any) {
				if (error?.response?.status === 401) {
					clearAuthData();
					return false;
				}
				// Offline/network errors must not destroy a valid local session.
				return true;
			}
		},
		logout: async () => {
			try {
				await api.post('/users/logout');
			} catch (error) {
				console.error('Logout API call failed:', error);
			} finally {
				clearAuthData();
			}
		},
		hasPermission
	};
};

export const auth = createAuthStore();
