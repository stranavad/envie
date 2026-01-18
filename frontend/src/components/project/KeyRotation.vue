<script setup lang="ts">
import { ref, onMounted, computed } from 'vue';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Avatar, AvatarFallback } from '@/components/ui/avatar';
import { KeyRound, RefreshCw, Check, X, Clock, AlertTriangle, ShieldCheck } from 'lucide-vue-next';
import { ProjectService, type ProjectDetail } from '@/services/project.service';
import { KeyRotationService, type PendingRotation } from '@/services/key-rotation.service';
import { EncryptionService } from '@/services/encryption.service';

const props = defineProps<{
    project: ProjectDetail;
    decryptedKey: string;
    teamKey: string;
}>();

const emit = defineEmits<{
    (e: 'rotated'): void;
}>();

const pendingRotation = ref<PendingRotation | null>(null);
const isLoading = ref(false);
const isRotating = ref(false);
const isApproving = ref(false);
const isVerifying = ref(false);
const error = ref('');
const success = ref('');
const staleRotationExists = ref(false);

// Verification state
const verificationResult = ref<{
    verified: boolean;
    failedItems: string[];
    newKey: string;
} | null>(null);

const canInitiateRotation = computed(() => props.project.canEdit && !pendingRotation.value);
const canApprove = computed(() => {
    if (!pendingRotation.value || !props.project.canEdit) return false;
    return true;
});

async function loadPendingRotation() {
    isLoading.value = true;
    error.value = '';
    try {
        const result = await KeyRotationService.getPendingRotation(props.project.id);
        pendingRotation.value = result.pending;
        staleRotationExists.value = result.staleRotationExists || false;
    } catch (e: any) {
        error.value = 'Failed to load rotation status: ' + (e.message || e);
    } finally {
        isLoading.value = false;
    }
}

async function initiateRotation() {
    isRotating.value = true;
    error.value = '';
    success.value = '';

    try {
        // 1. Generate new random project key
        const newProjectKey = EncryptionService.generateProjectKey();

        // 2. Load and re-encrypt all config items
        const configs = await ProjectService.getConfig(props.project.id);
        const reEncryptedItems: { id: string; value: string }[] = [];

        for (const item of configs) {
            // Decrypt with old key
            const decrypted = await EncryptionService.decryptValue(props.decryptedKey, item.value);
            // Re-encrypt with new key
            const reEncrypted = await EncryptionService.encryptValue(newProjectKey, decrypted);
            reEncryptedItems.push({ id: item.id, value: reEncrypted });
        }

        // 3. Load and re-encrypt all file FEKs
        const fileFEKs = await ProjectService.getFilesForRotation(props.project.id);
        const reEncryptedFileFEKs: { id: string; encryptedFek: string }[] = [];

        for (const file of fileFEKs) {
            // Decrypt FEK with old project key
            const decryptedFEK = await EncryptionService.decryptValue(props.decryptedKey, file.encryptedFek);
            // Re-encrypt FEK with new project key
            const reEncryptedFEK = await EncryptionService.encryptValue(newProjectKey, decryptedFEK);
            reEncryptedFileFEKs.push({ id: file.id, encryptedFek: reEncryptedFEK });
        }

        // 4. Get teams with access and encrypt new key for each
        const accessData = await ProjectService.getProjectTeams(props.project.id);
        const teamsWithAccess = accessData.teams || [];

        const teamEncryptedKeys: { teamId: string; encryptedProjectKey: string }[] = [];

        for (const team of teamsWithAccess) {
            // For multi-team scenario, we need each team's key
            // For now, we only have access to our team's key
            // TODO: For multi-team support, need to fetch each team's encrypted key and decrypt
            const encryptedProjectKey = await EncryptionService.encryptValue(props.teamKey, newProjectKey);
            teamEncryptedKeys.push({
                teamId: team.id,
                encryptedProjectKey
            });
        }

        // 5. Call API to initiate rotation
        const result = await KeyRotationService.initiateRotation(props.project.id, {
            teamEncryptedKeys,
            reEncryptedConfigItems: reEncryptedItems,
            reEncryptedFileFEKs: reEncryptedFileFEKs.length > 0 ? reEncryptedFileFEKs : undefined
        });

        if (result.committed) {
            success.value = 'Key rotation completed successfully!';
            emit('rotated');
        } else {
            success.value = `Key rotation initiated. Awaiting ${result.requiredApprovals} approval(s).`;
            await loadPendingRotation();
        }

    } catch (e: any) {
        error.value = 'Failed to initiate rotation: ' + (e.message || e);
    } finally {
        isRotating.value = false;
    }
}

