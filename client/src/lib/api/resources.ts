import api, { API_BASE_URL } from '$lib/api';
import { getCached, setCache, invalidateCache } from '$lib/cache';
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
	Promotion,
	ReorderSuggestion,
	StockAdjustment,
	Supplier,
	SupplierPerformance,
	UserSummary,
	PaginatedUsers,
	PaginatedProducts,
	PaginatedResponse,
	Notification,
	DashboardSummary,
	LoyaltyAccount,
	TimeClock,
	BasketAnalysisItem,
	PaymentRequest,
	PaymentResponse,
	PurchaseReturn,
	CreatePromotionRequest,
	UpdatePromotionRequest,
	TeamStatus,
	TeamOverview,
	WeeklySummary,
} from '$lib/types';
import type { AxiosRequestConfig } from 'axios';

async function cachedGet<T>(key: string, url: string, config?: AxiosRequestConfig): Promise<T> {
	const cached = getCached<T>(key);
	if (cached) return cached;
	const response = await api.get<T>(url, config);
	setCache(key, response.data);
	return response.data;
}

export const searchApi = {
	query: async (q: string) => (await api.get('/search', { params: { q } })).data,
};

export const dashboardApi = {
	getSummary: async () => cachedGet<DashboardSummary>('dashboard:summary', '/dashboard/summary'),
};

export const productsApi = {
	list: async (page: number = 1, limit: number = 100, search?: string) =>
		cachedGet<PaginatedProducts>(`products:list:${page}:${limit}:${search ?? ''}`, `/products?page=${page}&limit=${limit}${search ? `&search=${search}` : ''}`),
	get: async (id: number) => cachedGet<Product>(`products:get:${id}`, `/products/${id}`),
	getBySku: async (sku: string) => cachedGet<Product>(`products:sku:${sku}`, `/products/sku/${sku}`),
	create: async (payload: Record<string, unknown>) => {
		const data = (await api.post<Product>('/products', payload)).data;
		invalidateCache('products:');
		return data;
	},
	update: async (id: number, payload: Record<string, unknown>) => {
		const data = (await api.put<Product>(`/products/${id}`, payload)).data;
		invalidateCache('products:');
		return data;
	},
	remove: async (id: number) => {
		const data = (await api.delete(`/products/${id}`)).data;
		invalidateCache('products:');
		return data;
	},
	stock: async (id: number, params?: Record<string, string | number | undefined>) =>
		cachedGet<{ productId: number; currentQuantity: number; batches: Batch[] }>(`products:stock:${id}`, `/products/${id}/stock`, { params }),
	stockHistory: async (id: number) => cachedGet<StockAdjustment[]>(`products:history:${id}`, `/products/${id}/history`),
	createBatch: async (id: number, payload: Record<string, unknown>) => {
		const data = (await api.post<Batch>(`/products/${id}/stock/batches`, payload)).data;
		invalidateCache(`products:stock:${id}`);
		return data;
	},
	adjustStock: async (id: number, payload: Record<string, unknown>) => {
		const data = (await api.post(`/products/${id}/stock/adjustments`, payload)).data;
		invalidateCache(`products:stock:${id}`);
		return data;
	},
};

export const promotionsApi = {
	list: async (active?: boolean) => {
		const params = active ? { active: true } : {};
		const response = await api.get<{ promotions: Promotion[] }>('/promotions', { params });
		return response.data.promotions;
	},
	create: async (data: CreatePromotionRequest) => {
		const response = await api.post<Promotion>('/promotions', data);
		return response.data;
	},
	update: async (id: number, data: UpdatePromotionRequest) => {
		const response = await api.put<Promotion>(`/promotions/${id}`, data);
		return response.data;
	},
	delete: async (id: number) => {
		await api.delete(`/promotions/${id}`);
	}
};

export const categoriesApi = {
	list: async () => cachedGet<Category[]>('categories:list', '/categories'),
	get: async (id: number) => cachedGet<Category>(`categories:get:${id}`, `/categories/${id}`),
	getByName: async (name: string) => cachedGet<Category>(`categories:name:${name}`, `/categories/name/${name}`),
	create: async (payload: { name: string }) => {
		const data = (await api.post<Category>('/categories', payload)).data;
		invalidateCache('categories:');
		return data;
	},
	update: async (id: number, payload: { name: string }) => {
		const data = (await api.put<Category>(`/categories/${id}`, payload)).data;
		invalidateCache('categories:');
		return data;
	},
	remove: async (id: number) => {
		const data = (await api.delete(`/categories/${id}`)).data;
		invalidateCache('categories:');
		return data;
	},
};

