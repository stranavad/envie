import { ref } from 'vue';
import { FileMappingService } from '@/services/file-mapping.service';
import { ProjectService, type ConfigItem } from '@/services/project.service';
import { useConfigEncryption } from './useConfigEncryption';

/**
 * Composable for syncing between local .env files and remote config.
 * Centralizes Pull and Push operations used across multiple components.
 */
export function useFileSync() {
    const { fetchAndDecryptConfig, encryptConfigItems } = useConfigEncryption();

    const isPulling = ref(false);
    const isPushing = ref(false);
    const syncError = ref('');

    /**
     * Pull remote config to local file.
     * Fetches config, decrypts it, writes to local file, and updates sync state.
     */
    async function pullToLocal(
        projectId: string,
        projectKey: string,
        configChecksum: string
    ): Promise<void> {
        const mapping = await FileMappingService.getMapping(projectId);
        if (!mapping) {
            throw new Error('No file mapping found');
        }

        const decryptedItems = await fetchAndDecryptConfig(projectId, projectKey);

        const localChecksum = await FileMappingService.writeToLocalFile(
            mapping.filePath,
            decryptedItems
        );

        await FileMappingService.updateSyncState(
            projectId,
            localChecksum,
            configChecksum
        );
    }

    /**
     * Pull with state management and error handling.
     * Returns true on success, false on failure.
     */
    async function pullWithState(
        projectId: string,
        projectKey: string,
        configChecksum: string
    ): Promise<boolean> {
        isPulling.value = true;
        syncError.value = '';

        try {
            await pullToLocal(projectId, projectKey, configChecksum);
            return true;
        } catch (e: any) {
            console.error('Pull failed', e);
            syncError.value = 'Pull failed: ' + (e.message || e.toString());
            return false;
        } finally {
            isPulling.value = false;
        }
    }

    /**
     * Push local changes to remote.
     * Merges local items into remote, encrypts, saves, and updates sync state.
     * Returns the updated project on success.
     */
    async function pushToRemote(
        projectId: string,
        projectKey: string,
        localItems: { name: string; value: string }[],
        remoteItems: ConfigItem[]
    ): Promise<{ updatedProject: any; localChecksum: string }> {
        // Merge local items into remote: update existing by name, add new ones
        const mergedItems = [...remoteItems];

        for (const localItem of localItems) {
            const existingIndex = mergedItems.findIndex(r => r.name === localItem.name);
            if (existingIndex !== -1) {
                mergedItems[existingIndex] = {
                    ...mergedItems[existingIndex],
                    value: localItem.value,
                };
            } else {
                mergedItems.push({
                    id: crypto.randomUUID(),
                    projectId,
                    name: localItem.name,
                    value: localItem.value,
                    sensitive: true,
                    position: mergedItems.length,
                });
            }
        }

        // Encrypt and save
        const itemsToSave = await encryptConfigItems(projectKey, mergedItems);
        await ProjectService.syncConfig(projectId, itemsToSave);

        // Reload project to get new checksum
        const updatedProject = await ProjectService.getProject(projectId);

        // Compute local checksum
        const localChecksum = await FileMappingService.computeChecksum(
            localItems.map((item, index) => ({
                id: '',
                projectId,
                name: item.name,
                value: item.value,
                sensitive: false,
                position: index,
            }))
        );

        // Update sync state
        await FileMappingService.updateSyncState(
            projectId,
            localChecksum,
            updatedProject.configChecksum || ''
        );

        return { updatedProject, localChecksum };
    }

    /**
     * Push with state management and error handling.
     * Returns the updated project on success, null on failure.
     */
    async function pushWithState(
        projectId: string,
        projectKey: string,
        localItems: { name: string; value: string }[],
        remoteItems: ConfigItem[]
    ): Promise<{ updatedProject: any; localChecksum: string } | null> {
        isPushing.value = true;
        syncError.value = '';

        try {
            return await pushToRemote(projectId, projectKey, localItems, remoteItems);
        } catch (e: any) {
            console.error('Push failed', e);
            syncError.value = 'Push failed: ' + (e.message || e.toString());
            return null;
        } finally {
            isPushing.value = false;
        }
    }

    /**
     * Read local file items for preview/comparison.
     */
    async function readLocalItems(projectId: string): Promise<{ name: string; value: string }[]> {
        const mapping = await FileMappingService.getMapping(projectId);
        if (!mapping) {
            throw new Error('No file mapping found');
        }
        return FileMappingService.readLocalFile(mapping.filePath);
    }

    return {
        isPulling,
        isPushing,
        syncError,
        pullToLocal,
        pullWithState,
        pushToRemote,
        pushWithState,
        readLocalItems,
    };
}
