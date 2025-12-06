import { writable } from 'svelte/store';

interface User {
  ID: number;
  Username: string;
  Role: { Name: string; Permissions?: { Name: string }[] } | string;
  IsActive: boolean;
  Permissions?: string[];
  LegacyRole?: string;
}

interface AuthState {
  accessToken: string | null;
  refreshToken: string | null;
  csrfToken: string | null;
  user: User | null;
  isAuthenticated: boolean;
  permissions: string[];
}

const createAuthStore = () => {
  const { subscribe, set, update } = writable<AuthState>(getInitialAuthState());

  function extractPermissions(user: User, explicitPermissions?: string[]): string[] {
    // If explicit permissions are provided and not empty, use them
    if (explicitPermissions && explicitPermissions.length > 0) {
      return explicitPermissions;
    }

    // Otherwise try to extract from User object
    if (user && typeof user.Role === 'object' && Array.isArray(user.Role.Permissions)) {
      return user.Role.Permissions.map(p => p.Name);
    }

    return [];
  }

  function getInitialAuthState(): AuthState {
    if (typeof window === 'undefined') {
      return { accessToken: null, refreshToken: null, csrfToken: null, user: null, isAuthenticated: false, permissions: [] };
    }

    const accessToken = localStorage.getItem('accessToken');
    const refreshToken = localStorage.getItem('refreshToken');
    const csrfToken = localStorage.getItem('csrfToken');
    const userStr = localStorage.getItem('user');
    const permissionsStr = localStorage.getItem('permissions');

    if (accessToken && refreshToken && csrfToken && userStr) {
      try {
        const parsedUser: User = JSON.parse(userStr);
        let parsedPermissions: string[] = permissionsStr ? JSON.parse(permissionsStr) : [];

        // Fallback: extract from user if permissions are empty
        if (parsedPermissions.length === 0) {
          parsedPermissions = extractPermissions(parsedUser);
        }

        return { accessToken, refreshToken, csrfToken, user: parsedUser, isAuthenticated: true, permissions: parsedPermissions };
      } catch (e) {
        console.error("Failed to parse user from storage", e);
        clearAuthData();
        return { accessToken: null, refreshToken: null, csrfToken: null, user: null, isAuthenticated: false, permissions: [] };
      }
    }

    return { accessToken: null, refreshToken: null, csrfToken: null, user: null, isAuthenticated: false, permissions: [] };
  }

  function setAuthData(accessToken: string, refreshToken: string, csrfToken: string, user: User, permissions: string[] = []) {
    // Consolidate permission extraction logic
    const finalPermissions = extractPermissions(user, permissions);

    if (typeof window !== 'undefined') {
      localStorage.setItem('accessToken', accessToken);
      localStorage.setItem('refreshToken', refreshToken);
      localStorage.setItem('csrfToken', csrfToken);
      localStorage.setItem('user', JSON.stringify(user));
      localStorage.setItem('permissions', JSON.stringify(finalPermissions));
    }

    set({ accessToken, refreshToken, csrfToken, user, isAuthenticated: true, permissions: finalPermissions });
  }

  function clearAuthData() {
    if (typeof window !== 'undefined') {
      localStorage.removeItem('accessToken');
      localStorage.removeItem('refreshToken');
      localStorage.removeItem('csrfToken');
      localStorage.removeItem('user');
      localStorage.removeItem('permissions');
    }

    set({ accessToken: null, refreshToken: null, csrfToken: null, user: null, isAuthenticated: false, permissions: [] });
  }

  function hasPermission(permission: string): boolean {
    let currentPermissions: string[] = [];
    subscribe(state => {
      currentPermissions = state.permissions;
    })();
    return currentPermissions.includes(permission);
  }

  return {
    subscribe,
    login: (accessToken: string, refreshToken: string, csrfToken: string, user: User, permissions: string[]) => setAuthData(accessToken, refreshToken, csrfToken, user, permissions),
    logout: () => clearAuthData(),
    hasPermission,
    refreshTokens: async (currentRefreshToken: string): Promise<{ accessToken: string, refreshToken: string, csrfToken: string } | null> => {
      // This will be implemented later when we create the API client
      console.log("Refreshing tokens with:", currentRefreshToken);
      return null;
    }
  };
};

export const auth = createAuthStore();