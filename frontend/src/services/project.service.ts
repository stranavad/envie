import { api } from './api';

export interface Project {
    id: string;
    name: string;
    organizationId: string;
    organizationName: string;
    teamId: string;
    teamName: string;
    encryptedProjectKey: string;
    encryptedTeamKey?: string;
    keyVersion: number;
    configChecksum?: string;
    createdAt: string;
    updatedAt: string;
}

export interface ProjectDetail {
    id: string;
    name: string;
    organizationId: string;
    organizationName: string;
    createdAt: string;
    updatedAt: string;
    encryptedProjectKey: string;
    encryptedTeamKey?: string;
    teamId: string;
    teamName: string;
    teamRole?: string;
    orgRole?: string;
    canEdit: boolean;
    canDelete: boolean;
    keyVersion: number;
    configChecksum?: string;
}

export interface CreateProjectRequest {
    name: string;
    organizationId: string;
    teamId: string;
    encryptedKey: string;
}

export interface CreateProjectResponse {
    id: string;
    name: string;
    organizationId: string;
}

export interface TeamUser {
    id: string;
    name: string;
    email: string;
    avatarUrl: string;
    role: string;
}

export interface TeamWithUsers {
    id: string;
    name: string;
    memberCount: number;
    projectCount: number;
    users: TeamUser[];
}

export interface OrgUser {
    id: string;
    name: string;
    email: string;
    avatarUrl: string;
    role: string;
}

export interface ProjectAccessData {
    teams: TeamWithUsers[];
    organizationAdmins: OrgUser[];
    availableTeams: TeamWithUsers[];
}

export interface AddTeamToProjectRequest {
    teamId: string;
    encryptedProjectKey: string;
}

export interface ProjectFile {
    id: string;
    name: string;
    sizeBytes: number;
    mimeType: string;
    encryptedFek: string;
    checksum: string;
    uploadedBy: {
        id: string;
        name: string;
        email: string;
    };
    createdAt: string;
}

export interface UploadFileRequest {
    file: File;
    name: string;
    encryptedFek: string;
    checksum: string;
    mimeType: string;
    originalSize: number;
}

export interface DownloadFileResponse {
    data: string; // base64 encoded
    encryptedFek: string;
    checksum: string;
    name: string;
    mimeType: string;
}

export interface FileFEK {
    id: string;
    encryptedFek: string;
}

export class ProjectService {
    static async getProjects(): Promise<Project[]> {
        return api.get<Project[]>('/projects');
    }

    static async getProject(id: string): Promise<ProjectDetail> {
        return api.get<ProjectDetail>(`/projects/${id}`);
    }

    static async createProject(request: CreateProjectRequest): Promise<CreateProjectResponse> {
        return api.post<CreateProjectResponse>('/projects', request);
    }

    static async updateProject(id: string, name: string): Promise<void> {
        await api.put(`/projects/${id}`, { name });
    }

    static async deleteProject(id: string): Promise<void> {
        await api.delete(`/projects/${id}`);
    }

    static async getConfig(projectId: string): Promise<ConfigItem[]> {
        return api.get<ConfigItem[]>(`/projects/${projectId}/config`);
    }

    static async syncConfig(projectId: string, items: ConfigItem[]): Promise<void> {
        await api.put(`/projects/${projectId}/config`, { items });
    }

    static async getProjectTeams(projectId: string): Promise<ProjectAccessData> {
        return api.get<ProjectAccessData>(`/projects/${projectId}/teams`);
    }

    static async addTeamToProject(projectId: string, request: AddTeamToProjectRequest): Promise<void> {
        await api.post(`/projects/${projectId}/teams`, request);
    }

    static async getFiles(projectId: string): Promise<ProjectFile[]> {
        return api.get<ProjectFile[]>(`/projects/${projectId}/files`);
    }

    static async uploadFile(projectId: string, request: UploadFileRequest): Promise<{ id: string; name: string; sizeBytes: number }> {
        const formData = new FormData();
        formData.append('file', request.file);
        formData.append('name', request.name);
        formData.append('encryptedFek', request.encryptedFek);
        formData.append('checksum', request.checksum);
        formData.append('mimeType', request.mimeType);
        formData.append('originalSize', request.originalSize.toString());

        return api.postFormData<{ id: string; name: string; sizeBytes: number }>(`/projects/${projectId}/files`, formData);
    }

    static async downloadFile(projectId: string, fileId: string): Promise<DownloadFileResponse> {
        return api.get<DownloadFileResponse>(`/projects/${projectId}/files/${fileId}`);
    }

    static async deleteFile(projectId: string, fileId: string): Promise<void> {
        await api.delete(`/projects/${projectId}/files/${fileId}`);
    }

    static async getFilesForRotation(projectId: string): Promise<FileFEK[]> {
        return api.get<FileFEK[]>(`/projects/${projectId}/files-feks`);
    }
}

export interface ConfigItem {
    id: string;
    projectId: string;
    name: string;
    value: string;
    sensitive: boolean;
    position: number;
    category?: string;
    createdAt?: string;
    updatedAt?: string;

    secretManagerName?: string;
    secretManagerConfigId?: string;
    secretManagerLastSyncAt?: string;
    secretManagerVersion?: string | null;

    creator?: { id: string; name: string; email: string };
    updater?: { id: string; name: string; email: string };
}
