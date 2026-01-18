import { api } from './api';

export interface SecretManagerConfig {
    id: string;
    projectId: string;
    name: string;
    encryptedKey: string;
    createdAt: string;
    updatedAt: string;
    createdBy?: { id: string; name: string; email: string };
    updatedBy?: { id: string; name: string; email: string };
}

export class SecretManagerConfigService {
    static async getConfigs(projectId: string): Promise<SecretManagerConfig[]> {
        return api.get<SecretManagerConfig[]>(`/projects/${projectId}/secret-managers`);
    }

    static async createConfig(projectId: string, name: string, encryptedKey: string): Promise<SecretManagerConfig> {
        return api.post<SecretManagerConfig>(`/projects/${projectId}/secret-managers`, { name, encryptedKey });
    }

    static async updateConfig(projectId: string, configId: string, name: string, encryptedKey?: string): Promise<SecretManagerConfig> {
        return api.put<SecretManagerConfig>(`/projects/${projectId}/secret-managers/${configId}`, { name, encryptedKey });
    }

    static async deleteConfig(projectId: string, configId: string): Promise<void> {
        await api.delete(`/projects/${projectId}/secret-managers/${configId}`);
    }
}
