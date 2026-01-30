import { Store } from '@tauri-apps/plugin-store';
import { invoke } from '@tauri-apps/api/core';
import { parse as parseDotenv } from 'dotenv';
import { IdentityService } from './identity.service';
import type { ConfigItem } from './project.service';
import { buildEnvString } from '@/utils/env-format';

// Custom Tauri commands for file operations (bypass fs plugin scope restrictions)
async function readTextFile(path: string): Promise<string> {
    return invoke<string>('read_text_file_absolute', { path });
}

async function writeTextFile(path: string, contents: string): Promise<void> {
    return invoke('write_text_file_absolute', { path, contents });
}

async function exists(path: string): Promise<boolean> {
    return invoke<boolean>('file_exists_absolute', { path });
}

export interface FileMapping {
    projectId: string;
    filePath: string;
    lastLocalChecksum: string;
    lastRemoteChecksum: string;
    linkedAt: string;
    devicePublicKey: string;
}

export type SyncStatus =
    | 'synced'           // Both checksums match
    | 'local_changed'    // Local file changed, remote unchanged
    | 'remote_changed'   // Remote changed, local unchanged
    | 'both_changed'     // Both changed
    | 'file_missing'     // Linked file no longer exists
    | 'not_linked';      // No mapping exists

export interface SyncStatusResult {
    status: SyncStatus;
    mapping?: FileMapping;
    currentLocalChecksum?: string;
    currentRemoteChecksum?: string;
}

const STORE_NAME = 'file-mappings.json';

export class FileMappingService {
    private static store: Store | null = null;

    private static async getStore(): Promise<Store> {
        if (!this.store) {
            this.store = await Store.load(STORE_NAME);
        }
        return this.store;
    }

    private static getDevicePublicKey(): string {
        const keyPair = IdentityService.getMasterKeyPair();
        if (!keyPair) {
            throw new Error('Identity not loaded. Please unlock your vault first.');
        }
        return keyPair.publicKey;
    }

    private static getMappingKey(projectId: string): string {
        const deviceKey = this.getDevicePublicKey();
        return `mapping:${deviceKey}:${projectId}`;
    }

    /**
     * Link a project to a local .env file
     */
    static async linkFile(projectId: string, filePath: string, remoteChecksum: string): Promise<FileMapping> {
        const store = await this.getStore();
        const devicePublicKey = this.getDevicePublicKey();

        // Read and compute local file checksum
        let localChecksum = '';
        if (await exists(filePath)) {
            const content = await readTextFile(filePath);
            localChecksum = await this.computeLocalFileChecksum(content);
        }

        const mapping: FileMapping = {
            projectId,
            filePath,
            lastLocalChecksum: localChecksum,
            lastRemoteChecksum: remoteChecksum,
            linkedAt: new Date().toISOString(),
            devicePublicKey,
        };

        const key = this.getMappingKey(projectId);
        await store.set(key, mapping);
        await store.save();

        return mapping;
    }

    /**
     * Unlink a project from its local .env file
     */
    static async unlinkFile(projectId: string): Promise<void> {
        const store = await this.getStore();
        const key = this.getMappingKey(projectId);
        await store.delete(key);
        await store.save();
    }

    /**
     * Get the mapping for a specific project on this device
     */
    static async getMapping(projectId: string): Promise<FileMapping | null> {
        const store = await this.getStore();
        const key = this.getMappingKey(projectId);
        const mapping = await store.get<FileMapping>(key);
        return mapping ?? null;
    }

    /**
     * Get all mappings for the current device
     */
    static async getAllMappings(): Promise<FileMapping[]> {
        const store = await this.getStore();
        const devicePublicKey = this.getDevicePublicKey();
        const prefix = `mapping:${devicePublicKey}:`;

        const mappings: FileMapping[] = [];
        const keys = await store.keys();

        for (const key of keys) {
            if (key.startsWith(prefix)) {
                const mapping = await store.get<FileMapping>(key);
                if (mapping) {
                    mappings.push(mapping);
                }
            }
        }

        return mappings;
    }

