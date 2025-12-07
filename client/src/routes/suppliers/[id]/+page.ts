// client/src/routes/suppliers/[id]/+page.ts
import { error } from '@sveltejs/kit';
import { suppliersApi } from '$lib/api/resources';
import type { PageLoad } from './$types';

export const load: PageLoad = async ({ params }) => {
  try {
    const supplierId = Number(params.id);
    if (isNaN(supplierId)) {
      throw error(400, 'Invalid supplier ID');
    }
    const supplier = await suppliersApi.get(supplierId);
    const performance = await suppliersApi.performance(supplierId);
    return {
      supplier,
      performance,
    };
  } catch (e: any) {
    throw error(e.response?.status || 500, e.response?.data?.error || 'Unable to load supplier');
  }
};