export const subCategoriesApi = {
	list: async (categoryId: number) => cachedGet<SubCategory[]>(`subcategories:list:${categoryId}`, `/categories/${categoryId}/sub-categories`),
	get: async (id: number) => cachedGet<SubCategory>(`subcategories:get:${id}`, `/sub-categories/${id}`),
	create: async (data: { name: string; }, categoryId: number) => {
		const result = (await api.post<SubCategory>(`/categories/${categoryId}/sub-categories`, data)).data;
		invalidateCache('subcategories:');
		return result;
	},
	update: async (id: number, data: { name: string }) => {
		const result = (await api.put<SubCategory>(`/sub-categories/${id}`, data)).data;
		invalidateCache('subcategories:');
		return result;
	},
	remove: async (id: number) => {
		const result = (await api.delete(`/sub-categories/${id}`)).data;
		invalidateCache('subcategories:');
		return result;
	},
};

export const suppliersApi = {
	list: async () => cachedGet<Supplier[]>('suppliers:list', '/suppliers'),
	get: async (id: number) => cachedGet<Supplier>(`suppliers:get:${id}`, `/suppliers/${id}`),
	getByName: async (name: string) => cachedGet<Supplier>(`suppliers:name:${name}`, `/suppliers/name/${name}`),
	create: async (payload: Record<string, unknown>) => {
		const data = (await api.post<Supplier>('/suppliers', payload)).data;
		invalidateCache('suppliers:');
		return data;
	},
	update: async (id: number, payload: Record<string, unknown>) => {
		const data = (await api.put<Supplier>(`/suppliers/${id}`, payload)).data;
		invalidateCache('suppliers:');
		return data;
	},
	remove: async (id: number) => {
		const data = (await api.delete(`/suppliers/${id}`)).data;
		invalidateCache('suppliers:');
		return data;
	},
	performance: async (id: number) => cachedGet<SupplierPerformance>(`suppliers:perf:${id}`, `/suppliers/${id}/performance`),
};

export const locationsApi = {
	list: async () => cachedGet<Location[]>('locations:list', '/locations'),
	get: async (id: number) => cachedGet<Location>(`locations:get:${id}`, `/locations/${id}`),
	create: async (payload: Record<string, unknown>) => {
		const data = (await api.post<Location>('/locations', payload)).data;
		invalidateCache('locations:');
		return data;
	},
	update: async (id: number, payload: Record<string, unknown>) => {
		const data = (await api.put<Location>(`/locations/${id}`, payload)).data;
		invalidateCache('locations:');
		return data;
	},
	remove: async (id: number) => {
		const data = (await api.delete(`/locations/${id}`)).data;
		invalidateCache('locations:');
		return data;
	},
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
	getPO: async (id: number) => (await api.get<PurchaseOrder>(`/replenishment/purchase-orders/${id}`)).data,
	listSuggestions: async (params?: Record<string, unknown>) =>
		(await api.get<ReorderSuggestion[]>('/replenishment/suggestions', { params })).data,
	generateSuggestions: async () => (await api.post('/replenishment/suggestions/generate')).data,
	createPOFromSuggestion: async (suggestionId: number) =>
		(await api.post<PurchaseOrder>(`/replenishment/suggestions/${suggestionId}/create-po`)).data,
	approvePO: async (poId: number) => (await api.post(`/replenishment/purchase-orders/${poId}/approve`)).data,
	sendPO: async (poId: number) => (await api.post(`/replenishment/purchase-orders/${poId}/send`)).data,
	updatePO: async (poId: number, payload: Record<string, unknown>) =>
		(await api.put(`/replenishment/purchase-orders/${poId}`, payload)).data,
	receivePO: async (poId: number, payload: Record<string, unknown>) =>
		(await api.post(`/replenishment/purchase-orders/${poId}/receive`, payload)).data,
	cancelPO: async (poId: number) => (await api.post(`/replenishment/purchase-orders/${poId}/cancel`)).data,
	getDashboard: async () => (await api.get('/replenishment/dashboard')).data,
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
	// New Real-Time Reports
	stockAging: async () => (await api.get('/reports/stock-aging')).data,
	deadStock: async () => (await api.get('/reports/dead-stock')).data,
	supplierPerformance: async (params?: Record<string, unknown>) =>
		(await api.get('/reports/supplier-performance', { params })).data,
	hourlyHeatmap: async (params?: Record<string, unknown>) =>
		(await api.get('/reports/heatmap', { params })).data,
	salesByEmployee: async (params?: Record<string, unknown>) =>
		(await api.get('/reports/employee-sales', { params })).data,
	categoryDrilldown: async (params?: Record<string, unknown>) =>
		(await api.get('/reports/category-drilldown', { params })).data,
	gmroi: async (params?: Record<string, unknown>) =>
		(await api.get('/reports/gmroi', { params })).data,
	voidAudit: async (params?: Record<string, unknown>) =>
		(await api.get('/reports/audit/voids', { params })).data,
	taxLiability: async (params?: Record<string, unknown>) =>
		(await api.get('/reports/tax-liability', { params })).data,
	cashReconciliation: async (params?: Record<string, unknown>) =>
		(await api.get('/reports/cash-reconciliation', { params })).data,
	customerInsights: async (params?: Record<string, unknown>) =>
		(await api.get('/reports/customer-insights', { params })).data,
	shrinkage: async (params?: Record<string, unknown>) =>
		(await api.get('/reports/shrinkage', { params })).data,
	returnsAnalysis: async (params?: Record<string, unknown>) =>
		(await api.get('/reports/returns-analysis', { params })).data,
	basketAnalysis: async (params?: Record<string, unknown>) =>
		(await api.get<PaginatedResponse<BasketAnalysisItem>>('/reports/basket-analysis', { params })).data,
	exportReport: async (params: { format: string }) =>
		(await api.post('/reports/sales-trends/export', null, { params })).data,
	getJobStatus: async (jobId: string) =>
		(await api.get(`/reports/jobs/${jobId}`)).data,
	getDownloadUrl: (jobId: string) => `${API_BASE_URL}/reports/download/${jobId}`,
};

