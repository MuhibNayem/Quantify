import api from '$lib/api';

export const settingsApi = {
    getSettings: async () => {
        const response = await api.get('/settings');
        return response.data;
    },

    updateSetting: async (key: string, value: string) => {
        const response = await api.put('/settings', { key, value });
        return response.data;
    },

    getAllPermissions: async () => {
        const response = await api.get('/permissions');
        return response.data;
    },

    getRolePermissions: async (role: string) => {
        const response = await api.get(`/permissions/${role}`);
        return response.data;
    },

    updateRolePermissions: async (role: string, permissionIds: number[]) => {
        const response = await api.put(`/permissions/${role}`, { permissionIds });
        return response.data;
    }
};
