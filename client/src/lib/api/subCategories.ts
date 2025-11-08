import api  from '$lib/api';
import type { SubCategory } from '$lib/types';

export const subCategoriesApi = {
	list: async () => (await api.get<SubCategory[]>('/sub-categories')).data,
	create: async (data: { name: string; categoryId: number }) => (await api.post<SubCategory>('/sub-categories', data)).data,
	update: async (id: number, data: { name: string; categoryId: number }) => (await api.put<SubCategory>(`/sub-categories/${id}`, data)).data,
	remove: async (id: number) => (await api.delete(`/sub-categories/${id}`)).data,
};