/**
 * Zero-trust verification: Compare old decrypted values with new decrypted values
 */
async function verifyRotation() {
    if (!pendingRotation.value) return;

    isVerifying.value = true;
    error.value = '';
    verificationResult.value = null;

    try {
        // Parse the encrypted configs snapshot from the pending rotation
        const reEncryptedItems: { id: string; value: string }[] = JSON.parse(
            pendingRotation.value.encryptedConfigsSnapshot
        );

        // Parse team encrypted keys to get our team's new key
        const teamEncryptedKeys: { teamId: string; encryptedProjectKey: string }[] = JSON.parse(
            pendingRotation.value.teamEncryptedKeys
        );

        // Find our team's encrypted key (we need to determine our team ID)
        // For now, we'll try each one until one decrypts successfully
        let newProjectKey = '';
        for (const tk of teamEncryptedKeys) {
            try {
                newProjectKey = await EncryptionService.decryptValue(props.teamKey, tk.encryptedProjectKey);
                break;
            } catch {
                // Not our team's key, continue
            }
        }

        if (!newProjectKey) {
            throw new Error('Could not decrypt new project key - you may not have access');
        }

        // Load current config items
        const currentConfigs = await ProjectService.getConfig(props.project.id);

        const failedItems: string[] = [];

        // Verify each item
        for (const reEncryptedItem of reEncryptedItems) {
            const currentItem = currentConfigs.find(c => c.id === reEncryptedItem.id);
            if (!currentItem) {
                failedItems.push(`Item ${reEncryptedItem.id}: not found in current config`);
                continue;
            }

            try {
                // Decrypt current value with old key
                const currentDecrypted = await EncryptionService.decryptValue(props.decryptedKey, currentItem.value);

                // Decrypt proposed value with new key
                const proposedDecrypted = await EncryptionService.decryptValue(newProjectKey, reEncryptedItem.value);

                // Compare
                if (currentDecrypted !== proposedDecrypted) {
                    failedItems.push(`${currentItem.name}: values don't match (possible tampering)`);
                }
            } catch (e: any) {
                failedItems.push(`${currentItem.name}: decryption failed - ${e.message}`);
            }
        }

        verificationResult.value = {
            verified: failedItems.length === 0,
            failedItems,
            newKey: newProjectKey
        };

    } catch (e: any) {
        error.value = 'Verification failed: ' + (e.message || e);
    } finally {
        isVerifying.value = false;
    }
}

async function approveRotation() {
    if (!pendingRotation.value) return;

    // Require verification before approval
    if (!verificationResult.value?.verified) {
        error.value = 'Please verify the rotation before approving';
        return;
    }

    isApproving.value = true;
    error.value = '';
    success.value = '';

    try {
        const result = await KeyRotationService.approveRotation(
            props.project.id,
            pendingRotation.value.id,
            true // verifiedDecryption
        );

        if (result.committed) {
            success.value = 'Key rotation approved and completed!';
            pendingRotation.value = null;
            verificationResult.value = null;
            emit('rotated');
        } else {
            success.value = 'Approval recorded. Waiting for more approvals.';
            await loadPendingRotation();
        }
    } catch (e: any) {
        error.value = 'Failed to approve: ' + (e.message || e);
    } finally {
        isApproving.value = false;
    }
}

async function rejectRotation() {
    if (!pendingRotation.value) return;

    isApproving.value = true;
    error.value = '';

    try {
        const reason = verificationResult.value?.failedItems.length
            ? `Verification failed: ${verificationResult.value.failedItems.join(', ')}`
            : undefined;

        await KeyRotationService.rejectRotation(props.project.id, pendingRotation.value.id, reason);
        success.value = 'Rotation rejected.';
        pendingRotation.value = null;
        verificationResult.value = null;
    } catch (e: any) {
        error.value = 'Failed to reject: ' + (e.message || e);
    } finally {
        isApproving.value = false;
    }
}