    /**
     * Check the sync status of a linked project
     */
    static async checkSyncStatus(projectId: string, currentRemoteChecksum: string): Promise<SyncStatusResult> {
        const mapping = await this.getMapping(projectId);

        if (!mapping) {
            return { status: 'not_linked' };
        }

        // Check if file exists
        if (!await exists(mapping.filePath)) {
            return {
                status: 'file_missing',
                mapping,
                currentRemoteChecksum,
            };
        }

        // Read current local file and compute checksum
        const content = await readTextFile(mapping.filePath);
        const currentLocalChecksum = await this.computeLocalFileChecksum(content);

        const localChanged = currentLocalChecksum !== mapping.lastLocalChecksum;
        const remoteChanged = currentRemoteChecksum !== mapping.lastRemoteChecksum;

        let status: SyncStatus;
        if (localChanged && remoteChanged) {
            status = 'both_changed';
        } else if (localChanged) {
            status = 'local_changed';
        } else if (remoteChanged) {
            status = 'remote_changed';
        } else {
            status = 'synced';
        }

        return {
            status,
            mapping,
            currentLocalChecksum,
            currentRemoteChecksum,
        };
    }

    /**
     * Update the sync state after a successful sync operation
     */
    static async updateSyncState(projectId: string, localChecksum: string, remoteChecksum: string): Promise<void> {
        const mapping = await this.getMapping(projectId);
        if (!mapping) {
            throw new Error('No mapping found for project');
        }

        const store = await this.getStore();
        const updatedMapping: FileMapping = {
            ...mapping,
            lastLocalChecksum: localChecksum,
            lastRemoteChecksum: remoteChecksum,
        };

        const key = this.getMappingKey(projectId);
        await store.set(key, updatedMapping);
        await store.save();
    }

    /**
     * Write config items to a local .env file (Pull operation)
     * Includes category comments matching the copy .env format
     */
    static async writeToLocalFile(filePath: string, items: ConfigItem[]): Promise<string> {
        // Sort items by position
        const sortedItems = [...items].sort((a, b) => a.position - b.position);

        // Extract unique categories in order of first appearance
        const categories: string[] = [];
        sortedItems.forEach(item => {
            if (item.category && !categories.includes(item.category)) {
                categories.push(item.category);
            }
        });

        // Helper functions for buildEnvString
        const getCategoryItems = (category: string) =>
            sortedItems.filter(item => item.category === category);

        const getUncategorizedItems = () =>
            sortedItems.filter(item => !item.category);

        // Build .env content with category comments
        const content = buildEnvString(sortedItems, categories, getCategoryItems, getUncategorizedItems);
        await writeTextFile(filePath, content + '\n');

        return await this.computeLocalFileChecksum(content);
    }

    /**
     * Read a local .env file and parse it into config-like items (Push operation)
     */
    static async readLocalFile(filePath: string): Promise<{ name: string; value: string }[]> {
        const content = await readTextFile(filePath);
        return this.parseEnvContent(content);
    }

    /**
     * Parse .env file content into name-value pairs.
     * Uses dotenv library for parsing, then unescapes values.
     */
    static parseEnvContent(content: string): { name: string; value: string }[] {
        const parsed = parseDotenv(content);
        return Object.entries(parsed).map(([name, value]) => ({
            name,
            value: this.unescapeEnvValue(value ?? ''),
        }));
    }

    /**
     * Unescape a value that was read from a .env file.
     * The dotenv library preserves escape sequences, so we need to unescape them.
     */
    private static unescapeEnvValue(val: string): string {
        // Unescape common escape sequences that formatEnvValue creates
        return val
            .replace(/\\n/g, '\n')
            .replace(/\\r/g, '\r')
            .replace(/\\"/g, '"')
            .replace(/\\\\/g, '\\');
    }

    /**
     * Compute SHA256 checksum from config items (matches backend algorithm)
     */
    static async computeChecksum(items: ConfigItem[]): Promise<string> {
        // Sort by position
        const sortedItems = [...items].sort((a, b) => a.position - b.position);

        // Build string: name=value joined by newlines
        const lines = sortedItems.map(item => `${item.name}=${item.value}`);
        const content = lines.join('\n');

        // Compute SHA256
        const encoder = new TextEncoder();
        const data = encoder.encode(content);
        const hashBuffer = await crypto.subtle.digest('SHA-256', data);
        const hashArray = Array.from(new Uint8Array(hashBuffer));
        return hashArray.map(b => b.toString(16).padStart(2, '0')).join('');
    }

    /**
     * Compute checksum from raw .env file content
     */
    static async computeLocalFileChecksum(content: string): Promise<string> {
        const items = this.parseEnvContent(content);
        // Convert to ConfigItem-like format with positions
        const configItems = items.map((item, index) => ({
            ...item,
            id: '',
            projectId: '',
            sensitive: false,
            position: index,
        })) as ConfigItem[];

        return this.computeChecksum(configItems);
    }
}
