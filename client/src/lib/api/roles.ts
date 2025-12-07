import api from '$lib/api';

export interface Permission {
    ID: number;
    Name: string;
    Description: string;
    Group: string;
}

export interface Role {
    ID: number;
    CreatedAt: string;
    UpdatedAt: string;
    DeletedAt: string | null;
    Name: string;
    Description: string;
    IsSystem: boolean;
    Permissions: Permission[];
}

export const rolesApi = {
    list: async () => {
        const response = await api.get('/roles');
        return response.data as Role[];
    },

    create: async (data: { name: string; description: string }) => {
        const response = await api.post('/roles', data);
        return response.data as Role;
    },

    update: async (id: number, data: { name: string; description: string }) => {
        const response = await api.put(`/roles/${id}`, data);
        return response.data as Role;
    },

    delete: async (id: number) => {
        await api.delete(`/roles/${id}`);
    },

    listPermissions: async () => {
        const response = await api.get('/permissions');
        return response.data as Permission[];
    },

    updatePermissions: async (roleId: number, permissionIds: number[]) => {
        await api.put(
            `/roles/${roleId}/permissions`,
            { permission_ids: permissionIds }
        );
    }
};
