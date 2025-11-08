// client/src/routes/+layout.ts
import { get } from 'svelte/store';
import { auth } from '$lib/stores/auth';
import { redirect } from '@sveltejs/kit';
import type { LayoutLoad } from './$types';

export const ssr = false;

const publicRoutes = ['/login', '/register'];

export const load: LayoutLoad = async ({ url }) => {
  const state = get(auth);

  if (!state.isAuthenticated && !publicRoutes.includes(url.pathname)) {
    throw redirect(302, '/login');
  }

  if (state.isAuthenticated && publicRoutes.includes(url.pathname)) {
    throw redirect(302, '/');
  }

  return {
    user: state.user,
    isAuthenticated: state.isAuthenticated,
  };
};
