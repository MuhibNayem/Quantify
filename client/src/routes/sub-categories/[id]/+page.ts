// client/src/routes/sub-categories/[id]/+page.ts
import { error } from '@sveltejs/kit';
import { subCategoriesApi } from '$lib/api/resources';
import type { PageLoad } from './$types';

export const load: PageLoad = async ({ params }) => {
  try {
    const subCategoryId = Number(params.id);
    if (isNaN(subCategoryId)) {
      throw error(400, 'Invalid sub-category ID');
    }
    const subCategory = await subCategoriesApi.get(subCategoryId);
    return {
      subCategory,
    };
  } catch (e: any) {
    throw error(e.response?.status || 500, e.response?.data?.error || 'Unable to load sub-category');
  }
};
