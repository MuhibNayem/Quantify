// client/src/routes/products/[id]/+page.ts
import { error } from '@sveltejs/kit';
import { productsApi } from '$lib/api/resources';
import type { PageLoad } from './$types';

export const load: PageLoad = async ({ params }) => {
  try {
    const productId = Number(params.id);
    if (isNaN(productId)) {
      throw error(400, 'Invalid product ID');
    }
    const product = await productsApi.get(productId);
    const stockHistory = await productsApi.stockHistory(productId);
    return {
      product,
      stockHistory,
    };
  } catch (e: any) {
    throw error(e.response?.status || 500, e.response?.data?.error || 'Unable to load product');
  }
};
