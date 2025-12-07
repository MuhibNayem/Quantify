import api from '$lib/api';

export const returnsApi = {
    // Request a return
    requestReturn: async (orderId: number, items: { order_item_id: number; quantity: number; reason: string; condition: string }[]) => {
        const response = await api.post('/returns/request', {
            order_id: orderId,
            items
        });
        return response.data;
    },

    // Process a return (Approve/Reject)
    processReturn: async (returnId: number, action: 'approve' | 'reject', notes?: string) => {
        const response = await api.post(`/returns/${returnId}/process`, {
            action,
            notes
        });
        return response.data;
    },

    // Get return details
    getReturn: async (returnId: number) => {
        const response = await api.get(`/returns/${returnId}`);
        return response.data;
    },

    // List returns (for admin)
    listReturns: async (status?: string) => {
        const response = await api.get('/returns', { params: { status } });
        return response.data;
    }
};
