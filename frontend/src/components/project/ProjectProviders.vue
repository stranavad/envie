<script setup lang="ts">
import {ref, watch} from 'vue';
import {type SecretManagerConfig, SecretManagerConfigService} from '@/services/secret-manager-config.service';
import { SectionHeader } from '@/components/ui/section-header';
import { EmptyState } from '@/components/ui/empty-state';
import { Plus, Cloud } from 'lucide-vue-next';
import SecretManagerForm from '@/components/project/SecretManagerForm.vue';
import SecretManagerRow from '@/components/project/SecretManagerRow.vue';
import {Project} from "@/services/project.service.ts";
import { useSecretManagerStore } from '@/stores/secret-manager.store';
import { useConfigEncryption } from '@/composables/useConfigEncryption';

const props = defineProps<{
    project: Project;
    decryptedKey: string;
}>();

const configs = ref<SecretManagerConfig[]>([]);
const isLoading = ref(false);
const error = ref('');

const connectionStatuses = ref<Record<string, 'pending' | 'success' | 'error' | 'unknown'>>({});

const isAdding = ref(false);
const editingId = ref<string | null>(null);
const secretManagerStore = useSecretManagerStore();
const { decryptSecretManagerConfig } = useConfigEncryption();

watch(() => props.decryptedKey, (newKey) => {
    if (newKey && configs.value.length > 0) {
        checkAllConnections();
    }
});

async function fetchConfigs() {
    isLoading.value = true;
    try {
        configs.value = await SecretManagerConfigService.getConfigs(props.project.id);
        await checkAllConnections();
    } catch (e: any) {
        error.value = "Failed to load configurations: " + e.toString();
    } finally {
        isLoading.value = false;
    }
}

async function checkAllConnections() {
    await Promise.all(configs.value.map(async(config) => {
      connectionStatuses.value[config.id] = 'pending';

      try {
        const decryptedJson = await decryptSecretManagerConfig(props.decryptedKey, config.encryptedKey);

        const success = await secretManagerStore.testConnection(config.id, decryptedJson);
        connectionStatuses.value[config.id] = success ? 'success' : 'error';
      } catch (e) {
        console.error("Failed to check connection for config", config.id, e);
        connectionStatuses.value[config.id] = 'error';
      }
    }))
}

function startAdding() {
    editingId.value = null; // Close any edit
    isAdding.value = true;
}

function cancelAdding() {
    isAdding.value = false;
}

function startEditing(config: SecretManagerConfig) {
    isAdding.value = false;
    editingId.value = config.id;
}

function cancelEdit() {
    editingId.value = null;
}

async function onSaved() {
    isAdding.value = false;
    if (editingId.value) {
        secretManagerStore.clearCache(editingId.value);
    }
    editingId.value = null;
    await fetchConfigs();
}


async function confirmDelete(id: string) {
    try {
        await SecretManagerConfigService.deleteConfig(props.project.id, id);
        secretManagerStore.clearCache(id);
        await fetchConfigs();
    } catch (e: any) {
         error.value = "Failed to delete: " + e.toString();
    }
}

async function onCreated(){
  await fetchConfigs();
}

onCreated()
</script>

<template>
    <div class="space-y-6">
        <SectionHeader
            title="Google Secret Manager"
            description="Connect to external secret managers to sync and import secrets."
            :action-label="!isAdding && !editingId ? 'Add Configuration' : undefined"
            :action-icon="Plus"
            @action="startAdding"
        />

        <div v-if="error" class="p-4 text-sm text-destructive bg-destructive/15 rounded-md">
            {{ error }}
        </div>

        <SecretManagerForm
            v-if="isAdding" 
            :project="project" 
            :decrypted-key="decryptedKey" 
            @saved="onSaved" 
            @cancel="cancelAdding" 
        />

        <EmptyState
            v-if="!isAdding && configs.length === 0"
            :icon="Cloud"
            title="No Secret Manager configurations found"
            description="Add a configuration to sync secrets from Google Secret Manager."
        />

        <div v-else class="grid gap-4 md:grid-cols-1">
            <template v-for="config in configs" :key="config.id">
                <SecretManagerForm
                    v-if="editingId === config.id"
                    :project="project"
                    :decrypted-key="decryptedKey"
                    :initial-config="config"
                    @saved="onSaved"
                    @cancel="cancelEdit"
                />

                <SecretManagerRow
                    v-else
                    :config="config"
                    :status="connectionStatuses[config.id] || 'unknown'"
                    @edit="startEditing(config)"
                    @delete="confirmDelete(config.id)"
                />
            </template>
        </div>
    </div>
</template>
