<script setup lang="ts">
import {computed, onMounted, onUnmounted, ref} from 'vue';
import {useAuthStore} from '@/stores/auth';
import {useVaultStore} from '@/stores/vault';
import {IdentityService} from '@/services/identity.service';
import {Button} from '@/components/ui/button';
import {Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle} from '@/components/ui/card';
import {Input} from '@/components/ui/input';
import {Label} from '@/components/ui/label';
import {config} from '@/config';
import {Check, Copy, KeyRound, Loader2, Monitor, ShieldCheck, Smartphone, Trash2} from 'lucide-vue-next';
import {EncryptionService} from '@/services/encryption.service';

const auth = useAuthStore();
const vault = useVaultStore();

type Step = 'check' | 'options' | 'generate' | 'recover' | 'waiting_approval' | 'registering';
const currentStep = ref<Step>('check');

const mnemonic = ref<string>('');
const mnemonicWords = computed(() => mnemonic.value.split(' '));
const enteredMnemonic = ref<string>('');
const deviceName = ref<string>('My Device');
const isLoading = ref(false);
const error = ref('');

const hasExistingDevices = ref(false);
let pollingInterval: ReturnType<typeof setInterval> | null = null;

const emit = defineEmits(['completed']);

// ============ CASE A: Request Approval from Another Device ============
async function startWaitingForApproval() {
    if (!vault.publicKey) {
        error.value = "Device public key not found. Is vault unlocked?";
        return;
    }

    isLoading.value = true;
    error.value = '';

    try {
        // Register device as pending (without encryptedMasterKey)
        const response = await fetch(`${config.backendUrl}/devices`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${auth.token}`
            },
            body: JSON.stringify({
                name: deviceName.value,
                publicKey: vault.publicKey
                // No encryptedMasterKey = pending approval
            })
        });

        if (!response.ok) {
            throw new Error("Failed to register device");
        }

        // Start polling for approval
        currentStep.value = 'waiting_approval';
        startPolling();

    } catch (e: any) {
        error.value = "Failed to request approval: " + e.message;
    } finally {
        isLoading.value = false;
    }
}

function startPolling() {
    // Poll every 3 seconds
    pollingInterval = setInterval(checkApprovalStatus, 3000);
}

function stopPolling() {
    if (pollingInterval) {
        clearInterval(pollingInterval);
        pollingInterval = null;
    }
}

async function checkApprovalStatus() {
    try {
        const res = await fetch(`${config.backendUrl}/devices`, {
            headers: { 'Authorization': `Bearer ${auth.token}` }
        });

        if (!res.ok) return;

        const devices = await res.json();
        const myDevice = devices.find((d: any) => d.publicKey === vault.publicKey);

        if (myDevice?.encryptedMasterKey) {
            // We've been approved!
            stopPolling();

            try {
                let bundle = myDevice.encryptedMasterKey;
                if (bundle.startsWith('"')) {
                    bundle = JSON.parse(bundle);
                }

                const masterKey = await EncryptionService.decryptKey(vault.privateKey, bundle);
                IdentityService.setMasterKey(masterKey);
                await vault.saveMasterKey(masterKey);
                emit('completed');
            } catch (e) {
                error.value = "Failed to decrypt master key: " + e;
                currentStep.value = 'options';
            }
        }
    } catch (e) {
        console.error("Polling error:", e);
    }
}

// ============ CASE B: Recover with Recovery Phrase ============
function startRecover() {
    enteredMnemonic.value = '';
    error.value = '';
    currentStep.value = 'recover';
}

async function submitRecovery() {
    if (!IdentityService.validateRecoveryPhrase(enteredMnemonic.value)) {
        error.value = "Invalid recovery phrase. Please check your words.";
        return;
    }

    isLoading.value = true;
    error.value = '';

    try {
        const masterKey = await IdentityService.deriveMasterKey(enteredMnemonic.value);
        await registerDevice(masterKey);
    } catch (e: any) {
        error.value = e.message;
        currentStep.value = 'recover';
    } finally {
        isLoading.value = false;
    }
}

// ============ CASE C: Destructive Reset ============
async function startDestructiveReset() {
    isLoading.value = true;
    error.value = '';

    try {
        // Delete all existing devices
        const deleteRes = await fetch(`${config.backendUrl}/devices`, {
            method: 'DELETE',
            headers: { 'Authorization': `Bearer ${auth.token}` }
        });

        if (!deleteRes.ok) {
            throw new Error("Failed to delete existing devices");
        }

        // Generate new identity
        mnemonic.value = IdentityService.generateRecoveryPhrase();
        currentStep.value = 'generate';

    } catch (e: any) {
        error.value = "Reset failed: " + e.message;
    } finally {
        isLoading.value = false;
    }
}

// ============ New Account: Generate New Identity ============
function startGenerate() {
    mnemonic.value = IdentityService.generateRecoveryPhrase();
    error.value = '';
    currentStep.value = 'generate';
}

async function submitGenerate() {
    isLoading.value = true;
    error.value = '';

    try {
        const masterKey = await IdentityService.deriveMasterKey(mnemonic.value);
        await registerDevice(masterKey);
    } catch (e: any) {
        error.value = e.message;
        currentStep.value = 'generate';
    } finally {
        isLoading.value = false;
    }
}

// ============ Common: Register Device ============
async function registerDevice(masterKey: string) {
    if (!vault.publicKey) {
        throw new Error("Device public key not found. Is vault unlocked?");
    }

    currentStep.value = 'registering';

    const encryptedMasterKeyBundle = await EncryptionService.encryptKey(vault.publicKey, masterKey);

    const response = await fetch(`${config.backendUrl}/devices`, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
            'Authorization': `Bearer ${auth.token}`
        },
        body: JSON.stringify({
            name: deviceName.value,
            publicKey: vault.publicKey,
            encryptedMasterKey: encryptedMasterKeyBundle
        })
    });

    if (!response.ok) {
        throw new Error("Failed to register device");
    }

    IdentityService.setMasterKey(masterKey);
    await vault.saveMasterKey(masterKey);
    emit('completed');
}

// ============ Initial Check ============
async function checkStatus() {
    try {
        const res = await fetch(`${config.backendUrl}/devices`, {
            headers: { 'Authorization': `Bearer ${auth.token}` }
        });

        if (res.ok) {
            const devices = await res.json();

            if (devices && devices.length > 0) {
                hasExistingDevices.value = true;
            }

            const myDevice = devices.find((d: any) => d.publicKey === vault.publicKey);

            if (myDevice) {
                if (myDevice.encryptedMasterKey) {
                    // Already approved - decrypt and continue
                    try {
                        let bundle = myDevice.encryptedMasterKey;
                        if (bundle.startsWith('"')) {
                            bundle = JSON.parse(bundle);
                        }

                        const masterKey = await EncryptionService.decryptKey(vault.privateKey, bundle);
                        IdentityService.setMasterKey(masterKey);
                        await vault.saveMasterKey(masterKey);
                        emit('completed');
                        return;
                    } catch (e) {
                        console.error("Failed to decrypt master key", e);
                        error.value = "Failed to unlock identity. You may need to re-verify.";
                    }
                } else {
                    // Device registered but pending approval - resume waiting
                    currentStep.value = 'waiting_approval';
                    startPolling();
                    return;
                }
            }
        }
    } catch (e) {
        console.error("Check status failed", e);
    }

    currentStep.value = 'options';
}

// ============ Lifecycle ============
onMounted(() => {
    checkStatus();
});

onUnmounted(() => {
    stopPolling();
});

// Clipboard
const copied = ref(false);
function copyToClipboard() {
    navigator.clipboard.writeText(mnemonic.value);
    copied.value = true;
    setTimeout(() => copied.value = false, 2000);
}

function cancelWaiting() {
    stopPolling();
    currentStep.value = 'options';
}
</script>

<template>
    <div class="flex items-center justify-center min-h-screen bg-background">
        <Card class="w-full max-w-lg">
            <CardHeader>
                <CardTitle>{{ hasExistingDevices ? 'Add New Device' : 'Device Setup' }}</CardTitle>
                <CardDescription>
                    {{ hasExistingDevices
                        ? 'We found existing devices for this account. Choose how to authorize this device.'
                        : 'Secure your new account with a master identity.'
                    }}
                </CardDescription>
            </CardHeader>

            <CardContent>
                <!-- Loading / Check -->
                <div v-if="currentStep === 'check' || currentStep === 'registering'" class="flex flex-col items-center py-8">
                    <Loader2 class="h-8 w-8 animate-spin mb-4" />
                    <p class="text-muted-foreground">{{ currentStep === 'registering' ? 'Registering device...' : 'Checking registration...' }}</p>
                </div>

                <!-- Options (has existing devices) -->
                <div v-if="currentStep === 'options' && hasExistingDevices" class="space-y-3">
                    <!-- Option A: Approve from another device -->
                    <Button
                        variant="outline"
                        class="w-full h-auto py-4 flex items-start gap-4 justify-start text-left"
                        @click="startWaitingForApproval"
                        :disabled="isLoading"
                    >
                        <div class="p-2 bg-primary/10 rounded-lg shrink-0">
                            <Smartphone class="h-5 w-5 text-primary" />
                        </div>
                        <div class="space-y-1">
                            <div class="font-semibold">Approve from another device</div>
                            <div class="text-xs text-muted-foreground font-normal">
                                Open your existing device and approve this new one
                            </div>
                        </div>
                    </Button>

                    <!-- Option B: Recovery phrase -->
                    <Button
                        variant="outline"
                        class="w-full h-auto py-4 flex items-start gap-4 justify-start text-left"
                        @click="startRecover"
                        :disabled="isLoading"
                    >
                        <div class="p-2 bg-primary/10 rounded-lg shrink-0">
                            <KeyRound class="h-5 w-5 text-primary" />
                        </div>
                        <div class="space-y-1">
                            <div class="font-semibold">Enter recovery phrase</div>
                            <div class="text-xs text-muted-foreground font-normal">
                                Use your 12-word recovery phrase
                            </div>
                        </div>
                    </Button>

                    <!-- Option C: Destructive reset -->
                    <Button
                        variant="outline"
                        class="w-full h-auto py-4 flex items-start gap-4 justify-start text-left border-destructive/30 hover:bg-destructive/5"
                        @click="startDestructiveReset"
                        :disabled="isLoading"
                    >
                        <div class="p-2 bg-destructive/10 rounded-lg shrink-0">
                            <Trash2 class="h-5 w-5 text-destructive" />
                        </div>
                        <div class="space-y-1">
                            <div class="font-semibold text-destructive">Reset identity</div>
                            <div class="text-xs text-muted-foreground font-normal">
                                Delete all devices and start fresh (destructive)
                            </div>
                        </div>
                    </Button>

                    <div class="space-y-2 pt-2">
                        <Label>Device Name</Label>
                        <Input v-model="deviceName" placeholder="e.g. MacBook Pro" />
                    </div>
                </div>

                <!-- Options (new account) -->
                <div v-if="currentStep === 'options' && !hasExistingDevices" class="space-y-4">
                    <Button
                        variant="outline"
                        class="w-full h-32 flex flex-col gap-2 border-primary/20 hover:bg-primary/5"
                        @click="startGenerate"
                    >
                        <ShieldCheck class="h-8 w-8 text-primary" />
                        <span class="text-primary font-semibold">Create New Identity</span>
                        <span class="text-xs text-muted-foreground font-normal">Generate 12-word phrase to secure your account</span>
                    </Button>

                    <div class="space-y-2">
                        <Label>Device Name</Label>
                        <Input v-model="deviceName" placeholder="e.g. MacBook Pro" />
                    </div>
                </div>

                <!-- Waiting for Approval -->
                <div v-if="currentStep === 'waiting_approval'" class="space-y-6 py-4">
                    <div class="flex flex-col items-center text-center space-y-4">
                        <div class="relative">
                            <Monitor class="h-16 w-16 text-primary" />
                            <div class="absolute -bottom-1 -right-1 bg-yellow-500 rounded-full p-1">
                                <Loader2 class="h-4 w-4 animate-spin text-white" />
                            </div>
                        </div>
                        <div class="space-y-2">
                            <h3 class="font-semibold text-lg">Waiting for approval</h3>
                            <p class="text-sm text-muted-foreground">
                                Open the <strong>Identities</strong> page on one of your existing devices and approve this device.
                            </p>
                        </div>
                    </div>

                    <div class="p-4 bg-muted rounded-lg space-y-2">
                        <div class="text-sm font-medium">Device: {{ deviceName }}</div>
                        <div class="text-xs text-muted-foreground font-mono break-all">
                            {{ vault.publicKey }}
                        </div>
                    </div>

                    <Button variant="ghost" class="w-full" @click="cancelWaiting">
                        Cancel
                    </Button>
                </div>

                <!-- Generate Flow -->
                <div v-if="currentStep === 'generate'" class="space-y-4">
                    <div class="p-4 bg-yellow-100 dark:bg-yellow-900/30 text-yellow-800 dark:text-yellow-200 rounded-md border border-yellow-200 dark:border-yellow-900">
                        <h4 class="font-medium mb-1">Save this phrase!</h4>
                        <p class="text-sm">
                            If you lose this phrase, you will need to approve new devices from an existing one or reset your identity.
                        </p>
                    </div>

                    <div class="p-4 bg-muted rounded-lg relative group cursor-pointer" @click="copyToClipboard">
                        <div class="absolute top-2 right-2 opacity-0 group-hover:opacity-100 transition-opacity">
                            <Check v-if="copied" class="h-4 w-4 text-green-500" />
                            <Copy v-else class="h-4 w-4 text-muted-foreground" />
                        </div>
                        <div class="grid grid-cols-3 gap-2 text-sm font-mono">
                            <div v-for="(word, i) in mnemonicWords" :key="i" class="flex gap-2">
                                <span class="text-muted-foreground select-none">{{ i + 1 }}.</span>
                                <span class="font-bold">{{ word }}</span>
                            </div>
                        </div>
                    </div>

                    <div class="space-y-2">
                        <Label>Device Name</Label>
                        <Input v-model="deviceName" placeholder="e.g. MacBook Pro" />
                    </div>
                </div>

                <!-- Recover Flow -->
                <div v-if="currentStep === 'recover'" class="space-y-4">
                    <div class="space-y-2">
                        <Label>Enter your 12-word phrase</Label>
                        <Input v-model="enteredMnemonic" placeholder="apple banana cherry ..." class="font-mono" />
                        <p class="text-xs text-muted-foreground">Separate words with spaces.</p>
                    </div>
                    <div class="space-y-2">
                        <Label>Device Name</Label>
                        <Input v-model="deviceName" placeholder="e.g. MacBook Pro" />
                    </div>
                </div>

                <!-- Error -->
                <div v-if="error" class="text-sm text-destructive mt-4">{{ error }}</div>
            </CardContent>

            <CardFooter class="flex justify-between" v-if="currentStep !== 'check' && currentStep !== 'registering' && currentStep !== 'waiting_approval'">
                <Button variant="ghost" @click="currentStep = 'options'" v-if="currentStep !== 'options'">Back</Button>
                <div v-else></div>

                <Button v-if="currentStep === 'generate'" @click="submitGenerate" :disabled="isLoading">
                    {{ isLoading ? 'Creating...' : 'I have saved it' }}
                </Button>
                <Button v-if="currentStep === 'recover'" @click="submitRecovery" :disabled="isLoading">
                    {{ isLoading ? 'Verifying...' : 'Verify & Add Device' }}
                </Button>
            </CardFooter>
        </Card>
    </div>
</template>
