<script setup lang="ts">
import { ref } from 'vue';
import { Button } from '@/components/ui/button';
import { Trash2, ShieldCheck, ShieldAlert, Monitor, Check, Copy } from 'lucide-vue-next';
import type { Device } from '@/services/device.service';

const props = defineProps<{
    device: Device;
    isCurrentDevice: boolean;
}>();

const emit = defineEmits<{
    approve: [device: Device];
    delete: [device: Device];
}>();

const copiedKey = ref(false);

function copyKey() {
    navigator.clipboard.writeText(props.device.publicKey);
    copiedKey.value = true;
    setTimeout(() => copiedKey.value = false, 2000);
}

function truncate(str: string, n: number) {
    return (str.length > n) ? str.slice(0, n - 1) + '...' : str;
}
</script>

<template>
    <div class="grid grid-cols-12 gap-4 p-4 items-center hover:bg-muted/30 transition-colors">
        <!-- Name -->
        <div class="col-span-4 flex items-center gap-3">
            <div
                class="p-2 rounded-full"
                :class="isCurrentDevice ? 'bg-primary/20 text-primary' : 'bg-muted text-muted-foreground'"
            >
                <Monitor class="w-4 h-4" />
            </div>
            <div>
                <div class="font-medium flex items-center gap-2">
                    {{ device.name }}
                    <span
                        v-if="isCurrentDevice"
                        class="text-xs bg-primary/10 text-primary px-1.5 py-0.5 rounded"
                    >
                        You
                    </span>
                </div>
                <div
                    class="text-xs text-muted-foreground font-mono truncate max-w-[150px]"
                    :title="device.id"
                >
                    ID: {{ truncate(device.id, 8) }}
                </div>
            </div>
        </div>

        <!-- Status / Key -->
        <div class="col-span-4 space-y-1">
            <!-- Verification Status -->
            <div
                v-if="device.encryptedMasterKey"
                class="flex items-center gap-1.5 text-xs text-green-600 bg-green-500/15 px-2 py-1 rounded w-fit"
            >
                <ShieldCheck class="w-3 h-3" />
                Verified
            </div>
            <div
                v-else
                class="flex items-center gap-1.5 text-xs text-yellow-600 bg-yellow-500/15 px-2 py-1 rounded w-fit cursor-pointer hover:underline"
                @click="emit('approve', device)"
            >
                <ShieldAlert class="w-3 h-3" />
                Pending Approval
            </div>

            <!-- Public Key -->
            <div
                class="flex items-center gap-2 text-xs text-muted-foreground font-mono group cursor-pointer"
                @click="copyKey"
            >
                {{ truncate(device.publicKey, 16) }}
                <Check v-if="copiedKey" class="w-3 h-3 text-green-500" />
                <Copy v-else class="w-3 h-3 opacity-0 group-hover:opacity-100 transition-opacity" />
            </div>
        </div>

        <!-- Last Active -->
        <div class="col-span-3 text-sm text-muted-foreground">
            {{ new Date(device.lastActive).toLocaleDateString() }}
            <span class="text-xs opacity-50">
                {{ new Date(device.lastActive).toLocaleTimeString() }}
            </span>
        </div>

        <!-- Actions -->
        <div class="col-span-1 text-right flex justify-end gap-2">
            <Button
                v-if="!device.encryptedMasterKey && !isCurrentDevice"
                size="sm"
                variant="secondary"
                class="h-8 px-2"
                @click="emit('approve', device)"
                title="Approve Device"
            >
                Approve
            </Button>

            <Button
                variant="ghost"
                size="icon"
                class="h-8 w-8 text-muted-foreground hover:text-destructive hover:bg-destructive/10"
                @click="emit('delete', device)"
                :disabled="isCurrentDevice"
                :title="isCurrentDevice ? 'Cannot delete current device' : 'Remove Device'"
            >
                <Trash2 class="w-4 h-4" />
            </Button>
        </div>
    </div>
</template>
