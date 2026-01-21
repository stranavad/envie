import { ref } from 'vue';
import { SecretManagerConfigService, type SecretManagerConfig } from '@/services/secret-manager-config.service';
import { useSecretManagerStore } from '@/stores/secret-manager.store';
import { useConfigEncryption } from './useConfigEncryption';

export interface SecretOption {
    name: string;
    providerId: string;
    providerName: string;
}

export interface SyncedSecretResult {
    value: string;
    version?: string;
    lastSyncAt: string;
}

/**
 * Composable for secret manager operations.
 * Centralizes the logic for loading secret manager configs,
 * listing secrets, and syncing secret values.
 */
export function useSecretManager(projectId: string, projectKey: string) {
    const secretManagerStore = useSecretManagerStore();
    const { decryptSecretManagerConfig } = useConfigEncryption();

    const isLoading = ref(false);
    const configs = ref<SecretManagerConfig[]>([]);
    const availableSecrets = ref<SecretOption[]>([]);
    const error = ref('');

    /**
     * Load all secret manager configs for the project
     */
    async function loadConfigs(): Promise<SecretManagerConfig[]> {
        try {
            configs.value = await SecretManagerConfigService.getConfigs(projectId);
            return configs.value;
        } catch (e: any) {
            console.error('Failed to load secret manager configs', e);
            error.value = 'Failed to load secret manager configs: ' + e.message;
            return [];
        }
    }

    /**
     * Load available secrets from all configured providers.
     * This decrypts each config's encryption key and fetches secret names.
     */
    async function loadAvailableSecrets(): Promise<SecretOption[]> {
        isLoading.value = true;
        error.value = '';

        try {
            if (configs.value.length === 0) {
                await loadConfigs();
            }

            const secrets: SecretOption[] = [];

            for (const config of configs.value) {
                try {
                    // Decrypt the config's encryption key
                    const decryptedJson = await decryptSecretManagerConfig(projectKey, config.encryptedKey);

                    // Fetch secrets from the provider
                    const list = await secretManagerStore.listSecrets(config.id, decryptedJson);
                    list.forEach(name => {
                        secrets.push({
                            name,
                            providerId: config.id,
                            providerName: config.name
                        });
                    });
                } catch (e) {
                    console.error(`Failed to load secrets for provider ${config.name}`, e);
                }
            }

            availableSecrets.value = secrets;
            return secrets;
        } finally {
            isLoading.value = false;
        }
    }

    /**
     * Get the value of a specific secret from a provider.
     * Returns the value, version, and sync timestamp.
     */
    async function syncSecret(
        configId: string,
        secretName: string
    ): Promise<SyncedSecretResult> {
        isLoading.value = true;
        error.value = '';

        try {
            const config = configs.value.find(c => c.id === configId);
            if (!config) {
                // Try loading configs if not available
                await loadConfigs();
                const refreshedConfig = configs.value.find(c => c.id === configId);
                if (!refreshedConfig) {
                    throw new Error('Secret manager config not found');
                }
            }

            const targetConfig = configs.value.find(c => c.id === configId)!;

            // Decrypt the config's encryption key
            const decryptedJson = await decryptSecretManagerConfig(projectKey, targetConfig.encryptedKey);

            // Fetch the secret value
            const result = await secretManagerStore.getSecretValue(configId, decryptedJson, secretName);

            return {
                value: result.value,
                version: result.version,
                lastSyncAt: new Date().toISOString()
            };
        } catch (e: any) {
            error.value = 'Failed to sync secret: ' + e.message;
            throw e;
        } finally {
            isLoading.value = false;
        }
    }

    /**
     * Find the provider ID for a secret name (auto-select if secret exists)
     */
    function findProviderForSecret(secretName: string): string | undefined {
        const match = availableSecrets.value.find(s => s.name === secretName);
        return match?.providerId;
    }

    return {
        isLoading,
        configs,
        availableSecrets,
        error,
        loadConfigs,
        loadAvailableSecrets,
        syncSecret,
        findProviderForSecret,
    };
}
