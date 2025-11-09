import api from '$lib/api';
import type {
	Alert,
	Batch,
	BulkImportJob,
	Category,
	SubCategory,
	DemandForecast,
	Location,
	Product,
	PurchaseOrder,
	ReorderSuggestion,
	Supplier,
	UserSummary,
	PaginatedProducts,
} from '$lib/types';

export const productsApi = {
	list: async (page: number = 1, limit: number = 100) => (await api.get<PaginatedProducts>(`/products?page=${page}&limit=${limit}`)).data,
	get: async (id: number) => (await api.get<Product>(`/products/${id}`)).data,
	create: async (payload: Record<string, unknown>) => (await api.post<Product>('/products', payload)).data,
	update: async (id: number, payload: Record<string, unknown>) => (await api.put<Product>(`/products/${id}`, payload)).data,
	remove: async (id: number) => (await api.delete(`/products/${id}`)).data,
	stock: async (id: number, params?: Record<string, string | number | undefined>) =>
		(await api.get(`/products/${id}/stock`, { params })).data as { productId: number; currentQuantity: number; batches: Batch[] },
	createBatch: async (id: number, payload: Record<string, unknown>) => (await api.post<Batch>(`/products/${id}/stock/batches`, payload)).data,
	adjustStock: async (id: number, payload: Record<string, unknown>) =>
		(await api.post(`/products/${id}/stock/adjustments`, payload)).data,
};

export const categoriesApi = {
	list: async () => (await api.get<Category[]>('/categories')).data,
	create: async (payload: { name: string }) => (await api.post<Category>('/categories', payload)).data,
	update: async (id: number, payload: { name: string }) => (await api.put<Category>(`/categories/${id}`, payload)).data,
	remove: async (id: number) => (await api.delete(`/categories/${id}`)).data,
};

export const subCategoriesApi = {
	list: async (categoryId: number) => (await api.get<SubCategory[]>(`/categories/${categoryId}/sub-categories`)).data,
	create: async (data: { name: string; }, categoryId: number) => (await api.post<SubCategory>(`/categories/${categoryId}/sub-categories`, data)).data,
	update: async (id: number, data: { name: string }) => (await api.put<SubCategory>(`/sub-categories/${id}`, data)).data,
	remove: async (id: number) => (await api.delete(`/sub-categories/${id}`)).data,
};

export const suppliersApi = {
	list: async () => (await api.get<Supplier[]>('/suppliers')).data,
	create: async (payload: Record<string, unknown>) => (await api.post<Supplier>('/suppliers', payload)).data,
	update: async (id: number, payload: Record<string, unknown>) => (await api.put<Supplier>(`/suppliers/${id}`, payload)).data,
	remove: async (id: number) => (await api.delete(`/suppliers/${id}`)).data,
};

export const locationsApi = {
	list: async () => (await api.get<Location[]>('/locations')).data,
	create: async (payload: Record<string, unknown>) => (await api.post<Location>('/locations', payload)).data,
	update: async (id: number, payload: Record<string, unknown>) => (await api.put<Location>(`/locations/${id}`, payload)).data,
	remove: async (id: number) => (await api.delete(`/locations/${id}`)).data,
};

export const barcodeApi = {
	lookup: async (barcode: string, params?: Record<string, string>) =>
		(await api.get<Product>('/barcode/lookup', { params: { barcode, ...params } })).data,
	generate: async (params: Record<string, string | number>) =>
		(await api.get('/barcode/generate', { params, responseType: 'blob' })).data as Blob,
};

export const inventoryApi = {
	transfer: async (payload: Record<string, unknown>) => (await api.post('/inventory/transfers', payload)).data,
};

export const replenishmentApi = {
	generateForecast: async (payload: Record<string, unknown>) => (
		await api.post('/replenishment/forecast/generate', payload)
	).data,
	getForecast: async (id: number) => (await api.get<DemandForecast>(`/replenishment/forecast/${id}`)).data,
	listSuggestions: async (params?: Record<string, unknown>) =>
		(await api.get<ReorderSuggestion[]>('/replenishment/suggestions', { params })).data,
	createPOFromSuggestion: async (suggestionId: number) =>
		(await api.post<PurchaseOrder>(`/replenishment/suggestions/${suggestionId}/create-po`)).data,
	approvePO: async (poId: number) => (await api.post(`/replenishment/purchase-orders/${poId}/approve`)).data,
	sendPO: async (poId: number) => (await api.post(`/replenishment/purchase-orders/${poId}/send`)).data,
	updatePO: async (poId: number, payload: Record<string, unknown>) =>
		(await api.put(`/replenishment/purchase-orders/${poId}`, payload)).data,
	receivePO: async (poId: number, payload: Record<string, unknown>) =>
		(await api.post(`/replenishment/purchase-orders/${poId}/receive`, payload)).data,
	cancelPO: async (poId: number) => (await api.post(`/replenishment/purchase-orders/${poId}/cancel`)).data,
};

export const reportsApi = {
	salesTrends: async (payload: Record<string, unknown>) => (
		await api.post('/reports/sales-trends', payload)
	).data,
	inventoryTurnover: async (payload: Record<string, unknown>) => (
		await api.post('/reports/inventory-turnover', payload)
	).data,
	profitMargin: async (payload: Record<string, unknown>) => (
		await api.post('/reports/profit-margin', payload)
	).data,
};

export const alertsApi = {
	list: async (params?: Record<string, unknown>) => (await api.get<Alert[]>('/alerts', { params })).data,
	resolve: async (id: number) => (await api.patch(`/alerts/${id}/resolve`)).data,
	updateProductSettings: async (productId: number, payload: Record<string, unknown>) =>
		(await api.put(`/alerts/products/${productId}/settings`, payload)).data,
	updateUserNotifications: async (userId: number, payload: Record<string, unknown>) =>
		(await api.put(`/alerts/users/${userId}/notification-settings`, payload)).data,
	triggerCheck: async () => (await api.post('/alerts/check')).data,
};

export const bulkApi = {
	downloadTemplate: async () => (await api.get('/bulk/products/template', { responseType: 'blob' })).data as Blob,
	uploadImport: async (formData: FormData) =>
		(await api.post<BulkImportJob>('/bulk/products/import', formData)).data,
	status: async (jobId: string) => (await api.get<BulkImportJob>(`/bulk/products/import/${jobId}/status`)).data,
	confirm: async (jobId: string) => (await api.post(`/bulk/products/import/${jobId}/confirm`)).data,
	exportProducts: async (params: Record<string, string | number | undefined>) =>
		(await api.get('/bulk/products/export', { params, responseType: 'blob' })).data as Blob,
};

export const usersApi = {
	list: async (params?: Record<string, unknown>) => (await api.get<UserSummary[]>('/users', { params })).data,
	get: async (id: number) => (await api.get<UserSummary>(`/users/${id}`)).data,
	update: async (id: number, payload: Record<string, unknown>) => (await api.put(`/users/${id}`, payload)).data,
	approve: async (id: number) => (await api.put(`/users/${id}/approve`)).data,
	remove: async (id: number) => (await api.delete(`/users/${id}`)).data,
};