export const alertsApi = {
	list: async (params?: Record<string, unknown>) => (await api.get<Alert[]>('/alerts', { params })).data,
	get: async (id: number) => (await api.get<Alert>(`/alerts/${id}`)).data,
	resolve: async (id: number) => (await api.patch(`/alerts/${id}/resolve`)).data,
	updateProductSettings: async (productId: number, payload: Record<string, unknown>) =>
		(await api.put(`/alerts/products/${productId}/settings`, payload)).data,
	updateUserNotifications: async (userId: number, payload: Record<string, unknown>) =>
		(await api.put(`/alerts/users/${userId}/notification-settings`, payload)).data,
	triggerCheck: async () => (await api.post('/alerts/check')).data,
	getRoleSubscriptions: async () => (await api.get('/alerts/subscriptions')).data,
	createRoleSubscription: async (data: { alertType: string; roleName: string }) =>
		(await api.post('/alerts/subscriptions', data)).data,
	deleteRoleSubscription: async (id: number) => (await api.delete(`/alerts/subscriptions/${id}`)).data,
};

export const notificationsApi = {
	list: async (userId: number, params?: Record<string, unknown>) =>
		(await api.get<Notification[]>(`/users/${userId}/notifications`, { params })).data,
	unreadCount: async (userId: number) =>
		(await api.get<{ count: number }>(`/users/${userId}/notifications/unread/count`)).data.count,
	markRead: async (userId: number, notificationId: number) =>
		(await api.patch(`/users/${userId}/notifications/${notificationId}/read`)).data,
	markAllRead: async (userId: number) =>
		(await api.patch(`/users/${userId}/notifications/read-all`)).data,
	getSubscriptions: async (userId: number) =>
		(await api.get(`/users/${userId}/notifications/subscriptions`)).data,
	updateSubscription: async (userId: number, type: string, enabled: boolean) =>
		(await api.put(`/users/${userId}/notifications/subscriptions/${type}`, { enabled })).data,
	broadcast: async (data: { title: string; message: string; type: string }) =>
		(await api.post('/notifications/broadcast', data)).data,
};

export const bulkApi = {
	downloadTemplate: async () => (await api.get('/bulk/products/template', { responseType: 'blob' })).data as Blob,
	uploadImport: async (formData: FormData) =>
		(await api.post<BulkImportJob>('/bulk/products/import', formData)).data,
	status: async (jobId: number | string) => (await api.get<BulkImportJob>(`/bulk/products/import/${jobId}/status`)).data,
	confirm: async (jobId: number | string) => (await api.post(`/bulk/products/import/${jobId}/confirm`)).data,
	exportProducts: async (params: Record<string, string | number | undefined>) =>
		(await api.get('/bulk/products/export', { params, responseType: 'blob' })).data as Blob,
	listJobs: async () => (await api.get<BulkImportJob[]>('/bulk/jobs')).data,
};

