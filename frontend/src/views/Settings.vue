<script setup lang="ts">
import { ref, computed } from 'vue';
import { useRouter } from 'vue-router';
import { invoke } from '@tauri-apps/api/core';
import { Button } from '@/components/ui/button';
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { useAuthStore } from '../stores/auth';
import { useVaultStore } from '../stores/vault';
import { LogOut, Trash2, Copy, Check, AlertTriangle, RefreshCw } from 'lucide-vue-next';
import { IdentityService } from "@/services/identity.service.ts";
import { EncryptionService } from "@/services/encryption.service.ts";
import { TeamService } from "@/services/team.service.ts";
import { DeviceService } from "@/services/device.service.ts";
import type { RotateMasterKeyRequest } from "@/services/auth.service.ts";
import { x25519 } from '@noble/curves/ed25519.js';

const auth = useAuthStore();
const vaultStore = useVaultStore();
const router = useRouter();

// Display the Master Identity public key (used for team encryption), not the device vault key
const publicKey = computed(() => {
    const masterKeyPair = IdentityService.getMasterKeyPair();
    return masterKeyPair?.publicKey || '';
});
const isLoading = ref(false);
const rotationStatus = ref('');
const error = ref('');
const copiedPublic = ref(false);

// Key rotation modal state
const showRotationModal = ref(false);
const newRecoveryPhrase = ref('');
const copiedPhrase = ref(false);
const phraseConfirmed = ref(false);

async function generateKeys() {
    isLoading.value = true;
    error.value = '';
    rotationStatus.value = '';

    try {
        const oldMasterKeyPair = IdentityService.getMasterKeyPair();
        if (!oldMasterKeyPair) {
            throw new Error('Master Identity not loaded');
        }

        // 1. Generate new recovery phrase and derive new master key
        rotationStatus.value = 'Generating new recovery phrase...';
        const newPhrase = IdentityService.generateRecoveryPhrase();
        const newMasterKey = await IdentityService.deriveMasterKey(newPhrase);

        // Get new keypair from the new master key
        const newPrivateBytes = atob(newMasterKey);
        const newPrivateArray = Uint8Array.from(newPrivateBytes, c => c.charCodeAt(0));
        const newPublicBytes = x25519.getPublicKey(newPrivateArray);
        const newPublicKey = btoa(String.fromCharCode(...newPublicBytes));

        // 2. Fetch all approved devices/identities
        rotationStatus.value = 'Fetching devices...';
        const devices = await DeviceService.getDevices();
        const approvedDevices = devices.filter(d => d.encryptedMasterKey);

        if (approvedDevices.length === 0) {
            throw new Error('No approved devices found. Cannot rotate keys.');
        }

        // 3. Fetch all teams
        rotationStatus.value = 'Fetching team memberships...';
        const teams = await TeamService.getMyTeams();

        // 4. Decrypt all team keys with old master private key
        rotationStatus.value = `Processing ${teams.length} team keys...`;
        const decryptedTeamKeys: { teamId: string; teamKey: string }[] = [];

        for (const team of teams) {
            try {
                const teamKey = await EncryptionService.decryptKey(
                    oldMasterKeyPair.privateKey,
                    team.encryptedTeamKey
                );
                decryptedTeamKeys.push({ teamId: team.teamId, teamKey });
            } catch (e) {
                console.warn(`Failed to decrypt team key for ${team.teamName}:`, e);
                throw new Error(`Failed to decrypt team key for "${team.teamName}". Key rotation aborted.`);
            }
        }

        // 5. Re-encrypt team keys with new master public key
        rotationStatus.value = 'Re-encrypting team keys...';
        const newTeamKeys: Record<string, string> = {};
        for (const { teamId, teamKey } of decryptedTeamKeys) {
            const encrypted = await EncryptionService.encryptKey(newPublicKey, teamKey);
            newTeamKeys[teamId] = encrypted;
        }

        // 6. Encrypt new master private key for each device
        rotationStatus.value = 'Encrypting keys for devices...';
        const newIdentityKeys: Record<string, string> = {};
        for (const device of approvedDevices) {
            // Encrypt the new master key with each device's public key
            const encrypted = await EncryptionService.encryptKey(device.publicKey, newMasterKey);
            newIdentityKeys[device.id] = encrypted;
        }

        // 7. Send rotation request to backend
        rotationStatus.value = 'Updating keys on server...';
        const request: RotateMasterKeyRequest = {
            newPublicKey: newPublicKey,
            identityKeys: newIdentityKeys,
            teamKeys: newTeamKeys
        };

        await auth.rotateMasterKey(request);

        // 8. Update local master key in memory
        IdentityService.setMasterKey(newMasterKey);

        // 9. Save new master key to vault (encrypted with device vault key)
        if (vaultStore.status === 'unlocked') {
            await vaultStore.saveMasterKey(newMasterKey);
        }

        // 10. Show the new recovery phrase to user
        newRecoveryPhrase.value = newPhrase;
        showRotationModal.value = true;
        rotationStatus.value = 'Key rotation complete! Please save your new recovery phrase.';

    } catch (err: any) {
        console.error(err);
        error.value = err.message || err.toString();
        rotationStatus.value = '';
    } finally {
        isLoading.value = false;
    }
}

