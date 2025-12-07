// client/src/routes/locations/[id]/+page.ts
import { error } from '@sveltejs/kit';
import { locationsApi } from '$lib/api/resources';
import type { PageLoad } from './$types';

export const load: PageLoad = async ({ params }) => {
  try {
    const locationId = Number(params.id);
    if (isNaN(locationId)) {
      throw error(400, 'Invalid location ID');
    }
    const location = await locationsApi.get(locationId);
    return {
      location,
    };
  } catch (e: any) {
    throw error(e.response?.status || 500, e.response?.data?.error || 'Unable to load location');
  }
};
