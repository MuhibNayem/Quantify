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
	Promotion,
	ReorderSuggestion,
	StockAdjustment,
	Supplier,
	SupplierPerformance,
	UserSummary,
	PaginatedUsers,
	PaginatedProducts,
	Notification,
	DashboardSummary,
	LoyaltyAccount,
	TimeClock,
} from '$lib/types';

export const dashboardApi = {
	getSummary: async () => (await api.get<DashboardSummary>('/dashboard/summary')).data,
};

export const productsApi = {
	list: async (page: number = 1, limit: number = 100, search?: string) => (await api.get<PaginatedProducts>(`/products?page=${page}&limit=${limit}${search ? `&search=${search}` : ''}`)).data,
	get: async (id: number) => (await api.get<Product>(`/products/${id}`)).data,
	getBySku: async (sku: string) => (await api.get<Product>(`/products/sku/${sku}`)).data,
	create: async (payload: Record<string, unknown>) => (await api.post<Product>('/products', payload)).data,
	update: async (id: number, payload: Record<string, unknown>) => (await api.put<Product>(`/products/${id}`, payload)).data,
	remove: async (id: number) => (await api.delete(`/products/${id}`)).data,
	stock: async (id: number, params?: Record<string, string | number | undefined>) =>
		(await api.get(`/products/${id}/stock`, { params })).data as { productId: number; currentQuantity: number; batches: Batch[] },
	stockHistory: async (id: number) => (await api.get<StockAdjustment[]>(`/products/${id}/history`)).data,
	createBatch: async (id: number, payload: Record<string, unknown>) => (await api.post<Batch>(`/products/${id}/stock/batches`, payload)).data,
	adjustStock: async (id: number, payload: Record<string, unknown>) =>
		(await api.post(`/products/${id}/stock/adjustments`, payload)).data,
};

export const promotionsApi = {
	list: async (active?: boolean) => {
		const params = active ? { active: true } : {};
		const response = await api.get<{ promotions: Promotion[] }>('/promotions', { params });
		return response.data.promotions;
	},
	create: async (data: any) => {
		const response = await api.post<Promotion>('/promotions', data);
		return response.data;
	},
	update: async (id: number, data: any) => {
		const response = await api.put<Promotion>(`/promotions/${id}`, data);
		return response.data;
	},
	delete: async (id: number) => {
		await api.delete(`/promotions/${id}`);
	}
};

export const categoriesApi = {
	list: async () => (await api.get<Category[]>('/categories')).data,
	get: async (id: number) => (await api.get<Category>(`/categories/${id}`)).data,
	getByName: async (name: string) => (await api.get<Category>(`/categories/name/${name}`)).data,
	create: async (payload: { name: string }) => (await api.post<Category>('/categories', payload)).data,
	update: async (id: number, payload: { name: string }) => (await api.put<Category>(`/categories/${id}`, payload)).data,
	remove: async (id: number) => (await api.delete(`/categories/${id}`)).data,
};

export const subCategoriesApi = {
	list: async (categoryId: number) => (await api.get<SubCategory[]>(`/categories/${categoryId}/sub-categories`)).data,
	get: async (id: number) => (await api.get<SubCategory>(`/sub-categories/${id}`)).data,
	create: async (data: { name: string; }, categoryId: number) => (await api.post<SubCategory>(`/categories/${categoryId}/sub-categories`, data)).data,
	update: async (id: number, data: { name: string }) => (await api.put<SubCategory>(`/sub-categories/${id}`, data)).data,
	remove: async (id: number) => (await api.delete(`/sub-categories/${id}`)).data,
};

export const suppliersApi = {
	list: async () => (await api.get<Supplier[]>('/suppliers')).data,
	get: async (id: number) => (await api.get<Supplier>(`/suppliers/${id}`)).data,
	getByName: async (name: string) => (await api.get<Supplier>(`/suppliers/name/${name}`)).data,
	create: async (payload: Record<string, unknown>) => (await api.post<Supplier>('/suppliers', payload)).data,
	update: async (id: number, payload: Record<string, unknown>) => (await api.put<Supplier>(`/suppliers/${id}`, payload)).data,
	remove: async (id: number) => (await api.delete(`/suppliers/${id}`)).data,
	performance: async (id: number) => (await api.get<SupplierPerformance>(`/suppliers/${id}/performance`)).data,
};

export const locationsApi = {
	list: async () => (await api.get<Location[]>('/locations')).data,
	get: async (id: number) => (await api.get<Location>(`/locations/${id}`)).data,
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
		(await api.get('/reports/basket-analysis', { params })).data,
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
	getTeamStatus: async () => (await api.get<any[]>('/time-tracking/team-status')).data,
	getRecentActivities: async () => (await api.get<TimeClock[]>('/time-tracking/activities')).data,
	getWeeklySummary: async (userId: number) =>
		(await api.get<any>(`/time-tracking/weekly-summary/${userId}`)).data,
	getTeamOverview: async () => (await api.get<any>('/time-tracking/team-overview')).data
};