async function copyRecoveryPhrase() {
    try {
        await navigator.clipboard.writeText(newRecoveryPhrase.value);
        copiedPhrase.value = true;
        setTimeout(() => copiedPhrase.value = false, 2000);
    } catch (err) {
        console.error('Failed to copy recovery phrase', err);
    }
}

function closeRotationModal() {
    showRotationModal.value = false;
    newRecoveryPhrase.value = '';
    phraseConfirmed.value = false;
    rotationStatus.value = '';
}

async function copyToClipboard(text: string) {
    try {
        await navigator.clipboard.writeText(text);
        copiedPublic.value = true;
        setTimeout(() => copiedPublic.value = false, 2000);
    } catch (err) {
        console.error('Failed to copy to clipboard', err);
    }
}

async function handleLogout() {
    isLoading.value = true;
    error.value = '';
    try {
        // 1. Clear vault memory state (Keep files)
        vaultStore.reset();
        
        // 2. Clear Identity memory
        IdentityService.clear();

        // 3. Logout from store (clears token/user)
        auth.logout();

        // 4. Redirect to login
        router.push('/');
        
    } catch (err: any) {
        console.error(err);
        error.value = "Failed to logout: " + err.toString();
    } finally {
        isLoading.value = false;
    }
}

async function handleDeleteData() {
    isLoading.value = true;
    error.value = '';
    try {
        // 1. Nuke local vault files via Rust
        await invoke('nuke_vault', { userId: vaultStore.userId || "" });

        // 2. Clear vault memory state
        vaultStore.reset();
        
        // 3. Clear Identity memory
        IdentityService.clear();

        // 4. Logout from store (clears token/user)
        auth.logout();

        // 5. Redirect to login
        router.push('/');
        
    } catch (err: any) {
        console.error(err);
        error.value = "Failed to clear data: " + err.toString();
    } finally {
        isLoading.value = false;
    }
}

function onCreated(){
    // Keys are already loaded if we are here (guarded by App.vue -> VaultUnlock)
}

onCreated()
</script>

