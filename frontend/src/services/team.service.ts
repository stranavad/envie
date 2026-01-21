import { api } from './api';

export interface Team {
    id: string;
    name: string;
    organizationId: string;
    encryptedKey: string;
    createdAt: string;
    updatedAt: string;
}

export interface TeamListItem extends Team {
    memberCount: number;
    projectCount: number;
    users: TeamUser[];
    userEncryptedKey: string;
}

export interface TeamUser {
    id: string;
    name: string;
    email: string;
    avatarUrl: string;
}

export interface TeamMember {
    userId: string;
    name: string;
    email: string;
    avatarUrl: string;
    role: string;
    joinedAt: string;
}

export interface TeamWithKey {
    teamId: string;
    teamName: string;
    organizationId: string;
    encryptedTeamKey: string;
    encryptedKey: string;
}

export interface CreateTeamRequest {
    name: string;
    organizationId: string;
    encryptedKey: string;
    userEncryptedKey: string;
}

export interface AddTeamMemberRequest {
    userId: string;
    encryptedTeamKey: string;
    role?: string;
}

export interface UpdateTeamMemberRequest {
    role: string;
}

export interface UpdateMyTeamKeyRequest {
    encryptedTeamKey: string;
}

export class TeamService {
    static async getTeams(organizationId: string): Promise<TeamListItem[]> {
        return api.get<TeamListItem[]>(`/teams?organizationId=${organizationId}`);
    }

    static async createTeam(request: CreateTeamRequest): Promise<Team> {
        return api.post<Team>('/teams', request);
    }

    static async getTeamMembers(teamId: string): Promise<TeamMember[]> {
        return api.get<TeamMember[]>(`/teams/${teamId}/members`);
    }

    static async addTeamMember(teamId: string, request: AddTeamMemberRequest): Promise<void> {
        await api.post(`/teams/${teamId}/members`, request);
    }

    static async updateTeamMember(teamId: string, userId: string, request: UpdateTeamMemberRequest): Promise<void> {
        await api.put(`/teams/${teamId}/members/${userId}`, request);
    }

    static async removeTeamMember(teamId: string, userId: string): Promise<void> {
        await api.delete(`/teams/${teamId}/members/${userId}`);
    }

    static async getMyTeams(): Promise<TeamWithKey[]> {
        return api.get<TeamWithKey[]>('/teams/my');
    }

    static async updateMyTeamKey(teamId: string, request: UpdateMyTeamKeyRequest): Promise<void> {
        await api.put(`/teams/${teamId}/my-key`, request);
    }
}
