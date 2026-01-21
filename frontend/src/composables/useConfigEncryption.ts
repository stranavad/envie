import { EncryptionService } from '@/services/encryption.service';
import { ProjectService, type ConfigItem } from '@/services/project.service';

/**
 * Composable for encrypting and decrypting config items.
 * Centralizes the common patterns used across the app for handling
 * project configuration encryption/decryption.
 */
export function useConfigEncryption() {
    /**
     * Decrypt a single config item value
     */
    async function decryptConfigValue(
        projectKey: string,
        encryptedValue: string
    ): Promise<string> {
        return await EncryptionService.decryptValue(projectKey, encryptedValue);
    }

    /**
     * Encrypt a single config item value
     */
    async function encryptConfigValue(
        projectKey: string,
        plainValue: string
    ): Promise<string> {
        return await EncryptionService.encryptValue(projectKey, plainValue);
    }

    /**
     * Decrypt all config items in place.
     * Returns a new array with decrypted values.
     * Items that fail to decrypt will have value set to '[DECRYPTION FAILED]'
     */
    async function decryptConfigItems(
        projectKey: string,
        items: ConfigItem[]
    ): Promise<ConfigItem[]> {
        const decryptedItems: ConfigItem[] = [];

        for (const item of items) {
            try {
                const decryptedValue = await EncryptionService.decryptValue(projectKey, item.value);
                decryptedItems.push({ ...item, value: decryptedValue });
            } catch (e) {
                console.error(`Failed to decrypt item ${item.name}`, e);
                decryptedItems.push({ ...item, value: '[DECRYPTION FAILED]' });
            }
        }

        return decryptedItems;
    }

    /**
     * Encrypt all config items.
     * Returns a new array with encrypted values.
     */
    async function encryptConfigItems(
        projectKey: string,
        items: ConfigItem[]
    ): Promise<ConfigItem[]> {
        return await Promise.all(
            items.map(async (item) => ({
                ...item,
                value: await EncryptionService.encryptValue(projectKey, item.value),
            }))
        );
    }

    /**
     * Fetch config items from backend and decrypt them.
     * Convenience method that combines fetch + decrypt.
     */
    async function fetchAndDecryptConfig(
        projectId: string,
        projectKey: string
    ): Promise<ConfigItem[]> {
        const configs = await ProjectService.getConfig(projectId);
        const decrypted = await decryptConfigItems(projectKey, configs);
        return decrypted.sort((a, b) => a.position - b.position);
    }

    /**
     * Encrypt config items and sync to backend.
     * Convenience method that combines encrypt + save.
     */
    async function encryptAndSyncConfig(
        projectId: string,
        projectKey: string,
        items: ConfigItem[]
    ): Promise<void> {
        const encrypted = await encryptConfigItems(projectKey, items);
        await ProjectService.syncConfig(projectId, encrypted);
    }

    /**
     * Decrypt a secret manager config's encryption key.
     * The encryptedKey field contains JSON config encrypted with the project key.
     */
    async function decryptSecretManagerConfig(
        projectKey: string,
        encryptedConfig: string
    ): Promise<string> {
        return await EncryptionService.decryptValue(projectKey, encryptedConfig);
    }

    return {
        decryptConfigValue,
        encryptConfigValue,
        decryptConfigItems,
        encryptConfigItems,
        fetchAndDecryptConfig,
        encryptAndSyncConfig,
        decryptSecretManagerConfig,
    };
}
