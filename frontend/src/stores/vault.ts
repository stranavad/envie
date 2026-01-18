import { defineStore } from 'pinia';
import { ref } from 'vue';
import { Client, Stronghold } from '@tauri-apps/plugin-stronghold';
import { invoke } from '@tauri-apps/api/core';
import { EncryptionService } from "@/services/encryption.service.ts";
import { IdentityService } from "@/services/identity.service.ts";
import { appDataDir } from '@tauri-apps/api/path';

const CLIENT_NAME = 'primary';
const KEY_RECORD_KEY = 'keypair';
const MASTER_KEY_RECORD_KEY = 'master_key';
const REFRESH_TOKEN_KEY = 'refresh_token';

export type VaultStatus = 'loading' | 'uninitialized' | 'locked' | 'unlocked';

interface KeyPair {
    privateKey: string;
    publicKey: string;
}

export const useVaultStore = defineStore('vault', () => {
    const status = ref<VaultStatus>('loading');
    const error = ref('');
    const privateKey = ref<string>('');
    const publicKey = ref<string>('');
    const userId = ref<string>(''); // Current User ID target

    // Internal Client reference
    let client: Client | null = null;
    let stronghold: Stronghold | null = null;

    function setUserId(id: string) {
        userId.value = id;
    }

    async function loadStronghold(password: string): Promise<Stronghold> {
        if (!stronghold) {
            // User-specific vault file
            const vaultFile = userId.value ? `vault_${userId.value}.hold` : `snapshot.hold`;
            const vaultPath = `${await appDataDir()}/${vaultFile}`;
            stronghold = await Stronghold.load(vaultPath, password);
        }
        return stronghold;
    }

    async function checkStatus() {
        if (!userId.value) {
            status.value = 'uninitialized'; // Or 'unknown'?
            return;
        }

        status.value = 'loading';
        error.value = '';
        try {
            // Check if vault file for THIS user exists via Rust Command
            const exists = await invoke<boolean>('check_vault_exists', { userId: userId.value });
            if (exists) {
                status.value = 'locked';
            } else {
                status.value = 'uninitialized';
            }
        } catch (e: any) {
            console.error("Status check error", e);
            status.value = 'locked';
        }
    }

    async function initVault(password: string) {
        if (!userId.value) throw new Error("User ID not set for vault initialization");

        status.value = 'loading';
        error.value = '';
        try {
            const sh = await loadStronghold(password);

            client = await sh.createClient(CLIENT_NAME);

            // Generate Keys via Rust (stateless)
            const keys = EncryptionService.generateKeyPair()

            // Save to Stronghold Store
            const store = client.getStore();
            const value = Array.from(new TextEncoder().encode(JSON.stringify(keys)));
            await store.insert(KEY_RECORD_KEY, value);

            // Persist
            await sh.save();

            // Set State
            privateKey.value = keys.privateKey;
            publicKey.value = keys.publicKey;
            status.value = 'unlocked';

        } catch (e: any) {
            console.error(e);
            error.value = "Setup failed: " + e.toString();
            status.value = 'uninitialized';
            stronghold = null; // Reset
        }
    }

    async function saveMasterKey(keyB64: string) {
        if (status.value !== 'unlocked' || !client || !stronghold) {
            throw new Error("Vault must be unlocked to save master key.");
        }
        try {
            const store = client.getStore();
            const value = Array.from(new TextEncoder().encode(keyB64));
            await store.insert(MASTER_KEY_RECORD_KEY, value);
            await stronghold.save();
        } catch (e) {
            console.error("Failed to save master key to vault", e);
            throw e;
        }
    }

    async function saveRefreshToken(token: string) {
        if (status.value !== 'unlocked' || !client || !stronghold) {
            throw new Error("Vault must be unlocked to save refresh token.");
        }
        try {
            const store = client.getStore();
            const value = Array.from(new TextEncoder().encode(token));
            await store.insert(REFRESH_TOKEN_KEY, value);
            await stronghold.save();
        } catch (e) {
            console.error("Failed to save refresh token to vault", e);
            throw e;
        }
    }

    async function getRefreshToken(): Promise<string | null> {
        if (status.value !== 'unlocked' || !client) {
            return null;
        }
        try {
            const store = client.getStore();
            const data = await store.get(REFRESH_TOKEN_KEY);
            if (!data) return null;
            return new TextDecoder().decode(new Uint8Array(data));
        } catch (e) {
            console.warn("Failed to get refresh token from vault", e);
            return null;
        }
    }

    async function clearRefreshToken() {
        if (status.value !== 'unlocked' || !client || !stronghold) {
            return;
        }
        try {
            const store = client.getStore();
            // Remove by inserting empty value (Stronghold doesn't have explicit delete for store)
            await store.insert(REFRESH_TOKEN_KEY, []);
            await stronghold.save();
        } catch (e) {
            console.warn("Failed to clear refresh token from vault", e);
        }
    }

    async function unlockVault(password: string) {
        if (!userId.value) throw new Error("User ID not set for vault unlock");

        status.value = 'loading';
        error.value = '';
        try {
            // Attempt to load stronghold. 
            // If file exists & password wrong -> Throws error.
            // If file missing -> Creates new instance in memory (succeeds).
            const sh = await loadStronghold(password);

            try {
                client = await sh.loadClient(CLIENT_NAME);

                // Read Keys
                const store = client.getStore();
                const data = await store.get(KEY_RECORD_KEY);

                if (!data) {
                    throw new Error("No keys found.");
                }

                // Decode
                const decoded = new TextDecoder().decode(new Uint8Array(data));
                const keys: KeyPair = JSON.parse(decoded);

                privateKey.value = keys.privateKey;
                publicKey.value = keys.publicKey;

                // Try Load Master Key (if exists)
                try {
                    const masterKeyData = await store.get(MASTER_KEY_RECORD_KEY);
                    if (masterKeyData) {
                        const masterKey = new TextDecoder().decode(new Uint8Array(masterKeyData));
                        IdentityService.setMasterKey(masterKey);
                        console.log("Master Identity Key loaded from Vault.");
                    } else {
                        console.warn("No Master Identity Key found in Vault.");
                    }
                } catch (mkErr) {
                    console.warn("Failed to load Master Key", mkErr);
                }

                status.value = 'unlocked';

            } catch (innerErr: any) {
                // Determine if this is a "client missing" or "key missing" issue
                // which implies uninitialized vault (or corrupt).
                // Since we passed the password check (Stronghold loaded), 
                // we treat this as "Setup required".
                console.warn("Vault loaded but content missing. Switching to setup.", innerErr);
                status.value = 'uninitialized';
                stronghold = null; // Reset
            }

        } catch (e: any) {
            console.error(e);
            error.value = "Unlock failed. Incorrect password or corrupt vault.";
            status.value = 'locked';
            stronghold = null;
        }
    }

    async function rotateKeys() {
        if (status.value !== 'unlocked' || !client || !stronghold) {
            throw new Error("Vault must be unlocked to rotate keys.");
        }

        try {
            // Generate New Keys
            const keys = EncryptionService.generateKeyPair()

            // Overwrite in Store
            const store = client.getStore();
            const value = Array.from(new TextEncoder().encode(JSON.stringify(keys)));
            // Insert overwrites if key exists? Yes usually for KV. 
            // Documentation: "Insert a record...". If it exists?
            // "Stronghold Store allows... create, update...". 
            // We assume insert works for update.
            await store.insert(KEY_RECORD_KEY, value);

            await stronghold.save();

            privateKey.value = keys.privateKey;
            publicKey.value = keys.publicKey;

        } catch (e: any) {
            console.error("Rotate keys failed", e);
            throw e;
        }
    }

    function reset() {
        // Clears memory state only. File persistence remains. 
        // Real absolute reset would require deleting `vault.hold` and `salt.txt`.
        stronghold = null;
        client = null;
        status.value = 'loading';
        privateKey.value = '';
        publicKey.value = '';
        // Do NOT call checkStatus automatically here, user might be changing
    }

    return {
        status,
        error,
        privateKey,
        publicKey,
        userId,
        setUserId,
        checkStatus,
        initVault,
        unlockVault,
        rotateKeys,
        reset,
        saveMasterKey,
        saveRefreshToken,
        getRefreshToken,
        clearRefreshToken
    };
});
