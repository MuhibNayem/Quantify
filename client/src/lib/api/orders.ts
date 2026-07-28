import api from '$lib/api';
import type { PaginatedResponse } from '$lib/types';

export const ordersApi = {
    listOrders: async (params?: Record<string, unknown>) => {
        const response = await api.get<PaginatedResponse<any>>('/sales/orders', { params });
        return response.data;
    },

    getOrder: async (orderNumber: string) => {
        const response = await api.get(`/sales/orders/${orderNumber}`);
        return response.data;
    }
};
