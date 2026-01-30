<script setup lang="ts">
import { ref, computed } from 'vue';
import { type Device } from '@/services/device.service';
import { useVaultStore } from '@/stores/vault';
import { PageLoader } from '@/components/ui/spinner';
import { ErrorState } from '@/components/ui/error-state';
import DeviceListItem from '@/components/identity/DeviceListItem.vue';
import ApproveDeviceDialog from '@/components/identity/ApproveDeviceDialog.vue';
import DeleteDeviceDialog from '@/components/identity/DeleteDeviceDialog.vue';
import { useDevices, queryKeys } from '@/queries';
import { useQueryClient } from '@tanstack/vue-query';

const vault = useVaultStore();
const queryClient = useQueryClient();

// TanStack Query for devices
const { data: devices, isLoading, error: queryError, refetch } = useDevices();

const isApprovalOpen = ref(false);
const deviceToApprove = ref<Device | null>(null);

const isDeleteOpen = ref(false);
const deviceToDelete = ref<Device | null>(null);

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

function handleDeviceChanged() {
    queryClient.invalidateQueries({ queryKey: queryKeys.devices });
}

const errorMessage = computed(() => {
    if (queryError.value) {
        return queryError.value instanceof Error ? queryError.value.message : String(queryError.value);
    }
    return '';
});
</script>

<template>
    <div class="p-8 max-w-5xl mx-auto space-y-8">
        <div>
            <h1 class="text-3xl font-bold tracking-tight">Identities & Devices</h1>
            <p class="text-muted-foreground">Manage devices authorized to access your encrypted data.</p>
        </div>

        <ErrorState
            v-if="errorMessage"
            title="Failed to load devices"
            :message="errorMessage"
            :retry="refetch"
        />

        <PageLoader v-else-if="isLoading" message="Loading devices..." />

        <div v-else class="bg-card rounded-lg border shadow-sm">
            <div class="grid grid-cols-12 gap-4 p-4 border-b font-medium text-sm text-muted-foreground">
                <div class="col-span-4">Device Name</div>
                <div class="col-span-4">Public Key / Status</div>
                <div class="col-span-3">Last Active</div>
                <div class="col-span-1 text-right">Actions</div>
            </div>

            <div v-if="!devices || devices.length === 0" class="p-8 text-center text-muted-foreground">
                No devices found (Odd, you should be one).
            </div>

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
            @approved="handleDeviceChanged"
        />

        <DeleteDeviceDialog
            v-model:open="isDeleteOpen"
            :device="deviceToDelete"
            @deleted="handleDeviceChanged"
        />
    </div>
</template>
