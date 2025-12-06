// client/src/lib/stores/auth.ts
import { writable } from 'svelte/store';

interface User {
  ID: number;
  Username: string;
  Role: { Name: string } | string; // Support both for safety during migration
  IsActive: boolean;
}

interface AuthState {
  accessToken: string | null;
  refreshToken: string | null;
  csrfToken: string | null;
  user: User | null;
  isAuthenticated: boolean;
}

const createAuthStore = () => {
  const { subscribe, set } = writable<AuthState>(getInitialAuthState());

  function getInitialAuthState(): AuthState {
    if (typeof window === 'undefined') {
      return { accessToken: null, refreshToken: null, csrfToken: null, user: null, isAuthenticated: false };
    }

    const accessToken = localStorage.getItem('accessToken');
    const refreshToken = localStorage.getItem('refreshToken');
    const csrfToken = localStorage.getItem('csrfToken');
    const user = localStorage.getItem('user');

    if (accessToken && refreshToken && csrfToken && user) {
      try {
        const parsedUser: User = JSON.parse(user);
        return { accessToken, refreshToken, csrfToken, user: parsedUser, isAuthenticated: true };
      } catch (e) {
        console.error("Failed to parse user from storage", e);
        clearAuthData(); // Clear corrupted data
        return { accessToken: null, refreshToken: null, csrfToken: null, user: null, isAuthenticated: false };
      }
    }

    return { accessToken: null, refreshToken: null, csrfToken: null, user: null, isAuthenticated: false };
  }

  function setAuthData(accessToken: string, refreshToken: string, csrfToken: string, user: User) {
    if (typeof window !== 'undefined') {
      localStorage.setItem('accessToken', accessToken);
      localStorage.setItem('refreshToken', refreshToken);
      localStorage.setItem('csrfToken', csrfToken);
      localStorage.setItem('user', JSON.stringify(user));
    }

    set({ accessToken, refreshToken, csrfToken, user, isAuthenticated: true });
  }

  function clearAuthData() {
    if (typeof window !== 'undefined') {
      localStorage.removeItem('accessToken');
      localStorage.removeItem('refreshToken');
      localStorage.removeItem('csrfToken');
      localStorage.removeItem('user');
    }

    set({ accessToken: null, refreshToken: null, csrfToken: null, user: null, isAuthenticated: false });
  }

  return {
    subscribe,
    login: (accessToken: string, refreshToken: string, csrfToken: string, user: User) => setAuthData(accessToken, refreshToken, csrfToken, user),
    logout: () => clearAuthData(),
    refreshTokens: async (currentRefreshToken: string): Promise<{ accessToken: string, refreshToken: string, csrfToken: string } | null> => {
      // This will be implemented later when we create the API client
      console.log("Refreshing tokens with:", currentRefreshToken);
      return null;
    }
  };
};

export const auth = createAuthStore();