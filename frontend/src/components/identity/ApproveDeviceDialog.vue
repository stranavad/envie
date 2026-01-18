<script setup lang="ts">
import { ref, watch } from 'vue';
import { Button } from '@/components/ui/button';
import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogFooter,
    DialogHeader,
    DialogTitle,
} from '@/components/ui/dialog';
import type { Device } from '@/services/device.service';
import { DeviceService } from '@/services/device.service';
import { IdentityService } from '@/services/identity.service';
import { EncryptionService } from '@/services/encryption.service';

const props = defineProps<{
    device: Device | null;
    open: boolean;
}>();

const emit = defineEmits<{
    'update:open': [value: boolean];
    approved: [];
}>();

const isApproving = ref(false);
const error = ref('');

watch(() => props.open, (isOpen) => {
    if (isOpen) {
        error.value = '';
    }
});

async function handleApprove() {
    if (!props.device) return;

    isApproving.value = true;
    error.value = '';

    try {
        const masterKey = IdentityService.getMasterKey();
        if (!masterKey) {
            throw new Error('Your Master Identity Key is not available. Be sure your vault is unlocked.');
        }

        const encryptedBundle = await EncryptionService.encryptKey(
            props.device.publicKey,
            masterKey
        );

        await DeviceService.updateDevice(props.device.id, {
            encryptedMasterKey: encryptedBundle
        });

        emit('update:open', false);
        emit('approved');
    } catch (e: any) {
        error.value = e.message || 'Approval failed';
    } finally {
        isApproving.value = false;
    }
}

function handleClose() {
    emit('update:open', false);
}
</script>

<template>
    <Dialog :open="open" @update:open="emit('update:open', $event)">
        <DialogContent>
            <DialogHeader>
                <DialogTitle>Approve Device</DialogTitle>
                <DialogDescription>
                    Grant <strong>{{ device?.name }}</strong> access to your identity?
                </DialogDescription>
            </DialogHeader>

            <div class="text-sm text-muted-foreground space-y-2">
                <p>
                    This will encrypt your Master Identity Key with this device's Public Key.
                    This allows the device to decrypt your projects securely.
                </p>
                <div class="bg-muted p-2 rounded text-xs font-mono break-all">
                    {{ device?.publicKey }}
                </div>
            </div>

            <div v-if="error" class="text-sm text-destructive">
                {{ error }}
            </div>

            <DialogFooter>
                <Button variant="ghost" @click="handleClose">Cancel</Button>
                <Button @click="handleApprove" :disabled="isApproving">
                    {{ isApproving ? 'Approving...' : 'Approve Access' }}
                </Button>
            </DialogFooter>
        </DialogContent>
    </Dialog>
</template>
