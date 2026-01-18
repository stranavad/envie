<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { Loader2 } from 'lucide-vue-next';
import { DeviceService, type Device } from '@/services/device.service';
import { useVaultStore } from '@/stores/vault';
import DeviceListItem from '@/components/identity/DeviceListItem.vue';
import ApproveDeviceDialog from '@/components/identity/ApproveDeviceDialog.vue';
import DeleteDeviceDialog from '@/components/identity/DeleteDeviceDialog.vue';

const vault = useVaultStore();

const devices = ref<Device[]>([]);
const isLoading = ref(false);
const error = ref('');

const isApprovalOpen = ref(false);
const deviceToApprove = ref<Device | null>(null);

const isDeleteOpen = ref(false);
const deviceToDelete = ref<Device | null>(null);

async function loadDevices() {
    isLoading.value = true;
    error.value = '';
    try {
        devices.value = await DeviceService.getDevices();
    } catch (e: any) {
        error.value = 'Error loading devices: ' + e.message;
    } finally {
        isLoading.value = false;
    }
}

function openApproveDialog(device: Device) {
    deviceToApprove.value = device;
    isApprovalOpen.value = true;
}

function openDeleteDialog(device: Device) {
    deviceToDelete.value = device;
    isDeleteOpen.value = true;
}

function isCurrentDevice(device: Device): boolean {
    return device.publicKey === vault.publicKey;
}

onMounted(() => {
    loadDevices();
});
</script>

<template>
    <div class="p-8 max-w-5xl mx-auto space-y-8">
        <div>
            <h1 class="text-3xl font-bold tracking-tight">Identities & Devices</h1>
            <p class="text-muted-foreground">Manage devices authorized to access your encrypted data.</p>
        </div>

        <div v-if="error" class="bg-destructive/15 text-destructive p-4 rounded-md text-sm">
            {{ error }}
        </div>

        <div class="bg-card rounded-lg border shadow-sm">
            <!-- Header -->
            <div class="grid grid-cols-12 gap-4 p-4 border-b font-medium text-sm text-muted-foreground">
                <div class="col-span-4">Device Name</div>
                <div class="col-span-4">Public Key / Status</div>
                <div class="col-span-3">Last Active</div>
                <div class="col-span-1 text-right">Actions</div>
            </div>

            <div v-if="isLoading && devices.length === 0" class="flex flex-col items-center py-12 text-muted-foreground">
                <Loader2 class="h-8 w-8 animate-spin mb-4" />
                <p>Loading devices...</p>
            </div>

            <div v-else-if="!isLoading && devices.length === 0" class="p-8 text-center text-muted-foreground">
                No devices found (Odd, you should be one).
            </div>

            <!-- List -->
            <div v-else class="divide-y divide-border">
                <DeviceListItem
                    v-for="device in devices"
                    :key="device.id"
                    :device="device"
                    :is-current-device="isCurrentDevice(device)"
                    @approve="openApproveDialog"
                    @delete="openDeleteDialog"
                />
            </div>
        </div>

        <ApproveDeviceDialog
            v-model:open="isApprovalOpen"
            :device="deviceToApprove"
            @approved="loadDevices"
        />

        <DeleteDeviceDialog
            v-model:open="isDeleteOpen"
            :device="deviceToDelete"
            @deleted="loadDevices"
        />
    </div>
</template>
