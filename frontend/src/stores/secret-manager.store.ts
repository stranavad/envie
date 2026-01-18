import { defineStore } from 'pinia';
import { ref } from 'vue';
import { SecretManagerService, type ServiceAccountKey } from '@/services/secret-manager.service';

interface CachedToken {
    token: string;
    expiresAt: number; // timestamp in ms
}

export const useSecretManagerStore = defineStore('secret-manager', () => {
    // State
    const accessTokens = ref<Map<string, CachedToken>>(new Map()); // Key: providerId (configId)
    const secretsCache = ref<Map<string, string[]>>(new Map()); // Key: providerId

    // Actions
    async function getAccessToken(providerId: string, serviceAccountJson: string): Promise<string> {
        const now = Date.now();
        const cached = accessTokens.value.get(providerId);

        if (cached && cached.expiresAt > now) {
            return cached.token;
        }

        const serviceAccount: ServiceAccountKey = JSON.parse(serviceAccountJson);
        const token = await SecretManagerService.getAccessToken(serviceAccount);

        // Google tokens usually last 1 hour (3600 seconds). We'll cache for 55 minutes to be safe.
        accessTokens.value.set(providerId, {
            token,
            expiresAt: now + (55 * 60 * 1000)
        });

        return token;
    }

    async function listSecrets(providerId: string, serviceAccountJson: string, forceRefresh = false): Promise<string[]> {
        if (!forceRefresh && secretsCache.value.has(providerId)) {
            return secretsCache.value.get(providerId)!;
        }

        const token = await getAccessToken(providerId, serviceAccountJson);
        const serviceAccount: ServiceAccountKey = JSON.parse(serviceAccountJson);
        const secrets = await SecretManagerService.listSecretsWithToken(token, serviceAccount.project_id);

        secretsCache.value.set(providerId, secrets);
        return secrets;
    }

    async function getSecretValue(providerId: string, serviceAccountJson: string, secretName: string, version: string = 'latest') {
        const token = await getAccessToken(providerId, serviceAccountJson);
        const serviceAccount: ServiceAccountKey = JSON.parse(serviceAccountJson);
        return SecretManagerService.getSecretValueWithToken(token, serviceAccount.project_id, secretName, version);
    }

    async function testConnection(providerId: string, serviceAccountJson: string): Promise<boolean> {
        const token = await getAccessToken(providerId, serviceAccountJson);
        const serviceAccount: ServiceAccountKey = JSON.parse(serviceAccountJson);
        return SecretManagerService.testConnectionWithToken(token, serviceAccount.project_id);
    }

    // Clear cache if needed (e.g. on config update)
    function clearCache(providerId?: string) {
        if (providerId) {
            accessTokens.value.delete(providerId);
            secretsCache.value.delete(providerId);
        } else {
            accessTokens.value.clear();
            secretsCache.value.clear();
        }
    }

    return {
        getAccessToken,
        listSecrets,
        getSecretValue,
        testConnection,
        clearCache
    };
});
