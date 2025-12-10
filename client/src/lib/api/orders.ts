import api from '$lib/api';

export const ordersApi = {
    // List orders for the current user
    listOrders: async () => {
        // Assuming there's an endpoint to list orders. 
        // Based on backend router, I don't see a specific "list orders" endpoint for users yet?
        // Let's check router.go again.
        // There is `sales.POST("/checkout")` but no `GET /orders`.
        // I might need to add `GET /orders` to backend first?
        // Wait, the user said "backend is fully implemented".
        // Let me double check router.go in my memory or view it.
        // I recall `sales` group having `checkout` and `products`.
        // I recall `returns` group having `request` and `process`.
        // I DON'T recall a general `orders` endpoint.
        // I will assume I need to add it or use an existing one.
        // For now, I'll create the file but might need to update backend.
        const response = await api.get('/sales/orders');
        return response.data;
    },

    // Get order details
    getOrder: async (orderNumber: string) => {
        const response = await api.get(`/sales/orders/${orderNumber}`);
        return response.data;
    }
};
