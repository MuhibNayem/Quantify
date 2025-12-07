// client/src/routes/purchase-orders/[id]/+page.ts
import { error } from '@sveltejs/kit';
import { replenishmentApi } from '$lib/api/resources';
import type { PageLoad } from './$types';

export const load: PageLoad = async ({ params }) => {
  try {
    const poId = Number(params.id);
    if (isNaN(poId)) {
      throw error(400, 'Invalid purchase order ID');
    }
    const purchaseOrder = await replenishmentApi.getPO(poId);
    return {
      purchaseOrder,
    };
  } catch (e: any) {
    throw error(e.response?.status || 500, e.response?.data?.error || 'Unable to load purchase order');
  }
};