async function cancelRotation() {
    if (!pendingRotation.value) return;

    isApproving.value = true;
    error.value = '';

    try {
        await KeyRotationService.cancelRotation(props.project.id, pendingRotation.value.id);
        success.value = 'Rotation cancelled.';
        pendingRotation.value = null;
        verificationResult.value = null;
    } catch (e: any) {
        error.value = 'Failed to cancel: ' + (e.message || e);
    } finally {
        isApproving.value = false;
    }
}

function formatDate(dateStr: string): string {
    return new Date(dateStr).toLocaleString();
}

onMounted(() => {
    loadPendingRotation();
});
</script>

<template>
    <Card>
        <CardHeader>
            <CardTitle class="flex items-center gap-2">
                <KeyRound class="w-5 h-5" />
                Key Rotation
            </CardTitle>
            <CardDescription>
                Rotate the project encryption key for enhanced security.
            </CardDescription>
        </CardHeader>
        <CardContent class="space-y-4">
            <!-- Current Status -->
            <div class="flex items-center justify-between p-4 bg-muted/50 rounded-lg">
                <div>
                    <p class="text-sm font-medium">Current Key Version</p>
                    <p class="text-2xl font-bold">v{{ project.keyVersion || 1 }}</p>
                </div>
                <div class="text-right">
                    <p class="text-sm text-muted-foreground">Random key encryption</p>
                    <Check class="w-5 h-5 text-green-500 inline" />
                </div>
            </div>

            <!-- Stale Rotation Warning -->
            <div v-if="staleRotationExists" class="bg-orange-500/15 text-orange-700 p-3 rounded-md text-sm">
                <AlertTriangle class="w-4 h-4 inline mr-2" />
                A previous rotation became stale due to config changes. You can initiate a new rotation.
            </div>

            <!-- Error/Success Messages -->
            <div v-if="error" class="bg-destructive/15 text-destructive p-3 rounded-md text-sm">
                {{ error }}
            </div>
            <div v-if="success" class="bg-green-500/15 text-green-700 p-3 rounded-md text-sm">
                {{ success }}
            </div>

            <!-- Loading State -->
            <div v-if="isLoading" class="text-center py-4 text-muted-foreground">
                Loading rotation status...
            </div>

            <!-- Pending Rotation -->
            <div v-else-if="pendingRotation" class="border rounded-lg p-4 space-y-4">
                <div class="flex items-center justify-between">
                    <div class="flex items-center gap-2">
                        <Clock class="w-5 h-5 text-orange-500" />
                        <span class="font-medium">Pending Rotation to v{{ pendingRotation.newVersion }}</span>
                    </div>
                    <span class="inline-flex items-center rounded-full border border-orange-600 px-2.5 py-0.5 text-xs font-semibold text-orange-600">
                        Awaiting Approval
                    </span>
                </div>

                <div class="grid gap-2 text-sm">
                    <div class="flex justify-between">
                        <span class="text-muted-foreground">Initiated by:</span>
                        <div class="flex items-center gap-2">
                            <Avatar class="h-5 w-5">
                                <AvatarFallback>{{ pendingRotation.initiator.name.charAt(0) }}</AvatarFallback>
                            </Avatar>
                            <span>{{ pendingRotation.initiator.name }}</span>
                        </div>
                    </div>
                    <div class="flex justify-between">
                        <span class="text-muted-foreground">Created:</span>
                        <span>{{ formatDate(pendingRotation.createdAt) }}</span>
                    </div>
                    <div class="flex justify-between">
                        <span class="text-muted-foreground">Expires:</span>
                        <span>{{ formatDate(pendingRotation.expiresAt) }}</span>
                    </div>
                    <div class="flex justify-between">
                        <span class="text-muted-foreground">Approvals:</span>
                        <span>{{ pendingRotation.approvals.filter(a => a.approved).length }} / {{ pendingRotation.requiredApprovals }}</span>
                    </div>
                </div>

                <!-- Approvals List -->
                <div v-if="pendingRotation.approvals.length > 0" class="space-y-2">
                    <p class="text-sm font-medium">Votes:</p>
                    <div v-for="approval in pendingRotation.approvals" :key="approval.id" class="flex items-center gap-2 text-sm">
                        <Check v-if="approval.approved" class="w-4 h-4 text-green-500" />
                        <X v-else class="w-4 h-4 text-red-500" />
                        <span>{{ approval.user.name }}</span>
                        <span v-if="approval.verifiedDecryption" class="text-xs text-green-600">(verified)</span>
                    </div>
                </div>

                <!-- Verification Section -->
                <div v-if="canApprove" class="border-t pt-4 space-y-3">
                    <div class="flex items-start gap-2">
                        <ShieldCheck class="w-5 h-5 text-blue-500 mt-0.5" />
                        <div class="text-sm">
                            <p class="font-medium">Zero-Trust Verification</p>
                            <p class="text-muted-foreground">
                                Verify that all config values decrypt correctly with the new key before approving.
                            </p>
                        </div>
                    </div>

                    <Button
                        variant="outline"
                        size="sm"
                        @click="verifyRotation"
                        :disabled="isVerifying"
                    >
                        <ShieldCheck class="w-4 h-4 mr-2" />
                        {{ isVerifying ? 'Verifying...' : 'Verify Decryption' }}
                    </Button>

                    <!-- Verification Result -->
                    <div v-if="verificationResult" class="p-3 rounded-md text-sm" :class="verificationResult.verified ? 'bg-green-500/15 text-green-700' : 'bg-destructive/15 text-destructive'">
                        <div class="flex items-center gap-2 font-medium">
                            <Check v-if="verificationResult.verified" class="w-4 h-4" />
                            <X v-else class="w-4 h-4" />
                            {{ verificationResult.verified ? 'Verification passed - all values match' : 'Verification failed' }}
                        </div>
                        <div v-if="verificationResult.failedItems.length > 0" class="mt-2 space-y-1">
                            <p v-for="item in verificationResult.failedItems" :key="item">{{ item }}</p>
                        </div>
                    </div>
                </div>

                <!-- Action Buttons -->
                <div class="flex gap-2 pt-2">
                    <Button
                        v-if="canApprove"
                        @click="approveRotation"
                        :disabled="isApproving || !verificationResult?.verified"
                        size="sm"
                    >
                        <Check class="w-4 h-4 mr-2" />
                        {{ isApproving ? 'Processing...' : 'Approve' }}
                    </Button>
                    <Button
                        v-if="canApprove"
                        variant="destructive"
                        @click="rejectRotation"
                        :disabled="isApproving"
                        size="sm"
                    >
                        <X class="w-4 h-4 mr-2" />
                        Reject
                    </Button>
                    <Button
                        variant="outline"
                        @click="cancelRotation"
                        :disabled="isApproving"
                        size="sm"
                    >
                        Cancel
                    </Button>
                </div>
            </div>

            <!-- Initiate Rotation -->
            <div v-else-if="canInitiateRotation" class="space-y-4">
                <div class="bg-muted/30 p-4 rounded-lg space-y-2">
                    <div class="flex items-start gap-2">
                        <AlertTriangle class="w-5 h-5 text-orange-500 mt-0.5" />
                        <div class="text-sm">
                            <p class="font-medium">Before rotating:</p>
                            <ul class="list-disc list-inside text-muted-foreground mt-1 space-y-1">
                                <li>All config values will be re-encrypted with a new key</li>
                                <li>Another admin will need to verify and approve the rotation</li>
                                <li>If config changes during approval, the rotation becomes stale</li>
                                <li>Rotation cannot be undone once committed</li>
                            </ul>
                        </div>
                    </div>
                </div>

                <Button
                    @click="initiateRotation"
                    :disabled="isRotating || !teamKey"
                >
                    <RefreshCw class="w-4 h-4 mr-2" :class="{ 'animate-spin': isRotating }" />
                    {{ isRotating ? 'Rotating...' : 'Rotate Encryption Key' }}
                </Button>

                <p v-if="!teamKey" class="text-sm text-muted-foreground">
                    Team key not available. Unable to rotate.
                </p>
            </div>

            <div v-else class="text-sm text-muted-foreground">
                Only project admins can rotate keys.
            </div>
        </CardContent>
    </Card>
</template>
