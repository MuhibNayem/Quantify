// client/src/routes/products/[sku]/+page.ts
import { error } from '@sveltejs/kit';
import { productsApi } from '$lib/api/resources';
import type { PageLoad } from './$types';

export const load: PageLoad = async ({ params }) => {
	try {
		// SKU is canonical and always tried first, so a purely numeric SKU is
		// not mistaken for a database ID. Numeric IDs remain a legacy fallback.
		const key = params.sku;
		let product;
		try {
			product = await productsApi.getBySku(key);
		} catch (skuError) {
			if (!/^\d+$/.test(key)) throw skuError;
			product = await productsApi.get(Number(key));
		}
		const stockHistory = await productsApi.stockHistory(product.ID);
		return {
			product,
			stockHistory
		};
	} catch (e: any) {
		throw error(e.response?.status || 500, e.response?.data?.error || 'Unable to load product');
	}
};