<template>
    <div class="p-8 max-w-5xl mx-auto space-y-8">
        <div>
            <h1 class="text-3xl font-bold tracking-tight">Settings</h1>
            <p class="text-muted-foreground">Manage your account and security preferences.</p>
        </div>

        <div class="space-y-6">
            <!-- Key Pair Card -->
            <Card>
                <CardHeader>
                    <CardTitle>Key Pair</CardTitle>
                    <CardDescription>Manage your encryption keys.</CardDescription>
                </CardHeader>
                <CardContent class="space-y-4">
                    <div class="space-y-2">
                        <label class="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70">Public Key</label>
                        <div class="relative">
                            <input
                                :value="publicKey"
                                readonly
                                class="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50 pr-10"
                            />
                            <button
                                @click="copyToClipboard(publicKey)"
                                class="absolute right-2 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground transition-colors"
                                title="Copy Public Key"
                            >
                                <Check v-if="copiedPublic" class="w-4 h-4 text-green-500" />
                                <Copy v-else class="w-4 h-4" />
                            </button>
                        </div>
                    </div>
                    
                    
                    <div class="space-y-2">
                        <label class="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70">Private Key</label>
                        <div class="text-sm text-muted-foreground italic">
                            Securely stored in the vault.
                        </div>
                    </div>

                    <div v-if="error" class="text-sm text-destructive">
                        {{ error }}
                    </div>

                    <div v-if="rotationStatus" class="text-sm text-muted-foreground">
                        {{ rotationStatus }}
                    </div>

                    <div class="pt-2">
                        <Button @click="generateKeys" :disabled="isLoading" variant="outline">
                            <RefreshCw v-if="!isLoading" class="w-4 h-4 mr-2" />
                            {{ isLoading ? 'Rotating Keys...' : 'Rotate Master Key' }}
                        </Button>
                        <p class="text-xs text-muted-foreground mt-2">
                            Generates a new recovery phrase and re-encrypts all your team keys.
                            <br>
                            <span class="text-amber-500">You will need to write down the new recovery phrase.</span>
                        </p>
                    </div>
                </CardContent>
            </Card>

            <!-- Recovery Phrase Modal -->
            <div v-if="showRotationModal" class="fixed inset-0 bg-black/80 flex items-center justify-center z-50">
                <Card class="w-full max-w-lg mx-4">
                    <CardHeader>
                        <div class="flex items-center gap-2 text-amber-500">
                            <AlertTriangle class="w-5 h-5" />
                            <CardTitle>New Recovery Phrase</CardTitle>
                        </div>
                        <CardDescription>
                            Your master key has been rotated. Write down this new recovery phrase and store it safely.
                            <span class="font-bold text-destructive">You will not see this again!</span>
                        </CardDescription>
                    </CardHeader>
                    <CardContent class="space-y-4">
                        <div class="bg-muted p-4 rounded-lg">
                            <div class="font-mono text-lg leading-relaxed select-all">
                                {{ newRecoveryPhrase }}
                            </div>
                        </div>

                        <div class="flex gap-2">
                            <Button @click="copyRecoveryPhrase" variant="outline" class="flex-1">
                                <Check v-if="copiedPhrase" class="w-4 h-4 mr-2 text-green-500" />
                                <Copy v-else class="w-4 h-4 mr-2" />
                                {{ copiedPhrase ? 'Copied!' : 'Copy Phrase' }}
                            </Button>
                        </div>

                        <div class="flex items-start gap-2">
                            <input
                                type="checkbox"
                                id="confirmPhrase"
                                v-model="phraseConfirmed"
                                class="mt-1"
                            />
                            <label for="confirmPhrase" class="text-sm text-muted-foreground">
                                I have written down my recovery phrase and stored it in a safe place.
                            </label>
                        </div>

                        <Button @click="closeRotationModal" class="w-full" :variant="phraseConfirmed ? 'default' : 'outline'">
                            {{ phraseConfirmed ? 'Done' : 'Close (Not Recommended)' }}
                        </Button>
                    </CardContent>
                </Card>
            </div>

            <!-- Danger Zone Card -->
            <Card class="border-destructive/50">
                <CardHeader>
                    <CardTitle class="text-destructive">Danger Zone</CardTitle>
                    <CardDescription>
                        Manage sensitive actions for your account and device.
                    </CardDescription>
                </CardHeader>
                <CardContent class="space-y-6">
                    
                    <!-- Standard Logout -->
                    <div class="flex items-center justify-between">
                        <div class="space-y-1">
                            <div class="font-medium">Log Out</div>
                            <div class="text-sm text-muted-foreground">
                                Signs you out and locks your local vault. Your data remains on this device.
                            </div>
                        </div>
                        <Button 
                            variant="outline" 
                            @click="handleLogout" 
                            :disabled="isLoading"
                            class="ml-4"
                        >
                            <LogOut class="w-4 h-4 mr-2" />
                            {{ isLoading ? 'Processing...' : 'Log Out' }}
                        </Button>
                    </div>

                    <div class="border-t border-border"></div>

                    <!-- Delete All Data -->
                    <div class="flex items-center justify-between">
                        <div class="space-y-1">
                            <div class="font-medium text-destructive">Delete All Data</div>
                            <div class="text-sm text-muted-foreground">
                                Permanently deletes your local vault, keys, and logs you out. 
                                <br>
                                <span class="text-destructive text-xs font-bold">WARNING: This action cannot be undone.</span>
                            </div>
                        </div>
                        <Button 
                            variant="destructive" 
                            @click="handleDeleteData" 
                            :disabled="isLoading"
                            class="ml-4"
                        >
                            <Trash2 class="w-4 h-4 mr-2" />
                            {{ isLoading ? 'Processing...' : 'Delete Data' }}
                        </Button>
                    </div>

                </CardContent>
            </Card>
        </div>
    </div>
</template>
