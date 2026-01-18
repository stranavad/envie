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
import { LogOut, Trash2, Copy, Check } from 'lucide-vue-next';
import { IdentityService } from "@/services/identity.service.ts";
import { EncryptionService } from "@/services/encryption.service.ts";
import { TeamService } from "@/services/team.service";

const auth = useAuthStore();
const vaultStore = useVaultStore();
const router = useRouter();

const publicKey = computed(() => vaultStore.publicKey);
const isLoading = ref(false);
const rotationStatus = ref('');
const error = ref('');
const copiedPublic = ref(false);

async function generateKeys() {
    isLoading.value = true;
    error.value = '';
    rotationStatus.value = '';

    try {
        // 1. Get all teams the user is a member of
        rotationStatus.value = 'Fetching team memberships...';
        const teams = await TeamService.getMyTeams();

        // 2. Decrypt all team keys with current private key
        rotationStatus.value = `Decrypting ${teams.length} team keys...`;
        const decryptedTeamKeys: { teamId: string; teamKey: string }[] = [];

        for (const team of teams) {
            if (team.encryptedTeamKey) {
                try {
                    const teamKey = await EncryptionService.decryptKey(
                        vaultStore.privateKey!,
                        team.encryptedTeamKey
                    );
                    decryptedTeamKeys.push({ teamId: team.teamId, teamKey });
                } catch (e) {
                    console.warn(`Failed to decrypt team key for ${team.teamName}:`, e);
                }
            }
        }

        // 3. Generate new key pair
        rotationStatus.value = 'Generating new key pair...';
        await vaultStore.rotateKeys();

        // 4. Re-encrypt all team keys with new public key
        rotationStatus.value = `Re-encrypting ${decryptedTeamKeys.length} team keys...`;
        for (const { teamId, teamKey } of decryptedTeamKeys) {
            const newEncryptedTeamKey = await EncryptionService.encryptKey(
                vaultStore.publicKey!,
                teamKey
            );

            // 5. Save re-encrypted team key to backend
            await TeamService.updateMyTeamKey(teamId, {
                encryptedTeamKey: newEncryptedTeamKey
            });
        }

        // 6. Sync new public key to backend
        rotationStatus.value = 'Syncing new public key...';
        if (vaultStore.publicKey) {
            await auth.updatePublicKey(vaultStore.publicKey);
        }

        rotationStatus.value = 'Key rotation complete!';
        setTimeout(() => { rotationStatus.value = ''; }, 3000);

    } catch (err: any) {
        console.error(err);
        error.value = err.toString();
        rotationStatus.value = '';
    } finally {
        isLoading.value = false;
    }
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
    if(!confirm("Are you sure? This will delete your local vault, keys, and log you out. This action cannot be undone.")) {
        return;
    }

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
                        <Button @click="generateKeys" :disabled="isLoading">
                            {{ isLoading ? 'Rotating...' : 'Regenerate Key Pair' }}
                        </Button>
                        <p class="text-xs text-muted-foreground mt-2">
                            This will re-encrypt all your team keys with the new key pair.
                        </p>
                    </div>
                </CardContent>
            </Card>

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
