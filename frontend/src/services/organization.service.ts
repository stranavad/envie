import { api } from './api';

export interface Organization {
    id: string;
    name: string;
    createdAt: string;
    updatedAt: string;
}

export interface OrganizationListItem extends Organization {
    role: string;
    projectCount: number;
    memberCount: number;
}

export interface OrganizationDetail extends Organization {
    role: string;
    encryptedOrganizationKey?: string;
}

export interface OrganizationUser {
    id: string;
    name: string;
    email: string;
    avatarUrl: string;
    publicKey?: string;
    role: string;
}

export interface CreateOrganizationRequest {
    name: string;
    encryptedOrganizationKey: string;
    generalTeamEncryptedKey: string;
    generalTeamUserEncryptedKey: string;
}

export interface UpdateOrganizationRequest {
    name: string;
}

export class OrganizationService {
    static async getOrganizations(): Promise<OrganizationListItem[]> {
        return api.get<OrganizationListItem[]>('/organizations');
    }

    static async getOrganization(id: string): Promise<OrganizationDetail> {
        const data = await api.get<{
            organization: Organization;
            role: string;
            encryptedOrganizationKey?: string;
        }>(`/organizations/${id}`);

        return {
            ...data.organization,
            role: data.role,
            encryptedOrganizationKey: data.encryptedOrganizationKey,
        };
    }

    static async createOrganization(request: CreateOrganizationRequest): Promise<Organization> {
        return api.post<Organization>('/organizations', request);
    }

    static async updateOrganization(id: string, request: UpdateOrganizationRequest): Promise<void> {
        await api.put(`/organizations/${id}`, request);
    }

    static async getOrganizationUsers(organizationId: string): Promise<OrganizationUser[]> {
        return api.get<OrganizationUser[]>(`/organizations/${organizationId}/users`);
    }
}