export const usersApi = {
	list: async (params?: Record<string, unknown>) => (await api.get<PaginatedUsers>('/users', { params })).data,
	get: async (id: number) => (await api.get<UserSummary>(`/users/${id}`)).data,
	update: async (id: number, payload: Record<string, unknown>) => (await api.put(`/users/${id}`, payload)).data,
	approve: async (id: number) => (await api.put(`/users/${id}/approve`)).data,
	remove: async (id: number) => (await api.delete(`/users/${id}`)).data,
};

export const crmApi = {
	listCustomers: async (params?: Record<string, unknown>) =>
		(await api.get<PaginatedUsers>('/crm/customers', { params })).data,
	createCustomer: async (payload: Record<string, unknown>) =>
		(await api.post<UserSummary>('/crm/customers', payload)).data,
	getCustomer: async (id: number | string) => (await api.get<UserSummary>(`/crm/customers/${id}`)).data,
	updateCustomer: async (id: number, payload: Record<string, unknown>) =>
		(await api.put<UserSummary>(`/crm/customers/${id}`, payload)).data,
	deleteCustomer: async (id: number) => (await api.delete(`/crm/customers/${id}`)).data,
	getLoyalty: async (id: number) => (await api.get<LoyaltyAccount>(`/crm/loyalty/${id}`)).data,
	addPoints: async (id: number, points: number) =>
		(await api.post<LoyaltyAccount>(`/crm/loyalty/${id}/points`, { points })).data,
	redeemPoints: async (userId: number, points: number) =>
		(await api.post<LoyaltyAccount>(`crm/loyalty/${userId}/redeem`, { points })).data,
	getChurnRisk: async (userId: number) => (await api.get(`crm/customers/${userId}/churn-risk`)).data,
};

export const purchaseOrdersApi = {
	list: async (params?: Record<string, unknown>) =>
		(await api.get<PaginatedResponse<PurchaseOrder>>('/replenishment/purchase-orders', { params })).data,
	get: async (id: number) => (await api.get<PurchaseOrder>(`/replenishment/purchase-orders/${id}`)).data,
	create: async (data: Record<string, unknown>) =>
		(await api.post<PurchaseOrder>('/replenishment/purchase-orders', data)).data,
	approve: async (id: number) => (await api.post(`/replenishment/purchase-orders/${id}/approve`)).data,
	send: async (id: number) => (await api.post(`/replenishment/purchase-orders/${id}/send`)).data,
	receive: async (id: number, data: Record<string, unknown>) =>
		(await api.post(`/replenishment/purchase-orders/${id}/receive`, data)).data,
	cancel: async (id: number) => (await api.post(`/replenishment/purchase-orders/${id}/cancel`)).data,
};

export const timeTrackingApi = {
	clockIn: async (userId: number, notes?: string) =>
		(await api.post<TimeClock>(`/time-tracking/clock-in/${userId}`, { notes })).data,
	clockOut: async (userId: number, notes?: string) =>
		(await api.post<TimeClock>(`/time-tracking/clock-out/${userId}`, { notes })).data,
	startBreak: async (userId: number) => (await api.post<TimeClock>(`/time-tracking/break-start/${userId}`)).data,
	endBreak: async (userId: number) => (await api.post<TimeClock>(`/time-tracking/break-end/${userId}`)).data,
	getLastEntry: async (userId: number) => (await api.get<TimeClock>(`/time-tracking/last-entry/${userId}`)).data,
	getHistory: async (userId: number) =>
		(await api.get<{ history: TimeClock[] }>(`/time-tracking/history/${userId}`)).data.history,
	getTeamStatus: async () => (await api.get<TeamStatus[]>('/time-tracking/team-status')).data,
	getRecentActivities: async () => (await api.get<TimeClock[]>('/time-tracking/activities')).data,
	getWeeklySummary: async (userId: number) =>
		(await api.get<WeeklySummary>(`/time-tracking/weekly-summary/${userId}`)).data,
	getTeamOverview: async () => (await api.get<TeamOverview>('/time-tracking/team-overview')).data
};

export const paymentApi = {
	create: async (data: PaymentRequest) => (await api.post<PaymentResponse>('/payment/create', data)).data,
};

export const purchaseReturnsApi = {
	list: async (params?: Record<string, unknown>) =>
		(await api.get<PaginatedResponse<PurchaseReturn>>('/replenishment/returns', { params })).data,
	create: async (data: Record<string, unknown>) =>
		(await api.post<PurchaseReturn>('/replenishment/returns', data)).data,
	approve: async (id: number) => (await api.post(`/replenishment/returns/${id}/approve`)).data,
	complete: async (id: number) => (await api.post(`/replenishment/returns/${id}/complete`)).data,
	cancel: async (id: number) => (await api.post(`/replenishment/returns/${id}/cancel`)).data,
};
