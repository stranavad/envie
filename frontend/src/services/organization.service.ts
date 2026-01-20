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

export interface AddOrganizationMemberRequest {
    userId: string;
    role: string;
    encryptedOrganizationKey?: string;
}

export interface SearchUserResult {
    id: string;
    name: string;
    email: string;
    avatarUrl: string;
    publicKey?: string;
}

export interface UpdateOrganizationMemberRequest {
    role: string;
    encryptedOrganizationKey?: string;
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

    static async addOrganizationMember(orgId: string, request: AddOrganizationMemberRequest): Promise<void> {
        await api.post(`/organizations/${orgId}/members`, request);
    }

    static async searchUserByEmail(email: string): Promise<SearchUserResult> {
        const searchParams = new URLSearchParams()
        searchParams.set('email', email)
        return api.get<SearchUserResult>(`/users/search?${searchParams.toString()}`);
    }

    static async updateOrganizationMember(orgId: string, userId: string, request: UpdateOrganizationMemberRequest): Promise<void> {
        await api.put(`/organizations/${orgId}/members/${userId}`, request);
    }

    static async removeOrganizationMember(orgId: string, userId: string): Promise<void> {
        await api.delete(`/organizations/${orgId}/members/${userId}`);
    }
}
