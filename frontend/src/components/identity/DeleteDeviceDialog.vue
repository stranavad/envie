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

const props = defineProps<{
    device: Device | null;
    open: boolean;
}>();

const emit = defineEmits<{
    'update:open': [value: boolean];
    deleted: [];
}>();

const isDeleting = ref(false);
const error = ref('');

watch(() => props.open, (isOpen) => {
    if (isOpen) {
        error.value = '';
    }
});

async function handleDelete() {
    if (!props.device) return;

    isDeleting.value = true;
    error.value = '';

    try {
        await DeviceService.deleteDevice(props.device.id);
        emit('update:open', false);
        emit('deleted');
    } catch (e: any) {
        error.value = e.message || 'Delete failed';
    } finally {
        isDeleting.value = false;
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
                <DialogTitle class="text-destructive">Remove Device</DialogTitle>
                <DialogDescription>
                    Are you sure you want to remove <strong>{{ device?.name }}</strong>?
                </DialogDescription>
            </DialogHeader>

            <div class="text-sm">
                This action is permanent. This device will lose access to all your encrypted data immediately.
            </div>

            <div v-if="error" class="text-sm text-destructive">
                {{ error }}
            </div>

            <DialogFooter>
                <Button variant="ghost" @click="handleClose">Cancel</Button>
                <Button variant="destructive" @click="handleDelete" :disabled="isDeleting">
                    {{ isDeleting ? 'Removing...' : 'Remove Device' }}
                </Button>
            </DialogFooter>
        </DialogContent>
    </Dialog>
</template>
